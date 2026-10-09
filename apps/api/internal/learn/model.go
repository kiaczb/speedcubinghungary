package learn

type Node struct {
	ID             string   `json:"id"`
	ParentID       *string  `json:"parentId"`
	Slug           string   `json:"slug"`
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	Icon           string   `json:"icon"`
	ImageURL       string   `json:"imageUrl"`
	Path           []string `json:"path"`
	ChildCount     int      `json:"childCount"`
	AlgorithmCount int      `json:"algorithmCount"`
}

type Algorithm struct {
	ID       string `json:"id"`
	NodeID   string `json:"nodeId"`
	Notation string `json:"notation"`
	Name     string `json:"name"`
	Notes    string `json:"notes"`
}

type Level struct {
	Node       *Node       `json:"node"`
	Ancestors  []Node      `json:"ancestors"`
	Children   []Node      `json:"children"`
	Algorithms []Algorithm `json:"algorithms"`
}

type NodeInput struct {
	ParentID    *string `json:"parentId"`
	Title       string  `json:"title"`
	Slug        string  `json:"slug"`
	Description string  `json:"description"`
	Icon        string  `json:"icon"`
	ImageURL    string  `json:"imageUrl"`
}

type AlgorithmInput struct {
	Notation string `json:"notation"`
	Name     string `json:"name"`
	Notes    string `json:"notes"`
}
