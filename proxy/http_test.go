package proxy

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gpsamson/segment-redis-proxy/cache"
	"github.com/julienschmidt/httprouter"
	redis "github.com/segmentio/redis-go"
)

var httpproxy = NewHTTPProxy("", cache.New(5, 0), &redis.Client{Addr: ":6379"})

func TestGetHandler(t *testing.T) {
	// Setup router.
	router := httprouter.New()
	router.GET("/:key", httpproxy.getHandler)

	// Setup test item.
	k, v := "philz", []byte("coffee")
	httpproxy.cache.Set(k, v)

	t.Run("200", func(t *testing.T) {
		t.Parallel()

		req, _ := http.NewRequest("GET", "/philz", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		if status := rr.Code; status != http.StatusOK {
			t.Fatalf("wrong status code returned. actual: %v, expected: %v", status, http.StatusOK)
		}
		if body := rr.Body.Bytes(); !bytes.Equal(body, v) {
			t.Fatalf("wrong value returned. actual: %v, expected: %v", body, v)
		}
	})
	t.Run("404", func(t *testing.T) {
		t.Parallel()

		req, _ := http.NewRequest("GET", "/starbucks", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		if status := rr.Code; status != http.StatusNotFound {
			t.Fatalf("wrong status code returned. actual: %v, expected: %v", status, http.StatusNotFound)
		}
	})
}
