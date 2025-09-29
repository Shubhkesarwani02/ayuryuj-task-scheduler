import { NextResponse } from "next/server"
import { TaskExecutor } from "@/lib/task-executor"

export async function POST() {
  try {
    TaskExecutor.stop()
    return NextResponse.json({ message: "Task executor stopped" })
  } catch (error) {
    console.error("Error stopping task executor:", error)
    return NextResponse.json({ error: "Failed to stop task executor" }, { status: 500 })
  }
}
