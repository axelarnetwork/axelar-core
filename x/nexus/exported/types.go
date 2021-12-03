package exported

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AddressValidator defines a function that implements address verification upon a request to link addresses
type AddressValidator func(ctx sdk.Context, address CrossChainAddress) error

// Validate validates the TransferState
func (m TransferState) Validate() error {
	_, ok := TransferState_name[int32(m)]
	if !ok {
		return fmt.Errorf("unknown transfer state")
	}

	if m == TRANSFER_STATE_UNSPECIFIED {
		return fmt.Errorf("unspecified transfer state")
	}

	return nil
}

// Validate validates the CrossChainTransfer
func (m CrossChainTransfer) Validate() error {
	if err := m.Recipient.Validate(); err != nil {
		return err
	}

	if err := m.Asset.Validate(); err != nil {
		return err
	}

	if err := m.State.Validate(); err != nil {
		return err
	}

	return nil
}

// Validate validates the CrossChainAddress
func (m CrossChainAddress) Validate() error {
	if err := m.Chain.Validate(); err != nil {
		return err
	}

	// TODO: validate .Address according to .Chain
	if m.Address == "" {
		return fmt.Errorf("address must be set")
	}

	return nil
}

// NewPendingCrossChainTransfer returns a pending CrossChainTransfer
func NewPendingCrossChainTransfer(id uint64, recipient CrossChainAddress, asset sdk.Coin) CrossChainTransfer {
	return CrossChainTransfer{
		ID:        id,
		Recipient: recipient,
		Asset:     asset,
		State:     Pending,
	}
}

// Validate performs a stateless check to ensure the Chain object has been initialized correctly
func (m Chain) Validate() error {
	if m.Name == "" {
		return fmt.Errorf("missing chain name")
	}

	if m.NativeAsset == "" {
		return fmt.Errorf("missing native asset name")
	}

	if err := m.KeyType.Validate(); err != nil {
		return err
	}

	if m.Module == "" {
		return fmt.Errorf("missing module name")
	}

	return nil
}
