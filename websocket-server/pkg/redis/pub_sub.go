package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"matching-service/websocket-server/internal/models"
)

func (r *RedisCache) PublishLocationUpdate(location models.Location) error {
	locationJSON, err := json.Marshal(location)
	if err != nil {
		return fmt.Errorf("error marshaling location: %w", err)
	}

	channel := fmt.Sprintf("location_updates:%s", location.UserId)
	err = r.redisClient.Publish(r.ctx, channel, locationJSON).Err()
	if err != nil {
		return fmt.Errorf("error publishing to channel %s: %w", channel, err)
	}

	return nil
}

func (r *RedisCache) SubscribeToFriendUpdates(userId string, updateChan chan<- models.Location) {
	friends, err := r.GetFriends(userId)
	if err != nil {
		log.Printf("Error getting friends for user %s: %v", userId, err)
		return
	}

	channels := make([]string, len(friends))
	for i, friendId := range friends {
		channels[i] = fmt.Sprintf("location_updates:%s", friendId)
	}

	pubsub := r.redisClient.Subscribe(r.ctx, channels...)
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

func (r *RedisCache) StartLocationUpdateWorker(ctx context.Context, userId string) {
	updateChan := make(chan models.Location, 100)
	go r.SubscribeToFriendUpdates(userId, updateChan)

	for {
		select {
		case location := <-updateChan:
			log.Printf("Received location update for friend %s: lat %f, lon %f",
				location.UserId, location.CurrentLatitude, location.CurrentLongitude)
			// Process the location update (e.g., update UI, trigger notifications, etc.)
		case <-ctx.Done():
			log.Println("Shutting down worker for user:", userId)
			return
		}
	}
}

// Existing friend-related functions
func (r *RedisCache) AddFriend(userId, friendId string) error {
	key := fmt.Sprintf("friends:%s", userId)
	_, err := r.redisClient.SAdd(r.ctx, key, friendId).Result()
	if err != nil {
		return err
	}

	// Resubscribe to include the new friend's channel
	// This is a simplified approach; in practice, you might want to manage subscriptions more carefully
	go r.refreshSubscriptions(userId)

	return nil
}

func (r *RedisCache) RemoveFriend(userId, friendId string) error {
	key := fmt.Sprintf("friends:%s", userId)
	err := r.redisClient.SRem(r.ctx, key, friendId).Err()
	if err != nil {
		return err
	}

	// Resubscribe to exclude the removed friend's channel
	go r.refreshSubscriptions(userId)

	return nil
}

func (r *RedisCache) GetFriends(userId string) ([]string, error) {
	key := fmt.Sprintf("friends:%s", userId)
	return r.redisClient.SMembers(r.ctx, key).Result()
}

func (r *RedisCache) refreshSubscriptions(userId string) {
	// Implementation to refresh subscriptions when friend list changes
	// This is a placeholder and would need to be implemented based on your specific subscription management approach
}
