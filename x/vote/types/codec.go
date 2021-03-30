package types

import (
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// RegisterCodec registers concrete types on codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*exported.VotingData)(nil), nil)
	cdc.RegisterInterface((*exported.Vote)(nil), nil)

	// Default type for voting, i.e. yes/no vote. Modules need to register their own types if they need more elaborate VotingData
	cdc.RegisterConcrete(true, "vote/VotingData", nil)
}

// ModuleCdc defines the module codec
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
