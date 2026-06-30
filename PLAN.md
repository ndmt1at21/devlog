# Devnote — Coding Blog (from ui-design)

## ✅ Status: COMPLETE (2026-06-28)

Full-stack blog implemented and verified. Build/test/lint all green; end-to-end SSR
confirmed against the live backend.

| Epic | Status | Notes |
| --- | --- | --- |
| B — Backend foundations | ✅ | config, domain, storage factory (memory\|mysql), embedded MySQL migrations |
| C — Content API + seed | ✅ | articles/categories/comments + **server-side paywall**; seeded from the design |
| D — Auth (IAM BFF) | ✅ | IAM client, AES-GCM session cookie, login/register/forgot/logout/me + refresh middleware |
| E — Pro/subscription | ✅ | plans + subscribe (demo) → premium gating wired into paywall & `/me` |
| F — FE foundations | ✅ | design tokens + Tailwind `@theme`, Be Vietnam Pro, theme (no-FOUC), `/api` rewrite, Header/Footer |
| G — Home | ✅ | featured + cards + category filter + instant search |
| H — Article | ✅ | block renderers (code copy, diagram, quote), series box, paywall, series nav, comments |
| I — Auth pages | ✅ | login/register/forgot + account menu wiring |
| J — Pro / Coffee / Ads | ✅ | Pro page, coffee modal (amount→pay→done), in-content ad slot + premium gating |
| K — Analytics (GA4) | ✅ | gtag bootstrap + route page_view, `track()` wrapper, comprehensive events, scroll-depth + ad impressions |
| L — Polish | ✅ | responsive, scroll progress, SEO metadata, not-found + error boundary |
| A / M — IAM provisioning + verify | ◑ | code complete + `go test`/SSR verified; live IAM provisioning documented (needs Docker/IAM running — unavailable in this env) |

**Verified:** `go build`/`go vet`/`go test ./...` pass; `npm run build`/`npm run lint`
clean; both servers run and SSR renders home, articles, code/diagram blocks, comments,
and the `iam-2` paywall (2 blocks + upgrade CTA, comments hidden). **Deferred (env):**
live IAM login/register and a live MySQL run — Docker/MySQL/IAM are not available here; a
zero-infra in-memory backend was added so the app runs and is fully verifiable now (see
README Quickstart). MySQL implementation + migrations and the IAM BFF client are complete
and compile; enable them per the README when infra is available.

## Context

[ui-design/Devnote Blog.dc.html](ui-design/Devnote%20Blog.dc.html) is a complete, self-contained
SPA mockup of **devnote**, a Vietnamese coding blog. It is a design-tool export (`.dc.html`) whose
`<script type="text/x-dc">` block is essentially a React-style class component (state, handlers,
`renderVals`) with baked-in sample data. We need to turn this mockup into a real product in the
existing monorepo:

- **frontend/** — Next.js 16 + React 19 + Tailwind v4 (App Router) — currently the default
  create-next-app starter ([frontend/src/app/page.tsx](frontend/src/app/page.tsx)).
- **backend/** — Go stdlib `net/http` API — currently two demo routes
  ([backend/internal/handler/handler.go](backend/internal/handler/handler.go)).

**Decisions already made (by the user):** Full-stack (FE + Go API + DB) · **MySQL** ·
authentication **integrated with the existing IAM service** at `/home/ndmt1at21/iam`
(`github.com/teko/iam`), a mature multi-tenant OAuth2/OIDC provider.

**Outcome:** a working, faithful-to-design blog — browse/search articles, read articles (code blocks,
diagrams, quotes, series with a Pro paywall), comment anonymously, register/login/forgot-password via
IAM, upgrade to Pro (demo), and a "buy me a coffee" demo flow — with light/dark theming.

---

## Architecture

```
Browser ──(same-origin /api/*)──► Next.js (rewrites + route handlers) ──► Go API (BFF) ──► MySQL
                                                                            │
                                                                            └──► IAM (OAuth2/OIDC)
```

- **BFF (Backend-For-Frontend):** the Go backend is a *confidential OAuth2 client* of IAM. It holds
  `client_id`/`client_secret`, exchanges credentials for tokens, and stores them in **httpOnly,
  Secure, SameSite=Lax session cookies**. The browser never sees IAM tokens.
- **Same-origin to the browser:** Next.js `rewrites` map `/api/:path*` → Go backend, so session
  cookies are first-party on `:3000` and SSR works without CORS gymnastics. (Auth responses that set
  cookies go through thin Next.js route handlers / pass-through rewrites.)
- **Content (articles/series) lives in MySQL**, seeded from the design's data. **Comments** are
  anonymous (name + text). **Pro/premium** is blog-owned state keyed by the IAM user id (`sub`).
  **Coffee** is a client-side demo (optional logging endpoint).

---

## IAM integration (prerequisites — do first)

IAM is a separate service/module; the blog consumes it over HTTP (optionally via the Go SDK at
[../iam/sdk/go](../iam/sdk/go), `module github.com/teko-vn/iam-go-sdk`, pulled in with a `replace`
directive). Endpoints/conventions confirmed in [../iam/api/openapi.yaml](../iam/api/openapi.yaml).

1. **Run IAM locally:** `cd ../iam && make compose-up` (brings up MySQL/Redis/etc.) then `make run`.
2. **Provision a tenant + client + user** for the blog (extend `../iam/scripts/seed.go` or use the
   management API):
   - Tenant `devnote`.
   - A **confidential OAuth2 client** whose `grant_types` include **`password`** and
     **`refresh_token`** — the password grant is gated per-client
     ([../iam/internal/auth/grant/password.go](../iam/internal/auth/grant/password.go) line ~15), so
     this is mandatory for the design's custom login form to work.
   - The demo user from the design: `demo@blog.vn` / `123456` (active, password set).
3. **Capture config** for the blog backend: management base URL, the **tenant issuer / protocol base**
   (dev path-fallback form, e.g. `http://localhost:8080/t/devnote`), `tenant_id`, `client_id`,
   `client_secret`, and JWKS URL (`{issuer}/.well-known/jwks.json`).

**Auth flows the blog uses (preserves the design's own forms — no hosted-login redirect):**

| Design action | Blog backend route | IAM call |
|---|---|---|
| Login form | `POST /api/auth/login` | `POST /oauth2/token` `grant_type=password` → set session cookies |
| Register form | `POST /api/auth/register` | `POST /auth/register` (then auto-login via password grant) |
| Forgot password | `POST /api/auth/forgot-password` | `POST /auth/forgot-password` (show "check inbox") |
| Logout | `POST /api/auth/logout` | `POST /oauth2/revoke` + clear cookies |
| Current user | `GET /api/auth/me` | verify access token via **JWKS** (or `/oauth2/introspect`); refresh with `refresh_token` grant when expired |

The Go SDK covers all of these (`OAuth2Service.Password/RefreshToken/Revoke/Introspect/UserInfo/JWKS`
in [../iam/sdk/go/oauth2.go](../iam/sdk/go/oauth2.go); user-lifecycle service for register/forgot).

---

## Backend plan (`backend/`)

Keep the existing stdlib `net/http` + `http.ServeMux` style (Go 1.22 method-pattern routing, as in
[backend/internal/handler/handler.go](backend/internal/handler/handler.go)). Mirror IAM's MySQL
conventions.

**Proposed structure (extends what exists):**
```
backend/
  cmd/server/main.go                 # exists — wire config, DB, IAM client, router
  internal/config/                   # env config (DB DSN, IAM_* , session secret, ports)
  internal/domain/                   # entities: Article, Block, Series, Comment, Subscription
  internal/repository/mysql/         # article.go, comment.go, subscription.go, mysql.go (Repos+mapError)
  internal/iam/                      # IAM client wrapper (token exchange, register, JWKS verify)
  internal/session/                  # httpOnly cookie session (encode/verify, refresh)
  internal/handler/                  # article.go, comment.go, auth.go, subscription.go, response.go
  migrations/mysql/                  # 0001_init.up/down.sql, 0002_seed.up/down.sql (go:embed)
  scripts/seed/                      # load design's 10 articles + iam series into MySQL
```

**MySQL conventions to copy from [../iam/internal/repository/mysql/mysql.go](../iam/internal/repository/mysql/mysql.go):**
DSN `parseTime=true&loc=UTC`; IDs `CHAR(36)` UUIDv7 (app-side); `DATETIME(6)` with app-set UTC
timestamps; JSON columns for arrays/objects; migration files `NNNN_name.up.sql`/`.down.sql`
(golang-migrate naming) embedded via `go:embed`; `Repos` bundle of per-entity structs; `mapError`
(`sql.ErrNoRows`→NotFound, MySQL `1062`→Conflict); domain error sentinels.

**Schema (new tables — public id = `slug`, matching design ids like `ai-agents`):**
- `articles`: id, slug (unique), category, author_name, read_time, published_at, title, excerpt,
  cover (nullable), tags JSON, series_slug (nullable), series_part INT (nullable), part_title
  (nullable), body JSON (ordered blocks `{type,text|lang/code|steps/caption}`), created_at, updated_at.
- `series`: slug (PK), title, description.  (Order/parts derived from `articles.series_part`.)
- `comments`: id, article_slug, name, body, created_at.
- `subscriptions`: id, user_id (IAM `sub`), plan (`month|year`), status (`active`), created_at.
- *(optional)* `coffee_orders` for the demo.

**API endpoints:**
- `GET /api/articles?category=&q=` — list (computes `isSeries`/`seriesBadge`); `GET /api/articles/featured`.
- `GET /api/categories` — derived list (`Tất cả` + distinct categories).
- `GET /api/articles/{slug}` — detail with series info (parts, prev/next) and **server-side paywall**:
  if `series && part>1 && !premium`, return only the first 2 blocks + `locked:true` (mirrors
  `renderVals` in the design). Premium read from the session.
- `GET /api/articles/{slug}/comments` · `POST /api/articles/{slug}/comments` `{name,text}` (anonymous).
- Auth: `POST /api/auth/{login,register,forgot-password,logout}`, `GET /api/auth/me`.
- Pro: `GET /api/pro/plans`, `GET /api/me/subscription`, `POST /api/me/subscription` (demo → inserts
  active subscription, flips premium).
- Update CORS/middleware in [backend/internal/handler/handler.go](backend/internal/handler/handler.go)
  (it currently allows `*`) — since the browser talks to Next, the Go API only needs to trust the
  Next origin / accept forwarded cookies.

---

## Frontend plan (`frontend/`)

> ⚠️ **This is Next.js 16 (non-standard).** Per [frontend/AGENTS.md](frontend/AGENTS.md), **read the
> relevant guides in `frontend/node_modules/next/dist/docs/01-app/` before writing any code** —
> routing, server/client components, fonts, metadata, and rewrites may differ from training data.

**Routes (App Router) — replace the starter pages:**
- `/` → Home — featured card + category filter + search + responsive grid.
- `/articles/[slug]` → Article — tags, author meta, series box, cover, block renderer, paywall,
  series prev/next nav, comments. (Slug path is easily swapped to Vietnamese `/bai-viet/[slug]`.)
- `/login`, `/register`, `/forgot-password`, `/pro`.

**Theming — port the design's token system** (it already defines a full light/dark palette in the
mockup's `<style>` `:root` + `[data-theme="dark"]`): move those CSS variables into
[frontend/src/app/globals.css](frontend/src/app/globals.css), set `data-theme` on `<html>`, and map
the tokens into Tailwind v4 via `@theme` so components use utilities (`bg-surface`, `text-body`,
`border-default`, accent, etc.) backed by the variables. Persist theme in `localStorage` with a tiny
pre-hydration inline script to avoid FOUC. Accent defaults to `#f5c700`.

**Fonts:** swap Geist for **Be Vietnam Pro** via `next/font/google` in
[frontend/src/app/layout.tsx](frontend/src/app/layout.tsx); add `<Header/>`, `<Footer/>`, theme +
auth providers there.

**Components** (`src/components/`), mapped 1:1 from the mockup sections:
- layout: `Header` (logo, search, coffee button, `AccountMenu`), `Footer`, `ScrollProgress`.
- theme: `ThemeProvider`, `ThemeToggle`.
- home: `FeaturedCard`, `ArticleCard`, `CategoryFilter`, `SearchBox`.
- article: block renderers `Paragraph`/`Heading`/`Quote`/`CodeBlock` (with copy button)/`Diagram`/
  `AdSlot`, plus `SeriesBox`, `Paywall`, `SeriesNav`, `Comments` (form + list).
- coffee: `CoffeeModal` (amount → pay [card/MoMo] → done steps).
- pro: `Pricing` (month/year plans).
- auth: `LoginForm`, `RegisterForm`, `ForgotForm`.

**Data & rendering:** article list/detail pages are **Server Components** fetching the Go API
(SEO-friendly); interactive parts are **Client Components** (search, theme, account menu, coffee
modal, comment form, code-copy, scroll progress). `src/lib/`: `api.ts` (server+client fetch wrapper),
`types.ts`, `auth.ts` (`useUser` via `/api/auth/me`), `format.ts` (initials, etc.).

**Session/proxy:** add `rewrites` in [frontend/next.config.ts](frontend/next.config.ts) mapping
`/api/:path*` → Go backend; SSR uses an internal backend URL and forwards incoming cookies. Keeps
httpOnly session cookies first-party.

**Ads / premium gating (client):** insert the in-content `AdSlot` at block index 3 when
`showAds && !premium && blocks>3`; hide ads and unlock series when `premium` (from `/api/auth/me`),
mirroring the mockup's logic.

---

## Analytics & event tracking — GA4 (gtag.js)

Google Analytics 4 via **gtag.js**, loaded with Next's `@next/third-parties/google`
`<GoogleAnalytics gaId={process.env.NEXT_PUBLIC_GA_ID!} />` in
[frontend/src/app/layout.tsx](frontend/src/app/layout.tsx) (handles script load + SPA `page_view` on
App Router navigations). **No consent gate** — analytics loads immediately. Env:
`NEXT_PUBLIC_GA_ID`. *(Next 16 is non-standard — verify `@next/third-parties` is available; if not,
fall back to a manual gtag bootstrap via `next/script` + a route-change `page_view` effect.)*

**Central wrapper** `src/lib/analytics.ts`: a typed, SSR-safe `track(name, params)` that calls
`window.gtag('event', name, params)` (guards `typeof window`). All custom events go through it so
instrumentation is consistent and testable.

**Comprehensive event taxonomy** (GA4 recommended names where they fit, custom otherwise):

| Event | Where it fires | Key params |
|---|---|---|
| `page_view` | auto (every route) | — |
| `view_article` | Article page mount | slug, category, series_slug, is_premium |
| `select_article` | `FeaturedCard`/`ArticleCard`/`SeriesNav`/`SeriesBox` | slug, title, category, list (featured\|grid\|series), position |
| `search` | `SearchBox` (debounced) | search_term, results_count |
| `select_category` | `CategoryFilter` | category |
| `select_tag` | article tag pills | tag |
| `sign_up` / `login` / `logout` | auth forms / `AccountMenu` | method: `password` |
| `select_pro_plan` / `subscribe_pro` | `Pricing` / subscribe | plan, price, value, currency `VND` |
| `coffee_open` / `coffee_donate` | `CoffeeModal` | amount, method, value, currency `VND` |
| `copy_code` | `CodeBlock` copy button | language, slug |
| `toggle_theme` | `ThemeToggle` | theme |
| `series_nav` | `SeriesNav` prev/next | direction, from_slug, to_slug |
| `paywall_view` / `paywall_upgrade_click` | `Paywall` | slug, series_slug |
| `ad_impression` / `ad_click` | `AdSlot` | slot |
| `scroll_depth` | article pages | percent (25/50/75/100), slug |

**Helpers:** `components/analytics/ScrollDepthTracker.tsx` (client; fires `scroll_depth` thresholds on
article pages) and `hooks/useImpression.ts` (IntersectionObserver → one `ad_impression` per slot;
reusable for any view-tracked element). Wire `track(...)` calls into each component's existing click
handlers as they are built (the mockup already centralizes these handlers — e.g. `openArticle`,
`setCategory`, `copyCode`, `subscribePro`, `sendCoffee` — so each maps cleanly to one event).

---

## Task breakdown

Tasks are grouped into epics (A–M). IDs show dependencies. Each task is an independent unit with
steps, key files, and an acceptance check. Rough order: A → B → C → (D, F) → (E, G, H) → I → J → K → L → M.

### Epic A — IAM provisioning (unblocks auth)

- **A1. Run IAM locally.** `cd ../iam`, `cp .env.example .env` (set `IAM_DB_DRIVER=mysql`,
  `IAM_DB_DSN`, `IAM_ENCRYPTION_KEY`), `make compose-up`, `make run`. Files: `../iam/.env`.
  *Done:* `curl localhost:8080/readyz` returns ready.
- **A2. Provision blog tenant + OAuth2 client + demo user.** Extend `../iam/scripts/seed.go` (or call
  the management API) to create tenant `devnote`, a **confidential client with `grant_types` =
  [`password`,`refresh_token`]**, and user `demo@blog.vn`/`123456` (active). Files:
  `../iam/scripts/seed/…`. *Done:* password-grant `curl` to `/oauth2/token` returns an access token.
- **A3. Record IAM config** (issuer/protocol base, tenant_id, client_id/secret, JWKS URL) for Epic D.
  *Done:* values captured in `backend/.env`.

### Epic B — Backend foundations

- **B1. Config loader.** `backend/internal/config/config.go` — read `PORT`, `DB_DSN`, `IAM_*`,
  `SESSION_SECRET`, `APP_BASE_URL` from env; update `backend/.env.example`. *Done:* `config.Load()`
  validates required vars.
- **B2. MySQL connection + migrations runner.** `backend/internal/repository/mysql/mysql.go`
  (`Repos` bundle, `mapError`, querier — copy conventions from
  [../iam/internal/repository/mysql/mysql.go](../iam/internal/repository/mysql/mysql.go)); embed
  `backend/migrations/mysql/*.sql` via `go:embed` and run on startup. *Done:* server boots and applies
  migrations against MySQL.
- **B3. Domain entities.** `backend/internal/domain/` — `Article`, `Block`, `Series`, `Comment`,
  `Subscription` + repository interfaces + error sentinels (`ErrNotFound`, `ErrConflict`). *Done:*
  package compiles.
- **B4. Schema migration `0001_init`.** `articles`, `series`, `comments`, `subscriptions` tables
  (CHAR(36) ids, DATETIME(6), JSON for tags/body) per the schema above. *Done:* up/down apply cleanly.

### Epic C — Backend content API + seed

- **C1. Article repository.** CRUD/queries in `backend/internal/repository/mysql/article.go`: list
  (filter by category, search across title/excerpt/category/tags), get-by-slug, featured. *Done:* unit
  tests pass.
- **C2. Comment repository.** `comment.go` — list by article_slug, insert. *Done:* tests pass.
- **C3. Seed script.** `backend/scripts/seed/` — load the 10 articles + `iam` series + sample comments
  from the mockup's data ([ui-design/Devnote Blog.dc.html](ui-design/Devnote%20Blog.dc.html) lines
  ~563–708). *Done:* `go run ./scripts/seed` populates MySQL.
- **C4. Article handlers.** `backend/internal/handler/article.go` — `GET /api/articles?category=&q=`,
  `GET /api/articles/featured`, `GET /api/categories`, `GET /api/articles/{slug}` with **server-side
  paywall** (truncate to 2 blocks + `locked:true` when `series && part>1 && !premium`) and series
  prev/next. Register in `handler.go`. *Done:* curl returns expected JSON incl. locked behavior.
- **C5. Comment handlers.** `GET`/`POST /api/articles/{slug}/comments` (anonymous `{name,text}`,
  validation mirroring the mockup). *Done:* POST then GET shows the new comment.

### Epic D — Backend auth (IAM BFF) — *depends on A, B1*

- **D1. IAM client wrapper.** `backend/internal/iam/` — wrap token exchange (password/refresh/revoke),
  register, forgot-password, and JWKS verify; use the Go SDK ([../iam/sdk/go/oauth2.go](../iam/sdk/go/oauth2.go))
  via a `replace` in `backend/go.mod`, or direct HTTP. *Done:* `Login(email,pass)` returns a token set.
- **D2. Session cookies.** `backend/internal/session/` — httpOnly/Secure/SameSite=Lax signed cookie
  storing tokens; helpers to read/refresh/clear. *Done:* round-trips a session.
- **D3. Auth handlers + middleware.** `backend/internal/handler/auth.go` —
  `POST /api/auth/{login,register,forgot-password,logout}`, `GET /api/auth/me`; middleware that
  validates/refreshes the access token and exposes the user (`sub`, name, email, premium). *Done:*
  login sets a cookie; `/api/auth/me` returns the user; logout clears it.

### Epic E — Backend Pro/subscription — *depends on B, D*

- **E1. Subscription repo + handlers.** `subscription.go` repo (by user_id) + `GET /api/pro/plans`,
  `GET /api/me/subscription`, `POST /api/me/subscription` (demo → active row, sets premium). Wire
  premium into C4's paywall + `/api/auth/me`. *Done:* subscribing flips a locked article to unlocked.

### Epic F — Frontend foundations

> First: read `frontend/node_modules/next/dist/docs/01-app/` (routing, fonts, metadata, rewrites).

- **F1. Design tokens + Tailwind theme.** Port the mockup's `:root`/`[data-theme="dark"]` palette into
  [frontend/src/app/globals.css](frontend/src/app/globals.css) and map tokens via Tailwind v4 `@theme`.
  *Done:* utilities like `bg-surface`/`text-body` resolve in both themes.
- **F2. Root layout.** [frontend/src/app/layout.tsx](frontend/src/app/layout.tsx) — Be Vietnam Pro
  (`next/font/google`), `<ThemeProvider>`, `<Header/>`, `<Footer/>`, VN metadata. *Done:* shell renders
  on every route.
- **F3. Theme provider + toggle + no-FOUC.** `components/theme/*` — persist to `localStorage`,
  pre-hydration inline script sets `data-theme`. *Done:* toggle persists across reloads, no flash.
- **F4. API client + types + proxy.** `src/lib/{api.ts,types.ts}`; `rewrites` in
  [frontend/next.config.ts](frontend/next.config.ts) mapping `/api/:path*` → Go backend; SSR forwards
  cookies. *Done:* a server component can fetch `/api/articles`.
- **F5. Header & Footer.** `components/layout/{Header,Footer,AccountMenu}.tsx` — logo, search box,
  coffee button, account dropdown (login/register/pro/theme/logout states). *Done:* matches mockup
  header/footer.

### Epic G — Frontend Home — *depends on C, F*

- **G1. Home page.** [frontend/src/app/page.tsx](frontend/src/app/page.tsx) (Server Component) +
  `components/home/{FeaturedCard,ArticleCard}.tsx`. *Done:* featured + responsive grid from the API.
- **G2. Category filter + search (client islands).** `CategoryFilter`, `SearchBox` with the mockup's
  search summary / empty states. *Done:* filtering + search update the grid.

### Epic H — Frontend Article — *depends on C, F*

- **H1. Article route + header.** `app/articles/[slug]/page.tsx` — title, tags, author meta, cover,
  `generateMetadata`. *Done:* renders article header from the API.
- **H2. Block renderer.** `components/article/blocks/{Paragraph,Heading,Quote,CodeBlock,Diagram,AdSlot}.tsx`
  — `CodeBlock` has a working copy button. *Done:* all block types render per mockup.
- **H3. Series box + paywall + nav.** `SeriesBox`, `Paywall`, `SeriesNav` driven by the API's
  `locked`/series data. *Done:* locked part shows 2 blocks + paywall; nav works.
- **H4. Comments.** `components/article/Comments.tsx` — list + anonymous form wired to C5. *Done:*
  posting a comment shows it immediately.

### Epic I — Frontend auth — *depends on D, F*

- **I1. Auth pages.** `app/{login,register,forgot-password}/page.tsx` + `components/auth/*` forms,
  posting to `/api/auth/*`; show demo-account hint, error states. *Done:* login/register/forgot work
  end-to-end against IAM.
- **I2. Account menu wiring.** `AccountMenu` reflects `/api/auth/me` (initial/name/email), logout.
  *Done:* logged-in vs anonymous states match mockup.

### Epic J — Frontend Pro, Coffee, Ads — *depends on E, F, H2*

- **J1. Pro page.** `app/pro/page.tsx` + `components/pro/Pricing.tsx` (month/year), subscribe (demo)
  → premium; success state. *Done:* subscribing unlocks series + hides ads.
- **J2. Coffee modal.** `components/coffee/CoffeeModal.tsx` — amount → pay (card/MoMo) → done steps
  (client demo). *Done:* full flow matches mockup.
- **J3. Ad slots + gating.** `AdSlot` inserted at block index 3 when `showAds && !premium && blocks>3`.
  *Done:* ads show for free users, hidden for premium.

### Epic K — Analytics (GA4) — *cross-cutting, depends on F*

- **K1. GA4 bootstrap.** `@next/third-parties` `<GoogleAnalytics>` in layout (+ `NEXT_PUBLIC_GA_ID`),
  fallback gtag `<Script>` if unavailable. *Done:* `page_view` shows in GA4 Realtime.
- **K2. `track()` wrapper.** `src/lib/analytics.ts` (typed, SSR-safe). *Done:* a test event appears in
  DebugView.
- **K3. Wire events** into component handlers per the taxonomy table (article/search/category/tag/
  auth/pro/coffee/copy/theme/series_nav/paywall/ad). *Done:* each event fires with expected params.
- **K4. Scroll-depth + ad impressions.** `components/analytics/ScrollDepthTracker.tsx` +
  `hooks/useImpression.ts`. *Done:* `scroll_depth` and `ad_impression` fire.

### Epic L — Polish

- **L1. Responsive + scroll progress.** Mobile/tablet/desktop breakpoints (mockup `renderVals` rsp
  map) + article reading-progress bar. **L2. SEO + states.** sitemap/metadata, empty/error/loading
  states. *Done:* matches mockup across viewports.

---

## Epic M — End-to-end verification

- **IAM up:** `cd ../iam && make compose-up && make run`; provision tenant/client/user; confirm
  `curl -d 'grant_type=password&client_id=...&client_secret=...&username=demo@blog.vn&password=123456'
  {issuer}/oauth2/token` returns tokens.
- **Backend:** `cd backend && make run`; `curl /api/articles`, `/api/articles/ai-agents`,
  `/api/categories`; POST a comment then GET it; `go test ./...`.
- **Frontend:** `cd frontend && npm run dev`; click through every screen and check against the
  mockup — theme toggle (persists), search, category filter, article blocks (code **copy**, diagram,
  quote), responsive layout.
- **End-to-end paywall:** open series part `iam-2` while logged out → see **Paywall** (first 2 blocks
  only) → login as `demo@blog.vn` → **subscribe Pro** (demo) → content unlocks and ads disappear.
- **Auth e2e:** register a new user (created in IAM) → header shows initial/name → logout → login again.
- **Analytics:** with `NEXT_PUBLIC_GA_ID` set, open GA4 **Realtime / DebugView** and confirm
  `page_view` on navigation plus custom events (`select_article`, `search`, `copy_code`,
  `subscribe_pro`, `coffee_donate`, `scroll_depth`, `ad_impression`) fire with expected params.

---

## Open notes / risks

- **Next.js 16 is non-standard** — consult `frontend/node_modules/next/dist/docs/` (rewrites, fonts,
  RSC, metadata) before coding; APIs may differ from memory.
- **Password grant is per-client gated** — the blog's IAM client *must* enable `password` +
  `refresh_token` grants, or login returns `invalid_client`/`unauthorized_grant`.
- **Premium ownership** — kept in the blog DB (not IAM); `subscribe Pro` is a demo (no real payment),
  same for the coffee flow, matching the mockup ("bản demo — không phát sinh thanh toán thật").
- **Dev issuer URL** — confirm IAM's per-tenant protocol base in dev (issuer origin vs `/t/{tenant}`
  path fallback) from IAM config/docs.
- **Go SDK linkage** — `../iam/sdk/go` is a local module; use a `replace` directive in
  `backend/go.mod`, or call the documented HTTP endpoints directly (both are fine).
- **GA4 on Next 16** — confirm `@next/third-parties` exists/works in this non-standard Next; if not,
  bootstrap gtag manually via `next/script` and emit `page_view` on App Router route changes. No
  consent gate per decision — revisit if EU/GDPR traffic becomes a concern.
