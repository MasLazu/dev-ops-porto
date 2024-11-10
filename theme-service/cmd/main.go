package main

import (
	"context"
	"log"
	"os"

	"github.com/MasLazu/dev-ops-porto/theme-service/internal/server"
)

func main() {
	ctx := context.Background()
	if err := server.Run(ctx); err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}
}
