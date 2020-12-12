package types

import (
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// RegisterCodec registers concrete types on codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgBallot{}, "vote/SendBallot", nil)
	cdc.RegisterInterface((*exported.MsgVote)(nil), nil)
	cdc.RegisterInterface((*exported.VotingData)(nil), nil)
	cdc.RegisterInterface((*exported.Vote)(nil), nil)

	// Default type for voting, i.e. yes/no vote. Modules need to register their own types if they need more elaborate VotingData
	cdc.RegisterConcrete(true, "vote/VotingData", nil)
}

// ModuleCdc defines the module codec. For the vote module, this must be set from the app.go,
// because it needs access to a codec that has registered all concrete message types from all modules.
// Thus, the codec is not initialized in an init() function as in all the other modules.
var ModuleCdc *codec.Codec
