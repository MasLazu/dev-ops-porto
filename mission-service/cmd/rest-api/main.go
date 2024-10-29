package main

import (
	"context"
	"log"
	"mission-service/internal/server"
	"os"
)

func main() {
	ctx := context.Background()
	if err := server.Run(ctx); err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}
}
