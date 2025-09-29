"use client"

import type React from "react"

import { useEffect } from "react"

export function TaskExecutorProvider({ children }: { children: React.ReactNode }) {
  useEffect(() => {
    // Start the task executor when the app loads
    const startExecutor = async () => {
      try {
        await fetch("/api/executor/start", { method: "POST" })
        console.log("[TaskExecutor] Started successfully")
      } catch (error) {
        console.error("[TaskExecutor] Failed to start:", error)
      }
    }

    startExecutor()

    // Cleanup on unmount
    return () => {
      fetch("/api/executor/stop", { method: "POST" }).catch(console.error)
    }
  }, [])

  return <>{children}</>
}
