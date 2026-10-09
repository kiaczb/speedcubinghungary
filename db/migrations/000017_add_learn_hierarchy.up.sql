-- Tanulás: tetszőleges mélységű navigációs fa, a leveleken algoritmusokkal.
--
-- Szándékosan EGY önhivatkozó tábla, és nincs benne semmilyen szintfogalom
-- (kirakó / metódus / lépés / eset). A "3x3 → CFOP → OLL → OLL 24" és a
-- "Skewb → NS" ugyanaz az adatszerkezet, csak más mélységben — a szintek
-- jelentése az adat, nem a séma.
--
-- Az invariánsok, amiket a séma NEM tud kifejezni, és ezért a Go réteg tartja
-- (internal/learn/repository.go, egyetlen advisory lock alatt):
--   * egy elemnek VAGY gyerekei, VAGY algoritmusai vannak, a kettő egyszerre nem;
--   * áthelyezéskor nem keletkezhet kör (elem a saját leszármazottja alá);
--   * a fa mélysége legfeljebb learn.MaxDepth.
CREATE TABLE learn_nodes (
    id          UUID PRIMARY KEY,
    -- NULL = legfelső szint. ON DELETE CASCADE: egy elem törlése a teljes
    -- részfáját (és azok algoritmusait) viszi, árva elem nem maradhat.
    parent_id   UUID REFERENCES learn_nodes(id) ON DELETE CASCADE,
    -- Az URL-szegmens: /learn/algorithms/3x3/cfop/oll/oll-24. Testvérek között
    -- egyedi; a NULLS NOT DISTINCT kell, különben a legfelső szinten (parent_id
    -- NULL) a Postgres minden NULL-t különbözőnek venne, és két "3x3" is lehetne.
    slug        TEXT NOT NULL,
    title       TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    sort_order  INTEGER NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT learn_nodes_not_own_parent CHECK (parent_id <> id),
    CONSTRAINT learn_nodes_slug_unique UNIQUE NULLS NOT DISTINCT (parent_id, slug)
);

CREATE INDEX learn_nodes_parent_idx ON learn_nodes (parent_id, sort_order);

-- Egy levélhez egy vagy több algoritmus. A videó szándékosan NINCS itt: ha
-- megépül, külön tábla lesz (learn_algorithm_videos), ez a tábla változatlan
-- marad.
CREATE TABLE learn_algorithms (
    id          UUID PRIMARY KEY,
    node_id     UUID NOT NULL REFERENCES learn_nodes(id) ON DELETE CASCADE,
    notation    TEXT NOT NULL,
    name        TEXT NOT NULL DEFAULT '',
    notes       TEXT NOT NULL DEFAULT '',
    sort_order  INTEGER NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX learn_algorithms_node_idx ON learn_algorithms (node_id, sort_order);
