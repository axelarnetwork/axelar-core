package testutils

import (
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
)

// RandKeyID creates a random valid key ID
func RandKeyID() tss.KeyID {
	keyID := tss.KeyID(rand.StrBetween(tss.KeyIDLengthMin, tss.KeyIDLengthMax))
	if err := keyID.Validate(); err != nil {
		panic(err)
	}
	return keyID
}
