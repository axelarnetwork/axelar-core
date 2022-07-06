package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/multisig/exported"
)

// NewKeygenStarted is the constructor for event keygen started
func NewKeygenStarted(keyID exported.KeyID, participants []sdk.ValAddress) *KeygenStarted {
	return &KeygenStarted{
		Module:       ModuleName,
		KeyID:        keyID,
		Participants: participants,
	}
}

// NewKeygenCompleted is the constructor for event keygen completed
func NewKeygenCompleted(keyID exported.KeyID) *KeygenCompleted {
	return &KeygenCompleted{
		Module: ModuleName,
		KeyID:  keyID,
	}
}

// NewPubKeySubmitted is the constructor for event pub key submitted
func NewPubKeySubmitted(keyID exported.KeyID, participant sdk.ValAddress, pubKey PublicKey) *PubKeySubmitted {
	return &PubKeySubmitted{
		Module:      ModuleName,
		KeyID:       keyID,
		Participant: participant,
		PubKey:      pubKey,
	}
}
