package main

import (
	"fmt"
	"log"
	"os"

	"github.com/reynaldipane/flight-search/internal/handlers"
)

func main() {
	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Setup router
	router := handlers.SetupRouter()

	// Start server
	addr := fmt.Sprintf(":%s", port)
	log.Printf("Starting Flight Search API on %s", addr)
	log.Printf("API Documentation available at http://localhost%s/api/v1/health", addr)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
