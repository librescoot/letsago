package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	ipc "github.com/librescoot/redis-ipc"
)

var version = "dev"

const (
	vehicleHashName   = "vehicle"
	dashboardHashName = "dashboard"
	dashboardTopic    = "dashboard"
	stateField        = "state"
	readyField        = "ready"
	readyValue        = "true"
	previousState     = "stand-by"
	targetState       = "parked"
)

func main() {
	redisAddr := flag.String("redis", "localhost:6379", "Redis address (host:port)")
	versionFlag := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("letsago %s\n", version)
		os.Exit(0)
	}

	if os.Getenv("JOURNAL_STREAM") != "" {
		log.SetFlags(0)
	} else {
		log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	}

	log.Printf("letsago %s starting", version)
	log.Printf("monitoring Redis at %s for vehicle state changes", *redisAddr)

	client, err := ipc.New(
		ipc.WithURL(*redisAddr),
		ipc.WithOnConnect(func() {
			log.Println("connected to Redis")
		}),
		ipc.WithOnDisconnect(func(err error) {
			log.Printf("disconnected from Redis: %v", err)
		}),
	)
	if err != nil {
		log.Fatalf("failed to create Redis client: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	var mu sync.Mutex
	var lastState string

	dashPub := client.NewHashPublisherWithChannel(dashboardHashName, dashboardTopic)

	watcher := client.NewHashWatcher(vehicleHashName)
	watcher.OnField(stateField, func(value string) error {
		mu.Lock()
		prev := lastState
		lastState = value
		mu.Unlock()

		if prev == value {
			return nil
		}

		log.Printf("vehicle state change: %s -> %s", prev, value)

		if prev == previousState && value == targetState {
			log.Printf("detected transition %s -> %s, setting dashboard ready", previousState, targetState)

			if err := dashPub.Set(readyField, readyValue); err != nil {
				log.Printf("error setting dashboard ready state: %v", err)
			}
		}

		return nil
	})

	if err := watcher.StartWithSync(); err != nil {
		log.Fatalf("failed to start vehicle watcher: %v", err)
	}
	defer watcher.Stop()

	log.Println("watching vehicle state")

	select {
	case sig := <-sigChan:
		log.Printf("received signal %v, shutting down", sig)
		cancel()
	case <-ctx.Done():
	}

	log.Println("letsago stopped")
}
