package imp1

import "bufio"

// Imp1 foreign type to import.
type Imp1 struct {
}

// ImpT type to import.
type ImpT int

//ForeignEmbedded for mock impl.
type ForeignEmbedded interface {
	// The return value here also makes sure that the generated mock picks up the "bufio" import.
	ForeignEmbeddedMethod() *bufio.Reader

	// This method uses a type in this package,
	// which should be qualified when this interface is embedded.
	ImplicitPackage(s string, t ImpT, st []ImpT, pt *ImpT, ct chan ImpT)
}
