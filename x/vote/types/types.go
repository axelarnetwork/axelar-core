package types

import (
	"bytes"
	"crypto/sha256"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	reward "github.com/axelarnetwork/axelar-core/x/reward/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

//go:generate moq -out ./mock/types.go -pkg mock . Store

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

// Hash returns the hash of the value of the vote
func (m TalliedVote) Hash() string {
	return hash(m.Data.GetCachedValue().(codec.ProtoMarshaler))
}

// NewPollMetaData is the constructor for PollMetadata.
// It is not in the exported package to make it clear that only the vote module is supposed to use it.
func NewPollMetaData(key exported.PollKey, threshold utils.Threshold, voters []exported.Voter, totalVotingPower sdk.Int) exported.PollMetadata {
	return exported.PollMetadata{
		Key:              key,
		ExpiresAt:        -1,
		Result:           nil,
		VotingThreshold:  threshold,
		State:            exported.Pending,
		MinVoterCount:    0,
		Voters:           voters,
		TotalVotingPower: totalVotingPower,
	}
}

var _ exported.Poll = &Poll{}

// Poll represents a poll with write-in voting
type Poll struct {
	exported.PollMetadata
	Store
	logger     log.Logger
	rewardPool reward.RewardPool
}

// Store enables a poll to communicate with the keeper
type Store interface {
	SetVote(voter sdk.ValAddress, vote TalliedVote)
	GetVote(hash string) (TalliedVote, bool)
	HasVoted(voter sdk.ValAddress) bool
	GetVotes() []TalliedVote
	SetMetadata(metadata exported.PollMetadata)
	GetPoll(key exported.PollKey) exported.Poll
	DeletePoll()
}

// NewPoll creates a new poll
func NewPoll(ctx sdk.Context, meta exported.PollMetadata, store Store, rewarder Rewarder) *Poll {
	poll := &Poll{
		PollMetadata: meta,
		Store:        store,
		logger:       utils.NewNOPLogger(),
	}

	if meta.RewardPoolName != "" {
		poll.rewardPool = rewarder.GetPool(ctx, meta.RewardPoolName)
	}

	poll.updateExpiry(ctx.BlockHeight())
	return poll
}

// WithLogger sets a logger for the poll
func (p *Poll) WithLogger(logger log.Logger) *Poll {
	p.logger = logger
	return p
}

// Is checks if the poll is in the given state
func (p Poll) Is(state exported.PollState) bool {
	// this special case check is needed, because 0 & x == 0 is true for any x
	if state == exported.NonExistent {
		return p.State == exported.NonExistent
	}
	return state&p.State == state
}

// AllowOverride makes it possible to delete the poll, regardless of which state it is in
func (p Poll) AllowOverride() {
	if !p.Is(exported.NonExistent) {
		p.State |= exported.AllowOverride
	}
	p.SetMetadata(p.PollMetadata)
}

// GetResult returns the result of the poll. Returns nil if the poll is not completed.
func (p Poll) GetResult() codec.ProtoMarshaler {
	if p.Result == nil {
		return nil
	}

	return p.Result.GetCachedValue().(codec.ProtoMarshaler)
}

// Initialize initializes the poll
func (p Poll) Initialize() error {
	sumVotingPower := sdk.ZeroInt()
	for _, voter := range p.Voters {
		sumVotingPower = sumVotingPower.AddRaw(voter.VotingPower)
	}

	if utils.NewThreshold(sumVotingPower.Int64(), p.TotalVotingPower.Int64()).LT(p.VotingThreshold) {
		return fmt.Errorf("cannot create poll %s due to it being impossible to pass", p.Key.String())
	}

	other := p.Store.GetPoll(p.Key)
	if err := other.Delete(); err != nil {
		return err
	}

	p.SetMetadata(p.PollMetadata)
	return nil
}

func (p *Poll) getVotingPower(v sdk.ValAddress) int64 {
	for _, voter := range p.PollMetadata.Voters {
		if v.Equals(voter.Validator) {
			return voter.VotingPower
		}
	}

	return 0
}

func (p *Poll) handleRewards() error {
	majorityVote := p.getMajorityVote()

	for _, vote := range p.GetVotes() {
		data, err := vote.Data.Marshal()
		if err != nil {
			panic(err)
		}

		if bytes.Equal(data, majorityVote.Data.Value) {
			for _, voter := range vote.Voters {
				if err := p.rewardPool.ReleaseRewards(voter); err != nil {
					return err
				}
			}
		} else {
			for _, voter := range vote.Voters {
				p.logger.Debug("penalizing voter due to incorrect vote", "voter", voter.String(), "poll", p.PollMetadata.Key.String())
				p.rewardPool.ClearRewards(voter)
			}
		}
	}

	return nil
}

// Vote records the given vote
func (p *Poll) Vote(voter sdk.ValAddress, data codec.ProtoMarshaler) error {
	if p.Is(exported.NonExistent) {
		return fmt.Errorf("poll does not exist")
	}

	// if the poll is already decided there is no need to keep track of further votes
	if p.Is(exported.Completed) || p.Is(exported.Failed) {
		return nil
	}

	votingPower := p.getVotingPower(voter)
	if votingPower == 0 {
		return fmt.Errorf("address %s is not eligible to Vote in this poll", voter)
	}

	if p.HasVoted(voter) {
		return fmt.Errorf("voter %s has already voted", voter.String())
	}

	p.SetVote(voter, p.tally(voter, votingPower, data))

	majorityVote := p.getMajorityVote()
	if p.hasEnoughVotes(majorityVote.Tally) {
		if p.rewardPool != nil {
			p.handleRewards()
		}

		p.Result = majorityVote.Data
		p.State = exported.Completed
		p.logger.Debug(fmt.Sprintf("poll %s (threshold: %d/%d, min vouter count: %d) completed",
			p.Key,
			p.VotingThreshold.Numerator,
			p.VotingThreshold.Denominator,
			p.MinVoterCount,
		))
	} else if p.cannotWin(majorityVote.Tally) {
		p.State = exported.Failed | exported.AllowOverride
		p.logger.Debug(fmt.Sprintf("poll %s (threshold: %d/%d, min vouter count: %d) failed, voters could not agree on single value",
			p.Key,
			p.VotingThreshold.Numerator,
			p.VotingThreshold.Denominator,
			p.MinVoterCount,
		))
	}

	p.SetMetadata(p.PollMetadata)

	return nil
}

// Delete deletes the poll. Returns error if the poll is in a state that does not allow deletion
func (p Poll) Delete() error {
	switch {
	case p.Is(exported.NonExistent):
		return nil
	case p.Is(exported.AllowOverride):
		p.logger.Debug(fmt.Sprintf("deleting poll %s in state %s", p.Key.String(), p.State))
		p.Store.DeletePoll()
		return nil
	default:
		return fmt.Errorf("cannot delete existing poll %s with state %s, must be allowed to be overridden", p.Key, p.State)
	}
}

// GetKey returns the poll's key
func (p *Poll) GetKey() exported.PollKey {
	return p.Key
}

// GetVoters returns the poll's voters
func (p *Poll) GetVoters() []exported.Voter {
	return p.Voters
}

// GetTotalVotingPower returns the total voting power of the poll
func (p *Poll) GetTotalVotingPower() sdk.Int {
	return p.TotalVotingPower
}

func (p *Poll) updateExpiry(currentBlockHeight int64) {
	if p.ExpiresAt != -1 && p.ExpiresAt <= currentBlockHeight && p.Is(exported.Pending) {
		p.State |= exported.Expired | exported.AllowOverride

		if p.rewardPool != nil {
			// Penalize voters who failed to vote
			for _, voter := range p.Voters {
				if !p.HasVoted(voter.Validator) {
					p.logger.Debug("penalizing voter due to timeout", "voter", voter.Validator.String(), "poll", p.PollMetadata.Key.String())
					p.rewardPool.ClearRewards(voter.Validator)
				}
			}
		}
	}
}

func (p *Poll) tally(voter sdk.ValAddress, votingPower int64, data codec.ProtoMarshaler) TalliedVote {
	var talliedVote TalliedVote
	if existingVote, ok := p.GetVote(hash(data)); !ok {
		talliedVote = NewTalliedVote(voter, votingPower, data)
	} else {
		talliedVote = existingVote
		talliedVote.Tally = talliedVote.Tally.AddRaw(votingPower)
		talliedVote.Voters = append(talliedVote.Voters, voter)
	}
	return talliedVote
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
		(int64(len(p.Voters)) < p.MinVoterCount || p.getVoterCount() >= p.MinVoterCount)
}

func (p *Poll) cannotWin(majority sdk.Int) bool {
	alreadyTallied := p.getTalliedVotingPower()
	missingVotingPower := p.GetTotalVotingPower().Sub(alreadyTallied)

	return utils.NewThreshold(majority.Add(missingVotingPower).Int64(), p.GetTotalVotingPower().Int64()).LT(p.VotingThreshold)
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

func hash(data codec.ProtoMarshaler) string {
	bz, err := data.Marshal()
	if err != nil {
		panic(err)
	}
	h := sha256.Sum256(bz)

	return string(h[:])
}
