# Speedcubing Hungary

Monorepo:

| Path            | Service  | Stack                         |
| --------------- | -------- | ----------------------------- |
| `apps/web`      | frontend | Next.js 16 / React 19         |
| `apps/api`      | backend  | Go (chi, golang-migrate)      |
| `apps/scripts`  | jobs     | Python weekly report          |
| `db/migrations` | schema   | SQL migrations (schema only)  |

The runtime stack runs on Docker Compose (services: `web`, `api`, `db`).

## Environments are decoupled from Git branches

The runtime environment is **not** chosen by the current branch — it is selected
by which Compose override + env file you pass. The same code runs in any
environment; only the override file and `--env-file` change.

| Environment | Compose files                                              | Env file    |
| ----------- | ---------------------------------------------------------- | ----------- |
| Development | `infra/docker-compose.yml` + `infra/docker-compose.dev.yml` | `.env.dev`  |
| Staging     | `infra/docker-compose.yml` + `infra/docker-compose.test.yml`| `.env.test` |
| Production  | `infra/docker-compose.yml` + `infra/docker-compose.prod.yml`| `.env.prod` |

- `infra/docker-compose.yml` — everything shared (services, env mapping, network, db volume/healthcheck). Builds production image stages by default. Includes a Caddy reverse-proxy service gated behind the `proxy` compose profile.
- `infra/docker-compose.dev.yml` — dev only: bind mounts, hot reload (`dev` build target), exposed DB port, and host ports for web/api (no Caddy in dev — use `http://localhost`).
- `infra/docker-compose.test.yml` / `infra/docker-compose.prod.yml` — immutable built images, no bind mounts.
- `infra/docker/api/Dockerfile` / `infra/docker/web/Dockerfile` — multi-stage Dockerfiles for each service.
- `infra/caddy/Caddyfile` — reverse proxy with automatic HTTPS (Let's Encrypt). In test/prod, web/api are **not** published to the host; only Caddy (80/443) is public and proxies `WEB_DOMAIN` → web and `API_DOMAIN` → api. Enabled by `COMPOSE_PROFILES=proxy` in the env file.

No values are hardcoded in the compose files — they all come from the `--env-file`.

## Running

### Local development (hot reload)

```bash
cp .env.dev.example .env.dev   # first time only, then fill in real values
docker compose \
  -f infra/docker-compose.yml \
  -f infra/docker-compose.dev.yml \
  --env-file .env.dev up
```

Frontend: http://localhost:3000 · API: http://localhost:8080

Source is bind-mounted, so:

- editing a **frontend** file triggers a Next.js hot reload (`next dev --webpack`, polling enabled for Windows/WSL);
- editing a **backend** file triggers an `air` rebuild/restart;
- no `docker compose build` is needed during normal development.

### Staging / test

```bash
cp .env.test.example .env.test   # fill in real values
docker compose \
  -f infra/docker-compose.yml \
  -f infra/docker-compose.test.yml \
  --env-file .env.test up -d --build
```

### Production

```bash
cp .env.prod.example .env.prod   # fill in real values
docker compose \
  -f infra/docker-compose.yml \
  -f infra/docker-compose.prod.yml \
  --env-file .env.prod up -d --build
```

Each environment uses a distinct `COMPOSE_PROJECT_NAME` (in its env file), so
their containers, networks and volumes are isolated even on the same host.

## Branch strategy

| Branch       | Role                | CI / CD                                  |
| ------------ | ------------------- | ---------------------------------------- |
| `main`       | Production           | push → CI, then deploy to production     |
| `test`       | Staging/test server | push → CI, then deploy to staging        |
| `develop`    | Integration / local | push → CI (build, lint, test). No deploy. |
| `feature/*`  | Feature work        | CI on push / PR. No deploy.              |
| `hotfix/*`   | Emergency fixes      | CI on push / PR. No deploy.              |

`main` is the default branch. Promotion is linear and fast-forward:

```bash
# develop -> staging
git switch test && git merge --ff-only develop && git push origin test
# staging -> production
git switch main && git merge --ff-only test   && git push origin main
git switch develop
```

## CI/CD

Workflows in `.github/workflows/`:

- `ci.yml` — build + lint + test (Go and Next.js). Runs on `develop`/`feature/*`/`hotfix/*` and PRs, and is reused by the deploy workflows.
- `deploy-test.yml` — on push to `test`: run CI, then SSH-deploy with the `test` Compose files. Uses the `test` GitHub Environment.
- `deploy-prod.yml` — on push to `main`: run CI, then SSH-deploy with the `prod` Compose files. Uses the `production` GitHub Environment.

Each deploy Environment needs secrets `SSH_HOST`, `SSH_USER`, `SSH_KEY`,
`SSH_PORT` and variables `DEPLOY_PATH`. Add **Required reviewers** to the
`production` environment to gate prod deploys behind manual approval.

## Server setup (one-time, per server)

```bash
git clone -b test <repo-url> /srv/speedcubinghungary    # or -b main on production
cd /srv/speedcubinghungary
cp .env.test.example .env.test                          # or .env.prod, then fill in secrets
docker compose -f infra/docker-compose.yml -f infra/docker-compose.test.yml --env-file .env.test up -d --build
```

Schema migrations run automatically on API startup (golang-migrate). The
database starts empty — migrations are schema-only and contain no seed data.

## Secrets & configuration

All configuration is environment-driven. **No secrets or environment-specific
values are committed** — only the `.env*.example` templates are tracked; every
real `.env` file (root `.env.dev`/`.env.test`/`.env.prod` and the Python job's
`apps/scripts/weekly-report/.env`) is git-ignored.

- **Development:** copy `.env.dev.example` → `.env.dev` locally.
- **Test / Production:** by design, the real `.env.test` / `.env.prod` live
  **on the deploy server**, created once during server setup and filled with the
  real secrets. They are git-ignored, so `git reset --hard` during a deploy never
  touches them. CI/CD (see `.github/workflows/`) only needs the SSH secrets to
  reach the server; the app secrets are held on the host, not in the workflow.
  (Trade-off accepted deliberately: simplicity over centralizing secrets in the
  CI/CD platform.)

Rules: never hardcode a secret, password, API key, or environment URL in code,
Dockerfiles, or the compose files. Add new variables to the `.env.*.example`
templates and map them in `infra/docker-compose.yml`.

## Weekly report job (`apps/scripts/weekly-report`)

A standalone Python cron job (not part of the Docker stack). It reads the shared
Postgres and emails a weekly report over SMTP.

```bash
cd apps/scripts/weekly-report
python -m venv .venv && source .venv/bin/activate
pip install -r requirements.txt
cp .env.example .env     # git-ignored; fill in real values, never commit it
```

`config.py` loads `.env` (via `python-dotenv`) and validates required variables
fail-fast — mirroring `apps/api/internal/config/config.go`. All sensitive values
(`EMAIL_FROM`, `EMAIL_TO`, `EMAIL_APP_PASSWORD`, `DATABASE_URL` / `DB_*`) come
from `.env`, never from hardcoded literals. In a test/staging context, point
`EMAIL_TO` at a fixed test inbox so real recipients are never emailed. If a real
credential was ever pasted into a tracked file, rotate it.
