package redis

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

var redisClent *redis.Client

func InitClient(ctx context.Context) {
	redisClent = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	_, err := redisClent.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Redis server couln't connect %v", err)
	}

}

func GetClient() *redis.Client {
	return redisClent
}
