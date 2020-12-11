package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

type TalliedVote struct {
	Tally sdk.Int
	Data  exported.VotingData
}

type Poll struct {
	Meta                   exported.PollMeta
	ValidatorSnapshotRound int64
	Votes                  []TalliedVote
	// nil as long as the poll is undecided
	Result exported.VotingData
}
