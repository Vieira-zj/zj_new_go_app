package utils

import (
	"errors"
	"fmt"
	"reflect"
	"unsafe"
)

//
// Struct to Map
//

// ReflectStructToMapV1 returns struct fields type and value.
func ReflectStructToMapV1(object interface{}) (map[string]interface{}, error) {
	valueOf := reflect.ValueOf(object)
	if valueOf.Kind() == reflect.Ptr {
		valueOf = valueOf.Elem()
	}
	if valueOf.Kind() != reflect.Struct {
		return nil, fmt.Errorf("not support non struct type: [%s]", valueOf.Kind().String())
	}

	numField := valueOf.NumField()
	retMap := make(map[string]interface{}, numField)
	for i := 0; i < numField; i++ {
		fValue := valueOf.Field(i)
		fTypeName := fValue.Type().Name()
		if fValue.Kind() == reflect.Ptr {
			fValue = fValue.Elem()
		}
		switch fValue.Kind() {
		case reflect.Struct:
			res, err := ReflectStructToMapV1(fValue.Interface())
			if err != nil {
				return nil, err
			}
			retMap[fTypeName] = res
		case reflect.Slice:
			res, err := ReflectSliceToString(fValue.Interface())
			if err != nil {
				return nil, err
			}
			retMap[fTypeName] = res
		default:
			retMap[fTypeName] = fValue.Interface()
		}
	}
	return retMap, nil
}

// ReflectStructToMapV2 returns struct fields tag name and value.
func ReflectStructToMapV2(object interface{}, tag string) (map[string]interface{}, error) {
	valueOf := reflect.ValueOf(object)
	if valueOf.Kind() == reflect.Ptr {
		valueOf = valueOf.Elem()
	}
	if valueOf.Kind() != reflect.Struct {
		return nil, fmt.Errorf("ToMap only accepts struct or struct pointer; got %s", valueOf.Kind().String())
	}

	queue := make([]interface{}, 0, 4)
	queue = append(queue, object)

	retMap := make(map[string]interface{}, valueOf.NumField())
	for len(queue) > 0 {
		valueOf := reflect.ValueOf(queue[0])
		if valueOf.Kind() == reflect.Ptr {
			valueOf = valueOf.Elem()
		}
		queue = queue[1:]
		for i := 0; i < valueOf.NumField(); i++ {
			fValue := valueOf.Field(i)
			fType := valueOf.Type().Field(i)
			if fValue.Kind() == reflect.Ptr { // 内嵌指针
				fValue = fValue.Elem()
			}
			if fValue.Kind() == reflect.Struct { // 内嵌结构体
				queue = append(queue, fValue.Interface())
			} else {
				if tagValue := fType.Tag.Get(tag); tagValue != "" {
					retMap[tagValue] = fValue.Interface()
				}
			}
		}
	}
	return retMap, nil
}

//
// Map to Struct
//

// ReflectMapToStruct init a struct by map key and value.
func ReflectMapToStruct(data map[string]interface{}, inStructPtr interface{}) error {
	inValueOf := reflect.ValueOf(inStructPtr)
	if inValueOf.Kind() != reflect.Ptr {
		return errors.New("inStructPtr must be ptr to struct")
	}

	inValueOf = inValueOf.Elem()
	if inValueOf.Kind() != reflect.Struct {
		return fmt.Errorf("not support non struct type: [%s]", inValueOf.Kind().String())
	}

	for i := 0; i < inValueOf.NumField(); i++ {
		fValue := inValueOf.Field(i)
		fType := inValueOf.Type().Field(i)

		key := fType.Tag.Get("key")
		dataValue, ok := data[key]
		if !ok {
			return fmt.Errorf(fType.Name + " not found")
		}

		dataTypeOf := reflect.TypeOf(dataValue)
		if dataTypeOf.Kind() == fValue.Kind() {
			fValue.Set(reflect.ValueOf(dataValue))
		} else if dataTypeOf.ConvertibleTo(fValue.Type()) {
			fValue.Set(reflect.ValueOf(dataValue).Convert(fValue.Type()))
		} else {
			return fmt.Errorf("type mismatch: fieldType=%s, valueType=%s", fValue.Kind(), dataTypeOf.Kind())
		}
	}
	return nil
}

//
// Others
//

// ReflectSliceToString .
func ReflectSliceToString(object interface{}) (string, error) {
	TypeOf := reflect.TypeOf(object)
	if TypeOf.Kind() == reflect.Ptr {
		TypeOf = TypeOf.Elem()
	}
	if TypeOf.Kind() != reflect.Slice {
		return "", fmt.Errorf("not support non slice type: [%s]", TypeOf.Kind().String())
	}

	valueOf := reflect.ValueOf(object)
	resItems := make([]interface{}, 0, valueOf.Len())
	for i := 0; i < valueOf.Len(); i++ {
		item := valueOf.Index(i)
		if item.Kind() == reflect.Ptr {
			item = item.Elem()
		}
		switch item.Kind() {
		case reflect.Struct:
			res, err := ReflectStructToMapV1(item.Interface())
			if err != nil {
				return "", err
			}
			resItems = append(resItems, res)
		case reflect.Slice:
			res, err := ReflectSliceToString(item.Interface())
			if err != nil {
				return "", err
			}
			resItems = append(resItems, res)
		default:
			resItems = append(resItems, item.Interface())
		}
	}
	return fmt.Sprintf("%s:%v", TypeOf.Kind().String(), resItems), nil
}

// unsafe.Pointer
// string 是 immutable, 而 []byte 是 mutable 的，所以这么转换的时候会拷贝一次；如果要不拷贝转换的话，就需要用到 unsafe 了。
// 可以不拷贝数据而将 string 转换成 []byte, 不过要注意这样生成的 []byte 不可写，否则行为未定义。

// StringToSliceByte .
func StringToSliceByte(s string) []byte {
	l := len(s)
	data := (*reflect.StringHeader)(unsafe.Pointer(&s)).Data
	return *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: data,
		Len:  l,
		Cap:  l,
	}))
}
