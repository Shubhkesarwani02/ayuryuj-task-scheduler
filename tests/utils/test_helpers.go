package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"task-scheduler/internal/handlers"
	"task-scheduler/internal/repository"
)

// TestHelper provides common testing utilities
type TestHelper struct {
	t         *testing.T
	db        *TestDatabase
	mock      *MockHTTPServer
	scenarios *TestScenarios
}

// NewTestHelper creates a new test helper
func NewTestHelper(t *testing.T) *TestHelper {
	return &TestHelper{
		t:         t,
		scenarios: NewTestScenarios(),
	}
}

// SetupTestEnvironment sets up the complete test environment
func (h *TestHelper) SetupTestEnvironment(ctx context.Context) error {
	var err error

	// Setup test database
	h.db, err = SetupTestDatabase(ctx)
	if err != nil {
		return fmt.Errorf("failed to setup test database: %w", err)
	}

	// Setup mock HTTP server
	h.mock = NewMockHTTPServer()

	return nil
}

// Cleanup cleans up all test resources
func (h *TestHelper) Cleanup(ctx context.Context) {
	if h.mock != nil {
		h.mock.Close()
	}
	if h.db != nil {
		h.db.Cleanup(ctx)
	}
}

// GetDB returns the test database
func (h *TestHelper) GetDB() *gorm.DB {
	return h.db.DB
}

// GetMockServer returns the mock HTTP server
func (h *TestHelper) GetMockServer() *MockHTTPServer {
	return h.mock
}

// GetScenarios returns the test scenarios helper
func (h *TestHelper) GetScenarios() *TestScenarios {
	return h.scenarios
}

// CleanDatabase cleans all tables in the test database
func (h *TestHelper) CleanDatabase() {
	require.NoError(h.t, h.db.CleanTables())
}

// CreateTestRouter creates a Gin router for testing handlers
func (h *TestHelper) CreateTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add basic middleware
	router.Use(gin.Recovery())

	return router
}

// SetupTaskHandlers sets up task handlers with repositories
func (h *TestHelper) SetupTaskHandlers(router *gin.Engine) (*handlers.TaskHandler, *handlers.ResultHandler, *repository.TaskRepository, *repository.ResultRepository) {
	taskRepo := repository.NewTaskRepository(h.db.DB)
	resultRepo := repository.NewResultRepository(h.db.DB)
	taskHandler := handlers.NewTaskHandler(taskRepo, resultRepo)
	resultHandler := handlers.NewResultHandler(resultRepo)

	// Setup routes
	v1 := router.Group("/api/v1")
	{
		v1.POST("/tasks", taskHandler.CreateTask)
		v1.GET("/tasks", taskHandler.GetTasks)
		v1.GET("/tasks/:id", taskHandler.GetTask)
		v1.PUT("/tasks/:id", taskHandler.UpdateTask)
		v1.DELETE("/tasks/:id", taskHandler.DeleteTask)
		v1.GET("/tasks/:id/results", taskHandler.GetTaskResults)
		v1.GET("/results", resultHandler.GetResults)
	}

	return taskHandler, resultHandler, taskRepo, resultRepo
}

// MakeJSONRequest makes a JSON HTTP request for testing
func (h *TestHelper) MakeJSONRequest(method, url string, body interface{}) *http.Request {
	var reqBody *bytes.Buffer

	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(h.t, err)
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, url, reqBody)
	require.NoError(h.t, err)

	req.Header.Set("Content-Type", "application/json")
	return req
}

// PerformRequest performs a request and returns the response recorder
func (h *TestHelper) PerformRequest(router *gin.Engine, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// AssertJSONResponse asserts that the response contains expected JSON
func (h *TestHelper) AssertJSONResponse(w *httptest.ResponseRecorder, expectedStatus int, expectedBody interface{}) {
	assert.Equal(h.t, expectedStatus, w.Code)
	assert.Equal(h.t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

	if expectedBody != nil {
		expectedJSON, err := json.Marshal(expectedBody)
		require.NoError(h.t, err)

		assert.JSONEq(h.t, string(expectedJSON), w.Body.String())
	}
}

// AssertErrorResponse asserts that the response contains an error
func (h *TestHelper) AssertErrorResponse(w *httptest.ResponseRecorder, expectedStatus int, expectedError string) {
	assert.Equal(h.t, expectedStatus, w.Code)
	assert.Equal(h.t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

	var errorResp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &errorResp)
	require.NoError(h.t, err)

	assert.Contains(h.t, errorResp["error"], expectedError)
}

// AssertTaskInDatabase asserts that a task exists in the database with expected values
func (h *TestHelper) AssertTaskInDatabase(taskID string, expectedStatus string) {
	var count int64
	err := h.db.DB.Table("tasks").Where("id = ? AND status = ?", taskID, expectedStatus).Count(&count).Error
	require.NoError(h.t, err)
	assert.Equal(h.t, int64(1), count, "Task should exist in database with expected status")
}

// AssertTaskResultInDatabase asserts that a task result exists in the database
func (h *TestHelper) AssertTaskResultInDatabase(taskID string, success bool) {
	var count int64
	err := h.db.DB.Table("task_results").Where("task_id = ? AND success = ?", taskID, success).Count(&count).Error
	require.NoError(h.t, err)
	assert.Greater(h.t, count, int64(0), "Task result should exist in database")
}

// WaitForCondition waits for a condition to be true with timeout
func (h *TestHelper) WaitForCondition(condition func() bool, timeout time.Duration, message string) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	h.t.Fatalf("Condition not met within timeout: %s", message)
}

// AssertRequestReceived asserts that the mock server received a request
func (h *TestHelper) AssertRequestReceived(method, path string, timeout time.Duration) *CapturedRequest {
	var foundRequest *CapturedRequest

	h.WaitForCondition(func() bool {
		requests := h.mock.GetRequests()
		for _, req := range requests {
			if req.Method == method && req.URL == path {
				foundRequest = &req
				return true
			}
		}
		return false
	}, timeout, fmt.Sprintf("Expected %s request to %s", method, path))

	return foundRequest
}

// ParseJSONResponse parses JSON response into a struct
func (h *TestHelper) ParseJSONResponse(w *httptest.ResponseRecorder, target interface{}) {
	err := json.Unmarshal(w.Body.Bytes(), target)
	require.NoError(h.t, err)
}

// CreateTestContext creates a test context with timeout
func (h *TestHelper) CreateTestContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// LogTestStep logs a test step for better debugging
func (h *TestHelper) LogTestStep(step string) {
	h.t.Logf("=== TEST STEP: %s ===", step)
}

// SetupMockResponseScenario sets up common mock response scenarios
func (h *TestHelper) SetupMockResponseScenario(scenario string) {
	switch scenario {
	case "success":
		h.mock.SetSuccessResponse("POST", "/webhook", map[string]string{
			"status":  "received",
			"message": "webhook processed successfully",
		})
		h.mock.SetSuccessResponse("GET", "/api/endpoint", map[string]interface{}{
			"data":      []string{"item1", "item2"},
			"timestamp": time.Now().Format(time.RFC3339),
		})
	case "error":
		h.mock.SetErrorResponse("POST", "/webhook", http.StatusInternalServerError, "Internal server error")
		h.mock.SetErrorResponse("GET", "/api/endpoint", http.StatusBadRequest, "Bad request")
	case "timeout":
		h.mock.SetTimeoutResponse("POST", "/webhook", time.Second*35) // Longer than executor timeout
	}
}
