# Workflow Orchestration Platform

A powerful workflow orchestration system similar to n8n, featuring AI agent integration and real-time chat triggers.

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
- **SQLite Database**: Persistent storage for workflows and execution history

## Architecture

This is a monorepo with three packages:

- `packages/backend`: Node.js/Express API server with workflow engine
- `packages/frontend`: React/TypeScript UI with React Flow
- `packages/shared`: Shared TypeScript types and schemas

## Prerequisites

- Node.js >= 18.0.0
- npm or yarn

## Quick Start

1. **Install dependencies**:
```bash
npm install
```

2. **Configure environment variables**:
```bash
cd packages/backend
cp .env.example .env
# Edit .env and add your ANTHROPIC_API_KEY
```

3. **Start development servers**:
```bash
# From root directory
npm run dev
```

This will start:
- Backend API on http://localhost:3001
- Frontend UI on http://localhost:3000

4. **Open your browser**:
Navigate to http://localhost:3000

## Building for Production

```bash
npm run build
npm start
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
- Node.js + Express
- TypeScript
- better-sqlite3 (Database)
- ws (WebSocket)
- Anthropic SDK (AI)

### Frontend
- React 18
- TypeScript
- Vite
- React Flow (Node-based UI)
- TanStack Query (Data fetching)
- Tailwind CSS (Styling)
- Zustand (State management)

## Database Schema

The SQLite database includes tables for:
- `workflows`: Workflow definitions
- `executions`: Workflow execution records
- `execution_logs`: Node execution logs
- `chat_sessions`: Chat sessions
- `chat_messages`: Chat message history

## Development

### Project Structure
```
workflow-orchestration/
├── packages/
│   ├── backend/
│   │   ├── src/
│   │   │   ├── engine/         # Workflow execution engine
│   │   │   ├── routes/         # API routes
│   │   │   ├── database.ts     # Database layer
│   │   │   ├── websocket.ts    # WebSocket server
│   │   │   └── index.ts        # Entry point
│   │   └── package.json
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

1. Add node type to `packages/shared/src/types.ts`:
```typescript
export enum NodeType {
  // ... existing types
  MY_NEW_NODE = 'my_new_node'
}
```

2. Implement execution logic in `packages/backend/src/engine/NodeExecutor.ts`

3. Add UI component in frontend node palette and config panel

## Environment Variables

### Backend
- `PORT`: Server port (default: 3001)
- `ANTHROPIC_API_KEY`: Your Anthropic API key
- `DATABASE_PATH`: SQLite database path (default: ./data/workflows.db)

## Troubleshooting

### WebSocket Connection Issues
- Ensure backend is running on port 3001
- Check browser console for connection errors
- Verify proxy configuration in vite.config.ts

### Workflow Execution Errors
- Check execution logs in the database
- Verify node configurations are valid
- Ensure AI API key is configured correctly

### Database Issues
- Database is created automatically in `packages/backend/data/`
- Delete database file to reset all data
- Check file permissions for database directory

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
