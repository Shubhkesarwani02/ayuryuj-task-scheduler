package repository

import (
	"context"
	"fmt"
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

type TaskRepositoryTestSuite struct {
	suite.Suite
	helper   *utils.TestHelper
	repo     *repository.TaskRepository
	factory  *utils.TaskFactory
}

func (suite *TaskRepositoryTestSuite) SetupSuite() {
	suite.helper = utils.NewTestHelper(suite.T())
	ctx := context.Background()
	
	err := suite.helper.SetupTestEnvironment(ctx)
	require.NoError(suite.T(), err)
	
	suite.repo = repository.NewTaskRepository(suite.helper.GetDB())
	suite.factory = utils.NewTaskFactory()
}

func (suite *TaskRepositoryTestSuite) TearDownSuite() {
	ctx := context.Background()
	suite.helper.Cleanup(ctx)
}

func (suite *TaskRepositoryTestSuite) SetupTest() {
	// Clean database before each test
	suite.helper.CleanDatabase()
}

func (suite *TaskRepositoryTestSuite) TestCreate() {
	tests := []struct {
		name    string
		task    *models.Task
		wantErr bool
	}{
		{
			name: "create one-off task",
			task: suite.factory.CreateOneOffTask(
				"Test One-Off Task",
				"https://example.com/webhook",
				time.Now().Add(time.Hour),
			),
			wantErr: false,
		},
		{
			name: "create cron task",
			task: suite.factory.CreateCronTask(
				"Test Cron Task",
				"https://example.com/api",
				"0 0 * * *",
			),
			wantErr: false,
		},
		{
			name: "create task with payload",
			task: suite.factory.CreateHTTPTask(
				"POST",
				"https://example.com/webhook",
				models.Headers{"Content-Type": "application/json"},
				map[string]string{"key": "value"},
			),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := suite.repo.Create(tt.task)
			
			if tt.wantErr {
				assert.Error(suite.T(), err)
				return
			}
			
			require.NoError(suite.T(), err)
			assert.NotEqual(suite.T(), uuid.Nil, tt.task.ID)
			
			// Verify task was created in database
			var count int64
			err = suite.helper.GetDB().Table("tasks").Where("id = ?", tt.task.ID).Count(&count).Error
			require.NoError(suite.T(), err)
			assert.Equal(suite.T(), int64(1), count)
		})
	}
}

func (suite *TaskRepositoryTestSuite) TestGetByID() {
	// Create a test task
	task := suite.factory.CreateOneOffTask(
		"Test Task",
		"https://example.com/webhook",
		time.Now().Add(time.Hour),
	)
	
	err := suite.repo.Create(task)
	require.NoError(suite.T(), err)

	tests := []struct {
		name    string
		id      uuid.UUID
		wantErr bool
		want    *models.Task
	}{
		{
			name:    "existing task",
			id:      task.ID,
			wantErr: false,
			want:    task,
		},
		{
			name:    "non-existing task",
			id:      uuid.New(),
			wantErr: true,
			want:    nil,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result, err := suite.repo.GetByID(tt.id)
			
			if tt.wantErr {
				assert.Error(suite.T(), err)
				assert.Nil(suite.T(), result)
				return
			}
			
			require.NoError(suite.T(), err)
			require.NotNil(suite.T(), result)
			assert.Equal(suite.T(), tt.want.ID, result.ID)
			assert.Equal(suite.T(), tt.want.Name, result.Name)
			assert.Equal(suite.T(), tt.want.URL, result.URL)
			assert.Equal(suite.T(), tt.want.Method, result.Method)
		})
	}
}

func (suite *TaskRepositoryTestSuite) TestList() {
	// Create test tasks with different statuses
	tasks := []*models.Task{
		suite.factory.CreateOneOffTask("Task 1", "https://example.com/1", time.Now().Add(time.Hour)),
		suite.factory.CreateOneOffTask("Task 2", "https://example.com/2", time.Now().Add(time.Hour)),
		suite.factory.CreateCronTask("Task 3", "https://example.com/3", "0 0 * * *"),
	}
	
	// Set different statuses
	tasks[1].Status = models.TaskStatusCancelled
	tasks[2].Status = models.TaskStatusCompleted
	
	// Create tasks in database
	for _, task := range tasks {
		err := suite.repo.Create(task)
		require.NoError(suite.T(), err)
	}

	tests := []struct {
		name         string
		limit        int
		offset       int
		status       string
		expectedCount int
		expectedTotal int64
	}{
		{
			name:          "list all tasks",
			limit:         10,
			offset:        0,
			status:        "",
			expectedCount: 3,
			expectedTotal: 3,
		},
		{
			name:          "list scheduled tasks only",
			limit:         10,
			offset:        0,
			status:        string(models.TaskStatusScheduled),
			expectedCount: 1,
			expectedTotal: 1,
		},
		{
			name:          "list cancelled tasks only",
			limit:         10,
			offset:        0,
			status:        string(models.TaskStatusCancelled),
			expectedCount: 1,
			expectedTotal: 1,
		},
		{
			name:          "list with limit",
			limit:         2,
			offset:        0,
			status:        "",
			expectedCount: 2,
			expectedTotal: 3,
		},
		{
			name:          "list with offset",
			limit:         10,
			offset:        1,
			status:        "",
			expectedCount: 2,
			expectedTotal: 3,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result, total, err := suite.repo.List(tt.limit, tt.offset, tt.status)
			
			require.NoError(suite.T(), err)
			assert.Len(suite.T(), result, tt.expectedCount)
			assert.Equal(suite.T(), tt.expectedTotal, total)
			
			// Verify ordering (should be by created_at DESC)
			if len(result) > 1 {
				for i := 0; i < len(result)-1; i++ {
					assert.True(suite.T(), 
						result[i].CreatedAt.After(result[i+1].CreatedAt) || 
						result[i].CreatedAt.Equal(result[i+1].CreatedAt),
						"Tasks should be ordered by created_at DESC")
				}
			}
		})
	}
}

func (suite *TaskRepositoryTestSuite) TestUpdate() {
	// Create a test task
	task := suite.factory.CreateOneOffTask(
		"Original Task",
		"https://example.com/original",
		time.Now().Add(time.Hour),
	)
	
	err := suite.repo.Create(task)
	require.NoError(suite.T(), err)

	// Update the task
	task.Name = "Updated Task"
	task.URL = "https://example.com/updated"
	task.Status = models.TaskStatusCancelled
	task.UpdatedAt = time.Now()

	err = suite.repo.Update(task)
	require.NoError(suite.T(), err)

	// Verify the update
	updated, err := suite.repo.GetByID(task.ID)
	require.NoError(suite.T(), err)
	
	assert.Equal(suite.T(), "Updated Task", updated.Name)
	assert.Equal(suite.T(), "https://example.com/updated", updated.URL)
	assert.Equal(suite.T(), models.TaskStatusCancelled, updated.Status)
}

func (suite *TaskRepositoryTestSuite) TestDelete() {
	// Create a test task
	task := suite.factory.CreateOneOffTask(
		"Task to Delete",
		"https://example.com/delete",
		time.Now().Add(time.Hour),
	)
	
	err := suite.repo.Create(task)
	require.NoError(suite.T(), err)

	// Delete the task
	err = suite.repo.Delete(task.ID)
	require.NoError(suite.T(), err)

	// Verify deletion
	_, err = suite.repo.GetByID(task.ID)
	assert.Error(suite.T(), err)
}

func (suite *TaskRepositoryTestSuite) TestGetScheduledTasks() {
	now := time.Now()
	
	// Create tasks with different next run times
	tasks := []*models.Task{
		suite.factory.CreateOneOffTask("Past Task", "https://example.com/1", now.Add(-time.Hour)),
		suite.factory.CreateOneOffTask("Current Task", "https://example.com/2", now.Add(time.Minute)),
		suite.factory.CreateOneOffTask("Future Task", "https://example.com/3", now.Add(time.Hour)),
	}
	
	// Set next run times
	pastTime := now.Add(-time.Hour)
	currentTime := now.Add(time.Minute)
	futureTime := now.Add(time.Hour)
	
	tasks[0].NextRun = &pastTime
	tasks[1].NextRun = &currentTime
	tasks[2].NextRun = &futureTime
	
	// Create tasks in database
	for _, task := range tasks {
		err := suite.repo.Create(task)
		require.NoError(suite.T(), err)
	}

	// Get scheduled tasks - uses current time internally
	scheduledTasks, err := suite.repo.GetScheduledTasks()
	require.NoError(suite.T(), err)
	
	// Should return tasks with next_run <= current time or null
	assert.GreaterOrEqual(suite.T(), len(scheduledTasks), 1)
	
	// Verify all returned tasks are scheduled
	for _, task := range scheduledTasks {
		assert.Equal(suite.T(), models.TaskStatusScheduled, task.Status)
	}
}

func (suite *TaskRepositoryTestSuite) TestUpdateNextRun() {
	// Create a cron task
	task := suite.factory.CreateCronTask(
		"Cron Task",
		"https://example.com/cron",
		"0 0 * * *",
	)
	
	err := suite.repo.Create(task)
	require.NoError(suite.T(), err)

	// Update next run time
	newNextRun := time.Now().Add(24 * time.Hour)
	err = suite.repo.UpdateNextRun(task.ID, &newNextRun)
	require.NoError(suite.T(), err)

	// Verify the update
	updated, err := suite.repo.GetByID(task.ID)
	require.NoError(suite.T(), err)
	
	require.NotNil(suite.T(), updated.NextRun)
	assert.True(suite.T(), updated.NextRun.Equal(newNextRun))
}

func (suite *TaskRepositoryTestSuite) TestTaskWithHeaders() {
	// Create task with custom headers
	headers := models.Headers{
		"Authorization": "Bearer token123",
		"Content-Type":  "application/json",
		"X-Custom":      "custom-value",
	}
	
	task := suite.factory.CreateHTTPTask(
		"POST",
		"https://example.com/webhook",
		headers,
		map[string]string{"key": "value"},
	)
	
	err := suite.repo.Create(task)
	require.NoError(suite.T(), err)

	// Retrieve and verify headers
	retrieved, err := suite.repo.GetByID(task.ID)
	require.NoError(suite.T(), err)
	
	assert.Equal(suite.T(), headers, retrieved.Headers)
	assert.Equal(suite.T(), "Bearer token123", retrieved.Headers["Authorization"])
	assert.Equal(suite.T(), "application/json", retrieved.Headers["Content-Type"])
	assert.Equal(suite.T(), "custom-value", retrieved.Headers["X-Custom"])
}

func (suite *TaskRepositoryTestSuite) TestConcurrentAccess() {
	// Test concurrent creation and reading
	const numTasks = 10
	
	tasks := make([]*models.Task, numTasks)
	for i := 0; i < numTasks; i++ {
		tasks[i] = suite.factory.CreateOneOffTask(
			fmt.Sprintf("Concurrent Task %d", i),
			fmt.Sprintf("https://example.com/task%d", i),
			time.Now().Add(time.Hour),
		)
	}
	
	// Create tasks concurrently
	errChan := make(chan error, numTasks)
	for i := 0; i < numTasks; i++ {
		go func(task *models.Task) {
			errChan <- suite.repo.Create(task)
		}(tasks[i])
	}
	
	// Collect errors
	for i := 0; i < numTasks; i++ {
		err := <-errChan
		assert.NoError(suite.T(), err)
	}
	
	// Verify all tasks were created
	allTasks, total, err := suite.repo.List(100, 0, "")
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(numTasks), total)
	assert.Len(suite.T(), allTasks, numTasks)
}

func TestTaskRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(TaskRepositoryTestSuite))
}