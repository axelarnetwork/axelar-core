package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/types/msgservice"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/axelarnetwork/axelar-core/x/snapshot/exported"
)

// RegisterLegacyAminoCodec registers concrete types on codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterInterface((*exported.SDKValidator)(nil), nil)
	cdc.RegisterConcrete(&RegisterProxyRequest{}, "snapshot/RegisterProxy", nil)
	cdc.RegisterConcrete(&DeactivateProxyRequest{}, "snapshot/DeactivateProxy", nil)

	/* The snapshot keeper is dependent on the StakingKeeper interface, which returns validators through interfaces.
	However, the snapshot keeper has to marshal the validators, so it must register the actual concrete type that is returned. */
	cdc.RegisterConcrete(&stakingtypes.Validator{}, "snapshot/SDKValidator", nil)
	cdc.RegisterConcrete(&exported.Validator{}, "snapshot/Validator", nil)
}

// RegisterInterfaces registers types and interfaces with the given registry
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil), &RegisterProxyRequest{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &DeactivateProxyRequest{})

	msgservice.RegisterMsgServiceDesc(registry, &_MsgService_serviceDesc)
}

var amino = codec.NewLegacyAmino()

// ModuleCdc defines the module codec
var ModuleCdc = codec.NewAminoCodec(amino)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()
}
