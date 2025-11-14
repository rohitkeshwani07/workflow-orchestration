# Getting Started with Workflow Orchestration

This guide will help you get up and running with the Workflow Orchestration platform in just a few minutes.

## Installation

1. **Install dependencies**:
```bash
# Install Go dependencies
cd backend
go mod download
cd ..

# Install frontend dependencies
npm install
```

2. **Set up your API key**:
```bash
cd backend
cp .env.example .env
```

Edit `.env` and add your Anthropic API key:
```
ANTHROPIC_API_KEY=sk-ant-...
```

3. **Start the application**:
```bash
cd ..  # Back to root
npm run dev
```

The application will start:
- Go Backend: http://localhost:3001
- React Frontend: http://localhost:3000

## Your First AI Chatbot Workflow

Let's create a simple AI chatbot that responds to user messages.

### Step 1: Create a Workflow

1. Open http://localhost:3000 in your browser
2. Click **"Create Workflow"**
3. Change the workflow name to "AI Chatbot"

### Step 2: Add Nodes

1. From the left palette, click these nodes to add them:
   - **Chat Trigger**
   - **AI Agent**
   - **Response**

2. Connect the nodes:
   - Drag from the **bottom circle** of "Chat Trigger"
   - Drop on the **top circle** of "AI Agent"
   - Connect "AI Agent" to "Response" the same way

### Step 3: Configure the AI Agent

1. Click on the **AI Agent** node
2. In the right panel, configure:
   - **Model**: Claude 3.5 Sonnet (already selected)
   - **System Prompt**:
   ```
   You are a friendly and helpful AI assistant. Answer questions clearly and concisely. Be conversational and warm in your responses.
   ```
   - **Temperature**: 0.7
   - **Max Tokens**: 4096

### Step 4: Configure the Response

1. Click on the **Response** node
2. Set the **Response Message**:
```
{{[your-ai-agent-node-id].text}}
```

> **Tip**: Replace `[your-ai-agent-node-id]` with the actual ID of your AI Agent node. You can see the node ID in the panel when you click on it.

### Step 5: Save and Test

1. Click **"Save"** at the top right
2. Click **"Test Chat"**
3. Type a message like "Hello! Can you tell me a joke?"
4. Watch your AI chatbot respond!

## Advanced Example: Weather-Aware Assistant

Let's create a more advanced workflow that fetches weather data and uses AI to respond.

### Workflow Design
```
Chat Trigger → HTTP Request → AI Agent → Response
```

### Configuration

1. **HTTP Request** node:
   - Method: GET
   - URL: `https://api.open-meteo.com/v1/forecast?latitude=40.7128&longitude=-74.0060&current=temperature_2m,weather_code`
   - Headers: `{}`

2. **AI Agent** node:
   - System Prompt:
   ```
   You are a helpful weather assistant. The user asked: {{trigger.message}}

   Here is the current weather data: {{[http-request-node-id].data}}

   Provide a friendly, conversational response about the weather based on this data.
   ```

3. **Response** node:
   - Message: `{{[ai-agent-node-id].text}}`

## Using Conditions

Create workflows that branch based on user input:

```
Chat Trigger → Condition → [AI Agent A | AI Agent B] → Response
```

**Condition** node expression:
```javascript
trigger.message.toLowerCase().includes('weather')
```

This routes weather-related questions to one AI agent and other questions to another.

## Tips & Tricks

### Variable Interpolation
Use `{{variableName}}` to access data from previous nodes:
- `{{trigger.message}}` - The user's message
- `{{nodeId.fieldName}}` - Output from a specific node

### Node IDs
- Each node has a unique ID like "node-1234567890"
- Click on a node to see its ID in the config panel
- Use these IDs to reference node outputs

### Testing Workflows
1. Use the **Execute** button to run the workflow manually
2. Use **Test Chat** to interact with it in real-time
3. Check execution logs in the database for debugging

### Best Practices
1. **Start Simple**: Begin with Chat Trigger → AI Agent → Response
2. **Test Incrementally**: Test after adding each node
3. **Use Clear Labels**: Name your nodes descriptively
4. **Save Often**: Click Save frequently to avoid losing work

## Common Use Cases

### Customer Support Bot
```
Chat Trigger → Condition → [FAQ AI | Support Ticket Creator] → Response
```

### Content Generator
```
Chat Trigger → AI Agent (Generator) → Transform → AI Agent (Reviewer) → Response
```

### API Integration
```
Chat Trigger → Transform (Parse Input) → HTTP Request → AI Agent (Format) → Response
```

### Multi-Step Assistant
```
Chat Trigger → AI Agent (Understand) → Code (Process) → HTTP Request → AI Agent (Respond) → Response
```

## Next Steps

- Explore all node types in the palette
- Check out the API documentation in README.md
- Build complex multi-step workflows
- Integrate with external APIs
- Create custom transformations with JavaScript

## Need Help?

- Check the README.md for detailed documentation
- Review the API endpoints and WebSocket protocol
- Open an issue on GitHub for bugs or feature requests

Happy workflow building!
