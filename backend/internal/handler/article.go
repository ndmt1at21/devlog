package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ndmt1at21/devlog/backend/internal/apierr"
	"github.com/ndmt1at21/devlog/backend/internal/content"
	"github.com/ndmt1at21/devlog/backend/internal/domain"
	"github.com/ndmt1at21/devlog/backend/internal/platform/logger"
)

// articleCreatePermission is the IAM permission (resource:action) required to
// publish articles. Provision it in the tenant and bind it to an author role.
const articleCreatePermission = "articles:create"

// maxCoverURLLen bounds a cover image URL to the width of the articles.cover
// column (VARCHAR(500)).
const maxCoverURLLen = 500

// maxCoverAltLen bounds the cover alt text to the width of the
// articles.cover_alt column (VARCHAR(300)), matching the in-article image alt.
const maxCoverAltLen = 300

// articleSummary is the list/card projection of an article (no body).
type articleSummary struct {
	Slug string `json:"slug"`
	// Lang is the language of this summary's title/excerpt ("vi" | "en").
	Lang          string `json:"lang"`
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
	CoverAlt    string    `json:"coverAlt,omitempty"`
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

// translationDTO is one non-primary language variant in the detail response.
type translationDTO struct {
	Title    string         `json:"title"`
	Excerpt  string         `json:"excerpt"`
	CoverAlt string         `json:"coverAlt,omitempty"`
	Body     []domain.Block `json:"body"`
}

// articleDetail is the full article projection, including the (possibly
// truncated) body and series navigation. The base fields (title/excerpt/body)
// carry the primary language (Lang); Translations holds the other variants so
// the reader can toggle language client-side without a refetch.
type articleDetail struct {
	articleSummary
	Body []domain.Block `json:"body"`
	// Translations maps a locale ("en", …) to that language's content. Omitted
	// when the article exists in one language only.
	Translations map[string]translationDTO `json:"translations,omitempty"`
	// AvailableLangs lists every language the article is published in (the
	// primary Lang first), for rendering the language switcher.
	AvailableLangs []string `json:"availableLangs"`
	// Editable is true when the requesting user is the article's author (matched
	// by user id). It lets the client show the edit affordance without ever
	// seeing raw author ids; PUT re-checks ownership authoritatively.
	Editable        bool            `json:"editable"`
	Locked          bool            `json:"locked"`
	InSeries        bool            `json:"inSeries"`
	SeriesTitle     string          `json:"seriesTitle,omitempty"`
	SeriesPartLabel string          `json:"seriesPartLabel,omitempty"`
	SeriesParts     []seriesPartDTO `json:"seriesParts,omitempty"`
	Prev            *partLink       `json:"prevPart,omitempty"`
	Next            *partLink       `json:"nextPart,omitempty"`
}

// defaultLang is the language assumed for content that predates the lang
// column (seed/imported rows), matching the frontend's DEFAULT_LOCALE.
const defaultLang = "vi"

// validLocales are the languages an article may be authored in. Kept in sync
// with the frontend Locale type ("vi" | "en").
var validLocales = map[string]bool{"vi": true, "en": true}

// langOr normalizes an empty language to the default.
func langOr(lang string) string {
	if lang == "" {
		return defaultLang
	}
	return lang
}

func toSummary(a domain.Article) articleSummary {
	s := articleSummary{
		Slug: a.Slug, Lang: langOr(a.Lang), Title: a.Title, Excerpt: a.Excerpt, Category: a.Category,
		Author: a.Author, AuthorInitial: initial(a.Author), Read: a.ReadTime,
		Date: formatVNDate(a.PublishedAt), PublishedAt: a.PublishedAt,
		Tags: a.Tags, Cover: a.Cover, CoverAlt: a.CoverAlt, Featured: a.Featured,
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

// editableBy reports whether the request's user is the article's author. It
// keys off the stable author id (empty for ownerless seed content, which no
// signed-in user can match), never the display name.
func editableBy(ctx context.Context, art domain.Article) bool {
	u, ok := userFrom(ctx)
	return ok && art.AuthorID != "" && art.AuthorID == u.Sub
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

// localizedInput is one language's content within a create/update payload. The
// primary language is carried by createArticleInput's own fields; each entry in
// its Translations is an additional language.
type localizedInput struct {
	Lang     string `json:"lang"`
	Title    string `json:"title"`
	Excerpt  string `json:"excerpt"`
	CoverAlt string `json:"coverAlt"`
	// Format selects the body payload: "markdown" (README-style, in Content) or
	// "blocks" (structured rich-text editor output, in Body).
	Format  string       `json:"format"`
	Content string       `json:"content"`
	Body    []blockInput `json:"body"`
}

type createArticleInput struct {
	// Shared across all languages.
	Category string   `json:"category"`
	Cover    string   `json:"cover"`
	Tags     []string `json:"tags"`
	// Primary-language content. Lang defaults to "vi" when omitted (legacy
	// single-language clients).
	Lang     string `json:"lang"`
	Title    string `json:"title"`
	Excerpt  string `json:"excerpt"`
	CoverAlt string `json:"coverAlt"`
	// Format selects the body payload: "markdown" (README-style, in Content) or
	// "blocks" (structured rich-text editor output, in Body).
	Format  string       `json:"format"`
	Content string       `json:"content"`
	Body    []blockInput `json:"body"`
	// Translations carries the non-primary language variants (may be empty:
	// "publish now, translate later").
	Translations []localizedInput `json:"translations"`
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
		logger.From(r.Context()).Warn("article create denied", "sub", u.Sub)
		writeError(w, r, apierr.ErrArticleForbidden)
		return
	}

	var in createArticleInput
	if !decodeJSON(w, r, &in) {
		return
	}
	c, err := a.prepareArticle(in)
	if err != nil {
		writeError(w, r, err)
		return
	}
	art := domain.Article{
		Lang:         c.Lang,
		Category:     c.Category,
		Author:       u.Name,
		AuthorID:     u.Sub,
		Cover:        c.Cover,
		CoverAlt:     c.CoverAlt,
		ReadTime:     content.ReadTime(c.Body),
		PublishedAt:  time.Now().UTC(),
		Title:        c.Title,
		Excerpt:      c.Excerpt,
		Tags:         c.Tags,
		Body:         c.Body,
		Translations: c.Translations,
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

	logger.From(r.Context()).Info("article created", "slug", created.Slug, "sub", u.Sub)
	detail := articleDetail{
		articleSummary: toSummary(created),
		Body:           created.Body,
		Editable:       editableBy(r.Context(), created),
	}
	detail.Translations, detail.AvailableLangs = buildTranslations(created, false)
	writeJSON(w, r, http.StatusCreated, detail)
}

// articleContent is the validated, normalized result of a create/update payload:
// the shared metadata plus the primary language's content and any translations —
// ready to persist onto a domain.Article.
type articleContent struct {
	Category     string
	Cover        string
	Tags         []string
	Lang         string
	Title        string
	Excerpt      string
	CoverAlt     string
	Body         []domain.Block
	Translations map[string]domain.Translation
}

// prepareArticle validates a create/update payload and normalizes it. Shared
// metadata (category, cover, tags) is validated once; the primary language and
// each translation are validated independently, so an article can be published
// in one language and gain the other later. On invalid input it returns an
// *apierr.Error (already localized). Blank excerpts are derived per language.
func (a *API) prepareArticle(in createArticleInput) (articleContent, error) {
	category := strings.TrimSpace(in.Category)
	if category == "" || len(category) > 80 {
		return articleContent{}, apierr.ErrValidation.WithMessage("Vui lòng chọn danh mục hợp lệ.")
	}
	tags, err := normalizeTags(in.Tags)
	if err != nil {
		return articleContent{}, err
	}

	// Cover is shared across languages. Optional; when present it must be a
	// well-formed image URL under the configured image origin (i.e. it came
	// through the upload flow), and short enough for the cover column.
	cover := strings.TrimSpace(in.Cover)
	if cover != "" {
		if len(cover) > maxCoverURLLen {
			return articleContent{}, apierr.ErrValidation.WithMessage("Đường dẫn ảnh bìa quá dài.")
		}
		if err := content.ValidateImageURL(cover); err != nil {
			var invalid *content.ErrInvalid
			if errors.As(err, &invalid) {
				return articleContent{}, apierr.ErrValidation.WithMessage("Ảnh bìa: " + invalid.Reason + ".")
			}
			return articleContent{}, apierr.ErrValidation
		}
		if !a.imageHostOK(cover) {
			return articleContent{}, apierr.ErrImageHost
		}
	}
	hasCover := cover != ""

	// Primary language.
	primaryLang := langOr(in.Lang)
	if !validLocales[primaryLang] {
		return articleContent{}, apierr.ErrValidation.WithMessage("Ngôn ngữ không hợp lệ.")
	}
	title, excerpt, coverAlt, err := validateLocalizedMeta(in.Title, in.Excerpt, in.CoverAlt, hasCover)
	if err != nil {
		return articleContent{}, err
	}
	body, err := a.prepareBody(in.Format, in.Content, in.Body)
	if err != nil {
		return articleContent{}, err
	}
	if excerpt == "" {
		excerpt = content.DeriveExcerpt(body)
	}

	// Translations: each additional language is validated the same way. They may
	// be empty ("publish now, translate later").
	var translations map[string]domain.Translation
	for _, t := range in.Translations {
		tl := langOr(t.Lang)
		if !validLocales[tl] {
			return articleContent{}, apierr.ErrValidation.WithMessage("Ngôn ngữ không hợp lệ.")
		}
		if tl == primaryLang {
			return articleContent{}, apierr.ErrValidation.WithMessage("Bản dịch trùng ngôn ngữ chính.")
		}
		if _, dup := translations[tl]; dup {
			return articleContent{}, apierr.ErrValidation.WithMessage("Trùng lặp bản dịch cho một ngôn ngữ.")
		}
		tt, te, tca, err := validateLocalizedMeta(t.Title, t.Excerpt, t.CoverAlt, hasCover)
		if err != nil {
			return articleContent{}, err
		}
		tb, err := a.prepareBody(t.Format, t.Content, t.Body)
		if err != nil {
			return articleContent{}, err
		}
		if te == "" {
			te = content.DeriveExcerpt(tb)
		}
		if translations == nil {
			translations = make(map[string]domain.Translation)
		}
		translations[tl] = domain.Translation{Title: tt, Excerpt: te, CoverAlt: tca, Body: tb}
	}

	return articleContent{
		Category:     category,
		Cover:        cover,
		Tags:         tags,
		Lang:         primaryLang,
		Title:        title,
		Excerpt:      excerpt,
		CoverAlt:     coverAlt,
		Body:         body,
		Translations: translations,
	}, nil
}

// validateLocalizedMeta trims and bounds one language's title/excerpt/cover-alt.
// Cover alt is plain accessibility copy tied to the shared cover image, so it is
// dropped when the article has no cover.
func validateLocalizedMeta(title, excerpt, coverAlt string, hasCover bool) (string, string, string, error) {
	title = strings.TrimSpace(title)
	excerpt = strings.TrimSpace(excerpt)
	coverAlt = strings.TrimSpace(coverAlt)
	switch {
	case title == "":
		return "", "", "", apierr.ErrValidation.WithMessage("Vui lòng nhập tiêu đề.")
	case len(title) > 300:
		return "", "", "", apierr.ErrValidation.WithMessage("Tiêu đề tối đa 300 ký tự.")
	case len(excerpt) > 500:
		return "", "", "", apierr.ErrValidation.WithMessage("Mô tả ngắn tối đa 500 ký tự.")
	}
	if !hasCover {
		coverAlt = ""
	}
	if len(coverAlt) > maxCoverAltLen {
		return "", "", "", apierr.ErrValidation.WithMessage("Mô tả ảnh bìa tối đa 300 ký tự.")
	}
	return title, excerpt, coverAlt, nil
}

// prepareBody converts one language's body payload (markdown or structured
// blocks) into normalized domain blocks and enforces the image-host policy.
func (a *API) prepareBody(format, mdContent string, blocks []blockInput) ([]domain.Block, error) {
	var raw []domain.Block
	switch format {
	case "markdown":
		raw = content.BlocksFromMarkdown(mdContent)
	case "blocks":
		raw = make([]domain.Block, 0, len(blocks))
		for _, b := range blocks {
			raw = append(raw, domain.Block{
				Type: b.Type, Text: b.Text, Lang: b.Lang, Code: b.Code,
				Caption: b.Caption, Steps: b.Steps, Items: b.Items, Ordered: b.Ordered,
				Src: b.Src, Alt: b.Alt,
			})
		}
	default:
		return nil, apierr.ErrValidation.WithMessage("Định dạng nội dung phải là markdown hoặc blocks.")
	}
	body, err := content.NormalizeBlocks(raw)
	var invalid *content.ErrInvalid
	if errors.As(err, &invalid) {
		return nil, apierr.ErrValidation.WithMessage(invalid.Reason + ".")
	}
	if err != nil {
		return nil, apierr.ErrInternal
	}
	if err := a.checkImageHosts(body); err != nil {
		return nil, err
	}
	return body, nil
}

// buildTranslations projects an article's translations into the detail DTO,
// truncating each variant's body when the part is locked (same rule as the
// primary body). It returns the translation map (nil when single-language) and
// the ordered list of available languages (primary first).
func buildTranslations(a domain.Article, locked bool) (map[string]translationDTO, []string) {
	primary := langOr(a.Lang)
	langs := []string{primary}
	if len(a.Translations) == 0 {
		return nil, langs
	}
	out := make(map[string]translationDTO, len(a.Translations))
	for _, loc := range []string{"vi", "en"} {
		if loc == primary {
			continue
		}
		t, ok := a.Translations[loc]
		if !ok {
			continue
		}
		body := t.Body
		if locked && len(body) > 2 {
			body = body[:2]
		}
		out[loc] = translationDTO{Title: t.Title, Excerpt: t.Excerpt, CoverAlt: t.CoverAlt, Body: body}
		langs = append(langs, loc)
	}
	if len(out) == 0 {
		return nil, langs
	}
	return out, langs
}

// updateArticle edits an existing article. Requires a session AND the IAM
// "articles:create" permission AND that the requester is the article's author.
// Authorship is matched by the stable author id (AuthorID == session sub) via
// editableBy, never the display name. The slug, author, publish time, ordering
// and series placement are immutable; only the title, excerpt, category, tags
// and body change.
func (a *API) updateArticle(w http.ResponseWriter, r *http.Request) {
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
		logger.From(r.Context()).Warn("article edit denied", "sub", u.Sub)
		writeError(w, r, apierr.ErrArticleForbidden)
		return
	}

	slug := r.PathValue("slug")
	existing, err := a.Store.Articles().GetBySlug(r.Context(), slug)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, r, apierr.ErrArticleNotFound)
		return
	}
	if err != nil {
		writeError(w, r, apierr.ErrArticleLoad)
		return
	}
	// Ownership: only the author may edit their own article, keyed to the stable
	// author id (seed/imported content has an empty AuthorID and is uneditable).
	if !editableBy(r.Context(), existing) {
		logger.From(r.Context()).Warn("article edit forbidden: not author", "slug", slug, "sub", u.Sub)
		writeError(w, r, apierr.ErrArticleEditForbidden)
		return
	}

	var in createArticleInput
	if !decodeJSON(w, r, &in) {
		return
	}
	c, err := a.prepareArticle(in)
	if err != nil {
		writeError(w, r, err)
		return
	}

	existing.Lang = c.Lang
	existing.Title = c.Title
	existing.Excerpt = c.Excerpt
	existing.Category = c.Category
	existing.Cover = c.Cover
	existing.CoverAlt = c.CoverAlt
	existing.Tags = c.Tags
	existing.Body = c.Body
	existing.ReadTime = content.ReadTime(c.Body)
	existing.Translations = c.Translations

	updated, err := a.Store.Articles().Update(r.Context(), existing)
	if err != nil {
		writeError(w, r, apierr.ErrArticleUpdate)
		return
	}

	logger.From(r.Context()).Info("article updated", "slug", updated.Slug, "sub", u.Sub)
	detail := articleDetail{
		articleSummary: toSummary(updated),
		Body:           updated.Body,
		Editable:       editableBy(r.Context(), updated),
	}
	detail.Translations, detail.AvailableLangs = buildTranslations(updated, false)
	writeJSON(w, r, http.StatusOK, detail)
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
		Editable:       editableBy(r.Context(), art),
		Locked:         locked,
	}
	detail.Translations, detail.AvailableLangs = buildTranslations(art, locked)

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
