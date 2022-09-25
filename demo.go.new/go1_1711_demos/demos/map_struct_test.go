package demos

import (
	"testing"

	"github.com/imdario/mergo"
)

//
// 1. mergo 不会复制非导出字段
// 2. map 使用时候，对应的key字段默认是小写的
// 3. mergo 可以嵌套赋值
//

type Student struct {
	Name string
	Num  int
	Age  int
}

func TestMergoStructToMap(t *testing.T) {
	student := Student{
		Name: "foo",
		Num:  1,
		Age:  18,
	}

	m := make(map[string]interface{})
	if err := mergo.Map(&m, student); err != nil {
		t.Fatal(err)
	}
	t.Logf("map m: %+v", m)
}

func TestMergoMapToStruct(t *testing.T) {
	m := map[string]interface{}{
		"name": "bar",
		"num":  2,
		"age":  20,
	}

	s := Student{}
	if err := mergo.Map(&s, m); err != nil {
		t.Fatal(err)
	}
	t.Logf("struct student: %+v", s)
}
