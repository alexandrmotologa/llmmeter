# llmmeter Deployment Guide

`llmmeter` runs as a single static binary containing the Go proxy engine and the embedded React 19 web studio.

## Standalone Binary Deployment

### Prerequisites
- Linux, macOS, or Windows
- Upstream LLM provider API keys

### 1. Build from Source
```bash
# Build React web dashboard assets
cd web && npm install && npm run build && cd ..

# Compile single binary
go build -o bin/llmmeter ./cmd/llmmeter
```

### 2. Run the Service
```bash
export OPENAI_API_KEY="sk-..."
export ANTHROPIC_API_KEY="sk-ant-..."
export PORT=8080
export LLMMETER_DB="./data/finops.db"

./bin/llmmeter serve --port $PORT --db $LLMMETER_DB
```

The web studio is accessible at `http://localhost:8080` and the proxy is active at `http://localhost:8080/v1/chat/completions`.

---

## Docker Deployment

Build the container image:
```bash
docker build -t llmmeter:latest .
```

Run with persistent storage volume:
```bash
docker run -d \
  --name llmmeter \
  -p 8080:8080 \
  -v $(pwd)/finops-data:/data \
  -e LLMMETER_DB=/data/finops.db \
  -e OPENAI_API_KEY="sk-..." \
  -e ANTHROPIC_API_KEY="sk-ant-..." \
  llmmeter:latest
```

---

## Docker Compose

```yaml
version: "3.8"

services:
  llmmeter:
    image: llmmeter:latest
    build: .
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - LLMMETER_DB=/data/finops.db
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
      - GEMINI_API_KEY=${GEMINI_API_KEY}
    volumes:
      - ./data:/data
    restart: unless-stopped
```

Start the stack:
```bash
docker-compose up -d
```

---

## Self-Hosting with Coolify

1. Create a new service in Coolify from your GitHub repository `alexandrmotologa/llmmeter`.
2. Set the build pack to `Dockerfile`.
3. Configure environment variables for `OPENAI_API_KEY`, `ANTHROPIC_API_KEY`, etc.
4. Mount persistent storage at `/data` for SQLite persistence.
5. Set port to `8080`.
