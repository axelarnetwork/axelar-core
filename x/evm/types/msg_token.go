package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"

	"github.com/axelarnetwork/axelar-core/utils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// NewConfirmTokenRequest creates a message of type ConfirmTokenRequest
func NewConfirmTokenRequest(sender sdk.AccAddress, chain string, asset Asset, txID common.Hash) *ConfirmTokenRequest {
	return &ConfirmTokenRequest{
		Sender: sender.String(),
		Chain:  nexus.ChainName(utils.NormalizeString(chain)),
		Asset:  asset,
		TxID:   Hash(txID),
	}
}

// Route implements sdk.Msg
func (m ConfirmTokenRequest) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m ConfirmTokenRequest) Type() string {
	return "ConfirmERC20TokenDeployment"
}

// ValidateBasic implements sdk.Msg
func (m ConfirmTokenRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, errorsmod.Wrap(err, "sender").Error())
	}

	if err := m.Chain.Validate(); err != nil {
		return errorsmod.Wrap(err, "invalid chain")
	}

	if err := m.Asset.Validate(); err != nil {
		return err
	}

	return nil
}

// GetSignBytes implements sdk.Msg
func (m ConfirmTokenRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}
