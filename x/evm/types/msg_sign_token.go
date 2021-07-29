package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewSignDeployTokenRequest is the constructor for SignDeployTokenRequest
func NewSignDeployTokenRequest(sender sdk.AccAddress, chain, originChain, tokenName, symbol string, decimals uint8, capacity sdk.Int) *SignDeployTokenRequest {
	return &SignDeployTokenRequest{
		Sender:      sender,
		Chain:       chain,
		OriginChain: originChain,
		TokenName:   tokenName,
		Symbol:      symbol,
		Decimals:    decimals,
		Capacity:    capacity,
	}
}

// Route implements sdk.Msg
func (m SignDeployTokenRequest) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m SignDeployTokenRequest) Type() string {
	return "SignDeployToken"
}

// GetSignBytes  implements sdk.Msg
func (m SignDeployTokenRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements sdk.Msg
func (m SignDeployTokenRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}

// ValidateBasic implements sdk.Msg
func (m SignDeployTokenRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}
	if m.Chain == "" {
		return fmt.Errorf("missing chain")
	}
	if m.OriginChain == "" {
		return fmt.Errorf("missing origin chain")
	}
	if m.TokenName == "" {
		return fmt.Errorf("missing token name")
	}
	if m.Symbol == "" {
		return fmt.Errorf("missing token symbol")
	}
	if !m.Capacity.IsPositive() {
		return fmt.Errorf("token capacity must be a positive number")
	}

	return nil
}
