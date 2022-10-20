package utils

import (
	"crypto/sha256"

	"github.com/cosmos/cosmos-sdk/codec"
)

// Hash returns the sha256 hash of the given protobuf data
func Hash(data codec.ProtoMarshaler) []byte {
	bz, err := data.Marshal()
	if err != nil {
		panic(err)
	}

	hash := sha256.Sum256(bz)

	return hash[:]
}
