import { Node, NodeType, AIAgentConfig, HTTPRequestConfig, TransformConfig, ConditionConfig } from '@workflow/shared';
import Anthropic from 'anthropic';

export class NodeExecutor {
  private anthropic: Anthropic | null = null;

  constructor() {
    const apiKey = process.env.ANTHROPIC_API_KEY;
    if (apiKey) {
      this.anthropic = new Anthropic({ apiKey });
    }
  }

  async execute(node: Node, context: Record<string, any>): Promise<any> {
    switch (node.type) {
      case NodeType.CHAT_TRIGGER:
        return this.executeChatTrigger(node, context);
      case NodeType.AI_AGENT:
        return this.executeAIAgent(node, context);
      case NodeType.HTTP_REQUEST:
        return this.executeHTTPRequest(node, context);
      case NodeType.TRANSFORM:
        return this.executeTransform(node, context);
      case NodeType.CONDITION:
        return this.executeCondition(node, context);
      case NodeType.CODE:
        return this.executeCode(node, context);
      case NodeType.DELAY:
        return this.executeDelay(node, context);
      case NodeType.RESPONSE:
        return this.executeResponse(node, context);
      default:
        throw new Error(`Unknown node type: ${node.type}`);
    }
  }

  private executeChatTrigger(node: Node, context: Record<string, any>): any {
    // Chat trigger just passes through the trigger data
    return context.trigger || {};
  }

  private async executeAIAgent(node: Node, context: Record<string, any>): Promise<any> {
    if (!this.anthropic) {
      throw new Error('Anthropic API key not configured');
    }

    const config = node.data as AIAgentConfig;
    const userMessage = this.interpolateVariables(config.systemPrompt || '', context);

    // Get previous messages from context if available
    const messages: any[] = [];

    // Add user input from trigger
    if (context.trigger?.message) {
      messages.push({
        role: 'user',
        content: context.trigger.message
      });
    }

    const response = await this.anthropic.messages.create({
      model: config.model || 'claude-3-5-sonnet-20241022',
      max_tokens: config.maxTokens || 4096,
      temperature: config.temperature || 0.7,
      system: userMessage,
      messages: messages.length > 0 ? messages : [{ role: 'user', content: 'Hello' }]
    });

    const content = response.content[0];
    return {
      text: content.type === 'text' ? content.text : '',
      fullResponse: response
    };
  }

  private async executeHTTPRequest(node: Node, context: Record<string, any>): Promise<any> {
    const config = node.data as HTTPRequestConfig;
    const url = this.interpolateVariables(config.url, context);

    const headers = config.headers || {};
    const interpolatedHeaders: Record<string, string> = {};
    for (const [key, value] of Object.entries(headers)) {
      interpolatedHeaders[key] = this.interpolateVariables(value, context);
    }

    const options: RequestInit = {
      method: config.method,
      headers: interpolatedHeaders
    };

    if (config.body && ['POST', 'PUT', 'PATCH'].includes(config.method)) {
      options.body = JSON.stringify(this.interpolateObject(config.body, context));
    }

    const response = await fetch(url, options);
    const contentType = response.headers.get('content-type');

    let data;
    if (contentType?.includes('application/json')) {
      data = await response.json();
    } else {
      data = await response.text();
    }

    return {
      status: response.status,
      statusText: response.statusText,
      headers: Object.fromEntries(response.headers.entries()),
      data
    };
  }

  private executeTransform(node: Node, context: Record<string, any>): any {
    const config = node.data as TransformConfig;
    const expression = config.expression;

    // Create a safe execution context
    const safeContext = { ...context, console, JSON, Math, Date };

    try {
      const func = new Function(...Object.keys(safeContext), `return ${expression}`);
      return func(...Object.values(safeContext));
    } catch (error: any) {
      throw new Error(`Transform error: ${error.message}`);
    }
  }

  private executeCondition(node: Node, context: Record<string, any>): any {
    const config = node.data as ConditionConfig;
    const expression = config.expression;

    const safeContext = { ...context, console, JSON, Math, Date };

    try {
      const func = new Function(...Object.keys(safeContext), `return Boolean(${expression})`);
      const result = func(...Object.values(safeContext));
      return { condition: result, branch: result ? 'true' : 'false' };
    } catch (error: any) {
      throw new Error(`Condition error: ${error.message}`);
    }
  }

  private executeCode(node: Node, context: Record<string, any>): any {
    const code = node.data.code as string;

    const safeContext = { ...context, console, JSON, Math, Date };

    try {
      const func = new Function(...Object.keys(safeContext), code);
      return func(...Object.values(safeContext));
    } catch (error: any) {
      throw new Error(`Code execution error: ${error.message}`);
    }
  }

  private async executeDelay(node: Node, context: Record<string, any>): Promise<any> {
    const delayMs = node.data.delayMs || 1000;
    await new Promise(resolve => setTimeout(resolve, delayMs));
    return { delayed: delayMs };
  }

  private executeResponse(node: Node, context: Record<string, any>): any {
    const message = this.interpolateVariables(node.data.message || '', context);
    return { message };
  }

  private interpolateVariables(template: string, context: Record<string, any>): string {
    return template.replace(/\{\{(\w+(?:\.\w+)*)\}\}/g, (match, path) => {
      const value = this.getNestedValue(context, path);
      return value !== undefined ? String(value) : match;
    });
  }

  private interpolateObject(obj: any, context: Record<string, any>): any {
    if (typeof obj === 'string') {
      return this.interpolateVariables(obj, context);
    }
    if (Array.isArray(obj)) {
      return obj.map(item => this.interpolateObject(item, context));
    }
    if (obj && typeof obj === 'object') {
      const result: any = {};
      for (const [key, value] of Object.entries(obj)) {
        result[key] = this.interpolateObject(value, context);
      }
      return result;
    }
    return obj;
  }

  private getNestedValue(obj: any, path: string): any {
    return path.split('.').reduce((current, key) => current?.[key], obj);
  }
}
