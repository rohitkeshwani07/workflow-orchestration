# AI Agent Service Integration Guide

This guide explains how to integrate the new AI Agent Service with the workflow orchestration platform.

## Overview

The AI Agent Service is a separate Python microservice that provides advanced AI capabilities:
- **LangGraph**: Agent orchestration with state management
- **MCP Tools**: Bind external tools to agents via Model Context Protocol
- **Multi-Model Support**: Anthropic Claude, OpenAI, and more
- **Memory**: Persistent conversation history with Redis
- **RESTful API**: Easy integration with the Go backend

## Architecture

```
┌──────────────────────────────────────────────────────────┐
│                  Frontend (React)                         │
└────────────┬─────────────────────────────────────────────┘
             │
             ├──► Backend (Go/Gin)
             │    └──► Workflow Engine
             │         ├──► Node Executors
             │         │    └──► AI Agent Node ──┐
             │         └──► Database (PostgreSQL) │
             │                                    │
             └──────────────────────────────────┐ │
                                                │ │
                      ┌─────────────────────────┘ │
                      │                           │
                      ▼                           ▼
         ┌────────────────────────┐  ┌───────────────────────┐
         │  AI Agent Service      │  │   Redis               │
         │  (Python/FastAPI)      │──│   (State Storage)     │
         ├────────────────────────┤  └───────────────────────┘
         │ • LangGraph Agents     │
         │ • MCP Tool Manager     │
         │ • State Manager        │
         │ • LLM Providers        │
         └────────────────────────┘
                  │
                  ├──► Anthropic API
                  ├──► OpenAI API
                  └──► MCP Servers
```

## Quick Start

### 1. Start All Services

```bash
# Make sure you have your API keys in .env
docker-compose up -d

# Check all services are running
docker-compose ps
```

Expected services:
- `workflow-postgres` (port 5432)
- `workflow-backend` (port 3001)
- `workflow-ai-agent` (port 8000)
- `workflow-redis` (port 6379)
- `workflow-frontend` (port 3000)

### 2. Verify AI Agent Service

```bash
# Check health
curl http://localhost:8000/health

# Expected response:
{
  "status": "healthy",
  "service": "ai-agent-service",
  "version": "0.1.0",
  "providers": {
    "anthropic": true,
    "openai": false
  },
  "redis_connected": true
}
```

## Using AI Agents from Go Backend

### Method 1: Direct HTTP Calls

Update the Go backend's AI Agent node executor to call the AI Agent Service:

```go
// backend/engine/executor.go

func (ne *NodeExecutor) executeAIAgent(node *Node, context map[string]interface{}) (interface{}, error) {
    // Extract AI agent configuration from node
    agentConfig := node.Data["config"].(map[string]interface{})
    message := context["message"].(string)
    sessionID := context["sessionId"].(string)

    // Call AI Agent Service
    requestBody := map[string]interface{}{
        "agent_id": node.ID,
        "message": message,
        "session_id": sessionID,
        "context": context,
    }

    jsonData, _ := json.Marshal(requestBody)

    resp, err := http.Post(
        "http://ai-agent:8000/agents/" + node.ID + "/chat",
        "application/json",
        bytes.NewBuffer(jsonData),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to call AI agent: %w", err)
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)

    return result["response"], nil
}
```

### Method 2: Create Agent Client

Create a reusable client for the AI Agent Service:

```go
// backend/clients/ai_agent_client.go

package clients

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type AIAgentClient struct {
    baseURL string
    client  *http.Client
}

func NewAIAgentClient(baseURL string) *AIAgentClient {
    return &AIAgentClient{
        baseURL: baseURL,
        client:  &http.Client{},
    }
}

func (c *AIAgentClient) CreateAgent(agentID string, config map[string]interface{}) error {
    requestBody := map[string]interface{}{
        "agent_id": agentID,
        "config": config,
    }

    jsonData, _ := json.Marshal(requestBody)
    resp, err := c.client.Post(
        c.baseURL+"/agents",
        "application/json",
        bytes.NewBuffer(jsonData),
    )

    if err != nil {
        return err
    }
    defer resp.Body.Close()

    return nil
}

func (c *AIAgentClient) Chat(agentID, message, sessionID string, context map[string]interface{}) (string, error) {
    requestBody := map[string]interface{}{
        "agent_id": agentID,
        "message": message,
        "session_id": sessionID,
        "context": context,
    }

    jsonData, _ := json.Marshal(requestBody)
    resp, err := c.client.Post(
        fmt.Sprintf("%s/agents/%s/chat", c.baseURL, agentID),
        "application/json",
        bytes.NewBuffer(jsonData),
    )

    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)

    return result["response"].(string), nil
}
```

Usage:
```go
client := clients.NewAIAgentClient("http://ai-agent:8000")
response, err := client.Chat("my-agent", "Hello!", "session-123", context)
```

## Example Workflows

### Example 1: Simple Chat Agent

Create an agent and use it in a workflow:

```bash
# 1. Create an agent
curl -X POST http://localhost:8000/agents \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "customer-support",
    "config": {
      "agent_id": "customer-support",
      "provider": "anthropic",
      "model": "claude-3-5-sonnet-20241022",
      "temperature": 0.7,
      "system_prompt": "You are a helpful customer support agent.",
      "memory_enabled": true,
      "max_memory_messages": 50
    }
  }'

# 2. Chat with the agent
curl -X POST http://localhost:8000/agents/customer-support/chat \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "customer-support",
    "message": "I need help with my order",
    "session_id": "session-123"
  }'
```

### Example 2: Agent with MCP Tools

First, set up an MCP server (e.g., filesystem):

```bash
# Install MCP filesystem server
npm install -g @modelcontextprotocol/server-filesystem

# Add to .env
AI_MCP_SERVERS=filesystem:/usr/local/bin/mcp-server-filesystem:--root /app/data
```

Create agent with tools:

```bash
curl -X POST http://localhost:8000/agents \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "file-assistant",
    "config": {
      "agent_id": "file-assistant",
      "provider": "anthropic",
      "model": "claude-3-5-sonnet-20241022",
      "system_prompt": "You help users manage files.",
      "tools": [
        {
          "name": "read_file",
          "description": "Read a file",
          "server": "filesystem"
        },
        {
          "name": "write_file",
          "description": "Write to a file",
          "server": "filesystem"
        }
      ],
      "memory_enabled": true
    }
  }'
```

Chat with tool-enabled agent:

```bash
curl -X POST http://localhost:8000/agents/file-assistant/chat \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "file-assistant",
    "message": "Read the file at /app/data/config.json",
    "session_id": "session-456"
  }'
```

### Example 3: Multi-Turn Conversation

```bash
# First message
curl -X POST http://localhost:8000/agents/customer-support/chat \
  -d '{"agent_id": "customer-support", "message": "I ordered a laptop", "session_id": "order-789"}'

# Second message (agent remembers context)
curl -X POST http://localhost:8000/agents/customer-support/chat \
  -d '{"agent_id": "customer-support", "message": "When will it arrive?", "session_id": "order-789"}'

# Get conversation history
curl http://localhost:8000/agents/customer-support/sessions/order-789/history?limit=10
```

## Frontend Integration

Update the frontend to use AI agents:

```typescript
// packages/frontend/src/services/aiAgent.ts

const AI_AGENT_API = 'http://localhost:8000';

export async function createAgent(agentId: string, config: any) {
  const response = await fetch(`${AI_AGENT_API}/agents`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ agent_id: agentId, config }),
  });
  return response.json();
}

export async function chatWithAgent(
  agentId: string,
  message: string,
  sessionId?: string
) {
  const response = await fetch(`${AI_AGENT_API}/agents/${agentId}/chat`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      agent_id: agentId,
      message,
      session_id: sessionId,
    }),
  });
  return response.json();
}

export async function getAgentHistory(agentId: string, sessionId: string) {
  const response = await fetch(
    `${AI_AGENT_API}/agents/${agentId}/sessions/${sessionId}/history`
  );
  return response.json();
}
```

Use in React component:

```tsx
import { chatWithAgent } from '@/services/aiAgent';

function ChatInterface() {
  const [message, setMessage] = useState('');
  const [response, setResponse] = useState('');

  const handleSend = async () => {
    const result = await chatWithAgent(
      'customer-support',
      message,
      'session-123'
    );
    setResponse(result.response);
  };

  return (
    <div>
      <input
        value={message}
        onChange={(e) => setMessage(e.target.value)}
      />
      <button onClick={handleSend}>Send</button>
      <div>{response}</div>
    </div>
  );
}
```

## Available MCP Tools

Common MCP servers you can integrate:

### 1. Filesystem
```bash
npm install -g @modelcontextprotocol/server-filesystem
# In .env: filesystem:/usr/local/bin/mcp-server-filesystem:--root /data
```

Tools: `read_file`, `write_file`, `list_directory`, `search_files`

### 2. Git
```bash
npm install -g @modelcontextprotocol/server-git
# In .env: git:/usr/local/bin/mcp-server-git:--repo /repo
```

Tools: `git_status`, `git_diff`, `git_log`, `git_commit`

### 3. Brave Search
```bash
npm install -g @modelcontextprotocol/server-brave-search
# In .env: brave:/usr/local/bin/mcp-server-brave-search:--api-key YOUR_KEY
```

Tools: `web_search`, `local_search`

### 4. Postgres
```bash
npm install -g @modelcontextprotocol/server-postgres
# In .env: postgres:/usr/local/bin/mcp-server-postgres:--connection postgresql://user:pass@host/db
```

Tools: `query`, `list_tables`, `describe_table`

## Configuration Reference

### Agent Configuration

```json
{
  "agent_id": "unique-agent-id",
  "provider": "anthropic",  // or "openai"
  "model": "claude-3-5-sonnet-20241022",
  "temperature": 0.7,
  "max_tokens": 4096,
  "system_prompt": "You are a helpful assistant.",
  "tools": [
    {
      "name": "tool_name",
      "description": "Tool description",
      "server": "mcp-server-name"
    }
  ],
  "memory_enabled": true,
  "max_memory_messages": 50
}
```

### Supported Models

**Anthropic:**
- `claude-3-5-sonnet-20241022`
- `claude-3-opus-20240229`
- `claude-3-sonnet-20240229`
- `claude-3-haiku-20240307`

**OpenAI:**
- `gpt-4-turbo-preview`
- `gpt-4`
- `gpt-3.5-turbo`

## Monitoring and Debugging

### Check Service Logs

```bash
# AI Agent Service logs
docker-compose logs -f ai-agent

# Backend logs
docker-compose logs -f backend

# Redis logs
docker-compose logs -f redis
```

### Monitor Agent Activity

```bash
# List all agents
curl http://localhost:8000/agents

# Get agent status
curl http://localhost:8000/agents/{agent_id}

# Check available tools
curl http://localhost:8000/tools

# Check MCP servers
curl http://localhost:8000/mcp/servers
```

### Debug Redis State

```bash
# Connect to Redis
docker exec -it workflow-redis redis-cli

# View all agent keys
KEYS agent:*

# Get agent config
GET agent:my-agent:config

# Get session messages
GET agent:my-agent:session:session-123
```

## Performance Considerations

1. **Redis vs In-Memory**: Use Redis for production (persistent, scalable)
2. **Connection Pooling**: The service uses connection pooling for better performance
3. **Message Limits**: Configure `max_memory_messages` to control memory usage
4. **Tool Timeouts**: MCP tools have default 30s timeout
5. **Concurrent Requests**: Service handles concurrent agent invocations

## Troubleshooting

### Agent Service Won't Start

```bash
# Check logs
docker-compose logs ai-agent

# Common issues:
# - Missing API keys in .env
# - Redis not available
# - Port 8000 already in use
```

### Agent Creation Fails

```bash
# Verify provider API key
curl http://localhost:8000/health

# Check providers object in response
```

### MCP Tools Not Working

```bash
# List MCP servers
curl http://localhost:8000/mcp/servers

# Reconnect server
curl -X POST http://localhost:8000/mcp/servers/{server_name}/connect \
  -d '{"command": "/path/to/server", "args": ["--arg"]}'
```

### Memory Not Persisting

```bash
# Check Redis connection
docker-compose ps redis

# Verify USE_REDIS=true in docker-compose.yml
```

## Next Steps

1. **Extend Node Types**: Add AI Agent node type to frontend workflow editor
2. **Streaming Responses**: Implement SSE for real-time agent responses
3. **Agent Templates**: Create pre-configured agents for common use cases
4. **Metrics**: Add Prometheus metrics for agent performance
5. **Rate Limiting**: Implement rate limits per agent/session

## Resources

- [LangGraph Documentation](https://python.langchain.com/docs/langgraph)
- [MCP Specification](https://modelcontextprotocol.io/)
- [Anthropic API Docs](https://docs.anthropic.com/)
- [OpenAI API Docs](https://platform.openai.com/docs/)
