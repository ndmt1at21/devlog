package handler

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ndmt1at21/devlog/backend/internal/apierr"
	"github.com/ndmt1at21/devlog/backend/internal/domain"
)

// articleSummary is the list/card projection of an article (no body).
type articleSummary struct {
	Slug          string `json:"slug"`
	Title         string `json:"title"`
	Excerpt       string `json:"excerpt"`
	Category      string `json:"category"`
	Author        string `json:"author"`
	AuthorInitial string `json:"authorInitial"`
	Read          string `json:"read"`
	Date          string `json:"date"`
	// PublishedAt is the machine-readable (RFC 3339) counterpart of Date,
	// consumed by the frontend for canonical SEO metadata, JSON-LD
	// (datePublished) and the sitemap's lastmod.
	PublishedAt time.Time `json:"publishedAt"`
	Tags        []string  `json:"tags"`
	Cover       string    `json:"cover,omitempty"`
	Featured    bool      `json:"featured"`
	IsSeries    bool      `json:"isSeries"`
	Series      string    `json:"series,omitempty"`
	SeriesBadge string    `json:"seriesBadge,omitempty"`
}

type seriesPartDTO struct {
	Slug      string `json:"id"`
	Part      int    `json:"part"`
	PTitle    string `json:"ptitle"`
	IsCurrent bool   `json:"isCurrent"`
	Locked    bool   `json:"pLocked"`
}

type partLink struct {
	Slug   string `json:"id"`
	PTitle string `json:"ptitle"`
}

// articleDetail is the full article projection, including the (possibly
// truncated) body and series navigation.
type articleDetail struct {
	articleSummary
	Body            []domain.Block  `json:"body"`
	Locked          bool            `json:"locked"`
	InSeries        bool            `json:"inSeries"`
	SeriesTitle     string          `json:"seriesTitle,omitempty"`
	SeriesPartLabel string          `json:"seriesPartLabel,omitempty"`
	SeriesParts     []seriesPartDTO `json:"seriesParts,omitempty"`
	Prev            *partLink       `json:"prevPart,omitempty"`
	Next            *partLink       `json:"nextPart,omitempty"`
}

func toSummary(a domain.Article) articleSummary {
	s := articleSummary{
		Slug: a.Slug, Title: a.Title, Excerpt: a.Excerpt, Category: a.Category,
		Author: a.Author, AuthorInitial: initial(a.Author), Read: a.ReadTime,
		Date: formatVNDate(a.PublishedAt), PublishedAt: a.PublishedAt,
		Tags: a.Tags, Cover: a.Cover, Featured: a.Featured,
	}
	if a.Series != "" {
		s.IsSeries = true
		s.Series = a.Series
		s.SeriesBadge = fmt.Sprintf("Series · Phần %d", a.Part)
	}
	if s.Tags == nil {
		s.Tags = []string{}
	}
	return s
}

func (a *API) listArticles(w http.ResponseWriter, r *http.Request) {
	f := domain.ArticleFilter{
		Category: r.URL.Query().Get("category"),
		Query:    r.URL.Query().Get("q"),
	}
	arts, err := a.Store.Articles().List(r.Context(), f)
	if err != nil {
		writeError(w, r, apierr.ErrArticleList)
		return
	}
	out := make([]articleSummary, 0, len(arts))
	for _, x := range arts {
		out = append(out, toSummary(x))
	}
	writeJSON(w, r, http.StatusOK, out)
}

func (a *API) featuredArticle(w http.ResponseWriter, r *http.Request) {
	art, err := a.Store.Articles().Featured(r.Context())
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, r, apierr.ErrFeaturedNotFound)
		return
	}
	if err != nil {
		writeError(w, r, apierr.ErrArticleLoad.WithMessage("Không tải được bài viết nổi bật."))
		return
	}
	writeJSON(w, r, http.StatusOK, toSummary(art))
}

func (a *API) categories(w http.ResponseWriter, r *http.Request) {
	cats, err := a.Store.Articles().Categories(r.Context())
	if err != nil {
		writeError(w, r, apierr.ErrCategoryList)
		return
	}
	writeJSON(w, r, http.StatusOK, cats)
}

func (a *API) getArticle(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	art, err := a.Store.Articles().GetBySlug(r.Context(), slug)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, r, apierr.ErrArticleNotFound)
		return
	}
	if err != nil {
		writeError(w, r, apierr.ErrArticleLoad)
		return
	}

	premium := isPremium(r.Context())
	locked := art.Series != "" && art.Part > 1 && !premium

	body := art.Body
	if locked && len(body) > 2 {
		body = body[:2]
	}

	detail := articleDetail{
		articleSummary: toSummary(art),
		Body:           body,
		Locked:         locked,
	}

	if art.Series != "" {
		if meta, err := a.Store.Series().GetBySlug(r.Context(), art.Series); err == nil {
			parts, _ := a.Store.Articles().SeriesParts(r.Context(), art.Series)
			detail.InSeries = true
			detail.SeriesTitle = meta.Title
			detail.SeriesPartLabel = fmt.Sprintf("Phần %d / %d", art.Part, len(parts))
			for i, p := range parts {
				detail.SeriesParts = append(detail.SeriesParts, seriesPartDTO{
					Slug: p.Slug, Part: p.Part, PTitle: p.PartTitle,
					IsCurrent: p.Slug == art.Slug,
					Locked:    p.Part > 1 && !premium && p.Slug != art.Slug,
				})
				if p.Slug == art.Slug {
					if i > 0 {
						detail.Prev = &partLink{Slug: parts[i-1].Slug, PTitle: parts[i-1].PartTitle}
					}
					if i < len(parts)-1 {
						detail.Next = &partLink{Slug: parts[i+1].Slug, PTitle: parts[i+1].PartTitle}
					}
				}
			}
		}
	}

	writeJSON(w, r, http.StatusOK, detail)
}
