import { type NextRequest, NextResponse } from "next/server"
import { TaskService } from "@/lib/task-service"

export async function GET(request: NextRequest) {
  try {
    const { searchParams } = new URL(request.url)
    const page = Number.parseInt(searchParams.get("page") || "1")
    const limit = Number.parseInt(searchParams.get("limit") || "10")
    const status = searchParams.get("status") || undefined

    const { tasks, total } = await TaskService.getTasks(page, limit, status)

    return NextResponse.json({
      tasks,
      pagination: {
        page,
        limit,
        total,
        pages: Math.ceil(total / limit),
      },
    })
  } catch (error) {
    console.error("Error fetching tasks:", error)
    return NextResponse.json({ error: "Failed to fetch tasks" }, { status: 500 })
  }
}

export async function POST(request: NextRequest) {
  try {
    const data = await request.json()

    // Validate required fields
    if (!data.name || !data.trigger_type || !data.trigger_value || !data.url) {
      return NextResponse.json(
        { error: "Missing required fields: name, trigger_type, trigger_value, url" },
        { status: 400 },
      )
    }

    // Validate trigger type
    if (!["one-off", "cron"].includes(data.trigger_type)) {
      return NextResponse.json({ error: 'trigger_type must be either "one-off" or "cron"' }, { status: 400 })
    }

    // Validate one-off datetime
    if (data.trigger_type === "one-off") {
      const triggerDate = new Date(data.trigger_value)
      if (isNaN(triggerDate.getTime())) {
        return NextResponse.json({ error: "Invalid datetime format for one-off trigger" }, { status: 400 })
      }
      if (triggerDate <= new Date()) {
        return NextResponse.json({ error: "One-off trigger must be in the future" }, { status: 400 })
      }
    }

    const task = await TaskService.createTask(data)
    return NextResponse.json(task, { status: 201 })
  } catch (error) {
    console.error("Error creating task:", error)
    return NextResponse.json({ error: "Failed to create task" }, { status: 500 })
  }
}
