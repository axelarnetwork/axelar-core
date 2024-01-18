package encoding_override

import (
	"github.com/gogo/protobuf/proto"
	"google.golang.org/grpc/encoding"
	encproto "google.golang.org/grpc/encoding/proto"
)

// This registers a codec that can encode custom Golang types defined by gogoproto extensions, which newer versions of the grpc module cannot.
// The fix has been extracted into its own module in order to minimize the number of dependencies
// that get imported before this init() function is called.
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
