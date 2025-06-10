package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"
)

func IsNil(x any) bool {
	if x == nil {
		return true
	}
	return reflect.ValueOf(x).IsNil()
}

func IsEmptyStruct(x any) bool {
	valOf := reflect.ValueOf(x)
	if valOf.Kind() == reflect.Ptr {
		valOf = valOf.Elem()
	}
	return valOf.IsZero()
}

func TrimStringFields(obj any) error {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Pointer {
		return errors.New("input should be pointer")
	}

	v = v.Elem()
	for i := 0; i < v.NumField(); i++ {
		fv := v.Field(i)
		// Note: cannot set value for private field, or get error:
		// reflect.Value.SetString using value obtained using unexported field.
		// Fix: add CanSet() in if cond
		if fv.Kind() == reflect.String && fv.CanSet() {
			trim := strings.TrimSpace(fv.String())
			fv.SetString(trim)
		}
	}

	return nil
}

func SlicesContains[T ~int | string | bool](s1, s2 []T) bool {
	base := s1
	dest := s2
	if len(s1) > len(s2) {
		base = s2
		dest = s1
	}

	for i := range base {
		ele1 := base[i]
		idx := 0
		isFound := false
		for j := range dest {
			ele2 := dest[j]
			if isObjectEqual(ele1, ele2) {
				idx = j
				isFound = true
				break
			}
		}
		if !isFound {
			return false
		}
		dest = append(dest[:idx], dest[idx+1:]...)
	}

	if len(dest) > 0 {
		log.Printf("remained elements: %+v", dest)
	}
	return true
}

func isObjectEqual(src, dest any) bool {
	return reflect.TypeOf(src).Kind() == reflect.TypeOf(dest).Kind() && reflect.DeepEqual(src, dest)
}

type FnDeclaration struct {
	Name   string   `json:"name"`
	Input  []string `json:"input"`
	Output []string `json:"output"`
}

func GetFnDeclaration(fn any) (FnDeclaration, error) {
	fnType := reflect.TypeOf(fn)
	if fnType.Kind() != reflect.Func {
		return FnDeclaration{}, fmt.Errorf("input must be function, but got: %s", fnType.Kind().String())
	}

	input := make([]string, 0)
	for i := 0; i < fnType.NumIn(); i++ {
		arg := fnType.In(i)
		input = append(input, arg.String())
	}

	output := make([]string, 0)
	for i := 0; i < fnType.NumOut(); i++ {
		result := fnType.Out(i)
		output = append(output, result.String())
	}

	fullName := GetFnFullName(fn)
	shortName := fullName[strings.LastIndex(fullName, ".")+1:]

	return FnDeclaration{
		Name:   shortName,
		Input:  input,
		Output: output,
	}, nil
}

// Deep Copy by Reflect

func ReflectDeepCopy(in any) any {
	return reflectDeepCopy(reflect.ValueOf(in)).Interface()
}

func reflectDeepCopy(src reflect.Value) reflect.Value {
	switch src.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map:
		if src.IsNil() {
			return src
		}
	case reflect.Func, reflect.Chan, reflect.UnsafePointer, reflect.Uintptr:
		panic(fmt.Sprintf("cannot deep copy kind: %s", src.Kind()))
	}

	switch src.Kind() {
	case reflect.Interface:
		return reflectDeepCopy(src.Elem())
	case reflect.Ptr:
		dst := reflect.New(src.Type().Elem())
		dst.Elem().Set(reflectDeepCopy(src.Elem()))
		return dst
	case reflect.Array:
		dst := reflect.New(src.Type())
		for i := 0; i < src.Len(); i++ {
			dst.Elem().Index(i).Set(reflectDeepCopy(src.Index(i)))
		}
		return dst.Elem()
	case reflect.Slice:
		dst := reflect.MakeSlice(src.Type(), 0, src.Len())
		for i := 0; i < src.Len(); i++ {
			reflect.Append(dst, reflectDeepCopy(src.Index(i)))
		}
		return dst
	case reflect.Map:
		dst := reflect.MakeMapWithSize(src.Type(), src.Len())
		for _, key := range src.MapKeys() {
			dst.SetMapIndex(key, reflectDeepCopy(src.MapIndex(key)))
		}
		return dst
	case reflect.Struct:
		dst := reflect.New(src.Type())
		for i := 0; i < src.NumField(); i++ {
			if !dst.Elem().Field(i).CanSet() { // can't set private fields
				return dst
			}
			dst.Elem().Field(i).Set(reflectDeepCopy(src.Field(i)))
		}
		return dst.Elem()
	default:
		// basic value types like numbers, booleans, and strings
		return src
	}
}

// Wrap Func by Reflect

type ReflectFunc func(args []reflect.Value) (results []reflect.Value)

func LogWrappedFunc(fn any) ReflectFunc {
	return func(args []reflect.Value) (results []reflect.Value) {
		decl, err := GetFnDeclaration(fn)
		if err != nil {
			return []reflect.Value{reflect.ValueOf(err)}
		}
		b, err := json.Marshal(decl)
		if err != nil {
			return []reflect.Value{reflect.ValueOf(err)}
		}

		start := time.Now()
		log.Println("exec func:", string(b))
		results = reflect.ValueOf(fn).Call(args)
		log.Printf("exec duration: %d millisec", time.Since(start).Milliseconds())

		return results
	}
}

func ReflectWrapFunc(fn any) (any, error) {
	typ := reflect.TypeOf(fn)
	if typ.Kind() != reflect.Func {
		return nil, errors.New("input arg is not a func")
	}

	wrapFunc := LogWrappedFunc(fn)
	return reflect.MakeFunc(typ, wrapFunc).Interface(), nil
}
