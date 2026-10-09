package learn

import (
	"github.com/google/uuid"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func childPath(parent []string, slug string) []string {
	path := make([]string, len(parent), len(parent)+1)
	copy(path, parent)
	return append(path, slug)
}

func (s *Service) GetLevel(path []string) (Level, error) {
	if err := validatePath(path); err != nil {
		return Level{}, err
	}

	chain, err := s.repo.ResolvePath(path)
	if err != nil {
		return Level{}, err
	}
	if len(chain) != len(path) {
		return Level{}, ErrNodeNotFound
	}
	for i := range chain {
		chain[i].Path = append([]string{}, path[:i+1]...)
	}

	level := Level{
		Ancestors:  []Node{},
		Algorithms: []Algorithm{},
	}

	var parentID *string
	if len(chain) > 0 {
		node := chain[len(chain)-1]
		level.Node = &node
		level.Ancestors = chain[:len(chain)-1]
		parentID = &node.ID
	}

	level.Children, err = s.repo.GetChildren(parentID)
	if err != nil {
		return Level{}, err
	}
	for i := range level.Children {
		level.Children[i].Path = childPath(path, level.Children[i].Slug)
	}

	if level.Node != nil && level.Node.AlgorithmCount > 0 {
		level.Algorithms, err = s.repo.GetAlgorithms(level.Node.ID)
		if err != nil {
			return Level{}, err
		}
	}

	return level, nil
}

func (s *Service) GetTree() ([]Node, error) {
	nodes, err := s.repo.GetAllNodes()
	if err != nil {
		return nil, err
	}

	childrenOf := make(map[string][]Node, len(nodes))
	for _, n := range nodes {
		parent := ""
		if n.ParentID != nil {
			parent = *n.ParentID
		}
		childrenOf[parent] = append(childrenOf[parent], n)
	}

	tree := make([]Node, 0, len(nodes))
	var walk func(parent string, path []string)
	walk = func(parent string, path []string) {
		for _, n := range childrenOf[parent] {
			n.Path = childPath(path, n.Slug)
			tree = append(tree, n)
			walk(n.ID, n.Path)
		}
	}
	walk("", nil)

	return tree, nil
}

func (s *Service) CreateNode(in NodeInput) (Node, error) {
	in, err := normalizeNode(in)
	if err != nil {
		return Node{}, err
	}

	id := uuid.NewString()
	if err := s.repo.CreateNode(id, in); err != nil {
		return Node{}, err
	}
	return s.repo.GetNode(id)
}

func (s *Service) UpdateNode(id string, in NodeInput) (Node, error) {
	id, err := canonicalID(id, ErrInvalidNodeID)
	if err != nil {
		return Node{}, err
	}
	in, err = normalizeNode(in)
	if err != nil {
		return Node{}, err
	}

	if err := s.repo.UpdateNode(id, in); err != nil {
		return Node{}, err
	}
	return s.repo.GetNode(id)
}

func (s *Service) DeleteNode(id string) error {
	id, err := canonicalID(id, ErrInvalidNodeID)
	if err != nil {
		return err
	}
	return s.repo.DeleteNode(id)
}

func (s *Service) ReorderNodes(parentID *string, ids []string) error {
	if parentID != nil {
		id, err := canonicalID(*parentID, ErrInvalidParentID)
		if err != nil {
			return err
		}
		parentID = &id
	}
	ids, err := normalizeOrder(ids)
	if err != nil {
		return err
	}
	return s.repo.ReorderNodes(parentID, ids)
}

func (s *Service) CreateAlgorithm(nodeID string, in AlgorithmInput) (Algorithm, error) {
	nodeID, err := canonicalID(nodeID, ErrInvalidNodeID)
	if err != nil {
		return Algorithm{}, err
	}
	in, err = normalizeAlgorithm(in)
	if err != nil {
		return Algorithm{}, err
	}

	id := uuid.NewString()
	if err := s.repo.CreateAlgorithm(id, nodeID, in); err != nil {
		return Algorithm{}, err
	}
	return Algorithm{ID: id, NodeID: nodeID, Notation: in.Notation, Name: in.Name, Notes: in.Notes}, nil
}

func (s *Service) UpdateAlgorithm(id string, in AlgorithmInput) (Algorithm, error) {
	id, err := canonicalID(id, ErrInvalidAlgorithmID)
	if err != nil {
		return Algorithm{}, err
	}
	in, err = normalizeAlgorithm(in)
	if err != nil {
		return Algorithm{}, err
	}

	nodeID, err := s.repo.UpdateAlgorithm(id, in)
	if err != nil {
		return Algorithm{}, err
	}
	return Algorithm{ID: id, NodeID: nodeID, Notation: in.Notation, Name: in.Name, Notes: in.Notes}, nil
}

func (s *Service) DeleteAlgorithm(id string) error {
	id, err := canonicalID(id, ErrInvalidAlgorithmID)
	if err != nil {
		return err
	}
	return s.repo.DeleteAlgorithm(id)
}

func (s *Service) ReorderAlgorithms(nodeID string, ids []string) error {
	nodeID, err := canonicalID(nodeID, ErrInvalidNodeID)
	if err != nil {
		return err
	}
	ids, err = normalizeOrder(ids)
	if err != nil {
		return err
	}
	return s.repo.ReorderAlgorithms(nodeID, ids)
}
