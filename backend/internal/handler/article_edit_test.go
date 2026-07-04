package handler_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// putJSON mirrors postJSON but issues a PUT through the cookie-carrying client.
func putJSON(t *testing.T, client *http.Client, url, body string) (int, envelope) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPut, url, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var env envelope
	if err := json.NewDecoder(res.Body).Decode(&env); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	return res.StatusCode, env
}

const editArticleBody = `{
	"title": "Tối ưu Go nâng cao",
	"category": "Backend",
	"tags": ["go", "perf"],
	"format": "markdown",
	"content": "# Phần mới\n\nNội dung đã được cập nhật."
}`

// createSampleArticle publishes markdownArticle through the authed client and
// returns its slug.
func createSampleArticle(t *testing.T, baseURL string, client *http.Client) string {
	t.Helper()
	status, env := postJSON(t, client, baseURL+v1+"/articles", markdownArticle)
	if status != http.StatusCreated {
		t.Fatalf("seed create status = %d", status)
	}
	var created struct {
		Slug string `json:"slug"`
	}
	if err := json.Unmarshal(env.Data, &created); err != nil {
		t.Fatal(err)
	}
	return created.Slug
}

func TestUpdateArticleAuthorFlow(t *testing.T) {
	srv, client := newAuthedClient(t, true)
	slug := createSampleArticle(t, srv.URL, client)

	status, env := putJSON(t, client, srv.URL+v1+"/articles/"+slug, editArticleBody)
	if status != http.StatusOK || env.Code != 0 {
		t.Fatalf("edit: status=%d code=%d message=%q, want 200/0", status, env.Code, env.Message)
	}
	var updated struct {
		Slug     string `json:"slug"`
		Title    string `json:"title"`
		Editable bool   `json:"editable"`
		Body     []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"body"`
	}
	if err := json.Unmarshal(env.Data, &updated); err != nil {
		t.Fatal(err)
	}
	// The slug is immutable so existing links / comments stay valid.
	if updated.Slug != slug {
		t.Errorf("slug = %q, want unchanged %q", updated.Slug, slug)
	}
	if updated.Title != "Tối ưu Go nâng cao" {
		t.Errorf("title = %q, want edited title", updated.Title)
	}
	if !updated.Editable {
		t.Error("editable should be true for the author in the edit response")
	}
	if len(updated.Body) != 2 || updated.Body[0].Type != "h" || updated.Body[1].Type != "p" {
		t.Fatalf("body = %+v, want [h p]", updated.Body)
	}

	// The edit is persisted and visible at the same slug; the author sees the
	// edit affordance flag.
	var detail map[string]any
	getClientEnv(t, client, srv.URL+v1+"/articles/"+slug, &detail)
	if detail["title"] != "Tối ưu Go nâng cao" {
		t.Errorf("persisted title = %v", detail["title"])
	}
	if detail["editable"] != true {
		t.Errorf("author GET editable = %v, want true", detail["editable"])
	}

	// An anonymous reader is never the author → not editable.
	var anon map[string]any
	getEnv(t, srv.URL+v1+"/articles/"+slug, &anon)
	if anon["editable"] != false {
		t.Errorf("anonymous GET editable = %v, want false", anon["editable"])
	}
}

func TestUpdateArticleRequiresLogin(t *testing.T) {
	srv, client := newAuthedClient(t, true)
	slug := createSampleArticle(t, srv.URL, client)

	// Cookieless client → 401.
	status, env := putJSON(t, &http.Client{}, srv.URL+v1+"/articles/"+slug, editArticleBody)
	if status != http.StatusUnauthorized || env.Code != 1002 {
		t.Fatalf("anonymous edit: status=%d code=%d, want 401/1002", status, env.Code)
	}
}

func TestUpdateArticleForbiddenWithoutPermission(t *testing.T) {
	srv, client := newAuthedClient(t, false)
	// docker-101 is a seed article; without the permission the edit is rejected
	// before ownership is even considered.
	status, env := putJSON(t, client, srv.URL+v1+"/articles/docker-101", editArticleBody)
	if status != http.StatusForbidden || env.Code != 3005 {
		t.Fatalf("edit without permission: status=%d code=%d, want 403/3005", status, env.Code)
	}
}

func TestUpdateArticleForbiddenWhenNotAuthor(t *testing.T) {
	srv, client := newAuthedClient(t, true)
	// docker-101 is seed content with no owning account (author_id is NULL), so
	// even with the permission no signed-in user matches → forbidden.
	status, env := putJSON(t, client, srv.URL+v1+"/articles/docker-101", editArticleBody)
	if status != http.StatusForbidden || env.Code != 3013 {
		t.Fatalf("edit foreign article: status=%d code=%d, want 403/3013", status, env.Code)
	}
}

func TestUpdateArticleNotFound(t *testing.T) {
	srv, client := newAuthedClient(t, true)
	status, env := putJSON(t, client, srv.URL+v1+"/articles/does-not-exist", editArticleBody)
	if status != http.StatusNotFound || env.Code != 3000 {
		t.Fatalf("edit missing article: status=%d code=%d, want 404/3000", status, env.Code)
	}
}

func TestUpdateArticleValidation(t *testing.T) {
	srv, client := newAuthedClient(t, true)
	slug := createSampleArticle(t, srv.URL, client)

	cases := map[string]string{
		"empty title":    `{"title":"","category":"Backend","format":"markdown","content":"x"}`,
		"empty category": `{"title":"T","category":"","format":"markdown","content":"x"}`,
		"bad format":     `{"title":"T","category":"Backend","format":"html","content":"x"}`,
		"empty body":     `{"title":"T","category":"Backend","format":"markdown","content":"   "}`,
	}
	for name, body := range cases {
		status, env := putJSON(t, client, srv.URL+v1+"/articles/"+slug, body)
		if status != http.StatusBadRequest {
			t.Errorf("%s: status=%d code=%d, want 400", name, status, env.Code)
		}
	}
}
