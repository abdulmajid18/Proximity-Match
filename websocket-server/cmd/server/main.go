package main

import (
	"log"
	"os"
	"websocket-server/internal/handler"
	"websocket-server/internal/repository"
	"websocket-server/pkg/database"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

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

	cassandraSession := database.GetSession()
	locationRepo := repository.NewLocationRepo(cassandraSession, keyspace)
	webSocketHandler := handler.NewWebSocketHandler(locationRepo)
	handleLocationWebSocket := webSocketHandler.HandleWebSocket

	r := gin.Default()

	// Define a route for WebSocket connections for location data
	r.GET("/location", handleLocationWebSocket)

	// Start the HTTP server
	r.Run(":8081")
}
