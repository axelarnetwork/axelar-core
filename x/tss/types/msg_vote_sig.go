package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Route returns the route for this message
func (m VoteSigRequest) Route() string {
	return RouterKey
}

// Type returns the type of this message
func (m VoteSigRequest) Type() string {
	return "VoteSig"
}

// ValidateBasic performs a stateless validation of this message
func (m VoteSigRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, errorsmod.Wrap(err, "sender").Error())
	}

	if err := m.Result.Validate(); err != nil {
		return err
	}

	return nil
}

// GetSignBytes returns the bytes to sign for this message
func (m VoteSigRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}
