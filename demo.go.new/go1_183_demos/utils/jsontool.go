package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
)

func IsValidJson(b []byte) bool {
	b = bytes.Trim(b, " ")
	if len(b) == 0 {
		return false
	}
	return (b[0] == '{' || b[0] == '[') && json.Valid(b)
}

func JsonLoad(b []byte, s any) error {
	if reflect.TypeOf(s).Kind() != reflect.Ptr {
		return fmt.Errorf("input should be pointer")
	}

	decoder := json.NewDecoder(bytes.NewBuffer(b))
	decoder.UseNumber()
	return decoder.Decode(s)
}

func JsonMarshalStream(r io.Reader, object any) error {
	decoder := json.NewDecoder(r)
	decoder.UseNumber()
	return decoder.Decode(object)
}

func JsonUnmarshalStream(w io.Writer, object any) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(object)
}

// UpdateValueByJsonPath updates value by json path (from jsondiff).
func UpdateValueByJsonPath(obj any, path string, newVal any) error {
	// use JsonPatch instead
	var parent, pKey any
	for _, key := range strings.Split(path, "/") {
		parent, pKey = obj, key
		if len(key) == 0 {
			continue
		}

		switch reflect.TypeOf(obj).Kind() {
		case reflect.Slice:
			idx, err := strconv.Atoi(key)
			if err != nil {
				return fmt.Errorf("invalid index: %v", key)
			}
			obj = obj.([]any)[idx]
			pKey = idx
		case reflect.Map:
			obj = obj.(map[string]any)[key]
		default:
			return fmt.Errorf("object must be slice or map")
		}
	}

	srcType := reflect.TypeOf(obj).Kind().String()
	destType := reflect.TypeOf(newVal).Kind().String()
	if srcType != destType {
		return fmt.Errorf("mismatch type: src=%s, dest=%s", srcType, destType)
	}

	switch reflect.TypeOf(parent).Kind() {
	case reflect.Slice:
		key := pKey.(int)
		parent.([]any)[key] = newVal
	case reflect.Map:
		key := pKey.(string)
		parent.(map[string]any)[key] = newVal
	}

	return nil
}

// GetValueByJsonPath returns value by json path (from jsondiff).
func GetValueByJsonPath(obj any, path string) (any, error) {
	for _, key := range strings.Split(path, "/") {
		if len(key) == 0 {
			continue
		}

		switch reflect.TypeOf(obj).Kind() {
		case reflect.Slice:
			idx, err := strconv.Atoi(key)
			if err != nil {
				return nil, fmt.Errorf("invalid index: %v", key)
			}
			obj = obj.([]any)[idx]
		case reflect.Map:
			obj = obj.(map[string]any)[key]
		default:
			return nil, fmt.Errorf("object must be slice or map")
		}
	}

	return obj, nil
}
