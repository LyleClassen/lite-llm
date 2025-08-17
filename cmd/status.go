package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/lyleclassen/lite-llm/internal/monitor"
	"github.com/lyleclassen/lite-llm/internal/ollama"
	"github.com/lyleclassen/lite-llm/internal/system"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check system and service status",
	Long:  `Get detailed status information about the LLM deployment, system health, and service availability.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStatus()
	},
}

var (
	watch    bool
	interval int
)

func init() {
	rootCmd.AddCommand(statusCmd)
	
	statusCmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch status continuously")
	statusCmd.Flags().IntVarP(&interval, "interval", "i", 5, "Update interval in seconds (when watching)")
}

func runStatus() error {
	if watch {
		return runStatusWatch()
	}
	
	return printStatus()
}

func runStatusWatch() error {
	logrus.Infof("Watching status (updating every %d seconds, press Ctrl+C to stop)", interval)
	
	for {
		// Clear screen (works on most terminals)
		fmt.Print("\033[2J\033[H")
		
		if err := printStatus(); err != nil {
			logrus.Errorf("Error getting status: %v", err)
		}
		
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func printStatus() error {
	ctx := context.Background()
	
	logrus.Info("=== Lite LLM Status ===")
	logrus.Infof("Timestamp: %s", time.Now().Format("2006-01-02 15:04:05"))
	logrus.Info("")

	// System Information
	logrus.Info("=== System Information ===")
	checker := system.NewChecker()
	sysInfo, err := checker.GetSystemInfo()
	if err != nil {
		logrus.Errorf("Failed to get system info: %v", err)
	} else {
		logrus.Infof("Kernel: %s", sysInfo.KernelVersion)
		logrus.Infof("Docker: %v", formatStatus(sysInfo.HasDocker))
		logrus.Infof("AMD GPU: %v", formatStatus(sysInfo.HasAMDGPU))
		if sysInfo.HasAMDGPU {
			logrus.Infof("  Model: %s", sysInfo.GPUModel)
			logrus.Infof("  Memory: %d MB", sysInfo.GPUMemory)
		}
		logrus.Infof("ROCm: %v", formatStatus(sysInfo.HasROCm))
		logrus.Infof("System Memory: %d MB", sysInfo.SystemMemory)
	}
	logrus.Info("")

	// Container Services (via Portainer)
	logrus.Info("=== Container Services ===")
	logrus.Info("Note: Use Portainer to manage Docker containers")
	logrus.Info("Expected containers: ollama-amd, open-webui")
	logrus.Info("")

	// Ollama Service
	logrus.Info("=== Ollama Service ===")
	ollamaClient := ollama.NewClient("http://localhost:11434")
	
	err = ollamaClient.Health(ctx)
	if err != nil {
		logrus.Errorf("Ollama API: %v", formatStatus(false))
		logrus.Errorf("  Error: %v", err)
	} else {
		logrus.Infof("Ollama API: %v", formatStatus(true))
		logrus.Info("  Endpoint: http://localhost:11434")
		
		// Get models
		models, err := ollamaClient.ListModels(ctx)
		if err != nil {
			logrus.Errorf("  Models: Failed to list (%v)", err)
		} else {
			logrus.Infof("  Models: %d installed", len(models))
			for _, model := range models {
				logrus.Infof("    - %s (%.1f GB)", model.Name, float64(model.Size)/(1024*1024*1024))
			}
		}
	}
	logrus.Info("")

	// Web Interfaces
	logrus.Info("=== Web Interfaces ===")
	
	// Check Open WebUI
	webUIStatus := checkWebInterface("http://localhost:3000")
	logrus.Infof("Open WebUI: %v", formatStatus(webUIStatus))
	if webUIStatus {
		logrus.Info("  URL: http://localhost:3000")
	}
	
	// Check custom web interface (if it's running)
	customWebStatus := checkWebInterface("http://localhost:8080")
	logrus.Infof("Custom Web: %v", formatStatus(customWebStatus))
	if customWebStatus {
		logrus.Info("  URL: http://localhost:8080")
	}
	logrus.Info("")

	// Performance Metrics (if available)
	logrus.Info("=== Performance Metrics ===")
	metrics := monitor.GetPerformanceMetrics()
	if metrics != nil {
		logrus.Infof("CPU Usage: %.1f%%", metrics.CPUUsage)
		logrus.Infof("Memory Usage: %.1f%% (%d MB / %d MB)", 
			metrics.MemoryUsagePercent, 
			metrics.MemoryUsedMB, 
			metrics.MemoryTotalMB)
		if metrics.GPUUsage >= 0 {
			logrus.Infof("GPU Usage: %.1f%%", metrics.GPUUsage)
		}
	} else {
		logrus.Info("Performance metrics unavailable")
	}

	return nil
}

func formatStatus(status bool) string {
	if status {
		return "✓ Running"
	}
	return "✗ Not Running"
}

func checkWebInterface(url string) bool {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	
	return resp.StatusCode == 200
}