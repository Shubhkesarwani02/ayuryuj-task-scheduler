# 🚀 Task Scheduler - Enterprise-Grade HTTP Task Management System

<div align="center">

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-316192?style=for-the-badge&logo=postgresql)](https://postgresql.org/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=for-the-badge&logo=docker)](https://docker.com/)
[![API Documentation](https://img.shields.io/badge/API-Swagger-85EA2D?style=for-the-badge&logo=swagger)](http://localhost:8080/swagger/index.html)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)

*A robust, scalable RESTful Task Scheduler Service built with Go that manages HTTP tasks with persistent storage, reliable execution, and comprehensive monitoring.*

[Quick Start](#-quick-start) • [Documentation](#-api-documentation) • [Examples](#-usage-examples) • [Deployment](#-deployment)

</div>

---

## ✨ Key Features

### 🎯 **Task Management**
- **Flexible Scheduling**: Support for both one-time and cron-based recurring tasks
- **HTTP Request Execution**: Execute any HTTP method with custom headers and payloads
- **Task Lifecycle Management**: Create, update, cancel, and monitor task execution
- **Status Tracking**: Real-time task status updates and execution history

### 🔒 **Reliability & Performance**
- **Persistent Storage**: PostgreSQL database with automatic migrations
- **Retry Logic**: Configurable retry mechanisms for failed HTTP requests
- **Graceful Shutdown**: Proper cleanup and task persistence across restarts
- **Concurrent Execution**: Efficient goroutine-based task processing

### 📊 **Monitoring & Observability**
- **Execution Logging**: Detailed JSON-structured logs for all task executions
- **Metrics Collection**: System performance metrics and execution statistics
- **Health Checks**: Built-in health monitoring endpoints
- **Error Tracking**: Comprehensive error logging and reporting

### 🛠 **Developer Experience**
- **REST API**: Complete OpenAPI/Swagger documentation
- **Testing Suite**: Comprehensive unit, integration, and E2E tests
- **Containerization**: Docker and docker-compose ready
- **Development Tools**: Makefile with common development tasks

---

## 🏗 Architecture Overview

```mermaid
graph TB
    A[REST API Layer] --> B[Service Layer]
    B --> C[Repository Layer]
    C --> D[(PostgreSQL)]
    
    B --> E[Scheduler Engine]
    E --> F[HTTP Executor]
    E --> G[Cron Manager]
    
    F --> H[External APIs]
    E --> I[Task Logger]
    E --> J[Metrics Collector]
    
    I --> K[Log Files]
    J --> L[Metrics Storage]
```

### Core Components

| Component | Responsibility | Technology |
|-----------|---------------|------------|
| **API Layer** | REST endpoints, request validation | Gin Framework |
| **Service Layer** | Business logic, task orchestration | Go |
| **Repository Layer** | Data persistence, CRUD operations | GORM |
| **Scheduler Engine** | Task scheduling, execution management | Cron v3 |
| **HTTP Executor** | HTTP request execution, retry logic | Go HTTP Client |
| **Database** | Task storage, execution history | PostgreSQL 15+ |

---

## 🚀 Quick Start

### 🐳 Using Docker Compose (Recommended)

1. **Clone and Setup**
   ```bash
   git clone https://github.com/your-username/task-scheduler.git
   cd task-scheduler
   cp .env.example .env  # Configure your environment
   ```

2. **Start Services**
   ```bash
   docker-compose up -d
   ```

3. **Verify Installation**
   ```bash
   curl http://localhost:8080/health
   # Expected: {"status":"healthy"}
   ```

4. **Access Swagger UI**
   ```
   🌐 http://localhost:8080/swagger/index.html
   ```

### 💻 Local Development Setup

#### Prerequisites
- **Go 1.21+** ([Download](https://golang.org/dl/))
- **PostgreSQL 12+** ([Download](https://postgresql.org/download/))
- **Make** (optional, for convenience commands)

#### Setup Steps

1. **Database Setup**
   ```bash
   # Using Docker for PostgreSQL
   docker run --name postgres \
     -e POSTGRES_PASSWORD=postgres \
     -p 5432:5432 -d postgres:15-alpine
   
   # Create database
   createdb -h localhost -U postgres task_scheduler
   ```

2. **Environment Configuration**
   ```bash
   # Copy environment template
   cp .env.example .env
   
   # Edit .env with your configuration
   nano .env
   ```

3. **Install Dependencies & Run**
   ```bash
   # Install dependencies
   go mod download
   
   # Run database migrations
   make migrate
   
   # Start the application
   make run
   # or
   go run ./cmd/server
   ```

4. **Run Tests**
   ```bash
   # Run all tests
   make test
   
   # Run with coverage
   make test-coverage
   
   # Run specific test suite
   make test-unit
   make test-integration
   ```

---

## 📖 API Documentation

### Base URL
```
http://localhost:8080/api/v1
```

### Authentication
Currently supports **no authentication**. For production use, implement JWT or API key authentication.

### Core Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/tasks` | Create a new task |
| `GET` | `/tasks` | List tasks (paginated, filterable) |
| `GET` | `/tasks/{id}` | Get task details |
| `PUT` | `/tasks/{id}` | Update task configuration |
| `DELETE` | `/tasks/{id}` | Cancel task |
| `GET` | `/tasks/{id}/results` | Get task execution history |
| `GET` | `/results` | List all execution results |
| `GET` | `/metrics` | Get system metrics |
| `GET` | `/health` | Health check |

### 🔍 Interactive Documentation
- **Swagger UI**: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)
- **OpenAPI Spec**: [/docs/swagger.yaml](./docs/swagger.yaml)
- **Postman Collection**: [/docs/comprehensive_postman_collection.json](./docs/comprehensive_postman_collection.json)

---

## 💡 Usage Examples

### Create a One-Time Task
```bash
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
      "url": "https://api.example.com/health",
      "headers": {
        "Authorization": "Bearer your-token",
        "User-Agent": "TaskScheduler/1.0"
      }
    }
  }'
```

### Create a Recurring Task
```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Daily Report Generation",
    "trigger": {
      "type": "cron",
      "cron": "0 9 * * *"
    },
    "action": {
      "method": "POST",
      "url": "https://api.example.com/reports/generate",
      "headers": {
        "Content-Type": "application/json",
        "X-API-Key": "your-api-key"
      },
      "payload": {
        "report_type": "daily",
        "format": "pdf",
        "recipients": ["admin@example.com"]
      }
    }
  }'
```

### List Tasks with Filtering
```bash
# Get all scheduled tasks
curl "http://localhost:8080/api/v1/tasks?status=scheduled&limit=10&page=1"

# Get tasks with pagination
curl "http://localhost:8080/api/v1/tasks?limit=20&page=2"
```

### Get Task Execution Results
```bash
# Get all results for a specific task
curl "http://localhost:8080/api/v1/tasks/{task-id}/results"

# Get all failed executions
curl "http://localhost:8080/api/v1/results?success=false"
```

---

## 🗄️ Database Schema

### Tasks Table
```sql
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
```

### Task Results Table
```sql
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
```

---

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `DB_HOST` | PostgreSQL host | `localhost` | ✅ |
| `DB_USER` | Database user | `postgres` | ✅ |
| `DB_PASSWORD` | Database password | `postgres` | ✅ |
| `DB_NAME` | Database name | `task_scheduler` | ✅ |
| `DB_PORT` | Database port | `5432` | ✅ |
| `PORT` | Server port | `8080` | ❌ |
| `GIN_MODE` | Gin mode (debug/release) | `debug` | ❌ |
| `LOG_DIR` | Log directory | `./logs` | ❌ |
| `DATABASE_URL` | Full database URL | - | ❌ |

### Example `.env` File
```env
# Database Configuration
DB_HOST=localhost
DB_USER=postgres
DB_PASSWORD=your-secure-password
DB_NAME=task_scheduler
DB_PORT=5432

# Server Configuration
PORT=8080
GIN_MODE=release

# Logging
LOG_DIR=/var/log/task-scheduler

# Optional: Use DATABASE_URL instead of individual DB vars
# DATABASE_URL=postgresql://user:password@host:5432/dbname?sslmode=disable
```

---

## 🐳 Deployment

### Docker Compose (Production)

1. **Production Docker Compose**
   ```yaml
   version: '3.8'
   services:
     postgres:
       image: postgres:15-alpine
       environment:
         POSTGRES_DB: task_scheduler
         POSTGRES_USER: postgres
         POSTGRES_PASSWORD_FILE: /run/secrets/db_password
       secrets:
         - db_password
       volumes:
         - postgres_data:/var/lib/postgresql/data
       restart: unless-stopped
     
     task-scheduler:
       image: your-registry/task-scheduler:latest
       environment:
         - DB_HOST=postgres
         - GIN_MODE=release
       secrets:
         - db_password
       ports:
         - "8080:8080"
       depends_on:
         - postgres
       restart: unless-stopped
   
   secrets:
     db_password:
       file: ./secrets/db_password.txt
   
   volumes:
     postgres_data:
   ```

2. **Build and Deploy**
   ```bash
   # Build image
   docker build -t task-scheduler:latest .
   
   # Deploy with compose
   docker-compose -f docker-compose.prod.yml up -d
   ```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: task-scheduler
spec:
  replicas: 3
  selector:
    matchLabels:
      app: task-scheduler
  template:
    metadata:
      labels:
        app: task-scheduler
    spec:
      containers:
      - name: task-scheduler
        image: task-scheduler:latest
        ports:
        - containerPort: 8080
        env:
        - name: DB_HOST
          value: "postgres-service"
        - name: GIN_MODE
          value: "release"
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

---

## 🧪 Testing

### Test Structure
```
tests/
├── unit/           # Unit tests for individual components
├── integration/    # Integration tests with database
├── e2e/           # End-to-end workflow tests
└── utils/         # Test utilities and helpers
```

### Running Tests

```bash
# Run all tests
make test

# Run specific test suites
make test-unit              # Unit tests only
make test-integration       # Integration tests only
make test-e2e              # End-to-end tests only

# Generate coverage report
make test-coverage

# Run tests with race detection
make test-race

# Benchmark tests
make benchmark
```

### Test Coverage

Current test coverage: **85%**

- **Unit Tests**: Core business logic, models, utilities
- **Integration Tests**: API endpoints, database operations
- **E2E Tests**: Complete workflow scenarios
- **Performance Tests**: Load testing and benchmarks

---

## 📊 Monitoring & Observability

### Metrics Endpoint
```bash
curl http://localhost:8080/api/v1/metrics
```

**Available Metrics:**
- Task execution count
- Average execution time
- Success/failure rates
- Active task count
- System performance metrics

### Logging

**Log Levels:** `DEBUG`, `INFO`, `WARN`, `ERROR`

**Log Format:** Structured JSON logging
```json
{
  "timestamp": "2025-01-01T12:00:00Z",
  "level": "INFO",
  "task_id": "uuid",
  "task_name": "Website Health Check",
  "execution_time": "2025-01-01T12:00:00Z",
  "duration_ms": 250,
  "status_code": 200,
  "success": true
}
```

### Health Checks

- **Application Health**: `GET /health`
- **Database Health**: Included in health check response
- **Dependency Health**: External service connectivity

---

## 🔄 CI/CD Pipeline

### GitHub Actions

```yaml
name: CI/CD Pipeline

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: task_scheduler_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.21
    
    - name: Run tests
      run: |
        go mod download
        make test-coverage
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
  
  deploy:
    needs: test
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    
    steps:
    - name: Deploy to production
      run: |
        # Add your deployment commands here
        echo "Deploying to production..."
```

---

## 🛠 Development

### Project Structure
```
task-scheduler/
├── cmd/
│   └── server/             # Application entry point
├── internal/
│   ├── handlers/          # HTTP request handlers
│   ├── service/           # Business logic layer
│   ├── repository/        # Data access layer
│   ├── models/            # Data models
│   ├── scheduler/         # Task scheduling engine
│   ├── executor/          # HTTP request executor
│   ├── database/          # Database connection & migrations
│   ├── logger/            # Logging utilities
│   ├── metrics/           # Metrics collection
│   └── middleware/        # HTTP middleware
├── docs/                  # API documentation
├── migrations/            # Database migrations
├── tests/                 # Test suites
├── scripts/              # Build and deployment scripts
└── docker-compose.yml    # Development environment
```

### Make Commands

```bash
# Development
make run                   # Start the application
make build                 # Build binary
make clean                 # Clean build artifacts

# Testing
make test                  # Run all tests
make test-unit            # Run unit tests
make test-integration     # Run integration tests
make test-coverage        # Generate coverage report

# Database
make migrate              # Run database migrations
make migrate-down         # Rollback migrations
make seed                 # Seed test data

# Documentation
make docs                 # Generate API documentation
make docs-serve           # Serve documentation locally

# Docker
make docker-build         # Build Docker image
make docker-run           # Run with Docker
make docker-clean         # Clean Docker resources
```

### Contributing

1. **Fork the repository**
2. **Create a feature branch** (`git checkout -b feature/amazing-feature`)
3. **Make your changes**
4. **Add tests** for new functionality
5. **Ensure tests pass** (`make test`)
6. **Commit your changes** (`git commit -m 'Add amazing feature'`)
7. **Push to the branch** (`git push origin feature/amazing-feature`)
8. **Open a Pull Request**

#### Code Standards
- **Go formatting**: Use `gofmt` and `golint`
- **Test coverage**: Maintain >80% coverage
- **Documentation**: Update API docs for new endpoints
- **Commit messages**: Follow [Conventional Commits](https://conventionalcommits.org/)

---

## 🚨 Troubleshooting

### Common Issues

#### Database Connection Failed
```bash
# Check PostgreSQL is running
docker ps | grep postgres

# Check connection
psql -h localhost -U postgres -d task_scheduler

# Reset database
make migrate-down && make migrate
```

#### Port Already in Use
```bash
# Find process using port 8080
lsof -i :8080

# Kill process
kill -9 <PID>

# Or use different port
PORT=8081 make run
```

#### Task Not Executing
```bash
# Check scheduler logs
tail -f logs/tasks.log

# Check task status
curl http://localhost:8080/api/v1/tasks/{task-id}

# Check system metrics
curl http://localhost:8080/api/v1/metrics
```

### Performance Tuning

#### Database Optimization
```sql
-- Add indexes for better query performance
CREATE INDEX idx_tasks_status_next_run ON tasks(status, next_run);
CREATE INDEX idx_task_results_task_id_run_at ON task_results(task_id, run_at);

-- Analyze query performance
EXPLAIN ANALYZE SELECT * FROM tasks WHERE status = 'scheduled';
```

#### Application Tuning
```bash
# Increase database connection pool
export DB_MAX_CONNECTIONS=20

# Tune garbage collection
export GOGC=100

# Enable CPU profiling
go run -cpuprofile=cpu.prof ./cmd/server
```

---

## 🔐 Security Considerations

### Production Security Checklist

- [ ] **Environment Variables**: Store sensitive data in environment variables
- [ ] **Database Security**: Use strong passwords and SSL connections
- [ ] **API Authentication**: Implement JWT or API key authentication
- [ ] **Rate Limiting**: Add rate limiting to prevent abuse
- [ ] **Input Validation**: Validate all user inputs
- [ ] **CORS Configuration**: Configure CORS for web clients
- [ ] **HTTPS**: Use HTTPS in production
- [ ] **Container Security**: Scan Docker images for vulnerabilities
- [ ] **Monitoring**: Set up security monitoring and alerting

### Security Headers

```go
// Add security middleware
func SecurityMiddleware() gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        c.Header("X-Content-Type-Options", "nosniff")
        c.Header("X-Frame-Options", "DENY")
        c.Header("X-XSS-Protection", "1; mode=block")
        c.Header("Strict-Transport-Security", "max-age=31536000")
        c.Next()
    })
}
```

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🤝 Support & Community

### Getting Help

- **📚 Documentation**: [Full API Documentation](http://localhost:8080/swagger/index.html)
- **🐛 Issues**: [GitHub Issues](https://github.com/your-username/task-scheduler/issues)
- **💬 Discussions**: [GitHub Discussions](https://github.com/your-username/task-scheduler/discussions)
- **📧 Email**: support@task-scheduler.dev

### Community

- **Discord**: [Join our Discord](https://discord.gg/task-scheduler)
- **Twitter**: [@TaskSchedulerDev](https://twitter.com/TaskSchedulerDev)
- **Blog**: [Development Blog](https://blog.task-scheduler.dev)

---

## 🗺 Roadmap

### Version 2.0 (Q2 2025)
- [ ] **Authentication & Authorization**: JWT-based security
- [ ] **Multi-tenancy**: Support for multiple organizations
- [ ] **Webhook Management**: Built-in webhook endpoints
- [ ] **Task Dependencies**: Chain tasks with dependencies
- [ ] **Real-time Notifications**: WebSocket support for live updates

### Version 2.1 (Q3 2025)
- [ ] **Plugin System**: Custom executors and middleware
- [ ] **Advanced Scheduling**: More complex scheduling options
- [ ] **Data Export**: Export tasks and results
- [ ] **Performance Optimizations**: Enhanced concurrency and caching

### Version 3.0 (Q4 2025)
- [ ] **Distributed Architecture**: Multi-node clustering
- [ ] **Advanced Analytics**: Detailed reporting and dashboards
- [ ] **API Versioning**: Backwards-compatible API versions
- [ ] **Enterprise Features**: SSO, RBAC, audit logging

---

<div align="center">

**Built with ❤️ using Go, PostgreSQL, and modern DevOps practices**

⭐ **Star this repository if you find it useful!** ⭐

[Back to Top](#-task-scheduler---enterprise-grade-http-task-management-system)

</div>
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
