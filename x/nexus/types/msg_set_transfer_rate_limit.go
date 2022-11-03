package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// NewSetTransferEpochLimitRequest creates a message of type SetTransferEpochLimitRequest
func NewSetTransferEpochLimitRequest(sender sdk.AccAddress, chain exported.ChainName, limit sdk.Coin, window time.Duration) *SetTransferEpochLimitRequest {
	return &SetTransferEpochLimitRequest{
		Sender: sender,
		Chain:  chain,
		Limit:  limit,
		Window: window,
	}
}

// Route implements sdk.Msg
func (m SetTransferEpochLimitRequest) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m SetTransferEpochLimitRequest) Type() string {
	return "SetTransferEpochLimit"
}

// ValidateBasic implements sdk.Msg
func (m SetTransferEpochLimitRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if err := m.Chain.Validate(); err != nil {
		return err
	}

	if err := m.Limit.Validate(); err != nil {
		return err
	}

	if m.Window.Nanoseconds() <= 0 {
		return fmt.Errorf("rate limit window must be positive")
	}

	return nil
}

// GetSignBytes implements sdk.Msg
func (m SetTransferEpochLimitRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements sdk.Msg
func (m SetTransferEpochLimitRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
