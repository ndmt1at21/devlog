.PHONY: run build seed test tidy

run:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

# Load demo content into MySQL (needs DB_DRIVER=mysql and DB_DSN set).
seed:
	go run ./cmd/seed

test:
	go test ./...

tidy:
	go mod tidy
