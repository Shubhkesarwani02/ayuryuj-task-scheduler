import { type NextRequest, NextResponse } from "next/server"
import { TaskService } from "@/lib/task-service"

export async function GET(request: NextRequest, { params }: { params: { id: string } }) {
  try {
    const task = await TaskService.getTask(params.id)

    if (!task) {
      return NextResponse.json({ error: "Task not found" }, { status: 404 })
    }

    return NextResponse.json(task)
  } catch (error) {
    console.error("Error fetching task:", error)
    return NextResponse.json({ error: "Failed to fetch task" }, { status: 500 })
  }
}

export async function PUT(request: NextRequest, { params }: { params: { id: string } }) {
  try {
    const data = await request.json()
    const task = await TaskService.updateTask(params.id, data)

    if (!task) {
      return NextResponse.json({ error: "Task not found" }, { status: 404 })
    }

    return NextResponse.json(task)
  } catch (error) {
    console.error("Error updating task:", error)
    return NextResponse.json({ error: "Failed to update task" }, { status: 500 })
  }
}

export async function DELETE(request: NextRequest, { params }: { params: { id: string } }) {
  try {
    const success = await TaskService.cancelTask(params.id)

    if (!success) {
      return NextResponse.json({ error: "Task not found or already cancelled" }, { status: 404 })
    }

    return NextResponse.json({ message: "Task cancelled successfully" })
  } catch (error) {
    console.error("Error cancelling task:", error)
    return NextResponse.json({ error: "Failed to cancel task" }, { status: 500 })
  }
}
