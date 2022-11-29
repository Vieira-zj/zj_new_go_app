package codec

import "google.golang.org/grpc"

const tagJson = "json"

func WithJsonCodec() grpc.CallOption {
	return grpc.CallContentSubtype(tagJson)
}

type JsonFrame struct {
	RawData []byte
}

func (JsonFrame) Name() string {
	return tagJson
}

func (JsonFrame) Marshal(v interface{}) ([]byte, error) {
	frame, _ := v.(*JsonFrame)
	return frame.RawData, nil
}

func (JsonFrame) Unmarshal(b []byte, v interface{}) error {
	frame, _ := v.(*JsonFrame)
	frame.RawData = b
	return nil
}
