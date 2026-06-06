package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"zjin.goapp.demo/utils"
)

func TestSliceToMap(t *testing.T) {
	type Student struct {
		ID   int
		Name string
	}

	students := []Student{
		{101, "Foo"},
		{103, "Bar"},
		{112, "Jon"},
		{106, "James"},
	}
	results := utils.SliceToMap(students, func(s Student) string {
		return s.Name
	})
	t.Logf("map results: %+v", results)
}

func TestSliceDiff(t *testing.T) {
	type args struct {
		s1 []string
		s2 []string
	}
	tests := []struct {
		name string
		args args
		want utils.SliceDiffs[string]
	}{
		{
			name: "slice diff case1",
			args: args{
				s1: []string{"a", "b", "c"},
				s2: []string{"b", "c", "d"},
			},
			want: utils.SliceDiffs[string]{
				Added:   []string{"d"},
				Removed: []string{"a"},
				Matched: []string{"b", "c"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.SliceDiff(tt.args.s1, tt.args.s2)
			t.Log("diff results:", got.String())
			assert.True(t, tt.want.Equal(got))
		})
	}
}
