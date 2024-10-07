package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents the user model
type User struct {
	ID        string         `gorm:"type:TEXT;primaryKey" json:"id" example:"550e8400-e29b-41d4-a716-446655440000" swaggertype:"string"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at" example:"2024-08-31T14:30:00Z" swaggertype:"string" format:"date-time"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at" example:"2024-08-31T14:30:00Z" swaggertype:"string" format:"date-time"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-" swaggerignore:"true"`
	Username  string         `gorm:"unique;not null" json:"username" example:"johndoe"`
	Password  string         `gorm:"not null" json:"-" swaggerignore:"true"`
	Email     string         `gorm:"unique;not null" json:"email" example:"john@example.com"`

	// Many-to-many relationship to represent the friends
	Friends []*User `gorm:"many2many:user_friends" json:"friends"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = uuid.New().String()
	return
}

// UserInput represents the structure for user input in registration
type UserInput struct {
	Username string `json:"username" binding:"required" example:"johndoe"`
	Password string `json:"password" binding:"required" example:"secret123"`
	Email    string `json:"email" binding:"required,email" example:"john@example.com"`
}

// LoginInput represents the structure for login input
type LoginInput struct {
	Username string `json:"username" binding:"required" example:"johndoe"`
	Password string `json:"password" binding:"required" example:"secret123"`
}
