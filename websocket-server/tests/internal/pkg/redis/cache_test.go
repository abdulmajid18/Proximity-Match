package redis

import (
	"context"
	"fmt"
	"log"
	"matching-service/websocket-server/internal/models"
	"matching-service/websocket-server/internal/repository"
	"matching-service/websocket-server/pkg/database"
	"matching-service/websocket-server/pkg/redis"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func TestCache(t *testing.T) {
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

	// redis connection
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	redis.InitClient(context.Background(), redisPort, redisHost)

	redisClient := redis.GetClient()
	redisCache := redis.NewRedisCache(context.Background(), redisClient)

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
	savedLoc, err := redisCache.StoreLocation(location)
	if err != nil {
		t.Fatalf("Failed to cache location: %v", err)
	}
	if savedLoc.UserId != location.UserId {
		t.Errorf("Expected user ID %v, got %v", location.UserId, savedLoc.UserId)
	}

	// Optionally, retrieve and validate the record
	savedLocation, err := repo.GetByUserID(userId)
	if err != nil {
		t.Fatalf("Failed to get location: %v", err)
	}
	if savedLocation.UserId != location.UserId {
		t.Errorf("Expected user ID %v, got %v", location.UserId, savedLocation.UserId)
	}

	savedCacheLocation, err := redisCache.Getlocation(savedLoc.UserId)
	if err != nil {
		t.Fatalf("Failed to retrieve from cache location: %v", err)
	}
	if savedCacheLocation.UserId != location.UserId {
		t.Errorf("Expected user ID from cache %v, got %v", location.UserId, savedLoc.UserId)
	}
}
