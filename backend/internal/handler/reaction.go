package handler

import (
	"errors"
	"net/http"

	"github.com/ndmt1at21/devlog/backend/internal/apierr"
	"github.com/ndmt1at21/devlog/backend/internal/domain"
)

// reactionDTO mirrors domain.ReactionStatus for the API.
type reactionDTO struct {
	Likes      int  `json:"likes"`
	Liked      bool `json:"liked"`
	Bookmarked bool `json:"bookmarked"`
}

func toReactionDTO(st domain.ReactionStatus) reactionDTO {
	return reactionDTO{Likes: st.Likes, Liked: st.Liked, Bookmarked: st.Bookmarked}
}

// getReactions is public: anonymous readers see the like count with
// liked/bookmarked false.
func (a *API) getReactions(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	var userID string
	if u, ok := userFrom(r.Context()); ok {
		userID = u.Sub
	}
	st, err := a.Store.Reactions().Status(r.Context(), slug, userID)
	if err != nil {
		writeError(w, r, apierr.ErrReactionLoad)
		return
	}
	writeJSON(w, r, http.StatusOK, toReactionDTO(st))
}

// setReaction handles PUT (set) and DELETE (unset) of a like/bookmark. Login is
// required so reactions are deduplicated per user.
func (a *API) setReaction(w http.ResponseWriter, r *http.Request) {
	u, ok := userFrom(r.Context())
	if !ok {
		writeError(w, r, apierr.ErrUnauthorized)
		return
	}
	kind, ok := domain.ParseReactionKind(r.PathValue("kind"))
	if !ok {
		writeError(w, r, apierr.ErrReactionKind)
		return
	}
	slug := r.PathValue("slug")
	on := r.Method != http.MethodDelete

	// Ensure the article exists so reactions can't be orphaned (removals stay
	// idempotent and skip the check).
	if on {
		if _, err := a.Store.Articles().GetBySlug(r.Context(), slug); err != nil {
			writeError(w, r, apierr.ErrArticleNotFound)
			return
		}
	}

	if err := a.Store.Reactions().Set(r.Context(), kind, slug, u.Sub, on); err != nil {
		writeError(w, r, apierr.ErrReactionUpdate)
		return
	}
	st, err := a.Store.Reactions().Status(r.Context(), slug, u.Sub)
	if err != nil {
		writeError(w, r, apierr.ErrReactionLoad)
		return
	}
	writeJSON(w, r, http.StatusOK, toReactionDTO(st))
}

// myBookmarks lists the signed-in user's saved articles, newest first.
func (a *API) myBookmarks(w http.ResponseWriter, r *http.Request) {
	u, ok := userFrom(r.Context())
	if !ok {
		writeError(w, r, apierr.ErrUnauthorized)
		return
	}
	slugs, err := a.Store.Reactions().BookmarkedSlugs(r.Context(), u.Sub)
	if err != nil {
		writeError(w, r, apierr.ErrBookmarkList)
		return
	}
	out := make([]articleSummary, 0, len(slugs))
	for _, slug := range slugs {
		art, err := a.Store.Articles().GetBySlug(r.Context(), slug)
		if errors.Is(err, domain.ErrNotFound) {
			continue // article removed since it was saved
		}
		if err != nil {
			writeError(w, r, apierr.ErrBookmarkList)
			return
		}
		out = append(out, toSummary(art))
	}
	writeJSON(w, r, http.StatusOK, out)
}
