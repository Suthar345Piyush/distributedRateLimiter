package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Suthar345Piyush/limiter"
)

type keyFunc func(r *http.Request) string

// client IP as the rate limit key

func ByIP(r *http.Request) string {

	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}

	return r.RemoteAddr
}

// getting the user id from header

func ByUser(r *http.Request) string {
	if uid := r.Header.Get("X-User-ID"); uid != "" {
		return uid
	}

	return ByIP(r)
}

// rate limit function to return handler

func RateLimit(l *limiter.Limiter, keyFn keyFunc) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			key := keyFn(r)
			res, err := l.Allow(r.Context(), key)

			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			// attaching rate limiter header
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(res.Remaining))

			if !res.Allowed {
				w.Header().Set("Retry-After", strconv.Itoa(int(res.RetryAfter/time.Second)+1))

				http.Error(w, `{"error": "rate limit exceeded"}`, http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}

}
