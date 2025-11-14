# Database Migrations Guide

This project uses [golang-migrate](https://github.com/golang-migrate/migrate) for database migrations. Migrations ensure your database schema is version-controlled and can be reliably applied across environments.

## Overview

- **Migration files**: Located in `backend/migrations/`
- **Migration tool**: Custom CLI in `backend/cmd/migrate/`
- **Automatic migrations**: Run via Docker Compose before backend starts
- **Manual migrations**: Available via Makefile commands

## Quick Start

### Docker (Automatic)

Migrations run automatically when you start with docker-compose:

```bash
docker-compose up -d
```

The `migrate` service runs before the `backend` service starts, ensuring your database is always up-to-date.

### Local Development

```bash
# Apply all pending migrations
make migrate-up

# Rollback the last migration
make migrate-down

# Check current migration version
make migrate-version
```

## Migration Files

Migration files are located in `backend/migrations/` and follow this naming convention:

```
NNNNNN_description.up.sql    # Forward migration
NNNNNN_description.down.sql  # Rollback migration
```

Example:
```
000001_initial_schema.up.sql
000001_initial_schema.down.sql
000002_add_user_table.up.sql
000002_add_user_table.down.sql
```

### Current Migrations

**000001_initial_schema** (Initial database schema)
- Creates `workflows` table
- Creates `executions` table
- Creates `execution_logs` table
- Creates `chat_sessions` table
- Creates `chat_messages` table
- All necessary indexes and foreign keys

## Creating New Migrations

### Method 1: Manual Creation

1. **Create migration files** in `backend/migrations/`:

```bash
# Example: Adding a new field to workflows table
touch backend/migrations/000002_add_workflow_tags.up.sql
touch backend/migrations/000002_add_workflow_tags.down.sql
```

2. **Write the up migration** (`000002_add_workflow_tags.up.sql`):

```sql
ALTER TABLE workflows ADD COLUMN tags TEXT;
CREATE INDEX idx_workflows_tags ON workflows(tags);
```

3. **Write the down migration** (`000002_add_workflow_tags.down.sql`):

```sql
DROP INDEX IF EXISTS idx_workflows_tags;
ALTER TABLE workflows DROP COLUMN tags;
```

### Method 2: Using migrate CLI

```bash
# Install migrate CLI globally
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Create new migration
migrate create -ext sql -dir backend/migrations -seq add_workflow_tags
```

## Migration Commands

### Using Makefile (Recommended)

```bash
# Apply all pending migrations
make migrate-up

# Rollback the last migration
make migrate-down

# Show current migration version
make migrate-version

# Force set migration version (use with caution)
go run cmd/migrate/main.go force [version]
```

### Using the migrate tool directly

```bash
cd backend

# Apply all migrations
go run cmd/migrate/main.go up

# Rollback last migration
go run cmd/migrate/main.go down

# Check version
go run cmd/migrate/main.go version

# Force version (if migrations are stuck)
go run cmd/migrate/main.go force 1
```

### Docker

```bash
# Run migrations in Docker
docker-compose run --rm migrate ./migrate up

# Check version
docker-compose run --rm migrate ./migrate version

# Rollback
docker-compose run --rm migrate ./migrate down
```

## Migration States

### Normal State
```
Current version: 1
```
Database is at version 1, ready for next migration.

### Dirty State
```
Current version: 1 (dirty)
```
A migration failed partway through. You need to:

1. Fix the migration file
2. Force the version:
```bash
go run cmd/migrate/main.go force 1
```
3. Try again:
```bash
go run cmd/migrate/main.go up
```

## Environment Configuration

Migrations use the same environment variables as the application:

### PostgreSQL
```env
DB_DRIVER=postgres
DB_HOST=postgres
DB_PORT=5432
DB_USER=workflow
DB_PASSWORD=workflow_password
DB_NAME=workflows
DB_SSLMODE=disable
```

### SQLite
```env
DB_DRIVER=sqlite
DATABASE_PATH=./data/workflows.db
```

## Docker Compose Integration

The migration service in `docker-compose.yml`:

```yaml
migrate:
  build:
    context: ./backend
    dockerfile: Dockerfile
  environment:
    - DB_DRIVER=postgres
    - DB_HOST=postgres
    # ... other env vars
  depends_on:
    postgres:
      condition: service_healthy
  command: ["./migrate", "up"]
  restart: on-failure
```

**Features:**
- Runs automatically before backend starts
- Waits for PostgreSQL to be healthy
- Restarts on failure (handles transient connection issues)
- Uses same Docker image as backend

## Best Practices

### 1. Always Write Down Migrations
Every `up` migration must have a corresponding `down` migration for rollback capability.

### 2. Test Migrations
Test both up and down migrations:
```bash
make migrate-up
make migrate-down
make migrate-up
```

### 3. Make Migrations Idempotent
Use `IF EXISTS` and `IF NOT EXISTS`:
```sql
CREATE TABLE IF NOT EXISTS users (...);
DROP TABLE IF EXISTS users;
```

### 4. One Change Per Migration
Don't combine unrelated changes. Keep migrations focused.

### 5. Never Edit Applied Migrations
Once a migration is applied in production, never edit it. Create a new migration instead.

### 6. Version Control
Always commit migration files to git.

## PostgreSQL vs SQLite Differences

### PostgreSQL
- Supports `SERIAL` for auto-increment
- Full transaction support for DDL
- Better concurrent migration handling

### SQLite
- Use `INTEGER PRIMARY KEY AUTOINCREMENT`
- Limited ALTER TABLE support
- May require table recreation for some changes

**Example SQLite migration for adding a column:**
```sql
-- Works in both PostgreSQL and SQLite
ALTER TABLE workflows ADD COLUMN tags TEXT;
```

**Example that needs special handling:**
```sql
-- PostgreSQL
ALTER TABLE workflows DROP COLUMN tags;

-- SQLite (requires recreation)
-- Create new table
-- Copy data
-- Drop old table
-- Rename new table
```

## Troubleshooting

### Migration Version Mismatch

**Problem**: Database shows wrong version

**Solution**:
```bash
# Check version
make migrate-version

# Force correct version
go run cmd/migrate/main.go force [correct_version]
```

### Dirty Migration State

**Problem**: Migration failed partway

**Solution**:
1. Check what failed:
```bash
make migrate-version  # Shows "dirty"
```

2. Manually fix the database if needed

3. Force to last good version:
```bash
go run cmd/migrate/main.go force [last_good_version]
```

4. Try migration again:
```bash
make migrate-up
```

### No Migrations to Apply

**Message**: "no change"

This is normal - all migrations are already applied.

### Connection Refused

**Problem**: Can't connect to database

**Solution**:
- Check database is running: `docker-compose ps`
- Verify environment variables
- Check database health: `docker-compose logs postgres`

### Permission Denied

**Problem**: Can't create tables

**Solution**:
- Check database user permissions
- Ensure user has CREATE/ALTER privileges

## Rolling Back in Production

### Planned Rollback
```bash
# Rollback last migration
docker-compose run --rm migrate ./migrate down
```

### Emergency Rollback
If the migration broke something:

1. Stop the application:
```bash
docker-compose stop backend
```

2. Rollback migration:
```bash
docker-compose run --rm migrate ./migrate down
```

3. Restart:
```bash
docker-compose start backend
```

## Advanced Usage

### Skip to Specific Version

```bash
# Migrate to version 3
go run cmd/migrate/main.go up
# (will stop at version 3 if that's the latest)
```

### Force Dirty Recovery

If migration is stuck in dirty state:

```bash
# Check current state
go run cmd/migrate/main.go version

# Force to version before dirty state
go run cmd/migrate/main.go force [version-1]

# Apply migrations again
go run cmd/migrate/main.go up
```

### Separate Migration from Application

Run migrations separately from application startup:

```bash
# 1. Run migrations
docker-compose up migrate

# 2. Start application
docker-compose up backend frontend
```

## CI/CD Integration

### Example GitHub Actions

```yaml
- name: Run migrations
  env:
    DB_DRIVER: postgres
    DB_HOST: localhost
    DB_USER: postgres
    DB_PASSWORD: postgres
    DB_NAME: test_db
  run: |
    cd backend
    go run cmd/migrate/main.go up
```

### Example GitLab CI

```yaml
migrate:
  script:
    - cd backend
    - go run cmd/migrate/main.go up
```

## Monitoring Migrations

### Check Migration Status

```bash
# In production
docker exec workflow-migrate ./migrate version

# Or via API health endpoint (if implemented)
curl http://localhost:3001/health
```

### Log Migration History

Migrations are logged to stdout:
```bash
docker-compose logs migrate
```

## Development Workflow

1. **Create feature branch**
```bash
git checkout -b feature/add-tags
```

2. **Create migration**
```bash
touch backend/migrations/000002_add_tags.up.sql
touch backend/migrations/000002_add_tags.down.sql
```

3. **Write migrations**
```sql
-- up
ALTER TABLE workflows ADD COLUMN tags TEXT;

-- down
ALTER TABLE workflows DROP COLUMN tags;
```

4. **Test locally**
```bash
make migrate-up
make migrate-down
make migrate-up
```

5. **Update GORM models** (if needed)

6. **Commit and push**
```bash
git add backend/migrations/
git commit -m "Add tags to workflows"
git push
```

7. **Deployment**
- Docker Compose automatically runs migrations
- Or run manually: `docker-compose run --rm migrate ./migrate up`

## Additional Resources

- [golang-migrate documentation](https://github.com/golang-migrate/migrate)
- [Migration best practices](https://github.com/golang-migrate/migrate/blob/master/MIGRATIONS.md)
- Our [DATABASE.md](DATABASE.md) for database configuration
