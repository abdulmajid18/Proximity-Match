package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// Define a struct to hold location data
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// WebSocket handler for location data
func handleLocationWebSocket(c *gin.Context) {
	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer conn.Close()

	// Handle WebSocket communication
	for {
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		// Decode the received message into a Location struct
		var location Location
		if err := json.Unmarshal(msg, &location); err != nil {
			conn.WriteMessage(websocket.TextMessage, []byte("Error decoding location data"))
			continue
		}

		// Process the location data (for example, log it)
		fmt.Printf("Received location: Latitude %f, Longitude %f\n", location.Latitude, location.Longitude)

		// Optionally, send a response back to the client
		response := "Location received"
		if err := conn.WriteMessage(messageType, []byte(response)); err != nil {
			break
		}
	}
}

func main() {
	r := gin.Default()

	// Define a route for WebSocket connections for location data
	r.GET("/location", handleLocationWebSocket)

	// Start the HTTP server
	r.Run(":8081")
}
