package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCodec registers concrete types on codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgKeygenStart{}, "tss/MsgKeygenStart", nil)
	cdc.RegisterConcrete(MsgKeygenTraffic{}, "tss/MsgKeygenTraffic", nil)
	cdc.RegisterConcrete(MsgSignTraffic{}, "tss/MsgSignTraffic", nil)
	cdc.RegisterConcrete(MsgAssignNextKey{}, "tss/MsgAssignNextKey", nil)
	cdc.RegisterConcrete(MsgRotateKey{}, "tss/MsgRotateKey", nil)
	cdc.RegisterConcrete(&MsgVoteSig{}, "tss/MsgVoteSig", nil)
	cdc.RegisterConcrete(&MsgVotePubKey{}, "tss/MsgVotePubKey", nil)
	cdc.RegisterConcrete(MsgDeregister{}, "tss/MsgDeregister", nil)

	// this module's votes contain byte slices and for the VotingData interface
	cdc.RegisterConcrete([]byte{}, "tss/bytes", nil)
}

// ModuleCdc defines the module codec
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
