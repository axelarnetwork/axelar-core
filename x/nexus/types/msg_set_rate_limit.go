package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// NewSetRateLimitRequest creates a message of type SetRateLimitRequest
func NewSetRateLimitRequest(sender sdk.AccAddress, chain exported.ChainName, limit sdk.Coin, window time.Duration) *SetRateLimitRequest {
	return &SetRateLimitRequest{
		Sender: sender,
		Chain:  chain,
		Limit:  limit,
		Window: window,
	}
}

// Route implements sdk.Msg
func (m SetRateLimitRequest) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m SetRateLimitRequest) Type() string {
	return "SetRateLimit"
}

// ValidateBasic implements sdk.Msg
func (m SetRateLimitRequest) ValidateBasic() error {
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
func (m SetRateLimitRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements sdk.Msg
func (m SetRateLimitRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
