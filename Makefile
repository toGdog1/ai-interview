.PHONY: run test test-cover build tidy fmt

run:
	go run ./cmd/server

test:
	go test ./...

test-cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

build:
	go build -o bin/server ./cmd/server

tidy:
	go mod tidy

fmt:
	go fmt ./...
