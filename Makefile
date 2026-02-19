.PHONY: build test lint run

build:
	go build -o bin/server cmd/server/main.go

test:
	go test -v ./...

lint:
	golangci-lint run

run:
	go run cmd/server/main.go
