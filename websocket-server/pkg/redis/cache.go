package redis

import (
	"context"
	"fmt"
	"log"
	"matching-service/websocket-server/internal/models"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCacheHandler interface {
	StoreLocation(location models.Location) (models.Location, error)
	Getlocation(key string) (models.Location, error)
	RefreshTTL(key string, ttl time.Duration, interval time.Duration, stopChan chan bool)
}

type RedisCache struct {
	ctx         context.Context
	redisClient *redis.Client
}

func NewRedisCache(ctx context.Context, redisClient *redis.Client) RedisCacheHandler {
	return &RedisCache{ctx: ctx, redisClient: redisClient}
}

func (r *RedisCache) Getlocation(key string) (models.Location, error) {
	location := models.Location{}
	geoKey := "geo:" + key
	geoLocation, err := r.redisClient.GeoPos(r.ctx, geoKey, key).Result()
	if err == redis.Nil {
		return location, fmt.Errorf("user %s not found in Redis", key)
	} else if err != nil {
		return location, fmt.Errorf("error getting user location from Redis: %w", err)
	}

	if len(geoLocation) == 0 || geoLocation[0] == nil {
		return location, fmt.Errorf("no geolocation data found for user %s", key)
	}
	location.UserId = key
	location.CurrentLatitude = geoLocation[0].Latitude
	location.CurrentLongitude = geoLocation[0].Longitude

	destKey := "dest:" + location.UserId
	destination, err := r.redisClient.HMGet(r.ctx, destKey, "destination_lat", "destination_lon").Result()
	if err == redis.Nil {
		log.Printf("Destination for user %s not found in Redis\n", key)
	} else if err != nil {
		return location, fmt.Errorf("error getting user destination from Redis: %w", err)
	} else if len(destination) == 2 && destination[0] != nil && destination[1] != nil {
		location.DestinationLatitude = parseFloat(destination[0])
		location.DestinationLongitude = parseFloat(destination[1])
	}

	return location, nil
}

func parseFloat(v interface{}) float64 {
	f, _ := v.(float64)
	return f
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
	geoKey := "geo:" + location.UserId

	_, err = r.redisClient.GeoAdd(r.ctx, geoKey, &currentLocation).Result()
	if err != nil {
		return models.Location{}, fmt.Errorf("could not store user in Redis: %w", err)
	}
	r.redisClient.Expire(r.ctx, location.UserId, 60*time.Second)

	destKey := "dest:" + location.UserId
	err = r.redisClient.HSet(r.ctx, destKey, destination).Err()
	if err != nil {
		return models.Location{}, fmt.Errorf("error adding destination: %w", err)
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
