package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Ensure MsgTrackAddress implements sdk.Msg interface
var _ sdk.Msg = &MsgRegisterVoter{}

type MsgRegisterVoter struct {
	Validator sdk.ValAddress
	Voter     sdk.AccAddress
}

func NewMsgRegisterVoter(validator sdk.ValAddress, voter sdk.AccAddress) MsgRegisterVoter {
	return MsgRegisterVoter{
		Validator: validator,
		Voter:     voter,
	}
}

func (msg MsgRegisterVoter) Route() string {
	return RouterKey
}

func (msg MsgRegisterVoter) Type() string {
	return "RegisterVotingAccount"
}

func (msg MsgRegisterVoter) ValidateBasic() error {
	if msg.Validator.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing validator")
	}
	if msg.Voter.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing voter")
	}

	return nil
}

func (msg MsgRegisterVoter) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgRegisterVoter) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.AccAddress(msg.Validator)}
}
