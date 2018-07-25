PACKAGE := github.com/tfrench15/redis-proxy
DEPENDENCYONE := github.com/mediocregopher/radix.v2/redis
DEPENDENCYTWO := github.com/hashicorp/golang-lru


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
	$(GOGET) $(DEPENDENCYONE) $(DEPENDENCYTWO)


