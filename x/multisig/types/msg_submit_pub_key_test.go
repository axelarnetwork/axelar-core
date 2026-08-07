package types_test

import (
	"crypto/sha256"
	"testing"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/multisig/exported"
	"github.com/axelarnetwork/axelar-core/x/multisig/types"
	"github.com/axelarnetwork/utils/funcs"
)

func TestSubmitPubKeyRequest_VerifyPubKeyOwnership(t *testing.T) {
	keyID := exported.KeyID(rand.HexStr(5))
	newRequest := func(sender []byte, signedMessage string) *types.SubmitPubKeyRequest {
		sk := funcs.Must(btcec.NewPrivateKey())
		hash := sha256.Sum256([]byte(signedMessage))

		return types.NewSubmitPubKeyRequest(sender, keyID, sk.PubKey().SerializeCompressed(), ecdsa.Sign(sk, hash[:]).Serialize())
	}

	t.Run("should accept a proof that matches the sender", func(t *testing.T) {
		sender := rand.AccAddr()
		req := newRequest(sender, sender.String())

		assert.NoError(t, req.VerifyPubKeyOwnership())
	})

	t.Run("should reject a proof that does not match the sender, but only outside ValidateBasic", func(t *testing.T) {
		sender := rand.AccAddr()
		req := newRequest(sender, "not the sender")

		assert.NoError(t, req.ValidateBasic())
		assert.ErrorContains(t, req.VerifyPubKeyOwnership(), "signature does not match the public key")
	})
}
