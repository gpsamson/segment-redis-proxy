package proxy

import (
	"bytes"
	"testing"

	"github.com/gpsamson/segment-redis-proxy/cache"
	redis "github.com/segmentio/redis-go"
)

var respproxy = NewRESPProxy("", cache.New(5, 0), &redis.Client{Addr: ":6379"})

type MockedResponseWriter struct {
	Written interface{}
}

func (m *MockedResponseWriter) WriteStream(n int) error { return nil }
func (m *MockedResponseWriter) Write(v interface{}) error {
	m.Written = v
	return nil
}

func TestHandler(t *testing.T) {
	t.Run("unsupported command", func(t *testing.T) {
		t.Parallel()

		res := &MockedResponseWriter{}
		req := redis.NewRequest(":6379", "GET", redis.List("gabe"))

		respproxy.handler(res, req)

		if res.Written != nil {
			t.Fatal("GET not registered as a supported command")
		}
	})
}

func TestGet(t *testing.T) {
	// Setup test item.
	k, v := "philz", []byte("coffee")
	respproxy.cache.Set(k, v)

	t.Run("existing item", func(t *testing.T) {
		t.Parallel()
		res := &MockedResponseWriter{}
		args := redis.List(k)

		respproxy.get(res, args)

		if !bytes.Equal(res.Written.([]byte), v) {
			t.Fatalf("wrong value returned. actual: %v, expected: %v", res.Written, v)
		}
	})
	t.Run("nonexisting item", func(t *testing.T) {
		t.Parallel()
		res := &MockedResponseWriter{}
		args := redis.List("nope")

		respproxy.get(res, args)

		if res.Written != nil {
			t.Fatalf("wrong value returned. actual: %v, expected: %v", res.Written, nil)
		}
	})
}
