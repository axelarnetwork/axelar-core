package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCodec registers concrete types on codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgLink{}, "ethereum/MsgLink", nil)
	cdc.RegisterConcrete(&MsgVoteVerifiedTx{}, "ethereum/VoteVerifyTx", nil)
	cdc.RegisterConcrete(MsgVerifyErc20TokenDeploy{}, "ethereum/VerifyErc20TokenDeploy", nil)
	cdc.RegisterConcrete(MsgVerifyErc20Deposit{}, "ethereum/VerifyErc20Deposit", nil)
	cdc.RegisterConcrete(MsgSignTx{}, "ethereum/SignTx", nil)
	cdc.RegisterConcrete(MsgSignPendingTransfers{}, "ethereum/SignPendingTransfersTx", nil)
	cdc.RegisterConcrete(MsgSignDeployToken{}, "ethereum/SignDeployToken", nil)
	cdc.RegisterConcrete(MsgSignBurnTokens{}, "ethereum/SignBurnTokens", nil)
}

// ModuleCdc defines the module codec
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
