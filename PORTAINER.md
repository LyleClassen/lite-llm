# Portainer Deployment Guide

This guide walks you through deploying the LLM stack using Portainer with AMD GPU acceleration.

## Prerequisites

1. **ROCm Drivers**: AMD GPU drivers with ROCm support
2. **Docker**: Docker Engine with device access to GPU
3. **Portainer**: Portainer CE or EE installed and running

## Step 1: Generate Stack Template

```bash
# Generate the Portainer stack template
lite-llm stack generate

# Or with custom settings
lite-llm stack generate --name my-llm --ollama-port 11435 --webui-port 3001
```

This creates `portainer-stack.yml` optimized for AMD GPU acceleration.

## Step 2: Deploy in Portainer

1. **Access Portainer**: Open your Portainer web interface
2. **Navigate to Stacks**: Go to "Stacks" in the left sidebar
3. **Add Stack**: Click "Add stack"
4. **Upload Template**: 
   - Choose "Upload" method
   - Select the generated `portainer-stack.yml` file
   - Or copy/paste the file contents into the web editor
5. **Configure Stack**:
   - Name: `llm-stack` (or your preferred name)
   - Environment: Select your Docker environment
6. **Deploy**: Click "Deploy the stack"

## Step 3: Verify Deployment

```bash
# Check if services are running
lite-llm status

# Or check directly with Docker
docker ps | grep -E "(ollama|webui)"
```

Expected containers:
- `llm-stack-ollama`: Ollama server with ROCm support
- `llm-stack-webui`: Open WebUI interface

## Step 4: Download Models

```bash
# Download recommended models for RX 570/580
lite-llm models recommended

# Or download specific models
lite-llm models download llama3.1:8b-instruct-q4_K_M
```

## Step 5: Access Interfaces

- **Open WebUI**: http://localhost:3000
- **Ollama API**: http://localhost:11434
- **Custom Interface**: Run `lite-llm serve` for port 8080

## Stack Configuration Details

### Environment Variables

The generated stack includes optimal settings for AMD GPUs:

```yaml
environment:
  - HSA_OVERRIDE_GFX_VERSION=10.3.0  # RX 570/580 compatibility
  - HCC_AMDGPU_TARGET=gfx1030
  - ROCM_PATH=/opt/rocm
  - HIP_VISIBLE_DEVICES=0
```

### Device Access

```yaml
devices:
  - /dev/kfd    # ROCm kernel driver
  - /dev/dri    # Direct Rendering Infrastructure
```

### Volume Mounts

```yaml
volumes:
  - ollama_data:/root/.ollama        # Model storage
  - open_webui_data:/app/backend/data # WebUI data
```

## Managing the Stack

### Updating the Stack

1. Generate new template with updates
2. In Portainer, go to your stack
3. Click "Editor" 
4. Replace content with new template
5. Click "Update the stack"

### Scaling Services

In Portainer:
1. Go to "Containers"
2. Select your service container
3. Use "Duplicate/Edit" to scale

### Viewing Logs

1. Navigate to "Containers" in Portainer
2. Click on container name
3. Go to "Logs" tab
4. Or use CLI: `docker logs llm-stack-ollama`

### Resource Monitoring

Monitor resource usage in Portainer:
1. "Containers" view shows CPU/Memory usage
2. Click container for detailed stats
3. Or use: `lite-llm status --watch`

## Troubleshooting

### GPU Not Detected

```bash
# Check if GPU devices are accessible
ls -la /dev/dri /dev/kfd

# Verify ROCm installation
rocminfo

# Check container can access GPU
docker exec llm-stack-ollama ls -la /dev/dri
```

### Container Fails to Start

1. Check logs in Portainer
2. Verify ROCm environment variables
3. Ensure no port conflicts
4. Check available system resources

### Models Not Loading

```bash
# Check disk space
df -h

# Verify model downloads
docker exec llm-stack-ollama ollama list

# Manual model download
docker exec llm-stack-ollama ollama pull llama3.1:8b
```

### Performance Issues

1. **GPU Memory**: Monitor with `rocm-smi`
2. **System Memory**: Check available RAM
3. **Model Size**: Use smaller quantized models
4. **Concurrent Users**: Limit simultaneous requests

## Advanced Configuration

### Custom Network

If you need custom networking, modify the stack template:

```yaml
networks:
  llm-network:
    driver: bridge
    ipam:
      config:
        - subnet: 172.20.0.0/16
```

### Multiple GPU Support

For systems with multiple AMD GPUs:

```yaml
environment:
  - HIP_VISIBLE_DEVICES=0,1  # Use GPU 0 and 1
```

### Persistent Storage

The stack uses Docker volumes by default. For custom storage locations:

```yaml
volumes:
  - /host/path/ollama:/root/.ollama
  - /host/path/webui:/app/backend/data
```

## Security Considerations

1. **Network Access**: Consider using reverse proxy for external access
2. **Authentication**: Enable WebUI authentication in production
3. **Updates**: Regularly update container images
4. **Backups**: Backup model data and configurations

## Performance Optimization

### For RX 570/580 (8GB VRAM)

1. **Recommended Models**:
   - Llama 3.1 8B (Q4_K_M): ~4.4GB
   - Mistral 7B (Q4_K_M): ~4.4GB
   - Gemma 2:2B (Q4_K_M): ~1.7GB

2. **Memory Management**:
   - Leave 1-2GB VRAM free for system
   - Monitor with `rocm-smi`
   - Use quantized models

3. **Response Optimization**:
   - Adjust context length
   - Tune temperature settings
   - Limit concurrent requests

## Integration with Other Services

### Reverse Proxy

Example nginx configuration:

```nginx
server {
    listen 80;
    server_name your-domain.com;
    
    location / {
        proxy_pass http://localhost:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
    
    location /api/ {
        proxy_pass http://localhost:11434/api/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### Monitoring Stack

Consider adding monitoring services to your stack:

```yaml
services:
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    # ... configuration
    
  grafana:
    image: grafana/grafana:latest
    ports:
      - "3001:3000"
    # ... configuration
```

This guide should help you successfully deploy and manage your LLM stack using Portainer with optimal AMD GPU acceleration.