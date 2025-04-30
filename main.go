package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	redisHost      = "192.168.7.1"
	redisPort      = "6379"
	vehicleHash    = "vehicle"
	stateField     = "state"
	dashboardHash  = "dashboard"
	dashboardTopic = "dashboard"
	readyField     = "ready"
	readyValue     = "true"
	readyMessage   = "ready"
	previousState  = "stand-by"
	targetState    = "parked"
	pollInterval   = 500 * time.Millisecond
	connectTimeout = 5 * time.Second
)

func main() {
	// Set up logging
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("LetsAGo vehicle state watcher starting up")
	log.Printf("Monitoring Redis at %s:%s for vehicle state changes", redisHost, redisPort)

	// Create a Redis client
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// Test the connection
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Successfully connected to Redis")

	// Create a context that will be canceled on interrupt signal
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start monitoring loop
	monitorVehicleState(ctx, rdb)
}

func monitorVehicleState(ctx context.Context, rdb *redis.Client) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	var lastState string

	for {
		select {
		case <-ctx.Done():
			log.Println("Received termination signal, shutting down")
			return
		case <-ticker.C:
			// Get the current state
			state, err := rdb.HGet(ctx, vehicleHash, stateField).Result()
			if err != nil {
				if err == redis.Nil {
					log.Println("Vehicle state key not found, waiting...")
					continue
				}
				log.Printf("Error getting vehicle state: %v", err)
				continue
			}

			// Only log when state changes to reduce verbosity
			if state != lastState {
				log.Printf("Vehicle state change: %s -> %s", lastState, state)

				// Check for specific state transition (stand-by -> parked)
				if lastState == previousState && state == targetState {
					log.Printf("Detected transition from '%s' to '%s', setting dashboard ready to %s",
						previousState, targetState, readyValue)

					// Set the dashboard hash field
					err = rdb.HSet(ctx, dashboardHash, readyField, readyValue).Err()
					if err != nil {
						log.Printf("Error setting dashboard ready state: %v", err)
					} else {
						log.Printf("Successfully set %s %s to %s", dashboardHash, readyField, readyValue)
					}

					// Publish to the dashboard channel
					err = rdb.Publish(ctx, dashboardTopic, readyMessage).Err()
					if err != nil {
						log.Printf("Error publishing to %s channel: %v", dashboardTopic, err)
					} else {
						log.Printf("Successfully published '%s' to %s channel", readyMessage, dashboardTopic)
					}
				}

				// Update lastState after processing
				lastState = state
			}
		}
	}
}
