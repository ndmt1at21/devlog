package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ndmt1at21/devlog/backend/internal/config"
	"github.com/ndmt1at21/devlog/backend/internal/handler"
	"github.com/ndmt1at21/devlog/backend/internal/repository/memory"
)

func newServer(t *testing.T) *httptest.Server {
	t.Helper()
	api := &handler.API{Store: memory.New(), Cfg: config.Config{DBDriver: "memory"}}
	return httptest.NewServer(api.NewRouter())
}

func getJSON(t *testing.T, url string, dst any) int {
	t.Helper()
	res, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer res.Body.Close()
	if dst != nil {
		if err := json.NewDecoder(res.Body).Decode(dst); err != nil {
			t.Fatalf("decode %s: %v", url, err)
		}
	}
	return res.StatusCode
}

func TestListAndFeatured(t *testing.T) {
	srv := newServer(t)
	defer srv.Close()

	var list []map[string]any
	if code := getJSON(t, srv.URL+"/api/articles", &list); code != 200 {
		t.Fatalf("articles status = %d", code)
	}
	if len(list) != 10 {
		t.Fatalf("articles len = %d, want 10", len(list))
	}

	var featured map[string]any
	getJSON(t, srv.URL+"/api/articles/featured", &featured)
	if featured["slug"] != "ai-agents" {
		t.Fatalf("featured slug = %v, want ai-agents", featured["slug"])
	}
}

func TestPaywallLocksLaterSeriesParts(t *testing.T) {
	srv := newServer(t)
	defer srv.Close()

	// Part 2 of a series is locked for anonymous (non-premium) readers.
	var iam2 struct {
		Locked bool             `json:"locked"`
		Body   []map[string]any `json:"body"`
	}
	getJSON(t, srv.URL+"/api/articles/iam-2", &iam2)
	if !iam2.Locked {
		t.Fatal("iam-2 should be locked for anonymous readers")
	}
	if len(iam2.Body) != 2 {
		t.Fatalf("locked body len = %d, want 2", len(iam2.Body))
	}

	// A standalone article is never locked and returns its full body.
	var ai struct {
		Locked bool             `json:"locked"`
		Body   []map[string]any `json:"body"`
	}
	getJSON(t, srv.URL+"/api/articles/ai-agents", &ai)
	if ai.Locked {
		t.Fatal("ai-agents should not be locked")
	}
	if len(ai.Body) <= 2 {
		t.Fatalf("ai-agents body len = %d, want > 2", len(ai.Body))
	}
}

func TestCommentValidationAndCreate(t *testing.T) {
	srv := newServer(t)
	defer srv.Close()

	// Empty name → 400.
	res, err := http.Post(srv.URL+"/api/articles/ai-agents/comments", "application/json",
		strings.NewReader(`{"name":"","text":"hi"}`))
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("empty-name status = %d, want 400", res.StatusCode)
	}

	// Valid → 201.
	res, err = http.Post(srv.URL+"/api/articles/ai-agents/comments", "application/json",
		strings.NewReader(`{"name":"Tester","text":"Hay quá"}`))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("create status = %d, want 201", res.StatusCode)
	}
	var c map[string]any
	json.NewDecoder(res.Body).Decode(&c)
	if c["name"] != "Tester" || c["initial"] != "T" {
		t.Fatalf("unexpected comment dto: %v", c)
	}
}
