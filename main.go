package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gpsamson/segment-redis-proxy/cache"
	"github.com/gpsamson/segment-redis-proxy/proxy"
	redis "github.com/segmentio/redis-go"
)

// Main validates any relevant env vars and creates a new proxy instance.
// - It might be useful to have a cache (and redis) interface. Mainly for the
// cache so we can swap in different implementations.
func main() {
	// Start up the cache and redis pool.
	cache := cache.New(EnvOrDefaultInt("CACHE_CAPACITY", 100), EnvOrDefaultInt("CACHE_TTTL", 300))
	redis := &redis.Client{
		Addr: EnvOrDefault("REDIS_ADDR", "localhost:6379"),
	}

	// Serve that proxy, double time!
	protocol := EnvOrDefault("PROTOCOL", "HTTP")
	port := EnvOrDefault("PORT", "8080")
	addr := fmt.Sprintf(":%v", port)
	var p proxy.Proxy
	switch protocol {
	case "HTTP":
		p = proxy.NewHTTPProxy(addr, cache, redis)
	case "RESP":
		p = proxy.NewRESPProxy(addr, cache, redis)
	default:
		log.Fatal("error: PROTOCOL env var must equal 'RESP' or 'HTTP'\n")
	}

	log.Printf("initializing %s redis proxy on port %s", protocol, port)
	log.Fatal(p.Serve())

	// TODO Graceful shutdown by SIGTERM or SIGINT. This requires `Proxy` to have `Close()`.
}

// EnvOrDefault returns the env var as, or `defaultValue` if the env var is not set.
func EnvOrDefault(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

// EnvOrDefaultInt returns the env var as an integer, or `defaultValue` if the env
// var is not set. Exits the program if the integer can't be parsed.
func EnvOrDefaultInt(key string, defaultValue int) int {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}

	i, err := strconv.Atoi(value)
	if err != nil {
		log.Fatalf("error: failed to parse env var as integer: %s\n", key)
	}

	return i
}
