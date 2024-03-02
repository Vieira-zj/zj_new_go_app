package structs_test

import (
	"encoding/json"
	"sort"
	"strings"
	"testing"
)

type treeNode struct {
	Name         string               `json:"name"`
	ChildrenDict map[string]*treeNode `json:"-"`
	Children     []*treeNode          `json:"children,omitempty"`
}

func newTreeNode(name string) *treeNode {
	return &treeNode{
		Name:         name,
		ChildrenDict: make(map[string]*treeNode, 1),
		Children:     make([]*treeNode, 0, 1),
	}
}

func buildTreeList(groups []string) *treeNode {
	var root *treeNode
	for _, group := range groups {
		items := strings.Split(group, ".")
		curNode := root
		for _, name := range items {
			if root == nil {
				root = newTreeNode(name)
				curNode = root
				continue
			}
			if name == "root" {
				continue
			}

			node, ok := curNode.ChildrenDict[name]
			if ok {
				curNode = node
				continue
			}

			newNode := newTreeNode(name)
			curNode.ChildrenDict[name] = newNode
			curNode.Children = append(curNode.Children, newNode)
			curNode = newNode
		}
	}
	return root
}

type treeNodeSortedByName []*treeNode

func (n treeNodeSortedByName) Len() int           { return len(n) }
func (n treeNodeSortedByName) Less(i, j int) bool { return n[i].Name < n[j].Name }
func (n treeNodeSortedByName) Swap(i, j int)      { n[i], n[j] = n[j], n[i] }

func sortTreeList(node *treeNode) {
	sort.Sort(treeNodeSortedByName(node.Children))
	for _, node := range node.Children {
		sortTreeList(node)
	}
}

func TestSortSlice(t *testing.T) {
	t.Run("sort slice of string", func(t *testing.T) {
		s := []string{"foo", "shop", "bar"}
		sort.Strings(s)
		t.Log("sort string:", s)
	})

	t.Run("sort slice of node", func(t *testing.T) {
		nodes := []*treeNode{
			{Name: "foo"},
			{Name: "bar"},
			{Name: "shop"},
		}
		sort.Slice(nodes, func(i, j int) bool {
			return nodes[i].Name < nodes[j].Name
		})

		b, _ := json.Marshal(&nodes)
		t.Logf("sort node: %s", b)
	})
}

func TestBuildTreeList(t *testing.T) {
	// create case treelist from full case names
	groups := []string{
		"root.shop.order.test7",
		"root.foo.bar.group2",
		"root.foo.bar.group.subc",
		"root.shop.item.testy",
		"root.foo.bar.group1",
		"root.shop.order.test3",
		"root.shop.item.testx",
		"root.foo.bar.group.suba",
	}
	root := buildTreeList(groups)

	t.Run("build treelist", func(t *testing.T) {
		b, _ := json.MarshalIndent(root, "", "  ")
		t.Logf("root:\n%s", b)
	})

	t.Run("sort treelist", func(t *testing.T) {
		sortTreeList(root)
		b, _ := json.MarshalIndent(root, "", "  ")
		t.Logf("sort root:\n%s", b)
	})
}
