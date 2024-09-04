package repository

import (
	"time"
	"websocket-server/internal/models"

	"github.com/gocql/gocql"
	"github.com/google/uuid"
)

type LocationRepository interface {
	Create(location models.Location) error
	GetByUserID(userID string) (models.Location, error)
	UpdateDestination(userID uuid.UUID, latitude float64, longitude float64) error
	UpdateCurrentLocation(userID uuid.UUID, latitude float64, longitude float64) error
	Update(location models.Location) error
	Delete(userID uuid.UUID) error
}

type LocationRepo struct {
	session  *gocql.Session
	keyspace string
}

func NewLocationRepo(session *gocql.Session, keyspace string) LocationRepository {
	return &LocationRepo{session: session, keyspace: keyspace}
}

func (r *LocationRepo) Create(location models.Location) error {
	query := `
        INSERT INTO ` + r.keyspace + `.locations (
            user_id, current_latitude, current_longitude, 
            destination_latitude, destination_longitude, created_at, updated_at
        ) VALUES (?, ?, ?, ?, ?, ?, ?)
    `
	err := r.session.Query(query,
		location.UserId,
		location.CurrentLatitude,
		location.CurrentLongitude,
		location.DestinationLatitude,
		location.DestinationLongitude,
		location.CreatedAt,
		location.UpdatedAt,
	).Exec()

	return err
}

func (r *LocationRepo) GetByUserID(userID string) (models.Location, error) {
	var location models.Location
	query := `
	SELECT user_id, current_latitude, current_longitude, destination_latitude, destination_longitude, created_at, updated_at
	FROM ` + r.keyspace + `.locations
	WHERE user_id = ?`

	err := r.session.Query(query, userID).Scan(
		&location.UserId,
		&location.CurrentLatitude,
		&location.CurrentLongitude,
		&location.DestinationLatitude,
		&location.DestinationLongitude,
		&location.CreatedAt,
		&location.UpdatedAt,
	)
	return location, err
}

func (r *LocationRepo) Update(location models.Location) error {
	err := r.session.Query(`
        UPDATE locations
        SET current_latitude = ?, current_longitude = ?, destination_latitude = ?, destination_longitude = ?, updated_at = ?
        WHERE user_id = ?`,
		location.CurrentLatitude,
		location.CurrentLongitude,
		location.DestinationLatitude,
		location.DestinationLongitude,
		time.Now(),
		location.UserId,
	).Exec()
	return err
}

func (r *LocationRepo) Delete(userID uuid.UUID) error {
	err := r.session.Query(`
        DELETE FROM locations
        WHERE user_id = ?`,
		userID,
	).Exec()
	return err
}

func (r *LocationRepo) UpdateDestination(userID uuid.UUID, latitude float64, longitude float64) error {
	err := r.session.Query(`
		UPDATE locations
		SET destination_latitude = ?, destination_longitude = ?, updated_at = ?
		WHERE user_id = ?`,
		latitude,
		longitude,
		time.Now(),
		userID,
	).Exec()
	return err
}

func (r *LocationRepo) UpdateCurrentLocation(userID uuid.UUID, latitude float64, longitude float64) error {
	err := r.session.Query(`
		UPDATE locations
		SET current_latitude = ?, current_longitude = ?, updated_at = ?
		WHERE user_id = ?`,
		latitude,
		longitude,
		time.Now(),
		userID,
	).Exec()
	return err
}
