package main

import (
	"context"
	"log"
	"os"

	"websocket-server/database"
	"websocket-server/internal/handler"
	"websocket-server/internal/repository"
	"websocket-server/redis"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var ctx = context.Background()

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	keyspace := os.Getenv("CASSANDRA_KEYSPACE")
	if keyspace == "" {
		log.Fatalf("CASSANDRA_KEYSPACE environment variable is not set")
	}
	database.Init(keyspace)
	redis.InitClient(ctx)
	redisCache := redis.NewRedisCache(ctx, redis.GetClient())

	cassandraSession := database.GetSession()
	locationRepo := repository.NewLocationRepo(cassandraSession, keyspace)
	webSocketHandler := handler.NewWebSocketHandler(locationRepo, redisCache)
	handleLocationWebSocket := webSocketHandler.HandleWebSocket

	r := gin.Default()

	// Define a route for WebSocket connections for location data
	r.GET("/location", handleLocationWebSocket)

	// Start the HTTP server
	r.Run(":8081")
}
