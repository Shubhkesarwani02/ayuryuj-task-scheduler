# Task Scheduler API Test Script (PowerShell)
# This script tests all the API endpoints and functionality

param(
    [string]$BaseUrl = "http://localhost:8080"
)

$ApiBase = "$BaseUrl/api/v1"

Write-Host "=== Task Scheduler API Testing ===" -ForegroundColor Cyan
Write-Host "Base URL: $BaseUrl"
Write-Host "API Base: $ApiBase"
Write-Host ""

# Function to print test results
function Write-TestResult {
    param([bool]$Success, [string]$Message)
    if ($Success) {
        Write-Host "✓ PASS: $Message" -ForegroundColor Green
    } else {
        Write-Host "✗ FAIL: $Message" -ForegroundColor Red
    }
}

# Function to make HTTP requests
function Invoke-ApiRequest {
    param(
        [string]$Method,
        [string]$Uri,
        [string]$Body = $null,
        [string]$ContentType = "application/json"
    )
    
    try {
        $headers = @{ "Content-Type" = $ContentType }
        if ($Body) {
            $response = Invoke-RestMethod -Uri $Uri -Method $Method -Body $Body -Headers $headers -ErrorAction Stop
        } else {
            $response = Invoke-RestMethod -Uri $Uri -Method $Method -ErrorAction Stop
        }
        return @{ Success = $true; Data = $response; StatusCode = 200 }
    } catch {
        return @{ Success = $false; Error = $_.Exception.Message; StatusCode = $_.Exception.Response.StatusCode.value__ }
    }
}

# Variables to store created task IDs
$TaskId = $null
$CronTaskId = $null

Write-Host "=== 1. Health Check ===" -ForegroundColor Yellow
$result = Invoke-ApiRequest -Method "GET" -Uri "$BaseUrl/health"
if ($result.Success) {
    Write-TestResult $true "Health endpoint is working"
    Write-Host ($result.Data | ConvertTo-Json -Depth 2)
} else {
    Write-TestResult $false "Health endpoint failed: $($result.Error)"
}
Write-Host ""

Write-Host "=== 2. Create One-off Task ===" -ForegroundColor Yellow
$oneoffTask = @{
    name = "Test One-off Task"
    trigger = @{
        type = "one-off"
        datetime = "2025-09-29T20:00:00Z"
    }
    action = @{
        method = "GET"
        url = "https://httpbin.org/get"
        headers = @{
            "User-Agent" = "Task-Scheduler-Test"
        }
    }
} | ConvertTo-Json -Depth 3

$result = Invoke-ApiRequest -Method "POST" -Uri "$ApiBase/tasks" -Body $oneoffTask
if ($result.Success) {
    Write-TestResult $true "One-off task created successfully"
    $TaskId = $result.Data.id
    Write-Host "Task ID: $TaskId"
} else {
    Write-TestResult $false "One-off task creation failed: $($result.Error)"
}
Write-Host ""

Write-Host "=== 3. Create Cron Task ===" -ForegroundColor Yellow
$cronTask = @{
    name = "Test Cron Task"
    trigger = @{
        type = "cron"
        cron = "*/5 * * * *"
    }
    action = @{
        method = "POST"
        url = "https://httpbin.org/post"
        headers = @{
            "Content-Type" = "application/json"
            "User-Agent" = "Task-Scheduler-Test"
        }
        payload = @{
            message = "Hello from cron task"
            timestamp = "2025-09-29T18:45:00Z"
        }
    }
} | ConvertTo-Json -Depth 3

$result = Invoke-ApiRequest -Method "POST" -Uri "$ApiBase/tasks" -Body $cronTask
if ($result.Success) {
    Write-TestResult $true "Cron task created successfully"
    $CronTaskId = $result.Data.id
    Write-Host "Cron Task ID: $CronTaskId"
} else {
    Write-TestResult $false "Cron task creation failed: $($result.Error)"
}
Write-Host ""

Write-Host "=== 4. List All Tasks ===" -ForegroundColor Yellow
$result = Invoke-ApiRequest -Method "GET" -Uri "$ApiBase/tasks"
if ($result.Success) {
    Write-TestResult $true "Tasks listed successfully"
    $taskCount = $result.Data.Count
    Write-Host "Total tasks: $taskCount"
    if ($taskCount -gt 0) {
        Write-Host ($result.Data | ConvertTo-Json -Depth 2)
    }
} else {
    Write-TestResult $false "Tasks listing failed: $($result.Error)"
}
Write-Host ""

Write-Host "=== 5. Get Specific Task ===" -ForegroundColor Yellow
if ($TaskId) {
    $result = Invoke-ApiRequest -Method "GET" -Uri "$ApiBase/tasks/$TaskId"
    if ($result.Success) {
        Write-TestResult $true "Task details retrieved successfully"
        Write-Host ($result.Data | ConvertTo-Json -Depth 2)
    } else {
        Write-TestResult $false "Task details retrieval failed: $($result.Error)"
    }
} else {
    Write-TestResult $false "Skipping - no task ID available"
}
Write-Host ""

Write-Host "=== 6. Update Task ===" -ForegroundColor Yellow
if ($TaskId) {
    $updateData = @{
        name = "Updated Test Task"
    } | ConvertTo-Json
    
    $result = Invoke-ApiRequest -Method "PUT" -Uri "$ApiBase/tasks/$TaskId" -Body $updateData
    if ($result.Success) {
        Write-TestResult $true "Task updated successfully"
    } else {
        Write-TestResult $false "Task update failed: $($result.Error)"
    }
} else {
    Write-TestResult $false "Skipping - no task ID available"
}
Write-Host ""

Write-Host "=== 7. Get Task Results ===" -ForegroundColor Yellow
if ($TaskId) {
    $result = Invoke-ApiRequest -Method "GET" -Uri "$ApiBase/tasks/$TaskId/results"
    if ($result.Success) {
        Write-TestResult $true "Task results retrieved successfully"
        $resultCount = $result.Data.Count
        Write-Host "Result count: $resultCount"
    } else {
        Write-TestResult $false "Task results retrieval failed: $($result.Error)"
    }
} else {
    Write-TestResult $false "Skipping - no task ID available"
}
Write-Host ""

Write-Host "=== 8. Get All Results ===" -ForegroundColor Yellow
$result = Invoke-ApiRequest -Method "GET" -Uri "$ApiBase/results"
if ($result.Success) {
    Write-TestResult $true "All results retrieved successfully"
    $resultCount = $result.Data.Count
    Write-Host "Total results: $resultCount"
} else {
    Write-TestResult $false "All results retrieval failed: $($result.Error)"
}
Write-Host ""

Write-Host "=== 9. Get Metrics ===" -ForegroundColor Yellow
$result = Invoke-ApiRequest -Method "GET" -Uri "$ApiBase/metrics"
if ($result.Success) {
    Write-TestResult $true "Metrics retrieved successfully"
    Write-Host ($result.Data | ConvertTo-Json -Depth 2)
} else {
    Write-TestResult $false "Metrics retrieval failed: $($result.Error)"
}
Write-Host ""

Write-Host "=== 10. Test Invalid Task Creation ===" -ForegroundColor Yellow
$invalidTask = @{
    name = ""
    trigger = @{
        type = "invalid"
    }
    action = @{
        method = "GET"
    }
} | ConvertTo-Json -Depth 3

$result = Invoke-ApiRequest -Method "POST" -Uri "$ApiBase/tasks" -Body $invalidTask
if (-not $result.Success -and $result.StatusCode -eq 400) {
    Write-TestResult $true "Invalid task properly rejected"
} else {
    Write-TestResult $false "Invalid task should be rejected (Status: $($result.StatusCode))"
}
Write-Host ""

Write-Host "=== 11. Cancel Task ===" -ForegroundColor Yellow
if ($CronTaskId) {
    $result = Invoke-ApiRequest -Method "DELETE" -Uri "$ApiBase/tasks/$CronTaskId"
    if ($result.Success) {
        Write-TestResult $true "Task cancelled successfully"
    } else {
        Write-TestResult $false "Task cancellation failed: $($result.Error)"
    }
} else {
    Write-TestResult $false "Skipping - no cron task ID available"
}
Write-Host ""

Write-Host "=== 12. Test Pagination ===" -ForegroundColor Yellow
$result = Invoke-ApiRequest -Method "GET" -Uri "$ApiBase/tasks?page=1`&limit=10"
if ($result.Success) {
    Write-TestResult $true "Pagination works correctly"
} else {
    Write-TestResult $false "Pagination failed: $($result.Error)"
}
Write-Host ""

Write-Host "=== 13. Test Status Filter ===" -ForegroundColor Yellow
$result = Invoke-ApiRequest -Method "GET" -Uri "$ApiBase/tasks?status=scheduled"
if ($result.Success) {
    Write-TestResult $true "Status filtering works correctly"
} else {
    Write-TestResult $false "Status filtering failed: $($result.Error)"
}
}
Write-Host ""

Write-Host ""
Write-Host "=== Testing Complete ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "Note: To test actual task execution, wait for the scheduled times or" -ForegroundColor Yellow
Write-Host "create tasks with immediate execution times." -ForegroundColor Yellow
Write-Host ""
Write-Host "Swagger UI available at: $BaseUrl/swagger/index.html" -ForegroundColor Green
Write-Host "Health endpoint: $BaseUrl/health" -ForegroundColor Green
Write-Host ""
Write-Host ""