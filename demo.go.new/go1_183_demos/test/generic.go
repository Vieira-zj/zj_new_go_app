package test

import (
	"fmt"
	"reflect"
	"strconv"
)

// AddByReflect add fn supports for int32 and float32 by reflect.
func AddByReflect(a, b interface{}) (interface{}, error) {
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

// AddByGeneric add fn supports for int32 and float32 by generic.
func AddByGeneric[T int32 | float32](a, b T) T {
	return a + b
}

func InsertAt[T any](list []T, idx int, t T) ([]T, error) {
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

func RemoveAt[T any](list []T, idx int, t T) ([]T, error) {
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

func GetFieldInfo[T any /* int | string */](field T) string {
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
