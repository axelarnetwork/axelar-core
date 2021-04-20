package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
)

// RegisterLegacyAminoCodec registers concrete types on codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgLink{}, "ethereum/MsgLink", nil)
	cdc.RegisterConcrete(&MsgVoteConfirmToken{}, "ethereum/VoteConfirmToken", nil)
	cdc.RegisterConcrete(&MsgVoteConfirmDeposit{}, "ethereum/VoteConfirmDeposit", nil)
	cdc.RegisterConcrete(&MsgConfirmToken{}, "ethereum/ConfirmToken", nil)
	cdc.RegisterConcrete(&MsgConfirmDeposit{}, "ethereum/ConfirmDeposit", nil)
	cdc.RegisterConcrete(&MsgSignTx{}, "ethereum/SignTx", nil)
	cdc.RegisterConcrete(&MsgSignPendingTransfers{}, "ethereum/SignPendingTransfers", nil)
	cdc.RegisterConcrete(&MsgSignDeployToken{}, "ethereum/SignDeployToken", nil)
	cdc.RegisterConcrete(&MsgSignBurnTokens{}, "ethereum/SignBurnTokens", nil)
	cdc.RegisterConcrete(&MsgSignTransferOwnership{}, "ethereum/SignTransferOwnership", nil)
}

var amino = codec.NewLegacyAmino()

// ModuleCdc defines the module codec
var ModuleCdc = codec.NewAminoCodec(amino)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()
}
