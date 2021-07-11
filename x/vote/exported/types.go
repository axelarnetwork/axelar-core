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

var _ codectypes.UnpackInterfacesMessage = PollMetadata{}

// NewPollMetaData is the constructor for PollMetadata
func NewPollMetaData(key PollKey, snapshotSeqNo int64, expiresAt int64, threshold utils.Threshold) PollMetadata {
	return PollMetadata{
		Key:             key,
		SnapshotSeqNo:   snapshotSeqNo,
		ExpiresAt:       expiresAt,
		Result:          nil,
		VotingThreshold: threshold,
		State:           Pending,
	}
}

func (m PollMetadata) Is(state PollState) bool {
	if state == NonExistent {
		return m.State == NonExistent
	}
	return state&m.State == state
}

func (m PollMetadata) UpdateBlockHeight(height int64) PollMetadata {
	if m.ExpiresAt <= height && m.Is(Pending) {
		m.State |= Expired
	}
	return m
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

type Poll interface {
	Vote(voter sdk.ValAddress, data codec.ProtoMarshaler) error
	Is(state PollState) bool
	GetMetadata() PollMetadata
	Initialize() error
	Delete()
}
