SRC = ./
BINARY = qq
DESTDIR = ~/.local/bin

all: build


build:
	go build -o bin/$(BINARY) $(SRC)

test: build
	./tests/test.sh

clean:
	rm bin/$(BINARY)

install: test
	mkdir -p $(DESTDIR)
	cp bin/$(BINARY) $(DESTDIR)

perf: build
	time "./tests/test.sh"

docker-push:
	docker buildx build --platform linux/amd64,linux/arm64 . -t jfryy/qq:latest --push

.PHONY: all test clean publish

