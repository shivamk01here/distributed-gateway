package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/shivamk01here/distributed-gateway/internal/api"
	"github.com/shivamk01here/distributed-gateway/internal/limiter"
)

func main() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "super-secret-key-for-local-testing"
	}

	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})

	// You can easily swap to NewSlidingWindowLimiter or NewLeakyBucketLimiter here!
	leakyLimiter := limiter.NewLeakyBucketLimiter(rdb, 5, 1.0)

	backendURLs := []string{"http://localhost:9000"} // Update for local testing
	lb := api.NewLoadBalancer(backendURLs)

	// Middleware Chain
	mux := http.NewServeMux()
	secureHandler := api.AuthMiddleware(jwtSecret, api.RateLimitMiddleware(leakyLimiter, lb.ReverseProxyHandler()))

	mux.Handle("/api/", secureHandler)
	mux.Handle("/metrics", promhttp.Handler())

	port := ":8080"

	srv := &http.Server{
		Addr:    port,
		Handler: mux,
	}

	go func() {
		log.Printf("Distributed API Gateway live on port %s\n", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server forced to shutdown: %v", err)
		}
	}()

	// 2. channel to listen for OS signals (like Ctrl+C or Docker stop)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 3. Block the main thread until a signal is received
	<-quit
	log.Println("Shutting down Gateway gracefully...")

	// 4. Create a 5-second timeout context.
	// The server will stop accepting new requests, and give active ones 5 seconds to finish.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Gateway forced shutdown: %v", err)
	}

	log.Println("Gateway exited cleanly. No connections dropped.")
}
