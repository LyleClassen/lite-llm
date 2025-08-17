package cmd

import (
	"context"
	"fmt"

	"github.com/lyleclassen/lite-llm/internal/ollama"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "Manage LLM models",
	Long:  `Manage local language models including downloading, listing, and removing models.`,
}

var listModelsCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed models",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runListModels()
	},
}

var downloadModelCmd = &cobra.Command{
	Use:   "download [model-name]",
	Short: "Download a model",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDownloadModel(args[0])
	},
}

var removeModelCmd = &cobra.Command{
	Use:   "remove [model-name]",
	Short: "Remove a model",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRemoveModel(args[0])
	},
}

var recommendedCmd = &cobra.Command{
	Use:   "recommended",
	Short: "Download recommended models for your hardware",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDownloadRecommended()
	},
}

func init() {
	rootCmd.AddCommand(modelsCmd)
	modelsCmd.AddCommand(listModelsCmd)
	modelsCmd.AddCommand(downloadModelCmd)
	modelsCmd.AddCommand(removeModelCmd)
	modelsCmd.AddCommand(recommendedCmd)
}

func runListModels() error {
	client := ollama.NewClient("http://localhost:11434")
	
	models, err := client.ListModels(context.Background())
	if err != nil {
		return fmt.Errorf("failed to list models: %w", err)
	}

	if len(models) == 0 {
		logrus.Info("No models installed")
		logrus.Info("Use 'lite-llm models recommended' to download recommended models")
		return nil
	}

	logrus.Info("Installed models:")
	for _, model := range models {
		logrus.Infof("  - %s (%.1f GB)", model.Name, float64(model.Size)/(1024*1024*1024))
	}

	return nil
}

func runDownloadModel(modelName string) error {
	client := ollama.NewClient("http://localhost:11434")
	
	logrus.Infof("Downloading model: %s", modelName)
	
	err := client.PullModel(context.Background(), modelName, func(progress ollama.PullProgress) {
		if progress.Total > 0 {
			percent := float64(progress.Completed) / float64(progress.Total) * 100
			logrus.Infof("Progress: %.1f%% (%s)", percent, progress.Status)
		} else {
			logrus.Info(progress.Status)
		}
	})

	if err != nil {
		return fmt.Errorf("failed to download model: %w", err)
	}

	logrus.Infof("Successfully downloaded model: %s", modelName)
	return nil
}

func runRemoveModel(modelName string) error {
	client := ollama.NewClient("http://localhost:11434")
	
	logrus.Infof("Removing model: %s", modelName)
	
	err := client.DeleteModel(context.Background(), modelName)
	if err != nil {
		return fmt.Errorf("failed to remove model: %w", err)
	}

	logrus.Infof("Successfully removed model: %s", modelName)
	return nil
}

func runDownloadRecommended() error {
	// Recommended models for RX 570/580 (8GB VRAM)
	recommendedModels := []string{
		"llama3.1:8b-instruct-q4_K_M",  // ~4.4GB
		"mistral:7b-instruct-q4_K_M",   // ~4.4GB  
		"gemma2:2b-instruct-q4_K_M",    // ~1.7GB
	}

	client := ollama.NewClient("http://localhost:11434")

	logrus.Info("Downloading recommended models for AMD RX 570/580...")
	
	for _, model := range recommendedModels {
		logrus.Infof("Downloading %s...", model)
		
		err := client.PullModel(context.Background(), model, func(progress ollama.PullProgress) {
			if progress.Total > 0 {
				percent := float64(progress.Completed) / float64(progress.Total) * 100
				logrus.Infof("  Progress: %.1f%% (%s)", percent, progress.Status)
			}
		})

		if err != nil {
			logrus.Errorf("Failed to download %s: %v", model, err)
			continue
		}

		logrus.Infof("✓ Successfully downloaded %s", model)
	}

	logrus.Info("Recommended models download complete!")
	logrus.Info("You can now use these models via the web interface at http://localhost:3000")
	
	return nil
}