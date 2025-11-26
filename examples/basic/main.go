package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Mark-Bazylev/routerx"
)

func main() {
	// Create a router with logging middleware
	router := routerx.New().
		Use(LoggingMiddleware)

	// /api/v1 group
	apiV1 := router.
		Group("/api").
		Group("/v1")

	// Simple GET returning JSON
	apiV1.Path("/hello").
		Get(func(w http.ResponseWriter, r *http.Request) {
			JSON(w, 200, map[string]string{
				"message": "Hello from routerx!",
			})
		})

	// POST example
	apiV1.Path("/echo").
		Post(func(w http.ResponseWriter, r *http.Request) {
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				JSON(w, 400, map[string]string{"error": "invalid JSON"})
				return
			}
			JSON(w, 200, payload)
		})

	log.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

// LoggingMiddleware prints each request with execution time
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf(">> %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		log.Printf("<< %s %s (%s)", r.Method, r.URL.Path, time.Since(start))
	})
}

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Println("JSON encode error:", err)
	}
}
