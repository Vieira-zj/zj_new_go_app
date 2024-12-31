package structs

import (
	"encoding/json"
	"fmt"
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

func TestDirTree2(t *testing.T) {
	nodes := []*DirTreeNode2{
		{Name: "documents", ParentPath: "/", IsDir: true},
		{Name: "doc1", ParentPath: "/documents", IsDir: false},
		{Name: "doc2", ParentPath: "/documents", IsDir: false},
		{Name: "backup", ParentPath: "/documents", IsDir: true},
		{Name: "old_doc1", ParentPath: "/documents/backup", IsDir: false},
		{Name: "img1", ParentPath: "/image", IsDir: false},
		{Name: "image", ParentPath: "", IsDir: true},
	}

	tree, err := CreeateDirTree2(nodes)
	require.NoError(t, err)

	t.Run("print dir tree", func(t *testing.T) {
		PrettyPrintDirTree2(tree.Root)
	})

	t.Run("append node to dir tree", func(t *testing.T) {
		for _, node := range []*DirTreeNode2{
			{Name: "old_doc2", ParentPath: "/documents/backup", IsDir: false},
			{Name: "music", ParentPath: "", IsDir: true},
		} {
			err := tree.AppendNode(node)
			assert.NoError(t, err)
		}

		PrettyPrintDirTree2(tree.Root)
	})

	t.Run("print dir tree as json", func(t *testing.T) {
		b, err := json.MarshalIndent(tree.Root, "", "  ")
		assert.NoError(t, err)
		fmt.Println("tree:\n", string(b))
	})
}
