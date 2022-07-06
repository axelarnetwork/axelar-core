package types

import (
	"github.com/axelarnetwork/axelar-core/x/multisig/exported"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewKeygen is the constructor for event keygen started
func NewKeygen(action Keygen_Action, keyID exported.KeyID, participants []sdk.ValAddress) *Keygen {
	return &Keygen{
		Module:       ModuleName,
		Action:       action,
		KeyID:        keyID,
		Participants: participants,
	}
}
