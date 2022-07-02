package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/types"
	"github.com/axelarnetwork/utils/proto"
	"github.com/axelarnetwork/utils/slices"
)

var _ exported.Poll = &poll{}

type poll struct {
	exported.PollMetadata
	ctx sdk.Context
	k   Keeper
}

func newPoll(ctx sdk.Context, k Keeper, metadata exported.PollMetadata) *poll {
	return &poll{
		ctx:          ctx,
		k:            k,
		PollMetadata: metadata,
	}
}

func (p poll) logger() log.Logger {
	return p.k.Logger(p.ctx)
}

func (p poll) HasVotedCorrectly(voter sdk.ValAddress) bool {
	majorityVote := p.getMajorityVote()
	_, ok := majorityVote.IsVoterLate[voter.String()]

	return p.Is(exported.Completed) && ok
}

func (p poll) HasVoted(voter sdk.ValAddress) bool {
	return slices.Any(p.k.getTalliedVotes(p.ctx, p.ID), func(talliedVote types.TalliedVote) bool {
		_, ok := talliedVote.IsVoterLate[voter.String()]

		return ok
	})
}

// GetResult returns the result of the poll. Returns nil if the poll is not completed.
func (p poll) GetResult() codec.ProtoMarshaler {
	if p.Result == nil {
		return nil
	}

	return p.Result.GetCachedValue().(codec.ProtoMarshaler)
}

// GetRewardPoolName returns the name of the attached reward pool, if any
func (p poll) GetRewardPoolName() (string, bool) {
	return p.RewardPoolName, len(p.RewardPoolName) > 0
}

// GetVoters returns the poll's voters
func (p poll) GetVoters() []sdk.ValAddress {
	return p.Snapshot.GetParticipantAddresses()
}

// GetID returns the ID of the poll
func (p poll) GetID() exported.PollID {
	return p.ID
}

// GetState returns the state of the poll
func (p poll) GetState() exported.PollState {
	return p.State
}

// Vote records the given vote
func (p *poll) Vote(voter sdk.ValAddress, blockHeight int64, data codec.ProtoMarshaler) (exported.VoteResult, error) {
	if p.Is(exported.NonExistent) {
		return exported.NoVote, fmt.Errorf("poll does not exist")
	}

	if p.HasVoted(voter) {
		return exported.NoVote, fmt.Errorf("voter %s has voted already", voter)
	}

	if p.Snapshot.GetParticipantWeight(voter).IsZero() {
		return exported.NoVote, fmt.Errorf("address %s is not eligible to vote in this poll", voter)
	}

	if p.Is(exported.Failed) {
		return exported.NoVote, nil
	}

	if p.Is(exported.Completed) && p.isInGracePeriod(blockHeight) {
		p.voteLate(voter, data)

		return exported.VotedLate, nil
	}

	if p.Is(exported.Completed) {
		return exported.NoVote, nil
	}

	p.voteBeforeCompletion(voter, blockHeight, data)

	return exported.VoteInTime, nil
}

// GetModule returns the module the poll is associated with
func (p poll) GetModule() string {
	return p.Module
}

func (p poll) voteLate(voter sdk.ValAddress, data codec.ProtoMarshaler) {
	p.logger().Debug("received late vote for poll",
		"voter", voter.String(),
		"poll", p.ID.String(),
	)

	talliedVote, ok := p.k.getTalliedVote(p.ctx, p.ID, proto.Hash(data))
	if !ok {
		talliedVote = types.NewTalliedVote(p.ID, data)
	}

	talliedVote.TallyVote(voter, p.Snapshot.GetParticipantWeight(voter), true)
	p.k.setTalliedVote(p.ctx, talliedVote)
}

func (p *poll) voteBeforeCompletion(voter sdk.ValAddress, blockHeight int64, data codec.ProtoMarshaler) {
	p.logger().Debug("received vote for poll",
		"voter", voter.String(),
		"poll", p.ID.String(),
	)

	talliedVote, ok := p.k.getTalliedVote(p.ctx, p.ID, proto.Hash(data))
	if !ok {
		talliedVote = types.NewTalliedVote(p.ID, data)
	}

	talliedVote.TallyVote(voter, p.Snapshot.GetParticipantWeight(voter), false)
	p.k.setTalliedVote(p.ctx, talliedVote)

	majorityVote := p.getMajorityVote()
	switch {
	case p.hasEnoughVotes(majorityVote.Tally):
		p.Result = majorityVote.Data
		p.State = exported.Completed
		p.CompletedAt = blockHeight
		p.logger().Debug(fmt.Sprintf("poll %s (threshold: %d/%d, min voter count: %d) completed",
			p.ID,
			p.VotingThreshold.Numerator,
			p.VotingThreshold.Denominator,
			p.MinVoterCount,
		))

		p.k.setPollMetadata(p.ctx, p.PollMetadata)

	case p.cannotWin(majorityVote.Tally):
		p.State = exported.Failed
		p.logger().Debug(fmt.Sprintf("poll %s (threshold: %d/%d, min voter count: %d) failed, voters could not agree on single value",
			p.ID,
			p.VotingThreshold.Numerator,
			p.VotingThreshold.Denominator,
			p.MinVoterCount,
		))

		p.k.setPollMetadata(p.ctx, p.PollMetadata)
	}
}

func (p poll) hasEnoughVotes(majority sdk.Uint) bool {
	return majority.GTE(p.Snapshot.CalculateMinPassingWeight(p.VotingThreshold)) &&
		p.getVoterCount() >= p.MinVoterCount
}

func (p poll) cannotWin(majority sdk.Uint) bool {
	alreadyTallied := p.getTalliedVotingPower()
	missingVotingPower := p.Snapshot.GetParticipantsWeight().Sub(alreadyTallied)

	return majority.
		Add(missingVotingPower).
		LT(p.Snapshot.CalculateMinPassingWeight(p.VotingThreshold))
}

func (p poll) getTalliedVotingPower() sdk.Uint {
	result := sdk.ZeroUint()

	for _, talliedVote := range p.k.getTalliedVotes(p.ctx, p.ID) {
		result = result.Add(talliedVote.Tally)
	}

	return result
}

func (p poll) getVoterCount() int64 {
	var count int64

	for _, talliedVote := range p.k.getTalliedVotes(p.ctx, p.ID) {
		count += int64(len(talliedVote.IsVoterLate))
	}

	return count
}

func (p poll) isInGracePeriod(blockHeight int64) bool {
	return blockHeight <= p.CompletedAt+p.GracePeriod
}

func (p poll) getMajorityVote() types.TalliedVote {
	var result types.TalliedVote

	for i, talliedVote := range p.k.getTalliedVotes(p.ctx, p.ID) {
		if i == 0 || talliedVote.Tally.GT(result.Tally) {
			result = talliedVote
		}
	}

	return result
}
