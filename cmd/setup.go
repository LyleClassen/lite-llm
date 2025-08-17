package cmd

import (
	"fmt"
	"os"

	"github.com/lyleclassen/lite-llm/internal/templates"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Generate setup scripts and templates",
	Long:  `Generate setup scripts for ROCm installation and Portainer deployment templates.`,
}

var rocmScriptCmd = &cobra.Command{
	Use:   "rocm",
	Short: "Generate ROCm installation script",
	Long:  `Generate a bash script to install and configure ROCm for AMD GPU acceleration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGenerateROCmScript()
	},
}

var dockerComposeCmd = &cobra.Command{
	Use:   "docker-compose",
	Short: "Generate docker-compose.yml for reference",
	Long:  `Generate a docker-compose.yml file for reference (use Portainer stack for actual deployment).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGenerateDockerCompose()
	},
}

var (
	setupOutputDir string
)

func init() {
	rootCmd.AddCommand(setupCmd)
	setupCmd.AddCommand(rocmScriptCmd)
	setupCmd.AddCommand(dockerComposeCmd)
	
	setupCmd.PersistentFlags().StringVarP(&setupOutputDir, "output-dir", "d", ".", "Output directory for generated files")
}

func runGenerateROCmScript() error {
	script := templates.GenerateROCmSetupScript()
	
	filename := fmt.Sprintf("%s/setup-rocm.sh", setupOutputDir)
	err := os.WriteFile(filename, []byte(script), 0755)
	if err != nil {
		return fmt.Errorf("failed to write ROCm setup script: %w", err)
	}

	logrus.Infof("ROCm setup script generated: %s", filename)
	logrus.Info("")
	logrus.Info("To install ROCm on your system:")
	logrus.Infof("  chmod +x %s", filename)
	logrus.Infof("  ./%s", filename)
	logrus.Info("")
	logrus.Info("After running the script, reboot your system and verify with:")
	logrus.Info("  rocminfo")
	logrus.Info("  /opt/rocm/bin/rocm-smi")

	return nil
}

func runGenerateDockerCompose() error {
	config := templates.StackConfig{
		StackName:  "llm-stack",
		OllamaPort: 11434,
		WebUIPort:  3000,
	}

	compose := templates.GenerateDockerComposeForReference(config)
	
	filename := fmt.Sprintf("%s/docker-compose.yml", setupOutputDir)
	err := os.WriteFile(filename, []byte(compose), 0644)
	if err != nil {
		return fmt.Errorf("failed to write docker-compose file: %w", err)
	}

	logrus.Infof("Docker Compose file generated: %s", filename)
	logrus.Info("")
	logrus.Info("This file is for reference only.")
	logrus.Info("For Portainer deployment, use: lite-llm stack generate")
	logrus.Info("")
	logrus.Info("To use with docker-compose CLI:")
	logrus.Info("  cd " + setupOutputDir)
	logrus.Info("  docker-compose up -d")

	return nil
}