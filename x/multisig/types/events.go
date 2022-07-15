package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/multisig/exported"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
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

// NewKeygenExpired is the constructor for event keygen expired
func NewKeygenExpired(keyID exported.KeyID) *KeygenExpired {
	return &KeygenExpired{
		Module: ModuleName,
		KeyID:  keyID,
	}
}

// NewPubKeySubmitted is the constructor for event pub key submitted
func NewPubKeySubmitted(keyID exported.KeyID, participant sdk.ValAddress, pubKey exported.PublicKey) *PubKeySubmitted {
	return &PubKeySubmitted{
		Module:      ModuleName,
		KeyID:       keyID,
		Participant: participant,
		PubKey:      pubKey,
	}
}

// NewSigningStarted is the constructor for event signing started
func NewSigningStarted(sigID uint64, key Key, payloadHash exported.Hash, requestingModule string) *SigningStarted {
	return &SigningStarted{
		Module:           ModuleName,
		SigID:            sigID,
		KeyID:            key.GetID(),
		PubKeys:          key.GetPubKeys(),
		PayloadHash:      payloadHash,
		RequestingModule: requestingModule,
	}
}

// NewSigningExpired is the constructor for event signing expired
func NewSigningExpired(sigID uint64) *SigningExpired {
	return &SigningExpired{
		Module: ModuleName,
		SigID:  sigID,
	}
}

// NewSigningCompleted is the constructor for event signing completed
func NewSigningCompleted(sigID uint64) *SigningCompleted {
	return &SigningCompleted{
		Module: ModuleName,
		SigID:  sigID,
	}
}

// NewSignatureSubmitted is the constructor for event signature submitted
func NewSignatureSubmitted(sigID uint64, participant sdk.ValAddress, signature Signature) *SignatureSubmitted {
	return &SignatureSubmitted{
		Module:      ModuleName,
		SigID:       sigID,
		Participant: participant,
		Signature:   signature,
	}
}

// NewKeyAssigned is the constructor for event key assigned
func NewKeyAssigned(chain nexus.ChainName, keyID exported.KeyID) *KeyAssigned {
	return &KeyAssigned{
		Module: ModuleName,
		Chain:  chain,
		KeyID:  keyID,
	}
}

// NewKeyRotated is the constructor for event key rotated
func NewKeyRotated(chain nexus.ChainName, keyID exported.KeyID) *KeyRotated {
	return &KeyRotated{
		Module: ModuleName,
		Chain:  chain,
		KeyID:  keyID,
	}
}
