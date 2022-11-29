package pkg

import (
	"fmt"
	"log"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
)

// Coder .
type Coder struct {
	MethodDescs map[string]*desc.MethodDescriptor
}

func newCoder(mDescs map[string]*desc.MethodDescriptor) Coder {
	return Coder{
		MethodDescs: mDescs,
	}
}

func (c Coder) buildReqProtoMessage(method, body string) (proto.Message, error) {
	dynMsg, err := c.newProtoMessage(method, inputMsgType)
	if err != nil {
		return nil, err
	}

	if err := jsonpb.UnmarshalString(body, dynMsg); err != nil {
		return nil, fmt.Errorf("jsonpb unmarshal error: %v", err)
	}
	return dynMsg, nil
}

func (c Coder) newRespProtoMessage(method string) (proto.Message, error) {
	return c.newProtoMessage(method, outputMsgType)
}

const (
	inputMsgType  = "input"
	outputMsgType = "output"
)

func (c Coder) newProtoMessage(method, msgType string) (proto.Message, error) {
	md, ok := c.MethodDescs[method]
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
