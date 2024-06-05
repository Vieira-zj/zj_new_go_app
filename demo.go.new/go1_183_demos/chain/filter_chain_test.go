package chain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunFilter(t *testing.T) {
	persons := []Person{
		{Name: "foo1", Age: 31, Title: "Dev"},
		{Name: "foo1", Age: 35, Title: "QA"},
		{Name: "foo1", Age: 33, Title: "QA"},
		{Name: "foo1", Age: 27, Title: "Dev"},
		{Name: "foo1", Age: 29, Title: "PM"},
		{Name: "foo1", Age: 41, Title: "Dev"},
	}

	filters := []Filter{
		AgeFilter{AgeCond: 30},
		TitleFilter{TitleCond: "Dev"},
	}

	results := RunFilter(persons, filters)
	b, err := json.Marshal(results)
	assert.NoError(t, err)
	t.Log("filter results:", string(b))
}
