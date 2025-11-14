# Database Configuration Guide

This application uses GORM as the ORM and supports multiple database backends. You can easily switch between PostgreSQL and SQLite.

## Supported Databases

- **PostgreSQL** (Recommended for production)
- **SQLite** (Good for development and testing)

## Quick Start

### PostgreSQL (Default)

1. **Set environment variables in `.env`**:
```env
DB_DRIVER=postgres
DB_HOST=postgres
DB_PORT=5432
DB_USER=workflow
DB_PASSWORD=workflow_password
DB_NAME=workflows
DB_SSLMODE=disable
```

2. **Start with Docker Compose**:
```bash
docker-compose up -d
```

This will start:
- PostgreSQL database container
- Backend connected to PostgreSQL
- Frontend

### SQLite (Alternative)

1. **Set environment variables in `.env`**:
```env
DB_DRIVER=sqlite
DATABASE_PATH=./data/workflows.db
```

2. **Start with Docker Compose**:
```bash
docker-compose -f docker-compose.sqlite.yml up -d
```

Or for local development:
```bash
cd backend
DB_DRIVER=sqlite DATABASE_PATH=./data/workflows.db make dev
```

## Configuration Options

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `DB_DRIVER` | Database driver: `postgres` or `sqlite` | `sqlite` | No |
| `DB_HOST` | PostgreSQL host | `localhost` | Yes (for Postgres) |
| `DB_PORT` | PostgreSQL port | `5432` | Yes (for Postgres) |
| `DB_USER` | PostgreSQL user | - | Yes (for Postgres) |
| `DB_PASSWORD` | PostgreSQL password | - | Yes (for Postgres) |
| `DB_NAME` | PostgreSQL database name | `workflows` | Yes (for Postgres) |
| `DB_SSLMODE` | PostgreSQL SSL mode | `disable` | No |
| `DATABASE_PATH` | SQLite file path | `./data/workflows.db` | Yes (for SQLite) |

## GORM Features

### Automatic Migrations

The application automatically creates and updates database tables on startup using GORM's AutoMigrate feature. This means:
- No manual SQL scripts needed
- Schema updates are handled automatically
- Works across both PostgreSQL and SQLite

### Database Tables

The following tables are created automatically:

1. **workflows** - Workflow definitions
   - id (primary key)
   - name, description
   - nodes (JSON)
   - edges (JSON)
   - active (boolean)
   - created_at, updated_at

2. **executions** - Workflow execution records
   - id (primary key)
   - workflow_id (foreign key, indexed)
   - status, started_at, finished_at
   - error, context (JSON)

3. **execution_logs** - Node execution logs
   - id (auto-increment)
   - execution_id (foreign key, indexed)
   - node_id, status
   - output, error
   - executed_at

4. **chat_sessions** - Chat session records
   - id (primary key)
   - workflow_id (foreign key, indexed)
   - created_at

5. **chat_messages** - Chat message history
   - id (primary key)
   - session_id (foreign key, indexed)
   - role, content
   - timestamp

## Switching Databases

### From SQLite to PostgreSQL

1. **Export data from SQLite** (if you have existing data):
```bash
# This is manual - you'll need to export/import data
# Or start fresh with PostgreSQL
```

2. **Update `.env`**:
```env
DB_DRIVER=postgres
DB_HOST=postgres
DB_PORT=5432
DB_USER=workflow
DB_PASSWORD=workflow_password
DB_NAME=workflows
DB_SSLMODE=disable
```

3. **Restart services**:
```bash
docker-compose down
docker-compose up -d
```

### From PostgreSQL to SQLite

1. **Update `.env`**:
```env
DB_DRIVER=sqlite
DATABASE_PATH=./data/workflows.db
```

2. **Restart services**:
```bash
docker-compose down
docker-compose -f docker-compose.sqlite.yml up -d
```

## Production Recommendations

### PostgreSQL Configuration

For production, consider:

1. **Connection Pooling**:
GORM handles connection pooling automatically, but you can configure it:
```go
sqlDB, err := db.DB()
sqlDB.SetMaxIdleConns(10)
sqlDB.SetMaxOpenConns(100)
sqlDB.SetConnMaxLifetime(time.Hour)
```

2. **SSL/TLS**:
```env
DB_SSLMODE=require
```

3. **Separate Database Host**:
Use a managed PostgreSQL service (AWS RDS, Google Cloud SQL, etc.)

4. **Backups**:
```bash
# PostgreSQL backup
docker exec workflow-postgres pg_dump -U workflow workflows > backup.sql

# Restore
docker exec -i workflow-postgres psql -U workflow workflows < backup.sql
```

### SQLite Configuration

For SQLite in production:

1. **WAL Mode** (better concurrency):
SQLite driver enables WAL mode by default with GORM.

2. **Regular Backups**:
```bash
# Backup
cp backend/data/workflows.db backups/workflows-$(date +%Y%m%d).db

# Or from Docker volume
docker run --rm -v workflow-orchestration_workflow-data:/data \
  -v $(pwd)/backups:/backup alpine \
  cp /data/workflows.db /backup/backup-$(date +%Y%m%d).db
```

## Database Monitoring

### PostgreSQL

```bash
# Connect to database
docker exec -it workflow-postgres psql -U workflow -d workflows

# List tables
\dt

# Describe table
\d workflows

# Query workflows
SELECT id, name, active, created_at FROM workflows;

# Check database size
SELECT pg_size_pretty(pg_database_size('workflows'));
```

### SQLite

```bash
# Connect to database (local)
sqlite3 backend/data/workflows.db

# Or in Docker
docker exec -it workflow-backend sqlite3 /root/data/workflows.db

# List tables
.tables

# Describe table
.schema workflows

# Query workflows
SELECT id, name, active, created_at FROM workflows;

# Database size
.dbinfo
```

## Performance Considerations

### PostgreSQL
- **Pros**: Better for concurrent writes, horizontal scaling, advanced features
- **Cons**: Requires separate service, more complex setup
- **Best for**: Production, multiple users, high concurrency

### SQLite
- **Pros**: Zero configuration, single file, fast for reads, easy backups
- **Cons**: Limited concurrent writes, no network access
- **Best for**: Development, small deployments, embedded systems

## Troubleshooting

### PostgreSQL Connection Issues

```bash
# Check if PostgreSQL is running
docker-compose ps postgres

# View PostgreSQL logs
docker-compose logs postgres

# Test connection
docker exec workflow-backend nc -zv postgres 5432

# Check credentials
docker exec -it workflow-postgres psql -U workflow -d workflows
```

### SQLite Issues

```bash
# Check if file exists
ls -la backend/data/workflows.db

# Check permissions
docker exec workflow-backend ls -la /root/data/

# Test database
docker exec workflow-backend sqlite3 /root/data/workflows.db "SELECT 1;"
```

### Migration Issues

If migrations fail:

1. **Check logs**:
```bash
docker-compose logs backend
```

2. **Manual migration**:
Connect to database and check what went wrong.

3. **Reset database** (CAUTION: deletes all data):
```bash
# PostgreSQL
docker-compose down -v
docker-compose up -d

# SQLite
rm backend/data/workflows.db
make dev
```

## Adding Support for Other Databases

GORM supports many databases. To add support for MySQL, for example:

1. **Add driver to `go.mod`**:
```go
require gorm.io/driver/mysql v1.5.2
```

2. **Update `database/database.go`**:
```go
import "gorm.io/driver/mysql"

case "mysql":
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
        config.User, config.Password, config.Host, config.Port, config.DBName)
    dialector = mysql.Open(dsn)
```

3. **Add configuration**:
```env
DB_DRIVER=mysql
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=password
DB_NAME=workflows
```

## Schema Versioning

GORM AutoMigrate handles schema changes automatically, but for complex migrations:

1. **Use GORM Migrator**:
```go
db.Migrator().AddColumn(&Workflow{}, "new_field")
db.Migrator().RenameColumn(&Workflow{}, "old_name", "new_name")
```

2. **Consider using a migration tool**:
- [golang-migrate](https://github.com/golang-migrate/migrate)
- [goose](https://github.com/pressly/goose)

## Best Practices

1. **Always use environment variables** for database configuration
2. **Never commit credentials** to version control
3. **Use connection pooling** for PostgreSQL in production
4. **Enable SSL/TLS** for PostgreSQL in production
5. **Regular backups** are essential
6. **Monitor database size** and performance
7. **Use indexes** for frequently queried fields (GORM handles this)
8. **Test migrations** on staging before production
