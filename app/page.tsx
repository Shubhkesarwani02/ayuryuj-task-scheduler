"use client"

import { useState, useEffect } from "react"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Textarea } from "@/components/ui/textarea"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Clock, Play, Pause, CheckCircle, XCircle, Calendar, Activity, Database } from "lucide-react"
import type { Task, TaskResult } from "@/lib/database"

export default function TaskSchedulerDashboard() {
  const [tasks, setTasks] = useState<Task[]>([])
  const [results, setResults] = useState<(TaskResult & { task_name?: string })[]>([])
  const [loading, setLoading] = useState(true)
  const [createDialogOpen, setCreateDialogOpen] = useState(false)

  useEffect(() => {
    fetchTasks()
    fetchResults()
  }, [])

  const fetchTasks = async () => {
    try {
      const response = await fetch("/api/tasks")
      const data = await response.json()
      setTasks(data.tasks || [])
    } catch (error) {
      console.error("Error fetching tasks:", error)
    } finally {
      setLoading(false)
    }
  }

  const fetchResults = async () => {
    try {
      const response = await fetch("/api/results?limit=20")
      const data = await response.json()
      setResults(data.results || [])
    } catch (error) {
      console.error("Error fetching results:", error)
    }
  }

  const createTask = async (formData: FormData) => {
    try {
      const payload = {
        name: formData.get("name"),
        trigger_type: formData.get("trigger_type"),
        trigger_value: formData.get("trigger_value"),
        method: formData.get("method"),
        url: formData.get("url"),
        headers: formData.get("headers") ? JSON.parse(formData.get("headers") as string) : {},
        payload: formData.get("payload") ? JSON.parse(formData.get("payload") as string) : null,
      }

      const response = await fetch("/api/tasks", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      })

      if (response.ok) {
        setCreateDialogOpen(false)
        fetchTasks()
      }
    } catch (error) {
      console.error("Error creating task:", error)
    }
  }

  const cancelTask = async (id: string) => {
    try {
      await fetch(`/api/tasks/${id}`, { method: "DELETE" })
      fetchTasks()
    } catch (error) {
      console.error("Error cancelling task:", error)
    }
  }

  const getStatusBadge = (status: string) => {
    const variants = {
      scheduled: "default",
      cancelled: "destructive",
      completed: "secondary",
    } as const

    return <Badge variant={variants[status as keyof typeof variants] || "default"}>{status}</Badge>
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "scheduled":
        return <Clock className="h-4 w-4 text-primary" />
      case "cancelled":
        return <XCircle className="h-4 w-4 text-destructive" />
      case "completed":
        return <CheckCircle className="h-4 w-4 text-success" />
      default:
        return <Activity className="h-4 w-4" />
    }
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="text-center">
          <Database className="h-12 w-12 text-primary mx-auto mb-4 animate-pulse" />
          <p className="text-muted-foreground">Loading task scheduler...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <div className="border-b border-border bg-card">
        <div className="container mx-auto px-6 py-4">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-bold text-foreground">Task Scheduler</h1>
              <p className="text-muted-foreground">Manage and monitor HTTP task execution</p>
            </div>
            <Dialog open={createDialogOpen} onOpenChange={setCreateDialogOpen}>
              <DialogTrigger asChild>
                <Button className="bg-primary hover:bg-primary/90">
                  <Play className="h-4 w-4 mr-2" />
                  Create Task
                </Button>
              </DialogTrigger>
              <DialogContent className="max-w-2xl">
                <DialogHeader>
                  <DialogTitle>Create New Task</DialogTitle>
                  <DialogDescription>Schedule a new HTTP task for one-time or recurring execution</DialogDescription>
                </DialogHeader>
                <form action={createTask} className="space-y-4">
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <Label htmlFor="name">Task Name</Label>
                      <Input id="name" name="name" placeholder="My API Task" required />
                    </div>
                    <div>
                      <Label htmlFor="method">HTTP Method</Label>
                      <Select name="method" defaultValue="GET">
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="GET">GET</SelectItem>
                          <SelectItem value="POST">POST</SelectItem>
                          <SelectItem value="PUT">PUT</SelectItem>
                          <SelectItem value="DELETE">DELETE</SelectItem>
                          <SelectItem value="PATCH">PATCH</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                  </div>

                  <div>
                    <Label htmlFor="url">URL</Label>
                    <Input id="url" name="url" placeholder="https://api.example.com/endpoint" required />
                  </div>

                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <Label htmlFor="trigger_type">Trigger Type</Label>
                      <Select name="trigger_type" defaultValue="one-off">
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="one-off">One-off</SelectItem>
                          <SelectItem value="cron">Cron</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                    <div>
                      <Label htmlFor="trigger_value">Trigger Value</Label>
                      <Input
                        id="trigger_value"
                        name="trigger_value"
                        placeholder="2024-12-31T23:59:59Z or 0 */5 * * * *"
                        required
                      />
                    </div>
                  </div>

                  <div>
                    <Label htmlFor="headers">Headers (JSON)</Label>
                    <Textarea
                      id="headers"
                      name="headers"
                      placeholder='{"Authorization": "Bearer token", "Content-Type": "application/json"}'
                      className="font-mono text-sm"
                    />
                  </div>

                  <div>
                    <Label htmlFor="payload">Payload (JSON)</Label>
                    <Textarea
                      id="payload"
                      name="payload"
                      placeholder='{"key": "value", "data": "example"}'
                      className="font-mono text-sm"
                    />
                  </div>

                  <div className="flex justify-end gap-2">
                    <Button type="button" variant="outline" onClick={() => setCreateDialogOpen(false)}>
                      Cancel
                    </Button>
                    <Button type="submit">Create Task</Button>
                  </div>
                </form>
              </DialogContent>
            </Dialog>
          </div>
        </div>
      </div>

      <div className="container mx-auto px-6 py-8">
        <Tabs defaultValue="tasks" className="space-y-6">
          <TabsList className="grid w-full grid-cols-2 max-w-md">
            <TabsTrigger value="tasks" className="flex items-center gap-2">
              <Calendar className="h-4 w-4" />
              Tasks
            </TabsTrigger>
            <TabsTrigger value="results" className="flex items-center gap-2">
              <Activity className="h-4 w-4" />
              Results
            </TabsTrigger>
          </TabsList>

          <TabsContent value="tasks" className="space-y-4">
            <div className="grid gap-4">
              {tasks.length === 0 ? (
                <Card>
                  <CardContent className="flex flex-col items-center justify-center py-12">
                    <Calendar className="h-12 w-12 text-muted-foreground mb-4" />
                    <h3 className="text-lg font-semibold mb-2">No tasks scheduled</h3>
                    <p className="text-muted-foreground text-center mb-4">
                      Create your first task to start scheduling HTTP requests
                    </p>
                    <Button onClick={() => setCreateDialogOpen(true)}>Create Task</Button>
                  </CardContent>
                </Card>
              ) : (
                tasks.map((task) => (
                  <Card key={task.id} className="hover:bg-accent/50 transition-colors">
                    <CardHeader className="pb-3">
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-3">
                          {getStatusIcon(task.status)}
                          <div>
                            <CardTitle className="text-lg">{task.name}</CardTitle>
                            <CardDescription className="flex items-center gap-2 mt-1">
                              <Badge variant="outline" className="text-xs">
                                {task.method}
                              </Badge>
                              <span className="text-xs text-muted-foreground">{task.url}</span>
                            </CardDescription>
                          </div>
                        </div>
                        <div className="flex items-center gap-2">
                          {getStatusBadge(task.status)}
                          {task.status === "scheduled" && (
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => cancelTask(task.id)}
                              className="text-destructive hover:text-destructive"
                            >
                              <Pause className="h-4 w-4" />
                            </Button>
                          )}
                        </div>
                      </div>
                    </CardHeader>
                    <CardContent className="pt-0">
                      <div className="grid grid-cols-2 gap-4 text-sm">
                        <div>
                          <span className="text-muted-foreground">Trigger:</span>
                          <p className="font-mono text-xs mt-1">
                            {task.trigger_type === "one-off"
                              ? new Date(task.trigger_value).toLocaleString()
                              : task.trigger_value}
                          </p>
                        </div>
                        <div>
                          <span className="text-muted-foreground">Next Run:</span>
                          <p className="text-xs mt-1">
                            {task.next_run ? new Date(task.next_run).toLocaleString() : "N/A"}
                          </p>
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                ))
              )}
            </div>
          </TabsContent>

          <TabsContent value="results" className="space-y-4">
            <div className="grid gap-4">
              {results.length === 0 ? (
                <Card>
                  <CardContent className="flex flex-col items-center justify-center py-12">
                    <Activity className="h-12 w-12 text-muted-foreground mb-4" />
                    <h3 className="text-lg font-semibold mb-2">No execution results</h3>
                    <p className="text-muted-foreground text-center">
                      Task execution results will appear here once tasks start running
                    </p>
                  </CardContent>
                </Card>
              ) : (
                results.map((result) => (
                  <Card key={result.id} className="hover:bg-accent/50 transition-colors">
                    <CardHeader className="pb-3">
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-3">
                          {result.success ? (
                            <CheckCircle className="h-4 w-4 text-success" />
                          ) : (
                            <XCircle className="h-4 w-4 text-destructive" />
                          )}
                          <div>
                            <CardTitle className="text-lg">{result.task_name || "Unknown Task"}</CardTitle>
                            <CardDescription>{new Date(result.run_at).toLocaleString()}</CardDescription>
                          </div>
                        </div>
                        <div className="flex items-center gap-2">
                          {result.status_code && (
                            <Badge variant={result.success ? "default" : "destructive"}>{result.status_code}</Badge>
                          )}
                          {result.duration_ms && (
                            <Badge variant="outline" className="text-xs">
                              {result.duration_ms}ms
                            </Badge>
                          )}
                        </div>
                      </div>
                    </CardHeader>
                    {(result.error_message || result.response_body) && (
                      <CardContent className="pt-0">
                        <div className="bg-muted rounded-lg p-3">
                          <pre className="text-xs text-muted-foreground overflow-x-auto">
                            {result.error_message ||
                              result.response_body?.substring(0, 200) +
                                (result.response_body?.length > 200 ? "..." : "")}
                          </pre>
                        </div>
                      </CardContent>
                    )}
                  </Card>
                ))
              )}
            </div>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  )
}
