# Backend (Go)

A minimal HTTP API server built with Go's standard library (`net/http`).

## Requirements

- Go 1.25+

## Getting started

```bash
# Run the server (defaults to port 8080)
make run
# or
go run ./cmd/server
```

The server reads the `PORT` environment variable (see `.env.example`).

## Endpoints

| Method | Path          | Description          |
| ------ | ------------- | -------------------- |
| GET    | `/api/health` | Health check         |
| GET    | `/api/hello`  | Sample hello message |

## Project layout

```
backend/
├── cmd/server/        # Application entry point (main)
└── internal/handler/  # HTTP routes & handlers
```

## Common commands

```bash
make build   # build binary to ./bin/server
make test    # run tests
make tidy    # tidy go.mod / go.sum
```
