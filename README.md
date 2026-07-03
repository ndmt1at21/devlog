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
modal. Login/register/Pro need IAM (below); without it those endpoints return a
"chưa được cấu hình" message.

## Backend config (`backend/.env.example`)

| Var | Purpose |
| --- | --- |
| `DB_DRIVER` | `memory` (default) or `mysql` |
| `DB_DSN` | MySQL DSN (`parseTime=true&loc=UTC&multiStatements=true`) |
| `IAM_ISSUER_URL` | tenant issuer base, e.g. `http://localhost:8080/t/devnote` |
| `IAM_TENANT_ID` / `IAM_CLIENT_ID` / `IAM_CLIENT_SECRET` | OAuth2 client creds |
| `SESSION_SECRET` / `COOKIE_SECURE` | session cookie sealing |
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

### Auth via IAM (`/home/ndmt1at21/iam`)

The backend is a confidential OAuth2 client of IAM (password grant for login,
`/auth/register` + `/auth/forgot-password` for lifecycle, userinfo for identity;
tokens kept in an httpOnly session cookie). To enable it:

1. Run IAM: `cd ../iam && make compose-up && make run`.
2. Provision a tenant `devnote`, a **confidential client whose `grant_types`
   include `password` and `refresh_token`** (required — the password grant is
   gated per-client), and the demo user `demo@blog.vn` / `123456`.
3. Set `IAM_ISSUER_URL` (e.g. `http://localhost:8080/t/devnote`), `IAM_CLIENT_ID`,
   `IAM_CLIENT_SECRET` in `backend/.env` and restart the backend.

> Note: IAM register sends a verification email and doesn't accept a display name,
> so new accounts must verify before login; seed the demo user as active.

### Publishing articles (IAM permission `articles:create`)

Logged-in users holding the IAM permission **`articles:create`** can publish from
`/articles/new` (Markdown/README or a rich-text block editor; both normalize to
the same block model). The backend verifies the permission on every
`POST /api/v1/articles` via IAM's policy decision endpoint (`POST
{issuer}/authz/decision`); the account menu's "Viết bài mới" entry is only a UI
hint snapshotted into the session at login/refresh. To grant it, in the
`devnote` tenant:

1. Create resource `articles` with action `create`, then permission
   `articles:create`.
2. Bind the permission to a role (e.g. `author`) and assign the role to the
   writer's account. The author picks it up at next login (or token refresh).

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
