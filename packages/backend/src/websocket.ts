import { WebSocketServer, WebSocket } from 'ws';
import { Server } from 'http';
import { v4 as uuidv4 } from 'uuid';
import { WorkflowDB } from './database';
import { WorkflowEngine } from './engine/WorkflowEngine';
import { ChatMessage } from '@workflow/shared';

interface ChatSession {
  id: string;
  workflowId: string;
  ws: WebSocket;
}

export class ChatWebSocketServer {
  private wss: WebSocketServer;
  private sessions: Map<string, ChatSession> = new Map();
  private engine: WorkflowEngine;

  constructor(server: Server) {
    this.wss = new WebSocketServer({ server, path: '/ws/chat' });
    this.engine = new WorkflowEngine();
    this.setupWebSocket();
  }

  private setupWebSocket() {
    this.wss.on('connection', (ws: WebSocket) => {
      console.log('New WebSocket connection');

      ws.on('message', async (data: Buffer) => {
        try {
          const message = JSON.parse(data.toString());
          await this.handleMessage(ws, message);
        } catch (error: any) {
          console.error('WebSocket message error:', error);
          ws.send(JSON.stringify({
            type: 'error',
            error: error.message
          }));
        }
      });

      ws.on('close', () => {
        // Clean up session
        for (const [sessionId, session] of this.sessions.entries()) {
          if (session.ws === ws) {
            this.sessions.delete(sessionId);
            console.log(`Session ${sessionId} closed`);
            break;
          }
        }
      });
    });
  }

  private async handleMessage(ws: WebSocket, message: any) {
    switch (message.type) {
      case 'start_session':
        await this.startSession(ws, message.workflowId);
        break;

      case 'chat_message':
        await this.handleChatMessage(ws, message.sessionId, message.content);
        break;

      default:
        ws.send(JSON.stringify({
          type: 'error',
          error: `Unknown message type: ${message.type}`
        }));
    }
  }

  private async startSession(ws: WebSocket, workflowId: string) {
    const workflow = WorkflowDB.getWorkflow(workflowId);
    if (!workflow) {
      ws.send(JSON.stringify({
        type: 'error',
        error: 'Workflow not found'
      }));
      return;
    }

    const sessionId = uuidv4();
    WorkflowDB.createChatSession(sessionId, workflowId);

    this.sessions.set(sessionId, {
      id: sessionId,
      workflowId,
      ws
    });

    ws.send(JSON.stringify({
      type: 'session_started',
      sessionId,
      workflowId
    }));

    console.log(`Started chat session ${sessionId} for workflow ${workflowId}`);
  }

  private async handleChatMessage(ws: WebSocket, sessionId: string, content: string) {
    const session = this.sessions.get(sessionId);
    if (!session) {
      ws.send(JSON.stringify({
        type: 'error',
        error: 'Session not found'
      }));
      return;
    }

    // Save user message
    const userMessage: ChatMessage = {
      id: uuidv4(),
      sessionId,
      role: 'user',
      content,
      timestamp: new Date().toISOString()
    };
    WorkflowDB.addChatMessage(userMessage);

    // Send acknowledgment
    ws.send(JSON.stringify({
      type: 'message_received',
      message: userMessage
    }));

    try {
      // Execute workflow with chat message as trigger
      const execution = await this.engine.executeFromTrigger(session.workflowId, {
        message: content,
        sessionId
      });

      // Get response from workflow execution
      const responseNode = Object.keys(execution.context).find(key =>
        key !== 'trigger' && execution.context[key]?.message
      );

      let responseContent = 'Workflow executed successfully';
      if (responseNode && execution.context[responseNode]?.message) {
        responseContent = execution.context[responseNode].message;
      } else if (execution.context.response?.message) {
        responseContent = execution.context.response.message;
      } else {
        // Check for AI agent response
        const aiAgentNode = Object.keys(execution.context).find(key =>
          execution.context[key]?.text
        );
        if (aiAgentNode) {
          responseContent = execution.context[aiAgentNode].text;
        }
      }

      // Save assistant message
      const assistantMessage: ChatMessage = {
        id: uuidv4(),
        sessionId,
        role: 'assistant',
        content: responseContent,
        timestamp: new Date().toISOString()
      };
      WorkflowDB.addChatMessage(assistantMessage);

      // Send response
      ws.send(JSON.stringify({
        type: 'chat_response',
        message: assistantMessage,
        executionId: execution.id
      }));

    } catch (error: any) {
      console.error('Workflow execution error:', error);

      const errorMessage: ChatMessage = {
        id: uuidv4(),
        sessionId,
        role: 'assistant',
        content: `Error: ${error.message}`,
        timestamp: new Date().toISOString()
      };
      WorkflowDB.addChatMessage(errorMessage);

      ws.send(JSON.stringify({
        type: 'chat_response',
        message: errorMessage,
        error: error.message
      }));
    }
  }

  getActiveSessions(): number {
    return this.sessions.size;
  }
}
