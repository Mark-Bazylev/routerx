package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Mark-Bazylev/routerx"
)

func main() {
	router := routerx.New().
		Use(LoggingMiddleware)

	apiV1 := router.
		Group("/api").
		Group("/v1")

	// Example: GET /api/v1/users/123
	apiV1.Path("/users/{id}").
		Get(getUserHandler).
		Patch(updateUserHandler).
		Delete(deleteUserHandler)

	log.Println("Server with path params running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

// getUserHandler demonstrates reading a path parameter from the request.
// Pattern: GET /api/v1/users/{id}
func getUserHandler(responseWriter http.ResponseWriter, request *http.Request) {
	userID := request.PathValue("id")

	JSON(responseWriter, http.StatusOK, map[string]any{
		"id":      userID,
		"message": "fetched user by id",
	})
}

// updateUserHandler demonstrates PATCH /api/v1/users/{id}
func updateUserHandler(responseWriter http.ResponseWriter, request *http.Request) {
	userID := request.PathValue("id")

	JSON(responseWriter, http.StatusOK, map[string]any{
		"id":      userID,
		"message": "updated user by id (demo only)",
	})
}

// deleteUserHandler demonstrates DELETE /api/v1/users/{id}
func deleteUserHandler(responseWriter http.ResponseWriter, request *http.Request) {
	userID := request.PathValue("id")

	JSON(responseWriter, http.StatusOK, map[string]any{
		"id":      userID,
		"message": "deleted user by id (demo only)",
	})
}

// LoggingMiddleware prints each request with execution time.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		startTime := time.Now()
		log.Printf(">> %s %s", request.Method, request.URL.Path)
		next.ServeHTTP(responseWriter, request)
		log.Printf("<< %s %s (%s)", request.Method, request.URL.Path, time.Since(startTime))
	})
}

func JSON(responseWriter http.ResponseWriter, statusCode int, data any) {
	responseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")
	responseWriter.WriteHeader(statusCode)

	if err := json.NewEncoder(responseWriter).Encode(data); err != nil {
		log.Println("JSON encode error:", err)
	}
}
