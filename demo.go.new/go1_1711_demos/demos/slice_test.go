package demos

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Demo: slice 初始值判断

type TestSliceHolder struct {
	data []string
}

func TestSliceCheck(t *testing.T) {
	holder := TestSliceHolder{}
	if holder.data == nil {
		t.Log("slice is <nil>")
	}
	if len(holder.data) == 0 {
		t.Log("slice len:", len(holder.data))
	}

	holder.data = make([]string, 3)
	t.Log("slice len:", len(holder.data))

	s := holder.data[:0]
	t.Log("reset slice:", len(s), cap(s))
}

// Demo: slice

func TestSliceCopy(t *testing.T) {
	a := []int{9, 8, 7, 6, 5, 4, 3, 2, 1}
	b := a[4:7]

	// 向一个空的 slice 里面拷贝元素什么也不会发生。拷贝时，只有 min(len(a), len(b)) 个元素会被成功拷贝
	// 先创建一个指定长度的 slice, 再执行拷贝
	c := make([]int, 3)
	copy(c, a[4:7])

	sort.Sort(sort.IntSlice(a))
	t.Log("after sort:")
	fmt.Println(a)
	fmt.Println(b)
	fmt.Println(c)
}

func TestInitSliceByRef(t *testing.T) {
	s := make([]*int, 0, 3)
	// all elements have same pointer value and value is the last value of the iterator variable
	// because i is the same variable throughout the loop.
	for i := 0; i < 3; i++ {
		s = append(s, &i)
	}
	t.Log(*s[0], *s[1], *s[2]) // 3, 3, 3

	s = make([]*int, 0, 3)
	for i := 0; i < 3; i++ {
		i := i
		s = append(s, &i)
	}
	t.Log(*s[0], *s[1], *s[2]) // 0, 1, 2
}

func TestSliceAppend(t *testing.T) {
	a := make([]int, 3, 10)
	a[0], a[1], a[2] = 1, 2, 3

	// c == b because both refer to same underlying array and capacity of that is 10
	// so appending to a will not create new array.
	b := append(a, 4)
	c := append(a, 5)
	t.Log("b:", b) // [1 2 3 5]
	t.Log("c:", c) // [1 2 3 5]
	fmt.Println()

	b = make([]int, 3)
	copy(b, a)
	b = append(b, 4)
	t.Log("b:", b)

	c = make([]int, 3)
	copy(c, a)
	c = append(c, 5)
	t.Log("c:", c)
}

// Demo: sort slice of structs

type person struct {
	name string
	age  int
	sex  string
}

type sortByPersonFields []person

func (e sortByPersonFields) Swap(i, j int) { e[i], e[j] = e[j], e[i] }
func (e sortByPersonFields) Len() int      { return len(e) }
func (e sortByPersonFields) Less(i, j int) bool {
	eleX := e[i]
	eleY := e[j]
	if eleX.sex != eleY.sex {
		return eleX.sex < eleY.sex
	}
	if eleX.age != eleY.age {
		return eleX.age < eleY.age
	}
	return eleX.name < eleY.name
}

func TestSortSliceOfStructs01(t *testing.T) {
	// sort by: sex asc, age asc, name asc
	persons := []person{
		{name: "zh", age: 24, sex: "female"},
		{name: "foo", age: 30, sex: "male"},
		{name: "yx", age: 36, sex: "female"},
		{name: "jx", age: 27, sex: "female"},
		{name: "zht", age: 24, sex: "female"},
		{name: "bar", age: 33, sex: "male"},
		{name: "ja", age: 24, sex: "female"},
	}

	sort.Sort(sortByPersonFields(persons))
	for _, p := range persons {
		log.Println("sorted persons:", p.sex, p.age, p.name)
	}
}

func TestSortSliceOfStructs02(t *testing.T) {
	// sort by: sex desc, age asc, name desc
	persons := []person{
		{name: "zh", age: 24, sex: "female"},
		{name: "foo", age: 30, sex: "male"},
		{name: "yx", age: 36, sex: "female"},
		{name: "jx", age: 27, sex: "female"},
		{name: "zht", age: 24, sex: "female"},
		{name: "bar", age: 33, sex: "male"},
		{name: "ja", age: 24, sex: "female"},
	}

	sort.Slice(persons, func(i, j int) bool {
		eleX := persons[i]
		eleY := persons[j]
		if eleX.sex != eleY.sex {
			return eleX.sex > eleY.sex
		}
		if eleX.age != eleY.age {
			return eleX.age < eleY.age
		}
		return eleX.name > eleY.name
	})

	for _, p := range persons {
		t.Log("sorted persons:", p.sex, p.age, p.name)
	}
}

// Demo: group by requests (type dict)

type dict map[string]interface{}

func groupByRequests(results map[string]int, reqs []dict) error {
	for {
		if len(reqs) == 0 {
			return nil
		}
		if len(reqs) == 1 {
			curReq := reqs[0]
			b, err := json.Marshal(curReq)
			if err != nil {
				return err
			}
			results[string(b)] = 1
			return nil
		}

		count := 1
		curReq := reqs[0]
		notMatchedReqs := make([]dict, 0, 16)
		for _, req := range reqs[1:] {
			if isDictEqualByFields(curReq, req) {
				count += 1
			} else {
				notMatchedReqs = append(notMatchedReqs, req)
			}
		}
		b, err := json.Marshal(curReq)
		if err != nil {
			return err
		}
		results[string(b)] = count

		reqs = notMatchedReqs
	}
}

// isDictEqualByFields only compares dict fields of 1st level.
func isDictEqualByFields(src, dst interface{}) bool {
	srcDict, ok := src.(dict)
	if !ok {
		return false
	}
	dstDict, ok := dst.(dict)
	if !ok {
		return false
	}

	if len(srcDict) != len(dstDict) {
		return false
	}

	for k := range srcDict {
		if _, ok := dstDict[k]; !ok {
			return false
		}
	}
	return true
}

func TestGroupByRequests(t *testing.T) {
	reqs := []dict{
		{"uid": 1221654048, "name": "ab"},
		{"uid": 1221654049, "name": "cd"},
		{"sp_uid": 1221724968},
		{"uid": 1221654048, "otp_token": "111111"},
		{"uid": 1221654050, "name": "xy"},
		{"uid": 1221654051, "name": "xz"},
		{"sp_uid": 1221724967},
	}

	results := make(map[string]int, 4)
	err := groupByRequests(results, reqs)
	assert.NoError(t, err)
	for req, count := range results {
		t.Logf("%s:%d", req, count)
	}
	t.Log("groupby done")
}
