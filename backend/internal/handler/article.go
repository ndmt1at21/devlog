package handler

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ndmt1at21/devlog/backend/internal/apierr"
	"github.com/ndmt1at21/devlog/backend/internal/content"
	"github.com/ndmt1at21/devlog/backend/internal/domain"
)

// articleCreatePermission is the IAM permission (resource:action) required to
// publish articles. Provision it in the tenant and bind it to an author role.
const articleCreatePermission = "articles:create"

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

// featuredArticle returns every featured article (possibly empty) so the
// homepage can pin multiple heroes above the category filter.
func (a *API) featuredArticle(w http.ResponseWriter, r *http.Request) {
	arts, err := a.Store.Articles().Featured(r.Context())
	if err != nil {
		writeError(w, r, apierr.ErrArticleLoad.WithMessage("Không tải được bài viết nổi bật."))
		return
	}
	out := make([]articleSummary, 0, len(arts))
	for _, x := range arts {
		out = append(out, toSummary(x))
	}
	writeJSON(w, r, http.StatusOK, out)
}

func (a *API) categories(w http.ResponseWriter, r *http.Request) {
	cats, err := a.Store.Articles().Categories(r.Context())
	if err != nil {
		writeError(w, r, apierr.ErrCategoryList)
		return
	}
	writeJSON(w, r, http.StatusOK, cats)
}

// blockInput is the strict client-facing block shape. It deliberately has no
// `html` field (and decodeJSON rejects unknown fields), so clients can never
// inject markup — server-rendered Shiki HTML is the only HTML that ever exists.
type blockInput struct {
	Type    string   `json:"type"`
	Text    string   `json:"text"`
	Lang    string   `json:"lang"`
	Code    string   `json:"code"`
	Caption string   `json:"caption"`
	Steps   []string `json:"steps"`
	Items   []string `json:"items"`
	Ordered bool     `json:"ordered"`
	Src     string   `json:"src"`
	Alt     string   `json:"alt"`
}

type createArticleInput struct {
	Title    string   `json:"title"`
	Excerpt  string   `json:"excerpt"`
	Category string   `json:"category"`
	Tags     []string `json:"tags"`
	// Format selects the body payload: "markdown" (README-style, in Content) or
	// "blocks" (structured rich-text editor output, in Body).
	Format  string       `json:"format"`
	Content string       `json:"content"`
	Body    []blockInput `json:"body"`
}

// createArticle publishes a new article. Requires a session AND the IAM
// permission "articles:create", verified against IAM's policy decision
// endpoint on every call (the session's canWrite flag is only a UI hint).
func (a *API) createArticle(w http.ResponseWriter, r *http.Request) {
	u, ok := userFrom(r.Context())
	if !ok {
		writeError(w, r, apierr.ErrUnauthorized)
		return
	}
	if a.Auth == nil {
		writeError(w, r, apierr.ErrAuthNotConfigured)
		return
	}
	allowed, err := a.Auth.CheckPermissions(r.Context(), u.Access, []string{articleCreatePermission})
	if err != nil {
		writeError(w, r, apierr.ErrAuthUpstream)
		return
	}
	if !allowed {
		log.Printf("article create denied sub=%s trace=%s", u.Sub, traceIDFrom(r.Context()))
		writeError(w, r, apierr.ErrArticleForbidden)
		return
	}

	var in createArticleInput
	if !decodeJSON(w, r, &in) {
		return
	}
	in.Title = strings.TrimSpace(in.Title)
	in.Excerpt = strings.TrimSpace(in.Excerpt)
	in.Category = strings.TrimSpace(in.Category)
	switch {
	case in.Title == "":
		writeError(w, r, apierr.ErrValidation.WithMessage("Vui lòng nhập tiêu đề."))
		return
	case len(in.Title) > 300:
		writeError(w, r, apierr.ErrValidation.WithMessage("Tiêu đề tối đa 300 ký tự."))
		return
	case in.Category == "" || len(in.Category) > 80:
		writeError(w, r, apierr.ErrValidation.WithMessage("Vui lòng chọn danh mục hợp lệ."))
		return
	case len(in.Excerpt) > 500:
		writeError(w, r, apierr.ErrValidation.WithMessage("Mô tả ngắn tối đa 500 ký tự."))
		return
	}
	tags, err := normalizeTags(in.Tags)
	if err != nil {
		writeError(w, r, err)
		return
	}

	var raw []domain.Block
	switch in.Format {
	case "markdown":
		raw = content.BlocksFromMarkdown(in.Content)
	case "blocks":
		raw = make([]domain.Block, 0, len(in.Body))
		for _, b := range in.Body {
			raw = append(raw, domain.Block{
				Type: b.Type, Text: b.Text, Lang: b.Lang, Code: b.Code,
				Caption: b.Caption, Steps: b.Steps, Items: b.Items, Ordered: b.Ordered,
				Src: b.Src, Alt: b.Alt,
			})
		}
	default:
		writeError(w, r, apierr.ErrValidation.WithMessage("Định dạng nội dung phải là markdown hoặc blocks."))
		return
	}
	body, err := content.NormalizeBlocks(raw)
	var invalid *content.ErrInvalid
	if errors.As(err, &invalid) {
		writeError(w, r, apierr.ErrValidation.WithMessage(invalid.Reason+"."))
		return
	}
	if err != nil {
		writeError(w, r, apierr.ErrArticleCreate)
		return
	}
	if err := a.checkImageHosts(body); err != nil {
		writeError(w, r, err)
		return
	}

	excerpt := in.Excerpt
	if excerpt == "" {
		excerpt = content.DeriveExcerpt(body)
	}
	art := domain.Article{
		Category:    in.Category,
		Author:      u.Name,
		ReadTime:    content.ReadTime(body),
		PublishedAt: time.Now().UTC(),
		Title:       in.Title,
		Excerpt:     excerpt,
		Tags:        tags,
		Body:        body,
	}

	// Slug uniqueness: try the natural slug, then numbered variants. The unique
	// index is the arbiter, so concurrent creates can't race.
	base := content.Slugify(in.Title)
	created := domain.Article{}
	for i := 1; ; i++ {
		art.Slug = base
		if i > 1 {
			art.Slug = fmt.Sprintf("%s-%d", base, i)
		}
		created, err = a.Store.Articles().Create(r.Context(), art)
		if errors.Is(err, domain.ErrConflict) && i < 50 {
			continue
		}
		break
	}
	if err != nil {
		writeError(w, r, apierr.ErrArticleCreate)
		return
	}

	log.Printf("article created slug=%s sub=%s trace=%s", created.Slug, u.Sub, traceIDFrom(r.Context()))
	writeJSON(w, r, http.StatusCreated, articleDetail{
		articleSummary: toSummary(created),
		Body:           created.Body,
	})
}

// normalizeTags trims, de-duplicates, and bounds the tag list (≤ 8 × 40 chars).
func normalizeTags(in []string) ([]string, error) {
	out := make([]string, 0, len(in))
	seen := map[string]bool{}
	for _, t := range in {
		t = strings.TrimSpace(t)
		if t == "" || seen[strings.ToLower(t)] {
			continue
		}
		if len(t) > 40 {
			return nil, apierr.ErrValidation.WithMessage("Mỗi thẻ tối đa 40 ký tự.")
		}
		seen[strings.ToLower(t)] = true
		out = append(out, t)
	}
	if len(out) > 8 {
		return nil, apierr.ErrValidation.WithMessage("Tối đa 8 thẻ.")
	}
	return out, nil
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
