import type React from "react"
import type { Metadata } from "next"
import { GeistSans } from "geist/font/sans"
import { GeistMono } from "geist/font/mono"
import { Analytics } from "@vercel/analytics/next"
import "./globals.css"
import { TaskExecutorProvider } from "@/components/task-executor-provider"
import { Suspense } from "react"

export const metadata: Metadata = {
  title: "Task Scheduler - HTTP Task Management",
  description: "Professional task scheduler for managing and monitoring HTTP requests",
  generator: "v0.app",
}

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html lang="en">
      <body className={`font-sans ${GeistSans.variable} ${GeistMono.variable}`}>
        <Suspense fallback={<div>Loading...</div>}>
          <TaskExecutorProvider>{children}</TaskExecutorProvider>
        </Suspense>
        <Analytics />
      </body>
    </html>
  )
}
