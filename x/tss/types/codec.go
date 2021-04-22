package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterLegacyAminoCodec registers concrete types on codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgKeygenStart{}, "tss/MsgKeygenStart", nil)
	cdc.RegisterConcrete(&MsgKeygenTraffic{}, "tss/MsgKeygenTraffic", nil)
	cdc.RegisterConcrete(&MsgSignTraffic{}, "tss/MsgSignTraffic", nil)
	cdc.RegisterConcrete(&MsgAssignNextKey{}, "tss/MsgAssignNextKey", nil)
	cdc.RegisterConcrete(&MsgRotateKey{}, "tss/MsgRotateKey", nil)
	cdc.RegisterConcrete(&MsgVoteSig{}, "tss/MsgVoteSig", nil)
	cdc.RegisterConcrete(&MsgVotePubKey{}, "tss/MsgVotePubKey", nil)
	cdc.RegisterConcrete(&MsgDeregister{}, "tss/MsgDeregister", nil)

	// this module's votes contain byte slices and for the VotingData interface
	cdc.RegisterConcrete([]byte{}, "tss/bytes", nil)
}

// RegisterInterfaces registers types and interfaces with the given registry
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgKeygenStart{},
		&MsgKeygenTraffic{},
		&MsgSignTraffic{},
		&MsgAssignNextKey{},
		&MsgRotateKey{},
		&MsgVoteSig{},
		&MsgVotePubKey{},
		&MsgDeregister{},
	)
}

var amino = codec.NewLegacyAmino()

// ModuleCdc defines the module codec
var ModuleCdc = codec.NewAminoCodec(amino)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()
}
