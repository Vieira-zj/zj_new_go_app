package pkg

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/golang/protobuf/jsonpb"
	"github.com/jhump/protoreflect/desc/protoparse"
)

func TestBuildReqProtoMessage(t *testing.T) {
	parser := protoparse.Parser{
		ImportPaths: []string{
			filepath.Join(os.Getenv("PROJECT_ROOT"), "protoc/proto/greeter"),
		},
		InferImportPaths: true,
	}

	fileName := "helloworld.proto"
	mDescs, err := parseProtoFile(parser, fileName)
	if err != nil {
		t.Fatal(err)
	}

	method := "/greeter.Greeter/SayHello"
	body := `{"name":"foo"}`
	coder := newCoder(mDescs)
	req, err := coder.buildReqProtoMessage(method, body)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("request:", req.String())

	decoder := jsonpb.Marshaler{}
	reqStr, err := decoder.MarshalToString(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("decode request:", reqStr)
}
