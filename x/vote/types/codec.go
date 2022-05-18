package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	reward "github.com/axelarnetwork/axelar-core/x/reward/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
	exported0_17 "github.com/axelarnetwork/axelar-core/x/vote/exported017"
)

// RegisterLegacyAminoCodec registers concrete types on codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&VoteRequest{}, "vote/Vote", nil)
}

// RegisterInterfaces registers types and interfaces with the given registry
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil), &VoteRequest{})

	registry.RegisterImplementations((*reward.Refundable)(nil), &VoteRequest{})

	registry.RegisterImplementations((*codec.ProtoMarshaler)(nil), &vote.Vote{})
	registry.RegisterImplementations((*codec.ProtoMarshaler)(nil), &exported0_17.Vote{})
}

var amino = codec.NewLegacyAmino()

// ModuleCdc defines the module codec
var ModuleCdc = codec.NewAminoCodec(amino)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()
}
