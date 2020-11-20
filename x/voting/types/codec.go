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

// ModuleCdc defines the module codec
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
