# Workflow Orchestration Platform

A powerful workflow orchestration system similar to n8n, featuring AI agent integration and real-time chat triggers. Built with Go backend and React frontend.

## Features

- **Node-Based Workflow Designer**: Drag-and-drop interface for building workflows
- **AI Agent Integration**: Built-in support for Claude AI with configurable prompts
- **Chat Trigger**: Real-time WebSocket-based chat interface to trigger workflows
- **Multiple Node Types**:
  - Chat Trigger: Start workflows from chat messages
  - AI Agent: Integrate Claude AI for intelligent responses
  - HTTP Request: Make API calls
  - Transform: Transform data with JavaScript
  - Condition: Branch workflow based on conditions
  - Code: Execute custom JavaScript code
  - Delay: Add delays between steps
  - Response: Send responses back to chat
- **Workflow Execution Engine**: Reliable execution with state management
- **Execution Monitoring**: Track workflow runs and view logs
- **Flexible Database**: PostgreSQL or SQLite with GORM (easily switchable)

## Architecture

This is a monorepo with:

- `backend/`: Go API server with workflow engine
- `packages/frontend/`: React/TypeScript UI with React Flow
- `packages/shared/`: Shared TypeScript types and schemas

## Prerequisites

**For Docker (Recommended):**
- Docker Engine 20.10+
- Docker Compose 2.0+

**For Local Development:**
- Go >= 1.21
- Node.js >= 18.0.0
- npm or yarn

## Quick Start

### Option 1: Docker (Recommended)

1. **Configure environment**:
```bash
cp .env.example .env
# Edit .env and add your ANTHROPIC_API_KEY
```

2. **Start with Docker Compose**:
```bash
docker-compose up -d
```

3. **Access the application**:
- Frontend: http://localhost:3000
- Backend: http://localhost:3001

See [DOCKER.md](DOCKER.md) for detailed Docker documentation.

### Option 2: Local Development

1. **Install Go dependencies**:
```bash
cd backend
go mod download
cd ..
```

2. **Install frontend dependencies**:
```bash
npm install
```

3. **Configure environment variables**:
```bash
cd backend
cp .env.example .env
# Edit .env and add your ANTHROPIC_API_KEY
cd ..
```

4. **Start development servers**:
```bash
# From root directory
npm run dev
```

This will start:
- Go Backend API on http://localhost:3001
- React Frontend UI on http://localhost:3000

5. **Open your browser**:
Navigate to http://localhost:3000

## Building for Production

### Docker (Recommended)
```bash
docker-compose up --build -d
```

### Manual Build
```bash
# Build everything
npm run build

# Run backend
cd backend
./bin/workflow-server

# Serve frontend (use your preferred static file server)
cd ../packages/frontend/dist
```

## Creating Your First Workflow

1. Click "Create Workflow" on the workflows page
2. Add nodes from the palette on the left:
   - Start with a "Chat Trigger" node
   - Add an "AI Agent" node
   - Connect them by dragging from the bottom handle to the top handle
   - Add a "Response" node to send the result back
3. Click on each node to configure it:
   - AI Agent: Set system prompt, model, temperature
   - Response: Set response message template
4. Click "Save" to save your workflow
5. Click "Test Chat" to open the chat interface
6. Send a message to trigger your workflow!

## Example Workflows

### Simple AI Chatbot
```
Chat Trigger → AI Agent → Response
```

1. **Chat Trigger**: Receives user message
2. **AI Agent**:
   - Model: Claude 3.5 Sonnet
   - System Prompt: "You are a helpful assistant. Answer questions clearly and concisely."
3. **Response**:
   - Message: `{{AI Agent node ID.text}}`

### API Integration with AI
```
Chat Trigger → HTTP Request → AI Agent → Response
```

1. **Chat Trigger**: Receives user query
2. **HTTP Request**: Fetch data from API
3. **AI Agent**: Process API data and generate response
4. **Response**: Send result to user

### Conditional Workflow
```
Chat Trigger → Condition → [AI Agent | HTTP Request] → Response
```

Use conditions to branch workflows based on input or data.

## Node Configuration

### AI Agent
- **Model**: Choose Claude model (Sonnet, Opus, Haiku)
- **System Prompt**: Define AI behavior and context
- **Temperature**: Control randomness (0-1)
- **Max Tokens**: Limit response length
- **Variable Interpolation**: Use `{{variableName}}` to reference context

### HTTP Request
- **Method**: GET, POST, PUT, DELETE, PATCH
- **URL**: API endpoint (supports variable interpolation)
- **Headers**: JSON object of HTTP headers
- **Body**: Request body (for POST/PUT/PATCH)

### Transform
- **Expression**: JavaScript expression that returns a value
- Access context variables directly
- Example: `{ result: trigger.message.toUpperCase() }`

### Condition
- **Expression**: JavaScript boolean expression
- Example: `trigger.message.includes('hello')`
- Branches workflow based on true/false result

### Code
- **Code**: Full JavaScript code block
- Access all context variables
- Must return a value
- Example:
```javascript
const processed = trigger.message.split(' ');
return { words: processed, count: processed.length };
```

## API Endpoints

### Workflows
- `GET /api/workflows` - List all workflows
- `GET /api/workflows/:id` - Get workflow by ID
- `POST /api/workflows` - Create workflow
- `PUT /api/workflows/:id` - Update workflow
- `DELETE /api/workflows/:id` - Delete workflow
- `POST /api/workflows/:id/execute` - Execute workflow
- `GET /api/workflows/:id/executions` - Get execution history

### WebSocket
- `ws://localhost:3001/ws/chat` - Chat WebSocket endpoint

## WebSocket Protocol

### Start Session
```json
{
  "type": "start_session",
  "workflowId": "workflow-id"
}
```

### Send Message
```json
{
  "type": "chat_message",
  "sessionId": "session-id",
  "content": "Hello!"
}
```

### Receive Response
```json
{
  "type": "chat_response",
  "message": {
    "id": "msg-id",
    "role": "assistant",
    "content": "Response text",
    "timestamp": "2024-01-01T00:00:00.000Z"
  }
}
```

## Technology Stack

### Backend
- **Go 1.21+**
- Gin (HTTP framework)
- Gorilla WebSocket
- GORM (ORM with PostgreSQL/SQLite support)
- PostgreSQL 16 or SQLite3
- Anthropic API (AI integration)

### Frontend
- React 18
- TypeScript
- Vite
- React Flow (Node-based UI)
- TanStack Query (Data fetching)
- Tailwind CSS (Styling)
- Zustand (State management)

## Database

The application uses **GORM** as the ORM and supports multiple databases:
- **PostgreSQL** (default, recommended for production)
- **SQLite** (alternative, good for development)

Database tables (auto-migrated by GORM):
- `workflows`: Workflow definitions
- `executions`: Workflow execution records
- `execution_logs`: Node execution logs
- `chat_sessions`: Chat sessions
- `chat_messages`: Chat message history

See [DATABASE.md](DATABASE.md) for complete database configuration guide.

## Development

### Project Structure
```
workflow-orchestration/
├── backend/
│   ├── database/           # SQLite database layer
│   ├── engine/             # Workflow execution engine
│   ├── handlers/           # HTTP API handlers
│   ├── models/             # Data models
│   ├── websocket/          # WebSocket server
│   ├── main.go             # Entry point
│   ├── go.mod              # Go dependencies
│   └── Makefile            # Build commands
├── packages/
│   ├── frontend/
│   │   ├── src/
│   │   │   ├── components/     # React components
│   │   │   ├── pages/          # Page components
│   │   │   ├── api/            # API client
│   │   │   └── stores/         # State management
│   │   └── package.json
│   └── shared/
│       └── src/
│           └── types.ts         # Shared TypeScript types
└── package.json
```

### Adding New Node Types

1. Add node type to `backend/models/models.go`:
```go
const (
  // ... existing types
  NodeTypeMyNewNode NodeType = "my_new_node"
)
```

2. Implement execution logic in `backend/engine/executor.go`:
```go
func (ne *NodeExecutor) executeMyNewNode(node *models.Node, ctx map[string]interface{}) (interface{}, error) {
  // Your implementation
}
```

3. Add UI component in frontend node palette and config panel

## Environment Variables

### Backend (.env in backend/ directory)
- `PORT`: Server port (default: 3001)
- `ANTHROPIC_API_KEY`: Your Anthropic API key (required for AI nodes)
- `DATABASE_PATH`: SQLite database path (default: ./data/workflows.db)

## Running Backend Separately

```bash
cd backend
make deps      # Download Go dependencies
make dev       # Run in development mode
make build     # Build binary
make run       # Run built binary
```

## Deployment

### Docker Deployment (Recommended)
See [DOCKER.md](DOCKER.md) for complete Docker deployment guide including:
- Production configuration
- Data persistence and backups
- Scaling and load balancing
- Security best practices
- Monitoring and logging

### Manual Deployment
1. Build the Go binary: `cd backend && make build`
2. Build the frontend: `cd packages/frontend && npm run build`
3. Deploy binary and static files to your server
4. Set up reverse proxy (Nginx/Caddy) for HTTPS
5. Configure systemd service for backend process

## Troubleshooting

### Docker Issues
See [DOCKER.md](DOCKER.md) troubleshooting section for Docker-specific issues.

### WebSocket Connection Issues
- Ensure backend is running on port 3001
- Check browser console for connection errors
- Verify proxy configuration in vite.config.ts

### Workflow Execution Errors
- Check execution logs in the database
- Verify node configurations are valid
- Ensure AI API key is configured correctly

### Database Issues
- Database is created automatically in `backend/data/`
- Delete database file to reset all data
- Check file permissions for database directory

### Go Build Issues
- Ensure Go 1.21+ is installed: `go version`
- Run `go mod download` in backend directory
- On Linux, you may need to install gcc for SQLite: `apt-get install build-essential`

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## License

MIT License - See LICENSE file for details

## Support

For issues and questions, please open an issue on the GitHub repository.
