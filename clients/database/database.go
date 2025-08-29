package database

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	DB *gorm.DB
}

// New creates a new database connection
func New() (*Database, error) {
	// Get database connection details from environment variables
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}

	user := os.Getenv("DB_USER")
	if user == "" {
		return nil, fmt.Errorf("DB_USER not set in environment variables")
	}

	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		return nil, fmt.Errorf("DB_PASSWORD not set in environment variables")
	}

	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		return nil, fmt.Errorf("DB_NAME not set in environment variables")
	}

	// Create connection string for database connection
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		host, user, password, dbname, port)

	// Try to connect to the database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		// If database doesn't exist, try to create it
		if strings.Contains(err.Error(), "does not exist") {
			// Connect to PostgreSQL without specifying database name
			adminDSN := fmt.Sprintf("host=%s user=%s password=%s port=%s sslmode=disable TimeZone=UTC",
				host, user, password, port)
			adminDB, adminErr := gorm.Open(postgres.Open(adminDSN), &gorm.Config{})
			if adminErr != nil {
				return nil, fmt.Errorf("failed to connect to PostgreSQL server: %w", adminErr)
			}

			// Create the database
			sqlDB, _ := adminDB.DB()
			_, createErr := sqlDB.Exec(fmt.Sprintf("CREATE DATABASE \"%s\"", dbname))
			if createErr != nil {
				return nil, fmt.Errorf("failed to create database: %w", createErr)
			}

			// Close admin connection
			sqlDB.Close()

			// Now connect to the newly created database
			db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
			if err != nil {
				return nil, fmt.Errorf("failed to connect to newly created database: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to connect to database: %w", err)
		}
	}

	// Run migrations
	if err := db.AutoMigrate(&User{}, &TrackedCRN{}); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &Database{DB: db}, nil
}

// CreateUser creates a new user in the database
func (d *Database) CreateUser(telegramID int64, username string) (*User, error) {
	user := &User{
		TelegramID: telegramID,
		Username:   username,
		CreatedAt:  time.Now().Unix(),
	}

	result := d.DB.FirstOrCreate(user, User{TelegramID: telegramID})
	if result.Error != nil {
		return nil, result.Error
	}

	return user, nil
}

// GetUserByTelegramID retrieves a user by their Telegram ID
func (d *Database) GetUserByTelegramID(telegramID int64) (*User, error) {
	var user User
	result := d.DB.Where("telegram_id = ?", telegramID).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// GetUserByID retrieves a user by their ID
func (d *Database) GetUserByID(id int64) (*User, error) {
	var user User
	result := d.DB.Where("id = ?", id).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// AddTrackedCRN adds a CRN to track for a user
func (d *Database) AddTrackedCRN(userID int64, crn string, title string) (*TrackedCRN, error) {
	trackedCRN := &TrackedCRN{
		UserID:    userID,
		CRN:       crn,
		Title:     title,
		Active:    true,
		CreatedAt: time.Now().Unix(),
	}

	result := d.DB.FirstOrCreate(trackedCRN, TrackedCRN{UserID: userID, CRN: crn})
	if result.Error != nil {
		return nil, result.Error
	}

	// If the record already existed but was inactive, reactivate it
	if !trackedCRN.Active {
		result = d.DB.Model(trackedCRN).Update("active", true)
		if result.Error != nil {
			return nil, result.Error
		}
		trackedCRN.Active = true
	}

	return trackedCRN, nil
}

// RemoveTrackedCRN removes a CRN from tracking for a user
func (d *Database) RemoveTrackedCRN(userID int64, crn string) error {
	result := d.DB.Model(&TrackedCRN{}).Where("user_id = ? AND crn = ?", userID, crn).Update("active", false)
	return result.Error
}

// GetUserTrackedCRNs retrieves all active CRNs tracked by a user
func (d *Database) GetUserTrackedCRNs(userID int64) ([]TrackedCRN, error) {
	var crns []TrackedCRN
	result := d.DB.Where("user_id = ? AND active = ?", userID, true).Find(&crns)
	if result.Error != nil {
		return nil, result.Error
	}
	return crns, nil
}

// GetAllTrackedCRNs retrieves all active CRNs tracked by all users
func (d *Database) GetAllTrackedCRNs() ([]TrackedCRN, error) {
	var crns []TrackedCRN
	result := d.DB.Where("active = ?", true).Find(&crns)
	if result.Error != nil {
		return nil, result.Error
	}
	return crns, nil
}

// UpdateCRNTitle updates the title of a tracked CRN
func (d *Database) UpdateCRNTitle(userID int64, crn string, title string) error {
	result := d.DB.Model(&TrackedCRN{}).Where("user_id = ? AND crn = ?", userID, crn).Update("title", title)
	return result.Error
}
