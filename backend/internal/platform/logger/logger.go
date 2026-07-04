// Package logger builds the application's structured logger (log/slog, stdlib —
// no external dependency) and threads a request-scoped logger and trace id
// through context, so every line emitted while handling a request carries the
// same trace_id and correlates to a single client call.
package logger

import (
	"context"
	"io"
	"log/slog"
	"strings"
)

// Options configures New.
type Options struct {
	Level  string // debug|info|warn|error (default info)
	Format string // "json" for log aggregation, anything else for human-readable text
}

// New builds a slog.Logger writing to w. Format "json" emits one JSON object
// per line (production / log pipelines); "text" (the default) emits
// human-readable key=value lines for local development.
func New(w io.Writer, opts Options) *slog.Logger {
	handlerOpts := &slog.HandlerOptions{Level: ParseLevel(opts.Level)}
	var h slog.Handler
	if strings.ToLower(opts.Format) == "json" {
		h = slog.NewJSONHandler(w, handlerOpts)
	} else {
		h = slog.NewTextHandler(w, handlerOpts)
	}
	return slog.New(h)
}

// ParseLevel maps a level name to a slog.Level, defaulting to Info for unknown
// or empty input.
func ParseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

type ctxKey int

const (
	loggerKey ctxKey = iota
	traceKey
)

// WithTraceID returns a copy of ctx carrying the request's trace id.
func WithTraceID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, traceKey, id)
}

// TraceID returns the request's trace id, or "" if none is set.
func TraceID(ctx context.Context) string {
	id, _ := ctx.Value(traceKey).(string)
	return id
}

// WithContext returns a copy of ctx carrying a request-scoped logger, retrieved
// later by From.
func WithContext(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

// From returns the request-scoped logger stored on ctx, falling back to
// slog.Default() when none is present (e.g. outside an HTTP request).
func From(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey).(*slog.Logger); ok && l != nil {
		return l
	}
	return slog.Default()
}
