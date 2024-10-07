package repository

import (
	"fmt"
	"log"
	"matching-service/api-server/internal/models"

	"gorm.io/gorm"
)

type UserRepository interface {
	FindByUsername(username string) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	CreateUser(user *models.User) error
	AddFriend(userID, friendID string) error
	RemoveFriend(userID, friendID string) error
	GetFriends(userID string) ([]models.User, error)
	GetFriendshipRecords() ([]struct {
		UserID   string
		FriendID string
	}, error)
}

type userRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) FindByEmail(email string) (*models.User, error) {
	var user models.User
	result := r.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &user, nil
}

func (r *userRepo) FindByUsername(username string) (*models.User, error) {
	var user models.User
	result := r.db.Where("username = ?", username).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &user, nil
}

func (r *userRepo) CreateUser(user *models.User) error {
	result := r.db.Create(user)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *userRepo) RemoveFriend(userID, friendID string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var user, friend models.User

		if err := tx.First(&user, "id = ?", userID).Error; err != nil {
			return fmt.Errorf("error finding user: %w", err)
		}

		if err := tx.First(&friend, "id = ?", friendID).Error; err != nil {
			return fmt.Errorf("error finding friend: %w", err)
		}

		// Remove friend from user
		if err := tx.Exec("DELETE FROM user_friends WHERE user_id = ? AND friend_id = ?", userID, friendID).Error; err != nil {
			return fmt.Errorf("error removing friend from user: %w", err)
		}

		// Remove user from friend (for bidirectional friendship)
		if err := tx.Exec("DELETE FROM user_friends WHERE user_id = ? AND friend_id = ?", friendID, userID).Error; err != nil {
			return fmt.Errorf("error removing user from friend: %w", err)
		}

		log.Printf("Successfully removed friendship between %s and %s", userID, friendID)
		return nil
	})
}

func (r *userRepo) AddFriend(userID, friendID string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var user, friend models.User

		if err := tx.First(&user, "id = ?", userID).Error; err != nil {
			return fmt.Errorf("error finding user: %w", err)
		}

		if err := tx.First(&friend, "id = ?", friendID).Error; err != nil {
			return fmt.Errorf("error finding friend: %w", err)
		}

		// Add friend to user
		if err := tx.Exec("INSERT INTO user_friends (user_id, friend_id) VALUES (?, ?)", userID, friendID).Error; err != nil {
			return fmt.Errorf("error adding friend to user: %w", err)
		}

		// Add user to friend (for bidirectional friendship)
		if err := tx.Exec("INSERT INTO user_friends (user_id, friend_id) VALUES (?, ?)", friendID, userID).Error; err != nil {
			return fmt.Errorf("error adding user to friend: %w", err)
		}

		log.Printf("Successfully added friendship between %s and %s", userID, friendID)
		return nil
	})
}

func (r *userRepo) GetFriends(userID string) ([]models.User, error) {
	var user models.User

	if err := r.db.Preload("Friends").First(&user, "id = ?", userID).Error; err != nil {
		return nil, fmt.Errorf("error finding user: %w", err)
	}

	log.Printf("User found: %v, Friends count: %d", user, len(user.Friends))

	friends := make([]models.User, len(user.Friends))
	for i, friend := range user.Friends {
		friends[i] = *friend
	}

	return friends, nil
}

func (r *userRepo) GetFriendshipRecords() ([]struct {
	UserID   string
	FriendID string
}, error) {
	var friendships []struct {
		UserID   string
		FriendID string
	}
	err := r.db.Table("user_friends").Select("user_id", "friend_id").Scan(&friendships).Error
	if err != nil {
		return nil, fmt.Errorf("error retrieving friendship records: %w", err)
	}
	return friendships, nil
}
