-- A WCA Live-ról érkező TELJES eredmények (nem csak a DNF-ek).
--
-- Miért kell: a DNF-ek önmagukban nem elegendők a versenyzői nézethez. Ahhoz,
-- hogy a BpSpecial-höz hasonlóan minden időt, az összesített felhasznált és a
-- maradék időt is meg tudjuk mutatni, ismernünk kell a sikeres kísérleteket is.
-- A watcher úgyis teljes kör-pillanatképet kap, tehát ez nem jelent plusz
-- lekérdezést a WCA Live felé.
--
-- Ez a tábla TÜKÖR: a WCA Live hivatalos eredményének helyi másolata, amit
-- minden pillanatképnél felülírunk. Emberi adat NEM kerül bele — az továbbra is
-- kizárólag a live_dnf_attempts.elapsed_cs oszlopban él.
CREATE TABLE live_results (
    -- A WCA Live result id globálisan egyedi (verseny + kör + versenyző),
    -- ezért önmagában alkalmas kulcsnak: ugyanannak a pillanatképnek a
    -- többszöri feldolgozása egyetlen sort érint.
    live_result_id       INTEGER PRIMARY KEY,
    competition_wca_id   TEXT NOT NULL REFERENCES live_tracked_competitions(competition_wca_id) ON DELETE CASCADE,
    live_round_id        INTEGER NOT NULL,

    event_id             TEXT NOT NULL,
    round_number         INTEGER NOT NULL,
    round_wcif_id        TEXT NOT NULL,

    registrant_id        INTEGER NOT NULL,
    competitor_wca_id    TEXT,
    competitor_name      TEXT NOT NULL,

    -- WCIF AttemptResult értékek sorrendben: >0 = századmásodperc,
    -- -1 = DNF, -2 = DNS, 0 = kihagyott.
    attempts             INTEGER[] NOT NULL DEFAULT '{}',
    best                 INTEGER,
    average              INTEGER,

    time_limit_cs        INTEGER,
    cumulative_round_ids TEXT[] NOT NULL DEFAULT '{}',

    entered_at           TIMESTAMPTZ,
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- A versenyzői nézet versenyenként, versenyzőnként csoportosít.
CREATE INDEX live_results_competition_registrant_idx
    ON live_results (competition_wca_id, registrant_id);
