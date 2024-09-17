package redis

import (
	"context"
	"fmt"
	"log"
	"time"
	"websocket-server/internal/models"

	"github.com/redis/go-redis/v9"
)

type RedisCacheHandler interface {
	StoreLocation(location models.Location) (models.Location, error)
}

type RedisCache struct {
	ctx         context.Context
	redisClient *redis.Client
}

func NewRedisCache(ctx context.Context, redisClient *redis.Client) RedisCacheHandler {
	return &RedisCache{ctx: ctx, redisClient: redisClient}
}

// StoreLocation saves the location in Redis and returns the saved location
func (r *RedisCache) StoreLocation(location models.Location) (models.Location, error) {
	var err error
	currentLocation := redis.GeoLocation{
		Name:      location.UserId,
		Latitude:  location.CurrentLatitude,
		Longitude: location.CurrentLongitude,
	}
	destination := map[string]interface{}{
		"destination_lat": location.DestinationLatitude,
		"destination_lon": location.DestinationLongitude,
	}

	_, err = r.redisClient.GeoAdd(r.ctx, location.UserId, &currentLocation).Result()
	if err != nil {
		return models.Location{}, fmt.Errorf("could not store user in Redis: %w", err)
	}
	r.redisClient.Expire(r.ctx, location.UserId, 60*time.Second)

	err = r.redisClient.HSet(r.ctx, location.UserId, destination).Err()
	if err != nil {
		log.Fatalf("Error adding destination: %v", err)
	}

	log.Printf("Added user %s with current location and destination.\n", location.UserId)
	return location, nil
}

func (r *RedisCache) RefreshTTL(key string, ttl time.Duration, interval time.Duration, stopChan chan bool) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	select {
	case <-ticker.C:
		err := r.redisClient.Expire(r.ctx, key, ttl).Err()
		if err != nil {
			log.Printf("Could not refresh TTL for key %s : %v", key, err)
		} else {
			log.Printf("TTL refreshed for key %s", key)
		}
	case <-stopChan:
		log.Printf("Stopping TTL refresh for key %s", key)
		return
	}
}
