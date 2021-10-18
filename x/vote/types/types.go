package types

import (
	"crypto/sha256"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

//go:generate moq -out ./mock/types.go -pkg mock . Store

var _ codectypes.UnpackInterfacesMessage = TalliedVote{}

// Voters is a type alias necessary to unmarshal TalliedVote
type Voters []sdk.ValAddress

// NewTalliedVote is the constructor for TalliedVote
func NewTalliedVote(voter sdk.ValAddress, shareCount int64, data codec.ProtoMarshaler) TalliedVote {
	if voter == nil {
		panic("voter cannot be nil")
	}
	d, err := codectypes.NewAnyWithValue(data)
	if err != nil {
		panic(err)
	}

	return TalliedVote{
		Tally:  sdk.NewInt(shareCount),
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
func NewPollMetaData(key exported.PollKey, threshold utils.Threshold, snapshotSeqNo int64) exported.PollMetadata {
	return exported.PollMetadata{
		Key:             key,
		SnapshotSeqNo:   snapshotSeqNo,
		ExpiresAt:       -1,
		Result:          nil,
		VotingThreshold: threshold,
		State:           exported.Pending,
		MinVoterCount:   0,
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
	SetVote(voter sdk.ValAddress, vote TalliedVote)
	GetVote(hash string) (TalliedVote, bool)
	HasVoted(voter sdk.ValAddress) bool
	GetVotes() []TalliedVote
	GetShareCount(voter sdk.ValAddress) (int64, bool)
	GetTotalShareCount() sdk.Int
	SetMetadata(metadata exported.PollMetadata)
	GetPoll(key exported.PollKey) exported.Poll
	DeletePoll()
	GetTotalVoterCount() int64
}

// NewPoll creates a new poll
func NewPoll(meta exported.PollMetadata, currentBlock int64, store Store) *Poll {
	poll := &Poll{
		PollMetadata: meta,
		Store:        store,
		logger:       utils.NewNOPLogger(),
	}

	poll.updateExpiry(currentBlock)
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
	other := p.Store.GetPoll(p.Key)
	if err := other.Delete(); err != nil {
		return err
	}

	p.SetMetadata(p.PollMetadata)
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

	shareCount, ok := p.GetShareCount(voter)
	if !ok {
		return fmt.Errorf("address %s is not eligible to Vote in this poll", voter)
	}

	if p.HasVoted(voter) {
		return fmt.Errorf("voter %s has already voted", voter.String())
	}

	p.SetVote(voter, p.tally(voter, shareCount, data))

	majorityVote := p.getMajorityVote()
	if p.hasEnoughVotes(majorityVote.Tally) {
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

// GetSnapshotSeqNo returns the sequence number of the snapshot associated with the poll
func (p *Poll) GetSnapshotSeqNo() int64 {
	return p.SnapshotSeqNo
}

func (p *Poll) updateExpiry(currentBlockHeight int64) {
	if p.ExpiresAt != -1 && p.ExpiresAt <= currentBlockHeight && p.Is(exported.Pending) {
		p.State |= exported.Expired | exported.AllowOverride
	}
}

func (p *Poll) tally(voter sdk.ValAddress, shareCount int64, data codec.ProtoMarshaler) TalliedVote {
	var talliedVote TalliedVote
	if existingVote, ok := p.GetVote(hash(data)); !ok {
		talliedVote = NewTalliedVote(voter, shareCount, data)
	} else {
		talliedVote = existingVote
		talliedVote.Tally = talliedVote.Tally.AddRaw(shareCount)
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

func (p *Poll) hasEnoughVotes(majorityShare sdk.Int) bool {
	voterCount := p.getVoterCount()

	return utils.NewThreshold(majorityShare.Int64(), p.GetTotalShareCount().Int64()).GTE(p.VotingThreshold) &&
		(p.GetTotalVoterCount() < p.MinVoterCount || voterCount >= p.MinVoterCount)
}

func (p *Poll) cannotWin(majorityShare sdk.Int) bool {
	alreadyTallied := p.getTalliedShareCount()
	missingShares := p.GetTotalShareCount().Sub(alreadyTallied)

	return utils.NewThreshold(majorityShare.Add(missingShares).Int64(), p.GetTotalShareCount().Int64()).LT(p.VotingThreshold)
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

func (p Poll) getTalliedShareCount() sdk.Int {
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
