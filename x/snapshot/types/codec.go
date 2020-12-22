package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/staking"

	"github.com/axelarnetwork/axelar-core/x/snapshot/exported"
)

// RegisterCodec registers concrete types on codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgSnapshot{}, "snapshot/MsgSnapshot", nil)
	cdc.RegisterInterface((*exported.Validator)(nil), nil)

	/* The snapshot keeper is dependent on the StakingKeeper interface, which returns validators through interfaces.
	However, the snapshot keeper has to marshal the validators, so it must register the actual concrete type that is returned. */
	cdc.RegisterConcrete(&staking.Validator{}, "snapshot/Validator", nil)
}

// ModuleCdc defines the module codec
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
