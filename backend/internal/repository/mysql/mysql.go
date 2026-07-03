// Package mysql implements domain.Store against MySQL using database/sql. It
// mirrors the IAM service's conventions: CHAR(36) ids, DATETIME(6) timestamps,
// JSON columns for arrays/objects, and mapError translating driver errors to
// domain sentinels. The DSN must enable parseTime=true so DATETIME scans into
// time.Time.
package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"

	"github.com/ndmt1at21/devlog/backend/internal/domain"
	"github.com/ndmt1at21/devlog/backend/internal/platform/id"
	"github.com/ndmt1at21/devlog/backend/internal/seed"
	"github.com/ndmt1at21/devlog/backend/migrations"
)

// Store is the MySQL-backed implementation of domain.Store.
type Store struct {
	db *sql.DB
}

// New opens a MySQL connection, applies migrations, seeds initial content when
// the articles table is empty, and returns the store.
func New(ctx context.Context, dsn string) (*Store, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping mysql: %w", err)
	}

	s := &Store{db: db}
	if err := s.migrate(ctx); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	if err := s.seedIfEmpty(ctx); err != nil {
		return nil, fmt.Errorf("seed: %w", err)
	}
	return s, nil
}

func (s *Store) Articles() domain.ArticleRepository           { return &articleRepo{s.db} }
func (s *Store) Series() domain.SeriesRepository              { return &seriesRepo{s.db} }
func (s *Store) Comments() domain.CommentRepository           { return &commentRepo{s.db} }
func (s *Store) Reactions() domain.ReactionRepository         { return &reactionRepo{s.db} }
func (s *Store) Subscriptions() domain.SubscriptionRepository { return &subRepo{s.db} }
func (s *Store) CoffeeOrders() domain.CoffeeOrderRepository   { return &coffeeRepo{s.db} }
func (s *Store) Users() domain.UserRepository                 { return &userRepo{s.db} }
func (s *Store) RefreshTokens() domain.RefreshTokenRepository { return &refreshTokenRepo{s.db} }
func (s *Store) Ping(ctx context.Context) error               { return s.db.PingContext(ctx) }
func (s *Store) Close() error                                 { return s.db.Close() }

// ---- migrations ----

func (s *Store) migrate(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version BIGINT PRIMARY KEY, applied_at DATETIME(6) NOT NULL)`); err != nil {
		return err
	}

	entries, err := migrations.MySQL.ReadDir("mysql")
	if err != nil {
		return err
	}
	var ups []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".up.sql") {
			ups = append(ups, e.Name())
		}
	}
	sort.Strings(ups)

	for _, name := range ups {
		version, err := strconv.ParseInt(strings.SplitN(name, "_", 2)[0], 10, 64)
		if err != nil {
			return fmt.Errorf("bad migration name %q: %w", name, err)
		}
		var exists int
		if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations WHERE version = ?`, version).Scan(&exists); err != nil {
			return err
		}
		if exists > 0 {
			continue
		}
		content, err := migrations.MySQL.ReadFile("mysql/" + name)
		if err != nil {
			return err
		}
		for _, stmt := range splitStatements(string(content)) {
			if _, err := s.db.ExecContext(ctx, stmt); err != nil {
				return fmt.Errorf("apply %s: %w", name, err)
			}
		}
		if _, err := s.db.ExecContext(ctx, `INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)`, version, timeNow()); err != nil {
			return err
		}
	}
	return nil
}

// ---- seeding ----

func (s *Store) seedIfEmpty(ctx context.Context) error {
	var n int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM articles`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	return s.Seed(ctx)
}

// Seed upserts the design's series and articles and inserts seed comments when
// none exist. Safe to run repeatedly.
func (s *Store) Seed(ctx context.Context) error {
	for _, sr := range seed.Series() {
		if _, err := s.db.ExecContext(ctx,
			`INSERT INTO series (slug, title, description, created_at) VALUES (?, ?, ?, ?)
			 ON DUPLICATE KEY UPDATE title=VALUES(title), description=VALUES(description)`,
			sr.Slug, sr.Title, sr.Description, timeNow()); err != nil {
			return err
		}
	}
	for _, a := range seed.Articles() {
		tags, err := marshalJSON(a.Tags)
		if err != nil {
			return err
		}
		body, err := marshalJSON(a.Body)
		if err != nil {
			return err
		}
		now := timeNow()
		if _, err := s.db.ExecContext(ctx,
			`INSERT INTO articles
			 (id, slug, ord, featured, category, author, read_time, published_at, title, excerpt, cover, tags, series_slug, series_part, part_title, body, created_at, updated_at)
			 VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
			 ON DUPLICATE KEY UPDATE ord=VALUES(ord), featured=VALUES(featured), category=VALUES(category),
			   author=VALUES(author), read_time=VALUES(read_time), published_at=VALUES(published_at),
			   title=VALUES(title), excerpt=VALUES(excerpt), cover=VALUES(cover), tags=VALUES(tags),
			   series_slug=VALUES(series_slug), series_part=VALUES(series_part), part_title=VALUES(part_title),
			   body=VALUES(body), updated_at=VALUES(updated_at)`,
			id.NewV7(), a.Slug, a.Ord, a.Featured, a.Category, a.Author, a.ReadTime, a.PublishedAt,
			a.Title, a.Excerpt, nullStr(a.Cover), tags, nullStr(a.Series), nullInt(a.Part), nullStr(a.PartTitle),
			body, now, now); err != nil {
			return mapError(err)
		}
	}

	var nc int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM comments`).Scan(&nc); err != nil {
		return err
	}
	if nc == 0 {
		for _, c := range seed.Comments() {
			if _, err := s.db.ExecContext(ctx,
				`INSERT INTO comments (id, article_slug, name, body, created_at) VALUES (?,?,?,?,?)`,
				id.NewV7(), c.ArticleSlug, c.Name, c.Body, c.CreatedAt); err != nil {
				return err
			}
		}
	}
	return nil
}

// splitStatements breaks a migration file into executable statements: it drops
// whole-line "--" comments first so semicolons inside comment prose don't cut a
// statement short (see 0003_reactions.up.sql), then splits on ";" and discards
// empty fragments. Constraints for migration authors: don't put ";" in inline
// trailing comments or string literals (none of our DDL needs either).
func splitStatements(sql string) []string {
	var b strings.Builder
	for line := range strings.Lines(sql) {
		if strings.HasPrefix(strings.TrimSpace(line), "--") {
			continue
		}
		b.WriteString(line)
	}
	var out []string
	for _, stmt := range strings.Split(b.String(), ";") {
		if strings.TrimSpace(stmt) == "" {
			continue
		}
		out = append(out, stmt)
	}
	return out
}

// ---- shared helpers ----

func timeNow() time.Time { return time.Now().UTC() }

const mysqlDuplicateEntry = 1062

func mapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}
	var myErr *mysql.MySQLError
	if errors.As(err, &myErr) && myErr.Number == mysqlDuplicateEntry {
		return domain.ErrConflict
	}
	return err
}

func marshalJSON(v any) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal json: %w", err)
	}
	return b, nil
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullInt(i int) any {
	if i == 0 {
		return nil
	}
	return i
}
