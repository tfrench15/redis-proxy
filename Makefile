PACKAGE := github.com/tfrench15/redis-proxy

DEPENDENCIES := \
	github.com/mediocregopher/radix.v2/redis
	github.com/hashicorp/golang-lru

#Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

all: build test dep

build: 
	$(GOBUILD) 

test:
	$(GOTEST) -v 

dep:
	$(GOGET) $(DEPENDENCIES)

