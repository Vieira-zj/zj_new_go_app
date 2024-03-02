package demos

import (
	"fmt"
	"io"
	"net"
	"os"
	"reflect"
	"strconv"
)

// Refer: https://github.com/smallnest/talk-about-go-generics

// addByReflect an add fn supports for int32 and float32 by reflect.
func addByReflect(a, b interface{}) (interface{}, error) {
	// here use generic [T any] instead of interface{}
	aType := reflect.TypeOf(a).Name()
	bType := reflect.TypeOf(b).Name()
	fmt.Printf("param type is: %s, %s\n", aType, bType)
	if aType != bType {
		return -1, fmt.Errorf("param a and b type are mismatch")
	}

	aValueOf := reflect.ValueOf(a)
	bValueOf := reflect.ValueOf(b)
	switch aValueOf.Kind() {
	case reflect.Int32:
		return aValueOf.Int() + bValueOf.Int(), nil
	case reflect.Float32:
		return aValueOf.Float() + bValueOf.Float(), nil
	default:
		return -1, fmt.Errorf("param must be int32 or float32")
	}
}

// 泛型方法

// add adds fn supports for int32 and float32 by generic.
func add[T int32 | float32](a, b T) T {
	return a + b
}

// The special tilde ("~") symbol instructs Go that T should approximately match one of these types rather than precisely.
// That is, it should match one of these types or a derived-type extending these.
// Without the tilde this function will receive only the exact types declared in the constraints and no derived-types.
func min[T ~int | ~float64](x, y T) T {
	if x < y {
		return x
	}
	return y
}

func insertAt[T any](list []T, idx int, t T) ([]T, error) {
	l := len(list)
	if idx < 0 {
		idx = l + idx
	}

	if idx >= l || idx < 0 {
		return nil, fmt.Errorf("Out of index %d", idx)
	}

	retList := make([]T, 0, l+1)
	retList = append(retList, list[:idx]...)
	retList = append(retList, t)
	if idx < l+1 {
		retList = append(retList, list[idx:]...)
	}
	return retList, nil
}

//nolint:unused
func removeAt[T any](list []T, idx int, _ T) ([]T, error) {
	l := len(list)
	if idx < 0 {
		idx = l + idx
	}

	if idx >= l || idx < 0 {
		return nil, fmt.Errorf("Out of index %d", idx)
	}

	retList := make([]T, 0, l-1)
	retList = append(retList, list[:idx]...)
	if idx < l {
		retList = append(retList, list[idx+1:]...)
	}
	return retList, nil
}

func getFieldInfo[T any /* int | string */](field T) string {
	typeOf := reflect.TypeOf(field)
	valueOf := reflect.ValueOf(field)
	if typeOf.Kind() == reflect.Ptr {
		valueOf = valueOf.Elem()
	}

	var fType, fValue string
	switch valueOf.Kind() {
	case reflect.Int:
		fType = "int"
		fValue = strconv.FormatInt(valueOf.Int(), 10)
	case reflect.String:
		fType = "string"
		fValue = valueOf.String()
	default:
		fType = "any"
		fValue = fmt.Sprintf("%v", valueOf.Interface())
	}
	return fmt.Sprintf("type=%s | value=%s", fType, fValue)
}

// 类型约束

type myNumber interface {
	int | int32 | int64 | float32 | float64
}

//nolint:unused
type myReadWriterCloser interface {
	*os.File | *net.TCPConn | *net.UDPConn | *net.UnixConn | *net.IPConn

	// 嵌入接口
	io.Reader
	io.Writer
	io.Closer
}

// comparable means that the value of T supports comparison using the equality operator ("=").
type MyKvMap[K comparable, V myNumber] map[K]V

func (kv MyKvMap[K, V]) Set(k K, v V) MyKvMap[K, V] {
	if _, ok := kv[k]; !ok {
		kv[k] = v
	}
	return kv
}

func (kv MyKvMap[K, V]) Pprint() {
	for k, v := range kv {
		fmt.Printf("key=%v,value=%v\n", k, v)
	}
}
