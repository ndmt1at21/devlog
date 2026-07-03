package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ndmt1at21/devlog/backend/internal/authn"
	"github.com/ndmt1at21/devlog/backend/internal/config"
	"github.com/ndmt1at21/devlog/backend/internal/handler"
	"github.com/ndmt1at21/devlog/backend/internal/repository/memory"
	"github.com/ndmt1at21/devlog/backend/internal/session"
)

// fakeAuth is an authn.Provider stub: every login succeeds and the IAM
// permission decision is fixed to `allow`.
type fakeAuth struct{ allow bool }

func (f *fakeAuth) Login(context.Context, string, string) (*authn.TokenSet, error) {
	return &authn.TokenSet{AccessToken: "at", RefreshToken: "rt", ExpiresIn: 3600}, nil
}
func (f *fakeAuth) Register(context.Context, string, string) error { return nil }
func (f *fakeAuth) ForgotPassword(context.Context, string) error   { return nil }
func (f *fakeAuth) Logout(context.Context, string) error           { return nil }
func (f *fakeAuth) Refresh(context.Context, string) (*authn.TokenSet, error) {
	return &authn.TokenSet{AccessToken: "at2", ExpiresIn: 3600}, nil
}
func (f *fakeAuth) UserInfo(context.Context, string) (*authn.User, error) {
	return &authn.User{Sub: "u1", Name: "Tác Giả", Email: "author@example.com"}, nil
}
func (f *fakeAuth) FederatedLoginURL(string, string, string) string { return "" }
func (f *fakeAuth) ExchangeCode(context.Context, string, string) (*authn.TokenSet, error) {
	return nil, authn.ErrInvalidCredentials
}
func (f *fakeAuth) CheckPermissions(context.Context, string, []string) (bool, error) {
	return f.allow, nil
}

// newAuthedClient boots a server with the fake IAM provider and returns a
// cookie-carrying client that has already logged in.
func newAuthedClient(t *testing.T, allow bool) (*httptest.Server, *http.Client) {
	t.Helper()
	api := &handler.API{
		Store:    memory.New(),
		Cfg:      config.Config{DBDriver: "memory"},
		Auth:     &fakeAuth{allow: allow},
		Sessions: session.New("test-secret", false),
	}
	srv := httptest.NewServer(api.NewRouter())
	t.Cleanup(srv.Close)

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Jar: jar}
	res, err := client.Post(srv.URL+v1+"/auth/login", "application/json",
		strings.NewReader(`{"email":"author@example.com","password":"secret"}`))
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("login status = %d, want 200", res.StatusCode)
	}
	return srv, client
}

func postJSON(t *testing.T, client *http.Client, url, body string) (int, envelope) {
	t.Helper()
	res, err := client.Post(url, "application/json", strings.NewReader(body))
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

const markdownArticle = `{
	"title": "Hướng dẫn tối ưu Go",
	"category": "Backend",
	"tags": ["go", "performance"],
	"format": "markdown",
	"content": "# Mở đầu\n\nGo nhanh nhưng **có thể nhanh hơn**.\n\n- đo trước\n- tối ưu sau\n\n` + "```go\\nb.ReportAllocs()\\n```" + `"
}`

func TestCreateArticleRequiresLogin(t *testing.T) {
	srv, _ := newAuthedClient(t, true)
	// Plain client without the session cookie → 401.
	status, env := postJSON(t, http.DefaultClient, srv.URL+v1+"/articles", markdownArticle)
	if status != http.StatusUnauthorized || env.Code != 1002 {
		t.Fatalf("anonymous create: status=%d code=%d, want 401/1002", status, env.Code)
	}
}

func TestCreateArticleForbiddenWithoutPermission(t *testing.T) {
	srv, client := newAuthedClient(t, false)

	var me struct {
		User struct {
			CanWrite bool `json:"canWrite"`
		} `json:"user"`
	}
	getClientEnv(t, client, srv.URL+v1+"/auth/me", &me)
	if me.User.CanWrite {
		t.Error("canWrite should be false without the IAM permission")
	}

	status, env := postJSON(t, client, srv.URL+v1+"/articles", markdownArticle)
	if status != http.StatusForbidden || env.Code != 3005 {
		t.Fatalf("create without permission: status=%d code=%d, want 403/3005", status, env.Code)
	}
}

func TestCreateArticleMarkdownFlow(t *testing.T) {
	srv, client := newAuthedClient(t, true)

	var me struct {
		User struct {
			CanWrite bool `json:"canWrite"`
		} `json:"user"`
	}
	getClientEnv(t, client, srv.URL+v1+"/auth/me", &me)
	if !me.User.CanWrite {
		t.Error("canWrite should be true with the IAM permission")
	}

	status, env := postJSON(t, client, srv.URL+v1+"/articles", markdownArticle)
	if status != http.StatusCreated || env.Code != 0 {
		t.Fatalf("create: status=%d code=%d message=%q, want 201/0", status, env.Code, env.Message)
	}
	var created struct {
		Slug   string `json:"slug"`
		Author string `json:"author"`
		Read   string `json:"read"`
		Body   []struct {
			Type  string   `json:"type"`
			Text  string   `json:"text"`
			Lang  string   `json:"lang"`
			Items []string `json:"items"`
		} `json:"body"`
	}
	if err := json.Unmarshal(env.Data, &created); err != nil {
		t.Fatal(err)
	}
	if created.Slug != "huong-dan-toi-uu-go" {
		t.Errorf("slug = %q, want huong-dan-toi-uu-go", created.Slug)
	}
	if created.Author != "Tác Giả" {
		t.Errorf("author = %q, want session user name", created.Author)
	}
	if created.Read == "" {
		t.Error("read time should be derived")
	}
	types := make([]string, len(created.Body))
	for i, b := range created.Body {
		types[i] = b.Type
	}
	if got := strings.Join(types, ","); got != "h,p,list,code" {
		t.Errorf("body types = %s, want h,p,list,code", got)
	}

	// The article is publicly readable at its slug.
	var detail map[string]any
	if status, _ := getClientEnv(t, client, srv.URL+v1+"/articles/huong-dan-toi-uu-go", &detail); status != 200 {
		t.Fatalf("get created article status = %d", status)
	}
	if detail["title"] != "Hướng dẫn tối ưu Go" {
		t.Errorf("title = %v", detail["title"])
	}

	// Same title again → deduplicated slug.
	status, env = postJSON(t, client, srv.URL+v1+"/articles", markdownArticle)
	if status != http.StatusCreated {
		t.Fatalf("duplicate-title create status = %d", status)
	}
	created.Slug = ""
	if err := json.Unmarshal(env.Data, &created); err != nil {
		t.Fatal(err)
	}
	if created.Slug != "huong-dan-toi-uu-go-2" {
		t.Errorf("dedup slug = %q, want huong-dan-toi-uu-go-2", created.Slug)
	}
}

func TestCreateArticleValidation(t *testing.T) {
	srv, client := newAuthedClient(t, true)

	cases := map[string]string{
		"empty title":    `{"title":"","category":"Backend","format":"markdown","content":"x"}`,
		"empty category": `{"title":"T","category":"","format":"markdown","content":"x"}`,
		"bad format":     `{"title":"T","category":"Backend","format":"html","content":"x"}`,
		"empty body":     `{"title":"T","category":"Backend","format":"markdown","content":"   "}`,
		"bad block type": `{"title":"T","category":"Backend","format":"blocks","body":[{"type":"iframe","text":"x"}]}`,
		"unknown field":  `{"title":"T","category":"Backend","format":"blocks","body":[{"type":"p","text":"x","html":"<b>x</b>"}]}`,
	}
	for name, body := range cases {
		status, env := postJSON(t, client, srv.URL+v1+"/articles", body)
		if status != http.StatusBadRequest {
			t.Errorf("%s: status=%d code=%d, want 400", name, status, env.Code)
		}
	}
}

// getClientEnv mirrors getEnv but sends through a cookie-carrying client.
func getClientEnv(t *testing.T, client *http.Client, url string, dst any) (int, envelope) {
	t.Helper()
	res, err := client.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer res.Body.Close()
	var env envelope
	if err := json.NewDecoder(res.Body).Decode(&env); err != nil {
		t.Fatalf("decode envelope %s: %v", url, err)
	}
	if dst != nil && len(env.Data) > 0 {
		if err := json.Unmarshal(env.Data, dst); err != nil {
			t.Fatalf("unmarshal data %s: %v", url, err)
		}
	}
	return res.StatusCode, env
}
