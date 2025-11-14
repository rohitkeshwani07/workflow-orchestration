"""LangGraph AI Agent implementation with multi-model support."""
from typing import Annotated, Any, Literal, Optional, TypedDict
from uuid import uuid4

from langgraph.graph import StateGraph, END
from langgraph.graph.message import add_messages
from langgraph.prebuilt import ToolNode
from langchain_core.messages import BaseMessage, HumanMessage, AIMessage, SystemMessage
from langchain_core.tools import BaseTool
from langchain_anthropic import ChatAnthropic
from langchain_openai import ChatOpenAI

from config import settings
from models import AgentConfig, Message
from state_manager import state_manager
from mcp_manager import mcp_manager


class AgentState(TypedDict):
    """State for the agent graph."""

    messages: Annotated[list[BaseMessage], add_messages]
    session_id: str
    agent_id: str


class AIAgent:
    """LangGraph-based AI Agent with tool support and memory."""

    def __init__(self, config: AgentConfig):
        self.config = config
        self.llm = self._create_llm()
        self.tools: list[BaseTool] = []
        self.graph = None

        # Initialize tools if provided
        if config.tools:
            self._bind_tools()

        # Build the graph
        self._build_graph()

    def _create_llm(self):
        """Create LLM instance based on provider."""
        provider = self.config.provider.lower()
        model = self.config.model
        temperature = self.config.temperature
        max_tokens = self.config.max_tokens

        if provider == "anthropic":
            if not settings.anthropic_api_key:
                raise ValueError("Anthropic API key not configured")

            return ChatAnthropic(
                model=model,
                anthropic_api_key=settings.anthropic_api_key,
                temperature=temperature,
                max_tokens=max_tokens,
            )

        elif provider == "openai":
            if not settings.openai_api_key:
                raise ValueError("OpenAI API key not configured")

            return ChatOpenAI(
                model=model,
                openai_api_key=settings.openai_api_key,
                temperature=temperature,
                max_tokens=max_tokens,
            )

        else:
            raise ValueError(f"Unsupported provider: {provider}")

    def _bind_tools(self):
        """Bind MCP tools to the agent."""
        tool_names = [tool.name for tool in self.config.tools]

        # Get tools from MCP manager
        self.tools = mcp_manager.get_tools_by_names(tool_names)

        # Bind tools to LLM
        if self.tools:
            self.llm = self.llm.bind_tools(self.tools)

    def _build_graph(self):
        """Build the LangGraph workflow."""
        workflow = StateGraph(AgentState)

        # Define nodes
        workflow.add_node("agent", self._call_model)

        if self.tools:
            workflow.add_node("tools", ToolNode(self.tools))

        # Define edges
        workflow.set_entry_point("agent")

        if self.tools:
            workflow.add_conditional_edges(
                "agent",
                self._should_continue,
                {
                    "tools": "tools",
                    "end": END,
                },
            )
            workflow.add_edge("tools", "agent")
        else:
            workflow.add_edge("agent", END)

        # Compile the graph
        self.graph = workflow.compile()

    async def _call_model(self, state: AgentState) -> dict[str, Any]:
        """Call the LLM with the current state."""
        messages = state["messages"]

        # Add system prompt if configured
        if self.config.system_prompt:
            # Check if system message already exists
            has_system = any(isinstance(msg, SystemMessage) for msg in messages)
            if not has_system:
                messages = [SystemMessage(content=self.config.system_prompt)] + messages

        response = await self.llm.ainvoke(messages)
        return {"messages": [response]}

    def _should_continue(self, state: AgentState) -> Literal["tools", "end"]:
        """Determine if we should continue to tools or end."""
        messages = state["messages"]
        last_message = messages[-1]

        # Check if there are tool calls
        if hasattr(last_message, "tool_calls") and last_message.tool_calls:
            return "tools"
        return "end"

    async def invoke(
        self, message: str, session_id: Optional[str] = None, context: dict[str, Any] = None
    ) -> tuple[str, str, list[dict]]:
        """
        Invoke the agent with a message.

        Returns:
            Tuple of (response, session_id, tool_calls)
        """
        if session_id is None:
            session_id = str(uuid4())

        if context is None:
            context = {}

        # Load conversation history if memory is enabled
        messages = []
        if self.config.memory_enabled:
            history = await state_manager.get_messages(
                self.config.agent_id, session_id, limit=self.config.max_memory_messages
            )
            # Convert to LangChain messages
            for msg in history:
                if msg.role == "user":
                    messages.append(HumanMessage(content=msg.content))
                elif msg.role == "assistant":
                    messages.append(AIMessage(content=msg.content))
                elif msg.role == "system":
                    messages.append(SystemMessage(content=msg.content))

        # Add current message
        messages.append(HumanMessage(content=message))

        # Create initial state
        initial_state = {
            "messages": messages,
            "session_id": session_id,
            "agent_id": self.config.agent_id,
        }

        # Run the graph
        result = await self.graph.ainvoke(initial_state)

        # Extract response
        final_messages = result["messages"]
        last_message = final_messages[-1]

        response_text = last_message.content if hasattr(last_message, "content") else str(last_message)

        # Extract tool calls if any
        tool_calls = []
        if hasattr(last_message, "tool_calls"):
            tool_calls = last_message.tool_calls

        # Save messages to memory if enabled
        if self.config.memory_enabled:
            await state_manager.save_message(
                self.config.agent_id, session_id, Message(role="user", content=message)
            )
            await state_manager.save_message(
                self.config.agent_id,
                session_id,
                Message(role="assistant", content=response_text),
            )

        return response_text, session_id, tool_calls


class AgentManager:
    """Manages multiple AI agents."""

    def __init__(self):
        self.agents: dict[str, AIAgent] = {}

    async def create_agent(self, config: AgentConfig) -> AIAgent:
        """Create a new agent."""
        agent = AIAgent(config)
        self.agents[config.agent_id] = agent

        # Save config to state manager
        await state_manager.save_agent_config(
            config.agent_id, config.model_dump()
        )

        return agent

    def get_agent(self, agent_id: str) -> Optional[AIAgent]:
        """Get an existing agent."""
        return self.agents.get(agent_id)

    async def delete_agent(self, agent_id: str):
        """Delete an agent."""
        if agent_id in self.agents:
            del self.agents[agent_id]
            await state_manager.clear_agent(agent_id)

    def list_agents(self) -> list[str]:
        """List all agent IDs."""
        return list(self.agents.keys())

    async def reload_agent(self, agent_id: str) -> Optional[AIAgent]:
        """Reload an agent from stored config."""
        config_data = await state_manager.get_agent_config(agent_id)
        if config_data:
            config = AgentConfig(**config_data)
            return await self.create_agent(config)
        return None


# Global agent manager instance
agent_manager = AgentManager()
