# Lite LLM

A comprehensive LLM management tool designed for AMD GPU homelab systems. Supports deployment, monitoring, and management of local language models using Docker and ROCm acceleration.

## Features

- 🚀 **Easy Deployment**: One-command deployment of Ollama + Open WebUI with AMD GPU support
- 🔧 **System Checking**: Automatic validation of hardware and software requirements
- 📊 **Model Management**: Download, list, and manage LLM models optimized for your hardware
- 🌐 **Web Interface**: Custom web UI for interacting with your models
- 📈 **Monitoring**: Real-time system and service status monitoring
- 🎯 **AMD GPU Optimized**: Specifically designed for AMD RX 570/580 and similar GPUs

## Hardware Requirements

- **GPU**: AMD RX 570/580 (8GB VRAM) or similar
- **RAM**: 16GB system memory recommended
- **OS**: Ubuntu 24.04 LTS (or compatible Linux distribution)
- **Software**: Docker, ROCm drivers

## Quick Start

1. **Install Dependencies**:
   ```bash
   # Generate and run ROCm setup script
   lite-llm setup rocm
   chmod +x setup-rocm.sh
   ./setup-rocm.sh
   sudo reboot  # Required after ROCm installation
   ```

2. **Build and Install**:
   ```bash
   git clone <repository-url>
   cd lite-llm
   go build -o lite-llm
   sudo mv lite-llm /usr/local/bin/
   ```

3. **Deploy with Portainer**:
   ```bash
   # Generate Portainer stack template
   lite-llm stack generate
   
   # Deploy in Portainer:
   # 1. Go to Stacks > Add stack
   # 2. Upload portainer-stack.yml or copy/paste content
   # 3. Deploy the stack
   
   # Check deployment status
   lite-llm status
   ```

4. **Download Models**:
   ```bash
   # Download recommended models for your hardware
   lite-llm models recommended
   ```

5. **Access Web Interfaces**:
   - Open WebUI: http://localhost:3000
   - Custom Interface: http://localhost:8080 (run `lite-llm serve`)

## Commands

### Portainer Stack Management
```bash
lite-llm stack generate                    # Generate Portainer stack template
lite-llm stack generate -o my-stack.yml    # Custom output file
lite-llm stack generate --ollama-port 11435 --webui-port 3001  # Custom ports
```

### Setup and Configuration
```bash
lite-llm setup rocm                # Generate ROCm installation script
lite-llm setup docker-compose      # Generate docker-compose.yml for reference
```

### Model Management
```bash
lite-llm models list                    # List installed models
lite-llm models download llama3.1:8b    # Download specific model
lite-llm models remove llama3.1:8b      # Remove model
lite-llm models recommended             # Download recommended models
```

### Monitoring
```bash
lite-llm status           # Check system status
lite-llm status --watch   # Continuous monitoring
```

### Web Interface
```bash
lite-llm serve            # Start custom web interface
lite-llm serve --port 8080 --host 0.0.0.0
```

## Recommended Models for RX 570/580

The following models are optimized for 8GB VRAM GPUs:

- **Llama 3.1 8B** (Q4_K_M): ~4.4GB, excellent general performance
- **Mistral 7B** (Q4_K_M): ~4.4GB, fast and efficient for most tasks  
- **Gemma 2:2B** (Q4_K_M): ~1.7GB, lightweight option for simple tasks

## Configuration

Create `~/.lite-llm.yaml` for custom configuration:

```yaml
ollama:
  port: 11434
webui:
  port: 3000
gpu:
  type: amd
  override_version: "10.3.0"
models:
  default:
    - "llama3.1:8b"
    - "mistral:7b"
```

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   lite-llm CLI  │    │    Portainer    │    │ Custom Web UI   │
│ (Model Mgmt +   │    │ (Container Mgmt) │    │  (Port 8080)    │
│  Monitoring)    │    │                 │    │                 │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          │              ┌───────┴───────┐              │
          │              │ Docker Engine │              │
          │              └───────┬───────┘              │
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌─────────────┴───────────┐
                    │     Ollama Server       │
                    │     (Port 11434)        │
                    │   ROCm + AMD GPU        │
                    │                         │
                    │    Open WebUI          │
                    │    (Port 3000)         │
                    └─────────────────────────┘
```

## Performance Expectations

With AMD RX 570/580 (8GB VRAM):
- **Response Speed**: 15-30 tokens/second for 7B models
- **Model Loading**: 10-30 seconds depending on model size
- **Concurrent Users**: 2-3 simultaneous conversations
- **Memory Usage**: ~6-7GB GPU memory for models + overhead

## Troubleshooting

### GPU Not Detected
```bash
# Check if AMD GPU is recognized
lspci | grep -i amd

# Check ROCm installation
rocminfo

# Verify Docker can access GPU devices
ls -la /dev/dri /dev/kfd
```

### Ollama Not Starting
```bash
# Check container logs in Portainer or via Docker CLI
docker logs <stack-name>-ollama

# Verify ROCm environment
echo $HSA_OVERRIDE_GFX_VERSION

# Check if containers are running
docker ps | grep ollama
```

### Models Not Loading
```bash
# Check available disk space
df -h

# Check model download status
lite-llm models list

# Manual model download (replace <stack-name> with your stack name)
docker exec <stack-name>-ollama ollama pull llama3.1:8b
```

## Development

```bash
# Run tests
go test ./...

# Build for development
go build -o lite-llm

# Run with verbose logging
./lite-llm --verbose deploy
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

MIT License - see LICENSE file for details
