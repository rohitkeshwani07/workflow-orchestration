# AI Agent Service

A powerful AI agent service built with LangGraph, supporting MCP (Model Context Protocol) tools, memory management, and multiple LLM providers.

## Features

- **LangGraph Integration**: Advanced agent orchestration with state management
- **MCP Tool Support**: Bind and use Model Context Protocol tools
- **Multi-Provider Support**: Works with Anthropic Claude, OpenAI, and more
- **Memory Management**: Conversation history with Redis or in-memory storage
- **RESTful API**: Easy integration with other services
- **Docker Ready**: Containerized deployment

## Quick Start

### Using Docker Compose (Recommended)

From the project root:
```bash
docker-compose up ai-agent redis
```

### Local Development

1. **Install dependencies**:
```bash
pip install -r requirements.txt
```

2. **Configure environment**:
```bash
cp .env.example .env
# Edit .env with your API keys
```

3. **Run the service**:
```bash
uvicorn main:app --reload --port 8000
```

## API Endpoints

### Health Check
```bash
GET /health
```

### Agent Management

**Create Agent**:
```bash
POST /agents
Content-Type: application/json

{
  "agent_id": "my-agent",
  "config": {
    "agent_id": "my-agent",
    "provider": "anthropic",
    "model": "claude-3-5-sonnet-20241022",
    "temperature": 0.7,
    "max_tokens": 4096,
    "system_prompt": "You are a helpful assistant.",
    "tools": [],
    "memory_enabled": true,
    "max_memory_messages": 50
  }
}
```

**Chat with Agent**:
```bash
POST /agents/{agent_id}/chat
Content-Type: application/json

{
  "agent_id": "my-agent",
  "message": "Hello, how are you?",
  "session_id": "session-123",
  "context": {}
}
```

**List Agents**:
```bash
GET /agents
```

**Get Agent Status**:
```bash
GET /agents/{agent_id}
```

**Delete Agent**:
```bash
DELETE /agents/{agent_id}
```

### Session Management

**Get Session History**:
```bash
GET /agents/{agent_id}/sessions/{session_id}/history?limit=50
```

**Clear Session**:
```bash
POST /agents/{agent_id}/sessions/{session_id}/clear
```

### MCP Tools

**List Available Tools**:
```bash
GET /tools
```

**List MCP Servers**:
```bash
GET /mcp/servers
```

**Connect MCP Server**:
```bash
POST /mcp/servers/{server_name}/connect
Content-Type: application/json

{
  "command": "/usr/local/bin/mcp-server-filesystem",
  "args": ["--root", "/data"]
}
```

**Disconnect MCP Server**:
```bash
DELETE /mcp/servers/{server_name}
```

## Supported LLM Providers

### Anthropic Claude
```json
{
  "provider": "anthropic",
  "model": "claude-3-5-sonnet-20241022"
}
```

**Available models:**
- `claude-3-5-sonnet-20241022`
- `claude-3-opus-20240229`
- `claude-3-sonnet-20240229`
- `claude-3-haiku-20240307`

### OpenAI
```json
{
  "provider": "openai",
  "model": "gpt-4-turbo-preview"
}
```

**Available models:**
- `gpt-4-turbo-preview`
- `gpt-4`
- `gpt-3.5-turbo`

## MCP Tool Integration

MCP (Model Context Protocol) tools extend agent capabilities.

### Example MCP Servers

**1. Filesystem**
```bash
npm install -g @modelcontextprotocol/server-filesystem
```

Configure in `.env`:
```
MCP_SERVERS=filesystem:/usr/local/bin/mcp-server-filesystem:--root /data
```

**2. Git**
```bash
npm install -g @modelcontextprotocol/server-git
```

**3. Brave Search**
```bash
npm install -g @modelcontextprotocol/server-brave-search
```

**4. Postgres**
```bash
npm install -g @modelcontextprotocol/server-postgres
```

### Using Tools in Agents

```json
{
  "agent_id": "file-agent",
  "config": {
    "tools": [
      {
        "name": "read_file",
        "description": "Read a file from the filesystem",
        "server": "filesystem"
      },
      {
        "name": "write_file",
        "description": "Write content to a file",
        "server": "filesystem"
      }
    ]
  }
}
```

## Memory and State Management

### Redis (Production)
```bash
USE_REDIS=true
REDIS_HOST=redis
REDIS_PORT=6379
```

Benefits:
- Persistent storage
- Scalable across instances
- Automatic TTL for sessions

### In-Memory (Development)
```bash
USE_REDIS=false
```

Benefits:
- No external dependencies
- Fast for testing
- Simple setup

## Configuration

All configuration via environment variables. See `.env.example`.

Key settings:
- `ANTHROPIC_API_KEY`: Anthropic API key
- `OPENAI_API_KEY`: OpenAI API key
- `DEFAULT_PROVIDER`: Default LLM provider (anthropic, openai)
- `DEFAULT_MODEL`: Default model name
- `REDIS_HOST`: Redis hostname
- `USE_REDIS`: Enable Redis (true/false)
- `MCP_SERVERS`: MCP servers (comma-separated)

## Architecture

```
┌─────────────────────────────────────┐
│      FastAPI Application            │
│  (REST API Endpoints)               │
└──────────┬──────────────────────────┘
           │
           ├───► Agent Manager
           │     └───► LangGraph Agents
           │           ├─► LLM (Anthropic/OpenAI)
           │           └─► Tools (MCP)
           │
           ├───► State Manager
           │     └───► Redis / In-Memory
           │
           └───► MCP Manager
                 └───► MCP Tool Servers
```

## Example Usage

### Python Client
```python
import httpx

async with httpx.AsyncClient() as client:
    # Create agent
    response = await client.post(
        "http://localhost:8000/agents",
        json={
            "agent_id": "assistant",
            "config": {
                "agent_id": "assistant",
                "provider": "anthropic",
                "model": "claude-3-5-sonnet-20241022",
                "memory_enabled": True
            }
        }
    )

    # Chat
    response = await client.post(
        "http://localhost:8000/agents/assistant/chat",
        json={
            "agent_id": "assistant",
            "message": "Hello!",
            "session_id": "session-1"
        }
    )

    print(response.json()["response"])
```

### cURL
```bash
# Create agent
curl -X POST http://localhost:8000/agents \
  -H "Content-Type: application/json" \
  -d '{"agent_id":"assistant","config":{"agent_id":"assistant","provider":"anthropic","model":"claude-3-5-sonnet-20241022"}}'

# Chat
curl -X POST http://localhost:8000/agents/assistant/chat \
  -H "Content-Type: application/json" \
  -d '{"agent_id":"assistant","message":"Hello!","session_id":"session-1"}'
```

## Integration with Workflow Orchestration

This service integrates with the Go backend:

```go
// Call from Go
resp, err := http.Post(
    "http://ai-agent:8000/agents/my-agent/chat",
    "application/json",
    payload,
)
```

See `AI_AGENT_INTEGRATION.md` in the project root for complete integration guide.

## Development

### Running Tests
```bash
pytest
```

### Code Formatting
```bash
black .
ruff check .
```

### Docker Build
```bash
docker build -t ai-agent-service .
docker run -p 8000:8000 --env-file .env ai-agent-service
```

## Troubleshooting

### Service Won't Start
- Check API keys in `.env`
- Verify Redis connection
- Check port 8000 availability

### Agent Creation Fails
- Verify provider API key configured
- Check model name is correct
- Review logs: `docker-compose logs ai-agent`

### MCP Tools Not Working
- Verify MCP server installed
- Check server command path
- Test connection: `GET /mcp/servers`

### Memory Not Persisting
- Verify Redis running: `docker-compose ps redis`
- Check `USE_REDIS=true` in config
- Test Redis: `docker exec -it workflow-redis redis-cli ping`

## License

MIT
