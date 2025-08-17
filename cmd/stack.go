package cmd

import (
	"fmt"
	"os"

	"github.com/lyleclassen/lite-llm/internal/templates"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var stackCmd = &cobra.Command{
	Use:   "stack",
	Short: "Generate Portainer stack templates",
	Long:  `Generate optimized Portainer stack templates for deploying LLM services with AMD GPU support.`,
}

var generateStackCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate Portainer stack template",
	Long: `Generate a Portainer-compatible Docker Compose stack template optimized for 
AMD GPU acceleration with Ollama and Open WebUI.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGenerateStack()
	},
}

var (
	outputFile string
	stackName  string
	ollamaPort int
	webuiPort  int
	gpuType    string
)

func init() {
	rootCmd.AddCommand(stackCmd)
	stackCmd.AddCommand(generateStackCmd)
	
	generateStackCmd.Flags().StringVarP(&outputFile, "output", "o", "portainer-stack.yml", "Output file for the stack template")
	generateStackCmd.Flags().StringVar(&stackName, "name", "llm-stack", "Stack name for Portainer")
	generateStackCmd.Flags().IntVar(&ollamaPort, "ollama-port", 11434, "Port for Ollama service")
	generateStackCmd.Flags().IntVar(&webuiPort, "webui-port", 3000, "Port for Open WebUI")
	generateStackCmd.Flags().StringVar(&gpuType, "gpu", "amd", "GPU type: 'amd' or 'nvidia'")
}

func runGenerateStack() error {
	logrus.Info("Generating Portainer stack template...")

	// Validate GPU type
	if gpuType != "amd" && gpuType != "nvidia" {
		return fmt.Errorf("invalid GPU type: %s. Must be 'amd' or 'nvidia'", gpuType)
	}

	config := templates.StackConfig{
		StackName:  stackName,
		OllamaPort: ollamaPort,
		WebUIPort:  webuiPort,
		GPUType:    gpuType,
	}

	template, err := templates.GeneratePortainerStack(config)
	if err != nil {
		return fmt.Errorf("failed to generate stack template: %w", err)
	}

	// Write to file
	err = os.WriteFile(outputFile, []byte(template), 0644)
	if err != nil {
		return fmt.Errorf("failed to write template to file: %w", err)
	}

	logrus.Infof("Stack template generated: %s", outputFile)
	logrus.Info("")
	logrus.Info("To deploy in Portainer:")
	logrus.Info("1. Go to Stacks > Add stack")
	logrus.Info("2. Upload the generated file or copy/paste the content")
	logrus.Info("3. Deploy the stack")
	logrus.Info("")
	logrus.Info("Services will be available at:")
	logrus.Infof("  - Ollama API: http://localhost:%d", ollamaPort)
	logrus.Infof("  - Open WebUI: http://localhost:%d", webuiPort)

	return nil
}