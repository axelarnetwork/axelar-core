package exported2

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

// PollProperty is a modifier for PollMetadata. It should never be manually initialized
type PollProperty struct {
	do func(metadata PollMetadata) PollMetadata
}

func (p PollProperty) apply(metadata PollMetadata) PollMetadata {
	return p.do(metadata)
}

var _ codectypes.UnpackInterfacesMessage = PollMetadata{}

// UnpackInterfaces implements UnpackInterfacesMessage
func (m PollMetadata) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var data codec.ProtoMarshaler
	return unpacker.UnpackAny(m.Result, &data)
}

// With returns a new metadata object with all the given properties set
func (m PollMetadata) With(properties ...PollProperty) PollMetadata {
	newMetadata := m
	for _, property := range properties {
		newMetadata = property.apply(newMetadata)
	}
	return newMetadata
}
