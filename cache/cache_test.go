package cache

import (
	"bytes"
	"container/list"
	"reflect"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	exp := &Cache{
		capacity: 100,
		ttl:      0,
		lru:      list.New(),
		cache:    make(map[string]*list.Element),
	}

	act := New(100, 0)

	if !reflect.DeepEqual(exp, act) {
		t.Fatalf("constructor failed. actual: %v, expected: %v", act, exp)
	}
}

func TestSet(t *testing.T) {
	key1, key2, val1, val2 := "the", "not the", []byte("office"), []byte("avengers")

	t.Run("get and set", func(t *testing.T) {
		t.Parallel()

		c := New(2, 0)

		c.Set(key1, val1)
		if val, ok := c.Get(key1); !ok || !bytes.Equal(val, val1) {
			t.Fatalf("failed to set item: %s, %s", key1, val1)
		}

		c.Set(key1, val2)
		if val, ok := c.Get(key1); !ok || !bytes.Equal(val, val2) {
			t.Fatalf("failed to update item: %s, %s", key1, val2)
		}
	})
	t.Run("lru eviction", func(t *testing.T) {
		t.Parallel()
		c := New(1, 0)

		c.Set(key1, val1)
		c.Set(key2, val2)

		if _, ok := c.cache[key2]; !ok || len(c.cache) > 1 {
			t.Fatalf("failed to evict least recently used item and maintain capacity")
		}
	})
}

func TestGet(t *testing.T) {
	k, v := "color", []byte("yellow")

	t.Run("existing item", func(t *testing.T) {
		t.Parallel()
		c := New(1, 0)
		c.Set(k, v)

		if act, ok := c.Get(k); !ok || !bytes.Equal(act, v) {
			t.Fatalf("failed to get item. actual: %v, expected: %v", act, v)
		}
	})

	t.Run("expired item", func(t *testing.T) {
		t.Parallel()
		c := New(1, 1)
		c.Set(k, v)

		time.Sleep(time.Second)

		if act, ok := c.Get(k); ok {
			t.Fatalf("received expired item. actual: %v, expected: %v", act, nil)
		}
	})

	t.Run("nonexisting item", func(t *testing.T) {
		t.Parallel()
		c := New(0, 0)

		if act, ok := c.Get("no"); ok || act != nil {
			t.Fatalf("nonexistent key returned item. actual: %v, expected: %v", act, nil)
		}
	})
}
