package services

import (
	"errors"
	"matching-service/api-server/internal/auth"
	"matching-service/api-server/internal/models"
	"matching-service/api-server/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(user *models.User) error
	Login(username, password string) (string, error)
	GetUserByUsername(username string) (*models.User, error)
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) Register(user *models.User) error {
	// Check if user already exists
	existingUser, _ := s.userRepo.FindByUsername(user.Username)
	if existingUser != nil {
		return errors.New("username already exists")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	// Create the user
	return s.userRepo.CreateUser(user)
}

func (s *userService) Login(username, password string) (string, error) {
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("invalid credentials")
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user.Username)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *userService) GetUserByUsername(username string) (*models.User, error) {
	return s.userRepo.FindByUsername(username)
}
