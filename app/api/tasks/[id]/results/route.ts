import { type NextRequest, NextResponse } from "next/server"
import { TaskService } from "@/lib/task-service"

export async function GET(request: NextRequest, { params }: { params: { id: string } }) {
  try {
    const { searchParams } = new URL(request.url)
    const page = Number.parseInt(searchParams.get("page") || "1")
    const limit = Number.parseInt(searchParams.get("limit") || "10")

    const { results, total } = await TaskService.getTaskResults(params.id, page, limit)

    return NextResponse.json({
      results,
      pagination: {
        page,
        limit,
        total,
        pages: Math.ceil(total / limit),
      },
    })
  } catch (error) {
    console.error("Error fetching task results:", error)
    return NextResponse.json({ error: "Failed to fetch task results" }, { status: 500 })
  }
}
