package limiter

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type LeakyBucketLimiter struct {
	client   *redis.Client
	capacity float64
	leakRate float64 // drops leaked per second
}

func NewLeakyBucketLimiter(rdb *redis.Client, capacity int, leakRate float64) *LeakyBucketLimiter {
	return &LeakyBucketLimiter{
		client:   rdb,
		capacity: float64(capacity),
		leakRate: leakRate,
	}
}

// ARGV[1] = capacity, ARGV[2] = leak_rate_per_second, ARGV[3] = current_time_ms
var leakyBucketScript = redis.NewScript(`
	local key_water = KEYS[1] .. ':water'
	local key_ts = KEYS[1] .. ':ts'

	local capacity = tonumber(ARGV[1])
	local leak_rate_per_ms = tonumber(ARGV[2]) / 1000
	local now = tonumber(ARGV[3])

	local water = tonumber(redis.call('get', key_water)) or 0
	local last_leak_ts = tonumber(redis.call('get', key_ts)) or now

	-- 1. Calculate how much water leaked out since the last request
	local time_passed_ms = math.max(0, now - last_leak_ts)
	local leaked = time_passed_ms * leak_rate_per_ms
	
	-- 2. Update current water level (cannot go below 0)
	water = math.max(0, water - leaked)

	-- 3. If we add 1 drop (this request), does it overflow?
	if water + 1 <= capacity then
		water = water + 1
		redis.call('set', key_water, water)
		redis.call('set', key_ts, now)
		redis.call('expire', key_water, 60)
		redis.call('expire', key_ts, 60)
		return 1 -- Allowed
	end

	return 0 -- Throttled (Bucket is overflowing)
`)

func (l *LeakyBucketLimiter) Allow(key string) bool {
	ctx := context.Background()
	now := time.Now().UnixMilli()

	result, err := leakyBucketScript.Run(ctx, l.client, []string{key}, l.capacity, l.leakRate, now).Result()
	if err != nil {
		return false // Fail closed
	}

	return result.(int64) == 1
}
