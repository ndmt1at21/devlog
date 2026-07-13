package handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/ndmt1at21/devlog/backend/internal/apierr"
	"github.com/ndmt1at21/devlog/backend/internal/authn"
	"github.com/ndmt1at21/devlog/backend/internal/session"
)

var emailRe = regexp.MustCompile(`^\S+@\S+\.\S+$`)

// oauthStateCookie carries the CSRF state binding a federated-login start to its
// callback. Short-lived, httpOnly, SameSite=Lax so it survives the top-level
// redirect back from the IdP.
const oauthStateCookie = "devnote_oauth_state"

// googleCallbackPath is the blog redirect target Google sends the auth code
// to. {APP_BASE_URL}+googleCallbackPath must be registered as an authorized
// redirect URI on the Google OAuth client.
const googleCallbackPath = "/api/v1/auth/google/callback"

// internalReadHeader marks a request as a non-mutating, server-side read (set
// by the Next.js SSR API client on its RSC fetches). Those responses can't
// propagate our Set-Cookie back to the browser, so we must not rotate the
// session on them — see withSession.
const internalReadHeader = "X-Session-Read"

// isReadOnlySession reports whether the request is a safe, server-side read that
// must not mutate the session. Gated to GET so a stray header can never suppress
// the refresh a write actually needs.
func isReadOnlySession(r *http.Request) bool {
	return r.Method == http.MethodGet && r.Header.Get(internalReadHeader) == "1"
}

// withSession resolves the session cookie into a SessionUser on the request
// context, refreshing the access token when expired.
func (a *API) withSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.Auth == nil || a.Sessions == nil {
			next.ServeHTTP(w, r)
			return
		}
		data, ok := a.Sessions.Load(r)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}
		// Server-side reads (Next.js RSC fetches) can't write our rotated cookie
		// back to the browser, so refreshing here would revoke the refresh token
		// while the browser keeps the old cookie — logging the user out on their
		// next request. Such reads only need the sealed identity (sub / premium /
		// canWrite), never a live access token, so pass the session through
		// untouched and let the browser's own /auth/me revalidation refresh it.
		if isReadOnlySession(r) {
			next.ServeHTTP(w, r.WithContext(withUser(r.Context(), a.sessionUser(r, data))))
			return
		}
		if data.Expired() && data.Refresh != "" {
			ts, err := a.Auth.Refresh(r.Context(), data.Refresh)
			if err != nil {
				a.Sessions.Clear(w)
				next.ServeHTTP(w, r)
				return
			}
			data.Access = ts.AccessToken
			if ts.RefreshToken != "" {
				data.Refresh = ts.RefreshToken
			}
			data.Exp = time.Now().Add(time.Duration(ts.ExpiresIn) * time.Second).Unix()
			// Refresh the write-permission UI hint with the new token (best effort:
			// keep the previous snapshot when the check fails).
			if ok, err := a.Auth.CheckPermissions(r.Context(), data.Access, []string{articleCreatePermission}); err == nil {
				data.CanWrite = ok
			}
			_ = a.Sessions.Save(w, data)
		}
		next.ServeHTTP(w, r.WithContext(withUser(r.Context(), a.sessionUser(r, data))))
	})
}

func (a *API) sessionUser(r *http.Request, d session.Data) *SessionUser {
	return &SessionUser{
		Sub:      d.Sub,
		Access:   d.Access,
		Name:     d.Name,
		Email:    d.Email,
		Premium:  a.userPremium(r, d.Sub),
		CanWrite: d.CanWrite,
	}
}

// canCreateArticles snapshots the IAM write permission for the session cookie
// (UI hint only; the create endpoint re-checks authoritatively).
func (a *API) canCreateArticles(ctx context.Context, accessToken string) bool {
	ok, err := a.Auth.CheckPermissions(ctx, accessToken, []string{articleCreatePermission})
	return err == nil && ok
}

func (a *API) userPremium(r *http.Request, sub string) bool {
	if sub == "" {
		return false
	}
	s, err := a.Store.Subscriptions().GetByUser(r.Context(), sub)
	return err == nil && s != nil && s.Status == "active"
}

func (a *API) login(w http.ResponseWriter, r *http.Request) {
	if a.Auth == nil || a.Sessions == nil {
		writeError(w, r, apierr.ErrAuthNotConfigured)
		return
	}
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	email := strings.TrimSpace(in.Email)
	if email == "" || in.Password == "" {
		writeError(w, r, apierr.ErrValidation.WithMessage("Vui lòng nhập email và mật khẩu."))
		return
	}

	ts, err := a.Auth.Login(r.Context(), email, in.Password)
	if errors.Is(err, authn.ErrInvalidCredentials) {
		writeError(w, r, apierr.ErrInvalidCredentials)
		return
	}
	if err != nil {
		writeError(w, r, apierr.ErrAuthUpstream)
		return
	}

	user, err := a.Auth.UserInfo(r.Context(), ts.AccessToken)
	if err != nil {
		writeError(w, r, apierr.ErrUserInfo)
		return
	}
	name := user.Name
	if name == "" {
		name = emailLocal(user.Email)
	}
	data := session.Data{
		Access:   ts.AccessToken,
		Refresh:  ts.RefreshToken,
		Exp:      time.Now().Add(time.Duration(ts.ExpiresIn) * time.Second).Unix(),
		Sub:      user.Sub,
		Name:     name,
		Email:    user.Email,
		CanWrite: a.canCreateArticles(r.Context(), ts.AccessToken),
	}
	if err := a.Sessions.Save(w, data); err != nil {
		writeError(w, r, apierr.ErrSessionCreate)
		return
	}
	writeJSON(w, r, http.StatusOK, map[string]any{"authenticated": true, "user": a.sessionUser(r, data)})
}

func (a *API) register(w http.ResponseWriter, r *http.Request) {
	if a.Auth == nil {
		writeError(w, r, apierr.ErrAuthNotConfigured.WithMessage("Đăng ký chưa được cấu hình."))
		return
	}
	var in struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	in.Email = strings.TrimSpace(in.Email)
	if in.Name == "" || in.Email == "" || in.Password == "" {
		writeError(w, r, apierr.ErrValidation.WithMessage("Vui lòng điền đầy đủ thông tin."))
		return
	}
	if !emailRe.MatchString(in.Email) {
		writeError(w, r, apierr.ErrInvalidEmail)
		return
	}
	if len(in.Password) < 6 {
		writeError(w, r, apierr.ErrWeakPassword)
		return
	}

	err := a.Auth.Register(r.Context(), in.Email, in.Password, in.Name)
	if errors.Is(err, authn.ErrConflict) {
		writeError(w, r, apierr.ErrEmailTaken)
		return
	}
	if err != nil {
		writeError(w, r, apierr.ErrAuthUpstream)
		return
	}
	writeJSON(w, r, http.StatusOK, map[string]any{
		"status":  "registered",
		"message": "Đăng ký thành công. Bạn có thể đăng nhập ngay bây giờ.",
	})
}

func (a *API) forgotPassword(w http.ResponseWriter, r *http.Request) {
	if a.Auth == nil {
		writeError(w, r, apierr.ErrUnavailable)
		return
	}
	var in struct {
		Email string `json:"email"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	in.Email = strings.TrimSpace(in.Email)
	if !emailRe.MatchString(in.Email) {
		writeError(w, r, apierr.ErrInvalidEmail)
		return
	}
	// Always report success (anti-enumeration), matching IAM's behavior.
	_ = a.Auth.ForgotPassword(r.Context(), in.Email)
	writeJSON(w, r, http.StatusOK, map[string]any{"ok": true})
}

func (a *API) logout(w http.ResponseWriter, r *http.Request) {
	if a.Sessions != nil {
		if data, ok := a.Sessions.Load(r); ok && a.Auth != nil && data.Refresh != "" {
			_ = a.Auth.Logout(r.Context(), data.Refresh)
		}
		a.Sessions.Clear(w)
	}
	writeJSON(w, r, http.StatusOK, map[string]any{"ok": true})
}

func (a *API) me(w http.ResponseWriter, r *http.Request) {
	u, ok := userFrom(r.Context())
	if !ok {
		writeJSON(w, r, http.StatusOK, map[string]any{"authenticated": false})
		return
	}
	writeJSON(w, r, http.StatusOK, map[string]any{"authenticated": true, "user": u})
}

func emailLocal(email string) string {
	if i := strings.IndexByte(email, '@'); i > 0 {
		return email[:i]
	}
	if email == "" {
		return "Bạn"
	}
	return email
}

// googleLogin starts "Đăng nhập với Google": it sets a CSRF-state cookie and
// redirects the browser to Google's authorization URL; Google redirects back
// to googleCallback with an authorization code.
func (a *API) googleLogin(w http.ResponseWriter, r *http.Request) {
	if a.Auth == nil || a.Sessions == nil {
		http.Redirect(w, r, a.frontendURL("/login", "auth_unavailable"), http.StatusFound)
		return
	}
	state, err := randomState()
	if err != nil {
		http.Redirect(w, r, a.frontendURL("/login", "google_failed"), http.StatusFound)
		return
	}
	loginURL := a.Auth.FederatedLoginURL("google", state, a.googleRedirectURI())
	if loginURL == "" { // provider not configured (no Google client)
		http.Redirect(w, r, a.frontendURL("/login", "auth_unavailable"), http.StatusFound)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookie,
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   a.Cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600,
	})
	http.Redirect(w, r, loginURL, http.StatusFound)
}

// googleCallback is the redirect target for the federated flow. It verifies the
// state cookie, exchanges the authorization code for tokens, opens a session,
// and sends the browser back to the app. On any failure it redirects to /login
// with an error code the page surfaces inline.
func (a *API) googleCallback(w http.ResponseWriter, r *http.Request) {
	// Always clear the one-time state cookie.
	http.SetCookie(w, &http.Cookie{
		Name: oauthStateCookie, Value: "", Path: "/", MaxAge: -1,
		HttpOnly: true, Secure: a.Cfg.CookieSecure, SameSite: http.SameSiteLaxMode,
	})

	if a.Auth == nil || a.Sessions == nil {
		http.Redirect(w, r, a.frontendURL("/login", "auth_unavailable"), http.StatusFound)
		return
	}

	q := r.URL.Query()
	state, code := q.Get("state"), q.Get("code")
	c, err := r.Cookie(oauthStateCookie)
	if q.Get("error") != "" || err != nil || c.Value == "" || state == "" || code == "" || subtleMismatch(c.Value, state) {
		http.Redirect(w, r, a.frontendURL("/login", "google_failed"), http.StatusFound)
		return
	}

	ts, err := a.Auth.ExchangeCode(r.Context(), code, a.googleRedirectURI())
	if err != nil {
		http.Redirect(w, r, a.frontendURL("/login", "google_failed"), http.StatusFound)
		return
	}
	user, err := a.Auth.UserInfo(r.Context(), ts.AccessToken)
	if err != nil {
		http.Redirect(w, r, a.frontendURL("/login", "google_failed"), http.StatusFound)
		return
	}
	name := user.Name
	if name == "" {
		name = emailLocal(user.Email)
	}
	data := session.Data{
		Access:   ts.AccessToken,
		Refresh:  ts.RefreshToken,
		Exp:      time.Now().Add(time.Duration(ts.ExpiresIn) * time.Second).Unix(),
		Sub:      user.Sub,
		Name:     name,
		Email:    user.Email,
		CanWrite: a.canCreateArticles(r.Context(), ts.AccessToken),
	}
	if err := a.Sessions.Save(w, data); err != nil {
		http.Redirect(w, r, a.frontendURL("/login", "google_failed"), http.StatusFound)
		return
	}
	http.Redirect(w, r, strings.TrimRight(a.Cfg.AppBaseURL, "/")+"/", http.StatusFound)
}

func (a *API) googleRedirectURI() string {
	return strings.TrimRight(a.Cfg.AppBaseURL, "/") + googleCallbackPath
}

// frontendURL builds an app URL (optionally with an ?error= code) for redirects.
func (a *API) frontendURL(path, errCode string) string {
	base := strings.TrimRight(a.Cfg.AppBaseURL, "/") + path
	if errCode != "" {
		return base + "?error=" + url.QueryEscape(errCode)
	}
	return base
}

func randomState() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// subtleMismatch reports whether two non-empty strings differ, in constant time.
func subtleMismatch(a, b string) bool {
	if len(a) != len(b) {
		return true
	}
	var diff byte
	for i := 0; i < len(a); i++ {
		diff |= a[i] ^ b[i]
	}
	return diff != 0
}
