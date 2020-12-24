package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCodec registers concrete types on codec
func RegisterCodec(_ *codec.Codec) {
}

// ModuleCdc defines the module codec
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
