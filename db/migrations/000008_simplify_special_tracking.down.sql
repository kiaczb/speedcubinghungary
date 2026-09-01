-- Visszaállítás. A törölt oszlopok származtatottak voltak, ezért a
-- visszaállításuk üres/alapértelmezett értékekkel történik — az adatot a
-- következő WCA Live lekérdezés tölti fel újra.
ALTER TABLE special_dnf_attempts
    ALTER COLUMN elapsed_cs DROP NOT NULL;

ALTER TABLE special_dnf_attempts
    ADD COLUMN IF NOT EXISTS live_round_id        INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS event_id             TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS round_number         INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS round_wcif_id        TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS registrant_id        INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS competitor_wca_id    TEXT,
    ADD COLUMN IF NOT EXISTS competitor_name      TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS time_limit_cs        INTEGER,
    ADD COLUMN IF NOT EXISTS cumulative_round_ids TEXT[] NOT NULL DEFAULT '{}',
    ADD COLUMN IF NOT EXISTS max_elapsed_cs       INTEGER,
    ADD COLUMN IF NOT EXISTS completed_cs         INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS detected_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    ADD COLUMN IF NOT EXISTS entered_at           TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS resolved_at          TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS special_dnf_attempts_competition_idx
    ON special_dnf_attempts (competition_wca_id, detected_at DESC);
CREATE INDEX IF NOT EXISTS special_dnf_attempts_round_registrant_idx
    ON special_dnf_attempts (live_round_id, registrant_id);
DROP INDEX IF EXISTS special_dnf_attempts_competition_idx2;

CREATE TABLE IF NOT EXISTS special_results (
    live_result_id       INTEGER PRIMARY KEY,
    competition_wca_id   TEXT NOT NULL REFERENCES special_competitions(competition_wca_id) ON DELETE CASCADE,
    live_round_id        INTEGER NOT NULL,
    event_id             TEXT NOT NULL,
    round_number         INTEGER NOT NULL,
    round_wcif_id        TEXT NOT NULL,
    registrant_id        INTEGER NOT NULL,
    competitor_wca_id    TEXT,
    competitor_name      TEXT NOT NULL,
    attempts             INTEGER[] NOT NULL DEFAULT '{}',
    best                 INTEGER,
    average              INTEGER,
    time_limit_cs        INTEGER,
    cumulative_round_ids TEXT[] NOT NULL DEFAULT '{}',
    entered_at           TIMESTAMPTZ,
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);
