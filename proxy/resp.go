package proxy

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gpsamson/segment-redis-proxy/cache"
	redis "github.com/segmentio/redis-go"
)

// RESPProxy is a Redis proxy that uses the RESP (https://redis.io/topics/protocol)
// protocol to communicate with clients and handle `GET` commands.
type RESPProxy struct {
	Addr  string
	cache *cache.Cache
	redis *redis.Client
}

// NewRESPProxy creates a new TCP redis proxy instance.
func NewRESPProxy(Addr string, cache *cache.Cache, redis *redis.Client) *RESPProxy {
	return &RESPProxy{Addr, cache, redis}
}

// Serve registers the handler functions for specific routes
// and start listening on the given address.
func (p *RESPProxy) Serve() error {
	return redis.ListenAndServe(p.Addr, redis.HandlerFunc(p.handler))
}

// handler routes supports to the right handler and returns an error
// to the client if a command is unsupported.
func (p *RESPProxy) handler(res redis.ResponseWriter, req *redis.Request) {
	for _, cmd := range req.Cmds {
		switch strings.ToUpper(cmd.Cmd) {
		case "GET":
			p.get(res, cmd.Args)
		default:
			res.Write(fmt.Errorf("ERR unsupported command '%s'", cmd.Cmd))
		}
	}
}

// get gets an item from the cache or the backing redis instance
// and returns it to the client - updating the cache if needed.
func (p *RESPProxy) get(res redis.ResponseWriter, args redis.Args) {
	if args.Len() != 1 {
		res.Write(errors.New("ERR wrong number of arguments for 'get' command"))
	}

	key, keyerr := redis.String(args)
	if keyerr != nil {
		res.Write(fmt.Errorf("ERR can't parse key string: %v", args))
	}

	// Return the cached value if it exists.
	if val, ok := p.cache.Get(key); ok {
		res.Write(val)
		return
	}

	// Get the value from redis, cache it and return.
	resp := p.redis.Query(context.Background(), "GET", key)
	var val interface{}

	if resp.Next(&val) {
		// Nil value, peace out!
		if val == nil {
			res.Write(nil)
			return
		}

		// Update the cache and return the value.
		p.cache.Set(key, val.([]byte))
		res.Write(val.([]byte))
	}

	if err := resp.Close(); err != nil {
		res.Write(fmt.Errorf("ERR %s", err))
	}
}
