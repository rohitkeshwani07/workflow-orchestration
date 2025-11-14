# Docker Deployment Guide

This guide explains how to run the Workflow Orchestration platform using Docker and Docker Compose.

## Prerequisites

- Docker Engine 20.10+
- Docker Compose 2.0+

## Quick Start

1. **Clone the repository and set up environment variables**:
```bash
# Copy the example environment file
cp .env.example .env

# Edit .env and add your Anthropic API key
nano .env  # or use your preferred editor
```

2. **Build and start the containers**:
```bash
docker-compose up -d
```

3. **Access the application**:
- Frontend: http://localhost:3000
- Backend API: http://localhost:3001
- Health Check: http://localhost:3001/health

## Docker Compose Services

### Backend Service
- **Container**: `workflow-backend`
- **Port**: 3001
- **Technology**: Go 1.21 + Gin + SQLite
- **Volume**: `workflow-data` mounted at `/root/data` for database persistence

### Frontend Service
- **Container**: `workflow-frontend`
- **Port**: 3000 (mapped to internal port 80)
- **Technology**: React + Vite (served by Nginx)
- **Proxy**: Automatically proxies `/api` and `/ws` requests to backend

## Environment Variables

Set these in your `.env` file:

```env
# Required for AI Agent nodes
ANTHROPIC_API_KEY=sk-ant-your-key-here
```

## Docker Commands

### Start Services
```bash
# Start in foreground (see logs)
docker-compose up

# Start in background (detached)
docker-compose up -d

# Start and rebuild images
docker-compose up --build
```

### Stop Services
```bash
# Stop containers (preserves data)
docker-compose stop

# Stop and remove containers (preserves volumes)
docker-compose down

# Stop and remove containers and volumes (DELETES ALL DATA)
docker-compose down -v
```

### View Logs
```bash
# All services
docker-compose logs -f

# Backend only
docker-compose logs -f backend

# Frontend only
docker-compose logs -f frontend

# Last 100 lines
docker-compose logs --tail=100
```

### Rebuild Services
```bash
# Rebuild all services
docker-compose build

# Rebuild specific service
docker-compose build backend

# Force rebuild without cache
docker-compose build --no-cache
```

### Check Status
```bash
# List running containers
docker-compose ps

# Check health status
docker-compose ps
```

## Data Persistence

### Database Volume
The SQLite database is stored in a Docker volume named `workflow-data`. This ensures your workflows and execution history persist across container restarts.

**View volume location**:
```bash
docker volume inspect workflow-orchestration_workflow-data
```

**Backup database**:
```bash
# Create backup directory
mkdir -p ./backups

# Copy database from volume
docker run --rm -v workflow-orchestration_workflow-data:/data -v $(pwd)/backups:/backup alpine cp /data/workflows.db /backup/workflows-$(date +%Y%m%d-%H%M%S).db
```

**Restore database**:
```bash
# Restore from backup
docker run --rm -v workflow-orchestration_workflow-data:/data -v $(pwd)/backups:/backup alpine cp /backup/workflows-backup.db /data/workflows.db
```

## Development with Docker

### Live Development
For development, you may want to mount source code for live reloading:

```yaml
# docker-compose.dev.yml
version: '3.8'

services:
  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile
    volumes:
      - ./backend:/app
      - workflow-data:/root/data
    command: go run main.go

  frontend:
    build:
      context: ./packages/frontend
      dockerfile: Dockerfile.dev
    volumes:
      - ./packages/frontend:/app
      - /app/node_modules
```

Run with:
```bash
docker-compose -f docker-compose.dev.yml up
```

## Production Deployment

### Using Docker Compose
```bash
# Build optimized images
docker-compose build --no-cache

# Start in production mode
docker-compose up -d

# Monitor logs
docker-compose logs -f
```

### Environment Configuration
Create a production `.env` file:
```env
ANTHROPIC_API_KEY=your_production_key
```

### Resource Limits
Add resource limits to `docker-compose.yml`:

```yaml
services:
  backend:
    # ... existing config
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 512M
        reservations:
          cpus: '0.5'
          memory: 256M
```

## Networking

### Internal Network
Services communicate over Docker's internal network:
- Backend is accessible at `http://backend:3001`
- Frontend proxies requests to backend automatically

### External Access
- Frontend: Port 3000 (public)
- Backend: Port 3001 (public, can be restricted)

### Restrict Backend Access
To restrict direct backend access, remove the port mapping:

```yaml
backend:
  # Remove or comment out:
  # ports:
  #   - "3001:3001"
```

## Health Checks

Both services include health checks:

```bash
# Check health status
docker-compose ps

# Backend health endpoint
curl http://localhost:3001/health

# Frontend health (via Nginx)
curl http://localhost:3000
```

## Troubleshooting

### Container won't start
```bash
# Check logs
docker-compose logs backend
docker-compose logs frontend

# Check container status
docker-compose ps

# Restart specific service
docker-compose restart backend
```

### Database issues
```bash
# Enter backend container
docker-compose exec backend sh

# Check database file
ls -la /root/data/

# Test database connection
sqlite3 /root/data/workflows.db "SELECT * FROM workflows;"
```

### Network issues
```bash
# Check if services can communicate
docker-compose exec frontend ping backend

# Check listening ports
docker-compose exec backend netstat -tlnp
```

### Build issues
```bash
# Clean build cache
docker-compose build --no-cache

# Remove all containers and volumes
docker-compose down -v

# Rebuild from scratch
docker-compose up --build
```

### Permission issues
```bash
# Fix volume permissions
docker-compose exec backend chown -R root:root /root/data
```

## Scaling

### Multiple Backend Instances
Add a load balancer (Nginx or Traefik) and scale:

```bash
docker-compose up -d --scale backend=3
```

Note: You'll need to configure session affinity for WebSocket connections.

## Security Best Practices

1. **Use secrets for sensitive data**:
```yaml
services:
  backend:
    secrets:
      - anthropic_api_key

secrets:
  anthropic_api_key:
    file: ./secrets/anthropic_api_key.txt
```

2. **Run as non-root user**:
Add to Dockerfiles:
```dockerfile
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser
USER appuser
```

3. **Use read-only filesystem**:
```yaml
services:
  backend:
    read_only: true
    tmpfs:
      - /tmp
```

4. **Enable logging**:
```yaml
services:
  backend:
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

## Monitoring

### Docker Stats
```bash
# Real-time resource usage
docker stats
```

### Container Logs to File
```bash
# Export logs
docker-compose logs > logs/application-$(date +%Y%m%d).log
```

## Updates

### Updating the Application
```bash
# Pull latest code
git pull

# Rebuild and restart
docker-compose down
docker-compose up --build -d
```

### Updating Images
```bash
# Pull base image updates
docker-compose pull

# Rebuild with updated bases
docker-compose build --pull
docker-compose up -d
```

## Clean Up

### Remove Everything
```bash
# Stop and remove containers, networks
docker-compose down

# Also remove volumes (WARNING: deletes data)
docker-compose down -v

# Also remove images
docker-compose down --rmi all
```

### Prune Unused Resources
```bash
# Remove unused containers, networks, images
docker system prune

# Remove everything including volumes
docker system prune -a --volumes
```

## Support

For issues related to Docker deployment, check:
1. Container logs: `docker-compose logs`
2. Health checks: `docker-compose ps`
3. Resource usage: `docker stats`
4. Network connectivity: `docker network inspect workflow-orchestration_default`
