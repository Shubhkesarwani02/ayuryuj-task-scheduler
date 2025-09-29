package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"task-scheduler/internal/models"
	"task-scheduler/internal/repository"
	"task-scheduler/tests/utils"
)

type ResultRepositoryTestSuite struct {
	suite.Suite
	helper        *utils.TestHelper
	taskRepo      *repository.TaskRepository
	resultRepo    *repository.ResultRepository
	taskFactory   *utils.TaskFactory
	resultFactory *utils.ResultFactory
	testTask      *models.Task
}

func (suite *ResultRepositoryTestSuite) SetupSuite() {
	suite.helper = utils.NewTestHelper(suite.T())
	ctx := context.Background()

	err := suite.helper.SetupTestEnvironment(ctx)
	require.NoError(suite.T(), err)

	suite.taskRepo = repository.NewTaskRepository(suite.helper.GetDB())
	suite.resultRepo = repository.NewResultRepository(suite.helper.GetDB())
	suite.taskFactory = utils.NewTaskFactory()
	suite.resultFactory = utils.NewResultFactory()
}

func (suite *ResultRepositoryTestSuite) TearDownSuite() {
	ctx := context.Background()
	suite.helper.Cleanup(ctx)
}

func (suite *ResultRepositoryTestSuite) SetupTest() {
	// Clean database before each test
	suite.helper.CleanDatabase()

	// Create a test task for each test
	suite.testTask = suite.taskFactory.CreateOneOffTask(
		"Test Task",
		"https://example.com/webhook",
		time.Now().Add(time.Hour),
	)

	err := suite.taskRepo.Create(suite.testTask)
	require.NoError(suite.T(), err)
}

func (suite *ResultRepositoryTestSuite) TestCreate() {
	tests := []struct {
		name    string
		result  *models.TaskResult
		wantErr bool
	}{
		{
			name:    "create successful result",
			result:  suite.resultFactory.CreateSuccessResult(suite.testTask.ID),
			wantErr: false,
		},
		{
			name:    "create error result",
			result:  suite.resultFactory.CreateErrorResult(suite.testTask.ID, 500, "Internal server error"),
			wantErr: false,
		},
		{
			name:    "create timeout result",
			result:  suite.resultFactory.CreateTimeoutResult(suite.testTask.ID),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := suite.resultRepo.Create(tt.result)

			if tt.wantErr {
				assert.Error(suite.T(), err)
				return
			}

			require.NoError(suite.T(), err)
			assert.NotEqual(suite.T(), uuid.Nil, tt.result.ID)

			// Verify result was created in database
			var count int64
			err = suite.helper.GetDB().Table("task_results").Where("id = ?", tt.result.ID).Count(&count).Error
			require.NoError(suite.T(), err)
			assert.Equal(suite.T(), int64(1), count)
		})
	}
}

func (suite *ResultRepositoryTestSuite) TestGetByTaskID() {
	// Create multiple results for the test task
	results := []*models.TaskResult{
		suite.resultFactory.CreateSuccessResult(suite.testTask.ID),
		suite.resultFactory.CreateErrorResult(suite.testTask.ID, 404, "Not found"),
		suite.resultFactory.CreateSuccessResult(suite.testTask.ID),
	}

	// Set different run times to test ordering
	now := time.Now()
	results[0].RunAt = now.Add(-time.Hour)   // Oldest
	results[1].RunAt = now.Add(-time.Minute) // Middle
	results[2].RunAt = now                   // Newest

	// Create results in database
	for _, result := range results {
		err := suite.resultRepo.Create(result)
		require.NoError(suite.T(), err)
	}

	// Create results for a different task (should not be returned)
	otherTask := suite.taskFactory.CreateOneOffTask("Other Task", "https://example.com/other", time.Now().Add(time.Hour))
	err := suite.taskRepo.Create(otherTask)
	require.NoError(suite.T(), err)

	otherResult := suite.resultFactory.CreateSuccessResult(otherTask.ID)
	err = suite.resultRepo.Create(otherResult)
	require.NoError(suite.T(), err)

	tests := []struct {
		name          string
		taskID        uuid.UUID
		limit         int
		offset        int
		expectedCount int
		expectedTotal int64
	}{
		{
			name:          "get all results for task",
			taskID:        suite.testTask.ID,
			limit:         10,
			offset:        0,
			expectedCount: 3,
			expectedTotal: 3,
		},
		{
			name:          "get results with limit",
			taskID:        suite.testTask.ID,
			limit:         2,
			offset:        0,
			expectedCount: 2,
			expectedTotal: 3,
		},
		{
			name:          "get results with offset",
			taskID:        suite.testTask.ID,
			limit:         10,
			offset:        1,
			expectedCount: 2,
			expectedTotal: 3,
		},
		{
			name:          "get results for other task",
			taskID:        otherTask.ID,
			limit:         10,
			offset:        0,
			expectedCount: 1,
			expectedTotal: 1,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result, total, err := suite.resultRepo.GetByTaskID(tt.taskID, tt.limit, tt.offset)

			require.NoError(suite.T(), err)
			assert.Len(suite.T(), result, tt.expectedCount)
			assert.Equal(suite.T(), tt.expectedTotal, total)

			// Verify ordering (should be by run_at DESC)
			if len(result) > 1 {
				for i := 0; i < len(result)-1; i++ {
					assert.True(suite.T(),
						result[i].RunAt.After(result[i+1].RunAt) ||
							result[i].RunAt.Equal(result[i+1].RunAt),
						"Results should be ordered by run_at DESC")
				}
			}
		})
	}
}

func (suite *ResultRepositoryTestSuite) TestList() {
	// Create another task
	secondTask := suite.taskFactory.CreateCronTask("Second Task", "https://example.com/cron", "0 0 * * *")
	err := suite.taskRepo.Create(secondTask)
	require.NoError(suite.T(), err)

	// Create results for both tasks
	results := []*models.TaskResult{
		suite.resultFactory.CreateSuccessResult(suite.testTask.ID),
		suite.resultFactory.CreateErrorResult(suite.testTask.ID, 500, "Server error"),
		suite.resultFactory.CreateSuccessResult(secondTask.ID),
		suite.resultFactory.CreateTimeoutResult(secondTask.ID),
	}

	// Create results in database
	for _, result := range results {
		err := suite.resultRepo.Create(result)
		require.NoError(suite.T(), err)
	}

	tests := []struct {
		name          string
		limit         int
		offset        int
		taskID        *uuid.UUID
		success       *bool
		expectedCount int
		expectedTotal int64
	}{
		{
			name:          "list all results",
			limit:         10,
			offset:        0,
			taskID:        nil,
			success:       nil,
			expectedCount: 4,
			expectedTotal: 4,
		},
		{
			name:          "list results for specific task",
			limit:         10,
			offset:        0,
			taskID:        &suite.testTask.ID,
			success:       nil,
			expectedCount: 2,
			expectedTotal: 2,
		},
		{
			name:          "list successful results only",
			limit:         10,
			offset:        0,
			taskID:        nil,
			success:       boolPtr(true),
			expectedCount: 2,
			expectedTotal: 2,
		},
		{
			name:          "list failed results only",
			limit:         10,
			offset:        0,
			taskID:        nil,
			success:       boolPtr(false),
			expectedCount: 2,
			expectedTotal: 2,
		},
		{
			name:          "list with limit",
			limit:         2,
			offset:        0,
			taskID:        nil,
			success:       nil,
			expectedCount: 2,
			expectedTotal: 4,
		},
		{
			name:          "list with offset",
			limit:         10,
			offset:        2,
			taskID:        nil,
			success:       nil,
			expectedCount: 2,
			expectedTotal: 4,
		},
		{
			name:          "list successful results for specific task",
			limit:         10,
			offset:        0,
			taskID:        &suite.testTask.ID,
			success:       boolPtr(true),
			expectedCount: 1,
			expectedTotal: 1,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result, total, err := suite.resultRepo.List(tt.limit, tt.offset, tt.taskID, tt.success)

			require.NoError(suite.T(), err)
			assert.Len(suite.T(), result, tt.expectedCount)
			assert.Equal(suite.T(), tt.expectedTotal, total)

			// Verify filtering
			for _, res := range result {
				if tt.taskID != nil {
					assert.Equal(suite.T(), *tt.taskID, res.TaskID)
				}
				if tt.success != nil {
					assert.Equal(suite.T(), *tt.success, res.Success)
				}
				// Verify task is preloaded
				assert.NotEqual(suite.T(), uuid.Nil, res.Task.ID)
			}

			// Verify ordering (should be by run_at DESC)
			if len(result) > 1 {
				for i := 0; i < len(result)-1; i++ {
					assert.True(suite.T(),
						result[i].RunAt.After(result[i+1].RunAt) ||
							result[i].RunAt.Equal(result[i+1].RunAt),
						"Results should be ordered by run_at DESC")
				}
			}
		})
	}
}

func (suite *ResultRepositoryTestSuite) TestResultWithComplexData() {
	// Create a result with complex response headers and body
	result := &models.TaskResult{
		ID:         uuid.New(),
		TaskID:     suite.testTask.ID,
		RunAt:      time.Now(),
		Success:    true,
		StatusCode: intPtr(200),
		ResponseHeaders: models.Headers{
			"Content-Type":  "application/json",
			"Cache-Control": "no-cache",
			"X-RateLimit":   "100",
			"Authorization": "Bearer token123",
		},
		ResponseBody: stringPtr(`{
			"status": "success",
			"data": {
				"users": [
					{"id": 1, "name": "John"},
					{"id": 2, "name": "Jane"}
				],
				"total": 2
			},
			"metadata": {
				"timestamp": "2023-01-01T00:00:00Z",
				"version": "1.0"
			}
		}`),
		DurationMs: 250,
		CreatedAt:  time.Now(),
	}

	err := suite.resultRepo.Create(result)
	require.NoError(suite.T(), err)

	// Retrieve and verify
	results, total, err := suite.resultRepo.GetByTaskID(suite.testTask.ID, 10, 0)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), int64(1), total)
	require.Len(suite.T(), results, 1)

	retrieved := results[0]
	assert.Equal(suite.T(), result.ID, retrieved.ID)
	assert.Equal(suite.T(), result.TaskID, retrieved.TaskID)
	assert.Equal(suite.T(), result.Success, retrieved.Success)
	assert.Equal(suite.T(), result.StatusCode, retrieved.StatusCode)
	assert.Equal(suite.T(), result.ResponseHeaders, retrieved.ResponseHeaders)
	assert.Equal(suite.T(), result.ResponseBody, retrieved.ResponseBody)
	assert.Equal(suite.T(), result.DurationMs, retrieved.DurationMs)

	// Verify specific header values
	assert.Equal(suite.T(), "application/json", retrieved.ResponseHeaders["Content-Type"])
	assert.Equal(suite.T(), "Bearer token123", retrieved.ResponseHeaders["Authorization"])
}

func (suite *ResultRepositoryTestSuite) TestConcurrentResultCreation() {
	// Test concurrent result creation
	const numResults = 10

	resultChan := make(chan *models.TaskResult, numResults)
	errChan := make(chan error, numResults)

	// Create results concurrently
	for i := 0; i < numResults; i++ {
		go func(i int) {
			var result *models.TaskResult
			if i%2 == 0 {
				result = suite.resultFactory.CreateSuccessResult(suite.testTask.ID)
			} else {
				result = suite.resultFactory.CreateErrorResult(suite.testTask.ID, 500, "Error")
			}

			resultChan <- result
			errChan <- suite.resultRepo.Create(result)
		}(i)
	}

	// Collect results and errors
	createdResults := make([]*models.TaskResult, 0, numResults)
	for i := 0; i < numResults; i++ {
		result := <-resultChan
		err := <-errChan
		assert.NoError(suite.T(), err)
		createdResults = append(createdResults, result)
	}

	// Verify all results were created
	allResults, total, err := suite.resultRepo.GetByTaskID(suite.testTask.ID, 100, 0)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(numResults), total)
	assert.Len(suite.T(), allResults, numResults)

	// Verify success/failure distribution
	successCount := 0
	failureCount := 0
	for _, result := range allResults {
		if result.Success {
			successCount++
		} else {
			failureCount++
		}
	}

	assert.Equal(suite.T(), 5, successCount)
	assert.Equal(suite.T(), 5, failureCount)
}

// Helper functions
func boolPtr(b bool) *bool {
	return &b
}

func intPtr(i int) *int {
	return &i
}

func stringPtr(s string) *string {
	return &s
}

func TestResultRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(ResultRepositoryTestSuite))
}
