package types

import (
	"crypto/elliptic"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcutil"
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCodec registers concrete types on codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgVoteConfirmOutpoint{}, "bitcoin/VoteConfirmOutpoint", nil)
	cdc.RegisterConcrete(MsgConfirmOutpoint{}, "bitcoin/ConfirmOutpoint", nil)
	cdc.RegisterConcrete(MsgLink{}, "bitcoin/Link", nil)
	cdc.RegisterConcrete(MsgSignPendingTransfers{}, "bitcoin/SignPendingTransfers", nil)
	cdc.RegisterInterface((*btcutil.Address)(nil), nil)
	cdc.RegisterConcrete(&btcutil.AddressPubKeyHash{}, "bitcoin/pkhash", nil)
	cdc.RegisterInterface((*elliptic.Curve)(nil), nil)
	cdc.RegisterConcrete(btcec.S256(), "bitcoin/curve", nil)
}

// ModuleCdc defines the module codec
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
