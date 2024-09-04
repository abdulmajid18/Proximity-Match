package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
	"websocket-server/internal/models"
	"websocket-server/internal/repository"

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
}

func NewWebSocketHandler(repo repository.LocationRepository) *WebSocketHandler {
	return &WebSocketHandler{LocationRepo: repo}
}

func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Error while upgrading connection:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade connection"})
		return
	}
	defer conn.Close()

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
		if err := h.processMessage(conn, message); err != nil {
			log.Println("Error processing message:", err)
		}
	}
}

func (h *WebSocketHandler) processMessage(conn *websocket.Conn, message models.WebSocketMessage) error {
	var response models.WebSocketMessage
	var err error

	switch message.Action {
	case "create":
		return h.createLocation(message)
	case "update":
		return h.updateLocation(message)
	case "delete":
		return h.deleteLocation(message)
	case "update_destination":
		return h.updateDestination(message)
	case "update_current_location":
		return h.updateCurrentLocation(message)
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

func (h *WebSocketHandler) createLocation(message models.WebSocketMessage) error {
	randomUUID := uuid.New().String()
	location := models.Location{
		UserId:               randomUUID,
		CurrentLatitude:      message.Latitude,
		CurrentLongitude:     message.Longitude,
		DestinationLatitude:  message.DestinationLatitude,
		DestinationLongitude: message.DestinationLongitude,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}
	return h.LocationRepo.Create(location)
}

func (h *WebSocketHandler) updateLocation(message models.WebSocketMessage) error {
	location := models.Location{
		UserId:               message.UserID,
		CurrentLatitude:      message.Latitude,
		CurrentLongitude:     message.Longitude,
		DestinationLatitude:  message.DestinationLatitude,
		DestinationLongitude: message.DestinationLongitude,
		UpdatedAt:            time.Now(),
	}
	return h.LocationRepo.Update(location)
}

func (h *WebSocketHandler) deleteLocation(message models.WebSocketMessage) error {
	userIDParsed, err := uuid.Parse(message.UserID)
	if err != nil {
		return err
	}
	return h.LocationRepo.Delete(userIDParsed)
}

func (h *WebSocketHandler) updateDestination(message models.WebSocketMessage) error {
	userIDParsed, err := uuid.Parse(message.UserID)
	if err != nil {
		return err
	}
	return h.LocationRepo.UpdateDestination(userIDParsed, message.DestinationLatitude, message.DestinationLongitude)
}

func (h *WebSocketHandler) updateCurrentLocation(message models.WebSocketMessage) error {
	userIDParsed, err := uuid.Parse(message.UserID)
	if err != nil {
		return err
	}
	return h.LocationRepo.UpdateCurrentLocation(userIDParsed, message.Latitude, message.Longitude)
}

func (h *WebSocketHandler) getUserLocation(userId string) (models.WebSocketMessage, error) {
	location, err := h.LocationRepo.GetByUserID(userId)
	if err != nil {
		return models.WebSocketMessage{Error: "Failed to get location"}, err
	}
	return models.WebSocketMessage{
		UserID:               location.UserId,
		Latitude:             location.CurrentLatitude,
		Longitude:            location.CurrentLongitude,
		DestinationLatitude:  location.DestinationLatitude,
		DestinationLongitude: location.DestinationLongitude,
		CreatedAt:            location.CreatedAt,
		UpdatedAt:            location.UpdatedAt,
	}, nil
}
