package domain

import "time"

// Block is a single piece of article body content. Type is one of:
// "p" (paragraph), "h" (heading), "code", "diagram", "quote", "list", "img".
// The "ad" block type is synthesized on the frontend, never stored.
// Text-bearing fields (Text, Items) may carry inline markdown spans (**b**,
// *i*, `c`, [t](url)) which the frontend renders safely as React elements.
type Block struct {
	Type    string   `json:"type"`
	Text    string   `json:"text,omitempty"`
	Lang    string   `json:"lang,omitempty"`
	Code    string   `json:"code,omitempty"`
	Caption string   `json:"caption,omitempty"`
	Steps   []string `json:"steps,omitempty"`
	Items   []string `json:"items,omitempty"`
	Ordered bool     `json:"ordered,omitempty"`
	// Src/Alt describe an "img" block: Src is the image URL (must point at the
	// configured public image origin), Alt the accessibility description.
	Src string `json:"src,omitempty"`
	Alt string `json:"alt,omitempty"`
}

// Series groups multi-part articles. Parts are derived from Article.Part.
type Series struct {
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// Article is the core content entity. Slug is the public identifier used in URLs
// and the API (e.g. "ai-agents"); ID is the internal CHAR(36) UUID.
type Article struct {
	ID       string `json:"-"`
	Slug     string `json:"slug"`
	Ord      int    `json:"-"`
	Featured bool   `json:"featured"`
	Category string `json:"category"`
	Author   string `json:"author"`
	// AuthorID is the stable user id of the author (the session subject stamped
	// at create time). Empty for seed/imported content that has no owning
	// account. Never serialized — it is an internal authorization key, not
	// public data; the API exposes an Editable flag instead.
	AuthorID    string    `json:"-"`
	ReadTime    string    `json:"read"`
	PublishedAt time.Time `json:"-"`
	Title       string    `json:"title"`
	Excerpt     string    `json:"excerpt"`
	Cover       string    `json:"cover,omitempty"`
	CoverAlt    string    `json:"coverAlt,omitempty"`
	Tags        []string  `json:"tags"`
	Series      string    `json:"series,omitempty"`
	Part        int       `json:"part,omitempty"`
	PartTitle   string    `json:"ptitle,omitempty"`
	Body        []Block   `json:"body"`
}

// ArticleFilter narrows a List query. An empty/"Tất cả" Category and empty Query
// mean "no filter".
type ArticleFilter struct {
	Category string
	Query    string
}
