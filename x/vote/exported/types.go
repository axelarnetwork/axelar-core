package exported

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
)

//go:generate moq -out ./mock/types.go -pkg mock . Poll

// NewPollKey constructor for PollKey without nonce
func NewPollKey(module string, id string) PollKey {
	return PollKey{
		Module: module,
		ID:     id,
	}
}

func (m PollKey) String() string {
	return fmt.Sprintf("%s_%s", m.Module, m.ID)
}

// Validate performs a stateless validity check to ensure PollKey has been properly initialized
func (m PollKey) Validate() error {
	if m.Module == "" {
		return fmt.Errorf("missing module")
	}

	if m.ID == "" {
		return fmt.Errorf("missing poll ID")
	}

	return nil
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

// Poll provides an interface for other modules to interact with polls
type Poll interface {
	Vote(voter sdk.ValAddress, data codec.ProtoMarshaler) error
	Is(state PollState) bool
	AllowOverride()
	GetResult() codec.ProtoMarshaler
	GetKey() PollKey
	GetVoters() []Voter
	GetTotalVotingPower() sdk.Int
	Delete() error
}
