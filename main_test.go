package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// setup returns the same Proxy that will be used in production to test its functionality
func setup() *Proxy {
	p := NewProxy(*redisAddr, *proxyAddr, "tcp", time.Duration(*expiry)*time.Second, *capacity)
	redisClient.Cmd("SET", "sf", "SanFrancisco")                                                                          // set 'sf:SanFrancisco' key:value pair for testing
	redisClient.Cmd("SET", "ny", "NewYorkCity")                                                                           // set 'ny:NewYorkCity' key:value pair for testing
	redisClient.Cmd("SET", "old", "expired")                                                                              // set 'old:expired' key:value pair for testing
	p.Cache.Add("old", CachedItem{value: "expired", createdAt: time.Date(2000, time.November, 15, 1, 1, 1, 1, time.UTC)}) // test expiry works
	return p
}

func TestRetrieveFromRedis(t *testing.T) {
	p := setup()
	req, err := http.NewRequest("GET", "http://"+p.ProxyAddr+"/sf", nil)
	if err != nil {
		t.Fatalf("Error making request: %v", err)
	}
	rec := httptest.NewRecorder()
	p.ServeHTTP(rec, req)
	res := rec.Result()
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Errorf("Error reading response body: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		t.Errorf("Unexpected response code; got %v, expected %v", res.StatusCode, http.StatusOK)
	}

	expected := "SanFrancisco\nReturned from Redis"
	if v := string(bytes.TrimSpace(b)); v != expected {
		t.Errorf("Unexpected value; got %v, expected %v", v, expected)
	}
}

func TestRetrieveFromCache(t *testing.T) {
	// We test by making two requests -- the first will retrieve a key
	// from Redis and add it to the Cache.  The second request will
	// retrieve the key directly from the cache.
	p := setup()
	req1, err := http.NewRequest("GET", "http://"+p.ProxyAddr+"/ny", nil)
	if err != nil {
		t.Fatalf("Error making reqeust: %v", err)
	}
	req2, err := http.NewRequest("GET", "http://"+p.ProxyAddr+"/ny", nil)
	if err != nil {
		t.Fatalf("Error making request: %v", err)
	}

	rec1 := httptest.NewRecorder()
	rec2 := httptest.NewRecorder()
	p.ServeHTTP(rec1, req1)
	time.Sleep(1 * time.Second)
	p.ServeHTTP(rec2, req2)
	res2 := rec2.Result()
	defer res2.Body.Close()

	b2, err := ioutil.ReadAll(res2.Body)
	if err != nil {
		t.Errorf("Error reading response body: %v", err)
	}

	if res2.StatusCode != http.StatusOK {
		t.Errorf("Unexpected response code; got %v, expected %v", res2.StatusCode, http.StatusOK)
	}

	expected := "NewYorkCity\nReturned from Cache"
	if v := string(bytes.TrimSpace(b2)); v != expected {
		t.Errorf("Unexpected value; got %v, expected %v", v, expected)
	}
}

func TestExpiredKey(t *testing.T) {
	p := setup()
	req, err := http.NewRequest("GET", "http://"+p.ProxyAddr+"/old", nil)
	if err != nil {
		t.Fatalf("Error making request: %v", err)
	}
	rec := httptest.NewRecorder()
	p.ServeHTTP(rec, req)
	res := rec.Result()
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Errorf("Error reading response body: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		t.Errorf("Unexpected response code; got %v, expected %v", res.StatusCode, http.StatusOK)
	}

	expected := "expired\nReturned from Redis"
	if v := string(bytes.TrimSpace(b)); v != expected {
		t.Errorf("Unexpected value; got %v, expected %v", v, expected)
	}

}

func TestKeyNotInRedis(t *testing.T) {
	p := setup()
	req, err := http.NewRequest("GET", "http://"+p.ProxyAddr+"/hello", nil) // "hello" is not a key in Redis
	if err != nil {
		t.Fatalf("Error creating request: %v", err)
	}
	rec := httptest.NewRecorder()
	p.ServeHTTP(rec, req)
	if status := rec.Code; status != http.StatusNotFound {
		t.Errorf("Expected code %v; got code %v", http.StatusNotFound, rec.Code)
	}
}
