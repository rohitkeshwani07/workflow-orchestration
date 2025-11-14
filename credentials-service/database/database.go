package database

import (
	"fmt"
	"os"

	"github.com/workflow-orchestration/credentials-service/models"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DB struct {
	conn *gorm.DB
}

type Config struct {
	Driver   string
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
	FilePath string
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
			config.FilePath = "./data/credentials.db"
		}
		dialector = sqlite.Open(config.FilePath)

	default:
		return nil, fmt.Errorf("unsupported database driver: %s", config.Driver)
	}

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	conn, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db := &DB{conn: conn}

	// Auto-migrate
	if err := db.autoMigrate(); err != nil {
		return nil, fmt.Errorf("failed to auto-migrate database: %w", err)
	}

	return db, nil
}

func (db *DB) autoMigrate() error {
	return db.conn.AutoMigrate(
		&models.Credential{},
	)
}

func (db *DB) Close() error {
	sqlDB, err := db.conn.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Credential operations
func (db *DB) CreateCredential(c *models.Credential) error {
	return db.conn.Create(c).Error
}

func (db *DB) GetCredential(id string) (*models.Credential, error) {
	var credential models.Credential
	err := db.conn.First(&credential, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &credential, err
}

func (db *DB) GetAllCredentials() ([]models.Credential, error) {
	var credentials []models.Credential
	err := db.conn.Order("created_at DESC").Find(&credentials).Error
	return credentials, err
}

func (db *DB) UpdateCredential(c *models.Credential) error {
	return db.conn.Save(c).Error
}

func (db *DB) DeleteCredential(id string) error {
	return db.conn.Delete(&models.Credential{}, "id = ?", id).Error
}

// NewFromEnv creates a database connection from environment variables
func NewFromEnv() (*DB, error) {
	driver := os.Getenv("DB_DRIVER")
	if driver == "" {
		driver = "sqlite"
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
			config.DBName = "credentials"
		}
	}

	return New(config)
}
