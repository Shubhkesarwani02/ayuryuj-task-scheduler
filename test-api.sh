#!/usr/bin/env bash

# Task Scheduler API Test Script
# This script tests all the API endpoints and functionality

BASE_URL="http://localhost:8080"
API_BASE="$BASE_URL/api/v1"

echo "=== Task Scheduler API Testing ==="
echo "Base URL: $BASE_URL"
echo "API Base: $API_BASE"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print test results
print_test() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓ PASS${NC}: $2"
    else
        echo -e "${RED}✗ FAIL${NC}: $2"
    fi
}

# Function to make HTTP requests
make_request() {
    local method=$1
    local url=$2
    local data=$3
    local content_type=${4:-"application/json"}
    
    if [ -n "$data" ]; then
        curl -s -X "$method" "$url" \
             -H "Content-Type: $content_type" \
             -d "$data" \
             -w "\nHTTP_STATUS:%{http_code}\n"
    else
        curl -s -X "$method" "$url" \
             -w "\nHTTP_STATUS:%{http_code}\n"
    fi
}

echo "=== 1. Health Check ==="
response=$(make_request "GET" "$BASE_URL/health")
status=$(echo "$response" | grep "HTTP_STATUS:" | cut -d':' -f2)
if [ "$status" = "200" ]; then
    print_test 0 "Health endpoint is working"
    echo "$response" | grep -v "HTTP_STATUS:"
else
    print_test 1 "Health endpoint failed (Status: $status)"
fi
echo ""

echo "=== 2. Swagger Documentation ==="
response=$(make_request "GET" "$BASE_URL/swagger/index.html")
status=$(echo "$response" | grep "HTTP_STATUS:" | cut -d':' -f2)
if [ "$status" = "200" ]; then
    print_test 0 "Swagger documentation is accessible"
else
    print_test 1 "Swagger documentation failed (Status: $status)"
fi
echo ""

echo "=== 3. Create One-off Task ==="
oneoff_task='{
  "name": "Test One-off Task",
  "trigger": {
    "type": "one-off",
    "datetime": "2025-09-29T20:00:00Z"
  },
  "action": {
    "method": "GET",
    "url": "https://httpbin.org/get",
    "headers": {
      "User-Agent": "Task-Scheduler-Test"
    }
  }
}'

response=$(make_request "POST" "$API_BASE/tasks" "$oneoff_task")
status=$(echo "$response" | grep "HTTP_STATUS:" | cut -d':' -f2)
if [ "$status" = "201" ]; then
    print_test 0 "One-off task created successfully"
    task_id=$(echo "$response" | grep -v "HTTP_STATUS:" | jq -r '.id' 2>/dev/null || echo "")
    echo "Task ID: $task_id"
else
    print_test 1 "One-off task creation failed (Status: $status)"
    echo "$response" | grep -v "HTTP_STATUS:"
fi
echo ""

echo "=== 4. Create Cron Task ==="
cron_task='{
  "name": "Test Cron Task",
  "trigger": {
    "type": "cron",
    "cron": "*/5 * * * *"
  },
  "action": {
    "method": "POST",
    "url": "https://httpbin.org/post",
    "headers": {
      "Content-Type": "application/json",
      "User-Agent": "Task-Scheduler-Test"
    },
    "payload": {
      "message": "Hello from cron task",
      "timestamp": "2025-09-29T18:45:00Z"
    }
  }
}'

response=$(make_request "POST" "$API_BASE/tasks" "$cron_task")
status=$(echo "$response" | grep "HTTP_STATUS:" | cut -d':' -f2)
if [ "$status" = "201" ]; then
    print_test 0 "Cron task created successfully"
    cron_task_id=$(echo "$response" | grep -v "HTTP_STATUS:" | jq -r '.id' 2>/dev/null || echo "")
    echo "Cron Task ID: $cron_task_id"
else
    print_test 1 "Cron task creation failed (Status: $status)"
    echo "$response" | grep -v "HTTP_STATUS:"
fi
echo ""

echo "=== 5. List All Tasks ==="
response=$(make_request "GET" "$API_BASE/tasks")
status=$(echo "$response" | grep "HTTP_STATUS:" | cut -d':' -f2)
if [ "$status" = "200" ]; then
    print_test 0 "Tasks listed successfully"
    task_count=$(echo "$response" | grep -v "HTTP_STATUS:" | jq '. | length' 2>/dev/null || echo "0")
    echo "Total tasks: $task_count"
else
    print_test 1 "Tasks listing failed (Status: $status)"
    echo "$response" | grep -v "HTTP_STATUS:"
fi
echo ""

echo "=== 6. Get Specific Task ==="
if [ -n "$task_id" ]; then
    response=$(make_request "GET" "$API_BASE/tasks/$task_id")
    status=$(echo "$response" | grep "HTTP_STATUS:" | cut -d':' -f2)
    if [ "$status" = "200" ]; then
        print_test 0 "Task details retrieved successfully"
        echo "$response" | grep -v "HTTP_STATUS:" | jq '.' 2>/dev/null || echo "$response" | grep -v "HTTP_STATUS:"
    else
        print_test 1 "Task details retrieval failed (Status: $status)"
    fi
else
    print_test 1 "Skipping - no task ID available"
fi
echo ""

echo "=== 7. Update Task ==="
if [ -n "$task_id" ]; then
    update_data='{
      "name": "Updated Test Task"
    }'
    response=$(make_request "PUT" "$API_BASE/tasks/$task_id" "$update_data")
    status=$(echo "$response" | grep "HTTP_STATUS:" | cut -d':' -f2)
    if [ "$status" = "200" ]; then
        print_test 0 "Task updated successfully"
    else
        print_test 1 "Task update failed (Status: $status)"
        echo "$response" | grep -v "HTTP_STATUS:"
    fi
else
    print_test 1 "Skipping - no task ID available"
fi
echo ""

echo "=== 8. Get Task Results ==="
if [ -n "$task_id" ]; then
    response=$(make_request "GET" "$API_BASE/tasks/$task_id/results")
    status=$(echo "$response" | grep "HTTP_STATUS:" | cut -d':' -f2)
    if [ "$status" = "200" ]; then
        print_test 0 "Task results retrieved successfully"
        result_count=$(echo "$response" | grep -v "HTTP_STATUS:" | jq '. | length' 2>/dev/null || echo "0")
        echo "Result count: $result_count"
    else
        print_test 1 "Task results retrieval failed (Status: $status)"
    fi
else
    print_test 1 "Skipping - no task ID available"
fi
echo ""

echo "=== 9. Get All Results ==="
response=$(make_request "GET" "$API_BASE/results")
status=$(echo "$response" | grep "HTTP_STATUS:" | cut -d':' -f2)
if [ "$status" = "200" ]; then
    print_test 0 "All results retrieved successfully"
    result_count=$(echo "$response" | grep -v "HTTP_STATUS:" | jq '. | length' 2>/dev/null || echo "0")
    echo "Total results: $result_count"
else
    print_test 1 "All results retrieval failed (Status: $status)"
fi
echo ""

echo "=== 10. Get Metrics ==="
response=$(make_request "GET" "$API_BASE/metrics")
status=$(echo "$response" | grep "HTTP_STATUS:" | cut -d':' -f2)
if [ "$status" = "200" ]; then
    print_test 0 "Metrics retrieved successfully"
    echo "$response" | grep -v "HTTP_STATUS:" | jq '.' 2>/dev/null || echo "$response" | grep -v "HTTP_STATUS:"
else
    print_test 1 "Metrics retrieval failed (Status: $status)"
fi
echo ""

echo "=== 11. Test Invalid Task Creation ==="
invalid_task='{
  "name": "",
  "trigger": {
    "type": "invalid"
  },
  "action": {
    "method": "GET"
  }
}'

response=$(make_request "POST" "$API_BASE/tasks" "$invalid_task")
status=$(echo "$response" | grep "HTTP_STATUS:" | cut -d':' -f2)
if [ "$status" = "400" ]; then
    print_test 0 "Invalid task properly rejected"
else
    print_test 1 "Invalid task should be rejected (Status: $status)"
fi
echo ""

echo "=== 12. Cancel Task ==="
if [ -n "$cron_task_id" ]; then
    response=$(make_request "DELETE" "$API_BASE/tasks/$cron_task_id")
    status=$(echo "$response" | grep "HTTP_STATUS:" | cut -d':' -f2)
    if [ "$status" = "200" ]; then
        print_test 0 "Task cancelled successfully"
    else
        print_test 1 "Task cancellation failed (Status: $status)"
        echo "$response" | grep -v "HTTP_STATUS:"
    fi
else
    print_test 1 "Skipping - no cron task ID available"
fi
echo ""

echo "=== 13. Test Pagination ==="
response=$(make_request "GET" "$API_BASE/tasks?page=1&limit=10")
status=$(echo "$response" | grep "HTTP_STATUS:" | cut -d':' -f2)
if [ "$status" = "200" ]; then
    print_test 0 "Pagination works correctly"
else
    print_test 1 "Pagination failed (Status: $status)"
fi
echo ""

echo "=== 14. Test Status Filter ==="
response=$(make_request "GET" "$API_BASE/tasks?status=scheduled")
status=$(echo "$response" | grep "HTTP_STATUS:" | cut -d':' -f2)
if [ "$status" = "200" ]; then
    print_test 0 "Status filtering works correctly"
else
    print_test 1 "Status filtering failed (Status: $status)"
fi
echo ""

echo ""
echo "=== Testing Complete ==="
echo ""
echo "Note: To test actual task execution, wait for the scheduled times or"
echo "create tasks with immediate execution times."
echo ""
echo "Swagger UI available at: $BASE_URL/swagger/index.html"
echo "Health endpoint: $BASE_URL/health"
echo ""