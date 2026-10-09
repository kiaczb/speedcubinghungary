package learn

import (
	"database/sql"
	"os"
	"testing"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func smokeService(t *testing.T) (*Service, *sql.DB) {
	databaseURL := os.Getenv("SMOKE_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set SMOKE_DATABASE_URL to run the learn repository against a real Postgres")
	}

	db, err := sql.Open("pgx", databaseURL)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	require.NoError(t, db.Ping())

	return NewService(NewPostgresRepository(db)), db
}

func smokeRoot(t *testing.T, service *Service) Node {
	root, err := service.CreateNode(NodeInput{Title: "Smoke", Slug: "smoke-" + uuid.NewString()[:8]})
	require.NoError(t, err)
	t.Cleanup(func() { service.DeleteNode(root.ID) })
	return root
}

func child(t *testing.T, service *Service, parent Node, title string) Node {
	n, err := service.CreateNode(NodeInput{ParentID: &parent.ID, Title: title})
	require.NoError(t, err)
	return n
}

func TestSmokeDatabase_HierarchyOfAnyDepth(t *testing.T) {
	service, db := smokeService(t)
	root := smokeRoot(t, service)

	cfop := child(t, service, root, "CFOP")
	oll := child(t, service, cfop, "OLL")
	oll24 := child(t, service, oll, "OLL 24")
	assert.Equal(t, []string{root.Slug, "cfop", "oll", "oll-24"}, oll24.Path)

	first, err := service.CreateAlgorithm(oll24.ID, AlgorithmInput{Notation: "r U R' U' r' F R F'", Name: "Fő"})
	require.NoError(t, err)
	_, err = service.CreateAlgorithm(oll24.ID, AlgorithmInput{Notation: "(r U R' U') (r' F R F')"})
	require.NoError(t, err)

	level, err := service.GetLevel(oll24.Path)
	require.NoError(t, err)
	assert.Equal(t, "OLL 24", level.Node.Title)
	assert.Len(t, level.Ancestors, 3)
	assert.Empty(t, level.Children)
	require.Len(t, level.Algorithms, 2)
	assert.Equal(t, first.ID, level.Algorithms[0].ID)

	ns := child(t, service, root, "NS")
	_, err = service.CreateAlgorithm(ns.ID, AlgorithmInput{Notation: "R L R' L'"})
	require.NoError(t, err)
	short, err := service.GetLevel(ns.Path)
	require.NoError(t, err)
	assert.Len(t, short.Ancestors, 1, "két szint után közvetlenül algoritmus")
	assert.Len(t, short.Algorithms, 1)

	top, err := service.GetLevel([]string{})
	require.NoError(t, err)
	assert.Nil(t, top.Node)
	assert.Contains(t, slugs(top.Children), root.Slug)

	_, err = service.GetLevel([]string{root.Slug, "nincs-ilyen"})
	assert.ErrorIs(t, err, ErrNodeNotFound)

	require.NoError(t, service.DeleteNode(cfop.ID))
	_, err = service.GetLevel(oll24.Path)
	assert.ErrorIs(t, err, ErrNodeNotFound)

	var orphans int
	require.NoError(t, db.QueryRow(`SELECT count(*) FROM learn_algorithms WHERE node_id = $1`, oll24.ID).Scan(&orphans))
	assert.Zero(t, orphans, "a törlés a teljes részfát és az algoritmusait is viszi")
}

func TestSmokeDatabase_TreeInvariants(t *testing.T) {
	service, _ := smokeService(t)
	root := smokeRoot(t, service)

	cfop := child(t, service, root, "CFOP")
	oll := child(t, service, cfop, "OLL")
	leaf := child(t, service, oll, "OLL 21")
	_, err := service.CreateAlgorithm(leaf.ID, AlgorithmInput{Notation: "R U2 R' U' R U R' U' R U' R'"})
	require.NoError(t, err)

	_, err = service.CreateNode(NodeInput{ParentID: &leaf.ID, Title: "Alszint"})
	assert.ErrorIs(t, err, ErrParentHasAlgorithms)

	_, err = service.CreateAlgorithm(oll.ID, AlgorithmInput{Notation: "R U R'"})
	assert.ErrorIs(t, err, ErrNodeHasChildren)

	_, err = service.CreateNode(NodeInput{ParentID: &cfop.ID, Title: "oll"})
	assert.ErrorIs(t, err, ErrSlugTaken)

	_, err = service.UpdateNode(cfop.ID, NodeInput{ParentID: &oll.ID, Title: "CFOP"})
	assert.ErrorIs(t, err, ErrMoveIntoOwnSubtree)

	_, err = service.UpdateNode(cfop.ID, NodeInput{ParentID: &cfop.ID, Title: "CFOP"})
	assert.ErrorIs(t, err, ErrMoveIntoOwnSubtree)

	_, err = service.UpdateNode(oll.ID, NodeInput{ParentID: &leaf.ID, Title: "OLL"})
	assert.ErrorIs(t, err, ErrMoveIntoOwnSubtree)

	missing := uuid.NewString()
	_, err = service.CreateNode(NodeInput{ParentID: &missing, Title: "Árva"})
	assert.ErrorIs(t, err, ErrParentNotFound)

	_, err = service.CreateAlgorithm(missing, AlgorithmInput{Notation: "R"})
	assert.ErrorIs(t, err, ErrNodeNotFound)
}

func TestSmokeDatabase_MoveRenameAndReorder(t *testing.T) {
	service, _ := smokeService(t)
	root := smokeRoot(t, service)

	cfop := child(t, service, root, "CFOP")
	roux := child(t, service, root, "Roux")
	oll := child(t, service, cfop, "OLL")
	leaf := child(t, service, oll, "OLL 24")

	moved, err := service.UpdateNode(oll.ID, NodeInput{ParentID: &roux.ID, Title: "Orientáció", Slug: "orientacio", Icon: "event-333", ImageURL: "https://example.com/oll.svg"})
	require.NoError(t, err)
	assert.Equal(t, "event-333", moved.Icon)
	assert.Equal(t, "https://example.com/oll.svg", moved.ImageURL)
	assert.Equal(t, []string{root.Slug, "roux", "orientacio"}, moved.Path)
	assert.Equal(t, 1, moved.ChildCount)

	level, err := service.GetLevel([]string{root.Slug, "roux", "orientacio", "oll-24"})
	require.NoError(t, err)
	assert.Equal(t, leaf.ID, level.Node.ID, "a részfa a szülővel együtt költözik")
	assert.Equal(t, "event-333", level.Ancestors[2].Icon)

	renamedRoot, err := service.UpdateNode(root.ID, NodeInput{Title: "Smoke átnevezve", Slug: root.Slug})
	require.NoError(t, err)
	assert.Equal(t, "Smoke átnevezve", renamedRoot.Title)
	assert.Nil(t, renamedRoot.ParentID, "a szülő nélküli szerkesztés a legfelső szinten hagyja")

	require.NoError(t, service.ReorderNodes(&root.ID, []string{roux.ID, cfop.ID}))
	ordered, err := service.GetLevel([]string{root.Slug})
	require.NoError(t, err)
	assert.Equal(t, []string{"roux", "cfop"}, slugs(ordered.Children))

	err = service.ReorderNodes(&root.ID, []string{roux.ID})
	assert.ErrorIs(t, err, ErrOrderMismatch, "hiányzó elem")

	err = service.ReorderNodes(&root.ID, []string{roux.ID, leaf.ID})
	assert.ErrorIs(t, err, ErrOrderMismatch, "nem ide tartozó elem")

	a1, err := service.CreateAlgorithm(leaf.ID, AlgorithmInput{Notation: "R U R'"})
	require.NoError(t, err)
	a2, err := service.CreateAlgorithm(leaf.ID, AlgorithmInput{Notation: "L' U' L"})
	require.NoError(t, err)

	require.NoError(t, service.ReorderAlgorithms(leaf.ID, []string{a2.ID, a1.ID}))
	updated, err := service.UpdateAlgorithm(a1.ID, AlgorithmInput{Notation: "R U2 R'", Notes: "Ujjrend: mutató"})
	require.NoError(t, err)
	assert.Equal(t, leaf.ID, updated.NodeID)

	withAlgorithms, err := service.GetLevel(level.Node.Path)
	require.NoError(t, err)
	require.Len(t, withAlgorithms.Algorithms, 2)
	assert.Equal(t, a2.ID, withAlgorithms.Algorithms[0].ID)
	assert.Equal(t, "R U2 R'", withAlgorithms.Algorithms[1].Notation)
	assert.Equal(t, "Ujjrend: mutató", withAlgorithms.Algorithms[1].Notes)

	require.NoError(t, service.DeleteAlgorithm(a2.ID))
	assert.ErrorIs(t, service.DeleteAlgorithm(a2.ID), ErrAlgorithmNotFound)

	tree, err := service.GetTree()
	require.NoError(t, err)
	var found bool
	for _, n := range tree {
		if n.ID == leaf.ID {
			found = true
			assert.Equal(t, []string{root.Slug, "roux", "orientacio", "oll-24"}, n.Path)
		}
	}
	assert.True(t, found)
}

func TestSmokeDatabase_DepthLimit(t *testing.T) {
	service, _ := smokeService(t)
	root := smokeRoot(t, service)

	var almostDeepest Node
	deepest := root
	for depth := 2; depth <= MaxDepth; depth++ {
		almostDeepest, deepest = deepest, child(t, service, deepest, "Szint")
	}
	assert.Len(t, deepest.Path, MaxDepth)

	_, err := service.CreateNode(NodeInput{ParentID: &deepest.ID, Title: "Túl mély"})
	assert.ErrorIs(t, err, ErrTooDeep)

	level, err := service.GetLevel(deepest.Path)
	require.NoError(t, err, "a legmélyebb megengedett szint is elérhető útvonallal")
	assert.Equal(t, deepest.ID, level.Node.ID)

	branch := child(t, service, root, "Ág")
	child(t, service, branch, "Levele")
	_, err = service.UpdateNode(branch.ID, NodeInput{ParentID: &almostDeepest.ID, Title: "Ág"})
	assert.ErrorIs(t, err, ErrTooDeep, "a teljes részfa magasságát nézi, nem csak az áthelyezett elemet")

	leaf := child(t, service, root, "Levél")
	moved, err := service.UpdateNode(leaf.ID, NodeInput{ParentID: &almostDeepest.ID, Title: "Levél"})
	require.NoError(t, err, "egy levél ugyanoda még pont befér")
	assert.Len(t, moved.Path, MaxDepth)
}

func slugs(nodes []Node) []string {
	result := make([]string, len(nodes))
	for i, n := range nodes {
		result[i] = n.Slug
	}
	return result
}
