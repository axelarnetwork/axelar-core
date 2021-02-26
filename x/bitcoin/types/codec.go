package types

import (
	"github.com/btcsuite/btcutil"
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCodec registers concrete types on codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(&MsgVoteVerifiedTx{}, "bitcoin/VoteVerifiedTx", nil)
	cdc.RegisterConcrete(MsgVerifyTx{}, "bitcoin/VerifyTx", nil)
	cdc.RegisterConcrete(MsgLink{}, "bitcoin/Link", nil)
	cdc.RegisterConcrete(MsgSignPendingTransfers{}, "bitcoin/SignPendingTransfers", nil)
	cdc.RegisterInterface((*btcutil.Address)(nil), nil)
	cdc.RegisterConcrete(&btcutil.AddressPubKeyHash{}, "bitcoin/pkhash", nil)
}

// ModuleCdc defines the module codec
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
