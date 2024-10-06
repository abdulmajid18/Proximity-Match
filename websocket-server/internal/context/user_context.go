package context

import "matching-service/websocket-server/internal/models"

type UserContext struct {
	UserID   string
	Location models.Location
}
