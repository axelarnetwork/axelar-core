package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCodec registers concrete types on codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgWithdraw{}, "ethbridge/TrackAddress", nil)
	cdc.RegisterConcrete(MsgVerifyTx{}, "ethbridge/VerifyTx", nil)

	//	TODO: Replace bool with full bitcoin tx data when ethbridge uses write-in voting pattern
	cdc.RegisterConcrete(true, "ethbridge/VotingData", nil)
}

// ModuleCdc defines the module codec
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
