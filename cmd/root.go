package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	verbose bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "lite-llm",
	Short: "LLM management tool for AMD GPU homelab deployments",
	Long: `A comprehensive LLM management tool designed for AMD GPU homelab systems.
Supports deployment, monitoring, and management of local language models
using Docker and ROCm acceleration.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.lite-llm.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Bind flags to viper
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName(".lite-llm")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME")
	}

	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("ollama.port", 11434)
	viper.SetDefault("webui.port", 3000)
	viper.SetDefault("gpu.type", "amd")
	viper.SetDefault("gpu.override_version", "10.3.0")
	viper.SetDefault("models.default", []string{"llama3.1:8b", "mistral:7b"})

	if err := viper.ReadInConfig(); err == nil {
		// Config file found and successfully parsed
	}
}