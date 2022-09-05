package demos

import (
	"reflect"
	"testing"
)

/*
减少使用 FieldByName 方法

在需要使用反射进行成员变量访问的时候，尽可能的使用成员的序号。如果只知道成员变量的名称的时候，看具体代码的使用场景，
如果可以在启动阶段或在频繁访问前，通过 TypeOf(), Type.FieldByName() 和 StructField.Index 得到成员的序号。
注意：这里需要使用 reflect.Type 而不是 reflect.Value, 通过 reflect.Value 是得不到字段名称的。
*/

// Benchmark: New

func BenchmarkNew(b *testing.B) {
	var uInfo *UserInfo
	for i := 0; i < b.N; i++ {
		uInfo = new(UserInfo)
	}
	_ = uInfo
	// 30.12 ns/op
}

func BenchmarkReflectNew(b *testing.B) {
	var uInfo *UserInfo
	typeOf := reflect.TypeOf(UserInfo{})
	for i := 0; i < b.N; i++ {
		uInfo = reflect.New(typeOf).Interface().(*UserInfo)
	}
	_ = uInfo
	// 44.19 ns/op
}

// Benchmark: GetField

func BenchmarkGetField(b *testing.B) {
	var r int
	var uInfo = new(UserInfo)
	uInfo.Age = 1995
	for i := 0; i < b.N; i++ {
		r = uInfo.Age
	}
	_ = r
	// 0.2589 ns/op
}

func BenchmarkReflectGetFieldByIndex(b *testing.B) {
	var r int64
	var uInfo = new(UserInfo)
	valueOf := reflect.ValueOf(uInfo).Elem()
	for i := 0; i < b.N; i++ {
		r = valueOf.Field(1).Int()
	}
	_ = r
	// 3.067 ns/op
}

func BenchmarkReflectGetFieldByName(b *testing.B) {
	var r int64
	var uInfo = new(UserInfo)
	valueOf := reflect.ValueOf(uInfo).Elem()
	for i := 0; i < b.N; i++ {
		r = valueOf.FieldByName("Age").Int()
	}
	_ = r
	// 59.80 ns/op
}

// Benchmark: SetField

func BenchmarkSetField(b *testing.B) {
	var uInfo = new(UserInfo)
	for i := 0; i < b.N; i++ {
		uInfo.Age = i
	}
	_ = uInfo
	// 0.2315 ns/op
}

func BenchmarkReflectSetFieldByIndex(b *testing.B) {
	var uInfo = new(UserInfo)
	valueOf := reflect.ValueOf(uInfo).Elem()
	for i := 0; i < b.N; i++ {
		valueOf.Field(1).SetInt(int64(25))
	}
	// 5.024 ns/op
}

func BenchmarkReflectSetFieldByName(b *testing.B) {
	var uInfo = new(UserInfo)
	valueOf := reflect.ValueOf(uInfo).Elem()
	for i := 0; i < b.N; i++ {
		valueOf.FieldByName("Age").SetInt(int64(25))
	}
	// 60.97 ns/op
}

// Benchmark: CallMethod

func BenchmarkCallMethod(b *testing.B) {
	uInfo := &UserInfo{}
	for i := 0; i < b.N; i++ {
		uInfo.PrettyString()
	}
	// 219.7 ns/op
}

func BenchmarkReflectCallMethodByIndex(b *testing.B) {
	uInfo := reflect.ValueOf(&UserInfo{})
	for i := 0; i < b.N; i++ {
		uInfo.Method(0).Call(nil)
	}
	// 501.1 ns/op
}

func BenchmarkReflectCallMethodByName(b *testing.B) {
	uInfo := reflect.ValueOf(&UserInfo{})
	for i := 0; i < b.N; i++ {
		uInfo.MethodByName("PrettyString").Call(nil)
	}
	// 814.7 ns/op
}
