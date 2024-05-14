package pkg

import "net/http"

type RateLimiter struct {
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{}
}
func (r *RateLimiter) Limit(handler http.Handler) http.Handler {
	return handler
}
