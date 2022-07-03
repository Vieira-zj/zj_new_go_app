package test

import (
	"fmt"
	"reflect"
	"strconv"
)

// InsertAt .
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

// RemoveAt .
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

// GetFieldInfo .
func GetFieldInfo[T any /* int | string */](field T) string {
	typeOf := reflect.TypeOf(field)
	valueOf := reflect.ValueOf(field)
	if typeOf.Kind() == reflect.Ptr {
		valueOf = valueOf.Elem()
	}

	var (
		fType  = ""
		fValue = ""
	)
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
