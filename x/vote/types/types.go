package types

import (
	"bytes"
	"fmt"
	"sort"

	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"
	"golang.org/x/exp/maps"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/axelarnetwork/utils/slices"
)

//go:generate moq -out ./mock/types.go -pkg mock . Store VoteRouter

var _ codectypes.UnpackInterfacesMessage = TalliedVote{}

// Voters is a type alias necessary to unmarshal TalliedVote
type Voters []sdk.ValAddress

// NewTalliedVote is the constructor for TalliedVote
func NewTalliedVote(pollID exported.PollID, data codec.ProtoMarshaler) TalliedVote {
	d, err := codectypes.NewAnyWithValue(data)
	if err != nil {
		panic(err)
	}

	return TalliedVote{
		PollID: pollID,
		Tally:  sdk.ZeroUint(),
		Data:   d,
	}
}

// TallyVote adds the given voting power to the tallied vote
func (m *TalliedVote) TallyVote(voter sdk.ValAddress, votingPower sdk.Uint) {
	if voter == nil {
		panic("voter cannot be nil")
	}
	m.Voters = append(m.Voters, voter)
	m.Tally = m.Tally.Add(votingPower)
}

// UnpackInterfaces implements UnpackInterfacesMessage
func (m TalliedVote) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var data codec.ProtoMarshaler
	return unpacker.UnpackAny(m.Data, &data)
}

// NewPollMetaData is the constructor for PollMetadata.
// It is not in the exported package to make it clear that only the vote module is supposed to use it.
func NewPollMetaData(id exported.PollID, threshold utils.Threshold, snapshot snapshot.Snapshot) PollMetadata {
	return PollMetadata{
		VotingThreshold: threshold,
		State:           exported.Pending,
		ID:              id,
		Snapshot:        snapshot,
	}
}

var _ exported.Poll = &Poll{}

// Poll represents a poll with write-in voting
type Poll struct {
	PollMetadata
	Store
	logger log.Logger
}

// Store enables a poll to communicate with the keeper
type Store interface {
	SetVote(voter sdk.ValAddress, data codec.ProtoMarshaler, votingPower sdk.Uint, isLate bool)
	GetVote(hash string) (TalliedVote, bool)
	HasVoted(voter sdk.ValAddress) bool
	HasVotedLate(voter sdk.ValAddress) bool
	GetVotes() []TalliedVote
	SetMetadata(metadata PollMetadata)
	EnqueuePoll(metadata PollMetadata)
	GetPoll(id exported.PollID) Poll
	DeletePoll()
}

// NewPoll creates a new poll
func NewPoll(meta PollMetadata, store Store) *Poll {
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
	if err := p.ValidateBasic(); err != nil {
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

type VoteResult int

const (
	NoVote = iota
	VoteInTime
	VotedLate
)

// Vote records the given vote
func (p *Poll) Vote(voter sdk.ValAddress, blockHeight int64, data codec.ProtoMarshaler) (VoteResult, error) {
	if p.Is(exported.NonExistent) {
		return NoVote, fmt.Errorf("poll does not exist")
	}

	if p.HasVoted(voter) {
		return NoVote, fmt.Errorf("voter %s has voted already", voter)
	}

	if p.getVotingPower(voter).IsZero() {
		return NoVote, fmt.Errorf("address %s is not eligible to Vote in this poll", voter)
	}

	if p.Is(exported.Failed) {
		return NoVote, nil
	}

	if p.Is(exported.Completed) && p.isInGracePeriod(blockHeight) {
		p.voteLate(voter, data)
		return VotedLate, nil
	}

	if p.Is(exported.Completed) {
		return NoVote, nil
	}

	p.voteBeforeCompletion(voter, blockHeight, data)
	return VoteInTime, nil
}

func (p *Poll) isInGracePeriod(blockHeight int64) bool {
	return blockHeight <= p.CompletedAt+p.GracePeriod
}

func (p *Poll) voteBeforeCompletion(voter sdk.ValAddress, blockHeight int64, data codec.ProtoMarshaler) {
	p.logger.Debug("received vote for poll",
		"voter", voter.String(),
		"poll", p.GetID().String(),
	)

	p.SetVote(voter, data, p.getVotingPower(voter), false)

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

	case p.cannotWin(majorityVote.Tally):
		p.State = exported.Failed
		p.logger.Debug(fmt.Sprintf("poll %s (threshold: %d/%d, min voter count: %d) failed, voters could not agree on single value",
			p.ID,
			p.VotingThreshold.Numerator,
			p.VotingThreshold.Denominator,
			p.MinVoterCount,
		))

		p.SetMetadata(p.PollMetadata)
	}
}

func (p *Poll) voteLate(voter sdk.ValAddress, data codec.ProtoMarshaler) {
	p.logger.Debug("received late vote for poll",
		"voter", voter.String(),
		"poll", p.GetID().String(),
	)

	p.SetVote(voter, data, p.Snapshot.GetParticipantWeight(voter), true)
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
func (p Poll) GetVoters() []sdk.ValAddress {
	voters := slices.Map(maps.Values(p.Snapshot.Participants), snapshot.Participant.GetAddress)
	sort.SliceStable(voters, func(i, j int) bool { return bytes.Compare(voters[i], voters[j]) < 0 })
	return voters
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
func (p Poll) GetTotalVotingPower() sdk.Uint {
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

func (p *Poll) hasEnoughVotes(majority sdk.Uint) bool {
	return utils.NewThreshold(majority.Int64(), p.GetTotalVotingPower().Int64()).GTE(p.VotingThreshold) &&
		p.getVoterCount() >= p.MinVoterCount
}

func (p *Poll) cannotWin(majority sdk.Uint) bool {
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

func (p Poll) getTalliedVotingPower() sdk.Uint {
	result := sdk.ZeroUint()

	for _, talliedVote := range p.GetVotes() {
		result = result.Add(talliedVote.Tally)
	}

	return result
}

// Is checks if the poll is in the given state
func (m PollMetadata) Is(state exported.PollState) bool {
	return m.State == state
}

// ValidateBasic returns an error if the poll metadata is not valid; nil otherwise
func (m PollMetadata) ValidateBasic() error {
	if err := m.ModuleMetadata.ValidateBasic(); err != nil {
		return err
	}

	if m.ExpiresAt <= 0 {
		return fmt.Errorf("expires at must be >0")
	}

	if m.CompletedAt < 0 {
		return fmt.Errorf("completed at must be >=0")
	}

	if m.VotingThreshold.LTE(utils.ZeroThreshold) || m.VotingThreshold.GT(utils.OneThreshold) {
		return fmt.Errorf("voting threshold must be >0 and <=1")
	}

	if m.Is(Completed) == (m.Result == nil) {
		return fmt.Errorf("completed poll must have result set")
	}

	if m.Is(Completed) == (m.CompletedAt <= 0) {
		return fmt.Errorf("completed poll must have completed at set and non-completed poll must not")
	}

	if m.Is(NonExistent) {
		return fmt.Errorf("state cannot be non-existent")
	}

	if m.MinVoterCount < 0 || m.MinVoterCount > int64(len(m.Voters)) {
		return fmt.Errorf("invalid min voter count")
	}

	if len(m.Voters) == 0 {
		return fmt.Errorf("no voters set")
	}

	actualTotalVotingPower := sdk.ZeroInt()
	for _, voter := range m.Voters {
		if err := sdk.VerifyAddressFormat(voter.Validator); err != nil {
			return err
		}

		if voter.VotingPower <= 0 {
			return fmt.Errorf("voter's voting power must be >0")
		}

		actualTotalVotingPower = actualTotalVotingPower.AddRaw(voter.VotingPower)
	}

	if !m.TotalVotingPower.Equal(actualTotalVotingPower) {
		return fmt.Errorf("total voting power mismatch")
	}

	return nil
}

var _ codectypes.UnpackInterfacesMessage = PollMetadata{}

// UnpackInterfaces implements UnpackInterfacesMessage
func (m PollMetadata) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var data codec.ProtoMarshaler
	return unpacker.UnpackAny(m.Result, &data)
}
