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

| Secret          | Required | Description                                                             |
| --------------- | -------- | ----------------------------------------------------------------------- |
| `VPS_HOST`      | ✅       | VPS IP or hostname.                                                      |
| `VPS_USER`      | ✅       | SSH user with permission to run `docker` (e.g. `deploy`).               |
| `VPS_SSH_KEY`   | ✅       | **Private** SSH key (full PEM) whose public half is on the VPS.         |
| `VPS_PORT`      | ➖       | SSH port. Defaults to `22`.                                             |
| `DEPLOY_PATH`   | ✅       | Deploy directory on the VPS, e.g. `/opt/devlog-backend`.                |
| `HOST_PORT`     | ➖       | Host port the container publishes. Defaults to `8080`.                  |

The `DB_DSN` (with the DB password) is **not** a GitHub secret — it lives in
`backend.env` on the VPS alongside the other secrets. See the format below.

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

> All backend secrets (`DB_DSN`, `SESSION_SECRET`, Stripe/MoMo keys) live in
> `backend.env` on the VPS. They can be moved to GitHub secrets if you prefer —
> ask if you want that.

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

# 3. Runtime env — fill in DB_DSN (see the DSN host guidance above), SESSION_SECRET, etc.
#    Grab the template from the repo, or scp it up:
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
