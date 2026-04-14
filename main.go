package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Suthar345Piyush/limiter"
	rl "github.com/Suthar345Piyush/middleware"
	"github.com/redis/go-redis/v9"
)

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("redis: %v", err)
	}

	// limit is the 100 requests / minute / user

	lim := limiter.New(rdb, limiter.Config{
		Limit:  100,
		Window: time.Minute,
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/api/data", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"data": "ok"}`))
	})

	// applying rate limiting using user id header

	handler := rl.RateLimit(lim, rl.ByUser)(mux)

	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))

}
