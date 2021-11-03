package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/tss/exported"
)

// NewSubmitMultiSigPubKeysRequest constructor for SetMultisigPubKeyRequest
func NewSubmitMultiSigPubKeysRequest(sender sdk.AccAddress, keyID exported.KeyID, pubKeyInfos []exported.PubKeyInfo) *SubmitMultisigPubKeysRequest {
	return &SubmitMultisigPubKeysRequest{Sender: sender, KeyID: keyID, PubKeyInfos: pubKeyInfos}
}

// Route implements the sdk.Msg interface.
func (m SubmitMultisigPubKeysRequest) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
// naming convention follows x/staking/types/msgs.go
func (m SubmitMultisigPubKeysRequest) Type() string { return "SubmitPubKeys" }

// ValidateBasic implements the sdk.Msg interface.
func (m SubmitMultisigPubKeysRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if err := m.KeyID.Validate(); err != nil {
		return err
	}

	// check uniqueness
	seen := map[string]bool{}
	for _, info := range m.PubKeyInfos {
		if err := info.Validate(); err != nil {
			return nil
		}
		if seen[string(info.PubKey)] {
			return fmt.Errorf("duplicate key")
		}
		seen[string(info.PubKey)] = true
	}

	return nil
}

// GetSignBytes implements the sdk.Msg interface
func (m SubmitMultisigPubKeysRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners implements the sdk.Msg interface
func (m SubmitMultisigPubKeysRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
