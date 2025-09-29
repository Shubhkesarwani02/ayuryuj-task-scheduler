-- Create tasks table
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    trigger_type VARCHAR(20) NOT NULL CHECK (trigger_type IN ('one-off', 'cron')),
    trigger_time TIMESTAMPTZ NULL,          -- used if one-off
    cron_expr VARCHAR(255) NULL,            -- used if cron
    method VARCHAR(10) NOT NULL,
    url TEXT NOT NULL,
    headers JSONB,
    payload JSONB,
    status VARCHAR(20) NOT NULL DEFAULT 'scheduled' 
        CHECK (status IN ('scheduled', 'cancelled', 'completed')),
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    next_run TIMESTAMPTZ
);

-- Indexes for faster queries
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_next_run ON tasks(next_run);
CREATE INDEX idx_tasks_trigger_type ON tasks(trigger_type);
