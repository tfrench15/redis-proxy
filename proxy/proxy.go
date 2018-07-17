// Package proxy defines the Proxy type and provides functions
// for the creation and configuration of new proxy instances
package proxy

import (
	"time"
)

// Proxy is the mechanism used to interface with the backing Redis
type Proxy struct {
	RedisAddr string
	ProxyAddr string
	Network   string
	Expiry    time.Duration
	Capacity  int
}

// New creates a new instance of Proxy with the supplied configuration variables
func New(redisAddr, proxyAddr, network string, timeLimit time.Duration, capacity int) *Proxy {
	return &Proxy{
		RedisAddr: redisAddr,
		ProxyAddr: proxyAddr,
		Network:   network,
		Expiry:    timeLimit,
		Capacity:  capacity,
	}
}
