package web

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lyleclassen/lite-llm/internal/ollama"
	"github.com/sirupsen/logrus"
)

type Server struct {
	ollama *ollama.Client
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

type ChatResponse struct {
	Message ChatMessage `json:"message"`
	Done    bool        `json:"done"`
}

func NewServer(ollamaURL string) *Server {
	return &Server{
		ollama: ollama.NewClient(ollamaURL),
	}
}

func (s *Server) SetupRoutes() *gin.Engine {
	// Set gin to release mode for production
	gin.SetMode(gin.ReleaseMode)
	
	r := gin.Default()

	// Serve static files
	r.Static("/static", "./web/static")
	r.LoadHTMLGlob("web/templates/*")

	// Web interface routes
	r.GET("/", s.handleIndex)
	r.GET("/chat", s.handleChat)

	// API routes
	api := r.Group("/api")
	{
		api.GET("/models", s.handleListModels)
		api.POST("/chat", s.handleChatAPI)
		api.GET("/health", s.handleHealth)
	}

	return r
}

func (s *Server) handleIndex(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "Lite LLM",
	})
}

func (s *Server) handleChat(c *gin.Context) {
	c.HTML(http.StatusOK, "chat.html", gin.H{
		"title": "Chat - Lite LLM",
	})
}

func (s *Server) handleListModels(c *gin.Context) {
	models, err := s.ollama.ListModels(context.Background())
	if err != nil {
		logrus.Errorf("Failed to list models: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list models"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"models": models})
}

func (s *Server) handleChatAPI(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.Messages) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No messages provided"})
		return
	}

	// Get the last message as the prompt
	lastMessage := req.Messages[len(req.Messages)-1]
	
	// Build context from previous messages
	prompt := ""
	for _, msg := range req.Messages {
		if msg.Role == "user" {
			prompt += "User: " + msg.Content + "\n"
		} else if msg.Role == "assistant" {
			prompt += "Assistant: " + msg.Content + "\n"
		}
	}
	
	// Generate response
	resp, err := s.ollama.Generate(context.Background(), req.Model, prompt, nil)
	if err != nil {
		logrus.Errorf("Failed to generate response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate response"})
		return
	}

	chatResp := ChatResponse{
		Message: ChatMessage{
			Role:    "assistant",
			Content: resp.Response,
		},
		Done: resp.Done,
	}

	c.JSON(http.StatusOK, chatResp)
}

func (s *Server) handleHealth(c *gin.Context) {
	err := s.ollama.Health(context.Background())
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}