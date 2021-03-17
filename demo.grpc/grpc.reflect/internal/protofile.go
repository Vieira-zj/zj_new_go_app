package internal

import (
	"fmt"
	"strings"

	"github.com/fullstorydev/grpcurl"
)

// PrintProtoFileInfo prints proto files info.
func PrintProtoFileInfo(descSource grpcurl.DescriptorSource) error {
	allFiles, err := grpcurl.GetAllFiles(descSource)
	if err != nil {
		return err
	}

	for _, fd := range allFiles {
		name := fd.GetFullyQualifiedName()
		if strings.Index(name, "grpc_reflection_v1alpha") > -1 {
			continue
		}

		fmt.Println("protofile:", name)
		if strings.Index(name, "hello_msg.proto") > -1 {
			fmt.Println(fd.AsProto().String())
		}
	}
	return nil
}
