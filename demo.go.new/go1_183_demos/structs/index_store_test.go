package structs

import (
	"strconv"
	"testing"
)

func TestIndexStore(t *testing.T) {
	resources := make([]Resource, 0, 10)
	for i := 0; i < 10; i++ {
		resources = append(resources, Resource{
			ID:   i,
			Name: "resource_" + strconv.Itoa(i),
			Meta: ResourceMeta{
				Group: getResourceGroup(t, i),
				Type:  getResourceType(t, i),
			},
		})
	}

	indexStore := NewIndexStore()

	t.Run("add", func(t *testing.T) {
		for _, r := range resources {
			if err := indexStore.Add(r); err != nil {
				t.Fatal(err)
			}
		}
		indexStore.index.PrettyPrint()
	})

	t.Run("get", func(t *testing.T) {
		_, ok := indexStore.GetByIndex(KeyIndexByType, "type6")
		if ok {
			t.Fatal("get by index 'type6' should be not ok")
		}

		t.Log("get by index 'type3'")
		results, _ := indexStore.GetByIndex(KeyIndexByType, "type3")
		for _, r := range results {
			t.Log(r.String())
		}

		t.Log("get by index 'group4'")
		results, _ = indexStore.GetByIndex(KeyIndexByGroup, "group4")
		for _, r := range results {
			t.Log(r.String())
		}
	})

	t.Run("update", func(t *testing.T) {
		resources[9].ID = 19
		resources[9].Meta.Type = getResourceType(t, 19)
		if err := indexStore.Add(resources[9]); err != nil {
			t.Fatal(err)
		}

		indexStore.index.PrettyPrint()

		t.Log("get by index 'type1'")
		results, _ := indexStore.GetByIndex(KeyIndexByType, "type1")
		for _, r := range results {
			t.Log(r.String())
		}
	})

	t.Run("delete", func(t *testing.T) {
		for _, idx := range []int{8, 9} {
			r := resources[idx]
			indexStore.Delete(r)
		}
		indexStore.index.PrettyPrint()
		t.Log("total store:", len(indexStore.data))
	})
}

func getResourceType(t *testing.T, i int) string {
	switch {
	case i%2 == 0:
		return "type2"
	case i%3 == 0:
		return "type3"
	default:
		return "type1"
	}
}

func getResourceGroup(t *testing.T, i int) string {
	switch {
	case i%3 == 0:
		return "group3"
	case i%4 == 0:
		return "group4"
	default:
		return "group1"
	}
}
