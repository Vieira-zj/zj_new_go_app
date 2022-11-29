package codec

import (
	"encoding/json"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/encoding"
	"google.golang.org/protobuf/encoding/protojson"
	protoV2 "google.golang.org/protobuf/proto"
)

func init() {
	encoding.RegisterCodec(ProtoJson{})
}

type ProtoJson struct{}

func (ProtoJson) Name() string {
	return "protojson"
}

func (ProtoJson) Marshal(v interface{}) ([]byte, error) {
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
