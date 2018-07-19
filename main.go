package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"path"
	"time"

	"github.com/hashicorp/golang-lru"
	"github.com/mediocregopher/radix.v2/redis"
)

// Proxy defines the specification for a new read-through cache
type Proxy struct {
	RedisAddr string        // Address of backing Redis
	ProxyAddr string        // Port the proxy listens on
	Network   string        // tcp
	Expiry    time.Duration // duration for keeping a CachedItem
	Cache     *lru.Cache    // Cache
}

// CachedItem is the item type stored in the cache.
type CachedItem struct {
	value     string
	createdAt time.Time
}

// NewProxy returns a Proxy
func NewProxy(rAddr, pAddr, network string, expiry time.Duration, capacity int) *Proxy {
	cache, err := lru.New(capacity)
	if err != nil {
		log.Fatal("Error creating cache")
	}
	return &Proxy{
		RedisAddr: rAddr,
		ProxyAddr: pAddr,
		Network:   network,
		Expiry:    expiry,
		Cache:     cache,
	}
}

// NewRedisClient returns a new Redis client for use
func NewRedisClient() *redis.Client {
	client, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		log.Fatalf("Error: cannot connect to Redis server")
		return nil
	}
	return client
}

var (
	redisAddress = flag.String("ra", "localhost:6379", "address of the backing Redis")
	proxyAddress = flag.String("pa", "localhost:8080", "address Proxy listens on")
	network      = flag.String("network", "tcp", "communication protocol")
	expiry       = flag.Int("dur", 10, "duration for which keys are valid in the cache")
	capacity     = flag.Int("cap", 5, "capacity of the cache")
	proxy        = NewProxy(*redisAddress, *proxyAddress, *network, time.Duration(*expiry)*time.Second, *capacity)
	rc           = NewRedisClient()
)

// RetrieveFromCache checks the cache for a given key and, if
// it is present in the cache, returns the value and a bool.
func RetrieveFromCache(key string) (string, bool) {
	value, ok := proxy.Cache.Get(key)
	if !ok {
		return "", false
	}
	ci := value.(CachedItem)
	ciVal := ci.value
	if time.Now().Sub(ci.createdAt) < proxy.Expiry {
		return ciVal, true
	}
	return "", false
}

// RetrieveFromRedis checks Redis for a given key and, if it is
// present in Redis, adds the resultant key:value pair to the cache.
func RetrieveFromRedis(key string, rc *redis.Client) (string, bool) {
	value, err := rc.Cmd("GET", key).Str()
	if err != nil {
		fmt.Println(err)
		return "", false
	}
	if value == "" {
		fmt.Println("Error: key not found")
		return "", false
	}
	proxy.Cache.Add(key, CachedItem{value: value, createdAt: time.Now()})
	return value, true
}

// ProxyRedis is the core Handler for the service.
func ProxyRedis(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r)
	switch r.Method {
	case "GET":
		key := path.Base(r.URL.Path)
		value, ok := RetrieveFromCache(key) // check Cache first
		if ok {
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, value+"\n")
			io.WriteString(w, "Returned from Cache")
			return
		}
		value, ok = RetrieveFromRedis(key, rc) // check Redis second
		if ok {
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, value+"\n")
			io.WriteString(w, "Returned from Redis")
			return
		}
		http.Error(w, "Error: key not found", http.StatusNotFound) // can't find the key
	default:
		http.Error(w, "Please issue a GET request", http.StatusBadRequest)
	}
}

func main() {
	flag.Parse()
	http.HandleFunc("/", ProxyRedis)
	err := http.ListenAndServe("localhost:8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
