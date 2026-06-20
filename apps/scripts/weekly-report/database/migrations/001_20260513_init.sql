-- Migrációs napló tábla (ha még nem lenne)
CREATE TABLE IF NOT EXISTS schema_migrations (
    version_name TEXT PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Alaptáblák létrehozása
CREATE TABLE IF NOT EXISTS persons (
    wca_id VARCHAR(15) PRIMARY KEY,
    name TEXT NOT NULL,
    competition_count INTEGER DEFAULT 0,
    delegated_count INTEGER DEFAULT 0
);

CREATE TABLE IF NOT EXISTS records (
    region VARCHAR(50),
    event_id VARCHAR(20),
    single INTEGER,
    average INTEGER,
    PRIMARY KEY (region, event_id)
);