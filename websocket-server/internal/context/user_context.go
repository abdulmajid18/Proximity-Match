package context

import "websocket-server/internal/models"

type UserContext struct {
	UserID   string
	Location models.Location
}
