package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewCreateDeployTokenRequest is the constructor for CreateDeployTokenRequest
func NewCreateDeployTokenRequest(sender sdk.AccAddress, chain, originChain, tokenName, symbol, nativeAsset string, decimals uint8, capacity sdk.Int) *CreateDeployTokenRequest {
	return &CreateDeployTokenRequest{
		Sender:      sender,
		Chain:       chain,
		OriginChain: originChain,
		TokenName:   tokenName,
		Symbol:      symbol,
		NativeAsset: nativeAsset,
		Decimals:    decimals,
		Capacity:    capacity,
	}
}

// Route implements sdk.Msg
func (m CreateDeployTokenRequest) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m CreateDeployTokenRequest) Type() string {
	return "CreateDeployToken"
}

// GetSignBytes  implements sdk.Msg
func (m CreateDeployTokenRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements sdk.Msg
func (m CreateDeployTokenRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}

// ValidateBasic implements sdk.Msg
func (m CreateDeployTokenRequest) ValidateBasic() error {
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
	if m.NativeAsset == "" {
		return fmt.Errorf("missing token native asset")
	}
	if !m.Capacity.IsPositive() {
		return fmt.Errorf("token capacity must be a positive number")
	}

	return nil
}
