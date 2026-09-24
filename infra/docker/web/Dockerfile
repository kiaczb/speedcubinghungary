# syntax=docker/dockerfile:1
# Multi-stage build for the Next.js web app.
# Stage selection is driven by the compose `target:`:
#   - dev  -> `next dev` hot reload, source bind-mounted at runtime (docker-compose.dev.yml)
#   - prod -> minimal standalone runtime (default; used by test/prod overrides)

# ---------- dev: hot reload via `next dev` (source bind-mounted at runtime) ----------
FROM node:22-alpine AS dev
WORKDIR /app
COPY apps/web/package*.json ./
RUN npm install
EXPOSE 3000
# --webpack: Turbopack has no polling fallback; webpack + WATCHPACK_POLLING
# is what makes hot reload work over Docker bind mounts on Windows/WSL.
CMD ["npm", "run", "dev", "--", "--webpack"]

# ---------- deps: clean, cached dependency install for the build ----------
FROM node:22-alpine AS deps
WORKDIR /app
COPY apps/web/package.json apps/web/package-lock.json ./
RUN npm ci

# ---------- builder: produce the standalone Next.js output ----------
FROM node:22-alpine AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY apps/web/ .
# NEXT_PUBLIC_* are inlined into the browser bundle at BUILD time, so they must
# arrive as build args (provided by the test/prod compose overrides).
ARG NEXT_PUBLIC_API_URL
ARG NEXT_PUBLIC_BACKEND_LOGIN
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL \
    NEXT_PUBLIC_BACKEND_LOGIN=$NEXT_PUBLIC_BACKEND_LOGIN \
    NEXT_TELEMETRY_DISABLED=1 \
    NODE_ENV=production
RUN npm run build

# ---------- prod: minimal standalone runtime ----------
FROM node:22-alpine AS prod
WORKDIR /app
ENV NODE_ENV=production \
    NEXT_TELEMETRY_DISABLED=1 \
    PORT=3000 \
    HOSTNAME=0.0.0.0
RUN addgroup -g 1001 -S nodejs && adduser -S nextjs -u 1001
COPY --from=builder /app/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static
USER nextjs
EXPOSE 3000
CMD ["node", "server.js"]
