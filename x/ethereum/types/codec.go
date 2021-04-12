package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCodec registers concrete types on codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgLink{}, "ethereum/MsgLink", nil)
	cdc.RegisterConcrete(MsgVoteConfirmToken{}, "ethereum/VoteConfirmToken", nil)
	cdc.RegisterConcrete(MsgVoteConfirmDeposit{}, "ethereum/VoteConfirmDeposit", nil)
	cdc.RegisterConcrete(MsgConfirmToken{}, "ethereum/ConfirmToken", nil)
	cdc.RegisterConcrete(MsgConfirmDeposit{}, "ethereum/ConfirmDeposit", nil)
	cdc.RegisterConcrete(MsgSignTx{}, "ethereum/SignTx", nil)
	cdc.RegisterConcrete(MsgSignPendingTransfers{}, "ethereum/SignPendingTransfers", nil)
	cdc.RegisterConcrete(MsgSignDeployToken{}, "ethereum/SignDeployToken", nil)
	cdc.RegisterConcrete(MsgSignBurnTokens{}, "ethereum/SignBurnTokens", nil)
	cdc.RegisterConcrete(MsgSignTransferOwnership{}, "ethereum/SignTransferOwnership", nil)
}

// ModuleCdc defines the module codec
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
