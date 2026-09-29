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
Caddy (test/prod only) terminates HTTPS on a single domain: WEB_DOMAIN/api→api (prefix stripped), everything else→web.
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
│   │       ├── allrounder/  # Hungarian All-Rounder: cached WCA results + live leg + scoring
│   │       ├── special/     # Budapest Special: live WCA Live results + recorded DNF times
│   │       ├── wcalive/     # shared WCA Live transport (GraphQL, Phoenix socket, broadcaster)
│   │       ├── middleware/  # JSON content-type, JWT auth, RBAC
│   │       ├── httperr/     # hiba → HTTP válasz (5xx szövege naplóba, nem a kliensnek)
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
(business logic) → `handler.go` (HTTP). Wiring (repo → service → handler) happens
in `cmd/server/main.go`; `internal/router/router.go` takes the finished handlers
as parameters and only registers routes.

- `auth`: exchanges WCA OAuth code for a WCA token, calls WCA `/me`, upserts the
  user, mints the app JWT cookie. Owns the `users` table.
- `post`: news CRUD. Public reads; admin-only writes. Validates content in
  `validator.go`. Owns the `posts` table.
- `competition`: read-only WCA API proxy. Paginates the public WCA competitions
  endpoint for `country_iso2=HU`, maps to a slim DTO. No DB.
- `special`: **Budapest Special** — WCA Live integration. Official results are
  **not stored**: they are fetched live and held in memory, kept real-time by a
  `roundUpdated` websocket subscription that runs only while someone has the page
  open. The **only** persisted data is the human-recorded duration of a DNF
  attempt. Owns `special_competitions` and `special_dnf_attempts`. The
  competition is chosen **by the viewer**, per request (`?competition=<wcaId>`);
  there is no server-side "tracked competition" and no admin selection. See §18.
- `allrounder`: **Hungarian All-Rounder** — the season-long point standings of the
  competition series. Finished legs' official WCA results are **cached in
  Postgres** (`allrounder_syncs` + `allrounder_results`); the **currently running
  leg comes from WCA Live**, live, because official results only exist after the
  delegate uploads them. Both sources feed the *same* scoring, which runs
  server-side. The series definition (which competitions, points per position,
  how many results count) lives in **code**, `series.go`. See §19.
- `wcalive`: the **shared WCA Live transport and session lifecycle** — GraphQL
  client + `ResolveCompetitionID`, the Phoenix/Absinthe `SocketClient`, the keyed
  `Broadcaster`, and `Session` (viewer refcount, socket start/stop, subscription
  diffing, reconnect resync). Used by both `special` and `allrounder`.
  ⚠️ The **GraphQL queries stay in the feature packages**, because they ask for
  different things (`special` wants attempts + time limits, `allrounder` wants
  `ranking`), and one combined query would hit WCA Live's complexity limit.
  Each feature still owns its own `SocketClient` instance (two websockets is
  cheaper than sharing one connection's state), but the **lifecycle around it is
  shared** — see §18 for the `Session`/`Watcher` split.

---

## 3. Tech stack & key dependencies

| Layer     | Stack |
| --------- | ----- |
| Frontend  | Next.js **16.2.7** (App Router, Turbopack default), React **19.2.4**, TypeScript 5, Tailwind CSS **v4** (`@tailwindcss/postcss`), `react-icons`, `@cubing/icons` (WCA event icon font, ~13 KB, imported only by `/tools/allrounder`). Node 22. |
| Backend   | Go **1.25.5**, `go-chi/chi/v5` (router), `go-chi/cors`, `golang-jwt/jwt/v5`, `golang-migrate/migrate/v4`, `jackc/pgx/v5` (via `stdlib` driver) + `lib/pq` (for `pq.Array`), `google/uuid`, `coder/websocket` (WCA Live subscriptions). Tests: `stretchr/testify` + `DATA-DOG/go-sqlmock`. |
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

Wiring is manual constructor injection in `cmd/server/main.go` — `repo → service →
handler`, one block per feature, then the finished handlers are passed into
`router.New(cfg, postHandler, competitionHandler, predictionHandler, authHandler,
specialHandler, allrounderHandler)`, which only registers routes.
**There is no DI framework; follow this manual pattern.**

Other Go conventions observed:
- Errors are surfaced with `httperr.Write(w, r, status, err)` and returned up the
  stack; no structured error type. `main.go` uses `log.Fatal` for startup failures.
  ⚠️ **Don't go back to `http.Error(w, err.Error(), status)`.** `httperr.Write`
  passes the error text through for 4xx (those are our own sentinels, and the
  caller needs to know what it got wrong) but for **5xx it logs the error and
  sends a generic message** — a raw `pq:` error names schema tables and columns,
  a `*url.Error` names the upstream hosts and our query parameters, and neither
  helps the browser. The status is still the caller's decision: it knows which
  of its sentinels means what.
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
- **There is a design-token layer in `app/globals.css` (`@theme`) — use it.**
  Colors (`brand-red`/`brand-green`, `canvas`/`surface`/`surface-muted`/
  `surface-inverse`, `line`/`line-strong`, `foreground`/`muted-foreground`/
  `subtle-foreground`), radii (`rounded-control`/`card`/`panel`), shadows
  (`shadow-card`/`card-hover`/`panel`) and one easing (`ease-smooth`). **Don't
  reintroduce raw `gray-*`, `shadow-xl`, `rounded-2xl`** — match the tokens.
  Numeric content (times, dates, counts) uses `font-mono` + the `.tabular`
  helper; that's deliberate for a timing sport, not decoration.
- **Shared primitives live in `app/components/ui/`**: `Button` (variants
  `primary`/`secondary`/`ghost`/`danger`/`inverse`; renders `<button>`,
  `<Link>`, or external `<a>` based on `href`/`external`), `Field` (+ the
  exported `controlClass` for inputs), `PageHeader`, `StateMessage`,
  `Modal` (**the shell every modal is built on**: backdrop, panel, header bar,
  Escape, backdrop click, scroll lock — which also pads `body` by the
  scrollbar's width, otherwise the page shifts right when the scrollbar
  disappears — and the **focus trap**. Tab cycles
  inside the panel, focus lands there on open and returns to whatever opened it
  on close. The backdrop closes it only when the pointer went down **and** up
  on the backdrop — a text selection that starts in a field and ends outside
  would otherwise wipe a half-filled form. Props: `title`, `onClose`, `size`
  (`sm` confirmation / `md` reading, the default / `lg` form), `role`,
  `closeDisabled` (blocks every way out while a request is in flight),
  `hideHeader` for a dialog that renders its own title, `initialFocusRef`.
  ⚠️ Don't hand-roll a backdrop + panel again — the four that existed had each
  dropped the focus trap, so Tab walked out of the dialog into the page behind
  it),
  `ConfirmModal` (delete confirmation, built on `Modal`; takes `title`,
  `description` (any ReactNode — it renders in a `div`),
  `onConfirm: () => Promise<void>`, `onClose`; the buttons read
  "Mégse"/"Törlés"; manages its own submitting/error state and closes itself
  on a successful `onConfirm`). Plus
  `app/components/Cards/Card.tsx` (`padding`, `interactive`) and
  `app/components/Badge.tsx`. **Reuse these
  instead of hand-rolling a button/card/field/confirm dialog** — in
  particular, never use the native `window.confirm`/`window.alert` for a
  destructive action; use `ConfirmModal` (see `DeletePostButton.tsx` /
  `DeletePredictionButton.tsx` for the pattern: an icon button that opens the
  modal, an async `onConfirm` that does the fetch + `router.refresh()`).
- **Shared data lives in `app/lib/`**: `api.ts` (**the fetch layer** — the
  two-base-URL split below plus the four things every call site otherwise has
  to get right on its own: `serverJson(path, fallback)` for Server Components
  (`cache: "no-store"`, logs and falls back so an unreachable API leaves an
  empty page rather than an error screen; its `catch` starts with
  `unstable_rethrow`, because a no-store fetch signals "render dynamically"
  by throwing, and swallowing that is a Next misuse), `clientFetch(path, init)` /
  `clientJson(path, init)` for the browser (**always** sends the JWT cookie —
  forgetting `credentials: "include"` silently sends a logged-out request — and
  turns a non-ok response into an `ApiError` carrying `status` and the server's
  message, with friendly Hungarian text for 401/403 and 429). Pass a JSON body
  as `json:` so the `Content-Type` comes with it; `FormData` goes in `body:`,
  because the browser must set that header itself for the boundary.
  ⚠️ `EventSource` is not `fetch`, so the two SSE streams still build their URL
  from `clientApiUrl()` — that's expected, not a leftover), `site.ts`
  (`NAV_LINKS`, `CONTACT_ITEMS`, `SITE_DESCRIPTION`, used by the navbar, footer
  and contact page), `events.ts` (`eventName()` — the WCA event id → Hungarian
  name table, shared by the allrounder and prediction tools; **don't copy it
  back into a tool**), `date.ts` (every date/time the UI renders — `hu-HU` and
  **pinned to `Europe/Budapest`**, because `toLocaleDateString` otherwise uses
  the runtime's zone and a Server Component runs in the container's UTC while a
  client component runs on the visitor's machine; near midnight the same date
  rendered as two different days. Add a new shape here rather than writing
  another `toLocaleDateString` in a component), `time.ts` (WCA-style digit-only time entry, shared by
  the prediction form and the Special DNF cell: `formatTimeInput()` formats
  while typing and returns `null` for a keystroke it refuses — the caller
  keeps the previous value; `parseTimeInput()` is the **only** validity check
  and every caller must run it before using the value. ⚠️ The formatter
  deliberately lets `1:75.00` through as an intermediate state: digits fill
  from the right, so typing `16:05.00` passes through `1:60.50`, and rejecting
  that would make whole minute ranges untypeable). Add nav/contact entries in
  `site.ts`, not per-component.
- **Dark mode is implemented and follows the OS** (`prefers-color-scheme`; no
  toggle, no class strategy, so no JS and no hydration flash). It works by
  **redefining the design tokens** in a `@media (prefers-color-scheme: dark)`
  block in `globals.css` — so `bg-surface`, `text-foreground`, `border-line`
  etc. flip on their own and most components needed no change. Rules to keep:
  - **`dark:` is only for hardcoded colors** a token can't reach (`bg-white` on
    the navbar/hero, the badge tints in `posts/types.ts`, the modal scrim).
    Don't add `dark:` next to a token utility — the token already handles it.
  - **`--color-surface-inverse` stays dark in BOTH themes.** It's the feature-band
    colour (footer, news band, "next competition" card). That's why the many
    `text-white` classes sitting on it are correct in both themes and must not be
    "fixed".
  - **`bg-foreground` must pair with `text-background`, never `text-white`** —
    `foreground` inverts to near-white in dark, so `text-white` becomes invisible.
  - ⚠️ **Tailwind v4 inlines `--shadow-*` values into the utility at build time**,
    so overriding `--shadow-card` in the dark block does nothing. The shadow
    tokens therefore reference `var(--shadow-tint*)` custom properties, and *those*
    get redefined for dark. Keep that indirection if you touch the shadows.
  - The logo (`SpeedcubingFullAnimatedSVG`) draws its letters and black cube
    pieces with `fill="currentColor"` under `text-foreground`, so it turns white
    in dark mode. The red/green tiles stay brand-coloured. It no longer takes a
    `textColor` prop.
  - `color-scheme` is set per theme so native controls (scrollbars, `<select>`
    dropdown, file input, checkboxes) match.
- **Server Components fetch data server-side** (e.g. `app/posts/page.tsx`
  `async` component with `fetch(..., { cache: 'no-store' })`); interactive pieces
  are `"use client"` components.
- **Two API base URLs**: server-side fetches use `SERVER_API_URL`
  (`http://api:8080`, the Docker service name); browser fetches use
  `NEXT_PUBLIC_API_URL` (`http://localhost:8080` / public API domain).
  **Respect this split** — server code cannot reach `localhost:8080` inside the
  container. In practice you pick the side by picking the helper (`serverJson`
  vs `clientFetch`/`clientJson`, see above) rather than by building the URL.
  **No frontend code calls the WCA API directly** — the home page's "next
  competition" card reads our `/competitions/upcoming` like the calendar does.
- Auth on the client is via `UserProvider` (`app/components/UserProvider.tsx`), a
  context calling `/me` and `/logout` with `credentials: "include"`. Use
  `useUser()` to read `{ user, isAdmin, isLoading, logout }` — gate admin UI on
  `isAdmin` rather than re-checking `roles` (the server enforces it anyway).
- Styling is Tailwind v4 utility classes inline in JSX. UI copy is largely Hungarian.
- **The post editor is Markdown-based, deliberately.** `CreatePostModal` uses
  `app/posts/components/MarkdownEditor.tsx` — a formatting toolbar that inserts
  Markdown syntax plus an "Előnézet" tab that renders through the *same*
  `PostContent` component as the live page. Storage stays Markdown, so the
  no-`rehype-raw` XSS guarantee below is untouched. **Don't swap this for an
  HTML-storing WYSIWYG** without re-solving sanitization and migrating existing
  posts. The editor and the markdown renderer are both `next/dynamic`
  (`ssr: false`) so the ~144 KB markdown chunk and the modal ship only when an
  admin opens them, not to every `/posts` visitor — keep that lazy boundary.
- **Post content is Markdown.** The detail page renders it with
  `app/posts/components/PostContent.tsx` (`react-markdown` + `remark-gfm`),
  styled per-element via the `components` map (no `@tailwindcss/typography`
  plugin). ⚠️ Security: there is **no `rehype-raw`**, so raw HTML in admin content
  is escaped, not rendered — keep it that way. Lists/cards must **not** render
  Markdown — they show a stripped plain-text teaser via
  `app/posts/lib/postPreview.ts`.
- **`tag === "report"` posts are the exception: their `content` is JSON, not
  Markdown.** The detail page (`app/posts/[id]/page.tsx`) branches on the tag:
  `parseWeeklyReport()` (`app/posts/lib/report.ts`) → `WeeklyReportView.tsx`
  (a styled, structured renderer — competition cards, result chips, milestones);
  everything else → `PostContent`. Parse failure falls back to `PostContent` so
  the page never crashes. Lists/cards show `reportSummary()` for report posts, not
  `postPreview`. The payload shape is `WeeklyReport*` in `types.ts` — keep it in
  sync with `apps/scripts/weekly-report/reportPayload.py`.

### Python (weekly-report)
- **The weekly report is published into the existing `posts` table**, not a
  separate content store — so it reuses the normal listing/detail UI. Flow in
  `main.py`: `BuildReportData()` (`reportData.py`) builds the dataset **once**,
  then `WriteEmail()` renders it as HTML for email and
  `PublishWeeklyReportPost()` (`reportPost.py`) serialises the **same** data as
  **JSON** (`reportPayload.py` → `BuildReportPayload`) and inserts it with tag
  `report`. Don't recompute the milestones separately — they hit the DB *and* the
  WCA API.
- The report post `content` is **JSON**, rendered by the frontend's
  `WeeklyReportView` (see the Frontend section) — *not* Markdown and *not* the
  email HTML. The email still renders from Jinja (`templating.py` /
  `html_environment` / `templates/competitions.html`); only the website post uses
  the JSON payload. Keep `reportPayload.py` and `types.ts` (`WeeklyReport*`) in sync.
- **All Hungarian-interest competitors are included, not just record-setters**
  (`BuildReportData` no longer filters). `competitors.py` already returns every
  Hungarian competitor who had a result; those without a record get an empty
  `results` list (shown as a name with a "résztvevő" marker). This also affects
  the email, which shares the data.
- The reported period is a **rolling 7-day window** (run time − 1 week … run
  time), *not* a calendar week — `competition.py` queries the WCA API with that
  range. The post title must show that actual window; only the idempotency key
  (`WeekStart`) is the calendar-week Monday. Don't conflate the two.
- Idempotency lives in the DB: `posts.report_week` is UNIQUE and the insert uses
  `ON CONFLICT (report_week) DO NOTHING`, so a second run in the same week is a
  no-op. `main.py --force` switches it to `DO UPDATE` (overwrite in place) and
  `--skip-email` bypasses SMTP — both are manual/test aids, never used by cron.
  `insert_weekly_report_post` deliberately **raises** (unlike the other upserts,
  which swallow) so a failed publish aborts the job loudly.
- Order in `main.py` matters: the post + records are written **before** the
  e-mail, so an SMTP failure (which now logs and re-raises → exit 1) cannot lose
  the week's report.
- Scheduling: no in-app scheduler — host cron calls
  `infra/cron/weekly-report.sh` (`ENV=test|prod`), which runs the containerized
  job via the `jobs` compose profile.
- Package-per-concern under `apps/scripts/weekly-report/`
  (`CompetitionModels/`, `database/`, `templates/`), orchestrated by `main.py`.
  Uses `logging`, exits non-zero on failure. Reads the shared Postgres read-only
  and sends email via SMTP. HTML templates under `templates/`.

---

## 7. Database schema & data flow

Migrations: `000001_init_schema`, `000002_add_post_image` (nullable
`posts.image_path`), `000003_add_weekly_report_post` (adds the `'report'`
`post_tag` enum value + the nullable `posts.report_week DATE UNIQUE`),
`000004_add_user_avatar` (nullable `users.avatar_url TEXT` — the WCA profile
picture URL), `000005_add_live_dnf_tracking` + `000006_add_live_results` (the
Budapest Special tables, see §18), `000007_rename_live_to_special` (renames
those three tables `live_*` → `special_*`; **column** names keep the `live_`
prefix on purpose — those are genuine WCA Live identifiers),
`000008_simplify_special_tracking` (drops `special_results` and every derived
column off `special_dnf_attempts` — official results are no longer stored, §18)
`000009_special_competition_per_viewer` (drops `special_competitions.active`
— the viewer picks the competition, so there is no active one, §18) and
`000010_add_allrounder_results` (the All-Rounder WCA results cache, see §19) and
`000011_add_allrounder_competition_dates` (nullable `start_date`/`end_date`/
`live_competition_id` on `allrounder_syncs` — this is what bounds when we open a
WCA Live connection for a leg, §19) and `000012_add_post_images` (replaces
`posts.image_path` with `posts.image_paths TEXT[]` — a post is a gallery now,
see §20; the single cover image is backfilled into a one-element array) and
`000013_add_prediction_contests` (the Tippverseny tables, see §21) and
`000014_add_prediction_text_question_type` (adds the `'text'` value to
`prediction_question_type` — a free-text extra question) and
`000015_users_require_wca_id` (renames any blank-`wca_id` row left behind by the
auth bug described in §9, then adds `CHECK (btrim(wca_id) <> '')` so that state
cannot recur at the storage layer).
Applied automatically on API startup via
`migrate.New("file:///migrations", cfg.DatabaseURL)` (the prod image bakes
`db/migrations` to `/migrations`).

```sql
post_tag ENUM ('event','competition','news','announcement','report')

posts(id UUID PK, title TEXT, tag post_tag, content TEXT,
      image_paths TEXT[] NOT NULL DEFAULT '{}',
                              -- 000002 + 000012: a poszt képei ('/uploads/<uuid>.<ext>')
                              -- a MEGJELENÍTÉS sorrendjében. Az első elem a borítókép,
                              -- a többi a poszt oldali carousel képe (§20). A sorrend
                              -- maga az adat — ezért tömb és nem külön tábla.
      report_week DATE NULL UNIQUE,  -- 000003: heti riport idempotencia-kulcsa
                                     -- (a hét hétfője); normál posztnál NULL,
                                     -- és a Postgres a NULL-okat külön veszi
      created_at TIMESTAMP default now(), updated_at TIMESTAMP default now())

users(id SERIAL PK, wca_id VARCHAR(50) UNIQUE, name, email,
      avatar_url TEXT NULL,   -- 000004: WCA profilkép URL-je (nullable)
      roles JSONB default '[]', competition_count INT, delegated_count INT,
      created_at, updated_at)

records(region VARCHAR, event_id VARCHAR, single INT, average INT,
        PRIMARY KEY (region, event_id))

-- 000005 + 000007 + 000009: Budapest Special (§18). TIMESTAMPTZ, mert a WCA
-- Live UTC pillanatokat ad, és élő verseny közben a zóna nem lehet kérdéses.
-- NINCS "aktív verseny" (000009 eldobta az active oszlopot): a versenyt a néző
-- választja kérésenként. Ez a tábla ma már CSAK a rögzített idők szülője —
-- sor kizárólag akkor keletkezik, ha ember rögzített DNF-időt (EnsureCompetition),
-- publikus kérésre soha.
special_competitions(competition_wca_id TEXT PK, live_competition_id INT,
        name, start_date, end_date, added_by → users(id), ...)

-- 000008: a HIVATALOS eredmények NINCSENEK tárolva (a special_results tábla
-- megszűnt) — a WCA Live-tól jönnek kérésre, és a memóriában élnek (§18).
-- Itt CSAK az van, amit a WCA Live nem tud megmondani: a DNF tényleges hossza.
-- Sor kizárólag akkor keletkezik, ha ember rögzített időt.
special_dnf_attempts(id UUID PK, competition_wca_id → special_competitions,
        live_result_id INT,        -- a WCA Live result id globálisan egyedi
        attempt_number INT,        -- 1-alapú
        elapsed_cs INT NOT NULL,   -- a solve TÉNYLEGES ideje — EMBERI megfigyelés
        elapsed_note, recorded_by → users(id), recorded_at,
        UNIQUE (live_result_id, attempt_number))   -- idempotencia-kulcs

-- 000010: Hungarian All-Rounder (§19). A WCA hivatalos eredményeinek
-- GYORSÍTÓTÁRA — nem forrás. A versenysorozat konfigurációja (melyik verseny
-- melyik évadhoz tartozik, pontozás) KÓDBAN van: internal/allrounder/series.go.
allrounder_syncs(competition_wca_id TEXT PK, fetched_at TIMESTAMPTZ,
        result_count INT, last_error TEXT NULL, ...)  -- versenyenkénti szinkron-állapot;
                                    -- eredmény nélküli (jövőbeli) versenynek is van sora

allrounder_results(competition_wca_id → allrounder_syncs, ordinal INT,
        wca_id, name, event_id, round_type_id, pos INT, best INT,
        PRIMARY KEY (competition_wca_id, ordinal))
-- ⚠️ Az `ordinal` a WCA API válaszTÖMBJÉBEN elfoglalt hely, és a pontszámítás
-- SORRENDFÜGGŐ (§19) — nem dekoráció, hanem a számítás bemenete.

-- 000013: Tippverseny (§21). Csak a LÉTREHOZÁS + LISTÁZÁS van kész — a tippek
-- leadása és a kiértékelés még nincs megépítve. A verseny NEVE/városa/dátuma
-- NINCS itt tárolva (nincs FK sem competition_wca_id-re) — azt a listázáskor
-- a competition csomag adja hozzá élőben a WCA API-ból, ugyanúgy, mint a
-- /competitions/upcoming végpontnál. A `status` (nyitott/lezárt) számított
-- érték a prediction_deadline-ból, nem oszlop.
prediction_contests(id UUID PK, competition_wca_id TEXT, prediction_deadline TIMESTAMPTZ,
        created_by → users(id), created_at, updated_at)

prediction_podium_events(contest_id → prediction_contests, event_id TEXT,
        PRIMARY KEY (contest_id, event_id))  -- amikhez dobogó-tippet (1-2-3. hely) kell leadni

prediction_question_type ENUM ('person','time','number','text')  -- 'text': 000014

prediction_extra_questions(id UUID PK, contest_id → prediction_contests,
        prompt TEXT, type prediction_question_type, ordinal INT)
        -- admin szabad kérdései, pl. "Hány rekord születik?" (number)
```

Notes / current state:
- `posts.id` is a `uuid.NewString()` generated in the service (not a DB default).
- `users.roles` is JSONB; login upsert **preserves existing roles** and only seeds
  `["user"]` on first insert (see `auth/repository.go` `Upsert` `ON CONFLICT`).
- `records` and `users.competition_count`/`delegated_count` are written and read
  only by the Python job (`SaveRecords`, `upsert_person`); the Go API does not
  touch them.
- Competition data is **not** persisted — it is proxied live from WCA per request.

### Adding a migration
Create the next numbered pair `db/migrations/00000N_<name>.{up,down}.sql`. It runs
on the next API start. Keep migrations schema-focused (no seed data — the DB starts
empty by design).

---

## 8. API architecture & endpoints

chi router in `internal/router/router.go`. Global middleware:
`middleware.SecurityHeaders` (`nosniff`) and CORS (allows `cfg.FrontendURL`,
`AllowCredentials: true` for the cookie). `middleware.JSON` (sets
`Content-Type: application/json`) wraps **only the JSON route group**; the
non-JSON routes — `/uploads/*`, the two SSE streams and the `/login` redirects —
are registered outside it, so they set their own Content-Type (§17). Every
admin route sits in one `Auth + RBAC(admin)` group.

| Method | Path | Auth | Handler |
| ------ | ---- | ---- | ------- |
| GET  | `/posts` | public | `post.GetPosts` |
| GET  | `/posts/{id}` | public | `post.GetPostByID` (404 if the id doesn't exist, 400 if it isn't a UUID — `posts.id` is a `uuid` column, so a malformed id would otherwise reach Postgres as a type error and come back as a 500) |
| GET  | `/posts/latest` | public | `post.GetLatestPosts` (limit 3) |
| POST | `/posts` | JWT + `admin` | `post.CreatePost` (**multipart/form-data only**: `tag`, `title`, `content` + up to `MaxImages` (10) repeated `images` file parts — their order is the gallery order) |
| PUT  | `/posts/{id}` | JWT + `admin` | `post.UpdatePost` (multipart; the **required** `image_order` describes the whole resulting gallery, see §20; 404 if the id doesn't exist) |
| GET  | `/uploads/*` | public | static file server over `cfg.UploadDir` (post images) |
| DELETE | `/posts` | JWT + `admin` | `post.DeletePosts` (body `{ids:[]}`; also deletes the posts' image files from the volume, see §20) |
| GET  | `/competitions/upcoming` | public | `competition.GetUpcoming` (HU only; DTO includes `eventIds` — the WCA API's `event_ids` passed straight through, no extra request) |
| GET  | `/competitions/{id}/psych-sheet/{eventId}` | public | `competition.GetPsychSheet` — proxy of the WCA psych sheet (PR-ranked start list) for one event: `[{name, wcaId, position}]`, `position` 0 = no PR yet, `wcaId` empty for newcomers. Ids validated against `^[A-Za-z0-9]{1,64}$` (400), an id the WCA doesn't know 404, upstream failure 502. **Cached 15 min + singleflight**, expired entries swept, see §21 |
| GET  | `/competitions/{id}/competitors` | public | `competition.GetCompetitors` — every accepted, competing registrant from the public WCIF, sorted by name: `[{name, wcaId}]`. (The WCA `/registrations` endpoint only carries user ids, no names — that's why WCIF.) Same validation/caching as above |
| GET  | `/predictions` | public | `prediction.GetContests` — every tippverseny, `competitionName`/`City`/`Country`/`StartDate` enriched live from `competition.Service.GetUpcomingHungarian` (not stored, see §21). `status` is computed from `predictionDeadline`, never `"evaluated"` yet |
| POST | `/predictions` | JWT + `admin` | `prediction.CreateContest` (JSON body `{competitionId, predictionDeadline (RFC3339), podiumEventIds[], extraQuestions: [{prompt, type: "person"\|"time"\|"number"\|"text"}]}`; 400 on a past deadline, an empty/duplicated/malformed `podiumEventIds` (max 20), or an invalid/blank question (max 20, prompt ≤ 300 chars, stored trimmed). Records the admin in `created_by`. Every validation error wraps `prediction.ErrValidation`, which is the one thing the handler checks for 400 — add new rules as `validationError`, not to a status chain) |
| DELETE | `/predictions/{id}` | JWT + `admin` | `prediction.DeleteContest` — deletes the contest row; `prediction_podium_events`/`prediction_extra_questions` cascade via their FK (`ON DELETE CASCADE`), no transaction needed. 404 if the id doesn't exist, 400 if it isn't a UUID |
| GET  | `/allrounder/standings` | public | `allrounder.GetStandings` — a whole season's point standings in one payload (rules + competitions + people). `?year=` is optional; a **missing, unknown or malformed** year falls back to the default season, so there is no error branch — the response's `year` says which one you got |
| POST | `/allrounder/refresh` | JWT + `admin` | `allrounder.Refresh` — re-fetches the season from the WCA API ignoring the cache TTLs. Admin-only because it generates upstream traffic |
| GET  | `/allrounder/stream` | public | `allrounder.StreamUpdates` (**SSE**: data-less `update` event whenever the running leg changes on WCA Live; `?year=` optional). Opens successfully even when no leg is running. An open stream is what keeps the WCA Live subscription alive (§19) |
| GET  | `/login` | public | `auth.Login` — starts the WCA OAuth flow: issues the CSRF nonce cookie and redirects to WCA. Takes `?redirect=<path>` (sanitized). **This is the only login entry point**; the browser never builds the authorize URL (§9) |
| GET  | `/login/callback` | public | `auth.LoginCallback` (WCA OAuth redirect target; verifies `state` against the nonce cookie before exchanging the code) |
| GET  | `/me` | JWT | `auth.Me` |
| POST | `/logout` | JWT | `auth.Logout` |
| GET  | `/special/competitions` | public | `special.GetCompetitions` — the competition picker. `?query=` searches WCA Live by **name** (any country); without it, the **Hungarian** competitions of the last 60 days plus the upcoming ones. Sorted ongoing → upcoming → past |
| GET  | `/special/report` | public | `special.GetReport` (per-competitor view, assembled from **live** WCA Live results + our recorded DNF times). **`?competition=<wcaId>` is required** (400 without it). Optional `?round=N` narrows to the Nth round of **every** event (1 → `333-r1`, `222-r1`, …); invalid values 400, WCA Live unreachable 502, watch capacity exhausted 503 |
| GET  | `/special/stream` | public | `special.StreamUpdates` (**SSE**: `update` event on every change of `?competition=<wcaId>`, required). An open stream is also what keeps *that competition's* WCA Live subscription alive (§18) |
| PUT  | `/special/results/{liveResultId}/attempts/{attemptNumber}/elapsed` | JWT + `admin` | `special.RecordElapsed` (body `{competitionWcaId, elapsedCs, note}`; `elapsedCs: null` **deletes** the entry). Keyed by WCA Live ids, not our own — there is no row for an un-timed DNF |

**Reading `/special/*` is public; only recording a DNF duration is admin.** The
Budapest Special is a public leaderboard — the same competitor results are
already public on WCA Live, and the recorded DNF durations exist precisely so
competitors can see them. **Choosing the competition is NOT an admin action any
more:** any visitor picks one and the page goes live for them. Only writing the
human-observed duration is competition operation, hence admin-only. (Both this
and the earlier admin-only-everything decision were reversed deliberately —
don't "restore" either.)

⚠️ The WCA competition id comes from the client on every request and ends up in
the WCA Live `/link/` URL, so `Service.ResolveCompetition` validates it against
`^[A-Za-z0-9]{1,64}$` and always resolves the numeric WCA Live id **itself**
(cached for an hour). Never take `liveCompetitionId` from the client.

Integrations: **WCA API** — OAuth token + `/api/v0/me` (auth), and
`/api/v0/competitions` (competition proxy, paginated, `maxPages=20` guard, 10s
client timeout).

---

## 9. Authentication & authorization

**The flow starts on the API, not in the browser.** `GET /login` is the single
entry point; the frontend only navigates to it. Two reasons, both load-bearing:
the OAuth `state` must be an unguessable server-issued nonce kept in an
`HttpOnly` cookie (a browser-built URL cannot do that), and `client_id` /
`redirect_uri` / `scope` then live in exactly one place instead of drifting
between `NEXT_PUBLIC_WCA_*` and `WCA_*`.

Flow:
1. `LoginButton` navigates to `clientApiUrl() + "/login"` with
   `?redirect=<pathname>`.
2. `auth.Login` sanitizes that path, mints a 32-byte random nonce, stores
   `<nonce>:<path>` in the short-lived `oauth_state` cookie, and redirects to WCA
   `oauth/authorize` with `scope=public email` and `state=<nonce>`.
3. WCA redirects back to the API `/login/callback?code=...&state=...`.
4. `auth.LoginCallback` **verifies `state` against the cookie before touching the
   code**, exchanges the code for a WCA access token, calls WCA `/me`, **upserts**
   the user (including the WCA `avatar.url` → `users.avatar_url`), mints an HS256
   JWT (`middleware.UserClaims`: user_id, wca_id, name, avatarURL, roles; iat/nbf
   + 1-hour exp), and sets it as cookie `auth`. `/me` echoes the claim's `avatarURL`.
5. Cookie is `HttpOnly`, `SameSite=Strict`, `Path=/`, `MaxAge=3600`,
   `Secure` unless `DEPLOYMENT=development`. Then redirects to `FRONTEND_URL` +
   the path **taken from the cookie**. Any failure (bad state, missing code,
   upstream error) logs server-side and redirects to `FRONTEND_URL/?login=failed`
   — no upstream error text reaches the browser.

⚠️ Invariants that are easy to break (see `internal/auth/state.go`, and the tests
that pin each one):

- **The WCA response status codes MUST be checked** (`exchangeCode` / `fetchMe`).
  A 4xx body decodes "successfully" into an empty struct, so skipping the check
  let the whole flow run to completion and issue a **valid session cookie for a
  request that never authenticated** — an empty `wca_id` "ghost user" shared by
  every attacker. An empty `access_token` or `wca_id` is likewise fatal.
  Migration `000015` adds a `CHECK (btrim(wca_id) <> '')` so the storage layer
  can't hold that state either.
- **`oauth_state` is `SameSite=Lax`, not `Strict`.** The return from WCA is a
  cross-site top-level navigation; a Strict cookie would not be sent back and
  every login would fail.
- **`oauth_state` uses `Path=/`.** In test/prod Caddy strips the `/api` prefix,
  so the browser-visible path is `/api/login/callback` while the API sees
  `/login/callback`; a `/login`-scoped cookie would work in dev and never come
  back in prod.
- **The return path travels in the cookie, not in `state`.** `state` round-trips
  through the user agent and is attacker-controlled; the cookie is ours. It is
  sanitized on both ends anyway (`sanitizeRedirectPath` — allow-list, so
  `//evil`, `/\evil`, CR/LF and cookie-breaking characters all fall back to `/`).
- **Clear the state cookie BEFORE writing the response.** Every branch ends in
  `http.Redirect`, which calls `WriteHeader`; a `defer`-red `SetCookie` would
  silently never reach the browser.
- **`scope` is `public email`.** The WCA token is used for exactly one `/me` call
  and then dropped. `manage_competitions` was removed — don't restore it without
  an actual endpoint that needs it.

Authorization:
- `middleware.AuthMiddlewareWithConfig(cfg)` reads the `auth` cookie, validates the
  JWT, and puts `*UserClaims` in request context. Retrieve with
  `middleware.GetUserFromContext(r)`. It rejects non-HMAC algs **twice** (the
  manual `SigningMethodHMAC` check plus `jwt.WithValidMethods`), and passes
  `jwt.WithExpirationRequired()` — golang-jwt treats a token with **no** `exp` as
  valid, which would be a never-expiring session.
- `middleware.RBACMiddleware(types.ROLE_ADMIN, ...)` requires the user to have one
  of the given roles; returns 401 if unauthenticated, 403 if lacking the role.
- ⚠️ **`middleware.UserClaims` is the single definition of the token payload** —
  `auth` signs with it and the middleware parses with it. There used to be a
  second, identical `auth.UserClaims`, where a mismatched `json` tag silently
  dropped a field across the round trip (this happened once with `avatarURL`).
  Don't reintroduce a second copy.

---

## 10. Environment variables & configuration

Templates: `.env.dev.example`, `.env.test.example`, `.env.prod.example` at repo
root. Never commit real `.env*` files (git-ignored). Backend config is loaded and
**validated fail-fast** in `internal/config/config.go`.

**Required (API fails to start without these):** `DATABASE_URL`, `JWT_SECRET`
(≥32 chars — generate with `openssl rand -hex 32`), `WCA_CLIENT_ID`,
`WCA_CLIENT_SECRET`, `WCA_REDIRECT_URI`, `FRONTEND_URL`, `DEPLOYMENT`
(must be exactly `development`, `test` or `production`).

⚠️ **`DEPLOYMENT` and `FRONTEND_URL` deliberately have no default.** They used to
default to `development` / `http://localhost:3000`, which meant their entry in the
"required" map could never trigger: a missing `DEPLOYMENT` in production did not
stop the API, it silently produced **non-`Secure` auth cookies**, and a missing
`FRONTEND_URL` opened CORS to localhost and redirected there after login. The
value set is validated too, so a typo (`produciton`) fails fast instead of
quietly behaving like development. Don't add defaults back.

Other important vars:
- DB: the API uses **only `DATABASE_URL`** — for the connection pool
  (`db.NewPostgres`) and for golang-migrate alike. `DB_HOST`, `DB_PORT`,
  `DB_USER`, `DB_PASSWORD`, `DB_NAME` feed the `postgres` image, the weekly-report
  job and the `DATABASE_URL` interpolation; the api service doesn't receive them.
  **They are the single source of truth for the DB credentials.** The env
  files no longer also define `POSTGRES_USER`/`POSTGRES_PASSWORD`/`POSTGRES_DB`
  — `docker-compose.yml`'s `db` service sources those (which the `postgres`
  image itself requires under that name) from the same `DB_USER`/`DB_PASSWORD`/
  `DB_NAME`, i.e. `POSTGRES_USER: ${DB_USER}` etc. And `DATABASE_URL` in each
  env file is built from these via Compose's env-file variable interpolation
  (`postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable`)
  rather than being typed out separately. **Change the DB user/password/name in
  exactly one place** (`DB_USER`/`DB_PASSWORD`/`DB_NAME` in the env file) — don't
  reintroduce a parallel `POSTGRES_*` block or hand-type `DATABASE_URL`, both of
  which previously had to be kept in sync by hand and could silently drift.
- `DEPLOYMENT` (`development`/`test`/`production`) — controls cookie `Secure`
  and whether `X-Forwarded-For` is trusted. `NODE_ENV` (web only). There is no
  `LOG_LEVEL`: nothing ever read it, so it was removed.
- `UPLOAD_DIR` — where post images are written (default `/data/uploads`). Not
  required in any env file; the base compose mounts the `uploads_data` named
  volume there. Deliberately **not** under `/app`, because the dev override
  bind-mounts the source tree over `/app` and would shadow it.
- `WCA_LIVE_URL` — WCA Live base URL (default
  `https://live.worldcubeassociation.org`); both the GraphQL endpoint and the
  Phoenix socket are derived from it. Not required; mapped in the base compose
  with an empty default. (There is no `WCA_LIVE_WATCHER` kill switch any more —
  the subscription only exists while someone has the page open, §18.)
- Web: `SERVER_API_URL` (server→api, Docker service name), `NEXT_PUBLIC_API_URL`
  (browser→api; the login entry point is simply `NEXT_PUBLIC_API_URL + /login`).
  **`NEXT_PUBLIC_*` are baked into the browser bundle at build time** (Docker build
  args), so changing them requires a web rebuild. `NEXT_PUBLIC_WCA_CLIENT_ID`,
  `NEXT_PUBLIC_WCA_REDIRECT_URI` and `NEXT_PUBLIC_BACKEND_LOGIN` were
  **removed**: the browser no longer builds the WCA authorize URL (§9), so the
  OAuth client config lives only in `WCA_*` on the API side, and the login URL
  is derived from the API URL. Don't reintroduce them — two sources for the same
  value drift.
- Compose: `COMPOSE_PROJECT_NAME`, `COMPOSE_PROFILES=proxy` (test/prod),
  `WEB_DOMAIN`/`ACME_EMAIL` (Caddy — single domain, API path-routed under
  `WEB_DOMAIN/api`), `WEB_PORT`/`API_PORT`/
  `DB_PORT_HOST` (dev host ports). The `weekly-report` service sits behind the
  `jobs` profile (run it with `--profile jobs run --rm weekly-report`); its
  `EMAIL_FROM`/`EMAIL_TO`/`EMAIL_APP_PASSWORD` now live in the root
  `.env.<env>` templates too (defaulted to empty in the base compose so the main
  stack stays warning-free) — the job validates them fail-fast at runtime.
- Weekly-report job (Python): **uses the same root `.env.<ENV>` file as the rest
  of the stack — no per-app `.env`.** `config.py` reads `ENV` (default `dev`) and
  `load_dotenv`s the repo-root `.env.<ENV>`, with fail-fast validation mirroring
  `apps/api/internal/config/config.go`. It runs **either** standalone (venv on the
  host) **or** as a compose service behind the `jobs` profile. ⚠️ DB host caveat:
  the root file targets the Docker network (`DB_HOST=db`, `DATABASE_URL=...@db`),
  which only resolves inside a container. `config.py` detects whether it runs in a
  container (`/.dockerenv`): in a container it uses the root file's `DB_HOST`/
  `DATABASE_URL` as-is; **on the host it ignores `DATABASE_URL` and connects to
  `localhost`** (the dev compose exposes the DB there). In the container no env
  file is copied in (`.dockerignore` + build context is the job dir, not the repo
  root), so `load_dotenv` no-ops and the values arrive from the compose
  `--env-file`. All secrets come from the environment; **never hardcode them in
  `config.py` or any tracked file**. Point `EMAIL_TO` at a fixed test inbox in
  test/staging. Not scheduled by default — host cron runs
  `infra/cron/weekly-report.sh` (`ENV=test|prod`).

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
4. Register the route in `router.New` on the handler already passed in; wrap in
   the auth/RBAC `r.Group` if protected.
5. Add a table-driven / testify test next to the file (`*_test.go`). Mock the repo
   with testify `mock` (see `auth/service_test.go`) or `go-sqlmock` for SQL.

### Add a new backend feature package
Mirror `internal/post/`: create `internal/<feature>/{model,repository,service,handler}.go`
with the same constructor + interface shape, add any shared enum to
`internal/types/`, wire the repo → service → handler in `cmd/server/main.go`, and
pass the handler into `router.New` (add it as a new parameter) to register routes.

### Add a frontend route
Create `apps/web/app/<route>/page.tsx` (+ colocated `components/`, `lib/`,
`types.ts`). Fetch server-side with `serverJson` where possible; use
`"use client"` + `clientFetch`/`clientJson` for interactive/browser fetches. Add
nav links in `app/lib/site.ts` if needed. **Read the Next 16 docs first** (§3 warning).

---

## 12. Known constraints, assumptions & tech debt

- **Untracked working tree:** `apps/api`, `apps/web`, and other files show as
  untracked/added in git status — the project is mid-setup; commit history is thin.
- ~~`apps/web/app/layout.tsx` still has the default `create-next-app` metadata~~
  — **resolved**: real title template + description + OpenGraph, and `lang="hu"`.
- The navbar/footer link to `/posts` is now labelled **"Hírek"** (it was "Blog",
  which matched neither the route nor the page's own "Hírek" heading). Nav labels
  are Hungarian throughout and live in `app/lib/site.ts`.
- The contact form (`app/contact/components/ContactForm.tsx`) is still **not
  wired to any backend** — it renders and validates nothing, submitting does
  nothing. Needs a route handler or a mail service.
- **Tippverseny (`internal/prediction` + `/tools/prediction`) only covers
  contest CREATION, LISTING and DELETION so far — see §21.** Leaderboards/
  participant counts/answer submission are not built yet; `participantCount`
  is hardcoded to 0 and `status` only ever reports `open`/`closed` (never
  `evaluated`, though the frontend type still has that state for later).
- `/results`, `/rankings`, `/guide`, `/team`, `/registration` are linked from
  `QuickToolsSection` but **no such routes exist yet** (they 404).
- ~~The Hungarian All-Rounder tool card links out to `allrounderpoints.vercel.app`~~
  — **resolved**: the app is integrated at `/tools/allrounder`, see §19.
- ~~`auth/service.go` / `auth/handler.go` bypass the `config` struct and read env
  directly~~ — **resolved**: both take `*config.Config` (wired in `main.go`), so
  there are no `os.Getenv` calls left in the `auth` package (§9).
- No structured logging/observability beyond stdlib `log`.
- ~~No rate limiting anywhere, and no CSP~~ — **resolved**, see §14.
- `apps/api/tmp/` (air build output) and `__pycache__/` dirs are build artifacts.
- No Redis / message queue / object storage in the stack — the only data store is
  Postgres. Don't assume a cache/queue layer exists.
- ⚠️ **On the current staging host (`server`), the internet-facing Caddy is
  NOT the compose-managed `caddy` service from `infra/docker-compose.yml`.**
  It's a separate, manually-started `caddy` container (image `caddy:latest`,
  container name literally `caddy`, config bind-mounted from
  `~/caddy/Caddyfile` on the host — **not tracked in this repo**)
  that also proxies an unrelated `other.example.org` domain (Jellyfin). It
  exists because the compose `caddy` service can't bind :80/:443 — something
  already holds those ports on this host — so `up -d` for the `proxy` profile
  fails with "port is already allocated" and the compose `caddy` service stays
  `Created`/`Exited`, harmlessly. Two consequences if the API/web containers
  are ever recreated with a different Compose project name (i.e.
  `COMPOSE_PROJECT_NAME`, which sets the container-name prefix and the
  `<project>_appnet` network name) — this already happened once
  (`schu_test` → `speedcubinghungary_test`) and caused a same-day 502 on both
  login and `/competitions/upcoming`, because both proxy through this
  container:
  1. `~/caddy/Caddyfile`'s `reverse_proxy` targets hardcode the
     **container names** (not service names — this Caddy isn't in the compose
     network's service-DNS scope) and must be updated by hand to match.
  2. The `caddy` container must be connected to the **new** `<project>_appnet`
     network by hand (`docker network connect <project>_appnet caddy`) — it
     does not join it automatically, and without it DNS lookups for the
     container names fail (`no such host`) even with a correct Caddyfile.
  A `caddy reload` can report `"config is unchanged"` and silently keep
  serving the old config if the reload races the file edit — `docker restart
  caddy` is the reliable way to pick up a Caddyfile change here. Prefer fixing
  the underlying port conflict and switching back to the compose-managed
  `caddy` service over perpetuating this manual one, but until that happens,
  check both of the above after any staging redeploy that changes the project
  name or recreates `api`/`web`.
- `db/postgres.go` uses `sslmode=disable`; acceptable because the DB is only on the
  internal Docker network (never published in test/prod). Make it configurable if
  the DB is ever exposed.
- The weekly-report job previously hardcoded the Gmail app password + DB creds in
  a committed-looking `env.py` module. It's now `config.py`, which reads the shared
  root `.env.<ENV>` (see §10) — secret-free and env-driven; never hardcode values
  back into `config.py`.
- **One env file per environment, no per-app `.env`.** The only real env files are
  the root `.env.dev`/`.env.test`/`.env.prod` (git-ignored), loaded via
  `--env-file`; the tracked templates are `.env.<env>.example`. The Python job
  reads the *same* root file (see §10). The old per-directory leftovers (root
  `.env`, `apps/api/.env`, `apps/web/.env`, `db/.env`, `apps/web/.env.example`,
  `apps/scripts/weekly-report/.env[.example]`) have been removed — **don't
  reintroduce per-service env files.**

---

## 13. Performance considerations

- `/competitions/upcoming` paginates the WCA listing (up to 20 pages, 10s HTTP
  timeout) and is **cached for 15 minutes** in `competition.Service`, like the
  psych sheets and registrant lists — the home page, the calendar and the
  prediction modal all read it.
- All-Rounder results are the opposite case and a good precedent: a whole
  season's WCA results are cached in Postgres with per-competition TTLs, and the
  scoring runs server-side. Cold season ~2.5 s, warm ~50 ms. See §19.
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
- **Rate limiting is per-IP ONLY on `/login` + `/login/callback`**
  (`middleware.RateLimit(cfg, LoginRateLimit)`, 60/min — a login is two
  requests). ⚠️ **Never put a global or per-route per-IP limit on anything the
  pages use.** Two reasons, both learned the hard way: (1) Server Components
  call the API from the web container over the Docker network with no
  `X-Forwarded-For`, so *every visitor's* SSR request shares one key — the web
  container's IP — and when that bucket fills `serverJson` silently returns
  its fallback, i.e. empty news/posts/standings with no error; (2) the live
  tools are used at the venue, where everyone is behind one Wi-Fi IP and the
  Special page refetches the report per viewer on every live update.
  **WCA protection lives on the outbound side instead:**
  `competition.WCARepository` holds a shared semaphore (`maxConcurrentUpstream`,
  6 in flight, 15 s queue timeout, 16 MiB body cap), which bounds what we send
  the WCA regardless of who asks or from where — including deliberate cache
  misses with made-up ids. The limiter is in-memory and per-process (no Redis,
  §12). The client IP is the **last** `X-Forwarded-For` entry (Caddy appends
  what it actually saw) and the header is only trusted outside development.
- **CSP and the other response headers** live in `apps/web/next.config.ts`
  (`headers()`), plus `X-Content-Type-Options: nosniff` from
  `middleware.SecurityHeaders` on the API, which matters for the
  attacker-influenced bytes served at `/uploads/*`. ⚠️ Two things to know:
  `headers()` is evaluated at **build time**, so the API origin in
  `connect-src` **and `img-src`** (post images come from the API's
  `/uploads/*`) is baked in from `NEXT_PUBLIC_API_URL` like every other
  `NEXT_PUBLIC_*` (§10) — any new directive that reaches the API needs it too;
  and `script-src` deliberately keeps `'unsafe-inline'`, because the strict
  nonce variant needs a `proxy.ts` and forces **every** page to render
  dynamically — that is a separate architectural decision, not a side effect of
  adding a header.
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
  constructor, wired manually in `cmd/server/main.go`; `router.New` only routes.
- HTTP handler shape: decode into a local anonymous request struct → call service →
  `json.NewEncoder(w).Encode(...)` or `httperr.Write(w, r, status, err)` (§6).
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
- Real `.env.dev`, `.env.test`, `.env.prod` (root only) — secrets, git-ignored;
  edit the `.example` templates instead when documenting new vars. No per-app
  `.env` files — everything reads the root file.
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
- ⚠️ `middleware.JSON` sets `Content-Type: application/json`, and Go's
  `ServeContent` does **not** override an already-set Content-Type. That is why
  it wraps only the JSON route group in `router.New`: register a non-JSON route
  (files, SSE, redirects) **outside** that group, otherwise it goes out
  mislabeled as JSON. `router_test.go` pins this for `/uploads/*`.
- Uploaded images: the MIME type is detected from the file **contents**
  (`http.DetectContentType`), never from the client's Content-Type header, and the
  stored filename is a generated UUID — the client's filename is discarded to
  prevent path traversal. Limit is 5 MiB (`post.MaxImageSize`).
- ⚠️ The prod API image's `USER appuser` (uid 1001) can't write into a **freshly
  created** `uploads_data` named volume — Docker creates new named volumes owned
  by `root:root`, so the very first upload on a new environment fails with
  `open /data/uploads/...: permission denied`. Fixed via
  `infra/docker/api/entrypoint.sh`: the container starts as **root**, `chown -R`s
  `UPLOAD_DIR` to `appuser`, then `su-exec`s down to `appuser` to run the server
  (`Dockerfile` stays `ENTRYPOINT`, no more `USER appuser` line/`CMD`). This makes
  every fresh deploy (new server, new volume) self-healing — don't revert to a
  bare `USER appuser` + `CMD` without re-solving this.
- ⚠️ **Per-Dockerfile `.dockerignore` files can exclude a file the same
  Dockerfile `COPY`s.** `infra/docker/api/Dockerfile.dockerignore` blanket-excludes
  `infra` (to keep the API's build context — repo root — lean), but the Dockerfile
  also `COPY infra/docker/api/entrypoint.sh /entrypoint.sh` (added by the fix
  above, in a later commit that didn't touch the ignore file). The build only
  failed once a clean host actually had to build that layer fresh instead of
  hitting cache — `"/infra/docker/api/entrypoint.sh": not found"`. Fixed with a
  negation line, `!infra/docker/api/entrypoint.sh`, placed **right after** the
  blanket `infra` exclusion. If you add a new `COPY` from a directory that a
  `Dockerfile.dockerignore` excludes wholesale, you must add a matching `!`
  negation line, or a fresh (uncached) build will fail the same way. Test with
  `docker build -f infra/docker/api/Dockerfile --target prod .` — a cached build
  can hide this for a long time.
- Constructor naming is inconsistent across packages (`NewPostgresRepository` vs
  `postgresRepository`, `Service` interface in `auth` vs concrete struct in `post`).
  When editing a package, match **that package's** existing style.

---

## 18. Budapest Special (`internal/special`)

The public leaderboard for "Special"-format competitions, where a competitor's
score is the **sum of their times across every event**. That format needs the
one number the official results deliberately omit: how long a DNF attempt
actually took. Verified against the wca-live source (commit `f44ced1`,
2026-07-31) and probed against production.

⚠️ Naming: inside this package, **"live" always means the upstream WCA Live
service** — never our feature. That is why `LiveClient`, `SocketClient`, and the
`live_result_id` / `live_competition_id` **columns** keep the name, while the
package, tables and routes are `special`.

**The viewer picks the competition, per request.** There is no "active"/"tracked"
competition in the database and no admin selection step: `/special/report` and
`/special/stream` both take `?competition=<wcaId>`, and the page keeps that
choice in the URL. Several competitions can therefore be open at once, in
different browsers — the tracker, the broadcaster and the subscriptions are all
keyed by competition. (Earlier the admin set one active competition; migration
`000009` removed the flag. Don't reintroduce a server-side "current"
competition.)

Consequences worth keeping in mind:

- Any visitor can make us open a WCA Live session for any competition, so
  `maxTrackedCompetitions` (20) caps concurrent tracking; idle competitions are
  evicted after `idleRetention` (5 min), and when only viewed ones remain the
  request fails with `ErrTooManyCompetitions` (503) rather than growing.
- `special_competitions` rows are created **only** by `EnsureCompetition` when an
  admin records a duration — never by a public read, so browsing cannot fill the
  table.

### ⚠️ The single most important fact: WCA Live cannot tell you how long a DNF took

Do not spend time looking for this again. `lib/wca_live/scoretaking/attempt.ex`
is the *entire* attempt model:

```elixir
embedded_schema do
  field :result, :integer
  field :reconstruction, :string
end
```

There is **no attempt start/finish timestamp, no elapsed time, no timer state,
no live timer** — anywhere in WCA Live. Worse, `Result.changeset` →
`apply_time_limit/2` rewrites an over-the-limit time to DNF (`-1`) **before**
insert, so the entered time never reaches the database. The only timestamp is
`Result.entered_at`: **data-entry time**, second precision, shared by all five
attempts of one result — unrelated to when any solve happened. Never present it
as a solve time.

Yet WCA Regulation **A1a3** says the time counting toward a cumulative limit for
a DNF attempt is *"the time elapsed in the solve"* — a number the official data
model deliberately does not carry.

### The one thing we store

**The official results are NOT stored.** They belong to WCA Live; a local copy
would be a pure round-trip (`snapshotResults` → INSERT → SELECT → unchanged).
They are fetched live and held **in memory** by `Tracker` (tracker.go).

The **only** persisted data is what WCA Live cannot tell us: the actual duration
of a DNF, keyed by `(live_result_id, attempt_number)`. That is the whole
`special_dnf_attempts` table, and it is written **only when a human records a
time** — there are no rows for un-timed DNFs.

That single decision is why this package has no detection step, no
reconciliation, no `resolved_at`, and no backfill problem:

- The list of DNFs comes from the *current* WCA Live data, so it is never stale.
- A scoretaker fixing a typo means the attempt is no longer `-1`, so our row
  simply stops joining. Nothing to resolve, and the audit row survives.
- `max_elapsed_cs`, `completed_cs`, competitor names, event/round ids were all
  derived — they are recomputed per request instead (`attemptBound`).

**Don't reintroduce a results mirror or a background detector.** Migration
`000008_simplify_special_tracking` removed both deliberately.

### Transport (verified, not guessed)

| | |
| --- | --- |
| GraphQL | `POST https://live.worldcubeassociation.org/api` — public, no auth |
| Subscriptions | `wss://…/socket/websocket?vsn=2.0.0` — **Phoenix channels + Absinthe** (`absinthe_phoenix` 2.0.3), *not* graphql-ws |
| Handshake | `phx_join` on `__absinthe__:control` → `doc` push → `{"subscriptionId": "__absinthe__:doc:global:…"}` → data arrives on that topic as `subscription:data` |
| Frame format | Phoenix v2 array: `[join_ref, ref, topic, event, payload]` (`join_ref`/`ref` are `null` on broadcasts) |
| Auth | none — `UserSocket.connect/3` accepts every connection. Send **no** `Origin` header (Phoenix `check_origin` passes when it is absent). |
| Heartbeat | required every <60s on topic `phoenix`, event `heartbeat` |
| Subscription | there is exactly **one**: `roundUpdated(id: ID!): Round!` |
| Unsubscribe | `unsubscribe` on `__absinthe__:control` with `{"subscriptionId": …}` — dropping the key locally is **not** unsubscribing; without this frame WCA Live keeps pushing that round for the life of the connection |
| WCA ID → live ID | no GraphQL lookup exists; follow the `302` from `/link/competitions/{wcaId}` → `/competitions/{numericId}` |

⚠️ **You cannot fetch a whole competition's results in one query.** Adding
`results { … }` to `competitionEvents.rounds` is rejected by WCA Live's query
complexity limit (measured: complexity 7331, maximum 5000). Rounds must be
fetched one `round(id:)` query each — `Tracker.load` does that in parallel
(`roundFetchConcurrency = 4`). Measured against production: 10 rounds, 286
results, ~1.2s.

### The tracker's lifecycle: viewer-scoped, not a background job

There is **no start/stop watching system and no background watcher**. A viewer
opens the page for a competition; the data is fetched because someone is looking
at it:

- The tracker holds a **map of `trackedCompetition` entries** (competition wcaId
  → rounds + viewer count), not a single competition.
- `Tracker.Results(ctx, competition)` serves that entry's in-memory snapshot,
  re-fetching when it is not live-tracked and older than `cacheTTL` (15s). A
  **per-competition `loadMu`** keeps concurrent requests from multiplying
  upstream traffic.
- `Tracker.AddViewer(competition)` is **reference-counted by open SSE
  connections**, per competition *and* globally. The first viewer anywhere starts
  the session (one websocket); each competition's `roundUpdated` subscriptions
  are added while it has viewers and dropped when it loses its last one. The
  last viewer overall stops the session, so we hold no connection to WCA Live
  when nobody is watching.
- **One socket serves every competition.** `SocketClient` multiplexes over
  Phoenix channels, and WCA Live round ids are globally unique, so the
  `round:<id>` subscription key cannot collide between competitions. Don't
  create a socket per competition.
- `entry.trackedSince` exists so a snapshot cached *before* live tracking began
  is not mistaken for a self-updating one; only a load newer than that counts as
  live.
- `refreshInterval` (2 min) only re-checks whether new rounds opened, for the
  competitions that currently have viewers.

### ⚠️ The session lifecycle is SHARED with `allrounder` — don't copy it back

The socket lifecycle (viewer refcount, start/stop, reconnect resync,
subscription diffing) lives in **`wcalive.Session`** and is used by both
trackers. It used to be ~250 near-identical lines in each package, and the two
copies had already drifted.

The feature implements `wcalive.Watcher` — `Reload(ctx)`,
`WantedSubscriptions()`, `HandleEvent(event)` — and keeps only what genuinely
differs: **which** competitions are tracked (a map here, one leg in
`allrounder`), the GraphQL queries, and the snapshot shape. Rules the shared
layer encodes, each pinned by a test in `wcalive/session_test.go`:

- **The connection exists only while a viewer does.** `Session.AddViewer()`
  refcounts; the first viewer starts the socket, the last stops it.
- **`AddViewer` returns `started`, and the caller must use it.** The starter's
  work is covered by the session's own opening `Reload`; anyone joining a
  running session has to load what *they* watch. Deciding this from a separate
  "is it running?" call is a race — two simultaneous viewers can both conclude
  they don't need to load.
- **Sessions must not overlap.** `SocketClient` resets its per-connection state
  (`sentKeys`, `topicKey`) per session, so a new session waits on the previous
  one's `done`, and the session waits for `socket.Run` to return.
- **`SetOnConnect` is registered once, in `Session.Watch`** — not per session,
  which would race with the socket's own goroutine.
- **The `Session` never calls the `Watcher` while holding its own mutex**, and
  `Running()` is atomic — so a feature may ask "is the session live?" from
  inside its own lock without a lock-ordering hazard.

### Why the design is state-based

**Every `roundUpdated` event is a full Round snapshot, not a delta.** That single
fact makes duplicate, out-of-order and missed events self-healing, and is why
there is no event log, no sequence numbers and no dedup cache. Keep it that way.

Other invariants worth preserving:

- **Only `-1` is a DNF.** `-2` (DNS — never started, no elapsed time exists) and
  `0` (skipped/cutoff) never get an input. `333fm`/`333mbf` have no time limit,
  so no bound.
- `max_elapsed_cs` is a derived **upper bound** (per-attempt limit → the limit;
  cumulative → limit − prior positive attempts). It is a bound, not a
  measurement — label it as such in the UI.
- **The tracker loads every `open` round — including already-`finished` ones**
  (`openRounds` is just `round.Open`). This is deliberately *wider* than WCA
  Live's own client (`open && (!finished || active)`), and must stay that way: a
  Special score sums **all** events, but the page is typically opened
  mid-competition, and the morning's 2x2 is already finished.
- Round parameters (`timeLimit`) are cached per tracked round because the
  subscription payload carries only results.
- On every (re)connect the tracker **re-queries every round of every watched
  competition in full**, so nothing is lost while the socket was down. After an
  API restart there is nothing to recover: the open pages reconnect their SSE
  streams and that is what re-establishes the tracking.
- **WCA Live results have no group dimension.** Groups exist only in the WCIF
  schedule (`ActivityCode` `333-r1-g2`). Separation is by round + competitor +
  attempt number, which is strictly finer than per-group.
- Never write observations back into WCIF/WCA. The duration is not part of the
  official result.

### The competitor view

`BuildCompetitors` (competitor.go) merges the live snapshots with our recorded
durations into per-person cards (all attempts, best/average, used and remaining
time). Rules:

- **Cumulative budgets are keyed by the `cumulativeRoundIds` SET, not by round.**
  A "Budapest Special"-style limit spanning several events is therefore ONE
  shared budget, which is more correct than WCA Live's own per-round handling.
  `buildBudgets` is the single place that encodes Regulation A1a3 — do not add a
  second cumulative calculation.
- `remainingCs` is only exact once every DNF has a recorded duration;
  `missingDnfCount > 0` means it is an **upper estimate**, and the UI says so.
- **The round picker filters server-side, on purpose.** `GetReport(ctx, roundNumber)`
  narrows the snapshots *before* calling `BuildCompetitors`, so used/remaining
  time and the cumulative budget still come from the one `buildBudgets`
  implementation. Do **not** move this filter into the client — that would
  require a second A1a3 calculation in TypeScript. `availableRounds` is always
  computed from the **unfiltered** set, otherwise the picker would collapse to
  the selected round. ⚠️ A cumulative budget spanning round numbers (e.g.
  `["333-r1","333-r2"]`) shows only partial usage when filtered; the realistic
  Special setup is one budget per round across events, so this hasn't come up.

### Live updates to the browser

`Broadcaster` (broadcaster.go) + SSE at `/special/stream?competition=…`. The
event carries **no data**, only "something changed"; the client refetches the
report. That keeps one source of truth and makes dropped or duplicated
notifications harmless. The broadcaster **is keyed by competition** — several
competitions can be open at once, and one's update must not make the other's
viewers refetch. The stream route sits outside the JSON middleware group (§17).

The SSE handler is also what tells the tracker a viewer is present
(`AddViewer`), so the WCA Live subscription is tied to the page being open. The
tracker and the HTTP layer must share one `Broadcaster`, which is why
`router.New` takes a `*special.Handler` built in `main.go`.

Frontend: **`/tools/special`** (colocated `components/lib/types`), person cards
with search; SSE-driven with a 30s poll only as a fallback. Flow:

- With no `?competition=`, `SpecialView` renders `CompetitionPicker`: Hungarian
  competitions by default, debounced free-text search over all of WCA Live,
  "Most fut" badge on ongoing ones.
- Picking one does `router.replace("?competition=…")` — **the URL is the single
  source of truth for the selection**, so the view is shareable and survives a
  reload. `page.tsx` wraps the view in `<Suspense>` because `useSearchParams`
  requires it in a statically rendered route (Next 16).
- Round selection is stored **paired with the competition**
  (`{competition, round}`) and read back only when it matches, instead of being
  reset in an effect — the React Compiler lint (`react-hooks/set-state-in-effect`)
  rejects `setState` in an effect body. Loading/error states are derived the same
  way (`report.competition.competitionWcaId === competitionId`,
  `failedCompetition`), not stored as a separate status.
- The page is **public**; `SpecialView` checks `useUser()` for the `admin` role
  and only then turns the DNF cells into editable inputs. Non-admins see the
  recorded durations as read-only text. The server enforces this independently
  (§8) — the client check is UX only.

### Verifying against reality

`sqlmock` does not execute SQL and `psql PREPARE` declares parameter types, so
both can hide type-inference bugs (a `CASE ... THEN NULL ELSE $4 END` was
inferred as `text` against an `integer` column). Four opt-in smoke tests cover
this — `TestSmoke_AgainstRealWCALive` (wire format),
`TestSmokeTracker_LoadsWholeCompetition` (the whole-competition fetch and the
complexity limit), `TestSmokeCompetitions_AgainstRealWCALive` (the picker: the
`competitions(filter:, from:)` query, the country filter and the wcaId → live id
resolution) and `TestSmokeDatabase_EndToEnd` (real Postgres + real WCA Live, end
to end):

```bash
WCA_LIVE_SMOKE=HungarianAllRounder62026 go test ./internal/special -run Smoke -v

SMOKE_DATABASE_URL="postgres://…" SMOKE_USER_ID=<users.id> \
  WCA_LIVE_SMOKE=HungarianAllRounder62026 go test ./internal/special -run Smoke -v
```

---

## 19. Hungarian All-Rounder (`internal/allrounder` + `/tools/allrounder`)

The season-long point standings of the Hungarian All-Rounder series. Each leg is
its own WCA competition; a competitor's season score is the sum of the points
earned for their placements. Ported from the standalone `allrounderpoints`
Next.js app (previously hosted on Vercel) — the frontend `ToolCard` now links to
the internal route, not the external site.

### What moved where

The original fetched **every leg's full results from the WCA API in the browser
on every page load** (~100 KB × 10–12 legs) and scored them client-side. Now:

- The WCA results are **cached in Postgres** (`allrounder_results`, §7) and the
  scoring runs in Go. The page is a **Server Component** doing one fetch.
- `next-intl` (hu/en), `heap-js`, `next-themes` and `react-loading-skeleton` were
  **dropped**: the site is Hungarian-only, the scoring left the browser, and dark
  mode comes from the design tokens. Only `@cubing/icons` came along.

### ⚠️ There are TWO scoring implementations, and that is deliberate

**`ComputeStandingsByRound` (points.go) is the current one. Use it for anything
new.** `ComputeStandingsLegacy` exists only to keep already-published seasons
frozen. `Series.LegacyScoring` selects between them (true for 2024/2025).

**The modern path is round-structured and source-agnostic.** Its input is
`[]CompetitionRounds` (rounds.go): event → round → placements sorted by
position. Both sources normalize into it — `RoundsFromResults` for the cached
WCA API rows, `roundResultsFromLive` for WCA Live. Two rules, stated explicitly
rather than emerging from array order:

1. Per (competitor, event) the **latest round in which they scored** counts;
   whoever DNF'd the final (or fell outside the scoring cut there) drops back to
   their earlier round's placement.
2. An invalid result (`best <= 0`) earns nothing and **shifts nothing**.

*Verified:* the same leg scored from WCA Live and from the official WCA API
gives **identical standings** (AllRounder6 2026: 36 competitors, zero
difference). That is what makes the live merge safe — the numbers do not jump
when the delegate finally uploads.

### ⚠️ What the WCA API result array actually looks like

Measured across all seven uploaded 2026 legs. Two of these contradict what this
document used to claim:

- (event, round-type) blocks **are** contiguous ✓
- round types within an event **are** ascending ✓
- ⚠️ the array is grouped **round-type-major, not event-major** — every first
  round, *then* every final (`333/1, 333oh/c, …, 333/f, pyram/f`)
- ⚠️ **within a round the rows are NOT sorted by position** (AllRounder2 `555/c`
  arrives 8, 25, 15, 13, 23, …). AllRounder6 happened to be sorted, which is
  what made this easy to miss.

Because of the last point, **any order-dependent rule rests on WCA
serialization accident**. That is the whole reason the scoring was moved onto
the round structure.

**The legacy path's bug, for the record.** `ComputeStandingsLegacy` reads the
flat array backwards and counts invalid results into `invalidResultCount`,
which past position 20 rewrites the points index. Since rows inside a round are
unordered, that count is essentially arbitrary, and it awards **inverted**
points at the tail — measured on AllRounder6 2026 `sq1`: 21st place 3 points,
20th place 1 point, tying the 21st with the 18th. The original
`utils/pointCalculator.ts` had the same bug, which is why the two agreed when
the port was validated. Fixing it changes the 2026 season by 4 competitors and
1–3 points each (one rank swap, 41↔42).

**Do not "fix" `ComputeStandingsLegacy`.** 2024 and 2025 are closed and
published; reproducing the announced result is the job. The cached
`allrounder_results.ordinal` is what freezes them — that is the only remaining
reason the column is load-bearing. New seasons must leave `LegacyScoring` false.

### The series config lives in code, not the DB

`series.go` holds the seasons (competition ids + labels, points table, counting
results, podium count) — the direct port of `utils/competitionIds.ts`. It changes
about once a year, and keeping it next to the scoring rules is what makes both
readable. **The DB holds only what was downloaded from WCA.** Adding a leg =
one entry in `series.go`; the next request downloads and caches it.

### Cache freshness

`allrounder_syncs` records the last attempt per competition; `needsSync`
(service.go) decides staleness:

| state | TTL | why |
| --- | --- | --- |
| never synced | — | always fetch |
| has results | 7 days | WCA results are final once uploaded |
| no results yet | 6 hours | the leg is upcoming, or not uploaded yet |
| last attempt failed | 15 min | retry soon, but don't hammer |

Rules that matter:

- **A failed fetch never empties the table.** `MarkSyncFailed` only touches the
  sync row; the previously cached results keep serving. A whole-season wipe from
  one bad upstream response would be the worst possible failure mode.
- **A 404 means "no results yet" only if there were none before.** The WCA API
  404s for a competition whose results are not uploaded, so that is the normal
  state of a future leg — but a 404 for a leg that *already had* results is far
  more likely an upstream glitch, and is treated as a failure instead.
- Sync errors are **not fatal to the request**: whatever is cached is still
  scored and returned. Ten legs cached and one unreachable still renders.
- The sync runs on a **detached context** (`context.WithoutCancel` + a 2 min
  timeout). Otherwise a viewer navigating away mid-fetch would abort it and
  write a failure that suppresses retries for 15 minutes.
- A per-season mutex serialises refreshes, and the computed standings are held
  in memory for 5 min. Cold start (nothing cached) measured at ~2.5 s for a
  12-leg season; warm ~50 ms. ⚠️ **While a leg is live the standings cache is
  bypassed**, otherwise the page would show a 5-minute-old live score.
- Syncing a leg also fetches `/competitions/{id}` for its dates. Its failure is
  deliberately non-fatal: results are what matter, dates only gate live
  tracking. A 404 there is normal — `series.go` carries planned ids for legs the
  WCA does not know yet.

### The live leg

The running leg has no official results yet (they appear when the delegate
uploads), so it comes from **WCA Live** and is merged into the same scoring as
an ordinary `CompetitionRounds` with `Live: true`.

- **Which leg is live**: no official results (`result_count == 0`), dates known,
  and today is inside the competition day ±24 h (`liveGrace`). Nothing switches
  it off explicitly — the official results arriving is what makes it stop
  qualifying.
- **The dates are what bound WCA Live traffic.** Without them we would probe
  every unpublished leg (including months-away ones, and ids the WCA does not
  know) on every page load.
- `LiveTracker` (tracker.go) is a **single-competition** version of the special
  tracker — the series decides which leg runs, so the viewer-chosen multi-
  competition machinery is not needed. The lifecycle is literally the same code:
  both share `wcalive.Session` (see §18), so the WCA Live socket exists **only
  while an SSE stream is open**. No background watcher.
- ⚠️ **A failed round query keeps its previous snapshot** instead of dropping the
  round. Rebuilding the map from scratch made the live leg's points visibly fall
  and come back on the next refresh, which reads as lost points; a load where
  *every* round failed returns an error rather than committing an empty snapshot.
  Same rule, same reason, in `special` (there it corrupts the A1a3 budget).
- `GET /allrounder/stream?year=` opens even when no leg is running (it just
  stays quiet), so a competition starting mid-session lights up on its own.
- ⚠️ **WCA Live gives `best: 0` / `ranking: null` for a competitor who has not
  started.** `0` is not negative, so a bare `best < 0` filter would award them
  points for their position. `roundResultsFromLive` drops those, plus rows
  without a `wcaId`.
- Live points are provisional by nature (rankings churn mid-round, a delegate
  can edit, the official upload can differ) — `Standings.Live` exists so the UI
  says so. Don't remove that notice.

### Frontend notes

`/tools/allrounder` (colocated `components/lib/types`), Server Component:

- The season is in the URL (`?year=`), so the page is shareable. `YearPicker` is
  keyed on the season in `page.tsx` so a Back navigation resets its local state.
- The table's **identity column is one frozen cell** holding rank + name +
  total. One cell, not three, because several sticky columns would need their
  `left` offsets to match the rendered widths pixel for pixel. The table is
  `border-separate` on purpose: with `border-collapse` the browsers do not carry
  borders along with sticky cells.
- The admin-only `RefreshButton` POSTs `/allrounder/refresh`; the server checks
  the role independently, the client check is UX only.
- `LiveBanner` is the only client component that matters for live updates: it
  holds the SSE stream open and calls **`router.refresh()`** on `update`, so the
  *same* Server Component recomputes the table. The scoring never reaches the
  browser, and there is no second fetch path. It mounts even without a running
  leg (that's what keeps the stream open); it just renders nothing.

⚠️ **Dev container note:** `@cubing/icons` was added to `apps/web/package.json`,
but the dev override mounts `web_node_modules` as a named volume — a new
dependency needs `docker compose … exec web npm install` **and a `web` container
restart** (the Next dev server caches the failed resolution). Test/prod rebuild
the image, so `npm ci` handles it there.

## 20. Post képgaléria / carousel (`posts.image_paths` + `PostGallery`)

Egy poszthoz **legfeljebb 10 kép** tartozhat (`post.MaxImages`), és a poszt
oldalán carousellel lehet lapozni köztük. Korábban egyetlen borítókép volt
(`posts.image_path`), ezt váltotta a `000012_add_post_images`.

**A sorrend maga az adat.** Az `image_paths` tömb sorrendje egyszerre három
dolgot jelent: az **első elem a borítókép** (ez látszik a listákban és a
`NewsCard`-on), a teljes tömb pedig a carousel lapozási sorrendje. Ezért tömb
és nem külön `post_images` tábla: nem kell `ordinal` oszlop és join ahhoz, hogy
a sorrend megmaradjon. Ne normalizáld külön táblába, hacsak nem lesz képenkénti
metaadat (alt szöveg, szerző).

### A multipart szerződés

- **Létrehozás**: ismételt `images` fájlmezők, a beérkezés sorrendje = a galéria
  sorrendje.
- **Szerkesztés**: az `image_order` mező egy JSON tömb, ami a **teljes végső
  galériát** írja le. Minden eleme vagy a poszt egy MEGLÉVŐ képútvonala, vagy a
  `"new"` jelölő, amit sorban a most feltöltött `images` fájlok töltenek ki.
  Ez egyetlen mezőben fejezi ki a törlést, az átrendezést és azt is, hogy egy új
  kép a meglévők **közé** kerüljön — ez utóbbi miatt nem elég a „megtartandók
  listája + a végére fűzött újak" megoldás.
- ⚠️ `resolveImageOrder` (handler.go) **elutasít minden olyan útvonalat, ami nem
  ehhez a poszthoz tartozik** — enélkül a kliens más poszt (vagy tetszőleges)
  képére hivatkozhatna, és a takarítás azt törölné. Ugyanígy hibás, ha marad
  hivatkozás nélküli feltöltött fájl (árva fájl lenne a volume-on), vagy ha egy
  útvonal kétszer szerepel.
- Szerkesztéskor az `image_order` **kötelező** (hiányában 400). A régi
  kompatibilitási ágak — JSON törzs, egyes számú `image` mező,
  `remove_image=true` — megszűntek: az egyetlen kliens, a `CreatePostModal`,
  mindig multipartot és `image_order`-t küld. A létrehozás és a szerkesztés
  közös `parsePostForm`-on megy át, a hibák státuszát egyetlen `errorStatus` adja.
- A legfeljebb 10 kép × 5 MiB miatt a kérés törzsének korlátja `maxUploadBytes`;
  a **fájlonkénti** méretet a `FileHeader.Size` alapján utasítjuk el, mert a
  `SaveImage`-beli `LimitReader` csak csonkítana (néma félkép helyett 400).
- Minden hibaágon takarítunk: `SaveImages` a saját félig kész feltöltéseit, a
  handler pedig a sikertelen mentés utáni fájlokat törli. Sikeres mentés után a
  galériából **kiesett** képek fájljai törlődnek.

### Frontend

- `app/posts/lib/postImage.ts`: `postCoverUrl()` a listákhoz (első kép),
  `postImageUrls()` a carouselhez, `MAX_POST_IMAGES` a backend korlát tükre.
- `app/posts/components/PostGallery.tsx` (client): **egyetlen képnél
  szándékosan nincs vezérlő** — pontosan az a kép marad, ami korábban is látszott.
  Több képnél nyilak, pöttyök, `ArrowLeft`/`ArrowRight` és mobilon húzás lapoz,
  körbeérve. A képek egy eltolt sávban vannak (nem feltételes renderelés), így a
  böngésző előre betölti a következőt.
- A `CreatePostModal` galériaszerkesztője egyetlen listában kezeli a meglévő
  (útvonal) és az új (File) képeket, hogy a sorrendjük együtt legyen
  szerkeszthető; submitkor ebből áll elő az `image_order` és ebben a sorrendben
  fűzi hozzá az `images` fájlokat. A kettőnek **egyeznie kell** — a backend
  sorban tölti ki a `"new"` helyeket.
- A poszt TÖRLÉSE a képfájlokat is elviszi. A `Repository.Delete` a
  `DELETE … RETURNING image_paths`-szal EGY lekérdezésben adja vissza a
  felszabadult útvonalakat (külön SELECT + DELETE közben változhatna a
  galéria), a handler pedig ezekre hívja a `RemoveImages`-t — hibaágon is,
  mert a visszakapott sorok addigra már tényleg törlődtek.

## 21. Tippverseny (`internal/prediction` + `/tools/prediction`)

Admin által versenyenként indítható tippverseny: dobogó-tippek (1-2-3. hely)
a verseny kiválasztott versenyszámaira, plusz szabad extra kérdések
(versenyző/idő/szám/szabad szöveg típussal, pl. "Hány rekord születik?").
**Csak a
LÉTREHOZÁS, a LISTÁZÁS és a TÖRLÉS van megépítve** — a tippek leadása
(userenkénti válaszok tárolása) és a kiértékelés (pontozás, végeredmény) még
nincs, ez egy külön, nagyobb következő lépés lesz. A frontend típusa
(`PredictionStatus`) emiatt tartalmaz egy `"evaluated"` állapotot is, amit a
backend ma sosem ad vissza.

**A verseny metaadatai (név, város, dátum) NINCSENEK nálunk tárolva.** A
`prediction_contests.competition_wca_id` csak egy szöveges azonosító, FK
nélkül: a verseny a WCA-tól jön, mi csak az *ID*-t tartjuk meg. Listázáskor a
`prediction.Service.GetContests` minden versenyt **ID szerint** kér le
(`competition.Service.GetCompetition` → WCA `/competitions/{id}`, 15 percig
gyorsítótárazva) — ha a WCA API hibázik, a tippverseny akkor is megjelenik, csak
a verseny neve a nyers `competition_wca_id` lesz (lásd `buildContest`), és a
hiba a naplóba kerül. ⚠️ **Ne a `GetUpcomingHungarian`-ből töltsd:** az a
verseny végével kiejti a versenyt, és a tippverseny pont az
eredményhirdetéskor veszítette el a nevét, dátumát és helyszínét. Emiatt `prediction.Service`
egy szűk `CompetitionLookup` interfészen keresztül függ a competition csomagtól
(nem a konkrét `*competition.Service`-en), hogy a tesztelhetőség és a réteghatár
tiszta maradjon.

**A `status` (nyitott/lezárt) számított érték, nem oszlop.** A
`prediction_deadline`-ból számolódik minden listázáskor (`statusFor` a
`model.go`-ban) — így a leadási határidő és a megjelenített állapot soha nem
csúszhat szét.

Adatmodell (lásd §7 a teljes séma-blokkhoz):
- `prediction_contests` — egy sor egy tippversenyt jelent.
- `prediction_podium_events` — melyik versenyszámokhoz kell dobogót tippelni
  (gyerek tábla, `ON DELETE CASCADE`).
- `prediction_extra_questions` — az admin szabad kérdései, `ordinal`-lal
  megjelenítési sorrendben (mint az `allrounder_results.ordinal`, de itt csak
  sorrendezéshez kell, nem számítási bemenet).

A `POST /predictions` (`prediction.Handler.CreateContest`) mindhárom táblát
**egy tranzakcióban** írja (`Repository.Create`) — vagy mindhárom sikerül,
vagy egyik sem, hogy ne maradjon félig beírt tippverseny.

A `DELETE /predictions/{id}` (`prediction.Handler.DeleteContest`) NEM
tranzakció — elég egyetlen `DELETE FROM prediction_contests WHERE id = $1`,
mert a gyerek táblák FK-ja `ON DELETE CASCADE`. 0 törölt sor `ErrContestNotFound`-ot
ad vissza, amit a handler 404-re fordít (ugyanaz a minta, mint a `post.Update`
a `posts` táblán, lásd §6).

### Frontend

`/tools/prediction` (`app/tools/prediction/`), Server Component: a lista
oldal (`page.tsx`) szerveroldalon hívja a `GET /predictions`-t (mint a
`/posts`, `serverApiUrl()`-lel), és publikus — bárki látja a listát, csak a
létrehozás és a törlés admin (`useUser()` role-check a `PredictionList`-ben,
mint a posztoknál — `isAdmin` lejut a `PredictionListItem`-be, ami a törlés
gombot rendereli). `DeletePredictionButton.tsx` egyetlen tippversenyt töröl,
megerősítésre a közös `app/components/ui/ConfirmModal.tsx`-t használva (lásd
§6 — sem itt, sem a posztoknál nincs többes kijelölés/bulk törlés, mindkettő
elemenkénti kuka-ikon). A gomb a kártya `Link`-jének `after:absolute
after:inset-0` overlay-je fölött `relative z-10`-zel marad kattintható,
ugyanúgy, mint a `PostsListItem` szerkesztő/törlő gombjai.

Az admin `CreatePredictionModal` (`components/CreatePredictionModal.tsx`):
- Verseny-választó a `GET /competitions/upcoming`-ból (böngészőből,
  `clientApiUrl()`), amely most már `eventIds`-t is ad versenyenként (lásd §8)
  — ebből épül fel a dobogó-jelölők listája versenyenként, `@cubing/icons`
  ikonokkal, a KÖZÖS `app/lib/events.ts` névtérképpel (ugyanazt használja az
  allrounder tool is, §19 — a WCA versenyszám-nevei nem toolonkénti
  terméklogika, és két példánynál egy új versenyszám az egyik felületen némán
  nyers azonosítóként jelenne meg).
- Extra kérdés-szerkesztő (`ExtraQuestionsEditor.tsx`): tetszőleges számú sor,
  típusonként (Versenyző/Idő/Szám/Szabad szöveg) választható select-tel.
- A POST body RFC3339 `predictionDeadline`-t küld
  (`new Date(...).toISOString()`); a backend `time.Parse(time.RFC3339, ...)`-vel
  olvassa.

**A részletező oldal (`[id]/page.tsx`) kitölthető, de a submit még nem hív
végpontot.** `GET /predictions/{id}` nem létezik, ezért a szerveroldali
`getPrediction` a teljes `GET /predictions` listát kéri le és szűr `id`
szerint; ha a backend kap egyedi lekérő végpontot, ezt cseréld le. A form
állapota a `PredictionForm` client komponensben él (a `page.tsx` Server
Component marad). A "Tipp leadása" gomb **mindig `disabled`**, amíg nincs
leadási végpont, és a felület ki is mondja, hogy a mezők nem mentődnek.
⚠️ A `<form onSubmit={e => e.preventDefault()}>` nem hagyható el: nélküle a
gomb vagy egy Enter natív GET submitot indít, ami újratölti az oldalt, és a
kitöltött tipp elveszik. Ha a leadás megépül, ide kerül a valódi handler. A mezők a kérdés típusához igazodnak:

- **Versenyző** (`CompetitorInput`): typeahead, a választás a NEVET adja
  vissza. A dobogó-tippeknél a versenyszám **psych sheet-je** a lista
  (`/competitions/{id}/psych-sheet/{eventId}`, a `#N` pozíció csak
  tájékoztat), az extra "versenyző" kérdéseknél a **teljes nevezői lista**
  (`/competitions/{id}/competitors`). A listát nem az input tölti be, hanem
  a szülő a `useCompetitorOptions` hookkal: **egy versenyszám három mezője
  EGY kérés**, a nevezői lista pedig egyszer, csak ha van person kérdés.
  Hiba esetén a mező sima szöveges input marad (WCA-kiesés ne blokkoljon).
- **Idő** (`TimeInput` + a KÖZÖS `app/lib/time.ts`): csak számjegy, gépelés
  közben jobbról tagolva (`14289` → `1:42.89`), plafon 60:00.00, 6 számjegy.
  ⚠️ A `formatTimeInput` levágja a vezető nullákat — a megjelenített `0.01`
  nyers számjegyei `001`, és enélkül `001:23.45`-té halmozódna.
  Elutasított bevitelnél (`null`) a mező az előző értéken marad. Ugyanezt
  használja a Budapest Special DNF-cellája is (§18) — **ne másold vissza egy
  toolba**, korábban két, viselkedésében eltérő példány volt.
- **Szám**: `type="number"`.
- **Szabad szöveg** (`types.QUESTION_TEXT` / `"text"`, migráció
  `000014_add_prediction_text_question_type`): sima `type="text"` input,
  olyan kérdésekhez, amik nem versenyzőt, időt vagy számot várnak (pl.
  "Milyen új trükköt fogunk látni?"). A Postgres enumhoz hozzáadott érték —
  a `down` migráció szándékosan no-op, ugyanaz a minta, mint a `post_tag`
  `'report'` értékénél (000003): a Postgres nem tudja eltávolítani egy enum
  értékét, az árva érték ártalmatlan.

⚠️ **A WCA API kb. 15 párhuzamos kérés után 429-et ad.** A psych sheet
mezőnkénti lekérése (11 versenyszám × 3 hely = 33) ezt minden oldalbetöltésnél
túllépte, és 502-ként jött vissza. Ezért (1) a frontend versenyszámonként egy
kérést küld, (2) a `competition.Service` 15 percig memóriában cache-eli a
psych sheeteket és a nevezői listát, és `singleflight`-tal összevonja az
egyidejű azonos kéréseket (a React StrictMode dev-ben minden fetch-et
duplán futtat). Hibás válasz nem kerül a cache-be, a lejárt bejegyzéseket
pedig TTL-enként egyszer kisöpri (`sweepLocked`) — a kulcs a kliens által
adott azonosító, takarítás nélkül korlátlanul nőne. (3) A
`competition.WCARepository` egyszerre legfeljebb 6 kérést enged a WCA felé
(§14) — a singleflight csak AZONOS kulcsokat von össze, egy tucat különböző
versenyszám hideg cache mellett egyszerre ment volna ki. Ne tedd vissza a
lekérést az inputba, és ne vedd ki a cache-t.

⚠️ Amikor a tippleadás/kiértékelés megépül, ne bővítsd tovább ezt a modult
gondolkodás nélkül — a userenkénti válaszok tárolása és a pontozás logika
valószínűleg saját táblát/service-réteget igényel, hasonlóan ahhoz, ahogy az
allrounder is külön választja a "series config kódban" (§19) és a "letöltött
eredmény DB-ben" felelősséget.

## Instructions for Claude

When working in this repo:

1. **Follow existing patterns instead of introducing new ones.** New Go features
   use the `model/repository/service/handler` shape, wired in `cmd/server/main.go`
   and passed into `router.New` (which only registers routes). New frontend
   routes colocate `components/lib/types` under `app/<route>/`.
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
