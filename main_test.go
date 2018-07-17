package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mediocregopher/radix.v2/redis"
)

// req is the http.Request
// rec is the http.ResponseRecorder
// res is the response of rec

func setup() (*Proxy, *redis.Client) {
	// return a proxy
	p := NewProxy("localhost:6379", "localhost:8080/", "tcp", 5*time.Second, 5)
	rc := NewRedisClient()
	return p, rc
}

func TestRetrieveFromCache(t *testing.T) {
	req, err := http.NewRequest("GET", "localhost:8080/sf", nil)
	if err != nil {
		t.Fatalf("Error making request: %v", err)
	}
	fmt.Println(req)
	rec := httptest.NewRecorder()
	ProxyRedis(rec, req)
	res := rec.Result()
	fmt.Println(res)
}

func TestKeyNotInRedis(t *testing.T) {
	req := httptest.NewRequest("GET", "localhost:8080/hello", nil) // "hello" is not a key in Redis
	rec := httptest.NewRecorder()
	ProxyRedis(rec, req)
	if status := rec.Code; status != http.StatusNotFound {
		t.Errorf("Expected code %v; got code %v", http.StatusNotFound, rec.Code)
	}
}
