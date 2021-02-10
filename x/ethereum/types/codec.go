package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCodec registers concrete types on codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgVoteVerifiedToken{}, "ethereum/VoteVerifiedToken", nil)
	cdc.RegisterConcrete(&MsgVoteVerifiedTx{}, "ethereum/VoteVerifyTx", nil)
	cdc.RegisterConcrete(MsgVerifyToken{}, "ethereum/VoteVerifyToken", nil)
	cdc.RegisterConcrete(MsgVerifyTx{}, "ethereum/VerifyTx", nil)
	cdc.RegisterConcrete(MsgSignTx{}, "ethereum/SignTx", nil)
	cdc.RegisterConcrete(MsgSignPendingTransfers{}, "ethereum/SignPendingTransfersTx", nil)
	cdc.RegisterConcrete(MsgSignDeployToken{}, "ethereum/SignDeployToken", nil)
}

// ModuleCdc defines the module codec
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
