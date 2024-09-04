package models

import (
	"time"
)

type Location struct {
	UserId               string    `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000" swaggertype:"string"`
	CurrentLatitude      float64   `json:"current_latitude"`
	CurrentLongitude     float64   `json:"current_longitude"`
	DestinationLatitude  float64   `json:"destination_latitude"`
	DestinationLongitude float64   `json:"destination_longitude"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type WebSocketMessage struct {
	Action               string    `json:"action"` // Create, Update, Delete, etc.
	UserID               string    `json:"user_id,omitempty"`
	Latitude             float64   `json:"current_latitude,omitempty"`
	Longitude            float64   `json:"current_longitude,omitempty"`
	DestinationLatitude  float64   `json:"destination_latitude,omitempty"`
	DestinationLongitude float64   `json:"destination_longitude,omitempty"`
	CreatedAt            time.Time `json:"created_at,omitempty"`
	UpdatedAt            time.Time `json:"updated_at,omitempty"`
	Error                string    `json:"error,omitempty"`
}
