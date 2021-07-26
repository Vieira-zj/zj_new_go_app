package demos

import (
	"fmt"
	"testing"

	jsonpatch "github.com/evanphx/json-patch"
)

func TestCreateAndApplyMergePatch(t *testing.T) {
	original := []byte(`{"name": "John", "age": 24, "height": 3.21}`)
	target := []byte(`{"name": "Jane", "age": 24}`)

	patch, err := jsonpatch.CreateMergePatch(original, target)
	if err != nil {
		t.Fatal(err)
	}

	alternative := []byte(`{"name": "Tina", "age": 28, "height": 3.75}`)
	modifiedAlternative, err := jsonpatch.MergePatch(alternative, patch)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("patch document: %s\n", patch)
	fmt.Printf("updated alternative doc: %s\n", modifiedAlternative)
}

func TestApplyJsonPatch(t *testing.T) {
	original := []byte(`{"name": "John", "age": 24, "height": 3.21}`)
	patchJSON := []byte(`[
		{"op": "replace", "path": "/name", "value": "Jane"},
		{"op": "remove", "path": "/height"}
	]`)

	patch, err := jsonpatch.DecodePatch(patchJSON)
	if err != nil {
		t.Fatal(err)
	}

	modified, err := patch.Apply(original)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Original document: %s\n", original)
	fmt.Printf("Modified document: %s\n", modified)
}

func TestCompareJsonDocs(t *testing.T) {
	original := []byte(`{"name": "John", "age": 24, "height": 3.21}`)
	similar := []byte(`
		{
			"age": 24,
			"height": 3.21,
			"name": "John"
		}
	`)
	different := []byte(`{"name": "Jane", "age": 20, "height": 3.37}`)

	if jsonpatch.Equal(original, similar) {
		fmt.Println(`"original" is structurally equal to "similar"`)
	}
	if !jsonpatch.Equal(original, different) {
		fmt.Println(`"original" is _not_ structurally equal to "different"`)
	}
}

func TestCombineMergePatches(t *testing.T) {
	original := []byte(`{"name": "John", "age": 24, "height": 3.21}`)
	nameAndHeight := []byte(`{"height":null,"name":"Jane"}`)
	ageAndEyes := []byte(`{"age":23,"eyes":"blue"}`)

	combinedPatch, err := jsonpatch.MergeMergePatches(nameAndHeight, ageAndEyes)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("combined patch: %s\n", combinedPatch)

	withCombinedPatch, err := jsonpatch.MergePatch(original, combinedPatch)
	if err != nil {
		t.Fatal(err)
	}

	withoutCombinedPatch, err := jsonpatch.MergePatch(original, nameAndHeight)
	if err != nil {
		t.Fatal(err)
	}
	withoutCombinedPatch, err = jsonpatch.MergePatch(withoutCombinedPatch, ageAndEyes)
	if err != nil {
		t.Fatal(err)
	}

	if jsonpatch.Equal(withoutCombinedPatch, withCombinedPatch) {
		fmt.Println("Both JSON documents are structurally the same")
	}
	fmt.Printf("combined merge patch:\n%s\n", combinedPatch)
}
