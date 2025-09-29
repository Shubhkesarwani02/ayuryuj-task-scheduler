import { type NextRequest, NextResponse } from "next/server"
import { TaskService } from "@/lib/task-service"

export async function GET(request: NextRequest) {
  try {
    const { searchParams } = new URL(request.url)
    const page = Number.parseInt(searchParams.get("page") || "1")
    const limit = Number.parseInt(searchParams.get("limit") || "10")
    const success = searchParams.get("success")

    const successFilter = success === "true" ? true : success === "false" ? false : undefined

    const { results, total } = await TaskService.getAllResults(page, limit, successFilter)

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
    console.error("Error fetching results:", error)
    return NextResponse.json({ error: "Failed to fetch results" }, { status: 500 })
  }
}
