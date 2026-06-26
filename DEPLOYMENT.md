# Deployment & Branching

Single-developer, no feature branches. Three long-lived branches form a
promotion pipeline:

| Branch | Purpose                | Deploys to        | Trigger                |
| ------ | ---------------------- | ----------------- | ---------------------- |
| `dev`  | Local development only | nothing           | —                      |
| `test` | Staging / test server  | staging server    | push to `test`         |
| `main` | Production             | production server | push to `main`         |

`main` is the GitHub default branch. History is linear: every commit flows
`dev → test → main`, so each branch is an ancestor of the next.

## Day-to-day workflow

```bash
# 1. Develop locally on dev
git switch dev
# ...code, commit...
git push origin dev            # does NOT deploy anything

# 2. Promote to staging (fast-forward, then push -> deploys to staging)
git switch test
git merge --ff-only dev
git push origin test

# 3. Promote to production (fast-forward, then push -> deploys to prod)
git switch main
git merge --ff-only test
git push origin main

git switch dev                 # back to developing
```

If `merge --ff-only` ever fails, it means the target branch has commits the
source doesn't — investigate rather than force; with this solo flow it should
always fast-forward.

## CI/CD

Workflows live in `.github/workflows/`:

- `deploy-test.yml` — runs on push to `test`, uses the **`test`** environment.
- `deploy-prod.yml` — runs on push to `main`, uses the **`production`** environment.

Both SSH into the target server, `git reset --hard` the working tree to the
branch, and run `docker compose -f <COMPOSE_FILE> up -d --build`. `up` never
removes named volumes, so the PostgreSQL data is never wiped on deploy.

`dev` has no deploy workflow on purpose.

## Required GitHub setup (one-time)

Create two **Environments** (Settings → Environments): `test` and `production`.
For each, add:

**Secrets**

| Secret     | Description                                  |
| ---------- | -------------------------------------------- |
| `SSH_HOST` | server IP / hostname                         |
| `SSH_USER` | SSH user (must be in the `docker` group)     |
| `SSH_KEY`  | private key; public key in server authorized_keys |
| `SSH_PORT` | SSH port (optional; defaults to 22)          |

**Variables**

| Variable       | Description                                            | Example                   |
| -------------- | ----------------------------------------------------- | ------------------------- |
| `DEPLOY_PATH`  | path to the repo checkout on the server               | `/srv/speedcubinghungary` |
| `COMPOSE_FILE` | compose file to deploy (optional; default `docker-compose.yml`) | `docker-compose.yml` |

The old repo-level `DEV_*` secrets are no longer used and can be deleted. Reuse
the previous DEV server's values for the `test` environment.

Recommended: enable **Required reviewers** protection on the `production`
environment so prod deploys wait for manual approval.

## Server setup (one-time, per server)

```bash
git clone -b test <repo-url> /srv/speedcubinghungary   # or -b main on prod
cd /srv/speedcubinghungary
# create the git-ignored .env files the compose stack needs, then:
docker compose up -d --build
```
