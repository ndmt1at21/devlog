# devlog — devnote

A coding blog ("devnote") built from `ui-design/`: a **Next.js 16** frontend and a
**Go** backend (BFF), with content in **MySQL** and authentication via the **IAM**
OAuth2/OIDC service.

## Structure

```
devlog/
├── frontend/   # Next.js 16 + React 19 + Tailwind v4 (App Router)
├── backend/    # Go HTTP API (BFF): content, comments, auth, Pro
└── ui-design/  # source design mockup (Devnote Blog.dc.html)
```

## Quickstart (zero-infra)

The backend defaults to an in-memory store seeded from the design content, so you
can run the whole blog without MySQL or IAM:

```bash
# 1) backend  → http://localhost:8080
cd backend && go run ./cmd/server

# 2) frontend → http://localhost:3000
cd frontend && npm install && npm run dev
```

Open http://localhost:3000 — browse/search articles, read articles (code blocks,
diagrams, series + Pro paywall), comment, toggle dark mode, and open the coffee
modal. Login/register work out of the box: auth is embedded in the backend (no
external service). The first account you register becomes the author.

## Backend config (`backend/.env.example`)

| Var | Purpose |
| --- | --- |
| `DB_DRIVER` | `memory` (default) or `mysql` |
| `DB_DSN` | MySQL DSN (`parseTime=true&loc=UTC&multiStatements=true`) |
| `GOOGLE_CLIENT_ID` / `GOOGLE_CLIENT_SECRET` | Google login (optional; empty disables it) |
| `SESSION_SECRET` / `COOKIE_SECURE` | session cookie sealing (also keys access tokens) |
| `APP_BASE_URL` | frontend origin for payment redirects (default `http://localhost:3000`) |
| `STRIPE_SECRET_KEY` / `STRIPE_WEBHOOK_SECRET` | Stripe Checkout (coffee); empty → demo |
| `MOMO_PARTNER_CODE` / `MOMO_ACCESS_KEY` / `MOMO_SECRET_KEY` | MoMo (coffee); empty → demo |

### MySQL mode

```bash
DB_DRIVER=mysql DB_DSN='devlog:devlog@tcp(localhost:3306)/devlog?parseTime=true&loc=UTC&multiStatements=true' \
  go run ./cmd/server      # auto-creates schema + seeds on first run
```

Schema migrations live in `backend/migrations/mysql/` (embedded, applied on
startup).

### Auth (embedded)

Auth runs in-process, embedding the IAM service's core logic instead of calling
it over HTTP: accounts live in the blog's own store (`users` +
`refresh_tokens`), passwords are argon2id-hashed (PHC-encoded), access tokens
are short-lived signed JWTs, and refresh tokens are opaque, hashed at rest, and
rotated on every use. Tokens are kept in the httpOnly session cookie as before.
Nothing needs to be provisioned — register an account and log in.

For "Đăng nhập với Google", create a Google OAuth client
(https://console.cloud.google.com/apis/credentials), register the authorized
redirect URI `{APP_BASE_URL}/api/v1/auth/google/callback`, and set
`GOOGLE_CLIENT_ID` / `GOOGLE_CLIENT_SECRET` in `backend/.env`.

> Note: with no mail service embedded, register activates the account
> immediately (no verification email) and forgot-password is a no-op that
> always reports success (anti-enumeration).

### Publishing articles (permission `articles:create`)

Logged-in users holding the **`articles:create`** permission can publish from
`/articles/new` (Markdown/README or a rich-text block editor; both normalize to
the same block model). The backend verifies the permission on every
`POST /api/v1/articles`; the account menu's "Viết bài mới" entry is only a UI
hint snapshotted into the session at login/refresh.

Permissions derive from the user's role (`reader`, `author`, `admin`; roles
`author`/`admin` hold `articles:create`). **The first registered account is
bootstrapped as `author`**; everyone after that is a `reader`. To promote
someone later, update their row (MySQL mode):
`UPDATE users SET role='author' WHERE email='...'` — it takes effect on their
next request (decisions read the live role, not the token).

## Frontend config (`frontend/.env.local`)

| Var | Purpose |
| --- | --- |
| `BACKEND_INTERNAL_URL` | Go backend base for SSR + the `/api` rewrite (default `http://localhost:8080`) |
| `NEXT_PUBLIC_GA_ID` | GA4 measurement id; empty disables analytics |

The browser reaches the API same-origin via a `/api/*` rewrite (see
`next.config.ts`) so the session cookie stays first-party. Analytics is GA4
(gtag.js) with comprehensive event tracking (`src/lib/analytics.ts`).

## API

- `GET /api/articles` · `GET /api/articles/featured` · `GET /api/categories`
- `GET /api/articles/{slug}` (server-side Pro paywall)
- `GET|POST /api/articles/{slug}/comments`
- `POST /api/auth/{login,register,forgot-password,logout}` · `GET /api/auth/me`
- `GET /api/pro/plans` · `GET|POST /api/me/subscription`
- `POST /api/coffee/checkout` · `GET /api/coffee/{id}/status` (Stripe/MoMo; demo fallback)
- `POST /api/webhooks/{stripe,momo}` (provider callbacks — reach the backend directly, not via the Next rewrite)

### Payments — "buy me a coffee"

The coffee modal supports real **Stripe Checkout** (cards) and **MoMo** (Vietnam).
With no provider keys set it falls back to a no-charge demo. Configure keys in
`backend/.env` (see `STRIPE_*` / `MOMO_*` / `APP_BASE_URL` above); after payment the
provider redirects to the frontend `/coffee/result` page, and webhooks confirm the
order server-side.

See [backend/README.md](backend/README.md) and [frontend/README.md](frontend/README.md).
