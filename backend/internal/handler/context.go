package handler

import "context"

type ctxKey int

const userCtxKey ctxKey = iota

// SessionUser is the authenticated user attached to a request by the auth
// middleware. It is nil for anonymous requests.
type SessionUser struct {
	Sub     string `json:"-"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Premium bool   `json:"premium"`
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
