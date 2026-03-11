package limiter

import (
	"sync"
	"time"
)

// clientState holds the token bucket for a single, specific IP address.
type clientState struct {
	tokens     float64
	lastRefill time.Time
}

type TokenBucket struct {
	capacity   int
	tokens     float64
	refillRate float64
	clients    map[string]*clientState
	mu         sync.Mutex
}

func NewTokenBucket(capacity int, refillRate float64) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		refillRate: refillRate,
		clients:    make(map[string]*clientState),
	}
}

func (tb *TokenBucket) Allow(key string) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()

	// 1. Get the bucket for this specific IP
	client, exists := tb.clients[key]

	if !exists {
		tb.clients[key] = &clientState{
			// They use 1 token immediately for this request
			tokens:     tb.capacity - 1.0,
			lastRefill: now,
		}
		return true
	}

	timeElapsed := now.Sub(client.lastRefill).Seconds()
	
	// Multiply seconds passed by the refill rate to get new tokens.
	tokensToAdd := timeElapsed * tb.refillRate

	

// File: internal/limiter/token_bucket.go
package limiter

import (
	"sync"
	"time"
)

// clientState holds the token bucket for a single, specific IP address.
type clientState struct {
	tokens     float64
	lastRefill time.Time
}

// TokenBucket implements the RateLimiter interface.
type TokenBucket struct {
	capacity   float64
	refillRate float64
	
	// SDE2 Flex: The Map
	// We map an IP address (string) to its specific bucket state.
	clients map[string]*clientState
	
	// A single Mutex to protect the entire map from concurrent read/writes.
	mu sync.Mutex 
}

// NewTokenBucket initializes the global limiter.
func NewTokenBucket(capacity int, refillRate float64) *TokenBucket {
	return &TokenBucket{
		capacity:   float64(capacity),
		refillRate: refillRate,
		clients:    make(map[string]*clientState), // Initialize the map!
	}
}

// Allow checks if a specific IP has enough tokens to proceed.
func (tb *TokenBucket) Allow(key string) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()

	// 1. Get the bucket for this specific IP
	client, exists := tb.clients[key]
	
	// 2. If the IP is new, create a full bucket for them and allow the request.
	if !exists {
		tb.clients[key] = &clientState{
			// They use 1 token immediately for this request
			tokens:     tb.capacity - 1.0, 
			lastRefill: now,
		}
		return true
	}

	// 3. The Time-Delta Math (The Core Algorithm)
	// Calculate exactly how many seconds have passed since they last made a request.
	timeElapsed := now.Sub(client.lastRefill).Seconds()
	
	// Multiply seconds passed by the refill rate to get new tokens.
	tokensToAdd := timeElapsed * tb.refillRate

	// 4. Refill the bucket until max.
	return false
}
