package gormcache

import (
	"encoding/json"

	"github.com/vmihailenco/msgpack/v5"
)

// Serializer defines the interface for data serialization
type Serializer interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
}

// JSONSerializer implements JSON serialization
type JSONSerializer struct{}

// Marshal serializes v to JSON bytes
func (j *JSONSerializer) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// Unmarshal deserializes JSON bytes to v
func (j *JSONSerializer) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// MsgPackSerializer implements MessagePack serialization
type MsgPackSerializer struct{}

// Marshal serializes v to MessagePack bytes
func (m *MsgPackSerializer) Marshal(v interface{}) ([]byte, error) {
	return msgpack.Marshal(v)
}

// Unmarshal deserializes MessagePack bytes to v
func (m *MsgPackSerializer) Unmarshal(data []byte, v interface{}) error {
	return msgpack.Unmarshal(data, v)
}
