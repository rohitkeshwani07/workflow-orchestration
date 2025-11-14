"""Configuration management for AI Agent Service."""
from typing import Optional
from pydantic_settings import BaseSettings
from pydantic import Field


class Settings(BaseSettings):
    """Application settings."""

    # Service Configuration
    service_name: str = "ai-agent-service"
    host: str = "0.0.0.0"
    port: int = 8000
    debug: bool = False

    # Model Provider API Keys
    anthropic_api_key: Optional[str] = Field(None, env="ANTHROPIC_API_KEY")
    openai_api_key: Optional[str] = Field(None, env="OPENAI_API_KEY")

    # Default Model Configuration
    default_provider: str = "anthropic"
    default_model: str = "claude-3-5-sonnet-20241022"
    default_temperature: float = 0.7
    default_max_tokens: int = 4096

    # Redis Configuration (for state management)
    redis_host: str = "redis"
    redis_port: int = 6379
    redis_db: int = 0
    redis_password: Optional[str] = None
    use_redis: bool = True  # Set to False for in-memory state

    # MCP Server Configuration
    mcp_servers: list[str] = Field(default_factory=list)

    # CORS Configuration
    cors_origins: list[str] = Field(
        default_factory=lambda: ["http://localhost:3000", "http://localhost:3001"]
    )

    class Config:
        env_file = ".env"
        env_file_encoding = "utf-8"
        case_sensitive = False


# Global settings instance
settings = Settings()
