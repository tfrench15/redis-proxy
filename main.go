package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"

	"github.com/mediocregopher/radix.v2/redis"
)

const (
	cacheCapacity = 10
)

var cache map[string]interface{}

type response struct {
	value interface{}
	err   error
}

func helloProxy(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Welcome to my Redis Proxy")
}

func connectToRedis() (*redis.Client, error) {
	// TODO: grab a connection to the Redis instance
	client, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		return nil, errors.New("Error: cannot connect to Redis server")
	}
	return client, nil
}

func testCommands() {
	client, err := connectToRedis()
	if err != nil {
		fmt.Println(err)
	}
	sf, err := client.Cmd("GET", "SF").Str()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(sf)
	ny, err := client.Cmd("GET", "NY").Str()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(ny)
	love, err := client.Cmd("GET", "love").Str()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(love)
}

// Get handles incoming requests and maps HTTP GETs to
// Redis GETs
func Get(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		key := path.Base(r.URL.Path)
		fmt.Println(key)
		value, ok := cache[key]
		if !ok {
			// if len(cache) == cacheCapacity {
			// TODO: LRU eviction
			// TODO: Store key:value in cache
			// }
			client, err := connectToRedis()
			if err != nil {
				fmt.Println(err)
			}
			value, err = client.Cmd("GET", "key").Str()
			if err != nil {
				fmt.Println(err)
			}
			cache[key] = value
		}
	default:
		io.WriteString(w, "Sorry, given key does not exist.")
	}
}

func main() {
	http.HandleFunc("/", Get)
	testCommands()
	http.ListenAndServe("localhost:8000", nil)
}
