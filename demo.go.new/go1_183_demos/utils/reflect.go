package utils

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
)

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

	fullName := GetFullFnName(fn)
	shortName := fullName[strings.LastIndex(fullName, ".")+1:]

	return FnDeclaration{
		Name:   shortName,
		Input:  input,
		Output: output,
	}, nil
}
