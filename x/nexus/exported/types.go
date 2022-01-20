package exported

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
)

// AddressValidator defines a function that implements address verification upon a request to link addresses
type AddressValidator func(ctx sdk.Context, address CrossChainAddress) error

// TransferStateFromString converts a describing state string to the corresponding TransferState
func TransferStateFromString(s string) TransferState {
	state, ok := TransferState_value["TRANSFER_STATE_"+strings.ToUpper(s)]

	if !ok {
		return TRANSFER_STATE_UNSPECIFIED
	}

	return TransferState(state)
}

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

// TransferID represents the unique cross transfer identifier
type TransferID uint64

// String returns a string representation of TransferID
func (t TransferID) String() string {
	return strconv.FormatUint(uint64(t), 10)
}

// Bytes returns the byte array of TransferID
func (t TransferID) Bytes() []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(t))

	return bz
}

// NewPendingCrossChainTransfer returns a pending CrossChainTransfer
func NewPendingCrossChainTransfer(id uint64, recipient CrossChainAddress, asset sdk.Coin) CrossChainTransfer {
	return CrossChainTransfer{
		ID:        TransferID(id),
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

	if err := sdk.ValidateDenom(m.NativeAsset); err != nil {
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

// NewAsset returns an asset struct
func NewAsset(denom string, minAmount sdk.Int) Asset {
	return Asset{Denom: utils.NormalizeString(denom), MinAmount: minAmount}
}

// Validate checks the stateless validity of the asset
func (m Asset) Validate() error {
	if err := sdk.ValidateDenom(m.Denom); err != nil {
		return sdkerrors.Wrap(err, "invalid denomination")
	}

	if m.MinAmount.LTE(sdk.ZeroInt()) {
		return fmt.Errorf("minimum amount must be greater than zero")
	}

	return nil
}
