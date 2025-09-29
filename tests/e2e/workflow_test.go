package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"task-scheduler/internal/executor"
	"task-scheduler/internal/models"
	"task-scheduler/internal/repository"
	"task-scheduler/internal/service"
	"task-scheduler/tests/utils"
)

type WorkflowTestSuite struct {
	suite.Suite
	helper       *utils.TestHelper
	taskRepo     *repository.TaskRepository
	resultRepo   *repository.ResultRepository
	taskService  *service.TaskService
	httpExecutor *executor.HTTPExecutor
	scenarios    *utils.TestScenarios
}

func (suite *WorkflowTestSuite) SetupSuite() {
	suite.helper = utils.NewTestHelper(suite.T())
	ctx := context.Background()

	err := suite.helper.SetupTestEnvironment(ctx)
	require.NoError(suite.T(), err)

	// Setup repositories and services
	suite.taskRepo = repository.NewTaskRepository(suite.helper.GetDB())
	suite.resultRepo = repository.NewResultRepository(suite.helper.GetDB())
	suite.httpExecutor = executor.NewHTTPExecutor()
	suite.scenarios = suite.helper.GetScenarios()
}

func (suite *WorkflowTestSuite) TearDownSuite() {
	ctx := context.Background()
	suite.helper.Cleanup(ctx)
}

func (suite *WorkflowTestSuite) SetupTest() {
	suite.helper.CleanDatabase()
	suite.helper.GetMockServer().ClearRequests()
	suite.helper.GetMockServer().ClearResponses()
	suite.helper.SetupMockResponseScenario("success")
}

func (suite *WorkflowTestSuite) TestOneOffTaskWorkflow() {
	suite.helper.LogTestStep("Test complete one-off task workflow")

	// Step 1: Create a one-off task
	suite.helper.LogTestStep("Step 1: Create task")
	mockServerURL := suite.helper.GetMockServer().GetURL()
	task := suite.scenarios.TaskFactory.CreateOneOffTask(
		"E2E One-Off Task",
		mockServerURL+"/webhook",
		time.Now().Add(time.Second*2), // Execute in 2 seconds
	)

	err := suite.taskRepo.Create(task)
	require.NoError(suite.T(), err)
	suite.helper.LogTestStep("Task created with ID: " + task.ID.String())

	// Step 2: Verify task is in database
	suite.helper.LogTestStep("Step 2: Verify task persistence")
	savedTask, err := suite.taskRepo.GetByID(task.ID)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), task.Name, savedTask.Name)
	assert.Equal(suite.T(), models.TaskStatusScheduled, savedTask.Status)

	// Step 3: Execute the task manually (simulating scheduler)
	suite.helper.LogTestStep("Step 3: Execute task")
	result := suite.httpExecutor.Execute(savedTask)
	require.NotNil(suite.T(), result)

	// Step 4: Save execution result
	suite.helper.LogTestStep("Step 4: Save execution result")
	err = suite.resultRepo.Create(result)
	require.NoError(suite.T(), err)

	// Step 5: Verify execution was successful
	suite.helper.LogTestStep("Step 5: Verify execution success")
	assert.True(suite.T(), result.Success)
	assert.NotNil(suite.T(), result.StatusCode)
	assert.Equal(suite.T(), 200, *result.StatusCode)
	assert.Greater(suite.T(), result.DurationMs, 0)

	// Step 6: Verify mock server received the request
	suite.helper.LogTestStep("Step 6: Verify HTTP request was made")
	capturedRequest := suite.helper.AssertRequestReceived("POST", "/webhook", time.Second*5)
	require.NotNil(suite.T(), capturedRequest)
	assert.Equal(suite.T(), "POST", capturedRequest.Method)
	assert.Contains(suite.T(), capturedRequest.Body, "test")

	// Step 7: Verify result is stored in database
	suite.helper.LogTestStep("Step 7: Verify result persistence")
	results, total, err := suite.resultRepo.GetByTaskID(task.ID, 10, 0)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), total)
	assert.Len(suite.T(), results, 1)

	storedResult := results[0]
	assert.Equal(suite.T(), task.ID, storedResult.TaskID)
	assert.True(suite.T(), storedResult.Success)
	assert.Equal(suite.T(), 200, *storedResult.StatusCode)

	// Step 8: Update task status to completed (simulating scheduler)
	suite.helper.LogTestStep("Step 8: Mark task as completed")
	savedTask.Status = models.TaskStatusCompleted
	err = suite.taskRepo.Update(savedTask)
	require.NoError(suite.T(), err)

	// Step 9: Final verification
	suite.helper.LogTestStep("Step 9: Final verification")
	finalTask, err := suite.taskRepo.GetByID(task.ID)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), models.TaskStatusCompleted, finalTask.Status)

	suite.helper.LogTestStep("✅ One-off task workflow completed successfully")
}

func (suite *WorkflowTestSuite) TestCronTaskWorkflow() {
	suite.helper.LogTestStep("Test cron task workflow (simulated multiple executions)")

	// Step 1: Create a cron task
	suite.helper.LogTestStep("Step 1: Create cron task")
	mockServerURL := suite.helper.GetMockServer().GetURL()
	task := suite.scenarios.TaskFactory.CreateCronTask(
		"E2E Cron Task",
		mockServerURL+"/api/endpoint",
		"*/1 * * * *", // Every minute (for testing)
	)

	err := suite.taskRepo.Create(task)
	require.NoError(suite.T(), err)
	suite.helper.LogTestStep("Cron task created with ID: " + task.ID.String())

	// Step 2: Simulate multiple executions
	suite.helper.LogTestStep("Step 2: Simulate multiple executions")
	const numExecutions = 3

	for i := 0; i < numExecutions; i++ {
		suite.helper.LogTestStep(fmt.Sprintf("Execution %d/%d", i+1, numExecutions))

		// Execute the task
		result := suite.httpExecutor.Execute(task)
		require.NotNil(suite.T(), result)
		require.True(suite.T(), result.Success)

		// Save result
		err = suite.resultRepo.Create(result)
		require.NoError(suite.T(), err)

		// Simulate time passing between executions
		time.Sleep(100 * time.Millisecond)
	}

	// Step 3: Verify all executions were recorded
	suite.helper.LogTestStep("Step 3: Verify all executions recorded")
	results, total, err := suite.resultRepo.GetByTaskID(task.ID, 10, 0)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(numExecutions), total)
	assert.Len(suite.T(), results, numExecutions)

	// Step 4: Verify all results are successful
	suite.helper.LogTestStep("Step 4: Verify execution results")
	for i, result := range results {
		assert.True(suite.T(), result.Success, "Execution %d should be successful", i+1)
		assert.Equal(suite.T(), 200, *result.StatusCode)
		assert.Equal(suite.T(), task.ID, result.TaskID)
	}

	// Step 5: Verify mock server received all requests
	suite.helper.LogTestStep("Step 5: Verify all HTTP requests received")
	suite.helper.WaitForCondition(func() bool {
		return suite.helper.GetMockServer().GetRequestCount() >= numExecutions
	}, time.Second*5, "All requests should be received")

	requests := suite.helper.GetMockServer().GetRequestsByMethod("GET")
	assert.GreaterOrEqual(suite.T(), len(requests), numExecutions)

	// Step 6: Task should still be scheduled (cron tasks don't complete)
	suite.helper.LogTestStep("Step 6: Verify task remains scheduled")
	finalTask, err := suite.taskRepo.GetByID(task.ID)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), models.TaskStatusScheduled, finalTask.Status)

	suite.helper.LogTestStep("✅ Cron task workflow completed successfully")
}

func (suite *WorkflowTestSuite) TestErrorHandlingWorkflow() {
	suite.helper.LogTestStep("Test error handling workflow")

	// Step 1: Setup mock server to return errors
	suite.helper.LogTestStep("Step 1: Setup error scenario")
	suite.helper.SetupMockResponseScenario("error")

	// Step 2: Create a task that will fail
	mockServerURL := suite.helper.GetMockServer().GetURL()
	task := suite.scenarios.TaskFactory.CreateOneOffTask(
		"E2E Error Task",
		mockServerURL+"/webhook",
		time.Now().Add(time.Second),
	)

	err := suite.taskRepo.Create(task)
	require.NoError(suite.T(), err)

	// Step 3: Execute the task (should fail)
	suite.helper.LogTestStep("Step 3: Execute failing task")
	result := suite.httpExecutor.Execute(task)
	require.NotNil(suite.T(), result)

	// Step 4: Verify failure was captured
	suite.helper.LogTestStep("Step 4: Verify failure captured")
	assert.False(suite.T(), result.Success)
	assert.NotNil(suite.T(), result.StatusCode)
	assert.Equal(suite.T(), 500, *result.StatusCode)
	assert.NotNil(suite.T(), result.ResponseBody)
	assert.Contains(suite.T(), *result.ResponseBody, "Internal server error")

	// Step 5: Save error result
	err = suite.resultRepo.Create(result)
	require.NoError(suite.T(), err)

	// Step 6: Verify error result in database
	suite.helper.LogTestStep("Step 6: Verify error result persistence")
	results, total, err := suite.resultRepo.GetByTaskID(task.ID, 10, 0)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), total)

	storedResult := results[0]
	assert.False(suite.T(), storedResult.Success)
	assert.Equal(suite.T(), 500, *storedResult.StatusCode)

	// Step 7: Query failed results
	suite.helper.LogTestStep("Step 7: Query failed results")
	failedResults, total, err := suite.resultRepo.List(10, 0, nil, boolPtr(false))
	require.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), total, int64(1))
	assert.GreaterOrEqual(suite.T(), len(failedResults), 1)

	suite.helper.LogTestStep("✅ Error handling workflow completed successfully")
}

func (suite *WorkflowTestSuite) TestTaskCancellationWorkflow() {
	suite.helper.LogTestStep("Test task cancellation workflow")

	// Step 1: Create a task
	suite.helper.LogTestStep("Step 1: Create task for cancellation")
	mockServerURL := suite.helper.GetMockServer().GetURL()
	task := suite.scenarios.TaskFactory.CreateOneOffTask(
		"E2E Cancellation Task",
		mockServerURL+"/webhook",
		time.Now().Add(time.Hour), // Future execution
	)

	err := suite.taskRepo.Create(task)
	require.NoError(suite.T(), err)

	// Step 2: Verify task is scheduled
	suite.helper.LogTestStep("Step 2: Verify initial task status")
	savedTask, err := suite.taskRepo.GetByID(task.ID)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), models.TaskStatusScheduled, savedTask.Status)

	// Step 3: Cancel the task
	suite.helper.LogTestStep("Step 3: Cancel task")
	err = suite.taskRepo.Delete(task.ID) // This sets status to cancelled
	require.NoError(suite.T(), err)

	// Step 4: Verify task is cancelled
	suite.helper.LogTestStep("Step 4: Verify task cancellation")
	cancelledTask, err := suite.taskRepo.GetByID(task.ID)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), models.TaskStatusCancelled, cancelledTask.Status)

	// Step 5: Verify cancelled task is not in scheduled list
	suite.helper.LogTestStep("Step 5: Verify task not in scheduled list")
	scheduledTasks, err := suite.taskRepo.GetScheduledTasks()
	require.NoError(suite.T(), err)

	for _, scheduledTask := range scheduledTasks {
		assert.NotEqual(suite.T(), task.ID, scheduledTask.ID, "Cancelled task should not be in scheduled list")
	}

	// Step 6: Query cancelled tasks
	suite.helper.LogTestStep("Step 6: Query cancelled tasks")
	cancelledTasks, total, err := suite.taskRepo.List(10, 0, string(models.TaskStatusCancelled))
	require.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), total, int64(1))

	found := false
	for _, t := range cancelledTasks {
		if t.ID == task.ID {
			found = true
			assert.Equal(suite.T(), models.TaskStatusCancelled, t.Status)
			break
		}
	}
	assert.True(suite.T(), found, "Cancelled task should be found in cancelled tasks list")

	suite.helper.LogTestStep("✅ Task cancellation workflow completed successfully")
}

func (suite *WorkflowTestSuite) TestComplexPayloadWorkflow() {
	suite.helper.LogTestStep("Test complex payload workflow")

	// Step 1: Create task with complex payload
	suite.helper.LogTestStep("Step 1: Create task with complex payload")
	mockServerURL := suite.helper.GetMockServer().GetURL()

	complexPayload := map[string]interface{}{
		"user": map[string]interface{}{
			"id":       123,
			"name":     "Test User",
			"settings": []string{"notification", "dark-mode"},
		},
		"data": []interface{}{
			map[string]interface{}{"key": "value1", "count": 10},
			map[string]interface{}{"key": "value2", "count": 20},
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	task := suite.scenarios.TaskFactory.CreateHTTPTask(
		"POST",
		mockServerURL+"/api/complex",
		models.Headers{"Content-Type": "application/json"},
		complexPayload,
	)

	err := suite.taskRepo.Create(task)
	require.NoError(suite.T(), err)

	// Step 2: Execute task
	suite.helper.LogTestStep("Step 2: Execute task with complex payload")
	result := suite.httpExecutor.Execute(task)
	require.NotNil(suite.T(), result)
	assert.True(suite.T(), result.Success)

	// Step 3: Verify complex payload was sent correctly
	suite.helper.LogTestStep("Step 3: Verify complex payload transmission")
	capturedRequest := suite.helper.AssertRequestReceived("POST", "/api/complex", time.Second*5)
	require.NotNil(suite.T(), capturedRequest)

	// Verify payload structure in received request
	assert.Contains(suite.T(), capturedRequest.Body, "Test User")
	assert.Contains(suite.T(), capturedRequest.Body, "notification")
	assert.Contains(suite.T(), capturedRequest.Body, "value1")
	assert.Contains(suite.T(), capturedRequest.Body, "count")

	// Step 4: Save and verify result
	err = suite.resultRepo.Create(result)
	require.NoError(suite.T(), err)

	suite.helper.LogTestStep("✅ Complex payload workflow completed successfully")
}

// Helper functions
func boolPtr(b bool) *bool {
	return &b
}

func TestWorkflowTestSuite(t *testing.T) {
	suite.Run(t, new(WorkflowTestSuite))
}
