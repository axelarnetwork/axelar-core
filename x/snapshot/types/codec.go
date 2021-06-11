package types

import (
	"github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// RegisterLegacyAminoCodec registers concrete types on codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {}

// RegisterInterfaces registers types and interfaces with the given registry
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterInterface("exported.SDKValidator",
		(*exported.SDKValidator)(nil),
		&stakingtypes.Validator{},
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
