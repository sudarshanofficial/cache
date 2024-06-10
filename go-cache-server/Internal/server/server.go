package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"go-cache-server/Internal/cache"
	utils "go-cache-server/packageUtils/Utils"

	"github.com/gorilla/mux"
)

type Server struct {
	redisCache     *cache.RedisCache
	memcachedCache *cache.MemcachedCache
	mu             sync.Mutex
}

func NewServer(redisCache *cache.RedisCache, memcachedCache *cache.MemcachedCache) *Server {
	return &Server{
		redisCache:     redisCache,
		memcachedCache: memcachedCache,
	}
}

func (s *Server) GetCache(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["key"]

	// Try to get from Redis
	value, err := s.redisCache.Get(key)
	if err == nil {
		utils.RespondJSON(w, http.StatusOK, map[string]string{"key": key, "value": value})
		log.Printf("returning from RedisCache")
		return
	}

	// Try to get from Memcached
	value, err = s.memcachedCache.Get(key)
	if err == nil {
		utils.RespondJSON(w, http.StatusOK, map[string]string{"key": key, "value": value})
		log.Printf("returning from MemCachedCache")
		return
	}

	utils.RespondError(w, http.StatusNotFound, "Cache miss")
}

func (s *Server) SetCacheWithTTL(w http.ResponseWriter, r *http.Request) {
	ttl := mux.Vars(r)["ttl"]

	expireDuration, err := strconv.Atoi(ttl)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	// expireDuration = expireDuration* int(time.Second)
	mode := r.URL.Query().Get("cache")
	var payload struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.LogError("Error decoding JSON", err)
		utils.RespondError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	log.Printf("Setting cache for key: %s, value: %s", payload.Key, payload.Value)
	s.mu.Lock()
	defer s.mu.Unlock()
	switch mode {
	case "redis":
		if err := s.redisCache.SetWithTTL(payload.Key, payload.Value, time.Duration(expireDuration)*time.Second); err != nil {
			utils.LogError("Error setting Redis cache", err)
			utils.RespondError(w, http.StatusInternalServerError, "Failed to set Redis cache")
			return
		}
	case "memcatched":
		if err := s.memcachedCache.SetWithTTL(payload.Key, payload.Value, int32(expireDuration)); err != nil {
			utils.LogError("Error setting Memcached cache", err)
			utils.RespondError(w, http.StatusInternalServerError, "Failed to set Memcached cache")
			return
		}
	default:
		utils.RespondError(w, http.StatusInternalServerError, "Provide Valid Cache backend")

	}
	utils.RespondJSON(w, http.StatusOK, map[string]string{"status": "ok"})

}

func (s *Server) SetCache(w http.ResponseWriter, r *http.Request) {
	mode := r.URL.Query().Get("cache")
	var payload struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.LogError("Error decoding JSON", err)
		utils.RespondError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	log.Printf("Setting cache for key: %s, value: %s", payload.Key, payload.Value)
	s.mu.Lock()
	defer s.mu.Unlock()
	switch mode {
	case "redis":
		if err := s.redisCache.Set(payload.Key, payload.Value); err != nil {
			utils.LogError("Error setting Redis cache", err)
			utils.RespondError(w, http.StatusInternalServerError, "Failed to set Redis cache")
			return
		}
	case "memcatched":
		if err := s.memcachedCache.Set(payload.Key, payload.Value); err != nil {
			utils.LogError("Error setting Memcached cache", err)
			utils.RespondError(w, http.StatusInternalServerError, "Failed to set Memcached cache")
			return
		}
	default:
		utils.RespondError(w, http.StatusInternalServerError, "Provide Valid Cache backend")

	}

	utils.RespondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) DeleteCache(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["key"]
	mode := r.URL.Query().Get("cache")

	log.Printf("Deleting cache for key: %s", key+" "+mode)
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.redisCache.Delete(key); err != nil {
		utils.LogError("Error deleting Redis cache", err)
		utils.RespondError(w, http.StatusInternalServerError, "Failed to delete Redis cache")
		return
	}

	if err := s.memcachedCache.Delete(key); err != nil {
		utils.LogError("Error deleting Memcached cache", err)
		utils.RespondError(w, http.StatusInternalServerError, "Failed to delete Memcached cache")
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// func (s *Server) asyncDeleteCache(key string, w http.ResponseWriter) {

// 	if err := s.redisCache.Delete(key); err != nil {
// 		utils.LogError("Async: Error deleting Redis cache", err)
// 		utils.RespondError(w, http.StatusInternalServerError, "Failed to delete Redis cache")
// 		return
// 	}
// 	if err := s.memcachedCache.Delete(key); err != nil {
// 		utils.LogError("Async: Error deleting Memcached cache", err)
// 		utils.RespondError(w, http.StatusInternalServerError, "Failed to delete Memcached cache")
// 		return
// 	}
// 	utils.RespondJSON(w, http.StatusOK, map[string]string{"status": "ok"})

// }

func (s *Server) ClearAllCaches(w http.ResponseWriter, r *http.Request) {
	log.Printf("Clearing all caches")

	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.redisCache.Clear(); err != nil {
		utils.LogError("Error clearing Redis cache", err)
		utils.RespondError(w, http.StatusInternalServerError, "Failed to clear Redis cache")
		return
	}

	if err := s.memcachedCache.Clear(); err != nil {
		utils.LogError("Error clearing Memcached cache", err)
		utils.RespondError(w, http.StatusInternalServerError, "Failed to clear Memcached cache")
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
