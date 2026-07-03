package handler

import "context"

type ctxKey int

const (
	userCtxKey ctxKey = iota
	traceCtxKey
)

func withTraceID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, traceCtxKey, id)
}

// traceIDFrom returns the request's trace id, or "" if unset.
func traceIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(traceCtxKey).(string)
	return id
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
