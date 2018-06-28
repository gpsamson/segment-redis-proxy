package proxy

// Proxy represents a proxy service that adds caching to a single redis instance.
type Proxy interface {
	Serve() error
}
