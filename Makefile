SRC = ./
BINARY = qq
DESTDIR = ~/.local/bin
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS = -X github.com/JFryy/qq/cli.Version=$(VERSION)

all: build

build:
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) $(SRC)

test: build
	./tests/test.sh
	go test ./... -v -cover

clean:
	rm -f bin/$(BINARY) qq_test_binary coverage.out coverage.html
	go clean -testcache

install: build test
	mkdir -p $(DESTDIR)
	cp bin/$(BINARY) $(DESTDIR)

docker-push:
	docker buildx build --platform linux/amd64,linux/arm64 . -t jfryy/qq:latest --push

.PHONY: all build test clean install docker-push
