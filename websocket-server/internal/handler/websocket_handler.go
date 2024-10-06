package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"matching-service/websocket-server/internal/context"
	"matching-service/websocket-server/internal/models"
	"matching-service/websocket-server/internal/repository"
	"matching-service/websocket-server/pkg/redis"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type WebSocketHandler struct {
	LocationRepo repository.LocationRepository
	Cache        redis.RedisCacheHandler
}

func NewWebSocketHandler(repo repository.LocationRepository, cache redis.RedisCacheHandler) *WebSocketHandler {
	return &WebSocketHandler{LocationRepo: repo, Cache: cache}
}

func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Error while upgrading connection:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade connection"})
		return
	}
	defer conn.Close()

	userID := uuid.New().String()

	userContext := context.UserContext{
		UserID: userID,
	}

	log.Printf("Using dummy user ID: %s", userID)
	stopChan := make(chan bool)
	go h.Cache.RefreshTTL(userID, 60*time.Second, 30*time.Second, stopChan)

	for {
		log.Println("Websocket Server Reading mesaaage")
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error while reading message:", err)
			break
		}

		log.Println("Websocket Server Unmarshalling the message")
		var message models.WebSocketMessage
		if err := json.Unmarshal(msg, &message); err != nil {
			log.Println("Error while unmarshalling message:", err)
			continue
		}

		log.Println("Websocket server processing the message by relaying to the right handler")
		if err := h.processMessage(conn, message, &userContext); err != nil {
			log.Println("Error processing message:", err)
		}
	}
	stopChan <- true
}

func (h *WebSocketHandler) processMessage(conn *websocket.Conn, message models.WebSocketMessage, userContext *context.UserContext) error {
	var response models.WebSocketMessage
	var err error

	switch message.Action {
	case "create":
		return h.createLocation(message, userContext)
	case "update":
		return h.updateLocation(message, userContext)
	case "delete":
		return h.deleteLocation(userContext)
	case "update_destination":
		return h.updateDestination(message, userContext)
	case "update_current_location":
		return h.updateCurrentLocation(message, userContext)
	case "get_location":
		response, err = h.getUserLocation(message.UserID)
		if err != nil {
			log.Println("Error getting user location:", err)
			response = models.WebSocketMessage{Error: "Failed to get location"}
		}
		log.Printf("Message: %+v\n", response)
		err = conn.WriteJSON(response)
		if err != nil {
			log.Println("Error sending response:", err)
			return err
		}
	default:
		log.Println("Unknown action:", message.Action)
		return nil
	}

	return nil
}

func (h *WebSocketHandler) createLocation(message models.WebSocketMessage, userContext *context.UserContext) error {
	location := models.Location{
		UserId:               userContext.UserID,
		CurrentLatitude:      message.Latitude,
		CurrentLongitude:     message.Longitude,
		DestinationLatitude:  message.DestinationLatitude,
		DestinationLongitude: message.DestinationLongitude,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	err := h.LocationRepo.Create(location)
	if err != nil {
		return err
	}
	go h.cacheLocation(location)
	userContext.Location = location

	return nil
}

func (h *WebSocketHandler) updateLocation(message models.WebSocketMessage, userContext *context.UserContext) error {
	location := models.Location{
		UserId:               userContext.UserID,
		CurrentLatitude:      message.Latitude,
		CurrentLongitude:     message.Longitude,
		DestinationLatitude:  message.DestinationLatitude,
		DestinationLongitude: message.DestinationLongitude,
		UpdatedAt:            time.Now(),
	}
	err := h.LocationRepo.Update(location)
	if err != nil {
		return err
	}
	go h.cacheLocation(location)
	userContext.Location = location

	return nil
}

func (h *WebSocketHandler) deleteLocation(userContext *context.UserContext) error {
	userIDParsed, err := uuid.Parse(userContext.UserID)
	if err != nil {
		return err
	}
	return h.LocationRepo.Delete(userIDParsed)
}

func (h *WebSocketHandler) updateDestination(message models.WebSocketMessage, userContext *context.UserContext) error {
	userIDParsed, err := uuid.Parse(userContext.UserID)
	if err != nil {
		return err
	}
	return h.LocationRepo.UpdateDestination(userIDParsed, message.DestinationLatitude, message.DestinationLongitude)
}

func (h *WebSocketHandler) updateCurrentLocation(message models.WebSocketMessage, userContext *context.UserContext) error {
	userIDParsed, err := uuid.Parse(userContext.UserID)
	if err != nil {
		return err
	}
	return h.LocationRepo.UpdateCurrentLocation(userIDParsed, message.Latitude, message.Longitude)
}

func (h *WebSocketHandler) getUserLocation(userId string) (models.WebSocketMessage, error) {
	location, err := h.getLocationFromCacheOrDB(userId)
	if err != nil {
		return models.WebSocketMessage{}, fmt.Errorf("failed to get location: %w", err)
	}

	return h.locationToWebSocketMessage(location), nil
}

func (h *WebSocketHandler) getLocationFromCacheOrDB(userId string) (models.Location, error) {
	location, err := h.Cache.Getlocation(userId)
	if err == nil {
		return location, nil
	}

	location, err = h.LocationRepo.GetByUserID(userId)
	if err != nil {
		return models.Location{}, fmt.Errorf("failed to get location from database: %w", err)
	}

	go h.cacheLocation(location)

	return location, nil
}

func (h *WebSocketHandler) cacheLocation(location models.Location) {
	_, err := h.Cache.StoreLocation(location)
	if err != nil {
		log.Printf("Failed to cache location for user %s: %v", location.UserId, err)
	}
}

func (h *WebSocketHandler) locationToWebSocketMessage(location models.Location) models.WebSocketMessage {
	return models.WebSocketMessage{
		UserID:               location.UserId,
		Latitude:             location.CurrentLatitude,
		Longitude:            location.CurrentLongitude,
		DestinationLatitude:  location.DestinationLatitude,
		DestinationLongitude: location.DestinationLongitude,
		CreatedAt:            location.CreatedAt,
		UpdatedAt:            location.UpdatedAt,
	}
}
