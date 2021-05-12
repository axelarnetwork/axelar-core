package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterLegacyAminoCodec registers concrete types on codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgVoteConfirmOutpoint{}, "bitcoin/VoteConfirmOutpoint", nil)
	cdc.RegisterConcrete(&MsgConfirmOutpoint{}, "bitcoin/ConfirmOutpoint", nil)
	cdc.RegisterConcrete(&MsgLink{}, "bitcoin/Link", nil)
	cdc.RegisterConcrete(&MsgSignPendingTransfers{}, "bitcoin/SignPendingTransfers", nil)
}

// RegisterInterfaces registers types and interfaces with the given registry
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgVoteConfirmOutpoint{},
		&MsgConfirmOutpoint{},
		&MsgLink{},
		&MsgSignPendingTransfers{})
}

var amino = codec.NewLegacyAmino()

// ModuleCdc defines the module codec
var ModuleCdc = codec.NewAminoCodec(amino)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()
}
