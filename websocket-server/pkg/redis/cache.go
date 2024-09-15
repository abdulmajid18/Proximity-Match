package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"matching-service/websocket-server/internal/models"
	"time"

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
	locationJson, err := json.Marshal(location)
	if err != nil {
		return models.Location{}, fmt.Errorf("could not marshal location: %w", err)
	}

	err = r.redisClient.Set(r.ctx, location.UserId, locationJson, 60*time.Second).Err()
	if err != nil {
		return models.Location{}, fmt.Errorf("could not store user in Redis: %w", err)
	}

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
