"""Data models and schemas for AI Agent Service."""
from typing import Any, Optional, Literal
from pydantic import BaseModel, Field
from datetime import datetime


class Message(BaseModel):
    """Chat message model."""

    role: Literal["user", "assistant", "system"] = "user"
    content: str
    timestamp: Optional[datetime] = None


class Tool(BaseModel):
    """MCP Tool definition."""

    name: str
    description: str
    parameters: dict[str, Any] = Field(default_factory=dict)
    server: Optional[str] = None  # MCP server name


class AgentConfig(BaseModel):
    """Configuration for an AI agent."""

    agent_id: str
    provider: str = "anthropic"  # anthropic, openai, etc.
    model: str = "claude-3-5-sonnet-20241022"
    temperature: float = 0.7
    max_tokens: int = 4096
    system_prompt: Optional[str] = None
    tools: list[Tool] = Field(default_factory=list)
    memory_enabled: bool = True
    max_memory_messages: int = 50


class AgentRequest(BaseModel):
    """Request to interact with an agent."""

    agent_id: str
    message: str
    session_id: Optional[str] = None
    context: dict[str, Any] = Field(default_factory=dict)


class AgentResponse(BaseModel):
    """Response from an agent."""

    agent_id: str
    session_id: str
    response: str
    tool_calls: list[dict[str, Any]] = Field(default_factory=list)
    metadata: dict[str, Any] = Field(default_factory=dict)
    timestamp: datetime = Field(default_factory=datetime.now)


class CreateAgentRequest(BaseModel):
    """Request to create a new agent."""

    agent_id: str
    config: AgentConfig


class AgentStatus(BaseModel):
    """Status of an agent."""

    agent_id: str
    active: bool
    provider: str
    model: str
    tools_count: int
    memory_enabled: bool
    created_at: datetime


class HealthResponse(BaseModel):
    """Health check response."""

    status: str
    service: str
    version: str
    providers: dict[str, bool]  # Provider availability
    redis_connected: bool
