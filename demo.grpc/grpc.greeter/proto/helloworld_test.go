package proto

import (
	"encoding/json"
	"fmt"
	"testing"

	message "demo.grpc/grpc.greeter/proto/message"
	"github.com/golang/protobuf/proto"
)

func TestProtobufAndJson(t *testing.T) {
	// json to struct
	jsonStr := `{"name": "world"}`
	req := &message.HelloRequest{}
	if err := json.Unmarshal([]byte(jsonStr), req); err != nil {
		t.Fatal(err)
	}
	fmt.Println(proto.MarshalTextString(req))

	// struct to pb data
	data, err := proto.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(data))

	// pb data to struct
	newReq := message.HelloRequest{}
	err = proto.Unmarshal(data, &newReq)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(newReq.Name)

	// struct to json
	b, err := json.Marshal(&newReq)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(b))
}
