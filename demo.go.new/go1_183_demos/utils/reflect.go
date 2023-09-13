package utils

import (
	"errors"
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
	return reflect.TypeOf(src).Kind() == reflect.TypeOf(dest).Kind() && src == dest
}
