package structs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDirTree(t *testing.T) {
	nodes := []*DirTreeNode{
		{ID: "documents", ParentID: "", IsDir: true},
		{ID: "doc1", ParentID: "documents", IsDir: false},
		{ID: "doc2", ParentID: "documents", IsDir: false},
		{ID: "backup", ParentID: "documents", IsDir: true},
		{ID: "old_doc1", ParentID: "backup", IsDir: false},
		{ID: "img1", ParentID: "image", IsDir: false},
		{ID: "image", ParentID: "", IsDir: true},
	}

	tree, err := CreeateDirTree(nodes)
	require.NoError(t, err)

	t.Run("print dir tree", func(t *testing.T) {
		PrettyPrintDirTree(tree.Root)
	})

	t.Run("append node to dir tree", func(t *testing.T) {
		for _, node := range []*DirTreeNode{
			{ID: "old_doc2", ParentID: "backup", IsDir: false},
			{ID: "music", ParentID: "", IsDir: true},
		} {
			err := tree.AppendNode(node)
			assert.NoError(t, err)
		}

		PrettyPrintDirTree(tree.Root)
	})
}
