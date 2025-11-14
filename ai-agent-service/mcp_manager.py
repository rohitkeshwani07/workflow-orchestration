"""MCP (Model Context Protocol) tool manager."""
import asyncio
from typing import Any, Optional
from mcp import ClientSession, StdioServerParameters
from mcp.client.stdio import stdio_client
from langchain_core.tools import BaseTool, StructuredTool
from pydantic import BaseModel, Field


class MCPToolManager:
    """Manages MCP server connections and tool bindings."""

    def __init__(self):
        self.servers: dict[str, ClientSession] = {}
        self.tools: dict[str, list[BaseTool]] = {}  # server_name -> tools

    async def connect_server(self, server_name: str, command: str, args: list[str] = None):
        """Connect to an MCP server."""
        if args is None:
            args = []

        try:
            server_params = StdioServerParameters(command=command, args=args, env=None)

            async with stdio_client(server_params) as (read, write):
                async with ClientSession(read, write) as session:
                    await session.initialize()
                    self.servers[server_name] = session

                    # List available tools
                    response = await session.list_tools()
                    tools = await self._convert_mcp_tools(server_name, response.tools)
                    self.tools[server_name] = tools

                    print(f"✓ Connected to MCP server '{server_name}' with {len(tools)} tools")
                    return True
        except Exception as e:
            print(f"✗ Failed to connect to MCP server '{server_name}': {e}")
            return False

    async def _convert_mcp_tools(self, server_name: str, mcp_tools: list) -> list[BaseTool]:
        """Convert MCP tools to LangChain tools."""
        langchain_tools = []

        for tool in mcp_tools:
            # Create a LangChain tool from MCP tool
            async def tool_func(server=server_name, tool_name=tool.name, **kwargs):
                return await self.call_tool(server, tool_name, kwargs)

            # Create pydantic model for tool arguments
            tool_args_model = self._create_args_model(tool)

            langchain_tool = StructuredTool(
                name=tool.name,
                description=tool.description or f"MCP tool: {tool.name}",
                func=tool_func,
                coroutine=tool_func,
                args_schema=tool_args_model,
            )

            langchain_tools.append(langchain_tool)

        return langchain_tools

    def _create_args_model(self, mcp_tool) -> type[BaseModel]:
        """Create a Pydantic model for tool arguments."""
        # Extract parameters from MCP tool schema
        if hasattr(mcp_tool, "inputSchema") and mcp_tool.inputSchema:
            schema = mcp_tool.inputSchema
            properties = schema.get("properties", {})
            required = schema.get("required", [])

            # Dynamically create Pydantic model fields
            fields = {}
            for prop_name, prop_schema in properties.items():
                field_type = self._json_type_to_python(prop_schema.get("type", "string"))
                field_default = ... if prop_name in required else None
                fields[prop_name] = (
                    field_type,
                    Field(default=field_default, description=prop_schema.get("description")),
                )

            # Create dynamic model
            return type(f"{mcp_tool.name}Args", (BaseModel,), fields)
        else:
            # No schema, create empty model
            return type(f"{mcp_tool.name}Args", (BaseModel,), {})

    def _json_type_to_python(self, json_type: str) -> type:
        """Convert JSON schema type to Python type."""
        type_mapping = {
            "string": str,
            "number": float,
            "integer": int,
            "boolean": bool,
            "array": list,
            "object": dict,
        }
        return type_mapping.get(json_type, str)

    async def call_tool(self, server_name: str, tool_name: str, arguments: dict[str, Any]) -> Any:
        """Call an MCP tool."""
        if server_name not in self.servers:
            raise ValueError(f"MCP server '{server_name}' not connected")

        session = self.servers[server_name]

        try:
            result = await session.call_tool(tool_name, arguments=arguments)
            return result.content if hasattr(result, "content") else result
        except Exception as e:
            print(f"Error calling MCP tool '{tool_name}': {e}")
            raise

    def get_tools(self, server_name: Optional[str] = None) -> list[BaseTool]:
        """Get all tools or tools from a specific server."""
        if server_name:
            return self.tools.get(server_name, [])

        # Return all tools from all servers
        all_tools = []
        for tools_list in self.tools.values():
            all_tools.extend(tools_list)
        return all_tools

    def get_tools_by_names(self, tool_names: list[str]) -> list[BaseTool]:
        """Get specific tools by name."""
        all_tools = self.get_tools()
        return [tool for tool in all_tools if tool.name in tool_names]

    async def disconnect_server(self, server_name: str):
        """Disconnect from an MCP server."""
        if server_name in self.servers:
            # MCP session cleanup
            del self.servers[server_name]
            if server_name in self.tools:
                del self.tools[server_name]
            print(f"Disconnected from MCP server '{server_name}'")

    async def disconnect_all(self):
        """Disconnect from all MCP servers."""
        for server_name in list(self.servers.keys()):
            await self.disconnect_server(server_name)


# Global MCP manager instance
mcp_manager = MCPToolManager()
