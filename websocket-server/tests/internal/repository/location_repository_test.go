package repository

import (
	"fmt"
	"log"
	"matching-service/websocket-server/internal/models"
	"matching-service/websocket-server/internal/repository"
	"matching-service/websocket-server/pkg/database"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func TestCreateLocationIntegration(t *testing.T) {
	if err := godotenv.Load(".env.test"); err != nil {
		log.Fatalf("Error loading .env.test file")
	}

	// Set up real Cassandra session
	keyspace := os.Getenv("CASSANDRA_KEYSPACE")
	fmt.Println(keyspace)
	database.Init(keyspace)
	session := database.GetSession()
	// Create the repo with the real session
	repo := repository.NewLocationRepo(session, keyspace)

	userId := uuid.New().String()
	// Create a sample location
	location := models.Location{
		UserId:               userId,
		CurrentLatitude:      37.7749,
		CurrentLongitude:     -122.4194,
		DestinationLatitude:  40.7128,
		DestinationLongitude: -74.0060,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	// Test the Create method
	err := repo.Create(location)
	if err != nil {
		t.Fatalf("Failed to create location: %v", err)
	}

	// Optionally, retrieve and validate the record
	savedLocation, err := repo.GetByUserID(userId)
	if err != nil {
		t.Fatalf("Failed to get location: %v", err)
	}

	if savedLocation.UserId != location.UserId {
		t.Errorf("Expected user ID %v, got %v", location.UserId, savedLocation.UserId)
	}
}
