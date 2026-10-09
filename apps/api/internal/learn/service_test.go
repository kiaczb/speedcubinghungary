package learn

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubRepository struct {
	chain      []Node
	children   []Node
	all        []Node
	algorithms []Algorithm

	resolvedPath     []string
	childrenOf       *string
	childrenAsked    bool
	algorithmsAsked  string
	createdID        string
	created          NodeInput
	updated          NodeInput
	nodeAsked        string
	createdAlgorithm AlgorithmInput
}

func (s *stubRepository) ResolvePath(path []string) ([]Node, error) {
	s.resolvedPath = path
	if len(path) < len(s.chain) {
		return s.chain[:len(path)], nil
	}
	return s.chain, nil
}

func (s *stubRepository) GetChildren(parentID *string) ([]Node, error) {
	s.childrenAsked, s.childrenOf = true, parentID
	return s.children, nil
}

func (s *stubRepository) GetAlgorithms(nodeID string) ([]Algorithm, error) {
	s.algorithmsAsked = nodeID
	return s.algorithms, nil
}

func (s *stubRepository) GetAllNodes() ([]Node, error) { return s.all, nil }

func (s *stubRepository) GetNode(id string) (Node, error) {
	s.nodeAsked = id
	return Node{ID: id}, nil
}

func (s *stubRepository) CreateNode(id string, in NodeInput) error {
	s.createdID, s.created = id, in
	return nil
}

func (s *stubRepository) UpdateNode(_ string, in NodeInput) error {
	s.updated = in
	return nil
}

func (s *stubRepository) DeleteNode(string) error              { return nil }
func (s *stubRepository) ReorderNodes(*string, []string) error { return nil }

func (s *stubRepository) CreateAlgorithm(_, _ string, in AlgorithmInput) error {
	s.createdAlgorithm = in
	return nil
}

func (s *stubRepository) UpdateAlgorithm(string, AlgorithmInput) (string, error) {
	return "node", nil
}

func (s *stubRepository) DeleteAlgorithm(string) error             { return nil }
func (s *stubRepository) ReorderAlgorithms(string, []string) error { return nil }

func ptr(s string) *string { return &s }

func TestGetLevel_RootListsTheTopLevelWithPaths(t *testing.T) {
	repo := &stubRepository{
		chain:    []Node{},
		children: []Node{{ID: "a", Slug: "3x3"}, {ID: "b", Slug: "skewb"}},
	}

	level, err := NewService(repo).GetLevel([]string{})
	require.NoError(t, err)

	assert.Nil(t, level.Node)
	assert.Empty(t, level.Ancestors)
	assert.Nil(t, repo.childrenOf, "a legfelső szint gyerekei: parent_id IS NULL")
	require.Len(t, level.Children, 2)
	assert.Equal(t, []string{"3x3"}, level.Children[0].Path)
	assert.Equal(t, []string{"skewb"}, level.Children[1].Path)
	assert.Empty(t, repo.algorithmsAsked)
}

func TestGetLevel_AnyDepthUsesTheSameShape(t *testing.T) {
	repo := &stubRepository{
		chain: []Node{
			{ID: "n1", Slug: "3x3"},
			{ID: "n2", Slug: "cfop"},
			{ID: "n3", Slug: "oll"},
			{ID: "n4", Slug: "oll-24", AlgorithmCount: 2},
		},
		children:   []Node{},
		algorithms: []Algorithm{{ID: "a1"}, {ID: "a2"}},
	}
	path := []string{"3x3", "cfop", "oll", "oll-24"}

	level, err := NewService(repo).GetLevel(path)
	require.NoError(t, err)

	require.NotNil(t, level.Node)
	assert.Equal(t, "n4", level.Node.ID)
	assert.Equal(t, path, level.Node.Path)
	require.Len(t, level.Ancestors, 3)
	assert.Equal(t, []string{"3x3"}, level.Ancestors[0].Path)
	assert.Equal(t, []string{"3x3", "cfop", "oll"}, level.Ancestors[2].Path)
	assert.Equal(t, "n4", *repo.childrenOf)
	assert.Equal(t, "n4", repo.algorithmsAsked)
	assert.Len(t, level.Algorithms, 2)
}

func TestGetLevel_ShortPathSkewbNS(t *testing.T) {
	repo := &stubRepository{
		chain: []Node{
			{ID: "s1", Slug: "skewb"},
			{ID: "s2", Slug: "ns", AlgorithmCount: 1},
		},
		children:   []Node{},
		algorithms: []Algorithm{{ID: "a1"}},
	}

	level, err := NewService(repo).GetLevel([]string{"skewb", "ns"})
	require.NoError(t, err)

	assert.Equal(t, []string{"skewb", "ns"}, level.Node.Path)
	assert.Len(t, level.Ancestors, 1)
	assert.Len(t, level.Algorithms, 1)
}

func TestGetLevel_ChildPathsDoNotShareTheBackingArray(t *testing.T) {
	repo := &stubRepository{
		chain:    []Node{{ID: "n1", Slug: "3x3"}},
		children: []Node{{ID: "c1", Slug: "cfop"}, {ID: "c2", Slug: "roux"}},
	}

	level, err := NewService(repo).GetLevel([]string{"3x3"})
	require.NoError(t, err)

	assert.Equal(t, []string{"3x3", "cfop"}, level.Children[0].Path)
	assert.Equal(t, []string{"3x3", "roux"}, level.Children[1].Path)
}

func TestGetLevel_UnresolvedPathIsNotFound(t *testing.T) {
	repo := &stubRepository{chain: []Node{{ID: "n1", Slug: "3x3"}}}

	_, err := NewService(repo).GetLevel([]string{"3x3", "nincs-ilyen"})

	assert.ErrorIs(t, err, ErrNodeNotFound)
	assert.False(t, repo.childrenAsked)
}

func TestGetLevel_InvalidPathNeverReachesTheDatabase(t *testing.T) {
	repo := &stubRepository{}

	_, err := NewService(repo).GetLevel([]string{"3x3", "../etc"})

	assert.ErrorIs(t, err, ErrValidation)
	assert.Nil(t, repo.resolvedPath)
}

func TestGetTree_IsDepthFirstWithFullPaths(t *testing.T) {
	repo := &stubRepository{all: []Node{
		{ID: "a", Slug: "3x3"},
		{ID: "c", ParentID: ptr("b"), Slug: "ns"},
		{ID: "b", Slug: "skewb"},
		{ID: "d", ParentID: ptr("a"), Slug: "cfop"},
		{ID: "e", ParentID: ptr("d"), Slug: "oll"},
	}}

	tree, err := NewService(repo).GetTree()
	require.NoError(t, err)

	ids := make([]string, len(tree))
	for i, n := range tree {
		ids[i] = n.ID
	}
	assert.Equal(t, []string{"a", "d", "e", "b", "c"}, ids)
	assert.Equal(t, []string{"3x3", "cfop", "oll"}, tree[2].Path)
	assert.Equal(t, []string{"skewb", "ns"}, tree[4].Path)
}

func TestCreateNode_StoresTheNormalizedInputAndReturnsTheStoredNode(t *testing.T) {
	repo := &stubRepository{}

	node, err := NewService(repo).CreateNode(NodeInput{Title: " OLL 24 "})
	require.NoError(t, err)

	assert.Equal(t, "OLL 24", repo.created.Title)
	assert.Equal(t, "oll-24", repo.created.Slug)
	assert.NotEmpty(t, repo.createdID)
	assert.Equal(t, repo.createdID, repo.nodeAsked)
	assert.Equal(t, repo.createdID, node.ID)
}

func TestUpdateNode_RejectsANonUUIDBeforeTheDatabase(t *testing.T) {
	repo := &stubRepository{}

	_, err := NewService(repo).UpdateNode("nem-uuid", NodeInput{Title: "CFOP"})

	assert.ErrorIs(t, err, ErrInvalidNodeID)
	assert.Empty(t, repo.updated.Title)
}

func TestCreateAlgorithm_ReturnsTheNormalizedNotation(t *testing.T) {
	repo := &stubRepository{}
	nodeID := "7c9e6679-7425-40de-944b-e07fc1f90ae7"

	algorithm, err := NewService(repo).CreateAlgorithm(nodeID, AlgorithmInput{Notation: " R U  R' "})
	require.NoError(t, err)

	assert.Equal(t, "R U R'", algorithm.Notation)
	assert.Equal(t, "R U R'", repo.createdAlgorithm.Notation)
	assert.Equal(t, nodeID, algorithm.NodeID)
	assert.NotEmpty(t, algorithm.ID)
}
