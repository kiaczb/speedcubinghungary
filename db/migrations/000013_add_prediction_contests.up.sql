-- Tippverseny: adminok versenyenként tippversenyt hoznak létre — dobogó-tippek
-- a kiválasztott versenyszámokra, plusz szabad extra kérdések (versenyző/idő/
-- szám típussal). A tippek LEADÁSA és a KIÉRTÉKELÉS egyelőre nincs itt —
-- ez a migráció csak magát a tippverseny KONFIGURÁCIÓJÁT tárolja.
--
-- A `status` (nyitott/lezárt) NEM oszlop: a prediction_deadline-ból számolt
-- érték, hogy a leadási határidő ne csúszhasson szét a DB és a kiszámított
-- állapot között.
CREATE TABLE prediction_contests (
    id                   UUID PRIMARY KEY,
    competition_wca_id   TEXT NOT NULL,               -- pl. 'HungarianOpen2026' — nincs FK, a verseny a WCA-tól jön, nem a mi DB-nkből
    prediction_deadline  TIMESTAMPTZ NOT NULL,
    created_by           INTEGER REFERENCES users(id),
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Egy tippversenyhez tartozó versenyszámok, amikhez dobogó (1-2-3. hely) tippet
-- kell leadni. Sorrend nem számít — a UI a saját (WCA esemény-lista) sorrendje
-- szerint jeleníti meg.
CREATE TABLE prediction_podium_events (
    contest_id UUID NOT NULL REFERENCES prediction_contests(id) ON DELETE CASCADE,
    event_id   TEXT NOT NULL,                         -- WCA esemény id, pl. '333'
    PRIMARY KEY (contest_id, event_id)
);

CREATE TYPE prediction_question_type AS ENUM ('person', 'time', 'number');

-- Admin által megadott szabad kérdések (pl. "Hány magyar rekord születik?").
CREATE TABLE prediction_extra_questions (
    id          UUID PRIMARY KEY,
    contest_id  UUID NOT NULL REFERENCES prediction_contests(id) ON DELETE CASCADE,
    prompt      TEXT NOT NULL,
    type        prediction_question_type NOT NULL,
    ordinal     INTEGER NOT NULL                      -- admin által megadott sorrend, megjelenítéshez
);

CREATE INDEX prediction_podium_events_contest_idx ON prediction_podium_events (contest_id);
CREATE INDEX prediction_extra_questions_contest_idx ON prediction_extra_questions (contest_id, ordinal);
