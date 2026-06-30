package domain

import "context"

// ArticleRepository serves blog content.
type ArticleRepository interface {
	// List returns article summaries (Body omitted by callers as needed) matching
	// the filter, ordered by Ord ascending.
	List(ctx context.Context, f ArticleFilter) ([]Article, error)
	// GetBySlug returns a single article including its Body.
	GetBySlug(ctx context.Context, slug string) (Article, error)
	// Featured returns the single featured article.
	Featured(ctx context.Context) (Article, error)
	// Categories returns distinct category names ordered by first appearance.
	Categories(ctx context.Context) ([]string, error)
	// SeriesParts returns the ordered parts (by Part asc) of a series.
	SeriesParts(ctx context.Context, seriesSlug string) ([]Article, error)
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
	Subscriptions() SubscriptionRepository
	CoffeeOrders() CoffeeOrderRepository
	// Ping verifies the backing store is reachable.
	Ping(ctx context.Context) error
	Close() error
}
