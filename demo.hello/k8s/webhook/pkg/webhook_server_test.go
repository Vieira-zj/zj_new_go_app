package pkg

import (
	"encoding/json"
	"fmt"
	"testing"

	jsonpatch "github.com/evanphx/json-patch"
	"gopkg.in/yaml.v3"
)

/*
Load yaml configs
*/

func TestLoadConfig01(t *testing.T) {
	data := `
initContainers:
- name: init-busybox
  image: busybox:1.30
  imagePullPolicy: IfNotPresent
  command: ["sh", "-c", "sleep 3; echo \"init process done.\""]
volumes:
- name: busybox-conf
  configMap:
  name: busybox-configmap
`

	// Note: fields of struct Config should be public.
	var cfg Config
	if err := yaml.Unmarshal([]byte(data), &cfg); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("config: %+v", cfg)
}

func TestLoadConfig02(t *testing.T) {
	data := `
initContainers:
- name: init-busybox
  image: busybox:1.30
  imagePullPolicy: IfNotPresent
  command: ["sh", "-c", "sleep 3; echo \"init process done.\""]
# volumes:
# - name: busybox-conf
#   configMap:
#     name: busybox-configmap
`

	var cfg Config
	if err := yaml.Unmarshal([]byte(data), &cfg); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("config: %+v", cfg)
}

/*
Json Patch

github: https://github.com/evanphx/json-patch
jsonpatch syntax: https://tools.ietf.org/html/rfc6902#appendix-A.1
*/

func TestJsonPatch01(t *testing.T) {
	original := []byte(`{"name": "John", "age": 24, "height": 3.21}`)
	// jsonpatch op: replace and remove
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
	fmt.Println("Original document:", string(original))
	fmt.Println("MOdified document:", string(modified))
}

func TestJsonPatch02(t *testing.T) {
	original := []byte(`{"name": "John", "age": 24, "height": 3.21}`)

	// use json patch op struct
	patchReplace := patchOperation{
		Op:    "replace",
		Path:  "/name",
		Value: "Henry",
	}
	patchRemove := patchOperation{
		Op:   "remove",
		Path: "/height",
	}
	patches := []patchOperation{patchReplace, patchRemove}

	patchJSON, err := json.Marshal(patches)
	if err != nil {
		t.Fatal(err)
	}

	patch, err := jsonpatch.DecodePatch(patchJSON)
	if err != nil {
		t.Fatal(err)
	}

	modified, err := patch.Apply(original)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Original document:", string(original))
	fmt.Println("MOdified document:", string(modified))
}

func TestJsonPatch03(t *testing.T) {
	original := []byte(`{"name": "John", "age": 24, "skills": ["Java", "C#"]}`)
	// jsonpatch op: add a item to list
	patchJSON := []byte(`[
		{"op": "replace", "path": "/name", "value": "Henry"},
		{"op": "add", "path": "/skills/-", "value": "Golang"}
	]`)

	patch, err := jsonpatch.DecodePatch(patchJSON)
	if err != nil {
		t.Fatal(err)
	}

	modified, err := patch.Apply(original)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Original document:", string(original))
	fmt.Println("Modified document:", string(modified))
}
