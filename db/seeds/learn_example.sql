-- Példaadat a Tanulás fához — NEM migráció, és semmi nem futtatja automatikusan.
-- Kézi futtatás (dev, bash):
--
--   docker compose -f infra/docker-compose.yml -f infra/docker-compose.dev.yml \
--     --env-file .env.dev exec -T db sh -c 'psql -U "$POSTGRES_USER" -d "$POSTGRES_DB"' \
--     < db/seeds/learn_example.sql
--
-- Újrafuttatható: a meglévő elemeket (szülő + slug alapján) és algoritmusokat
-- (elem + jelölés alapján) nem duplikálja, és a már kitöltött ikont/képet nem
-- írja felül — csak az üreset tölti ki.
--
-- Két, szándékosan eltérő mélységű ág, hogy látszódjon: a felület ugyanúgy
-- kezeli a négy- és a kétszintes utat.
--   3x3 → CFOP → OLL → OLL 21…27      (négy szint, aztán algoritmusok)
--   Skewb → NS                         (két szint, aztán algoritmusok)
-- ⚠️ A Skewb ág csak a szerkezetet mutatja: a Sledgehammer egy alapszekvencia,
-- nem NS-eset. Cseréld valódi tartalomra, mielőtt bárhol megjelenik.
--
-- A képek VisualCube-ábrák (https://visualcube.api.cubing.net), az eset SAJÁT
-- fő algoritmusából generálva (case = az algoritmus inverze a kirakott
-- kockán), így kép és algoritmus nem csúszhat szét. A VisualCube szerint a
-- generált képek szabadon másolhatók és terjeszthetők.

CREATE FUNCTION pg_temp.visualcube(p_stage TEXT, p_alg TEXT DEFAULT '', p_plan BOOLEAN DEFAULT TRUE)
RETURNS TEXT AS $$
    SELECT 'https://visualcube.api.cubing.net/visualcube.php?fmt=svg&size=150&pzl=3'
        || '&stage=' || p_stage
        || CASE WHEN p_plan THEN '&view=plan' ELSE '' END
        || CASE WHEN p_alg = '' THEN ''
                ELSE '&case=' || replace(replace(p_alg, ' ', '%20'), '''', '%27') END
$$ LANGUAGE sql IMMUTABLE;

CREATE FUNCTION pg_temp.learn_node(
    p_parent UUID, p_slug TEXT, p_title TEXT, p_description TEXT,
    p_icon TEXT DEFAULT '', p_image TEXT DEFAULT ''
)
RETURNS UUID AS $$
DECLARE
    v_id UUID;
BEGIN
    SELECT id INTO v_id
    FROM learn_nodes
    WHERE parent_id IS NOT DISTINCT FROM p_parent AND slug = p_slug;

    IF v_id IS NULL THEN
        v_id := gen_random_uuid();
        INSERT INTO learn_nodes (id, parent_id, slug, title, description, icon, image_url, sort_order)
        VALUES (v_id, p_parent, p_slug, p_title, p_description, p_icon, p_image, (
            SELECT COALESCE(MAX(sort_order) + 1, 0)
            FROM learn_nodes
            WHERE parent_id IS NOT DISTINCT FROM p_parent
        ));
    ELSE
        UPDATE learn_nodes
        SET icon = CASE WHEN icon = '' THEN p_icon ELSE icon END,
            image_url = CASE WHEN image_url = '' THEN p_image ELSE image_url END
        WHERE id = v_id;
    END IF;

    RETURN v_id;
END
$$ LANGUAGE plpgsql;

CREATE FUNCTION pg_temp.learn_algorithm(p_node UUID, p_notation TEXT, p_name TEXT, p_notes TEXT)
RETURNS VOID AS $$
BEGIN
    INSERT INTO learn_algorithms (id, node_id, notation, name, notes, sort_order)
    SELECT gen_random_uuid(), p_node, p_notation, p_name, p_notes, (
        SELECT COALESCE(MAX(sort_order) + 1, 0) FROM learn_algorithms WHERE node_id = p_node
    )
    WHERE NOT EXISTS (
        SELECT 1 FROM learn_algorithms WHERE node_id = p_node AND notation = p_notation
    );
END
$$ LANGUAGE plpgsql;

-- Egy eset: az elem, a képe (a fő algoritmusból) és maga a fő algoritmus.
CREATE FUNCTION pg_temp.learn_case(
    p_parent UUID, p_slug TEXT, p_title TEXT, p_description TEXT,
    p_stage TEXT, p_alg TEXT, p_alg_name TEXT DEFAULT 'Fő'
)
RETURNS UUID AS $$
DECLARE
    v_id UUID;
BEGIN
    v_id := pg_temp.learn_node(p_parent, p_slug, p_title, p_description, '', pg_temp.visualcube(p_stage, p_alg));
    PERFORM pg_temp.learn_algorithm(v_id, p_alg, p_alg_name, '');
    RETURN v_id;
END
$$ LANGUAGE plpgsql;

DO $$
DECLARE
    cube UUID;
    cfop UUID;
    oll UUID;
    pll UUID;
    c UUID;
    skewb UUID;
    ns UUID;
BEGIN
    cube := pg_temp.learn_node(NULL, '3x3', '3x3', 'A klasszikus Rubik-kocka — a legtöbb módszer és algoritmus ide tartozik.', 'event-333');

    cfop := pg_temp.learn_node(cube, 'cfop', 'CFOP', 'Kereszt, F2L, OLL, PLL — a legelterjedtebb speedcubing-módszer.', '', pg_temp.visualcube('f2l', '', FALSE));
    PERFORM pg_temp.learn_node(cube, 'roux', 'Roux', 'Blokképítésre épülő módszer, kevés lépéssel.', '', pg_temp.visualcube('f2b', '', FALSE));
    PERFORM pg_temp.learn_node(cube, 'zz', 'ZZ', 'Élorientációval kezdő módszer, forgatás nélküli F2L-lel.', '', pg_temp.visualcube('line', '', FALSE));

    oll := pg_temp.learn_node(cfop, 'oll', 'OLL', 'Az utolsó réteg tájolása: a felső lap egyszínűvé tétele.', '', pg_temp.visualcube('oll', '', FALSE));
    pll := pg_temp.learn_node(cfop, 'pll', 'PLL', 'Az utolsó réteg permutációja: a darabok helyre rakása.', '', pg_temp.visualcube('pll', '', FALSE));

    c := pg_temp.learn_case(oll, 'oll-21', 'OLL 21', 'H — kereszt, egyik sarok sem néz felfelé (két fényszórópár).', 'oll', 'R U2 R'' U'' R U R'' U'' R U'' R''');
    PERFORM pg_temp.learn_algorithm(c, 'R U R'' U R U'' R'' U R U2 R''', '', 'Dupla Sune.');
    PERFORM pg_temp.learn_algorithm(c, 'F R U R'' U'' R U R'' U'' R U R'' U'' F''', '', 'Háromszor R U R'' U'', F-fel keretezve.');

    PERFORM pg_temp.learn_case(oll, 'oll-22', 'OLL 22', 'Pi — kereszt, egyik sarok sem néz felfelé (egy fényszórópár).', 'oll', 'R U2 R2 U'' R2 U'' R2 U2 R');

    c := pg_temp.learn_case(oll, 'oll-23', 'OLL 23', 'U / fényszóró — kereszt, két szomszédos sarok néz felfelé.', 'oll', 'R2 D'' R U2 R'' D R U2 R');
    PERFORM pg_temp.learn_algorithm(c, 'R2 D R'' U2 R D'' R'' U2 R''', '', '');

    PERFORM pg_temp.learn_case(oll, 'oll-24', 'OLL 24', 'T / kaméleon — kereszt, két szomszédos sarok néz felfelé.', 'oll', 'r U R'' U'' r'' F R F''');

    PERFORM pg_temp.learn_case(oll, 'oll-25', 'OLL 25', 'L / csokornyakkendő — kereszt, két átlós sarok néz felfelé.', 'oll', 'F'' r U R'' U'' r'' F R');

    c := pg_temp.learn_case(oll, 'oll-26', 'OLL 26', 'Antisune — kereszt, egy sarok néz felfelé.', 'oll', 'R U2 R'' U'' R U'' R''');
    PERFORM pg_temp.learn_algorithm(c, 'L'' U'' L U'' L'' U2 L', 'Balkezes', '');

    c := pg_temp.learn_case(oll, 'oll-27', 'OLL 27', 'Sune — kereszt, egy sarok néz felfelé.', 'oll', 'R U R'' U R U2 R''');
    PERFORM pg_temp.learn_algorithm(c, 'L'' U2 L U L'' U L', 'Balkezes', '');

    PERFORM pg_temp.learn_case(pll, 't-perm', 'T-perm', 'Két szomszédos sarok és két szemközti él cseréje.', 'pll', 'R U R'' U'' R'' F R2 U'' R'' U'' R U R'' F''');

    c := pg_temp.learn_case(pll, 'ua-perm', 'Ua-perm', 'Három él körbeforgatása.', 'pll', 'M2 U M U2 M'' U M2', 'M-szeletes');
    PERFORM pg_temp.learn_algorithm(c, 'R U'' R U R U R U'' R'' U'' R2', '', '');

    c := pg_temp.learn_case(pll, 'ub-perm', 'Ub-perm', 'Három él körbeforgatása, az Ua-val ellentétes irányban.', 'pll', 'M2 U'' M U2 M'' U'' M2', 'M-szeletes');
    PERFORM pg_temp.learn_algorithm(c, 'R2 U R U R'' U'' R'' U'' R'' U R''', '', '');

    PERFORM pg_temp.learn_case(pll, 'h-perm', 'H-perm', 'Két-két szemközti él cseréje.', 'pll', 'M2 U M2 U2 M2 U M2');

    PERFORM pg_temp.learn_case(pll, 'z-perm', 'Z-perm', 'Két-két szomszédos él cseréje.', 'pll', 'M'' U M2 U M2 U M'' U2 M2');

    skewb := pg_temp.learn_node(NULL, 'skewb', 'Skewb', 'Sarkok körül forgó kirakó.', 'event-skewb');
    ns := pg_temp.learn_node(skewb, 'ns', 'NS', 'Példa egy rövid útra: a kirakó után közvetlenül algoritmusok következnek.');
    PERFORM pg_temp.learn_algorithm(ns, 'R'' L R L''', 'Sledgehammer', 'Példaadat — cseréld a valódi NS-esetekre.');
END
$$;
