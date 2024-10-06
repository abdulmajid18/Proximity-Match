package main

import (
	"context"
	"log"
	"matching-service/websocket-server/internal/handler"
	"matching-service/websocket-server/internal/repository"
	"matching-service/websocket-server/pkg/database"
	"matching-service/websocket-server/pkg/redis"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Initialize database
	keyspace := os.Getenv("CASSANDRA_KEYSPACE")
	if keyspace == "" {
		log.Fatalf("CASSANDRA_KEYSPACE environment variable is not set")
	}
	database.Init(keyspace)

	// Initialize Redis
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	redis.InitClient(context.Background(), redisPort, redisHost)

	redisClient := redis.GetClient()
	redisCache := redis.NewRedisCache(context.Background(), redisClient)

	// Create repository and handler
	cassandraSession := database.GetSession()
	locationRepo := repository.NewLocationRepo(cassandraSession, keyspace)
	webSocketHandler := handler.NewWebSocketHandler(locationRepo, redisCache)

	// Initialize Gin router
	r := gin.Default()
	r.GET("/location", webSocketHandler.HandleWebSocket)

	// Start the HTTP server
	r.Run(":8081")
}
