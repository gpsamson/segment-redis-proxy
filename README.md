# Gabriel Samson - Segment Redis Proxy 🥊
Segment Redis Proxy is a service which adds read-through LRU caching on top of
a single Redis backing instance.

The service supports **only** the [GET](https://redis.io/commands/get) command
using either HTTP or [RESP](https://redis.io/topics/protocol) to communicate with
clients.

## HTTP Usage
Once you set the environment variable, `PROTOCOL`, to `HTTP` you can send a `GET` request from to the proxy (assuming the `gabriel` key is set):
```
curl -X GET http://localhost:8080/gabriel
```

## RESP Usage
Similarly, once you set the environment variable, `PROTOCOL`, to `RESP` you can connect to the proxy using `redis-cli` (assuming the `gabriel` key is set):
```
$ redis-cli -p 8080
127.0.0.1:8080> GET gabriel
"samson"
127.0.0.1:8080> SET gabe samson
(error) ERR unsupported command 'SET'
127.0.0.1:8080>
```

## Architecture
- **Proxy** - The proxy component is responsible for handling GET requests using read-through mechanisms - tying together the cache and Redis components. The proxy has two variations: the HTTP proxy and the RESP proxy. These two variations differ by communication protocol but still have the same functionality. At start time, the proxy variation is selected based on the coniguration.
- **Cache** - The cache component is a fixed-size thread safe LRU cache with a global time-to-live. This implementation is based on the [groupcache](https://github.com/golang/groupcache/blob/master/lru/lru.go) package.
- **Redis** - The Redis component is exactly that, Redis. This also includes the client packages needed to communicate with the backing instance.

## Configuration
Environment Variables:
- `PORT`: The port the proxy should listen on. (default: `8080`)
- `PROTOCOL`: The communication protocol to use when listening to clients. One of: `RESP` or `HTTP`. (default: `HTTP`)
- `REDIS_ADDR`: The server address of the backing Redis instance. (default: `localhost:8080`)
- `CACHE_CAPACITY`: Maximum number of items the cache should retain. (default: `100`)
- `CACHE_TTL`: Maximum number of seconds an item should be retained in the cache before being evicted. (default: `300`)

## Build and Test
The following make targets are provided: `deps`, `format`, `lint`, `test`, and `container-image`.

You can run the service locally using Docker Compose:
```
$ docker-compose up
```

To build a production-ready image use:
```
$ make container-image
```
Container images are named `gpsamson/segment-redis-proxy` by default, and they are tagged with the hash of the current commit. You can override these using `IMAGE` and `TAG` environment variables.
## Algorithmic Complexity
Cache operation complexities:
- `Get()` - time complexity of `O(1)`
- `Set()` - time complexity of `O(1)`
## Time Estimate
- Research and Design (~1hr)
- Cache (~1.5hr)
- HTTP Proxy (~1hr)
- TCP Proxy (~1.5hr)
- Documentation, build files, etc. (~30mins)
## Requirements
The service meets all requirements except for the "sequential concurrent processing" requirement. Nonetheless, Go's built-in packages already ship with great concurrency support so the bonus "parallel concurrent processing" requirement was prioritized. For a service like this parallel requests are favorable.
## Thoughts and TODOs
- Specific TODOs are scattered throughout the code. 👀
- I decided to roll out a custom cache package because using any open-source package would require too many modifications to meet the caching requirements. It also seemed more appropriate since this is a coding excercise.
- Instead of having the server handlers implement the read-through mechanisms we can
create a `RTCache` that takes care of it - passing this as the cache dependency. It's almost like a wrapper around the LRU cache.
- If we wanted the LRU cache to evict items as their TTL comes to an end we could create a goroutine to periodically remove expired items. I'm not sure if this would hog the mutex too often and affect others.
