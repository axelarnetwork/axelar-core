package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterLegacyAminoCodec registers concrete types on codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&UpdateGovernanceKeyRequest{}, "permission/UpdateGovernanceKey", nil)
	cdc.RegisterConcrete(&RegisterControllerRequest{}, "permission/RegisterController", nil)
	cdc.RegisterConcrete(&DeregisterControllerRequest{}, "permission/DeregisterController", nil)
	cdc.RegisterConcrete(&UpdateParamsRequest{}, "permission/UpdateParams", nil)
	cdc.RegisterConcrete(&PauseRequest{}, "permission/Pause", nil)
	cdc.RegisterConcrete(&UnpauseRequest{}, "permission/Unpause", nil)
	cdc.RegisterConcrete(&HaltChainRequest{}, "permission/HaltChain", nil)
}

// RegisterInterfaces registers types and interfaces with the given registry
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&UpdateGovernanceKeyRequest{},
		&RegisterControllerRequest{},
		&DeregisterControllerRequest{},
		&UpdateParamsRequest{},
		&PauseRequest{},
		&UnpauseRequest{},
		&HaltChainRequest{},
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
