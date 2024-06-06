package jsondiff

import "reflect"

func isSameKind(src, dst any) bool {
	srcKind, dstKind := reflect.TypeOf(src).Kind(), reflect.TypeOf(dst).Kind()
	return srcKind == dstKind
}
