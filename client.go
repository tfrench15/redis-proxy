package main

import (
	"log"

	"github.com/mediocregopher/radix.v2/pool"
	"github.com/mediocregopher/radix.v2/redis"
)

// CreatePool creates and returns a pool of Redis connections
func CreatePool() (*pool.Pool, error) {
	p, err := pool.New("tcp", "localhost:6379", 10)
	if err != nil {
		log.Fatal("Error: pool creation failed")
		return nil, err
	}
	return p, nil
}

// Connect grabs a connection from the pool, and tests a few Redis commands
func Connect() *redis.Client {
	p, err := CreatePool()
	if err != nil {
		panic(err)
	}
	conn, err := p.Get()
	if err != nil {
		panic(err)
	}
	conn.Cmd("SET", "Hi", "there")
	conn.Cmd("SET", "Tim", "French")
	return conn
}
