package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

var startTime = time.Now()

// ---- rate limiter ----
type visitor struct {
	lastSeen time.Time
	count    int
}

var (
	visitors = make(map[string]*visitor)
	mu       sync.Mutex
)

func rateLimit(next http.HandlerFunc) http.HandlerFunc {
	maxRequestsStr := getEnv("MAX_REQUESTS", "10")
	maxRequests, err := strconv.Atoi(maxRequestsStr)
	if err != nil {
		log.Printf("Invalid MAX_REQUESTS value, using default 10")
		maxRequests = 10
	}
	const window = time.Minute
	return func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		mu.Lock()
		v, exists := visitors[ip]
		if !exists || time.Since(v.lastSeen) > window {
			visitors[ip] = &visitor{lastSeen: time.Now(), count: 1}
			mu.Unlock()
			next(w, r)
			return
		}

		if v.count >= maxRequests {
			mu.Unlock()
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		v.count++
		v.lastSeen = time.Now()
		mu.Unlock()

		next(w, r)
	}
}

// ----------------------

func main() {
	port := getEnv("PORT", "8080")
	tz := getEnv("TIMEZONE", "UTC")
	appName := getEnv("APP_NAME", "Chronos")

	location, err := time.LoadLocation(tz)
	if err != nil {
		log.Fatalf("invalid timezone: %s", tz)
	}

	http.HandleFunc("/health", rateLimit(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
			"app":    appName,
		})
	}))

	http.HandleFunc("/time", rateLimit(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"server_time": time.Now().In(location).Format(time.RFC3339),
			"timezone":    tz,
			"uptime_sec":  int(time.Since(startTime).Seconds()),
		})
	}))

	log.Printf("%s running on port %s", appName, port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
