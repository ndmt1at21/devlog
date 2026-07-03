// Package content turns author input into the domain's block model: a scoped
// README-style markdown converter, a strict block validator, a Vietnamese-aware
// slugifier, and read-time/excerpt derivation. Inline markdown spans (**b**,
// *i*, `c`, [t](url)) are left inside text fields; the frontend renders them
// safely as React elements, so no HTML is ever produced or stored here.
package content

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/ndmt1at21/devlog/backend/internal/domain"
)

// Limits applied to a submitted article body. The HTTP layer additionally caps
// the whole request at 1 MiB.
const (
	MaxBlocks    = 200
	MaxTextLen   = 20000
	MaxLangLen   = 40
	MaxCaption   = 300
	MaxListItems = 50
	MaxItemLen   = 500
	MaxURLLen    = 2048
)

// ErrInvalid wraps a human-readable validation failure for the block body.
type ErrInvalid struct{ Reason string }

func (e *ErrInvalid) Error() string { return e.Reason }

func invalidf(format string, args ...any) error {
	return &ErrInvalid{Reason: fmt.Sprintf(format, args...)}
}

// ---- markdown → blocks ----

var (
	headingRe = regexp.MustCompile(`^(#{1,6})\s+(.*)$`)
	orderedRe = regexp.MustCompile(`^\d{1,4}[.)]\s+(.*)$`)
	hruleRe   = regexp.MustCompile(`^(-{3,}|\*{3,}|_{3,})$`)
	// imageRe matches a line that is exactly one ![alt](url) image. Images mixed
	// into paragraph text stay literal — only standalone lines become blocks.
	imageRe = regexp.MustCompile(`^!\[([^\]]*)\]\(([^()\s]+)\)$`)
)

// BlocksFromMarkdown converts README-style markdown into body blocks:
// headings → "h", fenced code (with optional language) → "code", "> " quotes →
// "quote", -/*/+ and "1." lists → "list", standalone ![alt](url) lines →
// "img", horizontal rules are dropped, and consecutive text lines merge into
// "p" paragraphs. Inline spans are preserved verbatim in the text. The output
// still goes through NormalizeBlocks.
func BlocksFromMarkdown(src string) []domain.Block {
	lines := strings.Split(strings.ReplaceAll(src, "\r\n", "\n"), "\n")
	var blocks []domain.Block
	var para []string

	flush := func() {
		if len(para) > 0 {
			blocks = append(blocks, domain.Block{Type: "p", Text: strings.Join(para, " ")})
			para = nil
		}
	}

	for i := 0; i < len(lines); {
		trimmed := strings.TrimSpace(lines[i])
		switch {
		case trimmed == "" || hruleRe.MatchString(trimmed):
			flush()
			i++

		case strings.HasPrefix(trimmed, "```"):
			flush()
			lang := strings.TrimSpace(strings.TrimPrefix(trimmed, "```"))
			i++
			var code []string
			for i < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[i]), "```") {
				code = append(code, lines[i])
				i++
			}
			if i < len(lines) {
				i++ // closing fence
			}
			blocks = append(blocks, domain.Block{Type: "code", Lang: lang, Code: strings.TrimRight(strings.Join(code, "\n"), "\n")})

		case headingRe.MatchString(trimmed):
			flush()
			m := headingRe.FindStringSubmatch(trimmed)
			blocks = append(blocks, domain.Block{Type: "h", Text: strings.TrimSpace(m[2])})
			i++

		case imageRe.MatchString(trimmed):
			flush()
			m := imageRe.FindStringSubmatch(trimmed)
			blocks = append(blocks, domain.Block{Type: "img", Alt: strings.TrimSpace(m[1]), Src: m[2]})
			i++

		case strings.HasPrefix(trimmed, ">"):
			flush()
			var quote []string
			for i < len(lines) {
				t := strings.TrimSpace(lines[i])
				if !strings.HasPrefix(t, ">") {
					break
				}
				if line := strings.TrimSpace(strings.TrimPrefix(t, ">")); line != "" {
					quote = append(quote, line)
				}
				i++
			}
			blocks = append(blocks, domain.Block{Type: "quote", Text: strings.Join(quote, " ")})

		case unorderedItem(trimmed) != "":
			flush()
			var items []string
			for i < len(lines) {
				item := unorderedItem(strings.TrimSpace(lines[i]))
				if item == "" {
					break
				}
				items = append(items, item)
				i++
			}
			blocks = append(blocks, domain.Block{Type: "list", Items: items})

		case orderedRe.MatchString(trimmed):
			flush()
			var items []string
			for i < len(lines) {
				m := orderedRe.FindStringSubmatch(strings.TrimSpace(lines[i]))
				if m == nil {
					break
				}
				items = append(items, strings.TrimSpace(m[1]))
				i++
			}
			blocks = append(blocks, domain.Block{Type: "list", Items: items, Ordered: true})

		default:
			para = append(para, trimmed)
			i++
		}
	}
	flush()
	return blocks
}

// unorderedItem returns the item text when the line is a "-", "*" or "+" list
// item, or "" otherwise.
func unorderedItem(line string) string {
	for _, marker := range []string{"- ", "* ", "+ "} {
		if strings.HasPrefix(line, marker) {
			return strings.TrimSpace(line[len(marker):])
		}
	}
	return ""
}

// ---- block validation ----

var langRe = regexp.MustCompile(`^[a-zA-Z0-9+#._-]*$`)

// NormalizeBlocks validates author-submitted blocks against the allowed types
// and size limits, trims whitespace, and drops empty blocks. It returns an
// *ErrInvalid describing the first violation.
func NormalizeBlocks(in []domain.Block) ([]domain.Block, error) {
	if len(in) > MaxBlocks {
		return nil, invalidf("bài viết vượt quá %d khối nội dung", MaxBlocks)
	}
	out := make([]domain.Block, 0, len(in))
	for i, b := range in {
		b.Text = strings.TrimSpace(b.Text)
		b.Code = strings.TrimRight(b.Code, "\n\t ")
		b.Lang = strings.TrimSpace(b.Lang)
		b.Caption = strings.TrimSpace(b.Caption)
		if len(b.Text) > MaxTextLen || len(b.Code) > MaxTextLen {
			return nil, invalidf("khối %d quá dài (tối đa %d ký tự)", i+1, MaxTextLen)
		}
		switch b.Type {
		case "p", "h", "quote":
			if b.Text == "" {
				continue
			}
			out = append(out, domain.Block{Type: b.Type, Text: b.Text})
		case "code":
			if b.Code == "" {
				continue
			}
			if len(b.Lang) > MaxLangLen || !langRe.MatchString(b.Lang) {
				return nil, invalidf("khối %d: ngôn ngữ code không hợp lệ", i+1)
			}
			out = append(out, domain.Block{Type: "code", Lang: b.Lang, Code: b.Code})
		case "diagram":
			steps, err := normalizeItems(b.Steps, i)
			if err != nil {
				return nil, err
			}
			if len(steps) == 0 {
				continue
			}
			if len(b.Caption) > MaxCaption {
				return nil, invalidf("khối %d: chú thích quá dài", i+1)
			}
			out = append(out, domain.Block{Type: "diagram", Steps: steps, Caption: b.Caption})
		case "list":
			items, err := normalizeItems(b.Items, i)
			if err != nil {
				return nil, err
			}
			if len(items) == 0 {
				continue
			}
			out = append(out, domain.Block{Type: "list", Items: items, Ordered: b.Ordered})
		case "img":
			b.Src = strings.TrimSpace(b.Src)
			b.Alt = strings.TrimSpace(b.Alt)
			if b.Src == "" {
				continue
			}
			if err := validateImageURL(b.Src); err != nil {
				return nil, invalidf("khối %d: %s", i+1, err)
			}
			if len(b.Alt) > MaxCaption || len(b.Caption) > MaxCaption {
				return nil, invalidf("khối %d: mô tả ảnh quá dài (tối đa %d ký tự)", i+1, MaxCaption)
			}
			out = append(out, domain.Block{Type: "img", Src: b.Src, Alt: b.Alt, Caption: b.Caption})
		default:
			return nil, invalidf("khối %d: loại %q không được hỗ trợ", i+1, b.Type)
		}
	}
	if len(out) == 0 {
		return nil, invalidf("nội dung bài viết đang trống")
	}
	return out, nil
}

// validateImageURL enforces the shape of an img block's Src: an absolute https
// URL (plain http only for loopback hosts, so MinIO works in dev) within the
// length cap. Which *origin* is allowed is deployment policy, checked in the
// handler against the configured image base URL.
func validateImageURL(src string) error {
	if len(src) > MaxURLLen {
		return fmt.Errorf("đường dẫn ảnh quá dài (tối đa %d ký tự)", MaxURLLen)
	}
	u, err := url.Parse(src)
	if err != nil || u.Host == "" {
		return fmt.Errorf("đường dẫn ảnh không hợp lệ")
	}
	local := u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1" || u.Hostname() == "::1"
	if u.Scheme != "https" && !(u.Scheme == "http" && local) {
		return fmt.Errorf("đường dẫn ảnh phải dùng https")
	}
	return nil
}

func normalizeItems(in []string, blockIdx int) ([]string, error) {
	if len(in) > MaxListItems {
		return nil, invalidf("khối %d vượt quá %d mục", blockIdx+1, MaxListItems)
	}
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if len(s) > MaxItemLen {
			return nil, invalidf("khối %d: mục quá dài (tối đa %d ký tự)", blockIdx+1, MaxItemLen)
		}
		out = append(out, s)
	}
	return out, nil
}

// ---- slug ----

const maxSlugLen = 80

// viFold maps lowercase Vietnamese letters to their ASCII base letter.
var viFold = func() map[rune]rune {
	m := make(map[rune]rune)
	for base, letters := range map[rune]string{
		'a': "àáảãạăằắẳẵặâầấẩẫậ",
		'e': "èéẻẽẹêềếểễệ",
		'i': "ìíỉĩị",
		'o': "òóỏõọôồốổỗộơờớởỡợ",
		'u': "ùúủũụưừứửữự",
		'y': "ỳýỷỹỵ",
		'd': "đ",
	} {
		for _, r := range letters {
			m[r] = base
		}
	}
	return m
}()

// Slugify derives a URL slug from a (typically Vietnamese) title: diacritics
// fold to ASCII, everything outside [a-z0-9] collapses to single dashes, and
// the result is capped at 80 chars (cut on a dash boundary when possible).
func Slugify(title string) string {
	var b strings.Builder
	prevDash := true
	for _, r := range strings.ToLower(strings.TrimSpace(title)) {
		if f, ok := viFold[r]; ok {
			r = f
		}
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			prevDash = false
		} else if !prevDash {
			b.WriteByte('-')
			prevDash = true
		}
	}
	slug := strings.TrimRight(b.String(), "-")
	if len(slug) > maxSlugLen {
		slug = slug[:maxSlugLen]
		if i := strings.LastIndexByte(slug, '-'); i > maxSlugLen/2 {
			slug = slug[:i]
		}
	}
	if slug == "" {
		return "bai-viet"
	}
	return slug
}

// ---- derived metadata ----

// inlineMarkerRe strips inline markdown markers so excerpts and word counts see
// prose, not syntax.
var inlineMarkerRe = regexp.MustCompile("[*_`]+|\\[([^\\]]*)\\]\\([^)]*\\)")

func plainText(s string) string {
	return inlineMarkerRe.ReplaceAllString(s, "$1")
}

// ReadTime estimates a Vietnamese read-time label ("N phút đọc") at ~200 wpm
// across all text-bearing fields.
func ReadTime(blocks []domain.Block) string {
	words := 0
	for _, b := range blocks {
		words += len(strings.Fields(b.Text)) + len(strings.Fields(b.Code))
		for _, s := range b.Steps {
			words += len(strings.Fields(s))
		}
		for _, s := range b.Items {
			words += len(strings.Fields(s))
		}
	}
	minutes := max((words+199)/200, 1)
	return fmt.Sprintf("%d phút đọc", minutes)
}

// DeriveExcerpt returns the first paragraph's prose truncated to ~160 runes on
// a word boundary — used when the author leaves the excerpt empty.
func DeriveExcerpt(blocks []domain.Block) string {
	for _, b := range blocks {
		if b.Type != "p" {
			continue
		}
		text := plainText(b.Text)
		r := []rune(text)
		if len(r) <= 160 {
			return text
		}
		cut := string(r[:160])
		if i := strings.LastIndexByte(cut, ' '); i > 80 {
			cut = cut[:i]
		}
		return cut + "…"
	}
	return ""
}
