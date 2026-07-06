# CLAUDE.md — Speedcubing Hungary

Permanent project guide for Claude Code. Read this before making changes so you
don't rediscover context each session. Update it when you learn something durable
(see **Instructions for Claude** at the end).

> Confidence: this document is derived directly from the code and infra as of
> 2026-07-01. Where something is inferred rather than stated in the repo, it is
> marked **(inferred)**.

---

## 1. Purpose & high-level architecture

Speedcubing Hungary is the web platform for the Hungarian speedcubing community.
It shows news/posts, an upcoming-competition calendar sourced from the World Cube
Association (WCA), and lets WCA members log in via WCA OAuth. A separate Python
job emails a weekly competition/records report.

It is a **monorepo** with three deployable apps plus shared infra, wired together
by Docker Compose:

```
Browser ──HTTP──▶ web (Next.js SSR + client) ──HTTP──▶ api (Go/chi) ──▶ Postgres
                                                          │
                                                          └──▶ WCA API (OAuth + competitions)
Caddy (test/prod only) terminates HTTPS and proxies WEB_DOMAIN→web, API_DOMAIN→api.
scripts (Python) runs out-of-band, reads the same Postgres, calls WCA, sends email.
```

Key architectural facts:

- **Runtime environment is decoupled from git branch.** The same image runs in
  dev/staging/prod; only the Compose override file + `--env-file` change. See §4.
- **Auth is cookie-based JWT.** The API mints its own HS256 JWT after exchanging a
  WCA OAuth code, stored in an `HttpOnly` `auth` cookie. See §9.
- **The `api` is a thin proxy for WCA competition data** (not stored in the DB) but
  the **owner of posts and users** (stored in Postgres).
- **DB migrations run automatically on API startup** (golang-migrate). Schema only,
  no seed data.

---

## 2. Directory structure

```
speedcubinghungary/
├── apps/
│   ├── api/                 # Go backend (chi router, layered architecture)
│   │   ├── cmd/server/main.go        # entrypoint: config → db → migrate → router → listen :8080
│   │   └── internal/
│   │       ├── auth/        # WCA OAuth login, JWT minting, /me, /logout, users table
│   │       ├── post/        # news posts CRUD (Postgres-backed)
│   │       ├── competition/ # WCA API proxy for upcoming Hungarian competitions
│   │       ├── middleware/  # JSON content-type, JWT auth, RBAC
│   │       ├── config/      # env var loading + validation (fail-fast)
│   │       ├── db/          # Postgres connection (pgx stdlib driver)
│   │       ├── router/      # wires repos→services→handlers + routes
│   │       └── types/       # shared enums: Role, PostTag
│   ├── web/                 # Next.js 16 / React 19 frontend (App Router)
│   │   └── app/             # routes + colocated components/lib/types
│   └── scripts/
│       └── weekly-report/   # Python weekly email job (WCA + Postgres + SMTP)
├── db/migrations/           # golang-migrate SQL files (schema, no seeds)
├── infra/
│   ├── docker-compose.yml            # base stack (shared, values from --env-file)
│   ├── docker-compose.dev.yml        # dev override: bind mounts, hot reload, host ports
│   ├── docker-compose.test.yml       # staging: built images, no bind mounts
│   ├── docker-compose.prod.yml       # prod: built images, no bind mounts
│   ├── docker/api/Dockerfile         # multi-stage (dev=air, prod=static binary)
│   ├── docker/web/Dockerfile         # multi-stage (dev=next dev, prod=standalone)
│   └── caddy/Caddyfile               # reverse proxy + automatic HTTPS (test/prod)
├── .github/workflows/       # ci.yml, deploy-test.yml, deploy-prod.yml
├── docs/                    # (currently empty)
├── README.md                # authoritative run/deploy/branch reference
└── .env*.example            # per-environment env templates
```

### Per-module responsibilities (api)

Each feature package (`auth`, `post`, `competition`) follows the same **layered
vertical slice**: `model.go` (types) → `repository.go` (data access) → `service.go`
(business logic) → `handler.go` (HTTP). Wiring happens in `internal/router/router.go`.

- `auth`: exchanges WCA OAuth code for a WCA token, calls WCA `/me`, upserts the
  user, mints the app JWT cookie. Owns the `users` table.
- `post`: news CRUD. Public reads; admin-only writes. Validates content in
  `validator.go`. Owns the `posts` table.
- `competition`: read-only WCA API proxy. Paginates the public WCA competitions
  endpoint for `country_iso2=HU`, maps to a slim DTO. No DB.

---

## 3. Tech stack & key dependencies

| Layer     | Stack |
| --------- | ----- |
| Frontend  | Next.js **16.2.7** (App Router, Turbopack default), React **19.2.4**, TypeScript 5, Tailwind CSS **v4** (`@tailwindcss/postcss`), `react-icons`. Node 22. |
| Backend   | Go **1.25.5**, `go-chi/chi/v5` (router), `go-chi/cors`, `golang-jwt/jwt/v5`, `golang-migrate/migrate/v4`, `jackc/pgx/v5` (via `stdlib` driver) + `lib/pq` (for `pq.Array`), `google/uuid`. Tests: `stretchr/testify` + `DATA-DOG/go-sqlmock`. |
| Scripts   | Python 3.11+ (weekly-report), Jinja-style HTML templates, SMTP email. `requirements.txt` at `apps/scripts/weekly-report/`. |
| Data      | PostgreSQL **17**. |
| Infra     | Docker Compose, Caddy 2 (alpine) reverse proxy w/ Let's Encrypt, GitHub Actions CI/CD, SSH-based deploy. |

⚠️ **Next.js 16 is not the Next.js in your training data.** See
`apps/web/AGENTS.md`: APIs, conventions, and file structure may differ. Read the
relevant guide in `apps/web/node_modules/next/dist/docs/` before writing frontend
code, and heed deprecation notices. `apps/web/CLAUDE.md` just re-exports
`AGENTS.md`.

---

## 4. Environments (decoupled from branches)

Environment = which Compose override + env file you pass. Same code everywhere.

| Environment | Compose files | Env file | Notes |
| ----------- | ------------- | -------- | ----- |
| Development | `docker-compose.yml` + `docker-compose.dev.yml` | `.env.dev` | bind mounts, hot reload, host ports, **no Caddy** (`http://localhost`) |
| Staging     | `docker-compose.yml` + `docker-compose.test.yml` | `.env.test` | immutable built images, Caddy via `COMPOSE_PROFILES=proxy` |
| Production  | `docker-compose.yml` + `docker-compose.prod.yml` | `.env.prod` | immutable built images, Caddy via `COMPOSE_PROFILES=proxy` |

- The **base** compose file (`docker-compose.yml`) defines structure only — **no
  values are hardcoded**; every value comes from `--env-file`. Do not hardcode
  values there.
- Each env uses a distinct `COMPOSE_PROJECT_NAME` so containers/networks/volumes
  are isolated on the same host.
- In test/prod, `web`/`api` are **not** published to the host; only Caddy (80/443)
  is public. The dev override re-publishes web:3000 / api:8080 for localhost work.

---

## 5. Build / dev / test / deploy workflows

### Local development (hot reload)
```bash
cp .env.dev.example .env.dev   # first time only
docker compose -f infra/docker-compose.yml -f infra/docker-compose.dev.yml --env-file .env.dev up
```
Frontend http://localhost:3000 · API http://localhost:8080. Source is bind-mounted:
editing a frontend file → Next.js hot reload (`next dev --webpack`, polling on);
editing a backend file → `air` rebuild/restart. No `docker compose build` needed
during normal dev.

### Run tests / lint locally (without Docker)
```bash
cd apps/api && go build ./... && go vet ./... && go test ./...   # backend CI parity
cd apps/web && npm ci && npm run lint && npm run build           # frontend CI parity
```

### Staging / production deploy
Push to `test` → deploy to staging; push to `main` → deploy to production (both run
CI first, then SSH-deploy with the matching Compose files). Manual local deploy:
```bash
docker compose -f infra/docker-compose.yml -f infra/docker-compose.prod.yml --env-file .env.prod up -d --build
```

### CI/CD (`.github/workflows/`)
- `ci.yml` — backend (`go build`/`vet`/`test`) + frontend (`npm ci`/`lint`/`build`).
  Runs on `develop`/`feature/**`/`hotfix/**` and PRs; reused by deploy workflows via
  `workflow_call`.
- `deploy-test.yml` — push to `test`: CI then SSH-deploy (GitHub `test` Environment).
- `deploy-prod.yml` — push to `main`: CI then SSH-deploy (GitHub `production`
  Environment; add **Required reviewers** to gate prod).
- Deploy Environments need secrets `SSH_HOST`/`SSH_USER`/`SSH_KEY`/`SSH_PORT` and
  variable `DEPLOY_PATH`.

### Branch strategy
`main` (prod, default) ← `test` (staging) ← `develop` (integration). `feature/*`
and `hotfix/*` get CI only. Promotion is linear fast-forward:
```bash
git switch test && git merge --ff-only develop && git push origin test
git switch main && git merge --ff-only test    && git push origin main
git switch develop
```

---

## 6. Coding conventions & architectural patterns

### Backend (Go) — the dominant pattern to follow
Every feature is a package under `internal/<feature>/` with the same four files:

1. **`model.go`** — plain structs. DB models tag `db:"col"`; API DTOs tag
   `json:"..."`. Note: WCA DTOs use `snake_case` JSON, but the `competition`
   response DTO deliberately uses **camelCase** to match the web client.
2. **`repository.go`** — a `Repository` interface + a concrete implementation
   (`postgresRepository` / `PostgresRepository` / `WCARepository`). All SQL lives
   here as raw `database/sql` queries (no ORM). `NewXxxRepository(db)` constructor.
3. **`service.go`** — a `Service` (interface in `auth`, concrete struct in
   `post`/`competition`) holding the repo. Business logic + validation. `NewService(repo)`.
4. **`handler.go`** — HTTP handlers on a `Handler` struct holding the service.
   Decode request, call service, encode JSON response / `http.Error`. `NewHandler(service)`.

Wiring is manual constructor injection in `router.New` — `repo → service → handler`,
then register routes. **There is no DI framework; follow this manual pattern.**

Other Go conventions observed:
- Errors are surfaced with `http.Error(w, err.Error(), status)` and returned up the
  stack; no structured error type. `main.go` uses `log.Fatal` for startup failures.
- Enums are string types in `internal/types` (`Role`, `PostTag`) with `UPPER_SNAKE`
  const names and lowercase string values.
- Comments and some identifiers are **in Hungarian** (e.g. `// kell a cookie-hoz`).
  This is normal here; match the surrounding language when editing a file.
- Nullable DB columns use `sql.NullString` then copy into the struct.
- `roles` is stored as JSONB and marshaled/unmarshaled manually with `encoding/json`.

### Frontend (Next.js App Router)
- Routes live in `apps/web/app/<route>/page.tsx`, with **colocated**
  `components/`, `lib/`, and `types.ts` per route (e.g. `app/calendar/`,
  `app/posts/`). Shared UI is in `app/components/`.
- **Server Components fetch data server-side** (e.g. `app/posts/page.tsx`
  `async` component with `fetch(..., { cache: 'no-store' })`); interactive pieces
  are `"use client"` components.
- **Two API base URLs**: server-side fetches use `SERVER_API_URL`
  (`http://api:8080`, the Docker service name); browser fetches use
  `NEXT_PUBLIC_API_URL` (`http://localhost:8080` / public API domain). The posts
  page falls back through both. **Respect this split** — server code cannot reach
  `localhost:8080` inside the container.
- Auth on the client is via `UserProvider` (`app/components/UserProvider.tsx`), a
  context calling `/me` and `/logout` with `credentials: "include"`. Use
  `useUser()` to read `{ user, isLoading, logout }`.
- Styling is Tailwind v4 utility classes inline in JSX. UI copy is largely Hungarian.

### Python (weekly-report)
- Package-per-concern under `apps/scripts/weekly-report/`
  (`CompetitionModels/`, `database/`, `templates/`), orchestrated by `main.py`.
  Uses `logging`, exits non-zero on failure. Reads the shared Postgres read-only
  and sends email via SMTP. HTML templates under `templates/`.

---

## 7. Database schema & data flow

Single migration so far: `db/migrations/000001_init_schema.{up,down}.sql`. Applied
automatically on API startup via `migrate.New("file:///migrations", cfg.DatabaseURL)`
(the prod image bakes `db/migrations` to `/migrations`).

```sql
post_tag ENUM ('event','competition','news','announcement')

posts(id UUID PK, title TEXT, tag post_tag, content TEXT,
      created_at TIMESTAMP default now(), updated_at TIMESTAMP default now())

users(id SERIAL PK, wca_id VARCHAR(50) UNIQUE, name, email,
      roles JSONB default '[]', competition_count INT, delegated_count INT,
      created_at, updated_at)

records(region VARCHAR, event_id VARCHAR, single INT, average INT,
        PRIMARY KEY (region, event_id))
```

Notes / current state:
- `posts.id` is a `uuid.NewString()` generated in the service (not a DB default).
- `users.roles` is JSONB; login upsert **preserves existing roles** and only seeds
  `["user"]` on first insert (see `auth/repository.go` `Upsert` `ON CONFLICT`).
- `records` is written by the Python job (`SaveRecords`); the Go API does not touch
  it yet. `competition_count`/`delegated_count` on users are read by the API but
  populated elsewhere **(inferred: by the Python job / future work)**.
- Competition data is **not** persisted — it is proxied live from WCA per request.

### Adding a migration
Create the next numbered pair `db/migrations/00000N_<name>.{up,down}.sql`. It runs
on the next API start. Keep migrations schema-focused (no seed data — the DB starts
empty by design).

---

## 8. API architecture & endpoints

chi router in `internal/router/router.go`. Global middleware: `middleware.JSON`
(sets `Content-Type: application/json`) and CORS (allows `cfg.FrontendURL`,
`AllowCredentials: true` for the cookie).

| Method | Path | Auth | Handler |
| ------ | ---- | ---- | ------- |
| GET  | `/posts` | public | `post.GetPosts` |
| GET  | `/posts/{id}` | public | `post.GetPostByID` |
| GET  | `/posts/latest` | public | `post.GetLatestPosts` (limit 3) |
| POST | `/posts` | JWT + `admin` | `post.CreatePost` |
| DELETE | `/posts` | JWT + `admin` | `post.DeletePosts` (body `{ids:[]}`) |
| GET  | `/competitions/upcoming` | public | `competition.GetUpcoming` (HU only) |
| GET  | `/login/callback` | public | `auth.LoginCallback` (WCA OAuth redirect target) |
| GET  | `/me` | JWT | `auth.Me` |
| POST | `/logout` | JWT | `auth.Logout` |

Integrations: **WCA API** — OAuth token + `/api/v0/me` (auth), and
`/api/v0/competitions` (competition proxy, paginated, `maxPages=20` guard, 10s
client timeout).

---

## 9. Authentication & authorization

Flow:
1. Frontend `LoginButton` redirects to WCA `oauth/authorize` with `client_id`,
   `redirect_uri`, `scope=public email manage_competitions`, and `state=<pathname>`.
2. WCA redirects back to the **API** `/login/callback?code=...&state=...`.
3. `auth.LoginCallback` exchanges the code for a WCA access token, calls WCA `/me`,
   **upserts** the user, mints an HS256 JWT (`UserClaims`: user_id, wca_id, name,
   roles; 1-hour expiry), and sets it as cookie `auth`.
4. Cookie is `HttpOnly`, `SameSite=Strict`, `Path=/`, `MaxAge=3600`,
   `Secure` unless `DEPLOYMENT=development`. Then redirects to
   `FRONTEND_URL + state` (state is sanitized: must start with a single `/`).

Authorization:
- `middleware.AuthMiddlewareWithConfig(cfg)` reads the `auth` cookie, validates the
  JWT (rejects non-HMAC algs), and puts `*UserClaims` in request context.
  Retrieve with `middleware.GetUserFromContext(r)`.
- `middleware.RBACMiddleware(types.ROLE_ADMIN, ...)` requires the user to have one
  of the given roles; returns 401 if unauthenticated, 403 if lacking the role.
- `middleware.AuthMiddleware` (no config) is **deprecated** and always errors — use
  `AuthMiddlewareWithConfig`.

⚠️ Note (as-is): `auth/service.go` reads WCA/JWT secrets directly from
`os.Getenv(...)` rather than from the loaded `cfg`, and `auth/handler.go` reads
`FRONTEND_URL`/`DEPLOYMENT` from env too. The `config` package validates these at
startup, so they are present — but if you refactor, prefer threading `cfg` through
rather than adding more `os.Getenv` calls **(inferred preference)**.

---

## 10. Environment variables & configuration

Templates: `.env.dev.example`, `.env.test.example`, `.env.prod.example` at repo
root. Never commit real `.env*` files (git-ignored). Backend config is loaded and
**validated fail-fast** in `internal/config/config.go`.

**Required (API fails to start without these):** `DATABASE_URL`, `JWT_SECRET`
(≥32 chars — generate with `openssl rand -hex 32`), `WCA_CLIENT_ID`,
`WCA_CLIENT_SECRET`, `WCA_REDIRECT_URI`, `FRONTEND_URL`.

Other important vars:
- DB connection (used by `db/postgres.go`): `DB_HOST`, `DB_PORT`, `DB_USER`,
  `DB_PASSWORD`, `DB_NAME` (separate from `DATABASE_URL`, which migrate uses).
- `DEPLOYMENT` (`development`/`test`/`production`) — controls cookie `Secure` +
  helpers. `LOG_LEVEL`. `NODE_ENV`.
- Web: `SERVER_API_URL` (server→api, Docker service name), `NEXT_PUBLIC_API_URL`
  (browser→api), `NEXT_PUBLIC_WCA_CLIENT_ID`, `NEXT_PUBLIC_WCA_REDIRECT_URI`,
  `NEXT_PUBLIC_BACKEND_LOGIN`. **`NEXT_PUBLIC_*` are baked into the browser bundle
  at build time** (Docker build args), so changing them requires a web rebuild.
- Compose: `COMPOSE_PROJECT_NAME`, `COMPOSE_PROFILES=proxy` (test/prod),
  `WEB_DOMAIN`/`API_DOMAIN`/`ACME_EMAIL` (Caddy), `WEB_PORT`/`API_PORT`/
  `DB_PORT_HOST` (dev host ports).
- Weekly-report job (Python): configured via `apps/scripts/weekly-report/.env`
  (git-ignored; template `.env.example`), loaded by `config.py` via
  `python-dotenv` with fail-fast validation mirroring
  `apps/api/internal/config/config.go`. It is **not** loaded by Docker Compose
  — it's a standalone cron. All secrets (`EMAIL_FROM`, `EMAIL_TO`,
  `EMAIL_APP_PASSWORD`, `DATABASE_URL`/`DB_*`) come from `.env`; **never
  hardcode them in `config.py` or any tracked file**. Point `EMAIL_TO` at a
  fixed test inbox in test/staging.

**Secret-delivery model (deliberate choice):** test/prod deploys use a
server-local `.env.test`/`.env.prod` placed once during server setup (git-ignored,
survives `git reset --hard`). The CI/CD workflows only carry SSH secrets, not the
app secrets. This trades "secrets centralized in the CI/CD platform" for
simplicity — **do not "fix" this** by moving env generation into the workflow
unless explicitly asked.

---

## 11. Common workflows for adding features

### Add a backend endpoint to an existing feature
1. Add the method to the `Repository` interface + implementation (`repository.go`).
2. Add business logic to `service.go` (validate in a `validator.go` if the feature
   has one).
3. Add the handler in `handler.go` (decode → call service → encode/`http.Error`).
4. Register the route in `router.New`; wrap in the auth/RBAC `r.Group` if protected.
5. Add a table-driven / testify test next to the file (`*_test.go`). Mock the repo
   with testify `mock` (see `auth/service_test.go`) or `go-sqlmock` for SQL.

### Add a new backend feature package
Mirror `internal/post/`: create `internal/<feature>/{model,repository,service,handler}.go`
with the same constructor + interface shape, add any shared enum to
`internal/types/`, and wire it in `router.New`.

### Add a frontend route
Create `apps/web/app/<route>/page.tsx` (+ colocated `components/`, `lib/`,
`types.ts`). Fetch server-side with `SERVER_API_URL` where possible; use
`"use client"` + `NEXT_PUBLIC_API_URL` for interactive/browser fetches. Add nav
links in `app/layout.tsx` if needed. **Read the Next 16 docs first** (§3 warning).

---

## 12. Known constraints, assumptions & tech debt

- **Untracked working tree:** `apps/api`, `apps/web`, and other files show as
  untracked/added in git status — the project is mid-setup; commit history is thin.
- `apps/web/app/layout.tsx` still has the default `create-next-app` `<title>` /
  metadata — likely should be updated.
- The navbar links to `/blog`, but the posts route is `/posts` (no `/blog` page
  exists) — likely a naming mismatch to reconcile.
- `auth/service.go` / `auth/handler.go` bypass the `config` struct and read env
  directly (§9). Some `AvatarURL` fields are commented out across auth structs —
  avatar support is stubbed, not implemented.
- `post` has `Update`/`UpdatePost` in service+repository but **no route** exposes
  it yet.
- `competition_count` / `delegated_count` / `records` are defined but not populated
  by the Go API.
- No structured logging/observability beyond stdlib `log` + `LOG_LEVEL` (the latter
  is loaded into `cfg` but **not consumed** by any logger yet).
- `apps/api/tmp/` (air build output) and `__pycache__/` dirs are build artifacts.
- No Redis / message queue / object storage in the stack — the only data store is
  Postgres. Don't assume a cache/queue layer exists.
- `db/postgres.go` uses `sslmode=disable`; acceptable because the DB is only on the
  internal Docker network (never published in test/prod). Make it configurable if
  the DB is ever exposed.
- The weekly-report job previously hardcoded the Gmail app password + DB creds in
  a committed-looking `env.py` module. It's now `config.py` + a git-ignored `.env`
  (template: tracked `.env.example`), loaded via `python-dotenv`. Keep it
  secret-free and env-driven — never hardcode values back into `config.py`.
- Legacy per-directory `.env` files (root `.env`, `apps/api/.env`, `apps/web/.env`,
  `db/.env`) and `apps/web/.env.example` are git-ignored leftovers from an older
  "single .env for all envs" model. The live model is one root `.env.<env>` loaded
  via `--env-file`; don't reintroduce per-service env files.

---

## 13. Performance considerations

- Competition data is fetched live from WCA on every `/competitions/upcoming`
  request with no caching, paginating up to 20 pages (10s HTTP timeout). If this
  becomes hot, add caching rather than hammering WCA.
- Posts list queries are unpaginated (`GetAll` returns everything); `/posts/latest`
  is capped at 3. Fine at current scale; revisit if `posts` grows large.
- DB access opens a single `*sql.DB` pool (default pool settings). No explicit
  connection tuning.
- Frontend prod uses Next.js `output: "standalone"` for a slim runtime image; dev
  uses filesystem **polling** (Docker bind-mount limitation on Windows/WSL), which
  is CPU-ish but required for hot reload.

---

## 14. Security considerations

- JWT is HS256 with a validated ≥32-char secret; the auth middleware **rejects
  non-HMAC signing algorithms** (alg-confusion guard) — keep that check.
- Auth cookie is `HttpOnly` + `SameSite=Strict` + `Secure` (outside dev). CORS
  allows only `FRONTEND_URL` with credentials.
- OAuth `state` is sanitized before use in the redirect (must be a single-slash
  path) to prevent open redirects — preserve this on changes.
- SQL uses parameterized queries throughout (`$1`, `pq.Array`) — **never build SQL
  with string concatenation**.
- Secrets come only from env/`--env-file`; nothing hardcoded in compose or code.
- In test/prod, only Caddy is internet-facing (auto HTTPS via Let's Encrypt); the
  DB and app services are not published to the host.

---

## 15. Recurring patterns cheat-sheet

- Go feature = `model/repository/service/handler` + interface + `NewXxx`
  constructor, wired manually in `router.New`.
- HTTP handler shape: decode into a local anonymous request struct → call service →
  `json.NewEncoder(w).Encode(...)` or `http.Error(w, err.Error(), status)`.
- Repository returns domain structs + `error`; SQL is raw and parameterized.
- Enums live in `internal/types` as string types.
- Frontend: server component fetches → passes props to client components; auth via
  `useUser()`; two API base URLs (server vs browser).
- Hungarian is used freely in comments and UI copy.

---

## 16. Files & directories to generally NOT modify

- `apps/*/node_modules/`, `apps/web/.next/`, `apps/api/tmp/`,
  `**/__pycache__/`, `*.tsbuildinfo` — build artifacts / vendored deps.
- `go.sum`, `package-lock.json` — change only via the proper tool
  (`go mod`, `npm`), never by hand.
- Real `.env`, `.env.dev`, `.env.test`, `.env.prod` — secrets, git-ignored; edit
  the `.example` templates instead when documenting new vars.
- `infra/docker-compose.yml` values — keep it value-free; put values in env files.
- Existing DB migrations are effectively immutable once applied — add a **new**
  numbered migration instead of editing `000001_*`.

---

## 17. Implicit conventions (not documented elsewhere)

- The API always listens on `0.0.0.0:8080`; the frontend on `3000`. These are
  assumed by Caddy, compose, and the `NEXT_PUBLIC_*` defaults.
- Docker build context is the **repo root** (so the API image can bake
  `db/migrations`), even though Dockerfiles live under `infra/docker/`.
- `next dev` must run with `--webpack` in Docker (Turbopack has no polling fallback
  for bind mounts). Don't "modernize" this back to Turbopack for the dev container.
- `air` and Next both **poll** the filesystem in dev by design (Windows/WSL Docker).
- JSON field casing: DB models/`post` use `snake_case`; the `competition` API DTO
  intentionally uses `camelCase` to match the web client. Match the neighboring
  file, don't unify blindly.
- Constructor naming is inconsistent across packages (`NewPostgresRepository` vs
  `postgresRepository`, `Service` interface in `auth` vs concrete struct in `post`).
  When editing a package, match **that package's** existing style.

---

## Instructions for Claude

When working in this repo:

1. **Follow existing patterns instead of introducing new ones.** New Go features
   use the `model/repository/service/handler` + `router.New` wiring shape. New
   frontend routes colocate `components/lib/types` under `app/<route>/`.
2. **Read the Next.js 16 docs before writing frontend code** — this Next version
   differs from training data. See `apps/web/AGENTS.md`.
3. **Minimize unnecessary refactoring.** Don't rename constructors, unify JSON
   casing, or replace raw SQL with an ORM unless asked. Match the neighboring
   file's style — including Hungarian comments/copy.
4. **Keep changes focused and localized.** Touch the feature package (and its
   router wiring) you're changing; avoid cross-cutting sweeps.
5. **Prefer consistency over cleverness.** Copy the closest existing example rather
   than inventing an abstraction.
6. **Preserve backward compatibility** unless explicitly told otherwise —
   especially the auth cookie/JWT contract, the two-API-base-URL split, the
   env-var contract in §10, and existing migrations (add new ones, don't edit old).
7. **Never hardcode secrets or values** in compose/code; add them to the `.env`
   templates and the base compose's env mapping.
8. **Keep security invariants:** parameterized SQL, the non-HMAC alg rejection, the
   OAuth `state` sanitization, and cookie flags.
9. **Ask for clarification before architectural changes** (new frameworks, new
   top-level services, DB engine/auth-model changes, DI frameworks).
10. **Verify before recommending:** parts of this file are marked **(inferred)** and
    the tree is mid-setup — confirm a file/route/flag still exists before relying
    on it.
11. **Update this CLAUDE.md** whenever you discover durable new project knowledge
    (new endpoints, migrations, env vars, conventions, or resolved tech debt).
