package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRetrieveFromRedis(t *testing.T) {
	req, err := http.NewRequest("GET", "http://localhost:8080/sf", nil)
	if err != nil {
		t.Fatalf("Error making request: %v", err)
	}
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

	expected := "SanFrancisco\nReturned from Redis"
	if v := string(bytes.TrimSpace(b)); v != expected {
		t.Errorf("Unexpected value; got %v, expected %v", v, expected)
	}
}

func TestRetrieveFromCache(t *testing.T) {
	// We test by making two requests -- the first will retrieve a key
	// from Redis and add it to the Cache.  The second request will
	// retrieve the key directly from the cache.
	req1, err := http.NewRequest("GET", "http://localhost:8080/ny", nil)
	if err != nil {
		t.Fatalf("Error making reqeust: %v", err)
	}
	req2, err := http.NewRequest("GET", "http://localhost:8080/ny", nil)
	if err != nil {
		t.Fatalf("Error making request: %v", err)
	}

	rec1 := httptest.NewRecorder()
	rec2 := httptest.NewRecorder()
	ProxyRedis(rec1, req1)
	time.Sleep(1 * time.Second)
	ProxyRedis(rec2, req2)
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

func TestKeyNotInRedis(t *testing.T) {
	req, err := http.NewRequest("GET", "http://localhost:8080/hello", nil) // "hello" is not a key in Redis
	if err != nil {
		t.Fatalf("Error creating request: %v", err)
	}
	rec := httptest.NewRecorder()
	ProxyRedis(rec, req)
	if status := rec.Code; status != http.StatusNotFound {
		t.Errorf("Expected code %v; got code %v", http.StatusNotFound, rec.Code)
	}
}
