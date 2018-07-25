# Redis-Proxy

Redis-Proxy provides an HTTP service that serves as a lightweight, read-through LRU cache for a single backing Redis instance.

### Overview

The proxy listens for incoming HTTP requests on a configurable port, and maps HTTP GET requests to Redis GET commands using the base of the URL path as the key.  Non-GET requests return a Bad Request (400) HTTP error.

For example, a GET request issued to "http://localhost:8000/hi" parses 'hi' as the key to look up in the cache, or to fetch from Redis if it is not yet cached.  Keys not in Redis return a Status Not Found (404) HTTP error.

### Features

#### Configuration
The proxy is configurable via command-line flags.  You may customize:
1. `redisAddr`: The address of the backing Redis
2. `proxyAddr`: The port the proxy will listen on
3. `capacity`: The size of the cache
4. `expiry`: The duration of time a key will be cached

#### How the Code works

The proxy maps HTTP GETs to Redis GETs, as detailed above.  First, it checks the cache via the `RetrieveFromCache()` method, which returns a value and `true` if the key is cached; an empty string and `false` otherwise.

`CachedItem`'s are structs containing the key's value and the time it was cached as the struct's fields.  `RetrieveFromCache` checks the `createdAt` field to determine whether the key is still valid, or whether it has expired.

If the key is cached, the value and its source (i.e. Cache or Redis) are written to the ResponseWriter.

Else, if the key is not cached, the key's value is fetched directly from the backing Redis via the `RetrieveFromRedis()` method.  This method also adds the resultant `CachedItem` to the cache for subsequent GETs.

Finally, if the key is not found in Redis, an error is returned.

#### Tests

The proxy comes with unit tests, leveraging Go's testing framework in `main_test.go`.

#### Algorithmic Complexity

The cache, imported from Hashicorp's `simplelru` cache, is a doubly-linked list. Looking up a key has linear time complexity (O(n)).  Adding a key to the cache has constant time complexity (O(1)).

#### Instructions (TODO)

To test the cache, enter the top-level directory and run `make test`.  Note that I'm currently learning how to publish the Proxy to Docker, so that testing and running the cache is doable with a single click.