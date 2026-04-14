package limiter

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// some structs

type Config struct {
	Limit     int           // maximum requests
	Window    time.Duration // rolling window
	keyPrefix string        // like - rl:
}

// final result (0 or 1)

type Result struct {
	Allowed    bool // (0 -> not allowed, 1 -> allowed)
	Remaining  int
	RetryAfter time.Duration // only when not allowed
}

type Limiter struct {
	rdb    *redis.Client // redis db client
	cfg    Config
	script *redis.Script
}

func New(rdb *redis.Client, cfg Config) *Limiter {
	if cfg.keyPrefix == "" {
		cfg.keyPrefix = "rl:"
	}

	return &Limiter{
		rdb:    rdb,
		cfg:    cfg,
		script: redis.NewScript(luaScript),
	}
}

// checking the given userid is within rate limit or not

func (l *Limiter) Allow(ctx context.Context, key string) (Result, error) {

	now := time.Now().UnixMilli()
	windowMs := l.cfg.Window.Milliseconds()
	member := fmt.Sprintf("%d:%s", now, randHex())

	// using redis hash tag to land all keys of one user in same slot of redis cluster

	redisKey := fmt.Sprintf("%s{%s}", l.cfg.keyPrefix, key)

	vals, err := l.script.Run(ctx, l.rdb, []string{redisKey}, windowMs, l.cfg.Limit, now, member).Int64Slice()

	if err != nil {
		return Result{}, fmt.Errorf("Rate limiter redis eval: %w", err)
	}

	// if first vals is 1, then allow it

	if vals[0] == 1 {
		return Result{
			Allowed:   true,
			Remaining: int(vals[1]),
		}, nil
	}

	// vals[2] = fix reset timestamps in ms

	retryAfter := time.Duration(vals[2]-now) * time.Millisecond

	if retryAfter < 0 {
		retryAfter = 0
	}

	return Result{
		Allowed:    false,
		Remaining:  0,
		RetryAfter: retryAfter,
	}, nil

}

// it returns encoded hex string

func randHex() string {
	b := make([]byte, 4)
	rand.Read(b)

	return hex.EncodeToString(b)
}
