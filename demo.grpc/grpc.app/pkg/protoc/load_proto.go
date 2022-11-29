package protoc

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
)

// LoadProtoFiles parses .proto files in paths, and loads method descriptors.
func LoadProtoFiles(paths ...string) (map[string]*desc.MethodDescriptor, error) {
	retMeDescs := make(map[string]*desc.MethodDescriptor, 16)
	parser := &protoparse.Parser{
		ImportPaths:      paths,
		InferImportPaths: true,
	}

	for _, path := range paths {
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			fname := entry.Name()
			if filepath.Ext(fname) == ".proto" {
				log.Println("parse proto file:", fname)
				mDescs, err := parseProtoFile(parser, fname)
				if err != nil {
					return nil, err
				}
				for k, md := range mDescs {
					if _, ok := retMeDescs[k]; ok {
						return nil, fmt.Errorf("duplicated method descriptor: %s", k)
					}
					retMeDescs[k] = md
				}
			}
		}
	}
	return retMeDescs, nil
}

func parseProtoFile(parser *protoparse.Parser, fname string) (map[string]*desc.MethodDescriptor, error) {
	fileDescs, err := parser.ParseFiles(fname)
	if err != nil {
		return nil, err
	}

	mDescs := map[string]*desc.MethodDescriptor{}
	fileDesc := fileDescs[0]
	for _, service := range fileDesc.GetServices() {
		srvName := service.GetFullyQualifiedName()
		for _, method := range service.GetMethods() {
			key := "/" + srvName + "/" + method.GetName()
			mDescs[key] = method

		}
	}
	return mDescs, nil
}
