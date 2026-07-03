# syntax=docker/dockerfile:1

# --- build stage -----------------------------------------------------------
FROM golang:1.25 AS build
WORKDIR /src

# Cache module downloads first so source-only changes don't re-download deps.
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

# Build a fully static binary. SQL migrations are embedded via go:embed, so the
# runtime image needs nothing but the binary itself.
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# --- runtime stage ---------------------------------------------------------
# distroless/static ships CA certificates (needed for outbound TLS to Stripe,
# MoMo, IAM, Google) and runs as a non-root user with no shell/package manager.
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/server /app/server
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/server"]
