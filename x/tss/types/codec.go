package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCodec registers concrete types on codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgKeygenStart{}, "tss/MsgKeygenStart", nil)
	cdc.RegisterConcrete(MsgKeygenTraffic{}, "tss/MsgKeygenTraffic", nil)
	cdc.RegisterConcrete(MsgSignStart{}, "tss/MsgSignStart", nil)
	cdc.RegisterConcrete(MsgSignTraffic{}, "tss/MsgSignTraffic", nil)

	// used in tss/keeper/querier.go
	// cdc.RegisterInterface((*elliptic.Curve)(nil), nil)
	// cdc.RegisterConcrete(&btcec.KoblitzCurve{}, "tss/elliptic.Curve", nil)
}

// ModuleCdc defines the module codec
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
