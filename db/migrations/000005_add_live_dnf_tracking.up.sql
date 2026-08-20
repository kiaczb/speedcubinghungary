-- WCA Live alapú DNF-megfigyelés.
--
-- Miért külön tábla és nem a hivatalos eredmény kiegészítése:
-- egy DNF tényleges ideje NEM része a hivatalos WCA eredménynek. A WCIF
-- Attempt-nek csak `result` és `reconstruction` mezője van, a WCA Live pedig
-- a Result.changeset-ben felülírja az időt DNF-re, ha átlépi a limitet — az
-- eredeti idő tehát sehol nem marad meg. Az itt tárolt idő a MI megfigyelésünk
-- (a bíró/adatrögzítő jegyzi fel), ezért saját, alkalmazás-specifikus adat.
--
-- Az időbélyegek TIMESTAMPTZ-ek (a posts/users TIMESTAMP-jaival ellentétben):
-- a WCA Live UTC pillanatokat ad vissza, és élő verseny közben a zónás
-- értelmezés nem lehet kérdéses.

-- Melyik versenyeket figyeljük. Újraindítás után innen töltjük vissza a
-- figyelt versenyek listáját (lásd internal/live/watcher.go).
CREATE TABLE live_tracked_competitions (
    competition_wca_id  TEXT PRIMARY KEY,          -- pl. 'HungarianAllRounder62026'
    live_competition_id INTEGER NOT NULL,          -- WCA Live numerikus id-ja
    name                TEXT NOT NULL,
    start_date          DATE,
    end_date            DATE,
    active              BOOLEAN NOT NULL DEFAULT TRUE,
    added_by            INTEGER REFERENCES users(id) ON DELETE SET NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Egy megfigyelt DNF kísérlet.
CREATE TABLE live_dnf_attempts (
    id UUID PRIMARY KEY,

    -- WCA Live azonosítók (a megfigyelés forrása)
    competition_wca_id  TEXT NOT NULL REFERENCES live_tracked_competitions(competition_wca_id) ON DELETE CASCADE,
    live_round_id       INTEGER NOT NULL,
    live_result_id      INTEGER NOT NULL,
    -- WCIF-kompatibilis leírók, hogy az adat WCA Live nélkül is értelmezhető legyen
    event_id            TEXT NOT NULL,             -- '333'
    round_number        INTEGER NOT NULL,          -- 1
    round_wcif_id       TEXT NOT NULL,             -- '333-r1' (WCIF Round.id / ActivityCode)
    attempt_number      INTEGER NOT NULL,          -- 1-alapú, WCIF szerinti

    -- versenyző
    registrant_id       INTEGER NOT NULL,          -- WCIF Person.registrantId
    competitor_wca_id   TEXT,                      -- újoncnál NULL
    competitor_name     TEXT NOT NULL,

    -- a kör paraméterei a megfigyelés pillanatában (a felső korlát számításához)
    time_limit_cs        INTEGER,                  -- NULL: 333fm / 333mbf
    cumulative_round_ids TEXT[] NOT NULL DEFAULT '{}',
    -- Felső KORLÁT, nem mérés: per-attempt limitnél maga a limit, kumulatívnál
    -- a limit mínusz a korábbi sikeres kísérletek összege. Soha nem tekintendő
    -- a solve tényleges idejének.
    max_elapsed_cs      INTEGER,
    -- A versenyző körön belüli sikeres (pozitív) kísérleteinek összege a
    -- megfigyelés pillanatában. Azért tároljuk, hogy a kumulatív limit
    -- elhasználása a WCA Live újbóli lekérdezése nélkül számolható legyen.
    completed_cs        INTEGER NOT NULL DEFAULT 0,

    -- időbélyegek
    detected_at         TIMESTAMPTZ NOT NULL DEFAULT now(),  -- mi mikor láttuk meg
    -- A WCA Live Result.enteredAt-je: az ADATRÖGZÍTÉS ideje (másodperc pontos,
    -- egész eredményre közös), NEM a solve ideje. Csak nyomkövetéshez tároljuk.
    entered_at          TIMESTAMPTZ,
    -- Ha az upstream kísérlet már nem DNF (elgépelés javítása), ezt állítjuk be
    -- törlés helyett: a felvett idő és az audit nyom megmarad.
    resolved_at         TIMESTAMPTZ,

    -- az emberi megfigyelés: a solve TÉNYLEGES ideje századmásodpercben
    elapsed_cs          INTEGER,
    elapsed_note        TEXT,
    recorded_by         INTEGER REFERENCES users(id) ON DELETE SET NULL,
    recorded_at         TIMESTAMPTZ,

    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- Idempotencia: a WCA Live result id globálisan egyedi (verseny+kör+versenyző),
    -- a kísérlet sorszámával együtt pontosan egy kísérletet azonosít. Ugyanannak
    -- az eseménynek a kétszeri feldolgozása így nem hoz létre két sort.
    UNIQUE (live_result_id, attempt_number)
);

-- A leggyakoribb lekérdezés: egy verseny nyitott megfigyelései, legfrissebb elöl.
CREATE INDEX live_dnf_attempts_competition_idx
    ON live_dnf_attempts (competition_wca_id, detected_at DESC);

-- A kumulatív limit számításához körönként/versenyzőnként kérdezünk.
CREATE INDEX live_dnf_attempts_round_registrant_idx
    ON live_dnf_attempts (live_round_id, registrant_id);
