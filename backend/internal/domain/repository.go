package domain

import "context"

// ArticleRepository serves blog content.
type ArticleRepository interface {
	// List returns article summaries (Body omitted by callers as needed) matching
	// the filter, ordered by Ord ascending.
	List(ctx context.Context, f ArticleFilter) ([]Article, error)
	// GetBySlug returns a single article including its Body.
	GetBySlug(ctx context.Context, slug string) (Article, error)
	// Featured returns all featured articles (summaries) ordered by Ord
	// ascending; empty when none are flagged.
	Featured(ctx context.Context) ([]Article, error)
	// Categories returns distinct category names ordered by first appearance.
	Categories(ctx context.Context) ([]string, error)
	// SeriesParts returns the ordered parts (by Part asc) of a series.
	SeriesParts(ctx context.Context, seriesSlug string) ([]Article, error)
	// Create persists a new article, assigning its ID and Ord (appended after the
	// current maximum). Returns ErrConflict when the slug is already taken.
	Create(ctx context.Context, a Article) (Article, error)
	// Update persists edits to an existing article's mutable fields (title,
	// excerpt, category, tags, body, read time), matched by Slug; the slug,
	// author, publish time, ordering and series placement are left untouched.
	// Callers that need not-found semantics should GetBySlug first — the memory
	// store returns ErrNotFound for an unknown slug, but the MySQL store cannot
	// distinguish "no row" from "no change" and treats a missing slug as a no-op.
	Update(ctx context.Context, a Article) (Article, error)
}

// SeriesRepository serves series metadata.
type SeriesRepository interface {
	GetBySlug(ctx context.Context, slug string) (Series, error)
}

// CommentRepository stores anonymous comments.
type CommentRepository interface {
	ListByArticle(ctx context.Context, slug string) ([]Comment, error)
	Create(ctx context.Context, c Comment) (Comment, error)
}

// ReactionRepository stores per-user article reactions (likes and bookmarks).
type ReactionRepository interface {
	// Set adds (on=true) or removes (on=false) the user's reaction of the given
	// kind on an article. Re-setting an existing state is a no-op.
	Set(ctx context.Context, kind ReactionKind, slug, userID string, on bool) error
	// Status returns the article's like count plus the user's own reaction
	// state. userID may be "" (anonymous): Liked/Bookmarked are then false.
	Status(ctx context.Context, slug, userID string) (ReactionStatus, error)
	// BookmarkedSlugs returns the article slugs the user bookmarked, most
	// recently saved first.
	BookmarkedSlugs(ctx context.Context, userID string) ([]string, error)
}

// SubscriptionRepository stores demo Pro memberships.
type SubscriptionRepository interface {
	GetByUser(ctx context.Context, userID string) (*Subscription, error)
	Create(ctx context.Context, s Subscription) (Subscription, error)
}

// CoffeeOrderRepository stores "buy me a coffee" donation orders.
type CoffeeOrderRepository interface {
	Create(ctx context.Context, o CoffeeOrder) (CoffeeOrder, error)
	GetByID(ctx context.Context, id string) (CoffeeOrder, error)
	// UpdateStatus sets the order status; when status is completed it also stamps
	// CompletedAt. Provider ids captured at creation are preserved.
	UpdateStatus(ctx context.Context, id, status string) error
}

// Store bundles the repositories a request handler needs.
type Store interface {
	Articles() ArticleRepository
	Series() SeriesRepository
	Comments() CommentRepository
	Reactions() ReactionRepository
	Subscriptions() SubscriptionRepository
	CoffeeOrders() CoffeeOrderRepository
	Users() UserRepository
	RefreshTokens() RefreshTokenRepository
	// Ping verifies the backing store is reachable.
	Ping(ctx context.Context) error
	Close() error
}
