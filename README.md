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

- `ci.yml` — build + lint + test (Go — including the weekly-report job — and Next.js). Runs on `develop`/`feature/*`/`hotfix/*` and PRs, and is reused by the deploy workflows.
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

All configuration is environment-driven. **One env file per environment, shared by
the whole stack** — there are no per-app `.env` files. Only the `.env*.example`
templates are tracked; the real root `.env.dev`/`.env.test`/`.env.prod` are
git-ignored.

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

## Weekly report job (`apps/api/cmd/weekly-report`)

A Go binary next to the API server (same module, same Docker image, package
`internal/weeklyreport`). In a single run it:

1. builds the weekly report from the WCA API + the shared database, and
   refreshes everyone's competition / delegation / successful-solve counts
   there from the WCA (official numbers + competitions whose results are only
   on WCA Live so far),
2. **publishes it as a regular Post** on the site (`post_tag = 'report'`),
3. refreshes the stored WR/ER/NR table from the WCA,
4. emails it over SMTP (`EMAIL_TO`).

Step 2 deliberately reuses the existing Post system rather than adding a parallel
content store, so the report inherits the normal listing (`/posts`), the detail
page (`/posts/<id>`) and the badge rendering with no extra UI.

The report post stores its data as **JSON** (`payload.go`) and the site renders it
with a dedicated component (`WeeklyReportView`). **Every** Hungarian-interest
competitor is listed in the post; the **e-mail** lists only the record-setters of
Hungarian competitions (and every Hungarian at foreign ones).

**Reported period:** a **rolling 7-day window** (run time − 1 week … run time,
Budapest time), not a calendar week. The post title shows exactly that window.

**Idempotency:** `posts.report_week` (the Monday of the week the job ran in) is
`UNIQUE` and the insert is `ON CONFLICT DO NOTHING`, so a second run in the same
week never creates a second post. The counters are recomputed from the WCA on
every run, so re-running is safe for them too.

**Flags** — never used by cron:

```bash
weekly-report -skip-email        # no SMTP, useful for testing
weekly-report -force             # overwrite this week's report with fresh content
weekly-report -counts-only       # only refresh the counters and records from the WCA (no post, no e-mail)
weekly-report -dry-run           # write NOTHING to the DB; log what would change, send the e-mail as "[TESZT] …"
```

The e-mail is sent **after** the post is stored, and any failure exits non-zero
with a logged error — the post is already persisted by then, so it is never lost
to an SMTP outage.

### Running it

The job runs **inside the running `api` container** (`docker compose exec`), so it
always uses the deployed code, the env vars the API already has, and a schema the
API has already migrated. `infra/cron/weekly-report.sh` wraps that:

```bash
ENV=test infra/cron/weekly-report.sh                 # staging, sends to its EMAIL_TO
ENV=test infra/cron/weekly-report.sh -skip-email     # flags are passed through
ENV=dev  infra/cron/weekly-report.sh -skip-email     # dev: `go run` in the dev container
```

`EMAIL_FROM` / `EMAIL_TO` (comma-separated for several recipients) /
`EMAIL_APP_PASSWORD` (a Gmail app password) live in the same `.env.<env>` as the
rest of the stack and are mapped to the `api` service. Point `EMAIL_TO` at a test
inbox on staging.

### Scheduling (cron)

There is no in-app scheduler; the host's cron calls the wrapper:

```bash
chmod +x infra/cron/weekly-report.sh
crontab -e
# every Monday at 08:00, staging:
0 8 * * 1 ENV=test /srv/speedcubinghungary-DEV/speedcubinghungary/infra/cron/weekly-report.sh >> /var/log/speedcubinghungary-weekly-report.log 2>&1
```

The cron line's time is in the **server's** time zone. A missed run can simply be
re-triggered by hand (`ENV=test infra/cron/weekly-report.sh`).

### Tests

```bash
cd apps/api && go test ./internal/weeklyreport
WEEKLY_REPORT_SMOKE=HungarianOpen2026 go test ./internal/weeklyreport -run Smoke -v   # real WCA API
```

### Importing the old standalone job's data (one-time, per environment)

The standalone `kiaczb/weekly-report` job kept its counters in its own
`persons` / `records` tables. Before the **first** Go run on an environment,
load its `pg_dump` (`INSERT` format) into the shared DB:

```bash
ENV=test infra/scripts/import-weekly-report-dump.sh /path/to/dump-weekly-report.sql
ENV=test infra/cron/weekly-report.sh -counts-only    # then correct every counter from the WCA
```

It needs the `db` container running and the API migrated (migration `000019`).
Only the dump's `persons` / `records` `INSERT` rows are used — its DDL is
ignored — and everything happens in one transaction. `persons` map onto `users`:
new people are inserted; existing users keep their name, roles and e-mail, but
their competition / delegation counts are **set** to the dump's (and their
solve count, where the dump has one). Records the DB already has are left
alone. The import mainly tells the job *who* to track — the `-counts-only` run
right after it replaces every count with the WCA-derived one. Stop the old job
once the Go one takes over.
