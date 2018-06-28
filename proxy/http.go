package proxy

import (
	"context"
	"net/http"

	"github.com/gpsamson/segment-redis-proxy/cache"
	"github.com/julienschmidt/httprouter"
	redis "github.com/segmentio/redis-go"
)

// HTTPProxy is a Redis proxy that uses HTTP to communicate with clients and
// handle `GET` commands.
type HTTPProxy struct {
	Addr  string
	cache *cache.Cache
	redis *redis.Client
}

// NewHTTPProxy creates a new HTTP redis proxy instance.
func NewHTTPProxy(Addr string, cache *cache.Cache, redis *redis.Client) *HTTPProxy {
	return &HTTPProxy{Addr, cache, redis}
}

// Serve registers the handler functions for specific routes
// and start listening on the given address.
func (p *HTTPProxy) Serve() error {
	router := httprouter.New()
	router.GET("/:key", p.getHandler)

	return http.ListenAndServe(p.Addr, router)
}

// getHandler gets an item from the cache or the backing redis instance
// and returns it to the client - updating the cache if needed.
func (p *HTTPProxy) getHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	key := ps.ByName("key")

	// Return the cached value if it exists.
	if val, ok := p.cache.Get(key); ok {
		p.respond(w, http.StatusOK, http.DetectContentType(val), val)
		return
	}

	// Get the value from redis, cache it and return.
	if val, ok := p.cache.Get(key); ok {
		p.respond(w, http.StatusOK, http.DetectContentType(val), val)
		return
	}

	// Get the value from redis, cache it and return.
	resp := p.redis.Query(context.Background(), "GET", key)
	var val interface{}

	if resp.Next(&val) {
		// Nil value, peace out!
		if val == nil {
			p.respond(w, http.StatusNotFound, "text/plain", []byte(""))
			return
		}

		// Update the cache and return the value.
		p.cache.Set(key, val.([]byte))
		p.respond(w, http.StatusOK, http.DetectContentType(val.([]byte)), val.([]byte))
		return
	}

	if err := resp.Close(); err != nil {
		p.respond(w, http.StatusInternalServerError, "text/plain", []byte(err.Error()))
	}
}

// respond saves space and writes standard things like the headers, status code and body.
func (p *HTTPProxy) respond(w http.ResponseWriter, code int, content string, body []byte) {
	w.Header().Set("Content-Type", content)
	w.WriteHeader(code)
	w.Write(body)
}
