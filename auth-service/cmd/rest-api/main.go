package main

import (
	"auth-service/internal/server"
	"context"
	"log"
	"os"
)

func main() {
	ctx := context.Background()
	if err := server.Run(ctx); err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}
}
