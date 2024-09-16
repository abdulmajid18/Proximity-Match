package matcher

import (
	"context"
	"fmt"
	"log"
	"math"

	"github.com/redis/go-redis/v9"
)

func matchUsers(ctx context.Context, rdb *redis.Client, userID string, radius float64) ([]string, error) {
	currentLocation, err := rdb.GeoPos(ctx, "user_locations", userID).Result()
	if err != nil {
		log.Fatalf("Error retrieving location: %v", err)
	}

	if len(currentLocation) == 0 || currentLocation[0] == nil {
		log.Fatalf("User location not found")
	}

	nearbyUsers, err := rdb.GeoRadius(ctx, "user_locations", currentLocation[0].Longitude, currentLocation[0].Latitude, &redis.GeoRadiusQuery{
		Radius:    radius, // radius in km
		Unit:      "km",
		WithCoord: true,
		WithDist:  true,
		Sort:      "ASC",
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("error finding nearby users: %v", err)
	}

	// Get the current user's destination
	currentUserDestination, err := rdb.HGet(ctx, "user_destinations", userID).Result()
	if err != nil {
		return nil, fmt.Errorf("error retrieving current user destination: %v", err)
	}
	currentDestination := parseCoordinates(currentUserDestination)

	matchedUsers := []string{}
	for _, user := range nearbyUsers {
		if user.Name == userID {
			continue
		}

		// Retrieve the destination for each nearby user
		destination, err := rdb.HGet(ctx, "user_destinations", user.Name).Result()
		if err != nil {
			log.Printf("Error retrieving destination for user %s: %v", user.Name, err)
			continue
		}
		userDestination := parseCoordinates(destination)

		if isSameDestination(currentDestination, userDestination) {
			matchedUsers = append(matchedUsers, user.Name)
		}
	}
	return matchedUsers, nil
}

func parseCoordinates(coordinates string) [2]float64 {
	var lat, lon float64
	fmt.Sscanf(coordinates, "%f,%f", &lat, &lon)
	return [2]float64{lat, lon}
}

func isSameDestination(dest1, dest2 [2]float64) bool {
	const tolerance = 0.01 // Tolerance for considering two destinations as the same
	return math.Abs(dest1[0]-dest2[0]) <= tolerance && math.Abs(dest1[1]-dest2[1]) <= tolerance
}
