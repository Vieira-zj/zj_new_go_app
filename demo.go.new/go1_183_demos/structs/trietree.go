package structs

type TrieTree struct {
	isPath   bool
	children map[byte]*TrieTree
}

func NewTrieTree() *TrieTree {
	return &TrieTree{false, make(map[byte]*TrieTree)}
}

// Add 添加一个路由到 Trie Tree.
func (t *TrieTree) Add(path string) {
	parent := t
	// 逐个 byte 加入到 Trie Tree
	for i := range path {
		if child, ok := parent.children[path[i]]; ok {
			// 如果子节点不为空, 继续向下遍历
			parent = child
		} else {
			child := NewTrieTree()
			parent.children[path[i]] = child
			parent = child
		}
	}

	// 更新当前路由的叶子节点的 IsPath 字段
	parent.isPath = true
}

// Find 返回指定路由是否存在于 Trie Tree 中.
func (t *TrieTree) Find(path string) bool {
	parent := t
	for i := range path {
		if child, ok := parent.children[path[i]]; ok {
			parent = child
		} else {
			return false
		}
	}

	return parent.isPath
}
