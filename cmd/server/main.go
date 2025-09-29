package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

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
	taskLogger, err := logger.NewTaskLogger("/var/log/task-scheduler/tasks.log")
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
		c.JSON(200, gin.H{"status": "healthy"})
	})

	// API routes
	api := r.Group("/api/v1")
	{
		// Task routes
		api.POST("/tasks", taskHandler.CreateTask)
		api.GET("/tasks", taskHandler.GetTasks)
		api.GET("/tasks/:id", taskHandler.GetTask)
		api.PUT("/tasks/:id", taskHandler.UpdateTask)
		api.DELETE("/tasks/:id", taskHandler.DeleteTask)
		api.GET("/tasks/:id/results", taskHandler.GetTaskResults)

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
