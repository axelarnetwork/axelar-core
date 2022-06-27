package exported

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogo "github.com/gogo/protobuf/types"

	"github.com/axelarnetwork/axelar-core/utils"
)

//go:generate moq -out ./mock/types.go -pkg mock . Poll VoteHandler

// VoteHandler defines a struct that can handle the poll result
type VoteHandler interface {
	IsFalsyResult(result codec.ProtoMarshaler) bool
	HandleExpiredPoll(ctx sdk.Context, poll Poll) error
	HandleCompletedPoll(ctx sdk.Context, poll Poll) error
	HandleResult(ctx sdk.Context, result codec.ProtoMarshaler) error
}

// PollID represents ID of polls
type PollID uint64

// String converts the given poll ID to string
func (id PollID) String() string {
	return strconv.FormatUint(uint64(id), 10)
}

func (m PollKey) String() string {
	return fmt.Sprintf("%s_%s", m.Module, m.ID)
}

type PollBuilder struct {
	ExpiresAt       int64
	VotingThreshold utils.Threshold
	MinVoterCount   int64
	RewardPoolName  string
	GracePeriod     int64
	Module          string
	Metadata        gogo.Any
}

// PollProperty is a modifier for PollMetadata. It should never be manually initialized
type PollProperty struct {
	do func(metadata PollBuilder) PollBuilder
}

func (p PollProperty) apply(metadata PollBuilder) PollBuilder {
	return p.do(metadata)
}

// With returns a new metadata object with all the given properties set
func (b PollBuilder) With(properties ...PollProperty) PollBuilder {
	builder := b
	for _, property := range properties {
		builder = property.apply(builder)
	}
	return builder
}

// ExpiryAt sets the expiry property on PollMetadata
func ExpiryAt(blockHeight int64) PollProperty {
	return PollProperty{do: func(builder PollBuilder) PollBuilder {
		builder.ExpiresAt = blockHeight
		return builder
	}}
}

// Threshold sets the threshold property on PollMetadata
func Threshold(threshold utils.Threshold) PollProperty {
	return PollProperty{do: func(builder PollBuilder) PollBuilder {
		builder.VotingThreshold = threshold
		return builder
	}}
}

// MinVoterCount sets the minimum number of voters that have to vote on PollMeta
// If not enough voters exist, then all of them have to vote
func MinVoterCount(minVoterCount int64) PollProperty {
	return PollProperty{do: func(builder PollBuilder) PollBuilder {
		builder.MinVoterCount = minVoterCount
		return builder
	}}
}

// RewardPool sets the name of a reward pool for the poll
func RewardPool(rewardPoolName string) PollProperty {
	return PollProperty{do: func(builder PollBuilder) PollBuilder {
		builder.RewardPoolName = rewardPoolName
		return builder
	}}
}

// GracePeriod sets the grace period after poll completion during which votes
// are still recorded
func GracePeriod(gracePeriod int64) PollProperty {
	return PollProperty{do: func(builder PollBuilder) PollBuilder {
		builder.GracePeriod = gracePeriod
		return builder
	}}
}

// Module sets the module name on the poll
func Module(module string) PollProperty {
	return PollProperty{do: func(builder PollBuilder) PollBuilder {
		builder.Module = module

		return builder
	}}
}

// ValidateBasic returns an error if the poll module metadata is not valid; nil otherwise
func (m PollModuleMetadata) ValidateBasic() error {
	if err := utils.ValidateString(m.Module); err != nil {
		return err
	}

	return nil
}

// Poll provides an interface for other modules to interact with polls
type Poll interface {
	HasVotedCorrectly(voter sdk.ValAddress) bool
	HasVoted(voter sdk.ValAddress) bool
	GetResult() codec.ProtoMarshaler
	GetRewardPoolName() (string, bool)
	GetVoters() []sdk.ValAddress
}
