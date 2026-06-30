package mysql

import (
	"context"
	"database/sql"
	"time"

	"github.com/ndmt1at21/devlog/backend/internal/domain"
	"github.com/ndmt1at21/devlog/backend/internal/platform/id"
)

type subRepo struct{ db *sql.DB }

func (r *subRepo) GetByUser(ctx context.Context, userID string) (*domain.Subscription, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, plan, status, created_at FROM subscriptions WHERE user_id = ?`, userID)
	var s domain.Subscription
	if err := row.Scan(&s.ID, &s.UserID, &s.Plan, &s.Status, &s.CreatedAt); err != nil {
		return nil, mapError(err)
	}
	return &s, nil
}

func (r *subRepo) Create(ctx context.Context, s domain.Subscription) (domain.Subscription, error) {
	s.ID = id.NewV7()
	if s.CreatedAt.IsZero() {
		s.CreatedAt = time.Now().UTC()
	}
	if s.Status == "" {
		s.Status = "active"
	}
	if _, err := r.db.ExecContext(ctx,
		`INSERT INTO subscriptions (id, user_id, plan, status, created_at) VALUES (?,?,?,?,?)
		 ON DUPLICATE KEY UPDATE plan=VALUES(plan), status=VALUES(status)`,
		s.ID, s.UserID, s.Plan, s.Status, s.CreatedAt); err != nil {
		return domain.Subscription{}, mapError(err)
	}
	return s, nil
}
