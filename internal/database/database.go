package database

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"task-scheduler/internal/models"
)

var DB *gorm.DB

func Connect() {
	var err error

	// Try to use DATABASE_URL first, fallback to individual parameters
	databaseURL := getEnv("DATABASE_URL", "")
	if databaseURL != "" {
		// Use the full DATABASE_URL if provided
		DB, err = gorm.Open(postgres.Open(databaseURL), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	} else {
		// Fallback to constructing DSN from individual parameters
		dsn := fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=require TimeZone=UTC",
			getEnv("PGHOST", "localhost"),
			getEnv("PGUSER", "postgres"),
			getEnv("PGPASSWORD", "postgres"),
			getEnv("PGDATABASE", "task_scheduler"),
			getEnv("DB_PORT", "5432"),
		)

		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	}

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Database connected successfully")
}

func Migrate() {
	err := DB.AutoMigrate(&models.Task{}, &models.TaskResult{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	log.Println("Database migration completed")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
