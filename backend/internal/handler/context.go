package handler

import (
	"context"

	"github.com/ndmt1at21/devlog/backend/internal/platform/logger"
)

type ctxKey int

const (
	userCtxKey ctxKey = iota
)

// withTraceID / traceIDFrom store the request trace id in context. They
// delegate to the logger package so the trace id lives in one place and stays
// consistent with the request-scoped logger attached alongside it.
func withTraceID(ctx context.Context, id string) context.Context {
	return logger.WithTraceID(ctx, id)
}

// traceIDFrom returns the request's trace id, or "" if unset.
func traceIDFrom(ctx context.Context) string {
	return logger.TraceID(ctx)
}

// SessionUser is the authenticated user attached to a request by the auth
// middleware. It is nil for anonymous requests.
type SessionUser struct {
	Sub     string `json:"-"`
	Access  string `json:"-"` // IAM access token, for permission checks; never serialized
	Name    string `json:"name"`
	Email   string `json:"email"`
	Premium bool   `json:"premium"`
	// CanWrite mirrors the session's IAM "articles:create" snapshot so the UI can
	// show the editor entry point; POST /articles re-checks IAM authoritatively.
	CanWrite bool `json:"canWrite"`
}

func withUser(ctx context.Context, u *SessionUser) context.Context {
	return context.WithValue(ctx, userCtxKey, u)
}

func userFrom(ctx context.Context) (*SessionUser, bool) {
	u, ok := ctx.Value(userCtxKey).(*SessionUser)
	return u, ok && u != nil
}

func isPremium(ctx context.Context) bool {
	u, ok := userFrom(ctx)
	return ok && u.Premium
}
