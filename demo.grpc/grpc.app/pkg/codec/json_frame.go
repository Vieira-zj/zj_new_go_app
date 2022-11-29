package codec

type JsonFrame struct {
	RawData []byte
}

func (JsonFrame) Name() string {
	return "json"
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
