package types

import (
	"crypto/sha256"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

//go:generate moq -out ./mock/types.go -pkg mock . Store

var _ types.UnpackInterfacesMessage = TalliedVote{}

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

func (m TalliedVote) Hash() string {
	return hash(m.Data.GetCachedValue().(codec.ProtoMarshaler))
}

var _ exported.Poll = &Poll{}

type Poll struct {
	exported.PollMetadata
	store Store
}

func (p *Poll) GetMetadata() exported.PollMetadata {
	return p.PollMetadata
}

type Store interface {
	SetVote(key exported.PollKey, vote TalliedVote)
	GetVote(key exported.PollKey, hash string) (TalliedVote, bool)
	GetVotes(key exported.PollKey) []TalliedVote
	SetVoted(key exported.PollKey, voter sdk.ValAddress)
	HasVoted(key exported.PollKey, voter sdk.ValAddress) bool
	GetShareCount(snapSeqNo int64, address sdk.ValAddress) (int64, bool)
	GetTotalShareCount(snapSeqNo int64) sdk.Int
	SetMetadata(metadata exported.PollMetadata)
	GetPoll(key exported.PollKey) exported.Poll
	DeletePoll(key exported.PollKey)
}

func NewPoll(meta exported.PollMetadata, store Store) *Poll {
	return &Poll{
		PollMetadata: meta,
		store:        store,
	}
}
func NewPollWithLogging(meta exported.PollMetadata, store Store, logger log.Logger) *PollWithLogging {
	return &PollWithLogging{
		Poll:   *NewPoll(meta, store),
		logger: logger,
	}
}

func (p Poll) Initialize() error {
	other := p.store.GetPoll(p.Key)

	switch {
	case other.Is(exported.Pending):
		return fmt.Errorf("poll %s already exists and has not expired yet", p.Key.String())
	case other.Is(exported.Completed):
		return fmt.Errorf("poll %s already exists and has a result", p.Key.String())
	case other.Is(exported.Failed) || other.Is(exported.Expired):
		other.Delete()
	}

	p.store.SetMetadata(p.PollMetadata)
	return nil
}

func (p *Poll) Vote(voter sdk.ValAddress, data codec.ProtoMarshaler) error {
	if p.Is(exported.NonExistent) {
		return fmt.Errorf("poll does not exist")
	}

	// if the poll is already decided there is no need to keep track of further votes
	if p.Is(exported.Completed) || p.Is(exported.Failed) {
		return nil
	}

	shareCount, ok := p.store.GetShareCount(p.SnapshotSeqNo, voter)
	if !ok {
		return fmt.Errorf("address %s is not eligible to Vote in this poll", voter)
	}

	if p.store.HasVoted(p.Key, voter) {
		return fmt.Errorf("voter %s has already voted", voter.String())
	}

	p.store.SetVoted(p.Key, voter)

	var talliedVote TalliedVote
	if existingVote, ok := p.store.GetVote(p.Key, hash(data)); !ok {
		talliedVote = NewTalliedVote(voter, shareCount, data)
		p.store.SetVote(p.Key, talliedVote)
	} else {
		talliedVote = existingVote
		talliedVote.Tally = talliedVote.Tally.AddRaw(shareCount)
		p.store.SetVote(p.Key, talliedVote)
	}

	if p.hasEnoughVotes(talliedVote) {
		p.Result = talliedVote.Data
		p.State = exported.Completed
	} else if p.cannotComplete() {
		p.State = exported.Failed
	}

	p.store.SetMetadata(p.PollMetadata)

	return nil
}

func (p Poll) Delete() {
	p.store.DeletePoll(p.Key)
}

func (p *Poll) hasEnoughVotes(talliedVote TalliedVote) bool {
	return p.VotingThreshold.IsMet(talliedVote.Tally, p.store.GetTotalShareCount(p.SnapshotSeqNo))
}

func (p *Poll) cannotComplete() bool {
	majorityShare := p.getMajorityVoteShareCount()
	totalShares := p.store.GetTotalShareCount(p.SnapshotSeqNo)
	alreadyTallied := p.getTalliedShareCount()
	missingShares := totalShares.Sub(alreadyTallied)

	return !p.VotingThreshold.IsMet(majorityShare.Add(missingShares), totalShares)
}

func (p Poll) getMajorityVoteShareCount() sdk.Int {
	var result TalliedVote

	for i, talliedVote := range p.store.GetVotes(p.Key) {
		if i == 0 || talliedVote.Tally.GT(result.Tally) {
			result = talliedVote
		}
	}

	return result.Tally
}

func (p Poll) getTalliedShareCount() sdk.Int {
	result := sdk.ZeroInt()

	for _, talliedVote := range p.store.GetVotes(p.Key) {
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

type PollWithLogging struct {
	Poll
	logger log.Logger
}

func (p *PollWithLogging) Vote(voter sdk.ValAddress, data codec.ProtoMarshaler) error {
	if err := p.Poll.Vote(voter, data); err != nil {
		return err
	}

	switch {
	case p.Is(exported.Completed):
		p.logger.Debug(fmt.Sprintf("poll %s (threshold: %d/%d) completed", p.Key,
			p.VotingThreshold.Numerator, p.VotingThreshold.Denominator))
	case p.Is(exported.Failed):
		p.logger.Debug(fmt.Sprintf("poll %s (threshold: %d/%d) failed, voters could not agree on single value", p.Key,
			p.VotingThreshold.Numerator, p.VotingThreshold.Denominator))
	}
	return nil
}

func (p PollWithLogging) Delete() {
	if p.Is(exported.Failed) || p.Is(exported.Expired) {
		p.logger.Debug(fmt.Sprintf("deleting poll %s due to expiry or failure", p.Key.String()))
		p.Poll.Delete()
	}
}
