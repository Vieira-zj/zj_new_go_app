package structs

import (
	"fmt"
)

const (
	RootNodeID   = "root"
	RootNodePath = "/root"

	Separate = "/"
)

type DirTreeNode struct {
	ID       string // dir name
	ParentID string // parent dir name
	IsDir    bool
	Children []*DirTreeNode
	FullPath string
}

func (n *DirTreeNode) IsRoot() bool {
	return len(n.ParentID) == 0
}

func (n *DirTreeNode) String() string {
	return fmt.Sprintf("{ name=%s, parent=%s, path=%s }", n.ID, n.ParentID, n.FullPath)
}

type DirTree struct {
	Root  *DirTreeNode
	Nodes map[string]*DirTreeNode
	Paths map[string]struct{}
}

// CreeateDirTree create dir tree for nodes list by parent id.
func CreeateDirTree(nodes []*DirTreeNode) (*DirTree, error) {
	nodesMap := make(map[string]*DirTreeNode, len(nodes)+1)
	for _, node := range nodes {
		nodesMap[node.ID] = node
		if node.IsRoot() {
			node.ParentID = RootNodeID
		}
	}

	root := createDirTreeRootNode()
	nodesMap[root.ID] = root

	paths := make(map[string]struct{}, len(nodes))
	for _, node := range nodes {
		if err := initNodeFullPath(nodesMap, node); err != nil {
			return nil, err
		}

		if _, ok := paths[node.FullPath]; ok {
			return nil, fmt.Errorf("dulplicated node, path=%s", node.FullPath)
		}
		paths[node.FullPath] = struct{}{}

		if parent, ok := nodesMap[node.ParentID]; ok && parent.IsDir {
			parent.Children = append(parent.Children, node)
		} else {
			return nil, fmt.Errorf("parent dir node is not found, id=%s", node.ParentID)
		}
	}

	return &DirTree{
		Root:  root,
		Nodes: nodesMap,
		Paths: paths,
	}, nil
}

func (t *DirTree) AppendNode(node *DirTreeNode) error {
	if node.IsRoot() {
		node.ParentID = RootNodeID
	}

	if err := initNodeFullPath(t.Nodes, node); err != nil {
		return err
	}

	if _, ok := t.Paths[node.FullPath]; ok {
		return fmt.Errorf("dulplicated node, path=%s", node.FullPath)
	}

	if parent, ok := t.Nodes[node.ParentID]; ok && parent.IsDir {
		parent.Children = append(parent.Children, node)
	} else {
		return fmt.Errorf("parent dir node is not found, id=%s", node.ParentID)
	}

	return nil
}

func createDirTreeRootNode() *DirTreeNode {
	return &DirTreeNode{
		ID:       RootNodeID,
		ParentID: "",
		IsDir:    true,
		Children: make([]*DirTreeNode, 0, 4),
		FullPath: RootNodePath,
	}
}

func initNodeFullPath(nodes map[string]*DirTreeNode, node *DirTreeNode) error {
	if node.IsRoot() {
		return nil
	}

	srcNode := node
	path := node.ID

	for !node.IsRoot() {
		if parent, ok := nodes[node.ParentID]; ok && parent.IsDir {
			path = parent.ID + Separate + path
			node = parent
		} else {
			return fmt.Errorf("parent dir node is not found, id=%s", node.ParentID)
		}
	}

	srcNode.FullPath = Separate + path
	return nil
}

func PrettyPrintDirTree(root *DirTreeNode) {
	prettyPrintDirTreeWithPrefix("", root)
}

func prettyPrintDirTreeWithPrefix(prefix string, root *DirTreeNode) {
	fmt.Println(prefix + root.String())

	if root.IsDir && len(root.Children) > 0 {
		for _, child := range root.Children {
			prettyPrintDirTreeWithPrefix(prefix+"\t", child)
		}
	}
}

//
// DirTree2
//

type DirTreeNode2 struct {
	Name       string          `json:"name"`
	ParentPath string          `json:"parent_path"`
	IsDir      bool            `json:"is_dir"`
	Children   []*DirTreeNode2 `json:"children,omitempty"`
}

func (n *DirTreeNode2) String() string {
	return fmt.Sprintf("{ name=%s, path=%s }", n.Name, n.GetPullPath())
}

func (n *DirTreeNode2) GetPullPath() string {
	if n.ParentPath == Separate {
		return n.ParentPath + n.Name
	}
	return n.ParentPath + Separate + n.Name
}

type DirTree2 struct {
	Root  *DirTreeNode2
	Nodes map[string]*DirTreeNode2 // key:full_path
}

func CreeateDirTree2(nodes []*DirTreeNode2) (*DirTree2, error) {
	nodesMap := make(map[string]*DirTreeNode2, len(nodes))
	for _, node := range nodes {
		if len(node.ParentPath) == 0 {
			node.ParentPath = Separate
		}
		if _, ok := nodesMap[node.GetPullPath()]; ok {
			return nil, fmt.Errorf("dulplicated node, path=%s", node.GetPullPath())
		}
		nodesMap[node.GetPullPath()] = node
	}

	root := &DirTreeNode2{
		Name:       "",
		ParentPath: Separate,
		IsDir:      true,
		Children:   make([]*DirTreeNode2, 0, 4),
	}
	nodesMap[root.GetPullPath()] = root

	for _, node := range nodes {
		if parent, ok := nodesMap[node.ParentPath]; ok && parent.IsDir {
			parent.Children = append(parent.Children, node)
		} else {
			return nil, fmt.Errorf("parent dir node is not found, path=%s", node.ParentPath)
		}
	}

	return &DirTree2{
		Root:  root,
		Nodes: nodesMap,
	}, nil
}

func (t *DirTree2) AppendNode(node *DirTreeNode2) error {
	if len(node.ParentPath) == 0 {
		node.ParentPath = Separate
	}

	if _, ok := t.Nodes[node.GetPullPath()]; ok {
		return fmt.Errorf("dulplicated node, path=%s", node.GetPullPath())
	}

	if parent, ok := t.Nodes[node.ParentPath]; ok && parent.IsDir {
		parent.Children = append(parent.Children, node)
		return nil
	}
	return fmt.Errorf("parent dir node is not found, path=%s", node.ParentPath)
}

func PrettyPrintDirTree2(root *DirTreeNode2) {
	prettyPrintDirTreeWithPrefix2("", root)
}

func prettyPrintDirTreeWithPrefix2(prefix string, root *DirTreeNode2) {
	fmt.Println(prefix + root.String())

	if root.IsDir && len(root.Children) > 0 {
		for _, child := range root.Children {
			prettyPrintDirTreeWithPrefix2(prefix+"\t", child)
		}
	}
}

/*
TD: Project Users Structure

Table:

- mgr_project_group_structure.tab: for level1-level3 user groups (e.g. id, department, group_name, parent_path, is_leaf).
- mgr_project_users.tab: for project users info (e.g. id, email, group_path).

Rest API:

- /get_prj_group_struct: returns level1-level3 user groups.
  - because of too much users, here only loads user groups structure.
  - first, it build a singlton user groups structure tree.
- /get_prj_users: returns project users info by group name.
*/
