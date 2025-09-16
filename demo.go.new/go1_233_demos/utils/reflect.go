package utils

import (
	"fmt"
	"reflect"
	"strconv"
)

const (
	added = iota + 1
	removed
	matched
)

type StructDiffResult struct {
	FieldName string
	OldValue  any
	NewValue  any
}

func (r StructDiffResult) String() string {
	return fmt.Sprintf("Field: %s, Old: %v, New: %v", r.FieldName, r.OldValue, r.NewValue)
}

// DiffStruct compares two flat structs, and returns a slice of diff results.
func DiffStruct(src, dst any) []StructDiffResult {
	srcFieldValueMapping := GetStructFieldValueMapping(src)
	dstFieldValueMapping := GetStructFieldValueMapping(dst)

	fieldDiffResults := make(map[string]int, len(srcFieldValueMapping)+len(dstFieldValueMapping))
	for fieldName := range dstFieldValueMapping {
		fieldDiffResults[fieldName] = added
	}
	for fieldName := range srcFieldValueMapping {
		if _, ok := fieldDiffResults[fieldName]; ok {
			fieldDiffResults[fieldName] = matched
		} else {
			fieldDiffResults[fieldName] = removed
		}
	}

	results := make([]StructDiffResult, 0, len(fieldDiffResults))

	for fieldName, result := range fieldDiffResults {
		switch result {
		case matched:
			srcValue, dstValue := srcFieldValueMapping[fieldName], dstFieldValueMapping[fieldName]
			if reflect.DeepEqual(srcValue, dstValue) {
				continue
			}
			results = append(results, StructDiffResult{
				FieldName: fieldName,
				OldValue:  srcValue,
				NewValue:  dstValue,
			})
		case added:
			results = append(results, StructDiffResult{
				FieldName: fieldName,
				OldValue:  nil,
				NewValue:  dstFieldValueMapping[fieldName],
			})
		case removed:
			results = append(results, StructDiffResult{
				FieldName: fieldName,
				OldValue:  srcFieldValueMapping[fieldName],
				NewValue:  nil,
			})
		}
	}
	return results
}

func GetStructFieldValueMapping(s any) map[string]any {
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	results := make(map[string]any, v.NumField())

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i).Interface()
		results[field.Name] = value
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
