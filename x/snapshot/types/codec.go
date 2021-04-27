package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/axelarnetwork/axelar-core/x/snapshot/exported"
)

// RegisterLegacyAminoCodec registers concrete types on codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterInterface((*exported.SDKValidator)(nil), nil)
	/* The snapshot keeper is dependent on the StakingKeeper interface, which returns validators through interfaces.
	However, the snapshot keeper has to marshal the validators, so it must register the actual concrete type that is returned. */
	cdc.RegisterConcrete(&stakingtypes.Validator{}, "snapshot/SDKValidator", nil)
	cdc.RegisterConcrete(&exported.Validator{}, "snapshot/Validator", nil)
}

var amino = codec.NewLegacyAmino()

// ModuleCdc defines the module codec
var ModuleCdc = codec.NewAminoCodec(amino)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()
}
