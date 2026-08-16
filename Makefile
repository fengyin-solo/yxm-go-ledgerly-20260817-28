.PHONY: build test clean lint

build:
	go build -o bin/ledgerly ./cmd/ledgerly

test:
	go test -v ./...

test-cover:
	go test -cover ./...

clean:
	rm -rf bin/

lint:
	go vet ./...

.DEFAULT_GOAL := build
