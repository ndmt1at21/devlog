// Package session manages the encrypted, httpOnly session cookie used by the
// BFF. The cookie carries the IAM token pair plus minimal identity so most
// requests don't need to call IAM. Contents are sealed with AES-GCM.
package session

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

var errShortCiphertext = errors.New("session: ciphertext too short")

// CookieName is the session cookie name.
const CookieName = "devnote_session"

const maxAge = 30 * 24 * time.Hour

// Data is the sealed session payload.
type Data struct {
	Access  string `json:"a"`
	Refresh string `json:"r"`
	Exp     int64  `json:"e"` // access-token expiry, unix seconds
	Sub     string `json:"s"`
	Name    string `json:"n"`
	Email   string `json:"m"`
}

// Expired reports whether the access token has passed its expiry.
func (d Data) Expired() bool { return time.Now().Unix() >= d.Exp }

// Manager seals and unseals session cookies.
type Manager struct {
	key    [32]byte
	secure bool
}

// New derives an AES key from secret. Set secure=false for plain-HTTP localhost.
func New(secret string, secure bool) *Manager {
	return &Manager{key: sha256.Sum256([]byte(secret)), secure: secure}
}

// Save seals data into the session cookie.
func (m *Manager) Save(w http.ResponseWriter, d Data) error {
	plain, err := json.Marshal(d)
	if err != nil {
		return err
	}
	sealed, err := m.seal(plain)
	if err != nil {
		return err
	}
	http.SetCookie(w, m.cookie(sealed, int(maxAge.Seconds())))
	return nil
}

// Load reads and unseals the session cookie.
func (m *Manager) Load(r *http.Request) (Data, bool) {
	c, err := r.Cookie(CookieName)
	if err != nil || c.Value == "" {
		return Data{}, false
	}
	plain, err := m.open(c.Value)
	if err != nil {
		return Data{}, false
	}
	var d Data
	if err := json.Unmarshal(plain, &d); err != nil {
		return Data{}, false
	}
	return d, true
}

// Clear expires the session cookie.
func (m *Manager) Clear(w http.ResponseWriter) {
	http.SetCookie(w, m.cookie("", -1))
}

func (m *Manager) cookie(value string, maxAgeSec int) *http.Cookie {
	return &http.Cookie{
		Name:     CookieName,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   m.secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   maxAgeSec,
	}
}

func (m *Manager) seal(plain []byte) (string, error) {
	block, err := aes.NewCipher(m.key[:])
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	out := gcm.Seal(nonce, nonce, plain, nil)
	return base64.RawURLEncoding.EncodeToString(out), nil
}

func (m *Manager) open(s string) ([]byte, error) {
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(m.key[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(raw) < gcm.NonceSize() {
		return nil, errShortCiphertext
	}
	nonce, ct := raw[:gcm.NonceSize()], raw[gcm.NonceSize():]
	return gcm.Open(nil, nonce, ct, nil)
}
