package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/voting/exported"
)

type TalliedVote struct {
	Tally sdk.Int
	Data  exported.VotingData
}

type Poll struct {
	Meta  exported.PollMeta
	Votes []TalliedVote
	// nil as long as the poll is undecided
	Result exported.Vote
}

type VoteResult struct {
	exported.PollMeta
	exported.VotingData
}

func (v VoteResult) Poll() exported.PollMeta {
	return v.PollMeta
}

func (v VoteResult) Data() exported.VotingData {
	return v.VotingData
}

type VotingThreshold struct {
	// split threshold into numerator and denominator to avoid floating point errors down the line
	Numerator   int64
	Denominator int64
}

func (t VotingThreshold) IsMet(accept sdk.Int, total sdk.Int) bool {
	return accept.MulRaw(t.Denominator).GT(total.MulRaw(t.Numerator))
}
