package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/ndmt1at21/devlog/backend/internal/domain"
	"github.com/ndmt1at21/devlog/backend/internal/platform/id"
)

type articleRepo struct{ db *sql.DB }

// scanner abstracts *sql.Row and *sql.Rows.
type scanner interface{ Scan(dest ...any) error }

const articleSummaryCols = `slug, ord, featured, category, author, read_time, published_at, title, excerpt, cover, tags, series_slug, series_part, part_title`

func scanArticleSummary(sc scanner) (domain.Article, error) {
	var a domain.Article
	var cover, seriesSlug, partTitle sql.NullString
	var seriesPart sql.NullInt64
	var tags []byte
	if err := sc.Scan(&a.Slug, &a.Ord, &a.Featured, &a.Category, &a.Author, &a.ReadTime,
		&a.PublishedAt, &a.Title, &a.Excerpt, &cover, &tags, &seriesSlug, &seriesPart, &partTitle); err != nil {
		return domain.Article{}, err
	}
	a.Cover = cover.String
	a.Series = seriesSlug.String
	a.Part = int(seriesPart.Int64)
	a.PartTitle = partTitle.String
	if len(tags) > 0 {
		_ = json.Unmarshal(tags, &a.Tags)
	}
	return a, nil
}

func (r *articleRepo) List(ctx context.Context, f domain.ArticleFilter) ([]domain.Article, error) {
	q := "SELECT " + articleSummaryCols + " FROM articles"
	var conds []string
	var args []any
	if cat := strings.TrimSpace(f.Category); cat != "" && cat != "Tất cả" {
		conds = append(conds, "category = ?")
		args = append(args, cat)
	}
	if query := strings.ToLower(strings.TrimSpace(f.Query)); query != "" {
		conds = append(conds, "LOWER(CONCAT(title,' ',excerpt,' ',category,' ',CAST(tags AS CHAR))) LIKE ?")
		args = append(args, "%"+query+"%")
	}
	if len(conds) > 0 {
		q += " WHERE " + strings.Join(conds, " AND ")
	}
	q += " ORDER BY ord"

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()

	var out []domain.Article
	for rows.Next() {
		a, err := scanArticleSummary(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *articleRepo) GetBySlug(ctx context.Context, slug string) (domain.Article, error) {
	row := r.db.QueryRowContext(ctx, "SELECT "+articleSummaryCols+", body FROM articles WHERE slug = ?", slug)
	var a domain.Article
	var cover, seriesSlug, partTitle sql.NullString
	var seriesPart sql.NullInt64
	var tags, body []byte
	if err := row.Scan(&a.Slug, &a.Ord, &a.Featured, &a.Category, &a.Author, &a.ReadTime,
		&a.PublishedAt, &a.Title, &a.Excerpt, &cover, &tags, &seriesSlug, &seriesPart, &partTitle, &body); err != nil {
		return domain.Article{}, mapError(err)
	}
	a.Cover = cover.String
	a.Series = seriesSlug.String
	a.Part = int(seriesPart.Int64)
	a.PartTitle = partTitle.String
	if len(tags) > 0 {
		_ = json.Unmarshal(tags, &a.Tags)
	}
	if len(body) > 0 {
		_ = json.Unmarshal(body, &a.Body)
	}
	return a, nil
}

func (r *articleRepo) Featured(ctx context.Context) ([]domain.Article, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT "+articleSummaryCols+" FROM articles WHERE featured = TRUE ORDER BY ord")
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	var out []domain.Article
	for rows.Next() {
		a, err := scanArticleSummary(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *articleRepo) Categories(ctx context.Context) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT category FROM articles GROUP BY category ORDER BY MIN(ord)")
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	out := []string{"Tất cả"}
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *articleRepo) Create(ctx context.Context, a domain.Article) (domain.Article, error) {
	tags, err := marshalJSON(a.Tags)
	if err != nil {
		return domain.Article{}, err
	}
	body, err := marshalJSON(a.Body)
	if err != nil {
		return domain.Article{}, err
	}
	a.ID = id.NewV7()
	now := timeNow()
	if a.PublishedAt.IsZero() {
		a.PublishedAt = now
	}
	// INSERT … SELECT computes Ord = max+1 atomically against the same table
	// (aggregate SELECT yields exactly one row even when the table is empty).
	if _, err := r.db.ExecContext(ctx, `INSERT INTO articles
		 (id, slug, ord, featured, category, author, read_time, published_at, title, excerpt, cover, tags, series_slug, series_part, part_title, body, created_at, updated_at)
		 SELECT ?, ?, COALESCE(MAX(ord), 0) + 1, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ? FROM articles`,
		a.ID, a.Slug, a.Featured, a.Category, a.Author, a.ReadTime, a.PublishedAt, a.Title, a.Excerpt,
		nullStr(a.Cover), tags, nullStr(a.Series), nullInt(a.Part), nullStr(a.PartTitle), body, now, now); err != nil {
		return domain.Article{}, mapError(err)
	}
	return a, nil
}

// Update rewrites the mutable columns of the article with the given Slug. Only
// the fields an author can change through the editor are touched; identity,
// ordering and series columns are left as-is. RowsAffected is not inspected: the
// default driver reports changed (not matched) rows, so an edit that changes
// nothing is indistinguishable from a missing slug — the handler verifies the
// article exists (and its ownership) before calling this.
func (r *articleRepo) Update(ctx context.Context, a domain.Article) (domain.Article, error) {
	tags, err := marshalJSON(a.Tags)
	if err != nil {
		return domain.Article{}, err
	}
	body, err := marshalJSON(a.Body)
	if err != nil {
		return domain.Article{}, err
	}
	if _, err := r.db.ExecContext(ctx, `UPDATE articles
		 SET category = ?, read_time = ?, title = ?, excerpt = ?, tags = ?, body = ?, updated_at = ?
		 WHERE slug = ?`,
		a.Category, a.ReadTime, a.Title, a.Excerpt, tags, body, timeNow(), a.Slug); err != nil {
		return domain.Article{}, mapError(err)
	}
	return a, nil
}

func (r *articleRepo) SeriesParts(ctx context.Context, seriesSlug string) ([]domain.Article, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT "+articleSummaryCols+" FROM articles WHERE series_slug = ? ORDER BY series_part", seriesSlug)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	var out []domain.Article
	for rows.Next() {
		a, err := scanArticleSummary(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}
