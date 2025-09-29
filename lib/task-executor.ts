import { TaskService } from "./task-service"
import type { Task } from "./database"

export class TaskExecutor {
  private static isRunning = false
  private static intervalId: NodeJS.Timeout | null = null

  static start() {
    if (this.isRunning) return

    this.isRunning = true
    console.log("[TaskExecutor] Starting task executor...")

    // Check for tasks every 30 seconds
    this.intervalId = setInterval(async () => {
      await this.processPendingTasks()
    }, 30000)

    // Process immediately on start
    this.processPendingTasks()
  }

  static stop() {
    if (!this.isRunning) return

    this.isRunning = false
    if (this.intervalId) {
      clearInterval(this.intervalId)
      this.intervalId = null
    }
    console.log("[TaskExecutor] Task executor stopped")
  }

  private static async processPendingTasks() {
    try {
      const tasks = await TaskService.getScheduledTasks()
      console.log(`[TaskExecutor] Found ${tasks.length} pending tasks`)

      for (const task of tasks) {
        await this.executeTask(task)
      }
    } catch (error) {
      console.error("[TaskExecutor] Error processing tasks:", error)
    }
  }

  private static async executeTask(task: Task) {
    const startTime = Date.now()
    const runAt = new Date().toISOString()

    console.log(`[TaskExecutor] Executing task: ${task.name} (${task.id})`)

    try {
      const response = await fetch(task.url, {
        method: task.method,
        headers: {
          "Content-Type": "application/json",
          ...task.headers,
        },
        body: task.payload ? JSON.stringify(task.payload) : undefined,
      })

      const duration = Date.now() - startTime
      const responseBody = await response.text()
      const responseHeaders = Object.fromEntries(response.headers.entries())

      // Create task result
      await TaskService.createTaskResult({
        task_id: task.id,
        run_at: runAt,
        status_code: response.status,
        success: response.ok,
        response_headers: responseHeaders,
        response_body: responseBody.substring(0, 10000), // Limit response body size
        error_message: response.ok ? null : `HTTP ${response.status}: ${response.statusText}`,
        duration_ms: duration,
      })

      // Update task for next run
      const nextRun = task.trigger_type === "cron" ? this.calculateNextCronRun(task.trigger_value) : null
      await TaskService.updateTaskAfterExecution(task.id, nextRun)

      console.log(`[TaskExecutor] Task completed: ${task.name} - Status: ${response.status}`)
    } catch (error) {
      const duration = Date.now() - startTime
      const errorMessage = error instanceof Error ? error.message : "Unknown error"

      // Create error result
      await TaskService.createTaskResult({
        task_id: task.id,
        run_at: runAt,
        success: false,
        response_headers: {},
        error_message: errorMessage,
        duration_ms: duration,
      })

      console.error(`[TaskExecutor] Task failed: ${task.name} - Error: ${errorMessage}`)
    }
  }

  private static calculateNextCronRun(cronExpression: string): string {
    // Simple implementation - add 1 hour for demo
    // In production, use a proper cron library
    const nextRun = new Date(Date.now() + 3600000)
    return nextRun.toISOString()
  }
}
