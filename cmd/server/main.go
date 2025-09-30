package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "task-scheduler/docs"
	"task-scheduler/internal/database"
	"task-scheduler/internal/executor"
	"task-scheduler/internal/handlers"
	"task-scheduler/internal/logger"
	"task-scheduler/internal/metrics"
	"task-scheduler/internal/middleware"
	"task-scheduler/internal/repository"
	"task-scheduler/internal/scheduler"
)

// @title Task Scheduler API
// @version 1.0
// @description A RESTful Task Scheduler Service for managing HTTP tasks
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1
func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found or couldn't be loaded: %v", err)
	}

	// Connect to database
	database.Connect()
	database.Migrate()

	// Initialize repositories
	taskRepo := repository.NewTaskRepository(database.DB)
	resultRepo := repository.NewResultRepository(database.DB)

	// Initialize logging and metrics
	logPath := "./logs/tasks.log"
	if logDir := os.Getenv("LOG_DIR"); logDir != "" {
		logPath = logDir + "/tasks.log"
	}

	taskLogger, err := logger.NewTaskLogger(logPath)
	if err != nil {
		log.Printf("Failed to initialize task logger: %v", err)
		// Create a no-op logger or handle gracefully
		taskLogger = nil
	}
	defer func() {
		if taskLogger != nil {
			taskLogger.Close()
		}
	}()

	systemMetrics := metrics.NewMetrics()

	// Initialize executor and scheduler
	httpExecutor := executor.NewHTTPExecutor()
	taskScheduler := scheduler.NewScheduler(taskRepo, resultRepo, httpExecutor, taskLogger, systemMetrics)

	// Initialize handlers
	taskHandler := handlers.NewTaskHandler(taskRepo, resultRepo)
	resultHandler := handlers.NewResultHandler(resultRepo)
	metricsHandler := handlers.NewMetricsHandler(systemMetrics)

	// Start scheduler
	if err := taskScheduler.Start(); err != nil {
		log.Fatal("Failed to start scheduler:", err)
	}

	// Setup graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("Shutting down gracefully...")
		taskScheduler.Stop()
		os.Exit(0)
	}()

	// Setup Gin router
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())
	r.Use(gin.Recovery())

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
			"version":   "1.0.0",
			"components": gin.H{
				"database":  "healthy",
				"scheduler": "healthy",
			},
		})
	})

	// API routes
	api := r.Group("/api/v1")
	{
		// Health check endpoint in API namespace too
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":    "healthy",
				"timestamp": time.Now().Format(time.RFC3339),
				"version":   "1.0.0",
				"components": gin.H{
					"database":  "healthy",
					"scheduler": "healthy",
				},
			})
		})

		// Task routes
		api.POST("/tasks", taskHandler.CreateTask)
		api.GET("/tasks", taskHandler.GetTasks)
		api.GET("/tasks/:id", taskHandler.GetTask)
		api.PUT("/tasks/:id", taskHandler.UpdateTask)
		api.DELETE("/tasks/:id", taskHandler.DeleteTask)
		api.GET("/tasks/:id/results", taskHandler.GetTaskResults)

		// Task control routes
		api.POST("/tasks/:id/execute", taskHandler.ExecuteTask)
		api.POST("/tasks/:id/pause", taskHandler.PauseTask)
		api.POST("/tasks/:id/resume", taskHandler.ResumeTask)

		// Result routes
		api.GET("/results", resultHandler.GetResults)

		// Metrics routes
		api.GET("/metrics", metricsHandler.GetMetrics)
	}

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
