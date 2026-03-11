package main

import (
	"log"
	"net/http"
	"os"

	"github.com/redis/go-redis/v9"
	"github.com/shivamk01here/distributed-gateway/internal/api"
	"github.com/shivamk01here/distributed-gateway/internal/limiter"
)

func main() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	log.Printf("Connecting to Redis Cluster at %s...", redisAddr)
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	log.Println("Initializing Distributed Redis Rate Limiter...")
	redisLimiter := limiter.NewRedisTokenBucket(rdb, 5, 1.0)

	log.Println("Configuring Round Robin Load Balancer...")
	backendURLs := []string{
		"http://backend_alpha:8080",
		"http://backend_beta:8080",
	}

	lb := api.NewLoadBalancer(backendURLs)

	gatewayRouter := api.RateLimitMiddleware(redisLimiter, lb.ReverseProxyHandler())

	http.HandleFunc("/", gatewayRouter)

	port := ":8080"
	log.Printf(" Distributed API Gateway live on port %s\n", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Gateway crashed: %v", err)
	}
}
