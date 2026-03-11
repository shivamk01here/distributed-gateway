// File: cmd/gateway/main.go
package main

import (
	"log"
	"net/http"

	"github.com/redis/go-redis/v9"
	"github.com/shivamk01here/distributed-gateway/internal/api"
	"github.com/shivamk01here/distributed-gateway/internal/limiter"
)

func main() {
	log.Println("Connecting to Redis Cluster...")
	// Connect to local Redis on port 6379
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	log.Println("Initializing Distributed Redis Rate Limiter...")
	// 5 tokens max, 1 token refill per second
	redisLimiter := limiter.NewRedisTokenBucket(rdb, 5, 1.0)

	log.Println("Configuring Reverse Proxy routing...")
	// send allowed traffic to our backend service running on port 9000
	backendServiceURL := "http://localhost:9000"
	proxyHandler := api.NewReverseProxy(backendServiceURL)

	// Wrap the proxy handler with our Redis rate limiting middleware
	gatewayRouter := api.RateLimitMiddleware(redisLimiter, proxyHandler)

	// root route so the gateway catches everything
	http.HandleFunc("/", gatewayRouter)

	port := ":8080"
	log.Printf("🔥 Distributed API Gateway live on port %s forwarding to %s\n", port, backendServiceURL)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Gateway crashed: %v", err)
	}
}
