package protoc

import (
	"fmt"
	"log"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
)

const (
	inputMsgType  = "input"
	outputMsgType = "output"
)

// Coder uses deprecated proto which compitable with dynamic message.
type Coder struct {
	methodDescs map[string]*desc.MethodDescriptor
}

func NewCoder(mDescs map[string]*desc.MethodDescriptor) Coder {
	return Coder{
		methodDescs: mDescs,
	}
}

// BuildReqProtoMessage creates grpc request (proto message) from json string.
func (c Coder) BuildReqProtoMessage(method, body string) (proto.Message, error) {
	dynMsg, err := c.newProtoMessage(method, inputMsgType)
	if err != nil {
		return nil, err
	}

	if err := jsonpb.UnmarshalString(body, dynMsg); err != nil {
		return nil, fmt.Errorf("jsonpb unmarshal error: %v", err)
	}
	return dynMsg, nil
}

// NewRespProtoMessage creates empty grpc response (proto message).
func (c Coder) NewRespProtoMessage(method string) (proto.Message, error) {
	return c.newProtoMessage(method, outputMsgType)
}

func (c Coder) newProtoMessage(method, msgType string) (proto.Message, error) {
	md, ok := c.methodDescs[method]
	if !ok {
		return nil, fmt.Errorf("method descriptor not found for: %s", method)
	}

	var msgDesc *desc.MessageDescriptor
	switch msgType {
	case inputMsgType:
		msgDesc = md.GetInputType()
	case outputMsgType:
		msgDesc = md.GetOutputType()
	default:
		return nil, fmt.Errorf("invalid msg type")
	}

	log.Printf("create new proto message for: %s", msgDesc.GetName())
	return dynamic.NewMessage(msgDesc), nil
}
