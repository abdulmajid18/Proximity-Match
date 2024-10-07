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

	// Drop existing tables
	db.Exec("DROP TABLE IF EXISTS user_friends")
	db.Exec("DROP TABLE IF EXISTS users")

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
func TestRemoveFriend(t *testing.T) {
	db, cleanup := setupTestDB()
	defer cleanup()

	userRepo := repository.NewUserRepo(db)

	// Create two users
	user1 := &models.User{Username: "user1", Email: "user1@example.com"}
	user2 := &models.User{Username: "user2", Email: "user2@example.com"}

	err := userRepo.CreateUser(user1)
	if err != nil {
		t.Fatalf("Error creating user 1: %v", err)
	}
	err = userRepo.CreateUser(user2)
	if err != nil {
		t.Fatalf("Error creating user 2: %v", err)
	}

	// Retrieve saved users
	savedUser1, err := userRepo.FindByEmail(user1.Email)
	if err != nil {
		t.Fatalf("Error finding user 1 by email: %v", err)
	}
	savedUser2, err := userRepo.FindByEmail(user2.Email)
	if err != nil {
		t.Fatalf("Error finding user 2 by email: %v", err)
	}

	// Add friend relationship
	err = userRepo.AddFriend(savedUser1.ID, savedUser2.ID)
	if err != nil {
		t.Fatalf("Error adding friend: %v", err)
	}

	// Verify friendship was added
	friendsUser1, err := userRepo.GetFriends(savedUser1.ID)
	if err != nil {
		t.Fatalf("Error getting friends for user1: %v", err)
	}
	if len(friendsUser1) != 1 || friendsUser1[0].ID != savedUser2.ID {
		t.Fatalf("Expected user1's friend to be %s, but got %v", savedUser2.ID, friendsUser1)
	}

	// Remove friend relationship
	err = userRepo.RemoveFriend(savedUser1.ID, savedUser2.ID)
	if err != nil {
		t.Fatalf("Error removing friend: %v", err)
	}

	// Verify friendship was removed for user1
	friendsUser1AfterRemoval, err := userRepo.GetFriends(savedUser1.ID)
	if err != nil {
		t.Fatalf("Error getting friends for user1 after removal: %v", err)
	}
	if len(friendsUser1AfterRemoval) != 0 {
		t.Fatalf("Expected user1 to have no friends after removal, but got %v", friendsUser1AfterRemoval)
	}

	// Verify friendship was removed for user2
	friendsUser2AfterRemoval, err := userRepo.GetFriends(savedUser2.ID)
	if err != nil {
		t.Fatalf("Error getting friends for user2 after removal: %v", err)
	}
	if len(friendsUser2AfterRemoval) != 0 {
		t.Fatalf("Expected user2 to have no friends after removal, but got %v", friendsUser2AfterRemoval)
	}

	// Verify no friendship records exist
	friendships, err := userRepo.GetFriendshipRecords()
	if err != nil {
		t.Fatalf("Error getting friendship records: %v", err)
	}
	if len(friendships) != 0 {
		t.Fatalf("Expected 0 friendship records after removal, but got %d", len(friendships))
	}
}
