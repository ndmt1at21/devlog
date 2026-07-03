# Backend deployment

Automated CI/CD for the Go backend: **push to `main` → build Docker image →
push to GHCR → deploy to the VPS**.

```
push to main ─▶ [Deploy workflow]
                  test (go vet + go test -race)
                  └▶ build & push image  ghcr.io/ndmt1at21/devlog-backend:{latest, sha-xxxxxxx}
                       └▶ ssh to VPS: docker compose pull + up -d  ──▶ /api/v1/health smoke test
```

Pull requests and feature branches run the **CI** workflow (build + vet + test)
only; nothing is deployed until it lands on `main`.

The deploy runs only the backend container and connects to MySQL via `DB_DSN`
(in `backend.env`). The DB can be **anything reachable from the VPS** — a MySQL
container, native MySQL on the host, or an external/managed server — so only the
DSN host changes (see below). Schema migrations are embedded in the binary and
applied automatically on startup.

---

## 1. Required GitHub repository secrets

Add these under **Settings → Secrets and variables → Actions → New repository secret**:

| Secret             | Required | Description                                                          |
| ------------------ | -------- | -------------------------------------------------------------------- |
| `VPS_HOST`         | ✅       | VPS IP or hostname.                                                   |
| `VPS_USER`         | ✅       | SSH user with permission to run `docker` (e.g. `deploy`).            |
| `VPS_SSH_KEY`      | ✅       | **Private** SSH key (full PEM) whose public half is on the VPS.      |
| `VPS_PORT`         | ➖       | SSH port. Defaults to `22`.                                          |
| `DEPLOY_PATH`      | ✅       | Deploy directory on the VPS, e.g. `/opt/devlog-backend`.             |
| `HOST_PORT`        | ➖       | Host port the container publishes. Defaults to `8080`.               |
| `BACKEND_ENV_FILE` | ➖       | Full content of `backend.env` (see `backend.env.example`). When set, every deploy rewrites `backend.env` on the VPS (0600) from this secret — single source of truth in GitHub. When unset, the file on the VPS is left as-is (manual management). |

With `BACKEND_ENV_FILE` set, `DB_DSN` and the rest of the backend settings live
in GitHub; the DSN format below still applies.

```
devlog:PASSWORD@tcp(HOST:3306)/devlog?parseTime=true&loc=UTC
```

`parseTime=true` & `loc=UTC` are required; add `tls=true` if MySQL enforces TLS.
The password is used literally (do **not** URL-encode); only `/` breaks the DSN,
so avoid it in the password. `HOST` depends on where MySQL lives — the backend
runs in a container, so **never** `127.0.0.1`:

- **MySQL on this VPS** (a container publishing 3306, or native on the host):
  use `host.docker.internal` — the compose file maps it to the host via an
  `extra_hosts: host-gateway` entry, so no shared network is needed.
- **External / managed MySQL:** use its real hostname or IP.

No registry credentials are needed — the workflow authenticates to GHCR with the
built-in `GITHUB_TOKEN` and passes it over SSH so the VPS can pull the (private)
image during the deploy.

> Recommended: paste the filled-in `backend.env` into the `BACKEND_ENV_FILE`
> secret (PROD environment) so all backend settings (`DB_DSN`,
> `SESSION_SECRET`, Stripe/MoMo keys, `S3_*`/`IMAGE_BASE_URL`) are managed from
> GitHub — editing the secret and re-running the Deploy workflow rolls out the
> change. Without the secret, `backend.env` is maintained by hand on the VPS.

### Frontend secrets (Deploy frontend workflow)

The frontend deploys to Cloudflare Workers via `.github/workflows/deploy-frontend.yml`
on every push to `main` touching `frontend/`. Its configuration also lives in
the `PROD` environment secrets:

| Secret                  | Required | Description                                                       |
| ----------------------- | -------- | ------------------------------------------------------------------ |
| `CLOUDFLARE_API_TOKEN`  | ✅       | API token with **Workers Scripts: Edit** (dash → My Profile → API Tokens). |
| `CLOUDFLARE_ACCOUNT_ID` | ✅       | Account id (dash → Workers & Pages → right sidebar).               |
| `BACKEND_INTERNAL_URL`  | ✅       | Public **HTTPS** origin of the Go backend (baked into the `/api/*` rewrite at build time). |
| `NEXT_PUBLIC_SITE_URL`  | ✅       | Public origin of the site (canonical/OG/sitemap).                  |
| `NEXT_PUBLIC_GA_ID` and other `NEXT_PUBLIC_*` | ➖ | Optional feature config (see `frontend/.env.example`); unset = feature off / default. |

Build-time values are inlined into the bundle, so changing one means editing
the secret and re-running the workflow. The **runtime** twin of
`BACKEND_INTERNAL_URL` is still read from `frontend/wrangler.jsonc` (`vars`) —
keep that committed value identical to the secret.

---

## 2. One-time VPS setup

On the VPS, as the `VPS_USER`:

```bash
# 1. Docker + compose plugin (skip if already installed)
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker "$USER"   # then log out/in so the group takes effect

# 2. Deploy directory (must match the DEPLOY_PATH secret)
sudo mkdir -p /opt/devlog-backend
sudo chown "$USER":"$USER" /opt/devlog-backend
cd /opt/devlog-backend

# 3. Runtime env — SKIP when the BACKEND_ENV_FILE secret is set (the deploy
#    workflow writes backend.env from it on every run). Manual alternative:
#    fill in DB_DSN (see the DSN host guidance above), SESSION_SECRET, etc.
#    curl -fsSL https://raw.githubusercontent.com/ndmt1at21/devlog/main/deployments/backend.env.example -o backend.env
nano backend.env
chmod 600 backend.env
```

The `docker-compose.prod.yml` file is copied up automatically by the workflow on
every deploy — you do **not** need to place it manually.

Add the SSH key pair:

```bash
# On your machine: generate a deploy key (no passphrase) and authorize it
ssh-keygen -t ed25519 -f devlog_deploy -C "github-actions-deploy" -N ""
ssh-copy-id -i devlog_deploy.pub "$VPS_USER@$VPS_HOST"
# Put the PRIVATE key (devlog_deploy) into the VPS_SSH_KEY secret.
```

Make sure the backend can reach MySQL at the `DB_DSN` host and that a `devlog`
database + user exist (the app creates the tables itself).

---

## 3. Deploy

- **Automatic:** merge/push to `main` (touching `backend/**`). The workflow tests,
  builds, pushes, deploys, and fails loudly if `/api/v1/health` doesn't come up.
- **Manual:** **Actions → Deploy → Run workflow** (`workflow_dispatch`).

Each build is tagged `latest` **and** `sha-<short-sha>` for traceability and
rollback.

---

## 4. Rollback

Redeploy a previous image by its short SHA tag on the VPS:

```bash
cd /opt/devlog-backend
# log in first if the image isn't cached locally (GHCR package is private):
#   echo <GHCR_PAT_with_read:packages> | docker login ghcr.io -u <github-user> --password-stdin
# DB_DSN is read from backend.env, so no need to export it for a manual run:
TAG=sha-abc1234 docker compose -f docker-compose.prod.yml up -d
```

Find available tags under the repo's **Packages** (`devlog-backend`).

---

## 5. HTTPS / reverse proxy

The container publishes plain HTTP on `HOST_PORT` (default `8080`). Put a reverse
proxy (Caddy, nginx, Traefik) in front to terminate TLS, and keep
`COOKIE_SECURE=true` in `backend.env`. Example Caddy one-liner:

```
your-domain.com {
    reverse_proxy 127.0.0.1:8080
}
```

To keep the backend off the public internet entirely, set the compose port
mapping to `127.0.0.1:${HOST_PORT:-8080}:8080` and proxy locally.

---

## 6. Image uploads (Cloudflare R2)

Article images live in an S3-compatible bucket and are served through a CDN —
the backend only presigns direct browser→bucket PUTs (`POST /api/v1/uploads`),
so image bytes never touch the VPS. Leave the `S3_*`/`IMAGE_BASE_URL` vars
blank to disable the feature.

### R2 setup (one-time)

1. **Create the bucket** — Cloudflare dashboard → **R2** → *Create bucket*
   (e.g. `devlog-images`). The free tier (10 GB storage, zero egress fees)
   comfortably covers a blog.
2. **Public access** — in the bucket's **Settings → Public access**, connect a
   **custom domain** on your zone (e.g. `img.your-domain`). This serves objects
   through Cloudflare's CDN with edge caching. (An `r2.dev` URL also works for
   testing, but it is rate-limited and not cached.)
3. **API token** — R2 → **Manage R2 API Tokens** → *Create API Token* with
   **Object Read & Write**, scoped to just this bucket. Note the
   *Access Key ID* / *Secret Access Key* pair.
4. **CORS** — the browser PUTs directly to the bucket, so the bucket must allow
   your frontend origin(s). Bucket **Settings → CORS policy**:

   ```json
   [
     {
       "AllowedOrigins": ["https://your-frontend-domain", "http://localhost:3000"],
       "AllowedMethods": ["PUT"],
       "AllowedHeaders": ["content-type"],
       "MaxAgeSeconds": 3600
     }
   ]
   ```

5. **Env** — fill in `backend.env`:

   ```
   S3_ENDPOINT=https://<account_id>.r2.cloudflarestorage.com
   S3_BUCKET=devlog-images
   S3_REGION=auto
   S3_ACCESS_KEY_ID=…
   S3_SECRET_ACCESS_KEY=…
   IMAGE_BASE_URL=https://img.your-domain
   ```

`IMAGE_BASE_URL` is also a policy: the backend rejects article bodies embedding
images from any other origin, so all article images go through the upload flow.

### Local development (MinIO)

Any S3-compatible store works. With Docker:

```sh
docker run -d --name minio -p 9000:9000 -p 9001:9001 \
  -e MINIO_ROOT_USER=devlog -e MINIO_ROOT_PASSWORD=devlog-secret \
  quay.io/minio/minio server /data --console-address ":9001"
docker run --rm --net=host --entrypoint /bin/sh quay.io/minio/mc -c '
  mc alias set m http://localhost:9000 devlog devlog-secret &&
  mc mb m/devlog-images && mc anonymous set download m/devlog-images'
```

Then run the backend with:

```
S3_ENDPOINT=http://localhost:9000  S3_BUCKET=devlog-images  S3_REGION=us-east-1
S3_ACCESS_KEY_ID=devlog  S3_SECRET_ACCESS_KEY=devlog-secret
IMAGE_BASE_URL=http://localhost:9000/devlog-images
```

(Plain-http image URLs are accepted for localhost only.)
