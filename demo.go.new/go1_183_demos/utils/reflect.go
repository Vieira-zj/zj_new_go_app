package utils

import (
	"errors"
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
