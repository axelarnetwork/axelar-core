package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCodec registers concrete types on codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(&MsgVoteVerifiedTx{}, "ethereum/VoteVerifyTx", nil)
	cdc.RegisterConcrete(MsgVerifyTx{}, "ethereum/VerifyTx", nil)
	cdc.RegisterConcrete(MsgRawTx{}, "ethereum/RawTx", nil)
	cdc.RegisterConcrete(MsgInstallSC{}, "ethereum/InstallSC", nil)
	cdc.RegisterConcrete(MsgSendTx{}, "ethereum/Send", nil)

}

// ModuleCdc defines the module codec
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
