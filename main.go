package main

import (
	"os"

	"github.com/lyleclassen/lite-llm/cmd"
	"github.com/sirupsen/logrus"
)

func main() {
	// Set up logging
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logrus.SetLevel(logrus.InfoLevel)

	// Execute root command
	if err := cmd.Execute(); err != nil {
		logrus.Fatal(err)
		os.Exit(1)
	}
}