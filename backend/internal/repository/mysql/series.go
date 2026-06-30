package mysql

import (
	"context"
	"database/sql"

	"github.com/ndmt1at21/devlog/backend/internal/domain"
)

type seriesRepo struct{ db *sql.DB }

func (r *seriesRepo) GetBySlug(ctx context.Context, slug string) (domain.Series, error) {
	row := r.db.QueryRowContext(ctx, `SELECT slug, title, description FROM series WHERE slug = ?`, slug)
	var s domain.Series
	if err := row.Scan(&s.Slug, &s.Title, &s.Description); err != nil {
		return domain.Series{}, mapError(err)
	}
	return s, nil
}
