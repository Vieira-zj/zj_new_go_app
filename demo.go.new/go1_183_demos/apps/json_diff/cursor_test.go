package jsondiff

import "testing"

func TestCursor(t *testing.T) {
	cursor := cursor{}
	cursor.appendKey("slice")
	cursor.appendIndex(1)
	t.Log("cursor:", cursor.string())

	cursor.rollback()
	cursor.rollback()
	t.Log("cursor:", cursor.string())
}
