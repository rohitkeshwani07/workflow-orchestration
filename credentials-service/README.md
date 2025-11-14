# Credentials Service

A standalone microservice for secure credential management with AES-256-GCM encryption.

## Features

- **Secure Storage**: Credentials are encrypted at rest using AES-256-GCM
- **Multiple Credential Types**: Supports API keys, OAuth, Basic Auth, Bearer Tokens, and custom credentials
- **RESTful API**: Simple HTTP API for CRUD operations
- **Database Support**: Works with both PostgreSQL and SQLite
- **Docker Ready**: Containerized for easy deployment

## API Endpoints

### GET /health
Health check endpoint

### GET /api/credentials
List all credentials (without decrypted values)

### GET /api/credentials/:id
Get a specific credential (includes decrypted value)

### POST /api/credentials
Create a new credential

**Request Body:**
```json
{
  "name": "My API Key",
  "type": "api_key",
  "description": "OpenAI API Key",
  "value": "sk-...",
  "metadata": "{}"
}
```

### PUT /api/credentials/:id
Update an existing credential

### DELETE /api/credentials/:id
Delete a credential

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| PORT | Service port | 3002 |
| DB_DRIVER | Database driver (postgres or sqlite) | sqlite |
| DB_HOST | Database host | localhost |
| DB_PORT | Database port | 5432 |
| DB_USER | Database user | - |
| DB_PASSWORD | Database password | - |
| DB_NAME | Database name | credentials |
| DATABASE_PATH | SQLite database path | ./data/credentials.db |
| ENCRYPTION_KEY | Base64-encoded 32-byte encryption key | Auto-generated |

## Encryption

Credentials are encrypted using AES-256-GCM before being stored in the database. The encryption key should be a 32-byte key encoded in base64.

Generate a secure encryption key:
```bash
openssl rand -base64 32
```

**Important**: Keep the `ENCRYPTION_KEY` secure and consistent across deployments. If you lose the key, you cannot decrypt existing credentials.

## Running Locally

```bash
# Install dependencies
go mod download

# Run the service
go run main.go
```

## Docker

Build and run with Docker:
```bash
docker build -t credentials-service .
docker run -p 3002:3002 \
  -e ENCRYPTION_KEY=your_key_here \
  -e DB_DRIVER=sqlite \
  credentials-service
```

## Architecture

The credentials service is designed to be a standalone microservice that:
- Handles all credential encryption/decryption
- Manages credential storage independently
- Can be scaled separately from other services
- Provides a simple HTTP API for credential operations

Other services (like the workflow backend) communicate with this service via HTTP to retrieve credentials when needed.
