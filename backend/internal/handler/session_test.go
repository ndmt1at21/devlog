package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/ndmt1at21/devlog/backend/internal/authn"
	"github.com/ndmt1at21/devlog/backend/internal/config"
	"github.com/ndmt1at21/devlog/backend/internal/handler"
	"github.com/ndmt1at21/devlog/backend/internal/repository/memory"
	"github.com/ndmt1at21/devlog/backend/internal/session"
)

// countingAuth is a fakeAuth whose access token is born expired and that counts
// how many times Refresh is invoked — enough to observe whether a request
// rotated the session.
type countingAuth struct {
	fakeAuth
	refreshes atomic.Int32
}

func (c *countingAuth) Login(context.Context, string, string) (*authn.TokenSet, error) {
	// ExpiresIn < 0 makes the session's stored expiry lie in the past, so the
	// very next request sees an expired access token and takes the refresh path.
	return &authn.TokenSet{AccessToken: "at", RefreshToken: "rt", ExpiresIn: -1}, nil
}

func (c *countingAuth) Refresh(context.Context, string) (*authn.TokenSet, error) {
	c.refreshes.Add(1)
	return &authn.TokenSet{AccessToken: "at2", RefreshToken: "rt2", ExpiresIn: 3600}, nil
}

// TestReadOnlySessionIsNotRotated verifies the X-Session-Read contract: a
// server-side read never rotates the session (its Set-Cookie could not reach the
// browser), yet still resolves the caller from the sealed cookie; a first-party
// browser request with the same expired cookie does refresh.
func TestReadOnlySessionIsNotRotated(t *testing.T) {
	auth := &countingAuth{}
	api := &handler.API{
		Store:    memory.New(),
		Cfg:      config.Config{DBDriver: "memory"},
		Auth:     auth,
		Sessions: session.New("test-secret", false),
	}
	srv := httptest.NewServer(api.NewRouter())
	defer srv.Close()

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

	// Read-only request: must not rotate, but must still report authenticated
	// from the sealed cookie identity.
	req, _ := http.NewRequest(http.MethodGet, srv.URL+v1+"/auth/me", nil)
	req.Header.Set("X-Session-Read", "1")
	roRes, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if got := decodeAuthenticated(t, roRes); !got {
		t.Fatal("read-only /auth/me reported unauthenticated; identity should come from the cookie")
	}
	if n := auth.refreshes.Load(); n != 0 {
		t.Fatalf("read-only request triggered %d refresh(es), want 0", n)
	}

	// A normal browser request with the same expired cookie refreshes exactly once.
	browserRes, err := client.Get(srv.URL + v1 + "/auth/me")
	if err != nil {
		t.Fatal(err)
	}
	if got := decodeAuthenticated(t, browserRes); !got {
		t.Fatal("browser /auth/me reported unauthenticated after refresh")
	}
	if n := auth.refreshes.Load(); n != 1 {
		t.Fatalf("browser request triggered %d refresh(es), want 1", n)
	}
}

// decodeAuthenticated reads the {authenticated,...} payload from a /auth/me
// envelope response and reports the authenticated flag.
func decodeAuthenticated(t *testing.T, res *http.Response) bool {
	t.Helper()
	defer res.Body.Close()
	var env struct {
		Data struct {
			Authenticated bool `json:"authenticated"`
		} `json:"data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&env); err != nil {
		t.Fatalf("decode /auth/me: %v", err)
	}
	return env.Data.Authenticated
}
