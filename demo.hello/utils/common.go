package utils

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"reflect"
	"time"
	"unsafe"
)

//
// Date time
//

// IsWeekDay .
func IsWeekDay(t time.Time) bool {
	switch t.Weekday() {
	case time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday:
		return true
	case time.Saturday, time.Sunday:
		return false
	}
	fmt.Println("Unrecognized day of the week:", t.Weekday().String())
	panic("Explicit Panic to avoid compiler error: missing return at end of function")
}

func nextWeekDay(loc *time.Location) time.Time {
	now := time.Now().In(loc)
	for !IsWeekDay(now) {
		now = now.AddDate(0, 0, 1)
	}
	return now
}

//
// Run shell
//

// RunShellCmd runs a shell command and returns output.
func RunShellCmd(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	defer stdout.Close()

	// run command
	if err = cmd.Start(); err != nil {
		return "", err
	}

	output, err := ioutil.ReadAll(stdout)
	if err != nil {
		return "", err
	}

	err = cmd.Wait()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

//
// Reflect
//

// GetStructFields returns struct field name and type desc.
func GetStructFields(ele reflect.Type) (map[string]interface{}, error) {
	if ele.Kind() == reflect.Ptr {
		ele = ele.Elem()
	}

	if ele.Kind() != reflect.Struct {
		return nil, fmt.Errorf("input element [%s] not struct", ele.Kind().String())
	}

	fieldsNum := ele.NumField()
	ret := make(map[string]interface{}, fieldsNum)
	for i := 0; i < fieldsNum; i++ {
		field := ele.Field(i)
		if field.Type.Kind() == reflect.Struct {
			value, err := GetStructFields(field.Type)
			if err != nil {
				return nil, err
			}
			ret[field.Name] = value
		} else if field.Type.Kind() == reflect.Slice {
			value, err := GetSliceElements(field.Type)
			if err != nil {
				return nil, err
			}
			ret[field.Name] = value
		} else {
			ret[field.Name] = field.Type.Kind().String()
		}
	}
	return ret, nil
}

// GetSliceElements gets slice element type desc.
func GetSliceElements(s reflect.Type) (string, error) {
	if s.Kind() == reflect.Ptr {
		s = s.Elem()
	}

	if s.Kind() != reflect.Slice {
		return "", fmt.Errorf("input element [%v] not slice", s.Kind().String())
	}

	ele := s.Elem()
	if ele.Kind() == reflect.Ptr {
		ele = ele.Elem()
	}

	var (
		res interface{}
		err error
	)
	if ele.Kind() == reflect.Struct {
		res, err = GetStructFields(ele)
	} else if ele.Kind() == reflect.Slice {
		res, err = GetSliceElements(ele)
	} else {
		res = ele.Kind().String()
	}
	return fmt.Sprintf("%s[%+v]", s.Kind().String(), res), err
}

// InitStructByReflect init a struct from map by reflection.
func InitStructByReflect(data map[string]interface{}, inStructPtr interface{}) error {
	inType := reflect.TypeOf(inStructPtr)
	inValue := reflect.ValueOf(inStructPtr)
	if inType.Kind() != reflect.Ptr {
		return fmt.Errorf("inStructPtr must be ptr to struct")
	}

	inType = inType.Elem()
	inValue = inValue.Elem()

	for i := 0; i < inType.NumField(); i++ {
		fType := inType.Field(i)   // struct field type
		fValue := inValue.Field(i) // struct field value

		key := fType.Tag.Get("key")
		v, ok := data[key]
		if !ok {
			return fmt.Errorf(fType.Name + " not found")
		}

		dataType := reflect.TypeOf(v)
		valueType := fValue.Type() // struct field value type
		if dataType == valueType {
			fValue.Set(reflect.ValueOf(v))
		} else if dataType.ConvertibleTo(valueType) {
			fValue.Set(reflect.ValueOf(v).Convert(valueType))
		} else {
			return fmt.Errorf(fType.Name + " type mismatch")
		}
	}
	return nil
}

// StringToSliceByte 可以不拷贝数据而将 string 转换成 []byte 了，不过要注意这样生成的 []byte 不可写，否则行为未定义。
func StringToSliceByte(s string) []byte {
	l := len(s)
	data := (*reflect.StringHeader)(unsafe.Pointer(&s)).Data
	return *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: data,
		Len:  l,
		Cap:  l,
	}))
}
