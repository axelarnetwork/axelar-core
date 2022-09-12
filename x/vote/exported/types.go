package exported

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
)

//go:generate moq -out ./mock/types.go -pkg mock . Poll VoteHandler

var _ codectypes.UnpackInterfacesMessage = PollMetadata{}

// UnpackInterfaces implements UnpackInterfacesMessage
func (m PollMetadata) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var data codec.ProtoMarshaler
	if err := unpacker.UnpackAny(m.Result, &data); err != nil {
		return err
	}

	if err := unpacker.UnpackAny(m.ModuleMetadata, &data); err != nil {
		return err
	}

	return nil
}

// VoteHandler defines a struct that can handle the poll result
type VoteHandler interface {
	IsFalsyResult(result codec.ProtoMarshaler) bool
	HandleExpiredPoll(ctx sdk.Context, poll Poll) error
	HandleFailedPoll(ctx sdk.Context, poll Poll) error
	HandleCompletedPoll(ctx sdk.Context, poll Poll) error
	HandleResult(ctx sdk.Context, result codec.ProtoMarshaler) error
}

// PollID represents ID of polls
type PollID uint64

// String converts the given poll ID to string
func (id PollID) String() string {
	return strconv.FormatUint(uint64(id), 10)
}

// Deprecated: String converts the given poll key to string
func (m PollKey) String() string {
	return fmt.Sprintf("%s_%s", m.Module, m.ID)
}

// PollBuilder is a builder that is used to build up the poll metadata
type PollBuilder struct {
	p PollMetadata
}

// NewPollBuilder is the constructor for the poll builder
func NewPollBuilder(module string, threshold utils.Threshold, snapshot snapshot.Snapshot, expiresAt int64) PollBuilder {
	return PollBuilder{
		p: PollMetadata{
			State:           Pending,
			Module:          module,
			VotingThreshold: threshold,
			Snapshot:        snapshot,
			ExpiresAt:       expiresAt,
		},
	}
}

// ID sets the poll ID
func (builder PollBuilder) ID(pollID PollID) PollBuilder {
	builder.p.ID = pollID
	return builder
}

// MinVoterCount sets the minimum number of voters that have to vote on PollMeta
// If not enough voters exist, then all of them have to vote
func (builder PollBuilder) MinVoterCount(minVoterCount int64) PollBuilder {
	builder.p.MinVoterCount = minVoterCount
	return builder
}

// RewardPoolName sets the name of a reward pool for the poll
func (builder PollBuilder) RewardPoolName(rewardPoolName string) PollBuilder {
	builder.p.RewardPoolName = rewardPoolName
	return builder
}

// GracePeriod sets the grace period after poll completion during which votes
// are still recorded
func (builder PollBuilder) GracePeriod(gracePeriod int64) PollBuilder {
	builder.p.GracePeriod = gracePeriod
	return builder
}

// ModuleMetadata sets the module metadata on the poll
func (builder PollBuilder) ModuleMetadata(moduleMetadata codec.ProtoMarshaler) PollBuilder {
	any, err := codectypes.NewAnyWithValue(moduleMetadata)
	if err != nil {
		panic(err)
	}

	builder.p.ModuleMetadata = any
	return builder
}

// Build returns the wrapped poll metadata, or an error if the poll metadata is not valid
func (builder PollBuilder) Build(blockHeight int64) (PollMetadata, error) {
	p := builder.p

	if err := p.ValidateBasic(); err != nil {
		return PollMetadata{}, err
	}

	if p.ExpiresAt <= blockHeight {
		return PollMetadata{}, fmt.Errorf(
			"cannot create poll that expires at block %d which is less than or equal to the current block height %d",
			p.ExpiresAt,
			blockHeight,
		)
	}

	if !p.Is(Pending) {
		return PollMetadata{}, fmt.Errorf("cannot create poll %s that is not pending", p.ID)
	}

	if p.CompletedAt != 0 {
		return PollMetadata{}, fmt.Errorf("cannot create poll %s that is already completed", p.ID)
	}

	return p, nil
}

// ValidateBasic returns an error if the poll metadata is not valid; nil otherwise
func (m PollMetadata) ValidateBasic() error {
	if len(m.Module) == 0 {
		return fmt.Errorf("module must be set")
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
		return fmt.Errorf("completed poll must have CompletedAt set and non-completed poll must not")
	}

	if m.Is(NonExistent) {
		return fmt.Errorf("state cannot be non-existent")
	}

	if m.MinVoterCount < 0 || m.MinVoterCount > int64(len(m.Snapshot.Participants)) {
		return fmt.Errorf("invalid min voter count")
	}

	if err := m.Snapshot.ValidateBasic(); err != nil {
		return err
	}

	if m.Snapshot.GetParticipantsWeight().LT(m.Snapshot.CalculateMinPassingWeight(m.VotingThreshold)) {
		return fmt.Errorf("invalid voting threshold")
	}

	return nil
}

// Is returns true if the poll metadata is in the given state, false otherwise
func (m PollMetadata) Is(state PollState) bool {
	return m.State == state
}

// VoteResult represents all possible results of vote
type VoteResult int

const (
	// NoVote means the voter is not allowed to vote for the poll anymore
	NoVote = iota
	// VoteInTime means the voter successfully voted for the poll before it completes
	VoteInTime
	// VotedLate means the voter successfully voted for the poll after it completes but within the grace period
	VotedLate
)

// Poll provides an interface for other modules to interact with polls
type Poll interface {
	GetID() PollID
	GetState() PollState
	HasVotedCorrectly(voter sdk.ValAddress) bool
	HasVoted(voter sdk.ValAddress) bool
	GetResult() codec.ProtoMarshaler
	GetRewardPoolName() (string, bool)
	GetVoters() []sdk.ValAddress
	Vote(voter sdk.ValAddress, blockHeight int64, data codec.ProtoMarshaler) (VoteResult, error)
	GetModule() string
	GetMetaData() (codec.ProtoMarshaler, bool)
}
