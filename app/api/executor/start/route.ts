import { NextResponse } from "next/server"
import { TaskExecutor } from "@/lib/task-executor"

export async function POST() {
  try {
    TaskExecutor.start()
    return NextResponse.json({ message: "Task executor started" })
  } catch (error) {
    console.error("Error starting task executor:", error)
    return NextResponse.json({ error: "Failed to start task executor" }, { status: 500 })
  }
}
