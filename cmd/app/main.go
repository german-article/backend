package main

import (
	"context"
	"github.com/DeryabinSergey/germanarticlebot/internal/domain"
	"github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
	"log"
	"os"
	"os/signal"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := funcframework.RegisterHTTPFunctionContext(ctx, "/", domain.Invoke); err != nil {
		log.Fatalf("RegisterHTTPFunctionContext: %v\n", err)
	}

	// Use PORT environment variable, or default to 8080.
	port := "8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	funcFrameworkError := make(chan error, 1)
	go func() {
		funcFrameworkError <- funcframework.Start(port)
	}()

	select {
	case <-ctx.Done():
		log.Println("Shutting down gracefully...")
	case err := <-funcFrameworkError:
		if err != nil {
			log.Fatalf("funcframework.Start: %v\n", err)
		}
	}
}
