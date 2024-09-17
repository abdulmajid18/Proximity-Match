package redis

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

var redisClent *redis.Client

func InitClient(ctx context.Context, port string, host string) {
	address := fmt.Sprintf("%s:%s", host, port)
	redisClent = redis.NewClient(&redis.Options{
		Addr:     address,
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
