package handler

import (
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/ndmt1at21/devlog/backend/internal/authn"
	"github.com/ndmt1at21/devlog/backend/internal/config"
	"github.com/ndmt1at21/devlog/backend/internal/domain"
	"github.com/ndmt1at21/devlog/backend/internal/payment"
	"github.com/ndmt1at21/devlog/backend/internal/platform/id"
	"github.com/ndmt1at21/devlog/backend/internal/platform/logger"
	"github.com/ndmt1at21/devlog/backend/internal/session"
)

// apiV1 is the version prefix every route lives under.
const apiV1 = "/api/v1"

// API holds the dependencies shared by HTTP handlers.
type API struct {
	Store    domain.Store
	Cfg      config.Config
	Logger   *slog.Logger     // base logger; nil falls back to slog.Default()
	Auth     authn.Provider   // nil when IAM is not configured
	Sessions *session.Manager // nil when IAM is not configured
	Stripe   *payment.Stripe  // nil when Stripe is not configured
	Momo     *payment.Momo    // nil when MoMo is not configured
}

// NewRouter wires up the application's HTTP routes (all under /api/v1).
func (a *API) NewRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET "+apiV1+"/health", a.health)

	// Content
	mux.HandleFunc("GET "+apiV1+"/articles", a.listArticles)
	// Publishing requires a session + the IAM "articles:create" permission.
	mux.HandleFunc("POST "+apiV1+"/articles", a.createArticle)
	// Presigned direct-to-bucket image upload for authors (same gate as above).
	mux.HandleFunc("POST "+apiV1+"/uploads", a.createUpload)
	mux.HandleFunc("GET "+apiV1+"/articles/featured", a.featuredArticle)
	mux.HandleFunc("GET "+apiV1+"/categories", a.categories)
	mux.HandleFunc("GET "+apiV1+"/articles/{slug}", a.getArticle)
	mux.HandleFunc("GET "+apiV1+"/articles/{slug}/comments", a.listComments)
	mux.HandleFunc("POST "+apiV1+"/articles/{slug}/comments", a.createComment)

	// Reactions: public like count + the signed-in user's like/bookmark toggles.
	mux.HandleFunc("GET "+apiV1+"/articles/{slug}/reactions", a.getReactions)
	mux.HandleFunc("PUT "+apiV1+"/articles/{slug}/reactions/{kind}", a.setReaction)
	mux.HandleFunc("DELETE "+apiV1+"/articles/{slug}/reactions/{kind}", a.setReaction)
	mux.HandleFunc("GET "+apiV1+"/me/bookmarks", a.myBookmarks)

	// Auth (IAM BFF) + Pro
	mux.HandleFunc("POST "+apiV1+"/auth/login", a.login)
	mux.HandleFunc("POST "+apiV1+"/auth/register", a.register)
	mux.HandleFunc("POST "+apiV1+"/auth/forgot-password", a.forgotPassword)
	mux.HandleFunc("POST "+apiV1+"/auth/logout", a.logout)
	mux.HandleFunc("GET "+apiV1+"/auth/me", a.me)
	// Federated "Đăng nhập với Google" (redirect flow via IAM).
	mux.HandleFunc("GET "+apiV1+"/auth/google/login", a.googleLogin)
	mux.HandleFunc("GET "+apiV1+"/auth/google/callback", a.googleCallback)
	mux.HandleFunc("GET "+apiV1+"/pro/plans", a.proPlans)
	mux.HandleFunc("GET "+apiV1+"/me/subscription", a.getSubscription)
	mux.HandleFunc("POST "+apiV1+"/me/subscription", a.subscribe)

	// "Buy me a coffee" payments (Stripe Checkout + MoMo). Anonymous-allowed.
	mux.HandleFunc("POST "+apiV1+"/coffee/checkout", a.coffeeCheckout)
	mux.HandleFunc("GET "+apiV1+"/coffee/{id}/status", a.coffeeStatus)
	// Provider webhooks (server-to-server). These must be reachable at the
	// backend's public URL directly, not via the Next.js /api rewrite. They
	// return provider-specific bodies, not the JSON envelope.
	mux.HandleFunc("POST "+apiV1+"/webhooks/stripe", a.stripeWebhook)
	mux.HandleFunc("POST "+apiV1+"/webhooks/momo", a.momoWebhook)

	// Outer→inner: assign trace id, log, CORS, resolve the session user.
	return a.traceMiddleware(a.logging(cors(a.withSession(mux))))
}

// baseLogger returns the configured base logger, or the slog default when the
// API was constructed without one (e.g. in tests).
func (a *API) baseLogger() *slog.Logger {
	if a.Logger != nil {
		return a.Logger
	}
	return slog.Default()
}

func (a *API) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, r, http.StatusOK, map[string]any{
		"status": "ok",
		"driver": a.Cfg.DBDriver,
		"auth":   a.Auth != nil,
		"time":   time.Now().UTC(),
	})
}

// traceMiddleware attaches a unique trace id to each request (honoring an
// inbound X-Request-Id), exposing it via the X-Trace-Id header and the request
// context. It also seeds the context with a request-scoped logger that carries
// the trace id, so every subsequent log line — via logger.From(ctx) —
// correlates to this request without threading the id by hand.
func (a *API) traceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := r.Header.Get("X-Request-Id")
		if traceID == "" {
			traceID = id.NewV7()
		}
		w.Header().Set("X-Trace-Id", traceID)
		ctx := withTraceID(r.Context(), traceID)
		ctx = logger.WithContext(ctx, a.baseLogger().With("trace_id", traceID))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// logging emits one structured line per request — method, path, status, bytes,
// duration, client ip, and the authenticated user when present — at a level
// keyed to the status class (5xx=error, 4xx=warn, else info). The trace id is
// already attached to the context logger by traceMiddleware.
func (a *API) logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		args := []any{
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", rec.status),
			slog.Int("bytes", rec.bytes),
			slog.Float64("duration_ms", float64(time.Since(start).Microseconds())/1000),
			slog.String("remote", clientIP(r)),
		}
		if u, ok := userFrom(r.Context()); ok {
			args = append(args, slog.String("user", u.Sub))
		}

		level := slog.LevelInfo
		switch {
		case rec.status >= 500:
			level = slog.LevelError
		case rec.status >= 400:
			level = slog.LevelWarn
		}
		logger.From(r.Context()).Log(r.Context(), level, "http_request", args...)
	})
}

// statusRecorder wraps an http.ResponseWriter to capture the response status
// code and byte count for request logging.
type statusRecorder struct {
	http.ResponseWriter
	status  int
	bytes   int
	written bool
}

func (r *statusRecorder) WriteHeader(code int) {
	if !r.written {
		r.status = code
		r.written = true
	}
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	r.written = true // an implicit 200 if WriteHeader was never called
	n, err := r.ResponseWriter.Write(b)
	r.bytes += n
	return n, err
}

// Flush forwards to the underlying writer so streaming responses still work.
func (r *statusRecorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// clientIP returns the request's client address, preferring the first hop in
// X-Forwarded-For when the backend runs behind a reverse proxy.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.IndexByte(xff, ','); i >= 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}

// cors allows the Next.js frontend to call the API during development.
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
