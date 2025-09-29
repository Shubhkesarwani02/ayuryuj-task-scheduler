export class cronParser {
  static getNextRun(cronExpression: string): string {
    // Simple cron parser - in production, use a library like node-cron
    const now = new Date()
    const parts = cronExpression.split(" ")

    if (parts.length !== 5) {
      throw new Error("Invalid cron expression")
    }

    // For demo purposes, add 1 minute to current time
    // In production, implement proper cron parsing
    const nextRun = new Date(now.getTime() + 60000)
    return nextRun.toISOString()
  }

  static isValid(cronExpression: string): boolean {
    const parts = cronExpression.split(" ")
    return parts.length === 5
  }
}
