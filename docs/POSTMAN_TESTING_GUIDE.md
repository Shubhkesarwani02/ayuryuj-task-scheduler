# Task Scheduler API - Comprehensive Testing Guide

## 📋 Overview

This guide provides comprehensive instructions for testing the Task Scheduler API using Postman with the complete test collection and environments provided.

## 🗂️ Available Files

### Collection Files
- `task-scheduler-api-complete.postman_collection.json` - Complete test suite with automated tests

### Environment Files
- `local-environment.postman_environment.json` - For local development testing
- `docker-environment.postman_environment.json` - For Docker-based testing
- `production-environment.postman_environment.json` - For production environment testing

## 🚀 Quick Start

### Step 1: Import Collection and Environment

1. **Open Postman**
2. **Import Collection:**
   - Click "Import" button
   - Select `task-scheduler-api-complete.postman_collection.json`
   - Confirm import

3. **Import Environment:**
   - Click "Import" button
   - Select the appropriate environment file:
     - `local-environment.postman_environment.json` for local testing
     - `docker-environment.postman_environment.json` for Docker testing
     - `production-environment.postman_environment.json` for production testing

### Step 2: Configure Environment

1. **Select Environment:**
   - In the top-right corner, select the imported environment from the dropdown

2. **Verify Variables:**
   - Click the eye icon next to the environment dropdown
   - Ensure `host` and `baseUrl` are correct for your setup

### Step 3: Start Your API Server

**For Local Development:**
```bash
# Start the Go server
go run cmd/server/main.go
```

**For Docker:**
```bash
# Build and run with Docker Compose
docker-compose up --build
```

## 🧪 Running Tests

### Option 1: Run Complete Test Suite

1. **Run Collection:**
   - Right-click on "Task Scheduler API - Complete Test Suite" collection
   - Select "Run collection"
   - Choose environment
   - Click "Run Task Scheduler API"

2. **Review Results:**
   - All tests should pass with green checkmarks
   - Failed tests will show in red with error details

### Option 2: Run Individual Test Categories

The collection is organized into logical categories:

#### 🔍 System Health & Info
- **Health Check** - Verifies API is running and healthy
- **API Metrics** - Tests metrics endpoint

#### 📋 Task Management - CRUD Operations
- **Create One-off Task** - Creates a one-time scheduled task
- **Create Cron Task** - Creates a recurring task with cron expression
- **Get All Tasks** - Retrieves paginated list of tasks
- **Get Specific Task** - Retrieves details of a specific task
- **Update Task** - Updates an existing task

#### ⚡ Task Execution & Control
- **Execute Task Immediately** - Triggers manual task execution
- **Pause Task** - Pauses a recurring task
- **Resume Task** - Resumes a paused task

#### 📊 Results & History
- **Get All Results** - Retrieves execution results with pagination
- **Get Task-Specific Results** - Gets results for a specific task

#### 🧪 Advanced Test Scenarios
- **Create Task with Complex Payload** - Tests complex JSON payloads
- **Test Invalid Cron Expression** - Validates error handling
- **Test Non-existent Task** - Tests 404 error responses

#### 🧹 Cleanup Test Data
- **Delete One-off Task** - Removes test data
- **Delete Cron Task** - Removes test data

### Option 3: Run Individual Requests

1. **Open any request** in the collection
2. **Click "Send"** to execute
3. **Review response** and test results in the "Test Results" tab

## 🔧 Advanced Configuration

### Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `protocol` | HTTP/HTTPS protocol | `http` or `https` |
| `host` | Server host and port | `localhost:8080` |
| `baseUrl` | Complete API base URL | `http://localhost:8080/api/v1` |
| `taskId` | Auto-populated task ID | (set by tests) |
| `cronTaskId` | Auto-populated cron task ID | (set by tests) |
| `oneOffTaskId` | Auto-populated one-off task ID | (set by tests) |
| `apiVersion` | API version | `v1` |
| `timeout` | Request timeout (ms) | `30000` |
| `testDataCleanup` | Auto-cleanup test data | `true`/`false` |

### Authentication (Production)

For production environments, you may need to configure authentication:

1. **Update Environment:**
   - Set `apiKey` or `authToken` in production environment
   - These are marked as "secret" type for security

2. **Collection-level Auth:**
   - The collection can be modified to include authentication
   - Add API key or Bearer token as needed

## 📊 Test Automation Features

### Automatic Variable Population
- Task IDs are automatically captured and stored
- Timestamps and random IDs are generated dynamically
- Test run IDs help identify related test data

### Comprehensive Validation
- **Response Status Codes** - Validates expected HTTP status
- **Response Structure** - Checks JSON structure and required fields
- **Data Integrity** - Ensures data consistency across requests
- **Error Handling** - Validates proper error responses
- **Performance** - Checks response times are acceptable

### Data Cleanup
- Tests automatically clean up created data
- Configurable via `testDataCleanup` environment variable
- Prevents test data accumulation

## 🐛 Troubleshooting

### Common Issues

#### 1. Connection Refused
**Error:** `Error: connect ECONNREFUSED`
**Solution:** 
- Ensure API server is running
- Verify host/port in environment variables
- Check firewall settings

#### 2. Invalid Cron Expression
**Error:** `400 Bad Request - invalid cron expression`
**Solution:**
- Use valid cron format: `minute hour day month weekday`
- Example: `0 9 * * *` (9 AM daily)

#### 3. Task Not Found
**Error:** `404 Not Found`
**Solution:**
- Ensure task was created successfully
- Check task ID is correctly populated
- Verify task wasn't deleted

#### 4. Database Connection Issues
**Error:** `500 Internal Server Error - database connection failed`
**Solution:**
- Check database is running
- Verify database configuration
- Check logs for detailed error

### Debugging Tips

1. **Enable Postman Console:**
   - View → Show Postman Console
   - See detailed request/response logs

2. **Check Environment Variables:**
   - Click eye icon to view current values
   - Ensure variables are populated correctly

3. **Review Test Scripts:**
   - Click on "Tests" tab in requests
   - Review validation logic

4. **Use Pre-request Scripts:**
   - Check "Pre-request Script" tab
   - Verify setup logic

## 📈 Performance Testing

### Load Testing with Newman

You can run the collection via command line for automated testing:

```bash
# Install Newman
npm install -g newman

# Run collection with local environment
newman run task-scheduler-api-complete.postman_collection.json \
  -e local-environment.postman_environment.json \
  --reporters cli,json \
  --reporter-json-export results.json

# Run with multiple iterations
newman run task-scheduler-api-complete.postman_collection.json \
  -e local-environment.postman_environment.json \
  -n 10 \
  --delay-request 1000
```

## 🔄 CI/CD Integration

### GitHub Actions Example

```yaml
name: API Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Start API Server
        run: |
          go run cmd/server/main.go &
          sleep 10
      
      - name: Install Newman
        run: npm install -g newman
      
      - name: Run Postman Tests
        run: |
          newman run docs/task-scheduler-api-complete.postman_collection.json \
            -e docs/local-environment.postman_environment.json \
            --reporters cli,junit \
            --reporter-junit-export test-results.xml
      
      - name: Publish Test Results
        uses: dorny/test-reporter@v1
        if: always()
        with:
          name: API Tests
          path: test-results.xml
          reporter: java-junit
```

## 📝 Contributing

When adding new API endpoints:

1. **Add new requests** to appropriate folder
2. **Include comprehensive tests** in the "Tests" tab
3. **Update environment variables** if needed
4. **Document new functionality** in this guide
5. **Test thoroughly** across all environments

## 📞 Support

For issues or questions:
- Check the troubleshooting section above
- Review API documentation in `swagger.yaml`
- Check server logs for detailed error information
- Open an issue in the project repository

---

**Happy Testing! 🚀**