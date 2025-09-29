package api

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"task-scheduler/internal/models"
	"task-scheduler/tests/utils"
)

type SimpleAPITestSuite struct {
	suite.Suite
	helper *utils.TestHelper
}

func (suite *SimpleAPITestSuite) SetupSuite() {
	suite.helper = utils.NewTestHelper(suite.T())
	ctx := context.Background()

	err := suite.helper.SetupTestEnvironment(ctx)
	require.NoError(suite.T(), err)
}

func (suite *SimpleAPITestSuite) TearDownSuite() {
	ctx := context.Background()
	suite.helper.Cleanup(ctx)
}

func (suite *SimpleAPITestSuite) SetupTest() {
	suite.helper.CleanDatabase()
}

func (suite *SimpleAPITestSuite) TestBasicTaskCreation() {
	suite.helper.LogTestStep("Test basic task creation via repository")

	// Setup
	_, _, taskRepo, _ := suite.helper.SetupTaskHandlers(suite.helper.CreateTestRouter())

	// Create a test task directly via repository
	factory := utils.NewTaskFactory()
	task := factory.CreateOneOffTask(
		"Simple Test Task",
		"https://example.com/webhook",
		time.Now().Add(time.Hour),
	)

	// Save task
	err := taskRepo.Create(task)
	require.NoError(suite.T(), err)

	// Verify task was saved
	savedTask, err := taskRepo.GetByID(task.ID)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), task.Name, savedTask.Name)
	assert.Equal(suite.T(), task.URL, savedTask.URL)
	assert.Equal(suite.T(), models.TaskStatusScheduled, savedTask.Status)

	suite.helper.LogTestStep("✅ Basic task creation test passed")
}

func (suite *SimpleAPITestSuite) TestTaskListing() {
	suite.helper.LogTestStep("Test basic task listing via repository")

	// Setup
	_, _, taskRepo, _ := suite.helper.SetupTaskHandlers(suite.helper.CreateTestRouter())

	// Create some test tasks
	factory := utils.NewTaskFactory()
	task1 := factory.CreateOneOffTask("Task 1", "https://example.com/1", time.Now().Add(time.Hour))
	task2 := factory.CreateCronTask("Task 2", "https://example.com/2", "0 0 * * *")

	// Save tasks
	require.NoError(suite.T(), taskRepo.Create(task1))
	require.NoError(suite.T(), taskRepo.Create(task2))

	// List tasks
	tasks, total, err := taskRepo.List(10, 0, "")
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), int64(2), total)
	assert.Len(suite.T(), tasks, 2)

	suite.helper.LogTestStep("✅ Basic task listing test passed")
}

func TestSimpleAPITestSuite(t *testing.T) {
	suite.Run(t, new(SimpleAPITestSuite))
}
