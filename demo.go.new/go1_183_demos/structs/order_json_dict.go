package structs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
)

// OrderedJsonDict prints 1-level json string in ordered key.
type OrderedJsonDict struct {
	m    map[string]json.RawMessage
	keys []string
}

func (o *OrderedJsonDict) MarshalJSON() ([]byte, error) {
	buf := &bytes.Buffer{}
	buf.WriteByte('{')

	sort.Strings(o.keys)
	for i, k := range o.keys {
		if i != 0 {
			buf.WriteByte(',')
		}
		buf.WriteByte('"')
		buf.WriteString(k)
		buf.WriteByte('"')
		buf.WriteByte(':')
		buf.Write(o.m[k])
	}
	buf.WriteByte('}')

	return buf.Bytes(), nil
}

func (o *OrderedJsonDict) UnmarshalJSON(b []byte) error {
	const delimLeft, delimRight = json.Delim('{'), json.Delim('}')
	dec := json.NewDecoder(bytes.NewReader(b))
	if t, err := dec.Token(); err != nil {
		return fmt.Errorf("decoder.Token error: %w", err)
	} else if t != delimLeft {
		return fmt.Errorf("not a JSON object")
	}

	if o.m == nil {
		o.m = make(map[string]json.RawMessage)
	}

	for dec.More() {
		var key string
		t, err := dec.Token()
		if err != nil {
			return fmt.Errorf("decoder.Token error: %w", err)
		}
		if t == delimRight {
			// not go here
			fmt.Println("unmarshal end of json")
			break
		}

		// check json key
		var ok bool
		if key, ok = t.(string); !ok {
			return fmt.Errorf("not a JSON object")
		}

		// handle json value
		var rm json.RawMessage
		if err := dec.Decode(&rm); err != nil {
			return fmt.Errorf("decoder.Decode value error: %w", err)
		} else {
			o.m[key] = rm
			o.keys = append(o.keys, key)
		}
	}

	return nil
}

func (o *OrderedJsonDict) Get(key string) (json.RawMessage, bool) {
	v, ok := o.m[key]
	return v, ok
}

func (o *OrderedJsonDict) Set(key string, value json.RawMessage) {
	o.m[key] = value
	for _, k := range o.keys {
		if k == key {
			return
		}
	}
	o.keys = append(o.keys, key)
}

func (o *OrderedJsonDict) Delete(key string) {
	delete(o.m, key)
	for i, k := range o.keys {
		if k == key {
			o.keys = append(o.keys[:i], o.keys[i+1:]...)
			return
		}
	}
}

// StreamJsonArray handles json array items in stream.
type StreamJsonArray struct {
	items []json.RawMessage
}

type JsonArrayStreamHandler func([]json.RawMessage) error

func (s *StreamJsonArray) MarshalJSON() ([]byte, error) {
	buf := &bytes.Buffer{}
	buf.WriteByte('[')

	for i, item := range s.items {
		if i != 0 {
			buf.WriteByte(',')
		}
		buf.Write(item)
	}
	buf.WriteByte(']')

	return buf.Bytes(), nil
}

func (s *StreamJsonArray) UnmarshalJSON(b []byte) error {
	const delimLeft = json.Delim('[')
	dec := json.NewDecoder(bytes.NewReader(b))
	if t, err := dec.Token(); err != nil {
		return fmt.Errorf("decoder.Token error: %w", err)
	} else if t != delimLeft {
		return fmt.Errorf("not a JSON array")
	}

	if s.items == nil {
		s.items = make([]json.RawMessage, 0)
	}

	for dec.More() {
		var rm json.RawMessage
		if err := dec.Decode(&rm); err != nil {
			return fmt.Errorf("decoder.Decode value error: %w", err)
		} else {
			s.items = append(s.items, rm)
		}
	}

	return nil
}

func (s *StreamJsonArray) StreamUnmarshalJSON(b []byte, batchSize int, handler JsonArrayStreamHandler) error {
	const delimLeft = json.Delim('[')
	dec := json.NewDecoder(bytes.NewReader(b))
	if t, err := dec.Token(); err != nil {
		return fmt.Errorf("decoder.Token error: %w", err)
	} else if t != delimLeft {
		return fmt.Errorf("not a JSON array")
	}

	if s.items == nil {
		s.items = make([]json.RawMessage, 0)
	}

	for dec.More() {
		var rm json.RawMessage
		if err := dec.Decode(&rm); err != nil {
			return fmt.Errorf("decoder.Decode value error: %w", err)
		} else {
			s.items = append(s.items, rm)
		}

		if len(s.items) >= batchSize {
			if err := handler(s.items); err != nil {
				return err
			}
			s.reset()
		}
	}

	if len(s.items) > 0 {
		if err := handler(s.items); err != nil {
			return err
		}
		s.reset()
	}
	return nil
}

func (s *StreamJsonArray) Index(index int) (json.RawMessage, error) {
	if index < 0 || index >= len(s.items) {
		return nil, fmt.Errorf("index out of range")
	}
	return s.items[index], nil
}

func (s *StreamJsonArray) Append(item json.RawMessage) {
	s.items = append(s.items, item)
}

func (s *StreamJsonArray) reset() {
	s.items = s.items[:0]
}
