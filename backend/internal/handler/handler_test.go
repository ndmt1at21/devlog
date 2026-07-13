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

const v1 = "/api/v1"

func newServer(t *testing.T) *httptest.Server {
	t.Helper()
	api := &handler.API{Store: memory.New(), Cfg: config.Config{DBDriver: "memory"}}
	return httptest.NewServer(api.NewRouter())
}

// envelope mirrors the uniform API response {code, message, traceId, data}.
type envelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	TraceID string          `json:"traceId"`
	Data    json.RawMessage `json:"data"`
}

// getEnv GETs url, decodes the envelope, and (when dst != nil) unmarshals the
// data field into dst. Returns the HTTP status and the envelope.
func getEnv(t *testing.T, url string, dst any) (int, envelope) {
	t.Helper()
	res, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer res.Body.Close()
	var env envelope
	if err := json.NewDecoder(res.Body).Decode(&env); err != nil {
		t.Fatalf("decode envelope %s: %v", url, err)
	}
	if env.TraceID == "" {
		t.Errorf("%s: envelope missing traceId", url)
	}
	if dst != nil && len(env.Data) > 0 {
		if err := json.Unmarshal(env.Data, dst); err != nil {
			t.Fatalf("unmarshal data %s: %v", url, err)
		}
	}
	return res.StatusCode, env
}

func TestListAndFeatured(t *testing.T) {
	srv := newServer(t)
	defer srv.Close()

	var list []map[string]any
	status, env := getEnv(t, srv.URL+v1+"/articles", &list)
	if status != 200 || env.Code != 0 {
		t.Fatalf("articles status=%d code=%d", status, env.Code)
	}
	if len(list) != 10 {
		t.Fatalf("articles len = %d, want 10", len(list))
	}

	// Featured returns every flagged article ordered by ord.
	var featured []map[string]any
	getEnv(t, srv.URL+v1+"/articles/featured", &featured)
	if len(featured) != 2 {
		t.Fatalf("featured len = %d, want 2", len(featured))
	}
	if featured[0]["slug"] != "ai-agents" || featured[1]["slug"] != "react-perf" {
		t.Fatalf("featured slugs = %v, want [ai-agents react-perf]", featured)
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
	getEnv(t, srv.URL+v1+"/articles/iam-2", &iam2)
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
	getEnv(t, srv.URL+v1+"/articles/ai-agents", &ai)
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

	// Empty name → 400 with the validation code (1001).
	res, err := http.Post(srv.URL+v1+"/articles/ai-agents/comments", "application/json",
		strings.NewReader(`{"name":"","text":"hi"}`))
	if err != nil {
		t.Fatal(err)
	}
	var errEnv envelope
	json.NewDecoder(res.Body).Decode(&errEnv)
	res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("empty-name status = %d, want 400", res.StatusCode)
	}
	if errEnv.Code != 1001 {
		t.Fatalf("empty-name code = %d, want 1001 (validation)", errEnv.Code)
	}

	// Valid → 201 with code 0 and the created comment in data.
	res, err = http.Post(srv.URL+v1+"/articles/ai-agents/comments", "application/json",
		strings.NewReader(`{"name":"Tester","text":"Hay quá"}`))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("create status = %d, want 201", res.StatusCode)
	}
	var env envelope
	json.NewDecoder(res.Body).Decode(&env)
	if env.Code != 0 {
		t.Fatalf("create code = %d, want 0", env.Code)
	}
	var c map[string]any
	json.Unmarshal(env.Data, &c)
	if c["name"] != "Tester" || c["initial"] != "T" {
		t.Fatalf("unexpected comment dto: %v", c)
	}
}

func TestReactions(t *testing.T) {
	srv := newServer(t)
	defer srv.Close()

	// Anonymous status: zero likes, nothing liked/bookmarked.
	var st struct {
		Likes      int  `json:"likes"`
		Liked      bool `json:"liked"`
		Bookmarked bool `json:"bookmarked"`
	}
	status, env := getEnv(t, srv.URL+v1+"/articles/ai-agents/reactions", &st)
	if status != 200 || env.Code != 0 {
		t.Fatalf("reactions status=%d code=%d", status, env.Code)
	}
	if st.Likes != 0 || st.Liked || st.Bookmarked {
		t.Fatalf("anonymous reactions = %+v, want zero state", st)
	}

	// Toggling requires login → 401 with the unauthorized code (1002).
	req, _ := http.NewRequest(http.MethodPut, srv.URL+v1+"/articles/ai-agents/reactions/like", nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	var errEnv envelope
	json.NewDecoder(res.Body).Decode(&errEnv)
	res.Body.Close()
	if res.StatusCode != http.StatusUnauthorized || errEnv.Code != 1002 {
		t.Fatalf("anonymous like status=%d code=%d, want 401/1002", res.StatusCode, errEnv.Code)
	}

	// Bookmark listing also requires login.
	status, env = getEnv(t, srv.URL+v1+"/me/bookmarks", nil)
	if status != http.StatusUnauthorized || env.Code != 1002 {
		t.Fatalf("anonymous bookmarks status=%d code=%d, want 401/1002", status, env.Code)
	}
}
