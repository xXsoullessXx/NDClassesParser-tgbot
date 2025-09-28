package database_test

import (
	"errors"
	"testing"

	"NDClasses/clients/database"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *database.Database {
	// Use in-memory SQLite for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Run migrations
	if err := db.AutoMigrate(&database.User{}, &database.TrackedCRN{}); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return &database.Database{DB: db}
}

func TestCreateUser(t *testing.T) {
	db := setupTestDB(t)

	// Test creating a new user
	user, err := db.CreateUser(12345, "testuser")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if user.TelegramID != 12345 {
		t.Errorf("Expected TelegramID 12345, got %d", user.TelegramID)
	}

	if user.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", user.Username)
	}

	// Test creating the same user again (should return existing)
	user2, err := db.CreateUser(12345, "testuser2")
	if err != nil {
		t.Fatalf("Failed to get existing user: %v", err)
	}

	if user2.ID != user.ID {
		t.Errorf("Expected same user ID, got %d vs %d", user2.ID, user.ID)
	}

	// Username should not change on second create
	if user2.Username != "testuser" {
		t.Errorf("Expected username to remain 'testuser', got '%s'", user2.Username)
	}
}

func TestGetUserByTelegramID(t *testing.T) {
	db := setupTestDB(t)

	// Create a user first
	createdUser, err := db.CreateUser(12345, "testuser")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Test getting the user
	user, err := db.GetUserByTelegramID(12345)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	if user.ID != createdUser.ID {
		t.Errorf("Expected user ID %d, got %d", createdUser.ID, user.ID)
	}

	if user.TelegramID != 12345 {
		t.Errorf("Expected TelegramID 12345, got %d", user.TelegramID)
	}

	// Test getting non-existent user
	_, err = db.GetUserByTelegramID(99999)
	if err == nil || !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("Expected ErrRecordNotFound when getting non-existent user, got: %v", err)
	}
}

func TestGetUserByID(t *testing.T) {
	db := setupTestDB(t)

	// Create a user first
	createdUser, err := db.CreateUser(12345, "testuser")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Test getting the user by ID
	user, err := db.GetUserByID(createdUser.ID)
	if err != nil {
		t.Fatalf("Failed to get user by ID: %v", err)
	}

	if user.ID != createdUser.ID {
		t.Errorf("Expected user ID %d, got %d", createdUser.ID, user.ID)
	}

	// Test getting non-existent user
	_, err = db.GetUserByID(99999)
	if err == nil || !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("Expected ErrRecordNotFound when getting non-existent user by ID, got: %v", err)
	}
}

func TestAddTrackedCRN(t *testing.T) {
	db := setupTestDB(t)

	// Create a user first
	user, err := db.CreateUser(12345, "testuser")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Test adding a CRN
	trackedCRN, err := db.AddTrackedCRN(user.ID, "12345", "Test Class")
	if err != nil {
		t.Fatalf("Failed to add tracked CRN: %v", err)
	}

	if trackedCRN.UserID != user.ID {
		t.Errorf("Expected UserID %d, got %d", user.ID, trackedCRN.UserID)
	}

	if trackedCRN.CRN != "12345" {
		t.Errorf("Expected CRN '12345', got '%s'", trackedCRN.CRN)
	}

	if trackedCRN.Title != "Test Class" {
		t.Errorf("Expected title 'Test Class', got '%s'", trackedCRN.Title)
	}

	if !trackedCRN.Active {
		t.Error("Expected CRN to be active")
	}

	// Test adding the same CRN again (should return existing without updating title)
	trackedCRN2, err := db.AddTrackedCRN(user.ID, "12345", "Updated Title")
	if err != nil {
		t.Fatalf("Failed to add existing CRN: %v", err)
	}

	if trackedCRN2.ID != trackedCRN.ID {
		t.Errorf("Expected same CRN ID, got %d vs %d", trackedCRN2.ID, trackedCRN.ID)
	}

	// Title should NOT be updated (current behavior)
	if trackedCRN2.Title != "Test Class" {
		t.Errorf("Expected title to remain 'Test Class', got '%s'", trackedCRN2.Title)
	}
}

func TestRemoveTrackedCRN(t *testing.T) {
	db := setupTestDB(t)

	// Create a user and add a CRN
	user, err := db.CreateUser(12345, "testuser")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	_, err = db.AddTrackedCRN(user.ID, "12345", "Test Class")
	if err != nil {
		t.Fatalf("Failed to add tracked CRN: %v", err)
	}

	// Test removing the CRN
	err = db.RemoveTrackedCRN(user.ID, "12345")
	if err != nil {
		t.Fatalf("Failed to remove tracked CRN: %v", err)
	}

	// Check that it's no longer active
	crns, err := db.GetUserTrackedCRNs(user.ID)
	if err != nil {
		t.Fatalf("Failed to get user tracked CRNs: %v", err)
	}

	if len(crns) != 0 {
		t.Errorf("Expected 0 tracked CRNs after removal, got %d", len(crns))
	}
}

func TestGetUserTrackedCRNs(t *testing.T) {
	db := setupTestDB(t)

	// Create a user
	user, err := db.CreateUser(12345, "testuser")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Add multiple CRNs
	_, err = db.AddTrackedCRN(user.ID, "12345", "Class 1")
	if err != nil {
		t.Fatalf("Failed to add CRN 1: %v", err)
	}

	_, err = db.AddTrackedCRN(user.ID, "67890", "Class 2")
	if err != nil {
		t.Fatalf("Failed to add CRN 2: %v", err)
	}

	// Test getting tracked CRNs
	crns, err := db.GetUserTrackedCRNs(user.ID)
	if err != nil {
		t.Fatalf("Failed to get user tracked CRNs: %v", err)
	}

	if len(crns) != 2 {
		t.Errorf("Expected 2 tracked CRNs, got %d", len(crns))
	}

	// Check CRNs are present
	foundCRN1 := false
	foundCRN2 := false
	for _, crn := range crns {
		if crn.CRN == "12345" && crn.Title == "Class 1" {
			foundCRN1 = true
		}
		if crn.CRN == "67890" && crn.Title == "Class 2" {
			foundCRN2 = true
		}
	}

	if !foundCRN1 {
		t.Error("CRN 12345 not found in tracked CRNs")
	}

	if !foundCRN2 {
		t.Error("CRN 67890 not found in tracked CRNs")
	}
}

func TestGetAllTrackedCRNs(t *testing.T) {
	db := setupTestDB(t)

	// Create two users
	user1, err := db.CreateUser(12345, "user1")
	if err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}

	user2, err := db.CreateUser(67890, "user2")
	if err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	// Add CRNs for both users
	_, err = db.AddTrackedCRN(user1.ID, "11111", "Class A")
	if err != nil {
		t.Fatalf("Failed to add CRN for user1: %v", err)
	}

	_, err = db.AddTrackedCRN(user2.ID, "22222", "Class B")
	if err != nil {
		t.Fatalf("Failed to add CRN for user2: %v", err)
	}

	// Test getting all tracked CRNs
	crns, err := db.GetAllTrackedCRNs()
	if err != nil {
		t.Fatalf("Failed to get all tracked CRNs: %v", err)
	}

	if len(crns) != 2 {
		t.Errorf("Expected 2 tracked CRNs total, got %d", len(crns))
	}
}

func TestUpdateCRNTitle(t *testing.T) {
	db := setupTestDB(t)

	// Create a user and add a CRN
	user, err := db.CreateUser(12345, "testuser")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	_, err = db.AddTrackedCRN(user.ID, "12345", "Original Title")
	if err != nil {
		t.Fatalf("Failed to add tracked CRN: %v", err)
	}

	// Update the title
	err = db.UpdateCRNTitle(user.ID, "12345", "Updated Title")
	if err != nil {
		t.Fatalf("Failed to update CRN title: %v", err)
	}

	// Verify the title was updated
	crns, err := db.GetUserTrackedCRNs(user.ID)
	if err != nil {
		t.Fatalf("Failed to get user tracked CRNs: %v", err)
	}

	if len(crns) != 1 {
		t.Errorf("Expected 1 tracked CRN, got %d", len(crns))
	}

	if crns[0].Title != "Updated Title" {
		t.Errorf("Expected title 'Updated Title', got '%s'", crns[0].Title)
	}
}
