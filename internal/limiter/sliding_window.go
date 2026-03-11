package limiter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type SlidingWindowLimiter struct {
	client   *redis.Client
	capacity int
	window   int
}

func NewSlidingWindowLimiter(rdb *redis.Client, capacity, windowSeconds int) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		client:   rdb,
		capacity: capacity,
		window:   windowSeconds,
	}
}

// KEYS[1] = IP address
// ARGV[1] = current time in milliseconds
// ARGV[2] = window size in milliseconds
// ARGV[3] = capacity
// ARGV[4] = unique request ID (to prevent ZSET collisions if two requests hit at the exact same millisecond)
var slidingWindowScript = redis.NewScript(`
	local key = KEYS[1]
	local now = tonumber(ARGV[1])
	local window = tonumber(ARGV[2])
	local capacity = tonumber(ARGV[3])
	local member = ARGV[4]

	local clearBefore = now - window

	-- 1. Remove all requests that are outside the sliding window
	redis.call('ZREMRANGEBYSCORE', key, 0, clearBefore)

	-- 2. Count active requests inside the window
	local count = redis.call('ZCARD', key)

	-- 3. If under capacity, allow it
	if count < capacity then
		redis.call('ZADD', key, now, member)
		-- Set TTL to the window size to prevent memory leaks in Redis
		redis.call('EXPIRE', key, math.ceil(window / 1000))
		return 1
	end

	-- 4. Throttled
	return 0
`)

func (s *SlidingWindowLimiter) Allow(key string) bool {
	ctx := context.Background()
	now := time.Now().UnixMilli()
	windowMillis := int64(s.window * 1000)
	uniqueMember := fmt.Sprintf("%d-%s", now, key)

	result, err := slidingWindowScript.Run(ctx, s.client, []string{key}, now, windowMillis, s.capacity, uniqueMember).Result()
	if err != nil {
		return false // Fail closed
	}

	return result.(int64) == 1
}
