package types

import (
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/axelarnetwork/axelar-core/x/voting/exported"
)

// RegisterCodec registers concrete types on codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgBallot{}, "voting/SendBallot", nil)
	cdc.RegisterConcrete(VoteResult{}, "voting/VoteResult", nil)
	cdc.RegisterInterface((*exported.MsgVote)(nil), nil)
	cdc.RegisterInterface((*exported.VotingData)(nil), nil)
	cdc.RegisterInterface((*exported.Vote)(nil), nil)
}

// ModuleCdc defines the module codec. For the voting module, this must be set from the app.go,
// because it needs access to a codec that has registered all concrete message types from all modules
var ModuleCdc *codec.Codec
