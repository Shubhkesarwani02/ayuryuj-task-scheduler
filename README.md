# Task Scheduler Backend

A robust RESTful Task Scheduler Service built in Go that manages HTTP tasks with persistent storage, reliable execution, and comprehensive monitoring.

## Features

- **Task Management**: Create, update, list, and cancel HTTP tasks
- **Flexible Scheduling**: Support for both one-off and cron-based recurring tasks
- **Persistent Storage**: PostgreSQL database with automatic migrations
- **Reliable Execution**: HTTP request execution with retry logic and timeout handling
- **Comprehensive Logging**: Detailed execution logs and metrics collection
- **REST API**: Full OpenAPI/Swagger documentation
- **Containerized**: Docker and docker-compose ready
- **Graceful Shutdown**: Proper cleanup and task persistence across restarts

## Quick Start

### Using Docker Compose (Recommended)

1. **Clone and setup**:
   \`\`\`bash
   git clone <repository-url>
   cd task-scheduler
   cp .env.example .env
   \`\`\`

2. **Start the services**:
   \`\`\`bash
   docker-compose up -d
   \`\`\`

3. **Verify the setup**:
   \`\`\`bash
   curl http://localhost:8080/health
   \`\`\`

4. **Access Swagger UI**:
   Open http://localhost:8080/swagger/index.html in your browser

### Local Development

1. **Prerequisites**:
   - Go 1.21+
   - PostgreSQL 12+
   - Make (optional)

2. **Setup database**:
   \`\`\`bash
   # Start PostgreSQL (or use Docker)
   docker run --name postgres -e POSTGRES_PASSWORD=postgres -p 5432:5432 -d postgres:15-alpine
   
   # Create database
   createdb -h localhost -U postgres task_scheduler
   
   # Run migrations
   make migrate
   \`\`\`

3. **Run the application**:
   \`\`\`bash
   make run
   # or
   go run ./cmd/server
   \`\`\`

## API Documentation

### Base URL
\`\`\`
http://localhost:8080/api/v1
\`\`\`

### Core Endpoints

#### Tasks
- `POST /tasks` - Create a new task
- `GET /tasks` - List tasks (with pagination and filtering)
- `GET /tasks/{id}` - Get task details
- `PUT /tasks/{id}` - Update task
- `DELETE /tasks/{id}` - Cancel task
- `GET /tasks/{id}/results` - Get task execution history

#### Results
- `GET /results` - List all execution results (with filtering)

#### System
- `GET /health` - Health check
- `GET /metrics` - System metrics
- `GET /swagger/*` - API documentation

### Example Usage

#### Create a One-off Task
\`\`\`bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Website Health Check",
    "trigger": {
      "type": "one-off",
      "datetime": "2025-01-01T12:00:00Z"
    },
    "action": {
      "method": "GET",
      "url": "https://httpbin.org/get",
      "headers": {
        "User-Agent": "TaskScheduler/1.0"
      }
    }
  }'
\`\`\`

#### Create a Recurring Task
\`\`\`bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Daily Report",
    "trigger": {
      "type": "cron",
      "cron": "0 9 * * *"
    },
    "action": {
      "method": "POST",
      "url": "https://httpbin.org/post",
      "headers": {
        "Content-Type": "application/json"
      },
      "payload": {
        "report_type": "daily",
        "timestamp": "2025-01-01T09:00:00Z"
      }
    }
  }'
\`\`\`

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_USER` | PostgreSQL user | `postgres` |
| `DB_PASSWORD` | PostgreSQL password | `postgres` |
| `DB_NAME` | PostgreSQL database | `task_scheduler` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `PORT` | Server port | `8080` |
| `GIN_MODE` | Gin mode (debug/release) | `debug` |

### Cron Expression Format

The service uses standard cron expressions with 5 fields:
\`\`\`
* * * * *
│ │ │ │ │
│ │ │ │ └─── Day of week (0-7, Sunday = 0 or 7)
│ │ │ └───── Month (1-12)
│ │ └─────── Day of month (1-31)
│ └───────── Hour (0-23)
└─────────── Minute (0-59)
\`\`\`

Examples:
- `0 9 * * *` - Every day at 9:00 AM
- `*/15 * * * *` - Every 15 minutes
- `0 0 1 * *` - First day of every month at midnight
- `0 18 * * 1-5` - Every weekday at 6:00 PM

## Database Schema

### Tasks Table
\`\`\`sql
CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    trigger_type VARCHAR(20) NOT NULL CHECK (trigger_type IN ('one-off', 'cron')),
    trigger_time TIMESTAMPTZ NULL,
    cron_expr VARCHAR(255) NULL,
    method VARCHAR(10) NOT NULL,
    url TEXT NOT NULL,
    headers JSONB,
    payload JSONB,
    status VARCHAR(20) NOT NULL DEFAULT 'scheduled',
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    next_run TIMESTAMPTZ
);
\`\`\`

### Task Results Table
\`\`\`sql
CREATE TABLE task_results (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    run_at TIMESTAMPTZ NOT NULL,
    status_code INT,
    success BOOLEAN,
    response_headers JSONB,
    response_body TEXT,
    error_message TEXT,
    duration_ms INT,
    created_at TIMESTAMPTZ DEFAULT now()
);
\`\`\`

## Architecture

\`\`\`
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   REST API      │    │   Scheduler     │    │  HTTP Executor  │
│   (Gin Router)  │────│   (Cron Jobs)   │────│  (HTTP Client)  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Repository    │    │   Database      │    │   Logging &     │
│   Layer         │────│   (PostgreSQL)  │    │   Metrics       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
\`\`\`

## Development

### Project Structure
\`\`\`
task-scheduler/
├── cmd/server/          # Application entry point
├── internal/
│   ├── database/        # Database connection and migration
│   ├── executor/        # HTTP request execution
│   ├── handlers/        # HTTP request handlers
│   ├── logger/          # Task execution logging
│   ├── metrics/         # System metrics collection
│   ├── middleware/      # HTTP middleware
│   ├── models/          # Data models and DTOs
│   ├── repository/      # Data access layer
│   ├── scheduler/       # Task scheduling engine
│   └── service/         # Business logic layer
├── migrations/          # Database migrations
├── docs/               # API documentation
├── docker-compose.yml  # Docker composition
├── Dockerfile          # Container definition
└── Makefile           # Build automation
\`\`\`

### Available Make Commands
\`\`\`bash
make build          # Build the application
make run            # Run locally
make test           # Run tests
make docker-build   # Build Docker image
make docker-run     # Start with Docker Compose
make docker-stop    # Stop Docker services
make migrate        # Run database migrations
make dev-setup      # Setup development environment
make swagger        # Generate Swagger docs
make fmt            # Format code
make lint           # Lint code
\`\`\`

### Testing with Postman

Import the provided Postman collection from `docs/postman_collection.json` to test all API endpoints with pre-configured requests.

## Monitoring and Observability

### Metrics Endpoint
Access system metrics at `GET /metrics`:
\`\`\`json
{
  "total_tasks_executed": 150,
  "successful_tasks": 142,
  "failed_tasks": 8,
  "success_rate_percent": 94.67,
  "average_execution_ms": 245,
  "tasks_per_minute": 2.5
}
\`\`\`

### Logging
- Application logs: Console output with structured logging
- Task execution logs: JSON format in `/var/log/task-scheduler/`
- Database query logs: Configurable via GORM logger

### Health Checks
- Application: `GET /health`
- Docker: Built-in health checks for both API and database containers

## Production Considerations

### Security
- Use environment variables for sensitive configuration
- Implement authentication/authorization as needed
- Configure CORS policies appropriately
- Use HTTPS in production

### Performance
- Database connection pooling is configured
- HTTP client with connection reuse
- Configurable timeouts and retry logic
- Efficient database indexing

### Scalability
- Stateless design allows horizontal scaling
- Database-backed persistence ensures consistency
- Consider using Redis for distributed locking in multi-instance deployments

### Monitoring
- Implement proper logging aggregation (ELK stack, etc.)
- Set up metrics collection (Prometheus, etc.)
- Configure alerting for failed tasks and system health

## License

MIT License - see LICENSE file for details.

## Support

For issues and questions:
1. Check the API documentation at `/swagger/`
2. Review the logs for error details
3. Open an issue in the repository
