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
- `infra/caddy/Caddyfile` — reverse proxy with automatic HTTPS (Let's Encrypt). In test/prod, web/api are **not** published to the host; only Caddy (80/443) is public. A single domain (`WEB_DOMAIN`) serves everything: `WEB_DOMAIN/api/*` → api (prefix stripped), everything else → web. One origin means the auth cookie is first-party and there's no CORS. Enabled by `COMPOSE_PROFILES=proxy` in the env file.

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

A Python job that reads the shared Postgres and, in a single run:

1. builds the weekly report from the WCA API + the shared database,
2. emails it over SMTP (`EMAIL_TO`),
3. **publishes it as a regular Post** on the site (`post_tag = 'report'`).

Step 3 deliberately reuses the existing Post system rather than adding a parallel
content store, so the report inherits the normal listing (`/posts`), the detail
page (`/posts/<id>`) and the badge rendering with no extra UI.

The report post stores its data as **JSON** (`reportPayload.py`) and the site
renders it with a dedicated, styled component (`WeeklyReportView`) — competition
cards, result chips, milestones. Normal (admin) posts are Markdown; only the
report uses the structured payload, selected by `tag === "report"`. **Every**
Hungarian-interest competitor is listed, not just record-setters (those without a
record show as a name only).

**Reported period:** the job queries a **rolling 7-day window** (run time − 1 week
… run time), not a calendar week. The post title shows exactly that window.

**Idempotency:** `posts.report_week` (the Monday of the week the job ran in) is
`UNIQUE` and the insert is `ON CONFLICT DO NOTHING`, so running the job twice in
the same week never creates a second report post. The job is therefore safe to
retry and safe to trigger manually.

**Manual / test runs** — two flags, never used by cron:

```bash
python main.py --skip-email        # no SMTP, useful for repeated local testing
python main.py --force             # overwrite this week's report with fresh content
python main.py --force --skip-email
```

Without `--force` a re-run leaves the week's existing report untouched; with it
the post is rewritten in place (still exactly one post per week).

The e-mail is sent **after** the post is stored, and a send failure aborts the job
with a logged error and a non-zero exit code — the post and records are already
persisted at that point, so they are never lost to an SMTP outage.

It is **not** part of the always-on stack — it runs on demand or from cron. Both
run paths below read the same env vars.

### Containerized (recommended on the servers)

The job has a `Dockerfile` and a compose service behind the `jobs` profile, so a
normal `up` never starts it. Run it against the shared DB using the same
`--env-file` model as the rest of the stack:

```bash
# staging (reads .env.test, sends to its EMAIL_TO — point that at a test inbox)
docker compose \
  -f infra/docker-compose.yml \
  -f infra/docker-compose.test.yml \
  --env-file .env.test \
  --profile jobs run --rm weekly-report
```

The `EMAIL_FROM` / `EMAIL_TO` / `EMAIL_APP_PASSWORD` vars live in the same
`.env.test` / `.env.prod` as the rest of the stack (see the templates); the DB
config is shared. Inside compose the job reaches the DB over the internal network
(`DB_HOST=db`).

### Standalone (local, no Docker)

```bash
cd apps/scripts/weekly-report
python -m venv .venv && source .venv/bin/activate
pip install -r requirements.txt
cp .env.example .env     # git-ignored; fill in real values, never commit it
python main.py
```

### Scheduling (cron)

The project has no in-app scheduler, so the existing cron approach is used. A
wrapper script (`infra/cron/weekly-report.sh`) runs the containerized job with the
right Compose files, so the crontab entry stays a one-liner:

```bash
chmod +x infra/cron/weekly-report.sh
crontab -e
# every Monday at 08:00, staging:
0 8 * * 1 ENV=test /srv/speedcubinghungary-DEV/speedcubinghungary/infra/cron/weekly-report.sh >> /var/log/schu-weekly-report.log 2>&1
```

`ENV` selects the environment (`dev` / `test` / `prod`) and therefore the Compose
override + `.env.<ENV>` file. Because publishing is idempotent per week, a missed
run can simply be re-triggered by hand:

```bash
ENV=test infra/cron/weekly-report.sh
```

### Tests

```bash
cd apps/scripts/weekly-report
pip install -r requirements-dev.txt
python -m pytest
```

Covered: the weekly idempotency key, the plain-text rendering (including that it
stays markup-free), and the competition filtering shared by the email and the post.
CI runs these on every push (`Scripts (weekly-report tests)` job).

### Configuration

`config.py` loads a local `.env` (via `python-dotenv`) when present and validates
required variables fail-fast — mirroring `apps/api/internal/config/config.go`. In
the container no `.env` is baked in (see `.dockerignore`); the values arrive from
the environment instead. All sensitive values (`EMAIL_FROM`, `EMAIL_TO`,
`EMAIL_APP_PASSWORD`, `DATABASE_URL` / `DB_*`) come from the environment, never
from hardcoded literals. In a test/staging context, point `EMAIL_TO` at a fixed
test inbox so real recipients are never emailed. If a real credential was ever
pasted into a tracked file, rotate it.
