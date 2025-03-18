package types

import (
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// NewSetTransferRateLimitRequest creates a message of type SetTransferRateLimitRequest
func NewSetTransferRateLimitRequest(sender sdk.AccAddress, chain exported.ChainName, limit sdk.Coin, window time.Duration) *SetTransferRateLimitRequest {
	return &SetTransferRateLimitRequest{
		Sender: sender.String(),
		Chain:  chain,
		Limit:  limit,
		Window: window,
	}
}

// Route implements sdk.Msg
func (m SetTransferRateLimitRequest) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m SetTransferRateLimitRequest) Type() string {
	return "SetTransferRateLimit"
}

// ValidateBasic implements sdk.Msg
func (m SetTransferRateLimitRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, errorsmod.Wrap(err, "sender").Error())
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
func (m SetTransferRateLimitRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}
