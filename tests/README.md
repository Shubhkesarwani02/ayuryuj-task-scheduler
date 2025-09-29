# Testing Configuration

## Overview
This document describes the comprehensive test suite for the Task Scheduler application.

## Test Structure

```
tests/
├── utils/                     # Test utilities and helpers
│   ├── test_containers.go     # PostgreSQL test containers setup
│   ├── mock_server.go         # HTTP mock server for task execution
│   ├── test_factory.go        # Test data factories
│   └── test_helpers.go        # Common test helper functions
├── unit/                      # Unit tests (isolated component testing)
│   ├── models/                # Model validation and serialization tests
│   ├── repository/            # Database layer tests
│   └── executor/              # HTTP executor tests
├── integration/               # Integration tests (component interaction)
│   └── api/                   # REST API endpoint tests
└── e2e/                       # End-to-end tests (full workflow)
    └── workflow_test.go       # Complete task lifecycle tests
```

## Test Categories

### 1. Unit Tests (`tests/unit/`)

**Purpose**: Test individual components in isolation

#### Models Tests (`tests/unit/models/`)
- Task model validation
- Headers serialization/deserialization
- Request/Response DTO validation
- Enum constants and types

#### Repository Tests (`tests/unit/repository/`)
- Task CRUD operations
- Result CRUD operations
- Database queries and filtering
- Pagination and sorting
- Concurrent access handling

#### Executor Tests (`tests/unit/executor/`)
- HTTP request execution
- Error handling (network, timeout, HTTP errors)
- Request/Response processing
- Headers and payload handling
- Concurrent execution

### 2. Integration Tests (`tests/integration/`)

**Purpose**: Test component interactions with real dependencies

#### API Tests (`tests/integration/api/`)
- Task creation, update, deletion
- Input validation and error handling
- Task listing and filtering
- Task result querying
- Complex payload handling
- Concurrent API operations

### 3. End-to-End Tests (`tests/e2e/`)

**Purpose**: Test complete workflows from start to finish

#### Workflow Tests
- One-off task complete lifecycle
- Cron task multiple executions
- Error handling and recovery
- Task cancellation workflow
- Complex payload processing

## Test Infrastructure

### Test Containers
- **PostgreSQL**: Real PostgreSQL database in Docker container
- **Automatic cleanup**: Containers are terminated after tests
- **Isolation**: Each test suite gets a fresh database

### Mock HTTP Server
- **Request capture**: All HTTP requests are captured for verification
- **Response configuration**: Configurable responses (success, error, timeout)
- **Concurrent handling**: Thread-safe request/response handling

### Test Data Factories
- **Task Factory**: Creates various types of test tasks
- **Result Factory**: Creates test task execution results
- **Scenario Factory**: Pre-configured test scenarios

## Running Tests

### Prerequisites
```bash
# Install dependencies
go mod tidy

# Ensure Docker is running (for test containers)
docker --version
```

### All Tests
```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run with race detection
make test-race
```

### By Category
```bash
# Unit tests only
make test-unit

# Integration tests only
make test-integration

# End-to-end tests only
make test-e2e
```

### By Component
```bash
# Model tests
make test-models

# Repository tests
make test-repository

# Executor tests
make test-executor

# API tests
make test-api

# Workflow tests
make test-workflow
```

### Quick Tests
```bash
# Short tests (faster execution)
make test-short

# Parallel execution
make test-parallel
```

## Test Configuration

### Environment Variables
Tests automatically handle environment setup. No manual configuration required.

### Test Database
- Uses testcontainers to spin up PostgreSQL
- Each test suite gets an isolated database
- Automatic migration execution
- Cleanup after test completion

### Timeouts
- Unit tests: 2-5 minutes
- Integration tests: 5-10 minutes
- E2E tests: 10 minutes
- Individual test cases: 30 seconds to 2 minutes

## Test Coverage Goals

| Component | Coverage Target | Current Status |
|-----------|----------------|----------------|
| Models    | >95%           | ✅ Achieved     |
| Repository| >90%           | ✅ Achieved     |
| Executor  | >85%           | ✅ Achieved     |
| Handlers  | >90%           | ✅ Achieved     |
| Overall   | >85%           | ✅ Achieved     |

## Test Scenarios Covered

### Positive Scenarios
- ✅ Valid one-off task creation and execution
- ✅ Valid cron task creation and multiple executions
- ✅ Task listing with various filters
- ✅ Task updates and modifications
- ✅ Complex JSON payload handling
- ✅ Custom HTTP headers processing
- ✅ Successful HTTP request execution

### Negative Scenarios
- ✅ Invalid input validation (URLs, cron expressions, dates)
- ✅ HTTP errors (4xx, 5xx status codes)
- ✅ Network errors (connection refused, DNS failures)
- ✅ Timeout handling
- ✅ Database constraint violations
- ✅ Concurrent access conflicts

### Edge Cases
- ✅ Past datetime validation
- ✅ Invalid cron expressions
- ✅ Large payload handling
- ✅ Special characters in headers
- ✅ Empty and null values
- ✅ UUID validation

## Performance Testing

### Load Testing Scenarios
- Concurrent task creation (10+ simultaneous)
- High-frequency task execution
- Large payload processing
- Database stress testing

### Benchmarks
```bash
# Run benchmarks
go test -bench=. -benchmem ./tests/...
```

## CI/CD Integration

### GitHub Actions (Recommended)
```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: '1.21'
      - run: make test-coverage
      - uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out
```

### Local CI Testing
```bash
# Simulate CI environment
make test-ci
```

## Debugging Tests

### Verbose Output
```bash
# Run with verbose output
go test -v ./tests/...
```

### Specific Test
```bash
# Run specific test
go test -v -run TestCreateTask ./tests/integration/api/
```

### Debug with Delve
```bash
# Debug specific test
dlv test ./tests/unit/models/ -- -test.run TestTaskCreation
```

## Best Practices

### Test Naming
- Test files: `*_test.go`
- Test functions: `Test<ComponentName><Functionality>`
- Benchmark functions: `Benchmark<ComponentName><Functionality>`

### Test Structure
1. **Arrange**: Set up test data and environment
2. **Act**: Execute the functionality being tested
3. **Assert**: Verify the expected outcomes

### Test Independence
- Each test should be independent
- Use setup/teardown for clean state
- Avoid shared state between tests

### Test Data
- Use factories for consistent test data
- Prefer explicit test data over random generation
- Include boundary and edge case values

## Troubleshooting

### Common Issues

#### Docker Permission Issues
```bash
# Fix Docker permissions (Linux/macOS)
sudo usermod -aG docker $USER
# Logout and login again
```

#### Port Conflicts
- TestContainers automatically assigns free ports
- No manual port configuration needed

#### Test Timeouts
- Increase timeout for slow environments
- Use `-timeout` flag: `go test -timeout 15m`

#### Memory Issues
```bash
# Run with more memory
GOMAXPROCS=2 go test ./tests/...
```

### Getting Help
1. Check test logs for detailed error messages
2. Run individual test suites to isolate issues
3. Verify Docker is running and accessible
4. Check Go version compatibility (requires Go 1.21+)

## Contributing

### Adding New Tests
1. Place tests in appropriate category (unit/integration/e2e)
2. Use existing test utilities and helpers
3. Follow naming conventions
4. Include both positive and negative scenarios
5. Add documentation for complex test scenarios

### Test Review Checklist
- [ ] Tests are properly categorized
- [ ] Test names are descriptive
- [ ] Both success and failure cases covered
- [ ] No hardcoded values (use factories)
- [ ] Proper cleanup and resource management
- [ ] Documentation updated if needed