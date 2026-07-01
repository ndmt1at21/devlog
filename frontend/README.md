# Frontend (Next.js)

The **devnote** blog UI — Next.js 16 (App Router) + React 19 + Tailwind v4,
ported from the `ui-design` mockup. It talks to the Go backend as a same-origin
BFF: the browser only ever hits Next on `:3000`, which rewrites `/api/*` to the
Go API so httpOnly session cookies stay first-party.

## Requirements

- Node 20+
- The Go backend running on `:8080` (see `../backend`). With the default
  in-memory store it needs zero infrastructure.

## Getting started

```bash
npm install
cp .env.example .env.local     # optional — sensible dev defaults are built in
npm run dev                    # http://localhost:3000
```

Start the backend in another terminal first:

```bash
cd ../backend && make run       # DB_DRIVER=memory by default
```

## Scripts

| Command         | Description                                  |
| --------------- | -------------------------------------------- |
| `npm run dev`   | Dev server (Turbopack)                       |
| `npm run build` | Production build                             |
| `npm start`     | Serve the production build                   |
| `npm run lint`  | ESLint (`eslint .` — Next 16 removed `next lint`) |

## Environment

| Var                     | Purpose                                                    |
| ----------------------- | ---------------------------------------------------------- |
| `BACKEND_INTERNAL_URL`        | Go API origin for SSR fetches + `/api/*` rewrite (`:8080`)     |
| `NEXT_PUBLIC_SITE_URL`        | Public origin for metadata + sitemap                          |
| `NEXT_PUBLIC_GA_ID`           | GA4 measurement id — omit to disable analytics                |
| `NEXT_PUBLIC_GAM_NETWORK_CODE`| Google Ad Manager network code — omit to keep the ad placeholder |
| `NEXT_PUBLIC_GAM_AD_UNIT`     | In-content ad-unit name (default `devnote_in_content`)        |

## Structure

```
src/
  app/                     # routes: /, /articles/[slug], /login, /register,
                           #         /forgot-password, /pro, /coffee/result
                           #         + sitemap.ts, robots.ts, not-found, error, loading
  components/
    layout/                # Header, Footer, AccountMenu, ScrollProgress
    theme/                 # ThemeProvider (no-FOUC), ThemeToggle
    home/                  # FeaturedCard, ArticleCard, CategoryFilter, HomeContent
    article/               # ArticleView, blocks/*, SeriesBox, Paywall, SeriesNav, Comments
    coffee/                # CoffeeModal (amount → pay → done)
    pro/                   # ProContent (pricing + subscribe)
    auth/                  # LoginForm, RegisterForm, ForgotForm
    analytics/             # ScrollDepthTracker
    ads/                   # GoogleAdManager (GPT loader), GamAdSlot
  lib/                     # api (client), server-api (SSR), auth, search, types,
                           # analytics, format
  hooks/                   # useImpression (ad_impression)
```

## Notes

- **Data & rendering** — Home and Article pages are Server Components that fetch
  the Go API and forward the session cookie (so the paywall is applied per-user);
  interactive parts (search, theme, comments, coffee, code-copy) are Client
  Components.
- **Auth** requires the backend's IAM integration to be configured
  (`IAM_ISSUER_URL` etc.). Without it, content/search/paywall/coffee-demo all work,
  and auth endpoints return a "not configured" message by design. **Đăng nhập với
  Google** is a federated (redirect) flow through IAM — "Continue with Google"
  navigates to `/api/auth/google/login`; when IAM/Google aren't configured it
  bounces back to `/login?error=…` with an inline message. See the backend
  `.env.example` for the IAM provisioning it needs.
- **Theming** is token-based: the mockup's light/dark palette lives in
  `globals.css` and is mapped into Tailwind utilities via `@theme inline`, so
  utilities re-resolve when `data-theme` flips.
```
