package types_test

import (
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/multisig/exported"
	exportedtestutils "github.com/axelarnetwork/axelar-core/x/multisig/exported/testutils"
	"github.com/axelarnetwork/axelar-core/x/multisig/types"
	"github.com/axelarnetwork/axelar-core/x/multisig/types/testutils"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
	. "github.com/axelarnetwork/utils/test"
)

func TestSig(t *testing.T) {
	var (
		multiSig types.MultiSig
	)

	givenRandomSig := Given("a random sig", func() {
		multiSig = testutils.MultiSig()
	})

	t.Run("ValidateBasic", func(t *testing.T) {
		givenRandomSig.
			When("", func() {}).
			Then("should return nil", func(t *testing.T) {
				assert.NoError(t, multiSig.ValidateBasic())
			}).
			Run(t, 5)

		givenRandomSig.
			When("key ID is invalid", func() {
				multiSig.KeyID = "___---"
			}).
			Then("should return error", func(t *testing.T) {
				assert.Error(t, multiSig.ValidateBasic())
			}).
			Run(t)

		givenRandomSig.
			When("payload hash is invalid", func() {
				multiSig.PayloadHash = make([]byte, types.HashLength+1)
			}).
			Then("should return error", func(t *testing.T) {
				assert.Error(t, multiSig.ValidateBasic())
			}).
			Run(t)

		givenRandomSig.
			When("module is invalid", func() {
				multiSig.Module = "___---"
			}).
			Then("should return error", func(t *testing.T) {
				assert.Error(t, multiSig.ValidateBasic())
			}).
			Run(t)

		givenRandomSig.
			When("some participant is invalid", func() {
				sk := funcs.Must(btcec.NewPrivateKey(btcec.S256()))
				multiSig.Sigs[rand.StrBetween(10, 50)] = funcs.Must(sk.Sign(multiSig.PayloadHash)).Serialize()
			}).
			Then("should return error", func(t *testing.T) {
				assert.Error(t, multiSig.ValidateBasic())
			}).
			Run(t, 5)

		givenRandomSig.
			When("some signature is invalid", func() {
				multiSig.Sigs[rand.ValAddr().String()] = rand.BytesBetween(1, 100)
			}).
			Then("should return error", func(t *testing.T) {
				assert.Error(t, multiSig.ValidateBasic())
			}).
			Run(t, 5)

		givenRandomSig.
			When("duplicate signatures exist", func() {
				sk := funcs.Must(btcec.NewPrivateKey(btcec.S256()))
				signature := funcs.Must(sk.Sign(multiSig.PayloadHash)).Serialize()

				multiSig.Sigs[rand.ValAddr().String()] = signature
				multiSig.Sigs[rand.ValAddr().String()] = signature
			}).
			Then("should return error", func(t *testing.T) {
				assert.Error(t, multiSig.ValidateBasic())
			}).
			Run(t)
	})
}

func TestSigningSession(t *testing.T) {
	var (
		signingSession types.SigningSession
		validators     []sdk.ValAddress
		privateKeys    map[string]*btcec.PrivateKey
	)

	givenNewSignSession := Given("new signing session", func() {
		validators = slices.Expand(func(int) sdk.ValAddress { return rand.ValAddr() }, 3)
		participants := map[string]snapshot.Participant{
			validators[0].String(): snapshot.NewParticipant(validators[0], sdk.NewUint(1)),
			validators[1].String(): snapshot.NewParticipant(validators[1], sdk.NewUint(2)),
			validators[2].String(): snapshot.NewParticipant(validators[2], sdk.NewUint(3)),
		}
		publicKeys := make(map[string]types.PublicKey, len(validators))
		privateKeys = make(map[string]*btcec.PrivateKey, len(validators))
		for _, v := range validators {
			privateKey := funcs.Must(btcec.NewPrivateKey(btcec.S256()))
			privateKeys[v.String()] = privateKey
			publicKeys[v.String()] = privateKey.PubKey().SerializeCompressed()
		}

		id := uint64(rand.PosI64())
		key := types.Key{
			ID: exportedtestutils.KeyID(),
			Snapshot: snapshot.Snapshot{
				Timestamp:    time.Now(),
				Height:       rand.PosI64(),
				Participants: participants,
				BondedWeight: sdk.NewUint(6),
			},
			PubKeys:          publicKeys,
			SigningThreshold: utils.NewThreshold(2, 3),
		}
		payloadHash := rand.Bytes(types.HashLength)
		expiresAt := rand.I64Between(1, 100)
		gracePeriod := int64(3)
		module := rand.NormalizedStr(5)

		signingSession = types.NewSigningSession(id, key, payloadHash, expiresAt, gracePeriod, module)
	})

	whenIsCompleted := When("is completed", func() {
		signingSession.State = exported.Completed
		signingSession.CompletedAt = signingSession.ExpiresAt - 1
		signingSession.MultiSig.Sigs = make(map[string]types.Signature)

		for p, sk := range privateKeys {
			signingSession.MultiSig.Sigs[p] = funcs.Must(sk.Sign(signingSession.MultiSig.PayloadHash)).Serialize()
		}
	})

	t.Run("ValidateBasic", func(t *testing.T) {
		givenNewSignSession.
			When("", func() {}).
			Then("should return nil", func(t *testing.T) {
				assert.NoError(t, signingSession.ValidateBasic())
			}).
			Run(t)

		givenNewSignSession.
			When("key is invalid", func() {
				signingSession.Key = types.Key{}
			}).
			Then("should return error", func(t *testing.T) {
				assert.Error(t, signingSession.ValidateBasic())
			}).
			Run(t)

		givenNewSignSession.
			When("multi sig is invalid", func() {
				signingSession.MultiSig = types.MultiSig{}
			}).
			Then("should return error", func(t *testing.T) {
				assert.Error(t, signingSession.ValidateBasic())
			}).
			Run(t)

		givenNewSignSession.
			When("state is invalid", func() {
				signingSession.State = exported.NonExistent
			}).
			Then("should return error", func(t *testing.T) {
				assert.Error(t, signingSession.ValidateBasic())
			}).
			Run(t)

		givenNewSignSession.
			When("key ID on multi sig mismatches with the key", func() {
				signingSession.Key.ID = exportedtestutils.KeyID()
			}).
			Then("should return error", func(t *testing.T) {
				assert.Error(t, signingSession.ValidateBasic())
			}).
			Run(t)

		givenNewSignSession.
			When("expires at is not set", func() {
				signingSession.ExpiresAt = 0
			}).
			Then("should return error", func(t *testing.T) {
				assert.Error(t, signingSession.ValidateBasic())
			}).
			Run(t)

		givenNewSignSession.
			When("completed at is set", func() {
				signingSession.CompletedAt = 10
			}).
			Then("should return error", func(t *testing.T) {
				assert.Error(t, signingSession.ValidateBasic())
			}).
			Run(t)

		givenNewSignSession.
			When("completed at is set", func() {
				signingSession.CompletedAt = 10
			}).
			Then("should return error", func(t *testing.T) {
				assert.Error(t, signingSession.ValidateBasic())
			}).
			Run(t)

		givenNewSignSession.
			When("some participant in multi sig is not found in the key", func() {
				sk := funcs.Must(btcec.NewPrivateKey(btcec.S256()))

				signingSession.MultiSig.Sigs = make(map[string]types.Signature)
				signingSession.MultiSig.Sigs[rand.ValAddr().String()] = funcs.Must(sk.Sign(signingSession.MultiSig.PayloadHash)).Serialize()
			}).
			Then("should return error", func(t *testing.T) {
				assert.Error(t, signingSession.ValidateBasic())
			}).
			Run(t)

		givenNewSignSession.
			When("some signature in multi sig mismatches with the corresponding public key", func() {
				sk := funcs.Must(btcec.NewPrivateKey(btcec.S256()))

				signingSession.MultiSig.Sigs = make(map[string]types.Signature)
				signingSession.MultiSig.Sigs[validators[0].String()] = funcs.Must(sk.Sign(signingSession.MultiSig.PayloadHash)).Serialize()
			}).
			Then("should return error", func(t *testing.T) {
				assert.Error(t, signingSession.ValidateBasic())
			}).
			Run(t)

		givenNewSignSession.
			When2(whenIsCompleted).
			Then("should return nil", func(t *testing.T) {
				assert.NoError(t, signingSession.ValidateBasic())
			}).
			Run(t)

		givenNewSignSession.
			When2(whenIsCompleted).
			When("completed at is not set", func() {
				signingSession.CompletedAt = 0
			}).
			Then("should return error", func(t *testing.T) {
				assert.Error(t, signingSession.ValidateBasic())
			}).
			Run(t)

		givenNewSignSession.
			When2(whenIsCompleted).
			When("signing threshold is not met", func() {
				delete(signingSession.MultiSig.Sigs, validators[2].String())
			}).
			Then("should return error", func(t *testing.T) {
				assert.Error(t, signingSession.ValidateBasic())
			}).
			Run(t)
	})
}
