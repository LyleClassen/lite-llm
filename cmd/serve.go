package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lyleclassen/lite-llm/internal/web"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the web interface server",
	Long: `Start a web server that provides a custom interface for interacting
with your LLM models. This serves as an alternative to Open WebUI.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runServe()
	},
}

var (
	port int
	host string
)

func init() {
	rootCmd.AddCommand(serveCmd)
	
	serveCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to serve on")
	serveCmd.Flags().StringVar(&host, "host", "0.0.0.0", "Host to bind to")
}

func runServe() error {
	logrus.Infof("Starting lite-llm web server on %s:%d", host, port)

	// Create web server
	server := web.NewServer("http://localhost:11434")
	router := server.SetupRoutes()

	httpServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", host, port),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("Failed to start server: %v", err)
		}
	}()

	logrus.Infof("Server started successfully!")
	logrus.Infof("Web interface: http://localhost:%d", port)
	logrus.Infof("API endpoint: http://localhost:%d/api", port)

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down server...")

	// Give it 30 seconds to finish existing requests
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		logrus.Errorf("Server forced to shutdown: %v", err)
		return err
	}

	logrus.Info("Server exited")
	return nil
}