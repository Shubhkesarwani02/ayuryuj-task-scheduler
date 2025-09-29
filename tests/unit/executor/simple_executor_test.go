package executor

import (
	"testing"

	"task-scheduler/internal/executor"
)

func TestExecutorCreation(t *testing.T) {
	httpExecutor := executor.NewHTTPExecutor()
	if httpExecutor == nil {
		t.Fatal("Expected HTTP executor to be created")
	}
}
