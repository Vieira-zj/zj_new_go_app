package codec

import (
	"encoding/json"
	"log"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
	"google.golang.org/protobuf/encoding/protojson"
	protoV2 "google.golang.org/protobuf/proto"
)

func init() {
	encoding.RegisterCodec(ProtoJson{})
}

const tagProtoJson = "protojson"

func WithProtoJsonCodec() grpc.CallOption {
	return grpc.CallContentSubtype(tagProtoJson)
}

type ProtoJson struct{}

func (ProtoJson) Name() string {
	return tagProtoJson
}

func (ProtoJson) Marshal(v interface{}) ([]byte, error) {
	log.Println("ProtoJson.Marshal")
	marshalOpts := protojson.MarshalOptions{
		UseProtoNames:   true,
		UseEnumNumbers:  true,
		EmitUnpopulated: true,
	}

	if msg, ok := v.(proto.Message); ok {
		return marshalOpts.Marshal(proto.MessageV2(msg))
	}
	if msg, ok := v.(protoV2.Message); ok {
		return marshalOpts.Marshal(msg)
	}
	return json.Marshal(v)
}

func (ProtoJson) Unmarshal(b []byte, v interface{}) error {
	log.Println("ProtoJson.Unmarshal")
	unmarshalOpts := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
	if msg, ok := v.(proto.Message); ok {
		return unmarshalOpts.Unmarshal(b, proto.MessageV2(msg))
	}
	if msg, ok := v.(protoV2.Message); ok {
		return unmarshalOpts.Unmarshal(b, msg)
	}
	return json.Unmarshal(b, v)
}
