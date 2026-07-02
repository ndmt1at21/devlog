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

The database is **external / managed MySQL** — the deploy runs only the backend
container and connects to your existing MySQL via `DB_DSN`. Schema migrations are
embedded in the binary and applied automatically on startup.

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
| `DB_DSN`        | ✅       | Full MySQL DSN **including the password** (see format below).           |
| `HOST_PORT`     | ➖       | Host port the container publishes. Defaults to `8080`.                  |

`DB_DSN` format (store the whole string as the secret value):

```
devlog:PASSWORD@tcp(your-db-host:3306)/devlog?parseTime=true&loc=UTC&multiStatements=true
```

It's injected straight into the container at deploy time and **never written to a
file on the VPS**. No registry credentials are needed — the workflow authenticates
to GHCR with the built-in `GITHUB_TOKEN` and passes it over SSH so the VPS can pull
the (private) image during the deploy.

> Other non-DB secrets (`SESSION_SECRET`, Stripe/MoMo keys) still live in
> `backend.env` on the VPS. They can be moved to GitHub secrets the same way —
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

# 3. Runtime env — fill in SESSION_SECRET, APP_BASE_URL, etc.
#    DB_DSN is NOT put here — it comes from the GitHub `DB_DSN` secret at deploy time.
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

Make sure the VPS can reach your MySQL server and that a `devlog` database + user
exist (the app creates the tables itself).

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
# DB_DSN must be exported for any manual run (compose fails fast without it):
export DB_DSN='devlog:PASSWORD@tcp(your-db-host:3306)/devlog?parseTime=true&loc=UTC&multiStatements=true'
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
