// File: cmd/gateway/main.go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/shivamk01here/distributed-gateway/internal/api"
	"github.com/shivamk01here/distributed-gateway/internal/limiter"
)

// dummyBackendService simulates the core service we are trying to protect.
func dummyBackendService(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Success! You reached the protected backend service.\n")
}

func main() {
	log.Println("Initializing Rate Limiter...")
	tb := limiter.NewTokenBucket(5, 1.0)

	log.Println("Setting up HTTP routes...")
	protectedHandler := api.RateLimitMiddleware(tb, dummyBackendService)

	// Register the route
	http.HandleFunc("/api/data", protectedHandler)

	port := ":8080"
	log.Printf("Starting Distributed API Gateway on port %s...\n", port)

	// Start the server
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
