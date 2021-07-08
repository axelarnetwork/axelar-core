package types

import (
	"crypto/sha256"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

var _ types.UnpackInterfacesMessage = TalliedVote{}
var _ types.UnpackInterfacesMessage = PollMetadata{}

// NewTalliedVote is the constructor for TalliedVote
func NewTalliedVote(voter snapshot.Validator, data codec.ProtoMarshaler) TalliedVote {
	d, err := codectypes.NewAnyWithValue(data)
	if err != nil {
		panic(err)
	}

	return TalliedVote{
		Tally:  sdk.NewInt(voter.ShareCount),
		Data:   d,
		Voters: Voters{voter.GetSDKValidator().GetOperator()},
	}
}

// UnpackInterfaces implements UnpackInterfacesMessage
func (m TalliedVote) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var data codec.ProtoMarshaler
	return unpacker.UnpackAny(m.Data, &data)
}

// NewPollMetaData is the constructor for PollMetadata
func NewPollMetaData(key exported.PollKey, snapshotSeqNo int64, expiresAt int64, threshold utils.Threshold) PollMetadata {
	return PollMetadata{
		Key:             key,
		SnapshotSeqNo:   snapshotSeqNo,
		ExpiresAt:       expiresAt,
		Result:          nil,
		VotingThreshold: threshold,
		State:           Pending,
	}
}

func (p PollMetadata) Is(state PollMetadata_State) bool {
	return state == p.State
}

func (p PollMetadata) UpdateBlockHeight(height int64) PollMetadata {
	if p.ExpiresAt <= height {
		p.State = Expired
	}
	return p
}

func (m PollMetadata) GetResult() codec.ProtoMarshaler {
	if m.Result == nil {
		return nil
	}

	return m.Result.GetCachedValue().(codec.ProtoMarshaler)
}

// UnpackInterfaces implements UnpackInterfacesMessage
func (m PollMetadata) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var data codec.ProtoMarshaler
	return unpacker.UnpackAny(m.Result, &data)
}

type Poll struct {
	PollMetadata
	mappedVotes map[string]int
	Votes       []FlaggedVote
	snapshot    snapshot.Snapshot
	voters      map[string]struct{}
}

type Voters []sdk.ValAddress
type FlaggedVote struct {
	Vote  TalliedVote
	Dirty bool
}

func NewPoll(meta PollMetadata, votes []TalliedVote, snapshot snapshot.Snapshot) Poll {
	mappedVotes := make(map[string]int)
	flaggedVotes := make([]FlaggedVote, 0, len(votes))
	voters := make(map[string]struct{})
	for i, vote := range votes {
		mappedVotes[hash(vote.Data.GetCachedValue().(codec.ProtoMarshaler))] = i
		flaggedVotes = append(flaggedVotes, FlaggedVote{Vote: vote, Dirty: false})
		for _, voter := range vote.Voters {
			voters[voter.String()] = struct{}{}
		}
	}

	return Poll{
		PollMetadata: meta,
		Votes:        flaggedVotes,
		mappedVotes:  mappedVotes,
		snapshot:     snapshot,
		voters:       voters,
	}
}

func (p *Poll) Vote(voterAddr sdk.ValAddress, data codec.ProtoMarshaler) error {
	voter, ok := p.snapshot.GetValidator(voterAddr)
	if !ok {
		return fmt.Errorf("address %s is not eligible to Vote in this poll", voter.String())
	}

	if p.hasVoted(voterAddr) {
		return fmt.Errorf("each validator can only Vote once")
	}

	p.rememberHasVoted(voter)

	var talliedVote TalliedVote
	// check if others match this vote, create a new unique entry if not, simply add voting power if match is found
	if voteIdx, ok := p.matchesExistingVote(data); !ok {
		talliedVote = NewTalliedVote(voter, data)

		p.Votes = append(p.Votes, FlaggedVote{Vote: talliedVote, Dirty: true})
		p.mappedVotes[hash(data)] = len(p.Votes) - 1
	} else {
		// this assignment copies the value, so we need to write it back into the array
		talliedVote = p.Votes[voteIdx].Vote
		talliedVote.Tally = talliedVote.Tally.AddRaw(voter.ShareCount)
		p.Votes[voteIdx] = FlaggedVote{Vote: talliedVote, Dirty: true}
	}

	if p.VotingThreshold.IsMet(talliedVote.Tally, p.snapshot.TotalShareCount) {
		p.Result = talliedVote.Data
		p.State = Completed
	} else if !p.VotingThreshold.IsMet(p.maxPossibleVoteShareCount(), p.snapshot.TotalShareCount) {
		p.State = Failed
	}

	return nil
}

func (p Poll) getTalliedShareCount() sdk.Int {
	result := sdk.ZeroInt()

	for _, talliedVote := range p.Votes {
		result = result.Add(talliedVote.Vote.Tally)
	}

	return result
}

func (p Poll) getMajorityVoteShareCount() sdk.Int {
	var result TalliedVote
	found := false

	for i, talliedVote := range p.Votes {
		if i == 0 || talliedVote.Vote.Tally.GT(result.Tally) {
			found = true
			result = talliedVote.Vote
		}
	}

	if found {
		return result.Tally
	}
	return sdk.ZeroInt()
}

func (p Poll) hasVoted(address sdk.ValAddress) bool {
	_, ok := p.voters[address.String()]
	return ok
}

func (p Poll) maxPossibleVoteShareCount() sdk.Int {
	majorityShare := p.getMajorityVoteShareCount()
	missingShares := p.snapshot.TotalShareCount.Sub(p.getTalliedShareCount())
	return majorityShare.Add(missingShares)
}

func (p Poll) matchesExistingVote(data codec.ProtoMarshaler) (int, bool) {
	h := hash(data)
	voteIdx, ok := p.mappedVotes[h]
	return voteIdx, ok
}

func (p Poll) rememberHasVoted(voter snapshot.Validator) {
	p.voters[voter.GetSDKValidator().GetOperator().String()] = struct{}{}
}

func hash(data codec.ProtoMarshaler) string {
	bz, err := data.Marshal()
	if err != nil {
		panic(err)
	}
	h := sha256.Sum256(bz)

	return string(h[:])
}
