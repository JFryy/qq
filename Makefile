SRC = ./
ifeq ($(OS),Windows_NT)
	BINARY = qq.exe
else
	BINARY = qq
endif
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
DESTDIR = ~/.local/bin

all: build

build:
	go build -ldflags "-s -w -X 'github.com/JFryy/qq/cli.Version=$(VERSION)'" -o bin/$(BINARY) $(SRC)

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
