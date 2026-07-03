package content

import (
	"strings"
	"testing"

	"github.com/ndmt1at21/devlog/backend/internal/domain"
)

func TestBlocksFromMarkdown(t *testing.T) {
	src := "# Tiêu đề chính\n" +
		"\n" +
		"Đoạn mở đầu dòng một\ntiếp nối dòng hai.\n" +
		"\n" +
		"> Trích dẫn hay\n> trên hai dòng.\n" +
		"\n" +
		"```go\nfmt.Println(\"hi\")\n```\n" +
		"\n" +
		"- một\n- hai\n" +
		"\n" +
		"1. bước đầu\n2. bước sau\n" +
		"\n" +
		"---\n" +
		"Kết thúc với **đậm** và [link](https://example.com).\n"

	got := BlocksFromMarkdown(src)
	want := []domain.Block{
		{Type: "h", Text: "Tiêu đề chính"},
		{Type: "p", Text: "Đoạn mở đầu dòng một tiếp nối dòng hai."},
		{Type: "quote", Text: "Trích dẫn hay trên hai dòng."},
		{Type: "code", Lang: "go", Code: `fmt.Println("hi")`},
		{Type: "list", Items: []string{"một", "hai"}},
		{Type: "list", Items: []string{"bước đầu", "bước sau"}, Ordered: true},
		{Type: "p", Text: "Kết thúc với **đậm** và [link](https://example.com)."},
	}
	if len(got) != len(want) {
		t.Fatalf("blocks = %d, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		g, w := got[i], want[i]
		if g.Type != w.Type || g.Text != w.Text || g.Lang != w.Lang || g.Code != w.Code || g.Ordered != w.Ordered {
			t.Errorf("block %d = %+v, want %+v", i, g, w)
		}
		if strings.Join(g.Items, "|") != strings.Join(w.Items, "|") {
			t.Errorf("block %d items = %v, want %v", i, g.Items, w.Items)
		}
	}
}

func TestBlocksFromMarkdownUnclosedFence(t *testing.T) {
	got := BlocksFromMarkdown("```js\nlet x = 1\n")
	if len(got) != 1 || got[0].Type != "code" || got[0].Code != "let x = 1" {
		t.Fatalf("unclosed fence = %+v", got)
	}
}

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Hướng dẫn tối ưu Go":      "huong-dan-toi-uu-go",
		"Đây là ĐỀ thi thử":        "day-la-de-thi-thu",
		"  Hello --- World!!  ":    "hello-world",
		"Tiếng Việt: ắ ổ ữ đ 100%": "tieng-viet-a-o-u-d-100",
		"":                         "bai-viet",
		"!!!":                      "bai-viet",
	}
	for in, want := range cases {
		if got := Slugify(in); got != want {
			t.Errorf("Slugify(%q) = %q, want %q", in, got, want)
		}
	}
	long := Slugify(strings.Repeat("xin chao ", 30))
	if len(long) > 80 {
		t.Errorf("long slug len = %d, want <= 80", len(long))
	}
}

func TestNormalizeBlocks(t *testing.T) {
	got, err := NormalizeBlocks([]domain.Block{
		{Type: "p", Text: "  giữ lại  "},
		{Type: "p", Text: "   "}, // dropped
		{Type: "code", Lang: "go", Code: "x := 1\n"},
		{Type: "list", Items: []string{" a ", "", "b"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("blocks = %d, want 3: %+v", len(got), got)
	}
	if got[0].Text != "giữ lại" || got[1].Code != "x := 1" || len(got[2].Items) != 2 {
		t.Fatalf("normalize = %+v", got)
	}

	if _, err := NormalizeBlocks([]domain.Block{{Type: "script", Text: "x"}}); err == nil {
		t.Error("unknown block type should be rejected")
	}
	if _, err := NormalizeBlocks([]domain.Block{{Type: "code", Code: "x", Lang: "js<script>"}}); err == nil {
		t.Error("invalid code lang should be rejected")
	}
	if _, err := NormalizeBlocks(nil); err == nil {
		t.Error("empty body should be rejected")
	}
}

func TestReadTimeAndExcerpt(t *testing.T) {
	long := strings.Repeat("từ ", 450) // ~450 words → 3 phút
	blocks := []domain.Block{{Type: "p", Text: long}}
	if got := ReadTime(blocks); got != "3 phút đọc" {
		t.Errorf("ReadTime = %q, want %q", got, "3 phút đọc")
	}
	if got := ReadTime([]domain.Block{{Type: "p", Text: "ngắn"}}); got != "1 phút đọc" {
		t.Errorf("ReadTime short = %q", got)
	}

	ex := DeriveExcerpt([]domain.Block{
		{Type: "h", Text: "Tiêu đề"},
		{Type: "p", Text: "Một **đoạn** với `code` và [liên kết](https://x.vn) sạch."},
	})
	if ex != "Một đoạn với code và liên kết sạch." {
		t.Errorf("DeriveExcerpt = %q", ex)
	}
	if got := DeriveExcerpt([]domain.Block{{Type: "p", Text: long}}); !strings.HasSuffix(got, "…") {
		t.Errorf("long excerpt should be truncated with ellipsis, got %q", got)
	}
}
