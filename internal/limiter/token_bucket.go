// File: internal/limiter/token_bucket.go
package limiter

import (
	"log"
	"sync"
	"time"
)

type clientState struct {
	tokens     float64
	lastRefill time.Time
}

type TokenBucket struct {
	capacity   float64
	refillRate float64
	clients    map[string]*clientState
	mu         sync.Mutex
}

// NewTokenBucket initializes the global limiter and starts the GC daemon.
func NewTokenBucket(capacity int, refillRate float64) *TokenBucket {
	tb := &TokenBucket{
		capacity:   float64(capacity),
		refillRate: refillRate,
		clients:    make(map[string]*clientState),
	}

	go tb.cleanupStaleClients()

	return tb
}

// cleanupStaleClients runs a continuous loop to prevent memory leaks.
func (tb *TokenBucket) cleanupStaleClients() {
	for {
		// Wait for 1 minute before running the next sweep
		time.Sleep(1 * time.Minute)

		tb.mu.Lock()
		for ip, client := range tb.clients {
			// If the client hasn't made a request in the last 3 minutes,
			// they have naturally refilled to max capacity anyway.
			// We can safely delete them from memory.
			if time.Since(client.lastRefill) > 3*time.Minute {
				delete(tb.clients, ip)
			}
		}

		// Log the current map size so we can monitor memory usage in production
		log.Printf("[MEMORY] GC Sweep complete. Active IPs in memory: %d", len(tb.clients))

		tb.mu.Unlock()
	}
}

func (tb *TokenBucket) Allow(key string) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()

	client, exists := tb.clients[key]

	if !exists {
		tb.clients[key] = &clientState{
			tokens:     tb.capacity - 1.0,
			lastRefill: now,
		}
		return true
	}

	timeElapsed := now.Sub(client.lastRefill).Seconds()
	tokensToAdd := timeElapsed * tb.refillRate

	client.tokens += tokensToAdd
	if client.tokens > tb.capacity {
		client.tokens = tb.capacity
	}

	client.lastRefill = now

	if client.tokens >= 1.0 {
		client.tokens -= 1.0
		return true
	}

	return false
}
