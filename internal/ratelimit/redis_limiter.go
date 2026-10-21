package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisLimiter struct {
	client *redis.Client
	limit  int64
	window time.Duration
}

func NewRedisLimiter(
	client *redis.Client,
	limit int64,
	window time.Duration,
) *RedisLimiter {
	return &RedisLimiter{
		client: client,
		limit:  limit,
		window: window,
	}
}

var fixedWindowScript = redis.NewScript(`
	local count = redis.call("INCR", KEYS[1])

	if count == 1 then
		redis.call("PEXPIRE", KEYS[1], ARGV[1])
	end

	return count
`)

func (l *RedisLimiter) Allow(
	ctx context.Context,
	connectionID string,
) (bool, error) {
	key := "live-chat:rate:" + connectionID

	count, err := fixedWindowScript.Run(
		ctx,
		l.client,
		[]string{key},
		l.window.Milliseconds(),
	).Int64()
	if err != nil {
		return false, fmt.Errorf("redis rate limit: %w", err)
	}

	return count <= l.limit, nil
}
