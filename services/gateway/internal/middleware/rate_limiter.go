package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

const slidingWindowLuaScript = `
local key = KEYS[1]
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])
local member = ARGV[4]

local window_start = now - window

redis.call("ZREMRANGEBYSCORE", key, 0, window_start)

local current_count = redis.call("ZCARD", key)

if current_count >= limit then
	return {0, current_count}
end

redis.call("ZADD", key, now, member)
redis.call("EXPIRE", key, math.ceil(window / 1000))

return {1, current_count + 1}
`

type RateLimiter struct {
	client       *redis.Client
	limit        int64
	window       time.Duration
	windowMillis int64
}

func NewRateLimiter(redisAddr string, limit int64, window time.Duration) *RateLimiter {
	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	return &RateLimiter{
		client:       client,
		limit:        limit,
		window:       window,
		windowMillis: window.Milliseconds(),
	}
}

func (r *RateLimiter) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *RateLimiter) Close() error {
	return r.client.Close()
}

func (r *RateLimiter) Limit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		userID, _ := req.Context().Value(UserIDKey).(string)
		if userID == "" {
			userID = "anonymous"
		}

		key := fmt.Sprintf("rate_limit:user:%s", userID)

		now := time.Now().UnixMilli()
		member := fmt.Sprintf("%d:%d", now, time.Now().UnixNano())

		result, err := r.client.Eval(
			req.Context(),
			slidingWindowLuaScript,
			[]string{key},
			now,
			r.windowMillis,
			r.limit,
			member,
		).Result()

		if err != nil {
			writeRateLimitError(w, http.StatusInternalServerError, "rate limiter failed")
			return
		}

		values, ok := result.([]any)
		if !ok || len(values) < 2 {
			writeRateLimitError(w, http.StatusInternalServerError, "invalid rate limiter response")
			return
		}

		allowed, ok := values[0].(int64)
		if !ok {
			writeRateLimitError(w, http.StatusInternalServerError, "invalid rate limiter decision")
			return
		}

		if allowed == 0 {
			w.Header().Set("Retry-After", fmt.Sprintf("%.0f", r.window.Seconds()))
			writeRateLimitError(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}

		next(w, req)
	}
}

func writeRateLimitError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
