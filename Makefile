SRC = ./
BINARY = qq
DESTDIR = ~/.local/bin

all: build

build:
	go build -o bin/$(BINARY) $(SRC)

test:
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
