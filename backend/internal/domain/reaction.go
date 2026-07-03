package domain

// ReactionKind is a per-user reaction on an article. Likes are public
// (aggregated into a count); bookmarks are personal (the reader's saved list).
type ReactionKind string

const (
	ReactionLike     ReactionKind = "like"
	ReactionBookmark ReactionKind = "bookmark"
)

// ParseReactionKind validates a client-supplied kind.
func ParseReactionKind(s string) (ReactionKind, bool) {
	switch ReactionKind(s) {
	case ReactionLike, ReactionBookmark:
		return ReactionKind(s), true
	}
	return "", false
}

// ReactionStatus is an article's aggregate like count plus the requesting
// user's own reaction state (false for anonymous readers).
type ReactionStatus struct {
	Likes      int
	Liked      bool
	Bookmarked bool
}
