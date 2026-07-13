# Series plan — "Designing a Multi-Tenant IAM Service in Go"

> Planning artifact (not a post). Produced by the blog-SKILL series-design stage.
> Source of truth for the codebase: [`iam/ARCHITECTURE.md`](https://github.com/ndmt1at21/iam),
> `iam/README.md`, and `iam/internal/**`. Approve/edit this before any drafting.

---

## 1. Series overview

| | |
|---|---|
| **Series slug** (`series.slug`) | `designing-a-multi-tenant-iam-service` |
| **Series title — vi** | Thiết kế một IAM service đa tenant với Go |
| **Series title — en** | Designing a Multi-Tenant IAM Service in Go |
| **Parts** | 7 core + 1 optional capstone |
| **Language** | Bilingual — `*.vi.md` (primary) + `*.en.md` (parallel, full parity) |
| **Length/part** | 2,000–2,500 words |
| **Reading level** | Backend / platform engineers, mid → senior |

**Series description** (`series.description`, ≤160 chars):
- vi: "Xây một OAuth2/OIDC provider đa tenant thật bằng Go: tenancy, grant registry, token rotation, RBAC động, federation, và PDP ở API gateway."
- en: "Build a real multi-tenant OAuth2/OIDC provider in Go — tenancy, grant registry, token rotation, dynamic RBAC, federation, and a gateway PDP."

**Audience & intent.** Engineers who have *consumed* OAuth2 (Google login, Auth0, Keycloak) but never *built* a provider, plus teams sizing up "buy vs. build." They can read Go but the lessons are language-agnostic.

**Content gap (the angle).** Public tutorials overwhelmingly build a **single-tenant toy** — "add a JWT to your API." Almost none design a **production multi-tenant OAuth2/OIDC provider**: signing-key rotation, refresh-token reuse detection, runtime-defined permissions, and externalized authorization at the gateway. This series uses one real codebase as the worked example and, for every decision, states **the alternative it rejected and why** — the part tutorials skip.

**Voice.** First-person devlog ("mình / I"), opinionated but sourced. Same register as the existing ScyllaDB post: answer-first, concrete numbers, no hype, honest tradeoffs. Author `ndmt1at21`.

---

## 2. Series-level SEO & internal-link map

**Head terms the series should own:** *multi-tenant IAM*, *build your own OAuth2 server*, *OIDC provider design*, *refresh token rotation*, *RBAC design*, *policy decision point*.

**Pillar–spoke linking (build these as `[INTERNAL-LINK]` zones):**

```
                 ┌─────────────────────────────┐
                 │  Part 1 — PILLAR (overview) │
                 └──────────────┬──────────────┘
   ┌───────────┬───────────┬───┴───────┬───────────┬───────────┐
  P2          P3          P4          P5          P6          P7
tenancy   grant reg    tokens       RBAC     federation    PDP
   └────────── each spoke links UP to P1 and ACROSS to prev/next ──────────┘
```

- Every spoke opens by linking back to the pillar and closes by linking to the next part (`nextPart`).
- The pillar's body carries one `[INTERNAL-LINK]` per spoke (7 total) — these become the series' internal-link spine.
- Cross-links that are natural: P4 (tokens) ↔ P3 (grants, since grants *mint* tokens); P5 (RBAC) → P7 (PDP consumes the permission claims); P6 (federation) → P3 (federated grant).

**hreflang / alternate:** each part's two files reference each other via `alternate` (already the repo convention) — SEO Stage 4 adds reciprocal links.

---

## 3. Frontmatter template (bilingual + series fields)

Extends the existing post frontmatter (see `posts/scylladb-la-gi.vi.md`) with the four series fields that map to the devlog schema (`articles.series_slug`, `articles.series_part`, `articles.part_title`, and the `series` table). `coverAlt` is **mandatory** (accessibility + SEO); `coverGen` holds the shared image-gen prompt.

**`iam-...-P1.vi.md`**
```yaml
---
title: "Tự viết OAuth2/OIDC provider đa tenant: bắt đầu từ đâu?"
description: "…150–160 chars, chứa 1 số liệu có nguồn…"
slug: "tu-viet-oauth2-oidc-provider-da-tenant"
lang: "vi"
alternate: "build-multi-tenant-oauth2-provider.en.md"
author: "ndmt1at21"
authorBio: "Backend engineer, viết devlog về distributed systems, database và identity."
date: "2026-07-11"
lastUpdated: "2026-07-11"
tags: ["iam", "oauth2", "oidc", "golang", "multi-tenant", "authentication"]
series: "designing-a-multi-tenant-iam-service"
seriesTitle: "Thiết kế một IAM service đa tenant với Go"
seriesDescription: "Xây một OAuth2/OIDC provider đa tenant thật bằng Go…"
part: 1
partTitle: "Kiến trúc tổng thể"
cover: "https://images.unsplash.com/…?w=1200&h=630&fit=crop&q=80"
coverAlt: "…câu mô tả rõ ảnh bìa cho screen reader…"
coverGen: "<shared series style block> … 1200x630, no text, no words, no logos"
---
```

The `.en.md` file is identical in shape with `lang: "en"`, `alternate:` pointing back to the `.vi.md`, English `title`/`description`/`partTitle`, and the same `series`, `part`, `date`, `cover` (image + `coverAlt` translated).

**Shared image style block** (put in every `coverGen`/`[IMAGE]` gen prompt for visual consistency across all 8 parts):
> "isometric flat vector illustration, dark navy background, cyan and orange accents, clean geometric lines, no gradients, [ratio], no text, no words, no logos"

Recurring visual motif for the series: a **hexagon = a tenant**, a **key = a token/signing key**, a **shield/gate = the PDP**. Reuse across covers so the series reads as one set.

---

## 4. Per-post design

Each spoke follows the blog-SKILL skeleton: crafted title (not "What is X?"), answer-first opener anchored to a **sourced** security stat, Key-Takeaways box, 6–8 H2s (60–70% phrased as questions), a citation capsule + a concrete `dẫn chứng` (code/schema/interface shown **inline**, not linked to the repo) per major H2, 2–3 information-gain markers, an `[IMAGE]`/`[CHART]`/`[CALLOUT]` every 300–500 words, a 3–5 Q FAQ, and 3–6 external tier-1–3 references. Stats below name the **source to verify in Stage 1** — no numbers are invented here.

**Two series-wide conventions (mandatory in every part):**

1. **Reason-first ("vì sao cần?" / "why do we need it?").** Before *any* how, each part answers *why this piece exists and what breaks without it*. Concretely: the answer-first opener frames the pain, and the **first substantive H2 is a "Vì sao cần…?" section** that names the failure the design prevents. Every subsequent design decision is introduced as *problem → why the obvious approach hurts → what we chose*. Goal: the reader remembers the **reason**, not just the mechanism.
2. **A Mermaid design diagram.** Every part carries **≥1 `mermaid` fenced code block** (flowchart / sequence / stateDiagram) that a reader can trace to follow the design end-to-end. This is a real diagram in the article body, distinct from illustrative `[IMAGE]` markers. The specific diagram is named in each part's **Mermaid** line below.
   - ✅ **Rendering (added 2026-07-11):** the devlog frontend now renders Mermaid. A ` ```mermaid ` fence still parses to a `code` block (`backend/internal/content/content.go`), but `BlockRenderer.tsx` routes `lang === "mermaid"` to a new client-only `frontend/src/components/article/blocks/Mermaid.tsx` (dynamic-imports mermaid@11, theme-aware via `useTheme`, falls back to raw source on error); the Shiki SSR pass in `articles/[slug]/page.tsx` skips mermaid. So the same ` ```mermaid ` blocks in these drafts render as diagrams on GitHub/IDE preview **and** on the published site.

Locked for this series: **code is shown inline only** (no `github.com/...` deep-links); **non-series `[INTERNAL-LINK]` targets stay as placeholders** (no sibling IAM posts exist yet); **all parts dated `2026-07-11`**; **slugs are frozen** (do not rename post-publish).

---

### Part 1 — Pillar / Overview
- **Title vi:** Tự viết OAuth2/OIDC provider đa tenant: bắt đầu từ đâu?
- **Title en:** Build a Multi-Tenant OAuth2 Provider: The Architecture
- **Slugs:** `tu-viet-oauth2-oidc-provider-da-tenant` / `build-multi-tenant-oauth2-provider`
- **Primary keyword:** multi-tenant IAM / OAuth2 provider design · **Intent:** informational + "should I build this?"
- **Reader takeaway:** the map of the whole system and *why* each piece exists before diving into any one.
- **Opener stat to source:** share of breaches involving stolen/compromised credentials → **Verizon DBIR (latest)**; OWASP Top 10 A01 Broken Access Control ranking.
- **H2 outline** *(ordered as a gentle on-ramp: establish IAM's importance first, defer tenancy depth to Part 2 — per reader feedback 2026-07-11)*
  1. Vì sao IAM là chỗ đáng đầu tư nhất? *(reason-first at the concept level: authN + authZ = the front door; anchor the breach stats here)*
  2. IAM này thực sự làm những gì? *(scope: OAuth2 AS + OIDC OP + RBAC + federation + passwordless + PDP — before build-vs-buy so terms are grounded)*
  3. Vì sao tự viết thay vì dùng Keycloak/Auth0? *(decision: build; rejected: managed — names the 3 multi-tenant needs but defers detail to later parts)*
  4. Ba nhóm API: Admin, OIDC/OAuth2, Auth — `dẫn chứng`: the 3 handler groups (`/api/v1`, `/oauth2`, `/auth`).
  4. Domain model trong một sơ đồ — `dẫn chứng`: `internal/domain/*.go` entities (Tenant→User/Role/Client/…).
  5. Vì sao chọn hexagonal + dual-backend? *(rejected: coupling SQL into services; `internal/storage` factory picks pg|mysql)*
  6. Stateless service scale ngang thế nào? *(state in DB/cache only)*
  7. Lộ trình series: 6 phần còn lại (7 `[INTERNAL-LINK]` xuống spokes).
- **Info-gain:** `[UNIQUE INSIGHT]` build-vs-buy decision table; `[ORIGINAL DATA]` LOC / package count from the repo.
- **Mermaid (design):** `flowchart` — HTTP clients → 3 API surfaces → service layer → adapters (pg/mysql/cache/IdP); the whole-system map readers keep in mind for the series.
- **Visuals:** `[IMAGE]` cover (hex mesh); `[CHART]` build-vs-buy decision table.
- **FAQ seeds:** "Keycloak có làm được multi-tenant không?" · "Có cần OIDC nếu chỉ làm login nội bộ?" · "Postgres hay MySQL?"
- **External refs:** RFC 6749, OpenID Connect Core, OWASP Top 10, Verizon DBIR.

### Part 2 — Multi-tenancy
- **Title vi:** Đa tenant bằng một cột tenant_id: được gì, mất gì?
- **Title en:** One `tenant_id` Column: Multi-Tenancy Tradeoffs
- **Slugs:** `da-tenant-bang-cot-tenant-id` / `multi-tenant-by-column-tenant-id`
- **Primary keyword:** multi-tenant database design / tenant isolation
- **Reader takeaway:** when shared-DB-by-column beats schema/DB-per-tenant — and the one invariant that makes it safe.
- **Opener stat to source:** SaaS tenant-isolation incidents / IDOR-style cross-tenant leaks (OWASP API Security Top 10 — BOLA/API1).
- **H2 outline**
  1. Ba kiểu multi-tenancy và cái giá của chúng *(column vs schema-per-tenant vs db-per-tenant)*
  2. Vì sao chọn shared DB + `tenant_id`? *(rejected: db-per-tenant — migration/ops blow-up)*
  3. Làm sao đảm bảo không rò rỉ chéo tenant? — `dẫn chứng`: every repo method takes `tenantID` first arg; invariant "every tenant-owned row carries `tenant_id`."
  4. Vì sao trả `ErrNotFound` thay vì `ErrForbidden`? *(anti-enumeration — the isolation invariant)*
  5. Tenant được resolve từ request thế nào? — `dẫn chứng`: `ResolveTenant` 3-stage fallback (trusted header → static domain map → `/t/{slug}`).
  6. Mỗi tenant một OIDC issuer nghĩa là gì? — per-tenant `.well-known/openid-configuration` + JWKS.
  7. `metadata jsonb` — thoát hiểm cho field tùy biến.
- **Info-gain:** `[UNIQUE INSIGHT]` the `ErrNotFound`-over-`ErrForbidden` rule; `[PERSONAL EXPERIENCE]` a cross-tenant bug caught by the invariant.
- **Mermaid (design):** `flowchart` — the 3-stage `ResolveTenant` fallback (trusted header → static domain map → `/t/{slug}`) ending at `Tenant{Slug,ID,Issuer}` on the context.
- **Visuals:** `[CHART]` 3-way isolation tradeoff table (column vs schema vs db-per-tenant).
- **FAQ:** "Column-based có đủ an toàn cho compliance?" · "RLS của Postgres thì sao?" · "Đổi tenant sang db riêng sau này có khó?"
- **External refs:** OWASP API Security Top 10 (BOLA), Postgres RLS docs (as the rejected alternative), a SaaS multi-tenancy reference (e.g. AWS SaaS lens).

### Part 3 — OAuth2 authorization server & grant registry
- **Title vi:** Grant registry: thêm kiểu đăng nhập không đụng endpoint token
- **Title en:** A Grant Registry for a Pluggable `/token` Endpoint
- **Slugs:** `oauth2-grant-registry` / `oauth2-grant-registry-design`
- **Primary keyword:** OAuth2 authorization server design / grant types
- **Reader takeaway:** model `/token` as a dispatcher over a small `Grant` interface so new auth methods are additive.
- **Opener stat to source:** OAuth2 ubiquity / adoption; PKCE now-required guidance from **OAuth 2.0 Security BCP (RFC 9700)**.
- **H2 outline**
  1. `/authorize` phải kiểm tra những gì? *(client, redirect_uri exact match, PKCE, scopes)*
  2. Vì sao `/token` nên là một dispatcher? *(rejected: one giant switch per grant)*
  3. Interface `Grant` trông ra sao? — `dẫn chứng`: `grant.Grant.Handle(ctx, *Request) (*TokenResponse, error)` + `Registry.Register(g Grant)`.
  4. Bảy grant đang hỗ trợ — `dẫn chứng`: `authorization_code`, `refresh_token`, `client_credentials`, `password`, `passwordless`, `federated`, `token_exchange` (files in `internal/auth/grant/`).
  5. PKCE cho public client: khi nào bắt buộc? *(enforced for public; optional for confidential)*
  6. Per-client gate: tắt/bật từng grant — `dẫn chứng`: `AuthMethodsConfig`, the `login_disabled` short-circuit (`unauthorized_client`).
  7. `token_exchange` (RFC 8693): khi service cần đổi token.
- **Info-gain:** `[ORIGINAL DATA]` the `Grant` interface diff that adds a new grant in ~1 file; `[UNIQUE INSIGHT]` why the registry beats a switch for testing (see `security_test.go`).
- **Mermaid (design):** `sequenceDiagram` — `/authorize` validation → browser login → `/token` dispatch through `Registry` to the matching `Grant.Handle` → `TokenResponse`.
- **Visuals:** `[CHART]` grant table (grant → who uses it → token out).
- **FAQ:** "Password grant còn nên dùng không?" · "PKCE có cần cho confidential client?" · "Thêm grant mới mất bao lâu?"
- **External refs:** RFC 6749, RFC 7636 (PKCE), RFC 8693 (token exchange), RFC 9700 (Security BCP), RFC 8252 (native apps).

### Part 4 — Token lifecycle
- **Title vi:** Refresh token rotation và chain revocation: vòng đời token
- **Title en:** Refresh Token Rotation and Reuse Detection, By Design
- **Slugs:** `refresh-token-rotation-chain-revocation` / `refresh-token-rotation-reuse-detection`
- **Primary keyword:** refresh token rotation / JWT access token / signing-key rotation
- **Reader takeaway:** short JWT access + opaque rotating refresh, with reuse detection that revokes the whole family, and key rotation with zero downtime.
- **Opener stat to source:** token-theft / replay in OAuth incidents; RFC 9700's explicit refresh-rotation + reuse-detection recommendation.
- **H2 outline**
  1. JWT cho access, opaque cho refresh — vì sao khác nhau? *(stateless verify vs revocable)*
  2. Lưu token thế nào cho an toàn? — `dẫn chứng`: auth codes & refresh tokens persisted **by SHA-256 hash only**; plaintext never re-derivable.
  3. Rotation là gì và vì sao mỗi lần dùng đổi token? *(rejected: long-lived static refresh)*
  4. Chain/family revocation: dùng lại token đã thu hồi thì sao? — `dẫn chứng`: reuse of a revoked token revokes all tokens from the same authorization.
  5. Xoay signing key mà không rớt token đang bay — `dẫn chứng`: `SigningKey` states `active/next/retired`; both active+retired published in JWKS during roll-over.
  6. ID token & nonce: chống replay trong code flow.
  7. RS256 hay ES256? *(alg choice, JWKS `kid`)*
- **Info-gain:** `[UNIQUE INSIGHT]` the family-revocation state machine; `[ORIGINAL DATA]` `iam_active_refresh_tokens` gauge over a rotation.
- **Mermaid (design):** `stateDiagram-v2` — signing-key lifecycle `next → active → retired` with JWKS-overlap window; plus a `sequenceDiagram` for refresh rotation + reuse → family revocation.
- **Visuals:** `[CHART]` access-vs-refresh token comparison (format, storage, lifetime, revocable?).
- **FAQ:** "Refresh token nên sống bao lâu?" · "Vì sao không dùng JWT cho refresh?" · "Rotate key có làm user logout?"
- **External refs:** RFC 7519 (JWT), RFC 9700 (Security BCP — rotation/reuse), RFC 7517 (JWK), OAuth.net token best practices.

### Part 5 — RBAC + dynamic permissions
- **Title vi:** RBAC với permission động: cho tenant tự định nghĩa quyền
- **Title en:** RBAC With Runtime, Tenant-Defined Permissions
- **Slugs:** `rbac-permission-dong-da-tenant` / `rbac-dynamic-tenant-permissions`
- **Primary keyword:** RBAC design / dynamic permissions / permission wildcards
- **Reader takeaway:** a permissions→roles→users model where tenants add their own `resource:action` permissions at runtime, resolved into JWT claims.
- **Opener stat to source:** OWASP A01 Broken Access Control (#1 in 2021 Top 10) as the risk this design targets.
- **H2 outline**
  1. Mô hình permissions→roles→users — `dẫn chứng`: `user_roles` / `role_permissions` join tables.
  2. `resource:action` và wildcard hoạt động ra sao? — `dẫn chứng`: `Enforcer.Has(held []string, required string)` with exact / `users:*` / `*`.
  3. System vs custom permission khác gì? — `dẫn chứng`: `Permission.TenantID == nil` → system (catalog); `!= nil` → tenant-custom.
  4. Vì sao cho tenant định nghĩa quyền lúc runtime? *(rejected: hardcoded enum — every tenant's product differs)*
  5. Cô lập custom permission giữa các tenant — `dẫn chứng`: `PermissionService.Delete` checks `*p.TenantID == tenantID`, returns `ErrNotFound` cross-tenant.
  6. System roles seed khi tạo tenant — `dẫn chứng`: `tenant_admin` (all), `member` (read subset).
  7. Permission vào token thế nào? — `dẫn chứng`: `EffectivePermissions` UNION query → JWT `permissions[]` + `/oauth2/userinfo`.
- **Info-gain:** `[UNIQUE INSIGHT]` wildcards let `tenant_admin` = `["*"]` instead of enumerating; `[ORIGINAL DATA]` the effective-permissions SQL.
- **Mermaid (design):** `flowchart` — `User → user_roles → Role → role_permissions → Permission`, then `EffectivePermissions` (UNION) → JWT `permissions[]`; branch showing `Enforcer.Has` exact / `users:*` / `*`.
- **Visuals:** `[CHART]` system-vs-custom permission table (`tenant_id NULL` vs `<uuid>`).
- **FAQ:** "RBAC vs ABAC — khi nào cần ABAC?" · "Permission trong token có bị stale không?" · "Bao nhiêu permission thì token quá to?"
- **External refs:** NIST RBAC model (INCITS 359), OWASP A01, OWASP ASVS access-control section.

### Part 6 — Federation + passwordless
- **Title vi:** Federated login và passwordless chung một luồng code
- **Title en:** Federated Login and Passwordless On One Code Flow
- **Slugs:** `federated-login-passwordless-mot-luong` / `federated-login-passwordless-one-flow`
- **Primary keyword:** federated login design / OIDC social login / passwordless OTP magic link
- **Reader takeaway:** a single IdP-adapter interface plus challenge objects, both resuming the *same* internal authorization-code flow.
- **Opener stat to source:** phishing/credential-stuffing prevalence (DBIR/FIDO) motivating passwordless & social login.
- **H2 outline**
  1. Adapter cho mỗi IdP trông thế nào? — `dẫn chứng`: `identity.Provider` = `Key() / AuthCodeURL(state,nonce) / Exchange(ctx,code,nonce) / Verify(ctx,rawToken)`.
  2. Luồng redirect và state chống CSRF — `dẫn chứng`: `fed:state:{state}` in cache, TTL 10 min, single-use on callback.
  3. Link hay tạo user? — `dẫn chứng`: `UserIdentity(provider, provider_subject)` unique per tenant.
  4. Credentials của IdP lưu ở đâu? — `dẫn chứng`: encrypted at rest with AES-GCM (`IAM_ENCRYPTION_KEY`).
  5. Passwordless: OTP và magic link khác gì? — `dẫn chứng`: `PasswordlessChallenge` w/ TTL + attempt counter (brute-force limit).
  6. Vì sao cả hai đều kết thúc bằng một authorization code nội bộ? *(unify → one token path; rejected: bespoke session per method)*
  7. `federated` grant: client tự cầm id_token thì sao? — `dẫn chứng`: `Provider.Verify`.
- **Info-gain:** `[UNIQUE INSIGHT]` "everything funnels into the code flow" as the unifying trick; `[PERSONAL EXPERIENCE]` an OTP brute-force attempt in metrics.
- **Mermaid (design):** `sequenceDiagram` — federation redirect (login → `fed:state` in cache → IdP → callback → link/create `UserIdentity` → internal auth code) shown side-by-side with the passwordless challenge path converging on the *same* code flow.
- **Visuals:** `[CHART]` OTP vs magic-link vs social comparison (UX, phishing-resistance, setup).
- **FAQ:** "Magic link có an toàn bằng OTP?" · "Nonce để làm gì trong OIDC?" · "Nhiều tài khoản social trỏ về một user được không?"
- **External refs:** OpenID Connect Core, RFC 6749 (state), NIST SP 800-63B (authenticators), FIDO Alliance passwordless data.

### Part 7 — API gateway as Policy Decision Point
- **Title vi:** IAM làm Policy Decision Point cho API gateway
- **Title en:** Your IAM as a Policy Decision Point for the Gateway
- **Slugs:** `iam-policy-decision-point-api-gateway` / `iam-policy-decision-point-gateway`
- **Primary keyword:** policy decision point / API gateway authorization / externalized authz
- **Reader takeaway:** move per-request authz to the gateway (PEP) calling a stateless PDP that holds no route policy.
- **Opener stat to source:** cost/latency of authz sprawl; zero-trust guidance (NIST SP 800-207) for externalized decisions.
- **H2 outline**
  1. PEP và PDP: ai làm gì? *(gateway enforces, IAM decides)*
  2. Vì sao PDP không giữ route→permission policy? *(rejected: policy in IAM — the gateway supplies requirements per route → stateless)*
  3. Decision contract — `dẫn chứng`: `DecisionRequest{token, method, path, required_scopes[], required_permissions[], match_all}` → `DecisionResponse{allow, reason, sub, tenant, client_id, scope, permissions[], exp}`.
  4. `match_all` và `tenant_mismatch`: chi tiết dễ sai.
  5. Reason codes → HTTP status ai map? — `dẫn chứng`: `ok/missing_token/invalid_token/tenant_mismatch/insufficient_scope/insufficient_permission` → 401/403.
  6. Kong `iam-decision` plugin làm gì? — `dẫn chứng`: access-phase plugin extracts bearer, calls PDP, injects `X-User-Id/X-Tenant-Id/…` upstream.
  7. Fail-open hay fail-closed? *(default 503 closed; `fail_open` trades safety for availability — use deliberately)*
- **Info-gain:** `[UNIQUE INSIGHT]` "the endpoint always returns 200; the gateway enforces" as the contract's key idea; `[ORIGINAL DATA]` a real `/authz/decision` request/response.
- **Mermaid (design):** `sequenceDiagram` — Client → Kong (`iam-decision`) → IAM PDP (`POST /authz/decision`) → allow/deny branches (inject identity headers + proxy upstream, vs 401/403, vs 503 on PDP-down).
- **Visuals:** `[CHART]` reason-code → gateway status table (`ok`/`missing_token`/… → 401/403).
- **FAQ:** "PDP một call mỗi request có chậm không?" · "Upstream có nên tin header X-User-Id?" · "OPA/Cedar khác gì cách này?"
- **External refs:** NIST SP 800-207 (Zero Trust), XACML PEP/PDP terminology, Kong plugin docs, OpenID Connect (token verify).

### Part 8 — Capstone: security & operability *(shipping — confirmed)*
- **Title vi:** Secrets-at-rest, rate limit, observability: phần không ai dạy
- **Title en:** Secrets at Rest, Rate Limits, Observability: The Rest
- **Slug:** `iam-security-hardening-observability` (both langs)
- **Primary keyword:** IAM security hardening / OAuth2 production checklist
- **Reader takeaway:** the production concerns that don't fit one feature but sink you if skipped — and *why* each is non-negotiable.
- **Covers:** AES-GCM secrets at rest + bcrypt password hashing; token binding by hash; sliding-window rate limits on `/oauth2/token`, `/oauth2/authorize`, `/auth/passwordless/start`; OTel traces + domain metrics (`iam_token_issued_total`, `iam_login_total`, `iam_grant_errors_total`, `iam_passwordless_otp_total`, `iam_active_refresh_tokens`); the dual-backend (pgx vs `database/sql`) adapter split; UUID v7 ids.
- **Mermaid (design):** `flowchart` — a request's cross-cutting path: rate-limit gate → handler → secrets decrypted (AES-GCM) → OTel span + domain metric emitted → adapter (pgx | database/sql); shows where each concern hooks in.
- **Visuals:** `[CHART]` the production checklist as a table (concern → what breaks without it → where it lives in the code).
- **FAQ:** "bcrypt hay argon2 cho password?" · "Rate limit theo IP hay theo client?" · "Metric nào cần alert?"
- **External refs:** OWASP ASVS, OWASP Cheat Sheet (Password Storage), RFC 6819 (threat model), OpenTelemetry semantic conventions.

---

## 5. Research to run in Stage 1 (shared, before drafting)

Verify each before use; **never invent a number.** Candidate tier-1 sources:

| Needed stat | Source to verify | Used in |
|---|---|---|
| % of breaches involving stolen/compromised credentials | Verizon **DBIR** (latest) | P1, P4, P6 |
| Broken Access Control = #1 web risk | **OWASP Top 10 (2021)** A01 | P1, P5 |
| BOLA / cross-tenant object access | **OWASP API Security Top 10** | P2 |
| Refresh-token rotation + reuse detection is recommended | **RFC 9700** (OAuth 2.0 Security BCP) | P4 |
| PKCE now recommended for all clients | RFC 9700 / RFC 7636 | P3 |
| Phishing-resistant / passwordless direction | **NIST SP 800-63B**, FIDO Alliance | P6 |
| Zero-trust externalized authorization | **NIST SP 800-207** | P7 |

RFC anchors to cite inline (descriptive anchors, not bare URLs): 6749, 6750, 7519, 7636, 8252, 8693, 9700, OIDC Core, plus NIST INCITS 359 (RBAC).

---

## 6. Production & delivery plan

- **File naming:** `posts/<en-slug>.en.md` + `posts/<vi-slug>.vi.md` per part (matches the existing ScyllaDB pair). Cross-link via `alternate`.
- **Draft order:** P1 (pillar) first — it defines the internal-link spine — then P2→P7 in order; P8 last.
- **Per-part pipeline:** research → write (outline approved unless one-shot) → review (≥85/100, zero Critical, clean fact-check, translation-parity) → SEO. The reviewer's `BLOCKING` verdict drives ≤3 fix loops.
- **Consistency to enforce across all parts:** the shared image style block (§3), the hex/key/shield motif, the `authorBio`, tag baseline `["iam","oauth2","oidc","golang","multi-tenant"]` + per-part tags, and the same series frontmatter block.
- **Effort estimate:** ~7–8 parts × 2 languages ≈ 14–16 files, 2–2.5k words each.

---

## 7. Decisions — resolved (2026-07-11)

| Question | Decision |
|---|---|
| Ship Part 8? | **Yes** — shipping as the capstone (8 parts total). |
| Link to non-series posts? | **No sibling posts exist yet** → leave `[INTERNAL-LINK]` placeholders for non-series targets; series (pillar↔spoke) links are concrete. |
| Publish dates | **All `2026-07-11`** (`date` = `lastUpdated`). |
| Repo deep-links | **No** — code/schema shown **inline only**. |
| Slugs | **Frozen** as listed per part — do not rename post-publish. |
| Reason-first framing | **Mandatory** — every part opens with "why we need it" before any how (see §4 convention 1). |
| Mermaid design diagram | **Mandatory** — ≥1 `mermaid` block per part (see §4 convention 2 and each part's **Mermaid** line). |

**Next action:** draft **Part 1 (pillar)** through research → write → check → SEO, then P2→P8 in order.
