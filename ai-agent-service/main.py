"""FastAPI application for AI Agent Service."""
import asyncio
from contextlib import asynccontextmanager
from uuid import uuid4

from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse

from config import settings
from models import (
    AgentConfig,
    AgentRequest,
    AgentResponse,
    AgentStatus,
    CreateAgentRequest,
    HealthResponse,
    Tool,
)
from agent import agent_manager
from state_manager import state_manager
from mcp_manager import mcp_manager


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Application lifespan events."""
    # Startup
    print("🚀 Starting AI Agent Service...")

    # Connect to Redis
    await state_manager.connect()

    # Connect to MCP servers if configured
    for server_config in settings.mcp_servers:
        # Parse server config: "name:command:args"
        parts = server_config.split(":", 2)
        if len(parts) >= 2:
            server_name = parts[0]
            command = parts[1]
            args = parts[2].split(",") if len(parts) > 2 else []
            await mcp_manager.connect_server(server_name, command, args)

    print(f"✓ AI Agent Service running on {settings.host}:{settings.port}")

    yield

    # Shutdown
    print("🛑 Shutting down AI Agent Service...")
    await state_manager.disconnect()
    await mcp_manager.disconnect_all()


# Create FastAPI app
app = FastAPI(
    title="AI Agent Service",
    description="LangGraph-based AI Agent Service with MCP tools and memory",
    version="0.1.0",
    lifespan=lifespan,
)

# Add CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.cors_origins,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


@app.get("/health", response_model=HealthResponse)
async def health_check():
    """Health check endpoint."""
    # Check provider availability
    providers = {
        "anthropic": bool(settings.anthropic_api_key),
        "openai": bool(settings.openai_api_key),
    }

    # Check Redis connection
    redis_connected = state_manager.redis_client is not None

    return HealthResponse(
        status="healthy",
        service=settings.service_name,
        version="0.1.0",
        providers=providers,
        redis_connected=redis_connected,
    )


@app.post("/agents", response_model=AgentStatus)
async def create_agent(request: CreateAgentRequest):
    """Create a new AI agent."""
    try:
        agent = await agent_manager.create_agent(request.config)

        return AgentStatus(
            agent_id=agent.config.agent_id,
            active=True,
            provider=agent.config.provider,
            model=agent.config.model,
            tools_count=len(agent.tools),
            memory_enabled=agent.config.memory_enabled,
            created_at=asyncio.get_event_loop().time(),
        )
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to create agent: {str(e)}")


@app.get("/agents/{agent_id}", response_model=AgentStatus)
async def get_agent_status(agent_id: str):
    """Get agent status."""
    agent = agent_manager.get_agent(agent_id)

    if not agent:
        raise HTTPException(status_code=404, detail=f"Agent {agent_id} not found")

    return AgentStatus(
        agent_id=agent.config.agent_id,
        active=True,
        provider=agent.config.provider,
        model=agent.config.model,
        tools_count=len(agent.tools),
        memory_enabled=agent.config.memory_enabled,
        created_at=asyncio.get_event_loop().time(),
    )


@app.get("/agents")
async def list_agents():
    """List all agents."""
    agent_ids = agent_manager.list_agents()
    return {"agents": agent_ids, "count": len(agent_ids)}


@app.delete("/agents/{agent_id}")
async def delete_agent(agent_id: str):
    """Delete an agent."""
    agent = agent_manager.get_agent(agent_id)

    if not agent:
        raise HTTPException(status_code=404, detail=f"Agent {agent_id} not found")

    await agent_manager.delete_agent(agent_id)
    return {"message": f"Agent {agent_id} deleted successfully"}


@app.post("/agents/{agent_id}/chat", response_model=AgentResponse)
async def chat_with_agent(agent_id: str, request: AgentRequest):
    """Send a message to an agent."""
    agent = agent_manager.get_agent(agent_id)

    if not agent:
        # Try to reload from state
        agent = await agent_manager.reload_agent(agent_id)
        if not agent:
            raise HTTPException(status_code=404, detail=f"Agent {agent_id} not found")

    try:
        response_text, session_id, tool_calls = await agent.invoke(
            message=request.message, session_id=request.session_id, context=request.context
        )

        return AgentResponse(
            agent_id=agent_id,
            session_id=session_id,
            response=response_text,
            tool_calls=tool_calls,
            metadata={
                "provider": agent.config.provider,
                "model": agent.config.model,
                "tools_used": len(tool_calls),
            },
        )
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Agent invocation failed: {str(e)}")


@app.post("/agents/{agent_id}/sessions/{session_id}/clear")
async def clear_session(agent_id: str, session_id: str):
    """Clear conversation history for a session."""
    await state_manager.delete_session(agent_id, session_id)
    return {"message": f"Session {session_id} cleared for agent {agent_id}"}


@app.get("/agents/{agent_id}/sessions/{session_id}/history")
async def get_session_history(agent_id: str, session_id: str, limit: int = 50):
    """Get conversation history for a session."""
    messages = await state_manager.get_messages(agent_id, session_id, limit=limit)
    return {
        "agent_id": agent_id,
        "session_id": session_id,
        "messages": [msg.model_dump() for msg in messages],
        "count": len(messages),
    }


@app.get("/tools")
async def list_tools():
    """List all available MCP tools."""
    all_tools = mcp_manager.get_tools()
    return {
        "tools": [
            {
                "name": tool.name,
                "description": tool.description,
            }
            for tool in all_tools
        ],
        "count": len(all_tools),
    }


@app.get("/mcp/servers")
async def list_mcp_servers():
    """List all connected MCP servers."""
    return {
        "servers": list(mcp_manager.servers.keys()),
        "count": len(mcp_manager.servers),
    }


@app.post("/mcp/servers/{server_name}/connect")
async def connect_mcp_server(server_name: str, command: str, args: list[str] = None):
    """Connect to an MCP server."""
    success = await mcp_manager.connect_server(server_name, command, args or [])
    if success:
        tools = mcp_manager.get_tools(server_name)
        return {
            "message": f"Connected to MCP server '{server_name}'",
            "tools_count": len(tools),
        }
    else:
        raise HTTPException(
            status_code=500, detail=f"Failed to connect to MCP server '{server_name}'"
        )


@app.delete("/mcp/servers/{server_name}")
async def disconnect_mcp_server(server_name: str):
    """Disconnect from an MCP server."""
    await mcp_manager.disconnect_server(server_name)
    return {"message": f"Disconnected from MCP server '{server_name}'"}


# Error handlers
@app.exception_handler(Exception)
async def global_exception_handler(request, exc):
    """Global exception handler."""
    return JSONResponse(
        status_code=500,
        content={
            "error": "Internal server error",
            "detail": str(exc),
            "type": type(exc).__name__,
        },
    )


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(
        "main:app",
        host=settings.host,
        port=settings.port,
        reload=settings.debug,
        log_level="info",
    )
