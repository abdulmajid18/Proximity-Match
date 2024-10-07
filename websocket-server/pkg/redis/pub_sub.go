package redis

import (
	"encoding/json"
	"fmt"
	"log"
	"matching-service/websocket-server/internal/models"
)

func (r *RedisCache) AddFriend(userId, friendId string) error {
	key := fmt.Sprintf("friends:%s", userId)
	_, err := r.redisClient.SAdd(r.ctx, key, friendId).Result()
	return err
}

func (r *RedisCache) GetFriends(userId string) ([]string, error) {
	key := fmt.Sprintf("friends:%s", userId)
	return r.redisClient.SMembers(r.ctx, key).Result()
}

func (r *RedisCache) RemoveFriend(userId, friendId string) error {
	key := fmt.Sprintf("friends:%s", userId)
	return r.redisClient.SRem(r.ctx, key, friendId).Err()
}

func (r *RedisCache) PublishLocationUpdate(location models.Location) error {
	friends, err := r.GetFriends(location.UserId)
	if err != nil {
		return fmt.Errorf("error getting friends: %w", err)
	}

	locationJSON, err := json.Marshal(location)
	if err != nil {
		return fmt.Errorf("error marshaling location: %w", err)
	}

	for _, friendId := range friends {
		channel := fmt.Sprintf("location_updates:%s", friendId)
		err = r.redisClient.Publish(r.ctx, channel, locationJSON).Err()
		if err != nil {
			log.Printf("Error publishing to channel %s: %v", channel, err)
		}
	}

	return nil
}

func (r *RedisCache) SubscribeToFriendUpdates(userId string, updateChan chan<- models.Location) {
	channel := fmt.Sprintf("location_updates:%s", userId)
	pubsub := r.redisClient.Subscribe(r.ctx, channel)
	defer pubsub.Close()

	for {
		msg, err := pubsub.ReceiveMessage(r.ctx)
		if err != nil {
			log.Printf("Error receiving message: %v", err)
			continue
		}

		var location models.Location
		err = json.Unmarshal([]byte(msg.Payload), &location)
		if err != nil {
			log.Printf("Error unmarshaling location: %v", err)
			continue
		}

		updateChan <- location
	}
}

func (r *RedisCache) StartLocationUpdateWorker(userId string) {
	updateChan := make(chan models.Location, 100)
	go r.SubscribeToFriendUpdates(userId, updateChan)

	for location := range updateChan {
		log.Printf("Received location update for friend %s: lat %f, lon %f",
			location.UserId, location.CurrentLatitude, location.CurrentLongitude)
		// Process the location update (e.g., update UI, trigger notifications, etc.)
	}
}
