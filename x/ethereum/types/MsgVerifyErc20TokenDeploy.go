package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
)

type MsgVerifyErc20TokenDeploy struct {
	Sender      sdk.AccAddress
	TxID        common.Hash
	Symbol      string
	GatewayAddr common.Address
}

func NewMsgVerifyErc20TokenDeploy(sender sdk.AccAddress, txID common.Hash, symbol string, gatewayAddr common.Address) sdk.Msg {
	return MsgVerifyErc20TokenDeploy{
		Sender:      sender,
		TxID:        txID,
		Symbol:      symbol,
		GatewayAddr: gatewayAddr,
	}
}

func (msg MsgVerifyErc20TokenDeploy) Route() string {
	return RouterKey
}

func (msg MsgVerifyErc20TokenDeploy) Type() string {
	return "VerifyErc20TokenDeploy"
}

func (msg MsgVerifyErc20TokenDeploy) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}

	return nil
}

func (msg MsgVerifyErc20TokenDeploy) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgVerifyErc20TokenDeploy) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
