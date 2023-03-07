package gotest

import (
	"encoding/json"
	"go1_1711_demo/utils"
	"math/rand"
	"sync"
	"testing"
)

// benchmark: 指定容器容量

func BenchmarkInitSlice(b *testing.B) {
	var nums []int
	for n := 0; n < b.N; n++ {
		nums = append(nums, rand.Intn(10000))
	}
}

func BenchmarkInitSliceWithCap(b *testing.B) {
	nums := make([]int, 0, b.N)
	for n := 0; n < b.N; n++ {
		nums = append(nums, rand.Intn(10000))
	}
}

// benchmark: 利用 unsafe 包避开内存 copy

func BenchmarkConvStr2bytes(b *testing.B) {
	s := "testString"
	var bs []byte
	for n := 0; n < b.N; n++ {
		bs = []byte(s)
	}
	_ = bs
}

func BenchmarkConvStr2bytesByUnsafe(b *testing.B) {
	s := "testString"
	var bs []byte
	for i := 0; i < b.N; i++ {
		bs = utils.Str2bytes(s)
	}
	_ = bs
}

func BenchmarkConvBytes2str(b *testing.B) {
	bs := utils.Str2bytes("testString")
	var s string
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		s = string(bs)
	}
	_ = s
}

func BenchmarkConvBytes2strByUnsafe(b *testing.B) {
	bs := utils.Str2bytes("testString")
	var s string
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		s = utils.Bytes2str(bs)
	}
	_ = s
}

// benchmark: sync.Pool

type RealTimeRuleStruct struct {
	Filter []*struct {
		PropertyId   int64    `json:"property_id"`
		PropertyCode string   `json:"property_code"`
		Operator     string   `json:"operator"`
		Value        []string `json:"value"`
	} `json:"filter"`
	ExtData [1024]byte `json:"ext_data"`
}

var realTimeRuleBytes = []byte(`{"filter":[{"property_id":2,"property_code":"search_poiid_industry","operator":"in","value":["yimei"]},{"property_id":4,"property_code":"request_page_id","operator":"in","value":["all"]}],"white_list":[{"property_id":1,"property_code":"white_list_for_adiu","operator":"in","value":["j838ef77bf227chcl89888f3fb0946","lb89bea9af558589i55559764bc83e"]}],"ipc_user_tag":[{"property_id":1,"property_code":"ipc_crowd_tag","operator":"in","value":["test_20227041152_mix_ipc_tag"]}],"relation_id":0,"is_copy":true}`)

func BenchmarkUnmarshal(b *testing.B) {
	// 每次都会生成一个新的临时对象
	for n := 0; n < b.N; n++ {
		s := RealTimeRuleStruct{}
		json.Unmarshal(realTimeRuleBytes, &s)
	}
}

var realTimeRulePool = sync.Pool{
	New: func() interface{} {
		return new(RealTimeRuleStruct)
	},
}

func BenchmarkUnmarshalWithPool(b *testing.B) {
	// 复用一个对象，不用每次都生成新的
	for n := 0; n < b.N; n++ {
		s := realTimeRulePool.Get().(*RealTimeRuleStruct)
		json.Unmarshal(realTimeRuleBytes, s)
		realTimeRulePool.Put(s)
	}
}
