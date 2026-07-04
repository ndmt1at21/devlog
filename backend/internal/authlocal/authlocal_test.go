package authlocal

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ndmt1at21/devlog/backend/internal/authn"
	"github.com/ndmt1at21/devlog/backend/internal/repository/memory"
)

func newProvider() *Provider {
	return New(memory.New(), "test-secret", GoogleConfig{})
}

func register(t *testing.T, p *Provider, email, pass, name string) {
	t.Helper()
	if err := p.Register(context.Background(), email, pass, name); err != nil {
		t.Fatalf("register %s: %v", email, err)
	}
}

func TestRegisterLoginUserInfo(t *testing.T) {
	p := newProvider()
	ctx := context.Background()
	register(t, p, "Author@Example.com", "secret123", "Tác Giả")

	// Login normalizes the email and returns a full token set.
	ts, err := p.Login(ctx, "author@example.com", "secret123")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if ts.AccessToken == "" || ts.RefreshToken == "" || ts.ExpiresIn <= 0 {
		t.Fatalf("token set incomplete: %+v", ts)
	}

	u, err := p.UserInfo(ctx, ts.AccessToken)
	if err != nil {
		t.Fatalf("userinfo: %v", err)
	}
	if u.Email != "author@example.com" || u.Name != "Tác Giả" || u.Sub == "" {
		t.Fatalf("userinfo = %+v", u)
	}
}

func TestLoginRejectsBadCredentials(t *testing.T) {
	p := newProvider()
	ctx := context.Background()
	register(t, p, "a@b.co", "secret123", "A")

	if _, err := p.Login(ctx, "a@b.co", "wrong-pass"); !errors.Is(err, authn.ErrInvalidCredentials) {
		t.Fatalf("wrong password err = %v, want ErrInvalidCredentials", err)
	}
	if _, err := p.Login(ctx, "nobody@b.co", "secret123"); !errors.Is(err, authn.ErrInvalidCredentials) {
		t.Fatalf("unknown email err = %v, want ErrInvalidCredentials", err)
	}
}

func TestRegisterConflict(t *testing.T) {
	p := newProvider()
	register(t, p, "a@b.co", "secret123", "A")
	if err := p.Register(context.Background(), "A@B.CO", "other-pass", "B"); !errors.Is(err, authn.ErrConflict) {
		t.Fatalf("duplicate register err = %v, want ErrConflict", err)
	}
}

func TestFirstUserBecomesAuthor(t *testing.T) {
	p := newProvider()
	ctx := context.Background()
	register(t, p, "first@b.co", "secret123", "First")
	register(t, p, "second@b.co", "secret123", "Second")

	firstTS, _ := p.Login(ctx, "first@b.co", "secret123")
	secondTS, _ := p.Login(ctx, "second@b.co", "secret123")

	perms := []string{"articles:create"}
	if ok, err := p.CheckPermissions(ctx, firstTS.AccessToken, perms); err != nil || !ok {
		t.Fatalf("first user CheckPermissions = (%v, %v), want (true, nil)", ok, err)
	}
	if ok, err := p.CheckPermissions(ctx, secondTS.AccessToken, perms); err != nil || ok {
		t.Fatalf("second user CheckPermissions = (%v, %v), want (false, nil)", ok, err)
	}
}

func TestCheckPermissionsDeniesBadTokens(t *testing.T) {
	p := newProvider()
	ctx := context.Background()

	// Garbage token: plain deny, no error (PDP semantics).
	if ok, err := p.CheckPermissions(ctx, "not-a-token", []string{"articles:create"}); err != nil || ok {
		t.Fatalf("garbage token = (%v, %v), want (false, nil)", ok, err)
	}

	// Well-signed but expired token: same.
	expired, err := signJWT(p.jwtKey, claims{
		Issuer:    tokenIssuer,
		Subject:   "u1",
		IssuedAt:  time.Now().Add(-time.Hour).Unix(),
		ExpiresAt: time.Now().Add(-time.Minute).Unix(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if ok, err := p.CheckPermissions(ctx, expired, []string{"articles:create"}); err != nil || ok {
		t.Fatalf("expired token = (%v, %v), want (false, nil)", ok, err)
	}
	if _, err := p.UserInfo(ctx, expired); !errors.Is(err, authn.ErrInvalidCredentials) {
		t.Fatalf("userinfo expired err = %v, want ErrInvalidCredentials", err)
	}
}

func TestRefreshRotatesToken(t *testing.T) {
	p := newProvider()
	ctx := context.Background()
	register(t, p, "a@b.co", "secret123", "A")
	ts, _ := p.Login(ctx, "a@b.co", "secret123")

	rotated, err := p.Refresh(ctx, ts.RefreshToken)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if rotated.RefreshToken == "" || rotated.RefreshToken == ts.RefreshToken {
		t.Fatal("refresh must rotate to a new token")
	}

	// The presented token is single-use.
	if _, err := p.Refresh(ctx, ts.RefreshToken); !errors.Is(err, authn.ErrInvalidCredentials) {
		t.Fatalf("reused refresh err = %v, want ErrInvalidCredentials", err)
	}
	// The rotated token still works.
	if _, err := p.Refresh(ctx, rotated.RefreshToken); err != nil {
		t.Fatalf("rotated refresh: %v", err)
	}
}

func TestLogoutRevokesRefresh(t *testing.T) {
	p := newProvider()
	ctx := context.Background()
	register(t, p, "a@b.co", "secret123", "A")
	ts, _ := p.Login(ctx, "a@b.co", "secret123")

	if err := p.Logout(ctx, ts.RefreshToken); err != nil {
		t.Fatalf("logout: %v", err)
	}
	if _, err := p.Refresh(ctx, ts.RefreshToken); !errors.Is(err, authn.ErrInvalidCredentials) {
		t.Fatalf("refresh after logout err = %v, want ErrInvalidCredentials", err)
	}
}

func TestFederatedLoginURL(t *testing.T) {
	p := newProvider()
	if got := p.FederatedLoginURL("google", "st", "https://app/cb"); got != "" {
		t.Fatalf("unconfigured google URL = %q, want empty", got)
	}

	p = New(memory.New(), "test-secret", GoogleConfig{ClientID: "cid", ClientSecret: "cs"})
	got := p.FederatedLoginURL("google", "st", "https://app/cb")
	for _, want := range []string{googleAuthURL, "client_id=cid", "state=st", "redirect_uri=https%3A%2F%2Fapp%2Fcb", "response_type=code"} {
		if !strings.Contains(got, want) {
			t.Fatalf("google URL %q missing %q", got, want)
		}
	}
	if got := p.FederatedLoginURL("facebook", "st", "https://app/cb"); got != "" {
		t.Fatalf("unknown provider URL = %q, want empty", got)
	}
}

// TestCallGroupCoalesces proves that overlapping calls for the same key run the
// work exactly once and every caller receives that one result — the property
// Refresh relies on so a burst of concurrent refreshes rotates the token once
// instead of tearing the session down.
func TestCallGroupCoalesces(t *testing.T) {
	g := callGroup{inflight: make(map[string]*refreshCall)}

	var calls atomic.Int32
	started := make(chan struct{})
	release := make(chan struct{})
	fn := func() (*authn.TokenSet, error) {
		if calls.Add(1) == 1 {
			close(started) // first (and only) execution has begun
		}
		<-release
		return &authn.TokenSet{AccessToken: "shared"}, nil
	}

	const n = 8
	results := make([]*authn.TokenSet, n)
	var wg sync.WaitGroup
	wg.Add(n)
	for i := range n {
		go func(i int) {
			defer wg.Done()
			ts, _ := g.do("k", fn)
			results[i] = ts
		}(i)
	}

	<-started                         // one caller is inside fn holding the entry
	time.Sleep(20 * time.Millisecond) // let the rest join the in-flight call
	close(release)
	wg.Wait()

	if got := calls.Load(); got != 1 {
		t.Fatalf("fn executed %d times, want 1 (calls not coalesced)", got)
	}
	for i, r := range results {
		if r == nil || r.AccessToken != "shared" {
			t.Fatalf("result[%d] = %+v, want the shared token set", i, r)
		}
	}
}

// TestConcurrentRefreshKeepsSession fires many refreshes of the same token at
// once (the page-load stampede) and requires that they don't collectively kill
// the session: at least one succeeds and every success returns an identical,
// usable rotated token.
func TestConcurrentRefreshKeepsSession(t *testing.T) {
	p := newProvider()
	ctx := context.Background()
	register(t, p, "a@b.co", "secret123", "A")
	ts, err := p.Login(ctx, "a@b.co", "secret123")
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	const n = 16
	got := make([]*authn.TokenSet, n)
	var wg sync.WaitGroup
	wg.Add(n)
	for i := range n {
		go func(i int) {
			defer wg.Done()
			out, err := p.Refresh(ctx, ts.RefreshToken)
			if err == nil {
				got[i] = out
			}
		}(i)
	}
	wg.Wait()

	var rotated string
	for _, out := range got {
		if out == nil {
			continue
		}
		if rotated == "" {
			rotated = out.RefreshToken
		} else if out.RefreshToken != rotated {
			t.Fatalf("concurrent refreshes returned different tokens: %q vs %q", rotated, out.RefreshToken)
		}
	}
	if rotated == "" {
		t.Fatal("no concurrent refresh succeeded; the session was torn down")
	}
	// The single rotated token is valid and usable.
	if _, err := p.Refresh(ctx, rotated); err != nil {
		t.Fatalf("rotated token unusable: %v", err)
	}
}
