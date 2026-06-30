// Package memory is an in-memory implementation of domain.Store, seeded from the
// design content. It is the zero-infrastructure default for local development and
// mirrors the behavior of the MySQL backend.
package memory

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ndmt1at21/devlog/backend/internal/domain"
	"github.com/ndmt1at21/devlog/backend/internal/platform/id"
	"github.com/ndmt1at21/devlog/backend/internal/seed"
)

// Store is a concurrency-safe in-memory data store.
type Store struct {
	mu       sync.RWMutex
	articles []domain.Article
	series   map[string]domain.Series
	comments map[string][]domain.Comment
	subs     map[string]domain.Subscription
	coffee   map[string]domain.CoffeeOrder
}

// New returns a Store preloaded with the seed content.
func New() *Store {
	s := &Store{
		series:   make(map[string]domain.Series),
		comments: make(map[string][]domain.Comment),
		subs:     make(map[string]domain.Subscription),
		coffee:   make(map[string]domain.CoffeeOrder),
	}
	s.articles = seed.Articles()
	sort.SliceStable(s.articles, func(i, j int) bool { return s.articles[i].Ord < s.articles[j].Ord })
	for _, sr := range seed.Series() {
		s.series[sr.Slug] = sr
	}
	for _, c := range seed.Comments() {
		c.ID = id.NewV7()
		s.comments[c.ArticleSlug] = append(s.comments[c.ArticleSlug], c)
	}
	return s
}

func (s *Store) Articles() domain.ArticleRepository           { return (*articleRepo)(s) }
func (s *Store) Series() domain.SeriesRepository              { return (*seriesRepo)(s) }
func (s *Store) Comments() domain.CommentRepository           { return (*commentRepo)(s) }
func (s *Store) Subscriptions() domain.SubscriptionRepository { return (*subRepo)(s) }
func (s *Store) CoffeeOrders() domain.CoffeeOrderRepository   { return (*coffeeRepo)(s) }
func (s *Store) Ping(context.Context) error                   { return nil }
func (s *Store) Close() error                                 { return nil }

// ---- series ----

type seriesRepo Store

func (r *seriesRepo) GetBySlug(_ context.Context, slug string) (domain.Series, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if sr, ok := r.series[slug]; ok {
		return sr, nil
	}
	return domain.Series{}, domain.ErrNotFound
}

// ---- articles ----

type articleRepo Store

func summary(a domain.Article) domain.Article { a.Body = nil; return a }

func (r *articleRepo) List(_ context.Context, f domain.ArticleFilter) ([]domain.Article, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cat := strings.TrimSpace(f.Category)
	q := strings.ToLower(strings.TrimSpace(f.Query))
	out := make([]domain.Article, 0, len(r.articles))
	for _, a := range r.articles {
		if cat != "" && cat != "Tất cả" && a.Category != cat {
			continue
		}
		if q != "" && !matches(a, q) {
			continue
		}
		out = append(out, summary(a))
	}
	return out, nil
}

func (r *articleRepo) GetBySlug(_ context.Context, slug string) (domain.Article, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, a := range r.articles {
		if a.Slug == slug {
			return a, nil
		}
	}
	return domain.Article{}, domain.ErrNotFound
}

func (r *articleRepo) Featured(_ context.Context) (domain.Article, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, a := range r.articles {
		if a.Featured {
			return summary(a), nil
		}
	}
	if len(r.articles) > 0 {
		return summary(r.articles[0]), nil
	}
	return domain.Article{}, domain.ErrNotFound
}

func (r *articleRepo) Categories(_ context.Context) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	seen := map[string]bool{}
	out := []string{"Tất cả"}
	for _, a := range r.articles {
		if !seen[a.Category] {
			seen[a.Category] = true
			out = append(out, a.Category)
		}
	}
	return out, nil
}

func (r *articleRepo) SeriesParts(_ context.Context, seriesSlug string) ([]domain.Article, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []domain.Article
	for _, a := range r.articles {
		if a.Series == seriesSlug {
			out = append(out, summary(a))
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Part < out[j].Part })
	return out, nil
}

func matches(a domain.Article, q string) bool {
	hay := strings.ToLower(a.Title + " " + a.Excerpt + " " + a.Category + " " + strings.Join(a.Tags, " "))
	return strings.Contains(hay, q)
}

// ---- comments ----

type commentRepo Store

func (r *commentRepo) ListByArticle(_ context.Context, slug string) ([]domain.Comment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	src := r.comments[slug]
	out := make([]domain.Comment, len(src))
	copy(out, src)
	sort.SliceStable(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

func (r *commentRepo) Create(_ context.Context, c domain.Comment) (domain.Comment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c.ID = id.NewV7()
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now().UTC()
	}
	r.comments[c.ArticleSlug] = append(r.comments[c.ArticleSlug], c)
	return c, nil
}

// ---- subscriptions ----

type subRepo Store

func (r *subRepo) GetByUser(_ context.Context, userID string) (*domain.Subscription, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if sub, ok := r.subs[userID]; ok {
		return &sub, nil
	}
	return nil, domain.ErrNotFound
}

func (r *subRepo) Create(_ context.Context, sub domain.Subscription) (domain.Subscription, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	sub.ID = id.NewV7()
	if sub.CreatedAt.IsZero() {
		sub.CreatedAt = time.Now().UTC()
	}
	if sub.Status == "" {
		sub.Status = "active"
	}
	r.subs[sub.UserID] = sub
	return sub, nil
}

// ---- coffee orders ----

type coffeeRepo Store

func (r *coffeeRepo) Create(_ context.Context, o domain.CoffeeOrder) (domain.CoffeeOrder, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if o.ID == "" {
		o.ID = id.NewV7()
	}
	if o.CreatedAt.IsZero() {
		o.CreatedAt = time.Now().UTC()
	}
	if o.Status == "" {
		o.Status = domain.CoffeePending
	}
	r.coffee[o.ID] = o
	return o, nil
}

func (r *coffeeRepo) GetByID(_ context.Context, oid string) (domain.CoffeeOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if o, ok := r.coffee[oid]; ok {
		return o, nil
	}
	return domain.CoffeeOrder{}, domain.ErrNotFound
}

func (r *coffeeRepo) UpdateStatus(_ context.Context, oid, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	o, ok := r.coffee[oid]
	if !ok {
		return domain.ErrNotFound
	}
	o.Status = status
	if status == domain.CoffeeCompleted && o.CompletedAt == nil {
		now := time.Now().UTC()
		o.CompletedAt = &now
	}
	r.coffee[oid] = o
	return nil
}
