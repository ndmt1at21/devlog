package handler

import (
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/ndmt1at21/devlog/backend/internal/authn"
	"github.com/ndmt1at21/devlog/backend/internal/iam"
	"github.com/ndmt1at21/devlog/backend/internal/session"
)

var emailRe = regexp.MustCompile(`^\S+@\S+\.\S+$`)

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
			_ = a.Sessions.Save(w, data)
		}
		next.ServeHTTP(w, r.WithContext(withUser(r.Context(), a.sessionUser(r, data))))
	})
}

func (a *API) sessionUser(r *http.Request, d session.Data) *SessionUser {
	return &SessionUser{
		Sub:     d.Sub,
		Name:    d.Name,
		Email:   d.Email,
		Premium: a.userPremium(r, d.Sub),
	}
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
		writeError(w, http.StatusServiceUnavailable, "Đăng nhập chưa được cấu hình.")
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
		writeError(w, http.StatusBadRequest, "Vui lòng nhập email và mật khẩu.")
		return
	}

	ts, err := a.Auth.Login(r.Context(), email, in.Password)
	if errors.Is(err, authn.ErrInvalidCredentials) {
		writeError(w, http.StatusUnauthorized, "Email hoặc mật khẩu không đúng.")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadGateway, "Không kết nối được dịch vụ xác thực.")
		return
	}

	user, err := a.Auth.UserInfo(r.Context(), ts.AccessToken)
	if err != nil {
		writeError(w, http.StatusBadGateway, "Không lấy được thông tin người dùng.")
		return
	}
	name := user.Name
	if name == "" {
		name = emailLocal(user.Email)
	}
	data := session.Data{
		Access:  ts.AccessToken,
		Refresh: ts.RefreshToken,
		Exp:     time.Now().Add(time.Duration(ts.ExpiresIn) * time.Second).Unix(),
		Sub:     user.Sub,
		Name:    name,
		Email:   user.Email,
	}
	if err := a.Sessions.Save(w, data); err != nil {
		writeError(w, http.StatusInternalServerError, "Không tạo được phiên đăng nhập.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"authenticated": true, "user": a.sessionUser(r, data)})
}

func (a *API) register(w http.ResponseWriter, r *http.Request) {
	if a.Auth == nil {
		writeError(w, http.StatusServiceUnavailable, "Đăng ký chưa được cấu hình.")
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
		writeError(w, http.StatusBadRequest, "Vui lòng điền đầy đủ thông tin.")
		return
	}
	if !emailRe.MatchString(in.Email) {
		writeError(w, http.StatusBadRequest, "Email chưa hợp lệ.")
		return
	}
	if len(in.Password) < 6 {
		writeError(w, http.StatusBadRequest, "Mật khẩu cần tối thiểu 6 ký tự.")
		return
	}

	err := a.Auth.Register(r.Context(), in.Email, in.Password)
	if iam.ErrConflict(err) {
		writeError(w, http.StatusConflict, "Email này đã được đăng ký.")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadGateway, "Không kết nối được dịch vụ xác thực.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "verification_email_sent",
		"message": "Đã gửi email xác thực. Vui lòng kiểm tra hộp thư để kích hoạt tài khoản.",
	})
}

func (a *API) forgotPassword(w http.ResponseWriter, r *http.Request) {
	if a.Auth == nil {
		writeError(w, http.StatusServiceUnavailable, "Tính năng chưa được cấu hình.")
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
		writeError(w, http.StatusBadRequest, "Email chưa hợp lệ.")
		return
	}
	// Always report success (anti-enumeration), matching IAM's behavior.
	_ = a.Auth.ForgotPassword(r.Context(), in.Email)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (a *API) logout(w http.ResponseWriter, r *http.Request) {
	if a.Sessions != nil {
		if data, ok := a.Sessions.Load(r); ok && a.Auth != nil && data.Refresh != "" {
			_ = a.Auth.Logout(r.Context(), data.Refresh)
		}
		a.Sessions.Clear(w)
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (a *API) me(w http.ResponseWriter, r *http.Request) {
	u, ok := userFrom(r.Context())
	if !ok {
		writeJSON(w, http.StatusOK, map[string]any{"authenticated": false})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"authenticated": true, "user": u})
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
