package handler_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/ndmt1at21/devlog/backend/internal/config"
)

// coverURL is a well-formed image URL (absolute https, within the length cap).
const coverURL = "https://cdn.example.com/img/cover-a.jpg"

func coverBody(cover string) string {
	return fmt.Sprintf(`{
		"title": "Bài có ảnh bìa",
		"category": "Backend",
		"tags": ["go"],
		"cover": %q,
		"format": "markdown",
		"content": "# Tiêu đề\n\nNội dung bài viết."
	}`, cover)
}

// coverOf GETs the article and returns its cover field ("" when absent).
func coverOf(t *testing.T, client *http.Client, baseURL, slug string) string {
	t.Helper()
	var detail struct {
		Cover string `json:"cover"`
	}
	if status, _ := getClientEnv(t, client, baseURL+v1+"/articles/"+slug, &detail); status != http.StatusOK {
		t.Fatalf("GET %s status = %d", slug, status)
	}
	return detail.Cover
}

// TestArticleCoverRoundTrip drives create → get → update → get → clear. The
// update leg specifically guards both stores' Update, which previously omitted
// the cover column entirely (so an author could never set or change a cover).
func TestArticleCoverRoundTrip(t *testing.T) {
	srv, client := newAuthedClient(t, true)

	// Create with a cover; the create response echoes it and it persists.
	status, env := postJSON(t, client, srv.URL+v1+"/articles", coverBody(coverURL))
	if status != http.StatusCreated || env.Code != 0 {
		t.Fatalf("create: status=%d code=%d message=%q, want 201/0", status, env.Code, env.Message)
	}
	var created struct {
		Slug  string `json:"slug"`
		Cover string `json:"cover"`
	}
	if err := json.Unmarshal(env.Data, &created); err != nil {
		t.Fatal(err)
	}
	if created.Cover != coverURL {
		t.Errorf("create response cover = %q, want %q", created.Cover, coverURL)
	}
	if got := coverOf(t, client, srv.URL, created.Slug); got != coverURL {
		t.Errorf("persisted cover = %q, want %q", got, coverURL)
	}

	// Change the cover — the previously-dropped update path.
	const newCover = "https://cdn.example.com/img/cover-b.png"
	status, env = putJSON(t, client, srv.URL+v1+"/articles/"+created.Slug, coverBody(newCover))
	if status != http.StatusOK || env.Code != 0 {
		t.Fatalf("update: status=%d code=%d message=%q, want 200/0", status, env.Code, env.Message)
	}
	if got := coverOf(t, client, srv.URL, created.Slug); got != newCover {
		t.Errorf("cover after edit = %q, want %q (Update must persist cover)", got, newCover)
	}

	// Editing with no cover clears it.
	if status, _ := putJSON(t, client, srv.URL+v1+"/articles/"+created.Slug, coverBody("")); status != http.StatusOK {
		t.Fatalf("clear-cover update status = %d, want 200", status)
	}
	if got := coverOf(t, client, srv.URL, created.Slug); got != "" {
		t.Errorf("cover after clear = %q, want empty", got)
	}
}

// TestArticleCoverHostEnforced verifies a cover must live under the configured
// public image origin — the same guard body images get — so covers can only come
// from the upload flow.
func TestArticleCoverHostEnforced(t *testing.T) {
	cfg := config.Config{DBDriver: "memory", ImageBaseURL: "https://cdn.example.com"}
	srv, client := newAuthedClientCfg(t, true, cfg)

	// A cover on a foreign origin is rejected with the image-host code (3011).
	status, env := postJSON(t, client, srv.URL+v1+"/articles", coverBody("https://evil.example.net/x.jpg"))
	if status != http.StatusBadRequest || env.Code != 3011 {
		t.Fatalf("foreign cover: status=%d code=%d, want 400/3011", status, env.Code)
	}

	// A cover under the configured origin is accepted.
	status, env = postJSON(t, client, srv.URL+v1+"/articles", coverBody("https://cdn.example.com/img/ok.jpg"))
	if status != http.StatusCreated || env.Code != 0 {
		t.Fatalf("valid cover: status=%d code=%d message=%q, want 201/0", status, env.Code, env.Message)
	}
}
