package structs_test

import (
	"testing"

	"demo.apps/structs"
)

func TestTrieTree(t *testing.T) {
	tree := structs.NewTrieTree()

	key := "hello"
	if got := tree.Find(key); got != false {
		t.Errorf("want: %v, got: %v", false, got)
	}

	tree.Add(key)
	if got := tree.Find(key); got != true {
		t.Errorf("want: %v, got: %v", true, got)
	}
	if got := tree.Find("he"); got != false {
		t.Errorf("want: %v, got: %v", false, got)
	}

	key = "he"
	tree.Add(key)
	if got := tree.Find(key); got != true {
		t.Errorf("want: %v, got: %v", true, got)
	}
	t.Log("test trie tree done")
}
