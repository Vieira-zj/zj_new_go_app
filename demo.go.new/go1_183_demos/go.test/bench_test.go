package gotest

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"
)

type MyTestPerson struct {
	Name string
	Id   int
	Addr string
}

// go test -benchmem -v -bench=BenchmarkJsonStrParse -run=^$ -benchtime=8s demo.apps/go.test
// 1206 ns/op   384 B/op   11 allocs/op
func BenchmarkJsonStrParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b := []byte(fmt.Sprintf(`{"name":"foo", "id":%d, "addr":"wuhan"}`, i))
		p := MyTestPerson{}
		json.Unmarshal(b, &p)
	}
}

// go test -benchmem -v -bench=BenchmarkTextStrParse -run=^$ -benchtime=8s demo.apps/go.test
// 238 ns/op   88 B/op   3 allocs/op
func BenchmarkTextStrParse(b *testing.B) {
	parse := func(s string, p *MyTestPerson) {
		items := strings.Split(s, ",")
		name := items[0][strings.Index(items[0], ":")+1:]
		id := items[1][strings.Index(items[1], ":")+1:]
		addr := items[2][strings.Index(items[2], ":")+1:]
		idNumer, _ := strconv.Atoi(id)

		p.Name = name
		p.Id = idNumer
		p.Addr = addr
	}

	for i := 0; i < b.N; i++ {
		s := fmt.Sprintf(`name:foo,id:%d,addr:wuhan`, i)
		p := MyTestPerson{}
		parse(s, &p)
	}
}
