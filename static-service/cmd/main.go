package main

import (
	"context"
	"log"
	"os"

	"github.com/MasLazu/dev-ops-porto/static-service/internal"
)

func main() {
	ctx := context.Background()
	if err := internal.Run(ctx); err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}
}
