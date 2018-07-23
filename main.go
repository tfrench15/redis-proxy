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

// Command-line flags for setting up the proxy
var (
	redisAddr   = flag.String("redis", "localhost:6379", "address of the backing Redis")
	proxyAddr   = flag.String("proxy", "localhost:8080", "port on which the proxy listens")
	expiry      = flag.Int("expiry", 10, "time duration of a key in the cache")
	capacity    = flag.Int("capacity", 5, "size of the cache")
	redisClient = NewRedisClient()
)

func main() {
	flag.Parse()
	p := NewProxy(*redisAddr, *proxyAddr, "tcp", time.Duration(*expiry), *capacity)
	http.Handle("/", p)
	err := http.ListenAndServe(p.ProxyAddr, p)
	if err != nil {
		log.Fatal(err)
	}
}

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
	client, err := redis.Dial("tcp", *redisAddr)
	if err != nil {
		log.Fatalf("Error: cannot connect to Redis server")
		return nil
	}
	return client
}

// RetrieveFromCache checks the cache for a given key and, if
// it is present in the cache, returns the value and a bool.
func (p *Proxy) RetrieveFromCache(key string) (string, bool) {
	value, ok := p.Cache.Get(key)
	if !ok {
		return "", false
	}
	ci := value.(CachedItem)
	ciVal := ci.value
	if time.Now().Sub(ci.createdAt) < p.Expiry {
		return ciVal, true
	}
	return "", false
}

// RetrieveFromRedis checks Redis for a given key and, if it is
// present in Redis, adds the resultant key:value pair to the cache
func (p *Proxy) RetrieveFromRedis(key string) (string, bool) {
	value, err := redisClient.Cmd("GET", key).Str()
	if err != nil {
		fmt.Println(err)
		return "", false
	}
	if value == "" {
		fmt.Println("Error: key not found")
		return "", false
	}
	p.Cache.Add(key, CachedItem{value: value, createdAt: time.Now()})
	return value, true
}

// ServeHTTP implements the Handler interface and provides the core
// caching service
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		key := path.Base(r.URL.Path)
		value, ok := p.RetrieveFromCache(key) // check Cache first
		fmt.Println(value, ok)
		if ok {
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, value+"\n")
			io.WriteString(w, "Returned from Cache")
			return
		}
		value, ok = p.RetrieveFromRedis(key) // check Redis second
		fmt.Println(value, ok)
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
