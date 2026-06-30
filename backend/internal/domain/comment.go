package domain

import "time"

// Comment is an anonymous, name-only comment on an article (no login required),
// matching the design.
type Comment struct {
	ID          string    `json:"-"`
	ArticleSlug string    `json:"-"`
	Name        string    `json:"name"`
	Body        string    `json:"text"`
	CreatedAt   time.Time `json:"-"`
}
