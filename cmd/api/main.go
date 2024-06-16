package main

import (
	"context"
	"log"
	"os/signal"
	"sync"
	"syscall"

	"github.com/danielperaltamadriz/tinyurl/api"
	"github.com/danielperaltamadriz/tinyurl/config"
)

func main() {
	server, err := api.NewAPI(config.Config{})
	if err != nil {
		log.Fatalf("failed to start server, err: %v", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		log.Println("Starting server...")
		err := server.Start()
		if err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
		wg.Done()
	}()

	<-ctx.Done()

	err = server.Shutdown()
	if err != nil {
		log.Fatalf("Failed to stop server: %v", err)
	}
	wg.Wait()
	log.Println("Server stopped")
}
