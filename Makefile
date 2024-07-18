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

.PHONY: all test clean publish

