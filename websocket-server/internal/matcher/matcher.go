package matcher

import (
	"context"
	"fmt"
	"log"
	"matching-service/websocket-server/internal/models"
	"matching-service/websocket-server/internal/repository"
	"math"

	"github.com/gocql/gocql"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type MatcherService struct {
	cassandraRepo repository.LocationRepository
	redisClient   *redis.Client
	ctx           context.Context
}

func NewMatcherService(cassandraRepo repository.LocationRepository, redisClient *redis.Client) *MatcherService {
	return &MatcherService{
		cassandraRepo: cassandraRepo,
		redisClient:   redisClient,
		ctx:           context.Background(),
	}
}

func (s *MatcherService) SyncLocationToRedis() error {
	locations, err := s.cassandraRepo.GetAllLocations()
	if err != nil {
		return fmt.Errorf("failed to fetch locations from Cassandra: %v", err)
	}

	// Add all locations to Redis
	pipe := s.redisClient.Pipeline()
	for _, loc := range locations {
		key := loc.UserId
		pipe.GeoAdd(s.ctx, "user_locations", &redis.GeoLocation{
			Name:      key,
			Longitude: loc.CurrentLongitude,
			Latitude:  loc.CurrentLatitude,
		})
		// Store additional user data
		pipe.HSet(s.ctx, key, map[string]interface{}{
			"destination_lat": loc.DestinationLatitude,
			"destination_lon": loc.DestinationLongitude,
		})
	}
	_, err = pipe.Exec(s.ctx)
	if err != nil {
		return fmt.Errorf("failed to sync locations to Redis: %v", err)
	}

	log.Printf("Synced %d locations to Redis", len(locations))
	return nil
}

func (s *MatcherService) FindPossibleMatches(userID string, radius float64) ([]models.Location, error) {
	// Get user's current location from Redis
	key := userID
	pos, err := s.redisClient.GeoPos(s.ctx, "user_locations", key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get user location: %v", err)
	}
	if len(pos) == 0 || pos[0] == nil {
		return nil, fmt.Errorf("user location not found")
	}

	// Find nearby users
	nearby, err := s.redisClient.GeoRadius(s.ctx, "user_locations", pos[0].Longitude, pos[0].Latitude, &redis.GeoRadiusQuery{
		Radius:    radius,
		Unit:      "km",
		WithCoord: true,
		WithDist:  true,
		Sort:      "ASC",
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to find nearby users: %v", err)
	}

	// Get user's destination
	userDest, err := s.redisClient.HMGet(s.ctx, key, "destination_lat", "destination_lon").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get user destination: %v", err)
	}

	matches := []models.Location{}
	for _, loc := range nearby {
		if loc.Name == key {
			continue // Skip the user themselves
		}

		// Get potential match's destination
		matchDest, err := s.redisClient.HMGet(s.ctx, loc.Name, "destination_lat", "destination_lon").Result()
		if err != nil {
			log.Printf("Failed to get destination for user %s: %v", loc.Name, err)
			continue
		}

		// Compare destinations
		if isDestinationMatch(userDest, matchDest) {
			uid, _ := gocql.ParseUUID(loc.Name) // Remove "user:" prefix
			matches = append(matches, models.Location{
				UserId:               uid.String(),
				CurrentLatitude:      loc.Latitude,
				CurrentLongitude:     loc.Longitude,
				DestinationLatitude:  parseFloat(matchDest[0]),
				DestinationLongitude: parseFloat(matchDest[1]),
			})
		}
	}

	return matches, nil
}

func isDestinationMatch(dest1, dest2 []interface{}) bool {
	const tolerance = 0.01 // Adjust as needed
	return math.Abs(parseFloat(dest1[0])-parseFloat(dest2[0])) <= tolerance &&
		math.Abs(parseFloat(dest1[1])-parseFloat(dest2[1])) <= tolerance
}

func parseFloat(v interface{}) float64 {
	f, _ := v.(float64)
	return f
}

func (s *MatcherService) UpdateUserLocation(userID string, lat, lon float64) error {
	// Update Cassandra
	uid, _ := uuid.Parse(userID)
	err := s.cassandraRepo.UpdateCurrentLocation(uid, lat, lon)
	if err != nil {
		return fmt.Errorf("failed to update location in Cassandra: %v", err)
	}

	// Update Redis
	key := fmt.Sprintf("user:%s", userID)
	_, err = s.redisClient.GeoAdd(s.ctx, "user_locations", &redis.GeoLocation{
		Name:      key,
		Latitude:  lat,
		Longitude: lon,
	}).Result()
	if err != nil {
		return fmt.Errorf("failed to update location in Redis: %v", err)
	}

	return nil
}
