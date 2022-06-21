package exported

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

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

// Is checks if the poll is in the given state
func (m PollMetadata) Is(state PollState) bool {
	// this special case check is needed, because 0 & x == 0 is true for any x
	if state == NonExistent {
		return m.State == NonExistent
	}
	return state&m.State == state
}

// Validate returns an error if the poll metadata is not valid; nil otherwise
func (m PollMetadata) Validate() error {
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

// PollID represents ID of polls
type PollID uint64

// String converts the give poll ID to string
func (id PollID) String() string {
	return strconv.FormatUint(uint64(id), 10)
}

func (m PollKey) String() string {
	return fmt.Sprintf("%s_%s", m.Module, m.ID)
}

// PollProperty is a modifier for PollMetadata. It should never be manually initialized
type PollProperty struct {
	do func(metadata PollMetadata) PollMetadata
}

func (p PollProperty) apply(metadata PollMetadata) PollMetadata {
	return p.do(metadata)
}

var _ codectypes.UnpackInterfacesMessage = PollMetadata{}

// UnpackInterfaces implements UnpackInterfacesMessage
func (m PollMetadata) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var data codec.ProtoMarshaler
	return unpacker.UnpackAny(m.Result, &data)
}

// With returns a new metadata object with all the given properties set
func (m PollMetadata) With(properties ...PollProperty) PollMetadata {
	newMetadata := m
	for _, property := range properties {
		newMetadata = property.apply(newMetadata)
	}
	return newMetadata
}

// ExpiryAt sets the expiry property on PollMetadata
func ExpiryAt(blockHeight int64) PollProperty {
	return PollProperty{do: func(metadata PollMetadata) PollMetadata {
		metadata.ExpiresAt = blockHeight
		return metadata
	}}
}

// Threshold sets the threshold property on PollMetadata
func Threshold(threshold utils.Threshold) PollProperty {
	return PollProperty{do: func(metadata PollMetadata) PollMetadata {
		metadata.VotingThreshold = threshold
		return metadata
	}}
}

// MinVoterCount sets the minimum number of voters that have to vote on PollMeta
// If not enough voters exist, then all of them have to vote
func MinVoterCount(minVoterCount int64) PollProperty {
	return PollProperty{do: func(metadata PollMetadata) PollMetadata {
		metadata.MinVoterCount = minVoterCount
		return metadata
	}}
}

// RewardPool sets the name of a reward pool for the poll
func RewardPool(rewardPoolName string) PollProperty {
	return PollProperty{do: func(metadata PollMetadata) PollMetadata {
		metadata.RewardPoolName = rewardPoolName
		return metadata
	}}
}

// GracePeriod sets the grace period after poll completion during which votes
// are still recorded
func GracePeriod(gracePeriod int64) PollProperty {
	return PollProperty{do: func(metadata PollMetadata) PollMetadata {
		metadata.GracePeriod = gracePeriod
		return metadata
	}}
}

// ModuleMetadata sets the module metadata on the poll
func ModuleMetadata(module string) PollProperty {
	return PollProperty{do: func(metadata PollMetadata) PollMetadata {
		// TODO: attach the actual module metadata
		metadata.ModuleMetadata = PollModuleMetadata{
			Module:   module,
			Metadata: nil,
		}

		return metadata
	}}
}

// Poll provides an interface for other modules to interact with polls
type Poll interface {
	HasVotedCorrectly(voter sdk.ValAddress) bool
	HasVoted(voter sdk.ValAddress) bool
	HasVotedLate(voter sdk.ValAddress) bool
	Vote(voter sdk.ValAddress, blockHeight int64, data codec.ProtoMarshaler) (result codec.ProtoMarshaler, voted bool, err error)
	Is(state PollState) bool
	SetExpired()
	GetResult() codec.ProtoMarshaler
	GetRewardPoolName() (string, bool)
	GetID() PollID
	GetVoters() []Voter
	GetTotalVotingPower() sdk.Int
	GetModuleMetadata() PollModuleMetadata
	Delete()
}
