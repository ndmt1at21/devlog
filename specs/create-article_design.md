# Create Article — technical design

Authenticated authors with the IAM permission `articles:create` can publish a new
article from the web UI, writing the body either as **Markdown (README-style)** or in a
**rich-text block editor**. Both formats normalize to the existing `[]domain.Block`
model, so rendering, SEO, ads, search and the paywall pipeline are unchanged and no DB
migration is needed.

## Authorization (IAM)

- Permission name follows IAM's `resource:action` convention: **`articles:create`**.
  Provision it in the `devnote` tenant (resource `articles`, action `create`), bind it
  to a role (e.g. `author`), and assign the role to writers.
- The Go backend (BFF) asks IAM's PDP: `POST {issuer}/authz/decision` with
  `{ token: <session access token>, required_permissions: ["articles:create"] }`.
  `authn.Provider` gains `CheckPermissions(ctx, accessToken, perms)` implemented by the
  IAM client.
- Enforcement is server-side on `POST /api/v1/articles`:
  - no session → `401` (code 1002); denied by PDP → `403` (new code 3005);
  - IAM unreachable → `502` (2002). Fail closed.
- UX hint only: the session cookie stores a `canWrite` snapshot (evaluated at
  login/callback and on token refresh) surfaced via `/auth/me`, so the UI can show or
  hide the editor entry point. The POST check remains authoritative.

## API

`POST /api/v1/articles` (auth + permission required)

```json
{
  "title": "…", "excerpt": "…?", "category": "…", "tags": ["go"],
  "format": "markdown" | "blocks",
  "content": "# markdown source (format=markdown)",
  "body": [ { "type": "p|h|quote|code|diagram|list", … } ]  // format=blocks
}
```

- `markdown` → converted server-side (`internal/content`): `#`/`##…` → `h`, fenced
  ``` → `code` (with lang), `>` → `quote`, `-`/`1.` lists → new `list` block,
  paragraphs → `p`. Inline spans (`**b**`, `*i*`, `` `c` ``, `[t](url)`) stay in text.
- `blocks` → validated against a strict allowlist DTO (unknown fields rejected via
  `DisallowUnknownFields`; no `html` field is accepted from clients).
- Server derives: slug (Vietnamese-aware slugify + uniqueness suffix), author (session
  name), publishedAt (now), read time (~200 wpm), `ord` = max+1.
- Limits: title ≤ 300, excerpt ≤ 500 (derived from first paragraph when empty),
  category ≤ 80, ≤ 8 tags × 40 chars, ≤ 200 blocks, ≤ 20k chars per block; 1 MiB body
  cap already enforced by `decodeJSON`.
- Response: `201` with the article detail (frontend redirects to `/articles/{slug}`).

## Frontend

- `/articles/new`: client editor page (server wrapper exports `robots: noindex`).
  Tabs (Markdown | Rich text) + preview. Rich text = structured block editor
  (paragraph/heading/quote/code/diagram/list) with an inline-format toolbar that inserts
  markdown spans — no `contentEditable`, no raw HTML anywhere.
- Rendering: `p`/`quote`/`list` text runs through a safe inline-markdown renderer that
  emits React elements (never `dangerouslySetInnerHTML`); link URLs restricted to
  http(s)/mailto/relative. New `list` block renders `<ul>/<ol>`.
- Entry point: "Viết bài" in the account menu, shown when `me.user.canWrite`.
- Accessibility: labelled inputs, `tablist/tab/tabpanel` with `aria-selected`,
  `role="alert"` errors, `aria-label` on icon-only controls.

## Security notes

- AuthZ enforced server-side via IAM PDP; UI flag is cosmetic.
- No user HTML is stored or rendered; code blocks are escaped by Shiki at SSR.
- Parameterized INSERT; slug charset `[a-z0-9-]`; strict input DTO + size limits.
- Audit: creations logged with trace id, sub, and slug.
