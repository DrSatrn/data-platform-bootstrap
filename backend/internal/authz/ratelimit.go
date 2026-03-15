// This file provides a tiny in-memory failed-login limiter so the platform can
// slow down password guessing without introducing another persistence system.
package authz

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type loginRateLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	records map[string][]time.Time
}

func newLoginRateLimiter(limit int, window time.Duration) *loginRateLimiter {
	return &loginRateLimiter{
		limit:   limit,
		window:  window,
		records: map[string][]time.Time{},
	}
}

func (l *loginRateLimiter) Allow(key string, now time.Time) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	key = normalizeRateLimitKey(key)
	windowStart := now.Add(-l.window)
	attempts := pruneTimes(l.records[key], windowStart)
	l.records[key] = attempts
	if len(attempts) < l.limit {
		return true, 0
	}
	retryAfter := attempts[0].Add(l.window).Sub(now)
	if retryAfter < 0 {
		retryAfter = 0
	}
	return false, retryAfter
}

func (l *loginRateLimiter) RecordFailure(key string, now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()

	key = normalizeRateLimitKey(key)
	windowStart := now.Add(-l.window)
	attempts := pruneTimes(l.records[key], windowStart)
	l.records[key] = append(attempts, now)
}

func (l *loginRateLimiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	delete(l.records, normalizeRateLimitKey(key))
}

func loginRateLimitKey(r *http.Request) string {
	forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	if forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return normalizeRateLimitKey(parts[0])
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil {
		return normalizeRateLimitKey(host)
	}
	return normalizeRateLimitKey(r.RemoteAddr)
}

func pruneTimes(times []time.Time, cutoff time.Time) []time.Time {
	out := times[:0]
	for _, item := range times {
		if item.After(cutoff) {
			out = append(out, item)
		}
	}
	if out == nil {
		return []time.Time{}
	}
	return out
}

func normalizeRateLimitKey(key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return "unknown"
	}
	return key
}
