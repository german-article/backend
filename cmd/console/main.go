package main

import (
	"context"
	"fmt"
	"github.com/DeryabinSergey/germanarticlebot/internal/infrastructure/container"
	"log"
	"os"
)

func main() {
	ctx := context.Background()

	// Initialize container
	appContainer, err := container.NewContainer(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// Get word and language from command line arguments
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run cmd/console/main.go <word> [language]")
		fmt.Println("Example: go run cmd/console/main.go Haus en")
		os.Exit(1)
	}

	word := os.Args[1]
	language := "en"
	if len(os.Args) > 2 {
		language = os.Args[2]
	}

	// Process request
	response, err := appContainer.ConsoleHandler.ProcessRequest(ctx, word, language)
	if err != nil {
		log.Fatalf("Failed to process request: %v", err)
	}

	fmt.Println(response)
}
