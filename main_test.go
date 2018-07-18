package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
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
	req, err := http.NewRequest("GET", "http://localhost:8080/sf", nil)
	if err != nil {
		t.Fatalf("Error making request: %v", err)
	}
	fmt.Println(req)
	rec := httptest.NewRecorder()
	ProxyRedis(rec, req)
	res := rec.Result()
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Errorf("Error reading response body: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		t.Errorf("Unexpected response code; got %v, expected %v", res.StatusCode, http.StatusOK)
	}

	expected := "SanFrancisco"
	if v := string(bytes.TrimSpace(b)); v != expected {
		t.Errorf("Unexpected value; got %v, expected %v", v, expected)
	}
}

func TestKeyNotInRedis(t *testing.T) {
	req := httptest.NewRequest("GET", "localhost:8080/hello", nil) // "hello" is not a key in Redis
	rec := httptest.NewRecorder()
	ProxyRedis(rec, req)
	if status := rec.Code; status != http.StatusNotFound {
		t.Errorf("Expected code %v; got code %v", http.StatusNotFound, rec.Code)
	}
}
