package templates

import (
	"fmt"
	"strings"
)

type StackConfig struct {
	StackName  string
	OllamaPort int
	WebUIPort  int
	GPUType    string // "amd" or "nvidia"
}

func GeneratePortainerStack(config StackConfig) (string, error) {
	var ollamaService string
	
	if config.GPUType == "nvidia" {
		ollamaService = `  ollama:
    image: ollama/ollama:latest
    container_name: %s-ollama
    restart: unless-stopped
    ports:
      - "%d:11434"
    volumes:
      - ollama_data:/root/.ollama
    environment:
      - NVIDIA_VISIBLE_DEVICES=all
      - NVIDIA_DRIVER_CAPABILITIES=compute,utility
    deploy:
      resources:
        reservations:
          devices:
            - driver: nvidia
              count: 1
              capabilities: [gpu]
    labels:
      - "io.portainer.accesscontrol.teams=administrators"
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:11434/api/version"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    networks:
      - llm-network`
	} else {
		// AMD ROCm configuration
		ollamaService = `  ollama:
    image: ollama/ollama:rocm
    container_name: %s-ollama
    restart: unless-stopped
    ports:
      - "%d:11434"
    volumes:
      - ollama_data:/root/.ollama
    devices:
      - /dev/kfd
      - /dev/dri
    environment:
      - HSA_OVERRIDE_GFX_VERSION=10.3.0  # For RX 570/580 compatibility
      - HCC_AMDGPU_TARGET=gfx1030
      - ROCM_PATH=/opt/rocm
      - HIP_VISIBLE_DEVICES=0
    labels:
      - "io.portainer.accesscontrol.teams=administrators"
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:11434/api/version"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    networks:
      - llm-network`
	}

	template := `version: '3.8'

services:
%s

  open-webui:
    image: ghcr.io/open-webui/open-webui:main
    container_name: %s-webui
    restart: unless-stopped
    ports:
      - "%d:8080"
    volumes:
      - open_webui_data:/app/backend/data
    environment:
      - OLLAMA_BASE_URL=http://ollama:11434
      - WEBUI_SECRET_KEY=%s-secret-key
      - WEBUI_AUTH=false  # Set to true if you want authentication
    labels:
      - "io.portainer.accesscontrol.teams=administrators"
    depends_on:
      ollama:
        condition: service_healthy
    networks:
      - llm-network

volumes:
  ollama_data:
    driver: local
    labels:
      - "io.portainer.accesscontrol.teams=administrators"
  open_webui_data:
    driver: local
    labels:
      - "io.portainer.accesscontrol.teams=administrators"

networks:
  llm-network:
    driver: bridge
    labels:
      - "io.portainer.accesscontrol.teams=administrators"

# Portainer Stack Configuration
# 
# This stack is optimized for AMD GPU acceleration with ROCm.
# 
# Prerequisites:
# 1. ROCm drivers installed on the host
# 2. Docker with device access to /dev/kfd and /dev/dri
# 3. At least 8GB GPU memory and 16GB system RAM
#
# Recommended Models for RX 570/580:
# - llama3.1:8b-instruct-q4_K_M (~4.4GB)
# - mistral:7b-instruct-q4_K_M (~4.4GB)
# - gemma2:2b-instruct-q4_K_M (~1.7GB)
#
# After deployment, download models using:
# docker exec %s-ollama ollama pull llama3.1:8b-instruct-q4_K_M
`

	// Format the ollama service first
	formattedOllamaService := fmt.Sprintf(ollamaService, config.StackName, config.OllamaPort)
	
	result := fmt.Sprintf(template, 
		formattedOllamaService, // formatted ollama service
		config.StackName,       // webui container name  
		config.WebUIPort,       // webui port
		config.StackName,       // secret key
		config.StackName,       // model download example
	)

	return result, nil
}

func GenerateROCmSetupScript() string {
	return `#!/bin/bash
# ROCm Setup Script for AMD GPU LLM Deployment
# Run this script on your Ubuntu 24.04 system before deploying the stack

set -e

echo "=== ROCm Setup for AMD GPU LLM Deployment ==="

# Check if running as root
if [[ $EUID -eq 0 ]]; then
   echo "Please don't run this script as root"
   exit 1
fi

# Update system
echo "Updating system packages..."
sudo apt update && sudo apt upgrade -y

# Install prerequisites
echo "Installing prerequisites..."
sudo apt install -y wget curl gnupg2 software-properties-common

# Add ROCm repository
echo "Adding ROCm repository..."
wget -q -O - https://repo.radeon.com/rocm/rocm.gpg.key | sudo apt-key add -
echo 'deb [arch=amd64] https://repo.radeon.com/rocm/apt/5.7/ ubuntu main' | sudo tee /etc/apt/sources.list.d/rocm.list

# Update package list
sudo apt update

# Install ROCm
echo "Installing ROCm..."
sudo apt install -y rocm-dev rocm-utils

# Add user to render and video groups
echo "Adding user to render and video groups..."
sudo usermod -aG render,video $USER

# Set environment variables
echo "Setting up environment variables..."
echo 'export PATH=$PATH:/opt/rocm/bin' >> ~/.bashrc
echo 'export HSA_OVERRIDE_GFX_VERSION=10.3.0' >> ~/.bashrc

# Create udev rules for device access
echo "Setting up device permissions..."
sudo tee /etc/udev/rules.d/70-rocm.rules > /dev/null <<EOF
SUBSYSTEM=="kfd", KERNEL=="kfd", TAG+="uaccess", GROUP="render"
SUBSYSTEM=="drm", KERNEL=="renderD*", GROUP="render", MODE="0664"
EOF

sudo udevadm control --reload-rules
sudo udevadm trigger

echo ""
echo "=== ROCm Setup Complete ==="
echo ""
echo "IMPORTANT: You need to reboot your system for changes to take effect."
echo "After rebooting, verify the installation by running:"
echo "  rocminfo"
echo "  /opt/rocm/bin/rocm-smi"
echo ""
echo "Then you can deploy the Portainer stack."
`
}

func GenerateDockerComposeForReference(config StackConfig) string {
	// This generates a standalone docker-compose.yml for reference
	// (not for Portainer, but for users who prefer docker-compose CLI)
	
	template := `# Docker Compose reference for %s
# This file is for reference only - use the Portainer stack template for deployment

version: '3.8'

services:
  ollama:
    image: ollama/ollama:rocm
    container_name: %s-ollama
    restart: unless-stopped
    ports:
      - "%d:11434"
    volumes:
      - ./data/ollama:/root/.ollama
    devices:
      - /dev/kfd
      - /dev/dri
    environment:
      - HSA_OVERRIDE_GFX_VERSION=10.3.0
      - HCC_AMDGPU_TARGET=gfx1030
      - ROCM_PATH=/opt/rocm
      - HIP_VISIBLE_DEVICES=0

  open-webui:
    image: ghcr.io/open-webui/open-webui:main
    container_name: %s-webui
    restart: unless-stopped
    ports:
      - "%d:8080"
    volumes:
      - ./data/webui:/app/backend/data
    environment:
      - OLLAMA_BASE_URL=http://ollama:11434
      - WEBUI_SECRET_KEY=%s-secret-key
      - WEBUI_AUTH=false
    depends_on:
      - ollama

# To use this file:
# 1. Save as docker-compose.yml
# 2. Run: docker-compose up -d
# 3. Download models: docker exec %s-ollama ollama pull llama3.1:8b
`

	return fmt.Sprintf(template,
		config.StackName,
		config.StackName,
		config.OllamaPort,
		config.StackName,
		config.WebUIPort,
		config.StackName,
		config.StackName,
	)
}