package exported

import (
	"fmt"

	"github.com/axelarnetwork/axelar-core/utils"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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

	if err := utils.ValidateString(m.Address); err != nil {
		return sdkerrors.Wrap(err, "invalid address")
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
	if err := utils.ValidateString(m.Name); err != nil {
		return sdkerrors.Wrap(err, "invalid chain name")
	}

	if err := utils.ValidateString(m.NativeAsset); err != nil {
		return sdkerrors.Wrap(err, "invalid native asset")
	}

	if err := m.KeyType.Validate(); err != nil {
		return err
	}

	if m.Module == "" {
		return fmt.Errorf("missing module name")
	}

	return nil
}
