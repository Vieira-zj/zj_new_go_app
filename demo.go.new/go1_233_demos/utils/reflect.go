package utils

import (
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"strings"
)

func SetStructField(obj any, name string, value any) error {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Pointer {
		return fmt.Errorf("input object must be a pointer")
	}

	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("input object must point to a struct")
	}

	field := v.FieldByName(name)
	if !field.IsValid() { // 检查 字段是否存在
		return fmt.Errorf("field %s is not found", name)
	}
	if !field.CanSet() { // 检查 字段是否可设置
		return fmt.Errorf("field %s cannot be set", name)
	}

	fieldValue := reflect.ValueOf(value)
	if field.Type() != fieldValue.Type() { // 检查 类型是否匹配
		return fmt.Errorf("type mismatch: field_type=%v, value_type=%v", field.Type(), fieldValue.Type())
	}

	field.Set(fieldValue)
	return nil
}

const (
	added = iota + 1
	removed
	matched
)

type StructDiffItem struct {
	FieldName string
	OldValue  any
	NewValue  any
}

func (r StructDiffItem) String() string {
	return fmt.Sprintf("Field: %s, Old: %v, New: %v", r.FieldName, r.OldValue, r.NewValue)
}

// DiffStruct compares two flat structs, and returns a slice of diff results.
func DiffStruct(src, dst any) []StructDiffItem {
	srcMapping := GetStructFieldValueMapping(src)
	dstMapping := GetStructFieldValueMapping(dst)

	diffs := make(map[string]int, len(srcMapping))
	for fieldName := range dstMapping {
		diffs[fieldName] = added
	}
	for fieldName := range srcMapping {
		if _, ok := diffs[fieldName]; ok {
			diffs[fieldName] = matched
		} else {
			diffs[fieldName] = removed
		}
	}

	results := make([]StructDiffItem, 0, len(diffs))

	for fieldName, diff := range diffs {
		switch diff {
		case matched:
			srcValue, dstValue := srcMapping[fieldName], dstMapping[fieldName]
			if !reflect.DeepEqual(srcValue, dstValue) {
				results = append(results, StructDiffItem{
					FieldName: fieldName,
					OldValue:  srcValue,
					NewValue:  dstValue,
				})
			}
		case added:
			results = append(results, StructDiffItem{
				FieldName: fieldName,
				OldValue:  nil,
				NewValue:  dstMapping[fieldName],
			})
		case removed:
			results = append(results, StructDiffItem{
				FieldName: fieldName,
				OldValue:  srcMapping[fieldName],
				NewValue:  nil,
			})
		}
	}
	return results
}

func GetStructFieldValueMapping(s any) map[string]any {
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Pointer && !v.IsNil() {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}

	results := make(map[string]any, v.NumField())

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field, value := t.Field(i), v.Field(i)
		results[field.Name] = value.Interface()
	}
	return results
}

type StructFieldInfo struct {
	Name  string
	Index int
	Value any
}

func GetStructFieldsInfo(s any) ([]StructFieldInfo, error) {
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("input is not a struct")
	}

	results := make([]StructFieldInfo, 0, v.NumField())

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field, value := t.Field(i), v.Field(i).Interface()
		if tag := field.Tag.Get("x_idx"); len(tag) > 0 {
			idx, err := strconv.Atoi(tag)
			if err != nil {
				return nil, fmt.Errorf("invalid x_idx tag value: %s", tag)
			}

			results = append(results, StructFieldInfo{
				Name:  field.Name,
				Index: idx,
				Value: value,
			})
		}
	}
	return results, nil
}

type FuncSignature struct {
	Pkg         string
	Name        string
	InputTypes  []string
	OutputTypes []string
}

func GetFuncSignature(fn any) (FuncSignature, error) {
	signature := FuncSignature{}

	fnValue := reflect.ValueOf(fn)
	fnType := fnValue.Type()

	if fnValue.Kind() != reflect.Func {
		return signature, fmt.Errorf("input is not a function")
	}

	fullName := runtime.FuncForPC(fnValue.Pointer()).Name()
	idx := strings.LastIndex(fullName, ".")
	signature.Pkg, signature.Name = fullName[:idx], fullName[idx+1:]

	paramTypes := make([]string, fnType.NumIn())
	for i := 0; i < fnType.NumIn(); i++ {
		paramTypes[i] = fnType.In(i).String()
	}
	signature.InputTypes = paramTypes

	resultTypes := make([]string, fnType.NumOut())
	for i := 0; i < fnType.NumOut(); i++ {
		resultTypes[i] = fnType.Out(i).String()
	}
	signature.OutputTypes = resultTypes

	// cannot get the function parameter names by reflect
	// need to use go/parser and go/ast to parse the source code file

	return signature, nil
}

func GetCallerFuncName() (pkg, fn string) {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		return
	}

	fullName := runtime.FuncForPC(pc).Name()
	idx := strings.LastIndex(fullName, ".")
	pkg, fn = fullName[:idx], fullName[idx+1:]
	return
}
