package protoc

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/jhump/protoreflect/desc/protoparse"
)

func TestParseAccountProtoFile(t *testing.T) {
	parser := &protoparse.Parser{
		ImportPaths: []string{
			filepath.Join(os.Getenv("PROJECT_ROOT"), "grpc.app/proto/account"), // 指定包含 proto 的目录
		},
		InferImportPaths: true,
	}

	fileName := "deposit.proto"
	mDescs, err := parseProtoFile(parser, fileName)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("load:", fileName)
	for _, mDesc := range mDescs {
		t.Log("method:", mDesc.GetName())
	}
}

func TestParseGreeterProtoFile(t *testing.T) {
	parser := &protoparse.Parser{
		ImportPaths: []string{
			filepath.Join(os.Getenv("PROJECT_ROOT"), "grpc.app/proto/greeter"),
		},
		InferImportPaths: true,
	}

	fileName := "helloworld.proto"
	mDescs, err := parseProtoFile(parser, fileName)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("load:", fileName)
	for _, mDesc := range mDescs {
		t.Log("method:", mDesc.GetName())
	}
}

func TestLoadProtoFiles01(t *testing.T) {
	paths := []string{
		filepath.Join(os.Getenv("PROJECT_ROOT"), "grpc.app/proto/account"),
		filepath.Join(os.Getenv("PROJECT_ROOT"), "grpc.app/proto/greeter"),
	}
	mDescs, err := LoadProtoFiles(paths...)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("methods desc:")
	for key, md := range mDescs {
		t.Log(key, md.GetName())
	}
}

func TestLoadProtoFiles02(t *testing.T) {
	path := filepath.Join(os.Getenv("PROJECT_ROOT"), "grpc.app/proto")
	dirPaths, err := GetAllProtoDirs(path)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("proto dirs:")
	for _, path := range dirPaths {
		fmt.Println(path)
	}

	mDescs, err := LoadProtoFiles(dirPaths...)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("methods desc:")
	for key, md := range mDescs {
		t.Log(key, md.GetName())
	}
}
