-- Algoritmus-tanulás: Versenyszám → Eset → Algoritmus (+ opcionális videó).
--
-- A tartalmat KIZÁRÓLAG admin viszi fel (a meglévő RBAC-cal, lásd
-- internal/middleware), az olvasás publikus. Nincs felhasználói tartalom, ezért
-- nincs szerző/moderáció/verziózás sem — a három tábla csak a hierarchiát írja le.
--
-- Miért ez a három szint: a kockás anyagok mindig így épülnek fel (3x3 → OLL 21
-- → R U R' U' R U2 R'). Az "esemény" itt szándékosan SZABAD SZÖVEG, nem a WCA
-- event-id enum: tanulni nem csak hivatalos versenyszámokat lehet (pl. "3x3
-- CFOP", "Roux", "Egykezes"), és a jövőbeli bővítést sem akarjuk enumhoz kötni.

CREATE TABLE learn_events (
    id          UUID PRIMARY KEY,
    name        TEXT NOT NULL,
    description TEXT,
    -- sort_order: az admin által megadott megjelenítési sorrend. NEM "order",
    -- mert az SQL kulcsszó, és minden lekérdezésben idézőjelezni kellene.
    sort_order  INTEGER NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE learn_cases (
    id          UUID PRIMARY KEY,
    -- ON DELETE CASCADE: a hierarchia felülről lefelé tulajdonol. Egy versenyszám
    -- törlése az alatta lévő eseteket (és azokon keresztül az algoritmusokat) is
    -- viszi — árva eset önmagában értelmezhetetlen lenne. A felület ezért
    -- megerősítést kér, és kiírja, hány elem tűnik el vele.
    event_id    UUID NOT NULL REFERENCES learn_events(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    description TEXT,
    sort_order  INTEGER NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE learn_algorithms (
    id          UUID PRIMARY KEY,
    case_id     UUID NOT NULL REFERENCES learn_cases(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    notation    TEXT NOT NULL,              -- pl. "R U R' U'"
    description TEXT,
    -- video_url: EGY opcionális, külső videó (ma YouTube-link). Szándékosan
    -- oszlop és nem külön tábla: most egy videó kell, feltöltés nincs. Ha később
    -- feltöltött/több videó kell, az egy új learn_algorithm_videos tábla lesz,
    -- ez az oszlop pedig migrálható oda — a frontend már ma is egy külön
    -- felismerő rétegen (app/learn/lib/video.ts) keresztül nyúl hozzá.
    video_url   TEXT,
    sort_order  INTEGER NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- A tipikus lekérdezés egy szülő gyerekeit kéri, megjelenítési sorrendben.
CREATE INDEX learn_cases_event_idx      ON learn_cases (event_id, sort_order, name);
CREATE INDEX learn_algorithms_case_idx  ON learn_algorithms (case_id, sort_order, name);
