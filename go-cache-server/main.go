package main

import (
	"fmt"
	"go-cache-server/Internal/cache"
	"go-cache-server/Internal/server"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	// Initialize caches
	redisCache := cache.NewRedisCache("localhost:6379", "", 0, 1*time.Minute)
	memcachedCache := cache.NewMemcachedCache("localhost:11211", 60)

	// Initialize server with caches
	srv := server.NewServer(redisCache, memcachedCache)

	// Create a new mux Router instance
	r := mux.NewRouter()

	// Define routes using the Server methods as handlers
	r.HandleFunc("/cache/{key}", srv.GetCache).Methods("GET")
	r.HandleFunc("/cache/{key}", srv.SetCache).Methods("POST")
	r.HandleFunc("/cache/{key}", srv.DeleteCache).Methods("DELETE")
	r.HandleFunc("/cache/clear", srv.ClearAllCaches).Methods("PUT")

	// Use your custom router instance with middleware or additional configuration if needed
	router := Router{r}

	// Start the HTTP server
	addr := "127.0.0.1:8080"
	fmt.Printf("Server started at %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}

// Router is a custom type for mux.Router that can be used to add additional methods if needed
type Router struct {
	*mux.Router
}