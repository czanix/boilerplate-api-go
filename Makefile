.PHONY: dev build test lint

dev:
	go run cmd/api/main.go

build:
	go build -o bin/api cmd/api/main.go

test:
	go test ./... -v -count=1

test-coverage:
	go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out

lint:
	golangci-lint run

docker-build:
	docker build -t czanix-api-go .
