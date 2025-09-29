import { sql, type Task, type TaskResult, type CreateTaskData } from "./database"
import { cronParser } from "./cron-parser"

export class TaskService {
  static async createTask(data: CreateTaskData): Promise<Task> {
    const nextRun = this.calculateNextRun(data.trigger_type, data.trigger_value)

    const [task] = await sql`
      INSERT INTO tasks (name, trigger_type, trigger_value, method, url, headers, payload, next_run)
      VALUES (${data.name}, ${data.trigger_type}, ${data.trigger_value}, ${data.method}, ${data.url}, 
              ${JSON.stringify(data.headers || {})}, ${JSON.stringify(data.payload)}, ${nextRun})
      RETURNING *
    `

    return task as Task
  }

  static async getTasks(page = 1, limit = 10, status?: string): Promise<{ tasks: Task[]; total: number }> {
    const offset = (page - 1) * limit
    const whereClause = status ? sql`WHERE status = ${status}` : sql``

    const tasks = await sql`
      SELECT * FROM tasks 
      ${whereClause}
      ORDER BY created_at DESC 
      LIMIT ${limit} OFFSET ${offset}
    `

    const [{ count }] = await sql`
      SELECT COUNT(*) as count FROM tasks ${whereClause}
    `

    return { tasks: tasks as Task[], total: Number.parseInt(count) }
  }

  static async getTask(id: string): Promise<Task | null> {
    const [task] = await sql`SELECT * FROM tasks WHERE id = ${id}`
    return (task as Task) || null
  }

  static async updateTask(id: string, data: Partial<CreateTaskData>): Promise<Task | null> {
    const updates = []
    const values = []

    if (data.name) {
      updates.push("name = $" + (values.length + 1))
      values.push(data.name)
    }
    if (data.trigger_type && data.trigger_value) {
      updates.push("trigger_type = $" + (values.length + 1))
      values.push(data.trigger_type)
      updates.push("trigger_value = $" + (values.length + 1))
      values.push(data.trigger_value)
      updates.push("next_run = $" + (values.length + 1))
      values.push(this.calculateNextRun(data.trigger_type, data.trigger_value))
    }
    if (data.method) {
      updates.push("method = $" + (values.length + 1))
      values.push(data.method)
    }
    if (data.url) {
      updates.push("url = $" + (values.length + 1))
      values.push(data.url)
    }
    if (data.headers) {
      updates.push("headers = $" + (values.length + 1))
      values.push(JSON.stringify(data.headers))
    }
    if (data.payload !== undefined) {
      updates.push("payload = $" + (values.length + 1))
      values.push(JSON.stringify(data.payload))
    }

    if (updates.length === 0) return null

    values.push(id)
    const query = `UPDATE tasks SET ${updates.join(", ")} WHERE id = $${values.length} RETURNING *`

    const [task] = await sql.unsafe(query, values)
    return (task as Task) || null
  }

  static async cancelTask(id: string): Promise<boolean> {
    const result = await sql`
      UPDATE tasks SET status = 'cancelled' WHERE id = ${id} AND status = 'scheduled'
    `
    return result.count > 0
  }

  static async getTaskResults(taskId: string, page = 1, limit = 10): Promise<{ results: TaskResult[]; total: number }> {
    const offset = (page - 1) * limit

    const results = await sql`
      SELECT * FROM task_results 
      WHERE task_id = ${taskId}
      ORDER BY run_at DESC 
      LIMIT ${limit} OFFSET ${offset}
    `

    const [{ count }] = await sql`
      SELECT COUNT(*) as count FROM task_results WHERE task_id = ${taskId}
    `

    return { results: results as TaskResult[], total: Number.parseInt(count) }
  }

  static async getAllResults(
    page = 1,
    limit = 10,
    success?: boolean,
  ): Promise<{ results: TaskResult[]; total: number }> {
    const offset = (page - 1) * limit
    const whereClause = success !== undefined ? sql`WHERE success = ${success}` : sql``

    const results = await sql`
      SELECT tr.*, t.name as task_name 
      FROM task_results tr
      JOIN tasks t ON tr.task_id = t.id
      ${whereClause}
      ORDER BY tr.run_at DESC 
      LIMIT ${limit} OFFSET ${offset}
    `

    const [{ count }] = await sql`
      SELECT COUNT(*) as count FROM task_results tr ${whereClause}
    `

    return { results: results as (TaskResult & { task_name: string })[], total: Number.parseInt(count) }
  }

  static async createTaskResult(data: Omit<TaskResult, "id" | "created_at">): Promise<TaskResult> {
    const [result] = await sql`
      INSERT INTO task_results (task_id, run_at, status_code, success, response_headers, response_body, error_message, duration_ms)
      VALUES (${data.task_id}, ${data.run_at}, ${data.status_code}, ${data.success}, 
              ${JSON.stringify(data.response_headers)}, ${data.response_body}, ${data.error_message}, ${data.duration_ms})
      RETURNING *
    `

    return result as TaskResult
  }

  static async getScheduledTasks(): Promise<Task[]> {
    const now = new Date().toISOString()
    const tasks = await sql`
      SELECT * FROM tasks 
      WHERE status = 'scheduled' AND next_run <= ${now}
      ORDER BY next_run ASC
    `

    return tasks as Task[]
  }

  static async updateTaskAfterExecution(id: string, nextRun?: string): Promise<void> {
    const now = new Date().toISOString()

    if (nextRun) {
      await sql`
        UPDATE tasks 
        SET last_run = ${now}, next_run = ${nextRun}
        WHERE id = ${id}
      `
    } else {
      await sql`
        UPDATE tasks 
        SET last_run = ${now}, status = 'completed', next_run = NULL
        WHERE id = ${id}
      `
    }
  }

  private static calculateNextRun(triggerType: string, triggerValue: string): string | null {
    if (triggerType === "one-off") {
      return triggerValue
    } else if (triggerType === "cron") {
      return cronParser.getNextRun(triggerValue)
    }
    return null
  }
}
