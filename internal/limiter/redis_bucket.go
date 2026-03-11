package limiter

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisTokenBucket struct {
	client     *redis.Client
	capacity   float64
	refillRate float64
}

func NewRedisTokenBucket(redisClient *redis.Client, capacity int, refillRate float64) *RedisTokenBucket {
	return &RedisTokenBucket{
		client:     redisClient,
		capacity:   float64(capacity),
		refillRate: refillRate,
	}
}

var tokenBucketScript = redis.NewScript(`
	local key_tokens = KEYS[1] .. ':tokens'
	local key_ts = KEYS[1] .. ':ts'

	local capacity = tonumber(ARGV[1])
	local refill_rate = tonumber(ARGV[2])
	local now = tonumber(ARGV[3])

	local tokens = tonumber(redis.call('get', key_tokens))
	local last_refill = tonumber(redis.call('get', key_ts))

	if tokens == nil then
		tokens = capacity
		last_refill = now
	else
		-- Time-delta math in Redis
		local time_passed = math.max(0, now - last_refill)
		tokens = math.min(capacity, tokens + (time_passed * refill_rate))
	end

	if tokens >= 1 then
		tokens = tokens - 1
		redis.call('set', key_tokens, tokens)
		redis.call('set', key_ts, now)
		-- Set a TTL (Time To Live) so Redis auto-deletes stale IPs! (Replaces our GC Daemon)
		redis.call('expire', key_tokens, 60) 
		redis.call('expire', key_ts, 60)
		return 1 -- Allowed
	end

	return 0 -- Throttled
`)

func (r *RedisTokenBucket) Allow(key string) bool {
	ctx := context.Background()
	now := time.Now().Unix()

	result, err := tokenBucketScript.Run(ctx, r.client, []string{key}, r.capacity, r.refillRate, now).Result()
	if err != nil {
		return false
	}

	return result.(int64) == 1
}
