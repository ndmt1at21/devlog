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

// reactionKey identifies one user's reaction of one kind on one article.
type reactionKey struct {
	slug, userID string
	kind         domain.ReactionKind
}

// Store is a concurrency-safe in-memory data store.
type Store struct {
	mu            sync.RWMutex
	articles      []domain.Article
	series        map[string]domain.Series
	comments      map[string][]domain.Comment
	reactions     map[reactionKey]time.Time
	subs          map[string]domain.Subscription
	coffee        map[string]domain.CoffeeOrder
	users         map[string]domain.User         // by id
	refreshTokens map[string]domain.RefreshToken // by token hash
}

// New returns a Store preloaded with the seed content.
func New() *Store {
	s := &Store{
		series:        make(map[string]domain.Series),
		comments:      make(map[string][]domain.Comment),
		reactions:     make(map[reactionKey]time.Time),
		subs:          make(map[string]domain.Subscription),
		coffee:        make(map[string]domain.CoffeeOrder),
		users:         make(map[string]domain.User),
		refreshTokens: make(map[string]domain.RefreshToken),
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
func (s *Store) Reactions() domain.ReactionRepository         { return (*reactionRepo)(s) }
func (s *Store) Subscriptions() domain.SubscriptionRepository { return (*subRepo)(s) }
func (s *Store) CoffeeOrders() domain.CoffeeOrderRepository   { return (*coffeeRepo)(s) }
func (s *Store) Users() domain.UserRepository                 { return (*userRepo)(s) }
func (s *Store) RefreshTokens() domain.RefreshTokenRepository { return (*refreshTokenRepo)(s) }
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

func summary(a domain.Article) domain.Article { a.Body = nil; a.Translations = nil; return a }

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

func (r *articleRepo) Featured(_ context.Context) ([]domain.Article, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []domain.Article
	for _, a := range r.articles { // already sorted by Ord
		if a.Featured {
			out = append(out, summary(a))
		}
	}
	return out, nil
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

func (r *articleRepo) Create(_ context.Context, a domain.Article) (domain.Article, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, ex := range r.articles {
		if ex.Slug == a.Slug {
			return domain.Article{}, domain.ErrConflict
		}
	}
	a.ID = id.NewV7()
	maxOrd := 0
	for _, ex := range r.articles {
		if ex.Ord > maxOrd {
			maxOrd = ex.Ord
		}
	}
	a.Ord = maxOrd + 1
	if a.PublishedAt.IsZero() {
		a.PublishedAt = time.Now().UTC()
	}
	r.articles = append(r.articles, a)
	return a, nil
}

func (r *articleRepo) Update(_ context.Context, a domain.Article) (domain.Article, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, ex := range r.articles {
		if ex.Slug == a.Slug {
			// Overlay only the editable fields, preserving the stored row's
			// identity, ordering and series placement.
			ex.Lang = a.Lang
			ex.Category = a.Category
			ex.Cover = a.Cover
			ex.CoverAlt = a.CoverAlt
			ex.ReadTime = a.ReadTime
			ex.Title = a.Title
			ex.Excerpt = a.Excerpt
			ex.Tags = a.Tags
			ex.Body = a.Body
			ex.Translations = a.Translations
			r.articles[i] = ex
			return ex, nil
		}
	}
	return domain.Article{}, domain.ErrNotFound
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

// ---- reactions ----

type reactionRepo Store

func (r *reactionRepo) Set(_ context.Context, kind domain.ReactionKind, slug, userID string, on bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := reactionKey{slug: slug, userID: userID, kind: kind}
	if !on {
		delete(r.reactions, key)
		return nil
	}
	if _, ok := r.reactions[key]; !ok {
		r.reactions[key] = time.Now().UTC()
	}
	return nil
}

func (r *reactionRepo) Status(_ context.Context, slug, userID string) (domain.ReactionStatus, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var st domain.ReactionStatus
	for k := range r.reactions {
		if k.slug != slug {
			continue
		}
		if k.kind == domain.ReactionLike {
			st.Likes++
			if userID != "" && k.userID == userID {
				st.Liked = true
			}
		}
		if k.kind == domain.ReactionBookmark && userID != "" && k.userID == userID {
			st.Bookmarked = true
		}
	}
	return st, nil
}

func (r *reactionRepo) BookmarkedSlugs(_ context.Context, userID string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	type saved struct {
		slug string
		at   time.Time
	}
	var marks []saved
	for k, at := range r.reactions {
		if k.kind == domain.ReactionBookmark && k.userID == userID {
			marks = append(marks, saved{slug: k.slug, at: at})
		}
	}
	sort.SliceStable(marks, func(i, j int) bool { return marks[i].at.After(marks[j].at) })
	out := make([]string, 0, len(marks))
	for _, m := range marks {
		out = append(out, m.slug)
	}
	return out, nil
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
