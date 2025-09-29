-- Create task_results table
CREATE TABLE task_results (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    run_at TIMESTAMPTZ NOT NULL,
    status_code INT,
    success BOOLEAN,
    response_headers JSONB,
    response_body TEXT,
    error_message TEXT,
    duration_ms INT,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- Indexes for faster queries
CREATE INDEX idx_task_results_task_id ON task_results(task_id);
CREATE INDEX idx_task_results_run_at ON task_results(run_at);
CREATE INDEX idx_task_results_success ON task_results(success);
