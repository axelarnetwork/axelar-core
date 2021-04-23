package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// TalliedVote represents a vote for a poll with the accumulated stake of all validators voting for the same VotingData
type TalliedVote struct {
	Tally sdk.Int
	Data  exported.VotingData
}

// Poll represents a poll with write-in voting, i.e. the result of the vote can have any data type
type Poll struct {
	Meta                     exported.PollMeta
	ValidatorSnapshotCounter int64
	Votes                    []TalliedVote
	// nil as long as the poll is undecided
	Result exported.VotingData
}
