package encoding_fix

import (
	"github.com/gogo/protobuf/proto"
	"google.golang.org/grpc/encoding"
	encproto "google.golang.org/grpc/encoding/proto"
)

func init() {
	encoding.RegisterCodec(GogoEnabled{Codec: encoding.GetCodec(encproto.Name)})
}

type GogoEnabled struct {
	encoding.Codec
}

func (c GogoEnabled) Marshal(v interface{}) ([]byte, error) {
	if vv, ok := v.(proto.Marshaler); ok {
		return vv.Marshal()
	}
	return c.Codec.Marshal(v)
}

func (c GogoEnabled) Unmarshal(data []byte, v interface{}) error {
	if vv, ok := v.(proto.Unmarshaler); ok {
		return vv.Unmarshal(data)
	}
	return c.Codec.Unmarshal(data, v)
}
