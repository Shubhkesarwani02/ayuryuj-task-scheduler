import { neon } from "@neondatabase/serverless"

if (!process.env.DATABASE_URL) {
  throw new Error("DATABASE_URL environment variable is not set")
}

export const sql = neon(process.env.DATABASE_URL)

export interface Task {
  id: string
  name: string
  trigger_type: "one-off" | "cron"
  trigger_value: string
  method: string
  url: string
  headers: Record<string, string>
  payload?: any
  status: "scheduled" | "cancelled" | "completed"
  created_at: string
  updated_at: string
  next_run?: string
  last_run?: string
}

export interface TaskResult {
  id: string
  task_id: string
  run_at: string
  status_code?: number
  success: boolean
  response_headers: Record<string, string>
  response_body?: string
  error_message?: string
  duration_ms?: number
  created_at: string
}

export interface CreateTaskData {
  name: string
  trigger_type: "one-off" | "cron"
  trigger_value: string
  method: string
  url: string
  headers?: Record<string, string>
  payload?: any
}
