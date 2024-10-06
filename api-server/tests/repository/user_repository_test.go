package repository

import (
	"matching-service/api-server/internal/models"
	"matching-service/api-server/internal/repository"
	"testing"

	"math/rand"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func GenerateRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func setupTestDB() (*gorm.DB, func()) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	err = db.AutoMigrate(&models.User{})
	if err != nil {
		panic("failed to migrate database")
	}

	cleanup := func() {
		sqlDB, err := db.DB()
		if err != nil {
			panic("failed to get database connection")
		}
		sqlDB.Close()
	}

	return db, cleanup
}

func TestCreateUser(t *testing.T) {
	db, cleanup := setupTestDB()
	defer cleanup()

	userRepo := repository.NewUserRepo(db)

	username := "user_" + GenerateRandomString(8)
	email := GenerateRandomString(10) + "@example.com"

	user := &models.User{
		Username: username,
		Password: "password123",
		Email:    email,
	}

	err := userRepo.CreateUser(user)
	if err != nil {
		t.Fatalf("Error creating user: %v", err)
	}

	createdUser, err := userRepo.FindByUsername(user.Username)
	if err != nil {
		t.Fatalf("Error finding user by username: %v", err)
	}
	if createdUser == nil {
		t.Fatalf("User with username %s not found", user.Username)
	}
	if createdUser.Email != user.Email {
		t.Errorf("Expected email %s, but got %s", user.Email, createdUser.Email)
	}
}

func TestFindByEmail(t *testing.T) {
	db, cleanup := setupTestDB()
	defer cleanup()

	userRepo := repository.NewUserRepo(db)

	username := "user_" + GenerateRandomString(8)
	email := GenerateRandomString(10) + "@example.com"

	user := &models.User{
		Username: username,
		Password: "password123",
		Email:    email,
	}

	db.Create(user)

	foundUser, err := userRepo.FindByEmail(email)
	if err != nil {
		t.Fatalf("Error finding user by email: %v", err)
	}
	if foundUser == nil {
		t.Fatalf("User with email %s not found", email)
	}
	if foundUser.Email != email {
		t.Errorf("Expected email %s, but got %s", email, foundUser.Email)
	}
}
