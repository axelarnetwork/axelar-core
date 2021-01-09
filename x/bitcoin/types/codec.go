package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCodec registers concrete types on codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgTrack{}, "bitcoin/MsgTrack", nil)
	cdc.RegisterConcrete(MsgVerifyTx{}, "bitcoin/VerifyTx", nil)
	cdc.RegisterConcrete(MsgRawTx{}, "bitcoin/RawTx", nil)
	cdc.RegisterConcrete(MsgSendTx{}, "bitcoin/Withdraw", nil)
	cdc.RegisterConcrete(&MsgVoteVerifiedTx{}, "bitcoin/MsgVoteVerifiedTx", nil)
}

// ModuleCdc defines the module codec
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
