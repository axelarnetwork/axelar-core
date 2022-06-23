package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/axelarnetwork/utils/slices"
)

//go:generate moq -out ./mock/types.go -pkg mock . Store VoteRouter

var _ codectypes.UnpackInterfacesMessage = TalliedVote{}

// Voters is a type alias necessary to unmarshal TalliedVote
type Voters []sdk.ValAddress

// NewTalliedVote is the constructor for TalliedVote
func NewTalliedVote(voter sdk.ValAddress, votingPower int64, data codec.ProtoMarshaler) TalliedVote {
	if voter == nil {
		panic("voter cannot be nil")
	}
	d, err := codectypes.NewAnyWithValue(data)
	if err != nil {
		panic(err)
	}

	return TalliedVote{
		Tally:  sdk.NewInt(votingPower),
		Data:   d,
		Voters: Voters{voter},
	}
}

// UnpackInterfaces implements UnpackInterfacesMessage
func (m TalliedVote) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var data codec.ProtoMarshaler
	return unpacker.UnpackAny(m.Data, &data)
}

// NewPollMetaData is the constructor for PollMetadata.
// It is not in the exported package to make it clear that only the vote module is supposed to use it.
func NewPollMetaData(id exported.PollID, threshold utils.Threshold, voters []exported.Voter) exported.PollMetadata {
	return exported.PollMetadata{
		ID:              id,
		ExpiresAt:       0,
		Result:          nil,
		VotingThreshold: threshold,
		State:           exported.Pending,
		MinVoterCount:   0,
		Voters:          voters,
		TotalVotingPower: slices.Reduce(voters, sdk.ZeroInt(), func(total sdk.Int, voter exported.Voter) sdk.Int {
			return total.AddRaw(voter.VotingPower)
		}),
		GracePeriod: 0,
		CompletedAt: 0,
	}
}

var _ exported.Poll = &Poll{}

// Poll represents a poll with write-in voting
type Poll struct {
	exported.PollMetadata
	Store
	logger log.Logger
}

// Store enables a poll to communicate with the keeper
type Store interface {
	SetVote(voter sdk.ValAddress, data codec.ProtoMarshaler, votingPower int64, isLate bool)
	GetVote(hash string) (TalliedVote, bool)
	HasVoted(voter sdk.ValAddress) bool
	HasVotedLate(voter sdk.ValAddress) bool
	GetVotes() []TalliedVote
	SetMetadata(metadata exported.PollMetadata)
	EnqueuePoll(metadata exported.PollMetadata)
	GetPoll(id exported.PollID) exported.Poll
	DeletePoll()
}

// NewPoll creates a new poll
func NewPoll(meta exported.PollMetadata, store Store) *Poll {
	return &Poll{
		PollMetadata: meta,
		Store:        store,
		logger:       utils.NewNOPLogger(),
	}
}

// WithLogger sets a logger for the poll
func (p *Poll) WithLogger(logger log.Logger) *Poll {
	p.logger = logger
	return p
}

// Is checks if the poll is in the given state
func (p Poll) Is(state exported.PollState) bool {
	return p.PollMetadata.Is(state)
}

// GetResult returns the result of the poll. Returns nil if the poll is not completed.
func (p Poll) GetResult() codec.ProtoMarshaler {
	if p.Result == nil {
		return nil
	}

	return p.Result.GetCachedValue().(codec.ProtoMarshaler)
}

// GetModuleMetadata returns the module metadata
func (p Poll) GetModuleMetadata() exported.PollModuleMetadata {
	return p.ModuleMetadata
}

// Initialize initializes the poll
func (p Poll) Initialize(blockHeight int64) error {
	if err := p.Validate(); err != nil {
		return err
	}

	if p.ExpiresAt <= blockHeight {
		return fmt.Errorf(
			"cannot create poll %s that expires at block %d which is less than or equal to the current block height",
			p.ID,
			p.ExpiresAt,
		)
	}

	if !p.Is(exported.Pending) {
		return fmt.Errorf("cannot create poll %s that is not pending", p.ID)
	}

	if p.CompletedAt != 0 {
		return fmt.Errorf("cannot create poll %s that is already completed", p.ID)
	}

	if !p.Store.GetPoll(p.ID).Is(exported.NonExistent) {
		return fmt.Errorf("poll with ID %s already exists", p.ID.String())
	}

	p.EnqueuePoll(p.PollMetadata)

	return nil
}

func (p Poll) getVotingPower(v sdk.ValAddress) int64 {
	for _, voter := range p.PollMetadata.Voters {
		if v.Equals(voter.Validator) {
			return voter.VotingPower
		}
	}

	return 0
}

// Vote records the given vote
func (p *Poll) Vote(voter sdk.ValAddress, blockHeight int64, data codec.ProtoMarshaler) (codec.ProtoMarshaler, bool, error) {
	if p.Is(exported.NonExistent) {
		return nil, false, fmt.Errorf("poll does not exist")
	}

	if p.HasVoted(voter) {
		return nil, false, fmt.Errorf("voter %s has voted already", voter)
	}

	votingPower := p.getVotingPower(voter)
	if votingPower == 0 {
		return nil, false, fmt.Errorf("address %s is not eligible to Vote in this poll", voter)
	}

	switch {
	case p.Is(exported.Failed):
		return nil, false, nil
	case p.Is(exported.Completed) && blockHeight <= p.CompletedAt+int64(p.GracePeriod):
		// poll is completed but within the grace period for late votes
		p.SetVote(voter, data, votingPower, true)
		p.logger.Debug("received late vote for poll",
			"voter", voter.String(),
			"poll", p.GetID().String(),
		)

		return nil, true, nil
	case p.Is(exported.Completed):
		return nil, false, nil
	}

	p.SetVote(voter, data, votingPower, false)
	p.logger.Debug("received vote for poll",
		"voter", voter.String(),
		"poll", p.GetID().String(),
	)

	majorityVote := p.getMajorityVote()
	switch {
	case p.hasEnoughVotes(majorityVote.Tally):
		p.Result = majorityVote.Data
		p.State = exported.Completed
		p.CompletedAt = blockHeight
		p.logger.Debug(fmt.Sprintf("poll %s (threshold: %d/%d, min voter count: %d) completed",
			p.ID,
			p.VotingThreshold.Numerator,
			p.VotingThreshold.Denominator,
			p.MinVoterCount,
		))

		p.SetMetadata(p.PollMetadata)

		return p.GetResult(), true, nil
	case p.cannotWin(majorityVote.Tally):
		p.State = exported.Failed
		p.logger.Debug(fmt.Sprintf("poll %s (threshold: %d/%d, min voter count: %d) failed, voters could not agree on single value",
			p.ID,
			p.VotingThreshold.Numerator,
			p.VotingThreshold.Denominator,
			p.MinVoterCount,
		))

		p.SetMetadata(p.PollMetadata)
		fallthrough
	default:
		return nil, true, nil
	}
}

// Delete deletes the poll
func (p Poll) Delete() {
	if p.Is(exported.NonExistent) {
		return
	}

	p.logger.Debug(fmt.Sprintf("deleting poll %s in state %s", p.ID.String(), p.State))
	p.Store.DeletePoll()
}

// GetID returns the poll's ID
func (p Poll) GetID() exported.PollID {
	return p.ID
}

// GetVoters returns the poll's voters
func (p Poll) GetVoters() []exported.Voter {
	return p.Voters
}

// HasVotedCorrectly returns true if the give voter has voted correctly for the poll, false otherwise
func (p Poll) HasVotedCorrectly(voter sdk.ValAddress) bool {
	majorityVote := p.getMajorityVote()

	return p.Is(exported.Completed) &&
		p.HasVoted(voter) &&
		slices.Any(majorityVote.Voters, func(v sdk.ValAddress) bool {
			return voter.Equals(v)
		})
}

// GetTotalVotingPower returns the total voting power of the poll
func (p Poll) GetTotalVotingPower() sdk.Int {
	return p.TotalVotingPower
}

// GetRewardPoolName returns the name of the attached reward pool, if any
func (p Poll) GetRewardPoolName() (string, bool) {
	return p.RewardPoolName, len(p.RewardPoolName) > 0
}

func (p Poll) getVoterCount() int64 {
	var count int64

	for _, talliedVote := range p.GetVotes() {
		count += int64(len(talliedVote.Voters))
	}

	return count
}

func (p *Poll) hasEnoughVotes(majority sdk.Int) bool {
	return utils.NewThreshold(majority.Int64(), p.GetTotalVotingPower().Int64()).GTE(p.VotingThreshold) &&
		p.getVoterCount() >= p.MinVoterCount
}

func (p *Poll) cannotWin(majority sdk.Int) bool {
	alreadyTallied := p.getTalliedVotingPower()
	missingVotingPower := p.GetTotalVotingPower().Sub(alreadyTallied)

	return !p.hasEnoughVotes(majority.Add(missingVotingPower))
}

func (p Poll) getMajorityVote() TalliedVote {
	var result TalliedVote

	for i, talliedVote := range p.GetVotes() {
		if i == 0 || talliedVote.Tally.GT(result.Tally) {
			result = talliedVote
		}
	}

	return result
}

func (p Poll) getTalliedVotingPower() sdk.Int {
	result := sdk.ZeroInt()

	for _, talliedVote := range p.GetVotes() {
		result = result.Add(talliedVote.Tally)
	}

	return result
}
