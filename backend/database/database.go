package database

import (
	"fmt"
	"os"
	"time"

	"github.com/workflow-orchestration/backend/models"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DB struct {
	conn *gorm.DB
}

type Config struct {
	Driver   string // "postgres" or "sqlite"
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
	FilePath string // for SQLite
}

func New(config Config) (*DB, error) {
	var dialector gorm.Dialector

	switch config.Driver {
	case "postgres":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
			config.Host, config.User, config.Password, config.DBName, config.Port, config.SSLMode)
		dialector = postgres.Open(dsn)

	case "sqlite":
		if config.FilePath == "" {
			config.FilePath = "./data/workflows.db"
		}
		dialector = sqlite.Open(config.FilePath)

	default:
		return nil, fmt.Errorf("unsupported database driver: %s", config.Driver)
	}

	// GORM config
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	conn, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db := &DB{conn: conn}

	// Auto-migrate schemas
	if err := db.autoMigrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

func (db *DB) autoMigrate() error {
	return db.conn.AutoMigrate(
		&models.Workflow{},
		&models.Execution{},
		&models.NodeExecutionLog{},
		&models.ChatSession{},
		&models.ChatMessage{},
	)
}

func (db *DB) Close() error {
	sqlDB, err := db.conn.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (db *DB) GetDB() *gorm.DB {
	return db.conn
}

// Workflow operations
func (db *DB) CreateWorkflow(w *models.Workflow) error {
	return db.conn.Create(w).Error
}

func (db *DB) GetWorkflow(id string) (*models.Workflow, error) {
	var workflow models.Workflow
	err := db.conn.First(&workflow, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &workflow, err
}

func (db *DB) GetAllWorkflows() ([]models.Workflow, error) {
	var workflows []models.Workflow
	err := db.conn.Order("updated_at DESC").Find(&workflows).Error
	return workflows, err
}

func (db *DB) UpdateWorkflow(w *models.Workflow) error {
	return db.conn.Save(w).Error
}

func (db *DB) DeleteWorkflow(id string) error {
	return db.conn.Delete(&models.Workflow{}, "id = ?", id).Error
}

// Execution operations
func (db *DB) CreateExecution(e *models.Execution) error {
	return db.conn.Create(e).Error
}

func (db *DB) UpdateExecution(e *models.Execution) error {
	return db.conn.Save(e).Error
}

func (db *DB) GetExecution(id string) (*models.Execution, error) {
	var execution models.Execution
	err := db.conn.First(&execution, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &execution, err
}

func (db *DB) GetWorkflowExecutions(workflowID string, limit int) ([]models.Execution, error) {
	var executions []models.Execution
	err := db.conn.Where("workflow_id = ?", workflowID).
		Order("started_at DESC").
		Limit(limit).
		Find(&executions).Error
	return executions, err
}

func (db *DB) AddExecutionLog(executionID, nodeID, status string, output *string, errMsg *string) error {
	log := &models.NodeExecutionLog{
		ExecutionID: executionID,
		NodeID:      nodeID,
		Status:      status,
		Output:      output,
		Error:       errMsg,
		ExecutedAt:  time.Now(),
	}
	return db.conn.Create(log).Error
}

func (db *DB) GetExecutionLogs(executionID string) ([]models.NodeExecutionLog, error) {
	var logs []models.NodeExecutionLog
	err := db.conn.Where("execution_id = ?", executionID).
		Order("executed_at ASC").
		Find(&logs).Error
	return logs, err
}

// Chat operations
func (db *DB) CreateChatSession(sessionID, workflowID string) error {
	session := &models.ChatSession{
		ID:         sessionID,
		WorkflowID: workflowID,
		CreatedAt:  time.Now(),
	}
	return db.conn.Create(session).Error
}

func (db *DB) AddChatMessage(msg *models.ChatMessage) error {
	return db.conn.Create(msg).Error
}

func (db *DB) GetChatMessages(sessionID string) ([]models.ChatMessage, error) {
	var messages []models.ChatMessage
	err := db.conn.Where("session_id = ?", sessionID).
		Order("timestamp ASC").
		Find(&messages).Error
	return messages, err
}

// NewFromEnv creates a database connection from environment variables
func NewFromEnv() (*DB, error) {
	driver := os.Getenv("DB_DRIVER")
	if driver == "" {
		driver = "sqlite" // default to SQLite
	}

	config := Config{
		Driver:   driver,
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
		FilePath: os.Getenv("DATABASE_PATH"),
	}

	// Set defaults for PostgreSQL
	if config.Driver == "postgres" {
		if config.Host == "" {
			config.Host = "localhost"
		}
		if config.Port == "" {
			config.Port = "5432"
		}
		if config.SSLMode == "" {
			config.SSLMode = "disable"
		}
		if config.DBName == "" {
			config.DBName = "workflows"
		}
	}

	return New(config)
}
