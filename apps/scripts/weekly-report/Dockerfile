# syntax=docker/dockerfile:1
# Weekly-report job image. Standalone Python cron that reads the shared Postgres
# and emails a weekly report over SMTP. It is NOT part of the always-on stack —
# it runs on demand via the compose `jobs` profile:
#
#   docker compose -f infra/docker-compose.yml -f infra/docker-compose.test.yml \
#     --env-file .env.test --profile jobs run --rm weekly-report
#
# Build context is this directory (apps/scripts/weekly-report), unlike the api
# image whose context is the repo root — this job needs no db/migrations.

FROM python:3.11-slim

WORKDIR /app

# Install deps first so the layer is cached across source changes.
# psycopg-binary ships wheels, so no build toolchain is needed.
COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt

# Copy the job source. See .dockerignore: the git-ignored .env, virtualenv and
# __pycache__ are excluded so no secrets or stale caches are baked into the image.
COPY . .

# Run as a non-root user, matching the api/web images.
RUN useradd -m -u 1001 appuser
USER appuser

# All configuration comes from the environment (compose --env-file). config.py's
# load_dotenv() silently no-ops when no .env file is present in the image, so the
# container reads EMAIL_* / DB_* straight from the process environment.
CMD ["python", "main.py"]
