package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"

	"task-scheduler/internal/models"
)

// TestDatabase represents a test database setup
type TestDatabase struct {
	Container *postgres.PostgresContainer
	DB        *gorm.DB
	DSN       string
}

// SetupTestDatabase creates a PostgreSQL test container and returns the database connection
func SetupTestDatabase(ctx context.Context) (*TestDatabase, error) {
	// Create PostgreSQL container
	postgresContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase("test_task_scheduler"),
		postgres.WithUsername("test_user"),
		postgres.WithPassword("test_password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Minute)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	// Get connection string
	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("failed to get connection string: %w", err)
	}

	// Connect to database
	db, err := gorm.Open(gormPostgres.Open(connStr), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Run migrations
	err = runMigrations(db)
	if err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &TestDatabase{
		Container: postgresContainer,
		DB:        db,
		DSN:       connStr,
	}, nil
}

// runMigrations runs the database migrations
func runMigrations(db *gorm.DB) error {
	// Enable UUID extension
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		return fmt.Errorf("failed to create uuid extension: %w", err)
	}

	// Auto migrate models
	if err := db.AutoMigrate(&models.Task{}, &models.TaskResult{}); err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}

	return nil
}

// Cleanup terminates the test database container
func (td *TestDatabase) Cleanup(ctx context.Context) error {
	if td.Container != nil {
		return td.Container.Terminate(ctx)
	}
	return nil
}

// CleanTables truncates all tables for test isolation
func (td *TestDatabase) CleanTables() error {
	return td.DB.Transaction(func(tx *gorm.DB) error {
		// Disable foreign key checks
		if err := tx.Exec("SET session_replication_role = 'replica'").Error; err != nil {
			return err
		}

		// Truncate tables
		if err := tx.Exec("TRUNCATE TABLE task_results CASCADE").Error; err != nil {
			return err
		}
		if err := tx.Exec("TRUNCATE TABLE tasks CASCADE").Error; err != nil {
			return err
		}

		// Re-enable foreign key checks
		if err := tx.Exec("SET session_replication_role = 'origin'").Error; err != nil {
			return err
		}

		return nil
	})
}

// SeedTestData inserts test data into the database
func (td *TestDatabase) SeedTestData() error {
	// This can be extended to insert common test data
	return nil
}

// GetDatabaseURL returns the database connection URL for external tools
func (td *TestDatabase) GetDatabaseURL() string {
	return td.DSN
}

// ExecSQL executes raw SQL for testing purposes
func (td *TestDatabase) ExecSQL(sql string, args ...interface{}) error {
	return td.DB.Exec(sql, args...).Error
}

// CountRecords counts records in a table
func (td *TestDatabase) CountRecords(tableName string) (int64, error) {
	var count int64
	err := td.DB.Table(tableName).Count(&count).Error
	return count, err
}

// WaitForContainerReady waits for the container to be ready
func (td *TestDatabase) WaitForContainerReady(ctx context.Context, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		var result int
		err := td.DB.Raw("SELECT 1").Scan(&result).Error
		if err == nil {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("database not ready within timeout")
}
