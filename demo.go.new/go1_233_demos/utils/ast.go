package utils

import "go/types"

func IsImplement(baseType types.Type, iface *types.Interface) bool {
	if types.Implements(baseType, iface) {
		// base type implements the interface
		return true
	}

	pointerType := types.NewPointer(baseType)
	if types.Implements(pointerType, iface) {
		// only the pointer type implements the interface
		return true
	}
	return false
}
