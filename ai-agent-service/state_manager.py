"""State and memory management for AI agents."""
import json
from typing import Any, Optional
from datetime import datetime, timedelta
import redis.asyncio as redis
from config import settings
from models import Message


class StateManager:
    """Manages agent state and conversation memory."""

    def __init__(self):
        self.redis_client: Optional[redis.Redis] = None
        self.memory_store: dict[str, list[Message]] = {}  # Fallback in-memory store

    async def connect(self):
        """Connect to Redis if enabled."""
        if settings.use_redis:
            try:
                self.redis_client = redis.Redis(
                    host=settings.redis_host,
                    port=settings.redis_port,
                    db=settings.redis_db,
                    password=settings.redis_password,
                    decode_responses=True,
                )
                await self.redis_client.ping()
                print("✓ Connected to Redis for state management")
            except Exception as e:
                print(f"⚠ Redis connection failed: {e}. Using in-memory state.")
                self.redis_client = None
        else:
            print("Using in-memory state management")

    async def disconnect(self):
        """Disconnect from Redis."""
        if self.redis_client:
            await self.redis_client.close()

    def _get_session_key(self, agent_id: str, session_id: str) -> str:
        """Generate Redis key for session."""
        return f"agent:{agent_id}:session:{session_id}"

    def _get_agent_config_key(self, agent_id: str) -> str:
        """Generate Redis key for agent config."""
        return f"agent:{agent_id}:config"

    async def save_message(
        self, agent_id: str, session_id: str, message: Message, ttl: int = 86400
    ):
        """Save a message to conversation history."""
        key = self._get_session_key(agent_id, session_id)

        if self.redis_client:
            try:
                # Get existing messages
                existing = await self.redis_client.get(key)
                messages = json.loads(existing) if existing else []

                # Add new message
                messages.append(
                    {
                        "role": message.role,
                        "content": message.content,
                        "timestamp": (
                            message.timestamp.isoformat()
                            if message.timestamp
                            else datetime.now().isoformat()
                        ),
                    }
                )

                # Limit message history
                max_messages = 100  # Configurable
                if len(messages) > max_messages:
                    messages = messages[-max_messages:]

                # Save to Redis with TTL
                await self.redis_client.setex(key, ttl, json.dumps(messages))
            except Exception as e:
                print(f"Error saving message to Redis: {e}")
                # Fallback to in-memory
                self._save_to_memory(agent_id, session_id, message)
        else:
            self._save_to_memory(agent_id, session_id, message)

    def _save_to_memory(self, agent_id: str, session_id: str, message: Message):
        """Save message to in-memory store."""
        key = f"{agent_id}:{session_id}"
        if key not in self.memory_store:
            self.memory_store[key] = []
        self.memory_store[key].append(message)

        # Limit in-memory history
        if len(self.memory_store[key]) > 100:
            self.memory_store[key] = self.memory_store[key][-100:]

    async def get_messages(
        self, agent_id: str, session_id: str, limit: Optional[int] = None
    ) -> list[Message]:
        """Retrieve conversation history."""
        key = self._get_session_key(agent_id, session_id)

        if self.redis_client:
            try:
                data = await self.redis_client.get(key)
                if data:
                    messages_data = json.loads(data)
                    messages = [
                        Message(
                            role=msg["role"],
                            content=msg["content"],
                            timestamp=datetime.fromisoformat(msg["timestamp"]),
                        )
                        for msg in messages_data
                    ]
                    if limit:
                        return messages[-limit:]
                    return messages
            except Exception as e:
                print(f"Error retrieving messages from Redis: {e}")

        # Fallback to in-memory
        mem_key = f"{agent_id}:{session_id}"
        messages = self.memory_store.get(mem_key, [])
        if limit:
            return messages[-limit:]
        return messages

    async def save_agent_config(self, agent_id: str, config: dict[str, Any]):
        """Save agent configuration."""
        key = self._get_agent_config_key(agent_id)

        if self.redis_client:
            try:
                await self.redis_client.set(key, json.dumps(config))
            except Exception as e:
                print(f"Error saving agent config: {e}")
        else:
            # Store in memory
            self.memory_store[f"config:{agent_id}"] = config

    async def get_agent_config(self, agent_id: str) -> Optional[dict[str, Any]]:
        """Retrieve agent configuration."""
        key = self._get_agent_config_key(agent_id)

        if self.redis_client:
            try:
                data = await self.redis_client.get(key)
                if data:
                    return json.loads(data)
            except Exception as e:
                print(f"Error retrieving agent config: {e}")

        # Fallback to in-memory
        return self.memory_store.get(f"config:{agent_id}")

    async def delete_session(self, agent_id: str, session_id: str):
        """Delete a conversation session."""
        key = self._get_session_key(agent_id, session_id)

        if self.redis_client:
            try:
                await self.redis_client.delete(key)
            except Exception as e:
                print(f"Error deleting session: {e}")

        # Also delete from in-memory
        mem_key = f"{agent_id}:{session_id}"
        if mem_key in self.memory_store:
            del self.memory_store[mem_key]

    async def clear_agent(self, agent_id: str):
        """Clear all data for an agent."""
        if self.redis_client:
            try:
                # Find all keys for this agent
                pattern = f"agent:{agent_id}:*"
                keys = []
                async for key in self.redis_client.scan_iter(match=pattern):
                    keys.append(key)
                if keys:
                    await self.redis_client.delete(*keys)
            except Exception as e:
                print(f"Error clearing agent data: {e}")

        # Clear in-memory data
        keys_to_delete = [k for k in self.memory_store.keys() if k.startswith(agent_id)]
        for key in keys_to_delete:
            del self.memory_store[key]


# Global state manager instance
state_manager = StateManager()
