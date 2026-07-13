package memory

import (
	"context"
	"time"

	"github.com/ndmt1at21/devlog/backend/internal/domain"
	"github.com/ndmt1at21/devlog/backend/internal/platform/id"
)

// ---- users ----

type userRepo Store

func (r *userRepo) Create(_ context.Context, u domain.User) (domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, existing := range r.users {
		if existing.Email == u.Email || (u.GoogleSub != "" && existing.GoogleSub == u.GoogleSub) {
			return domain.User{}, domain.ErrConflict
		}
	}
	u.ID = id.NewV7()
	now := time.Now().UTC()
	u.CreatedAt, u.UpdatedAt = now, now
	r.users[u.ID] = u
	return u, nil
}

func (r *userRepo) GetByID(_ context.Context, userID string) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if u, ok := r.users[userID]; ok {
		return u, nil
	}
	return domain.User{}, domain.ErrNotFound
}

func (r *userRepo) GetByEmail(_ context.Context, email string) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.users {
		if u.Email == email {
			return u, nil
		}
	}
	return domain.User{}, domain.ErrNotFound
}

func (r *userRepo) GetByGoogleSub(_ context.Context, sub string) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if sub != "" {
		for _, u := range r.users {
			if u.GoogleSub == sub {
				return u, nil
			}
		}
	}
	return domain.User{}, domain.ErrNotFound
}

func (r *userRepo) Update(_ context.Context, u domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	stored, ok := r.users[u.ID]
	if !ok {
		return domain.ErrNotFound
	}
	u.CreatedAt = stored.CreatedAt
	u.UpdatedAt = time.Now().UTC()
	r.users[u.ID] = u
	return nil
}

func (r *userRepo) Count(_ context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.users), nil
}

// ---- refresh tokens ----

type refreshTokenRepo Store

func (r *refreshTokenRepo) Create(_ context.Context, t domain.RefreshToken) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if t.ID == "" {
		t.ID = id.NewV7()
	}
	r.refreshTokens[t.TokenHash] = t
	return nil
}

func (r *refreshTokenRepo) GetByHash(_ context.Context, hash string) (domain.RefreshToken, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if t, ok := r.refreshTokens[hash]; ok {
		return t, nil
	}
	return domain.RefreshToken{}, domain.ErrNotFound
}

func (r *refreshTokenRepo) Revoke(_ context.Context, hash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if t, ok := r.refreshTokens[hash]; ok && t.RevokedAt.IsZero() {
		t.RevokedAt = time.Now().UTC()
		r.refreshTokens[hash] = t
	}
	return nil
}
