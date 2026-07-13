package memory

import (
	"context"
	"testing"

	"github.com/ndmt1at21/devlog/backend/internal/domain"
)

func TestReactionSetStatusAndBookmarks(t *testing.T) {
	s := New()
	ctx := context.Background()
	r := s.Reactions()

	// Two users like the same article; one bookmarks it.
	if err := r.Set(ctx, domain.ReactionLike, "ai-agents", "u1", true); err != nil {
		t.Fatal(err)
	}
	if err := r.Set(ctx, domain.ReactionLike, "ai-agents", "u2", true); err != nil {
		t.Fatal(err)
	}
	// Re-liking is idempotent.
	if err := r.Set(ctx, domain.ReactionLike, "ai-agents", "u1", true); err != nil {
		t.Fatal(err)
	}
	if err := r.Set(ctx, domain.ReactionBookmark, "ai-agents", "u1", true); err != nil {
		t.Fatal(err)
	}

	st, err := r.Status(ctx, "ai-agents", "u1")
	if err != nil {
		t.Fatal(err)
	}
	if st.Likes != 2 || !st.Liked || !st.Bookmarked {
		t.Fatalf("u1 status = %+v, want likes=2 liked bookmarked", st)
	}

	// Anonymous sees the count but no personal state.
	st, err = r.Status(ctx, "ai-agents", "")
	if err != nil {
		t.Fatal(err)
	}
	if st.Likes != 2 || st.Liked || st.Bookmarked {
		t.Fatalf("anonymous status = %+v, want likes=2 only", st)
	}

	// Unlike removes only that user's like.
	if err := r.Set(ctx, domain.ReactionLike, "ai-agents", "u1", false); err != nil {
		t.Fatal(err)
	}
	st, _ = r.Status(ctx, "ai-agents", "u1")
	if st.Likes != 1 || st.Liked {
		t.Fatalf("after unlike status = %+v, want likes=1 not liked", st)
	}

	// Bookmarks list newest-first and only the user's own.
	if err := r.Set(ctx, domain.ReactionBookmark, "go-generics", "u1", true); err != nil {
		t.Fatal(err)
	}
	if err := r.Set(ctx, domain.ReactionBookmark, "other", "u2", true); err != nil {
		t.Fatal(err)
	}
	slugs, err := r.BookmarkedSlugs(ctx, "u1")
	if err != nil {
		t.Fatal(err)
	}
	if len(slugs) != 2 || slugs[0] != "go-generics" || slugs[1] != "ai-agents" {
		t.Fatalf("bookmarks = %v, want [go-generics ai-agents]", slugs)
	}
}
