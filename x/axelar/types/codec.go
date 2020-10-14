package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCodec registers concrete types on codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgTrackAddress{}, "axelar/TrackAddress", nil)
	cdc.RegisterConcrete(MsgVerifyTx{}, "axelar/VerifyTx", nil)
	cdc.RegisterConcrete(MsgBatchVote{}, "axelar/BatchVote", nil)
	cdc.RegisterConcrete(MsgRegisterVoter{}, "axelar/RegisterVoter", nil)
}

// ModuleCdc defines the module codec
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
