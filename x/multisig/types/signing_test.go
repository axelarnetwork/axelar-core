package types_test

import (
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	ec "github.com/btcsuite/btcd/btcec/v2/ecdsa"
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
				multiSig.PayloadHash = make([]byte, exported.HashLength+1)
			}).
			Then("should return error", func(t *testing.T) {
				assert.Error(t, multiSig.ValidateBasic())
			}).
			Run(t)

		givenRandomSig.
			When("some participant is invalid", func() {
				sk := funcs.Must(btcec.NewPrivateKey())
				multiSig.Sigs[rand.StrBetween(10, 50)] = ec.Sign(sk, multiSig.PayloadHash).Serialize()
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
				sk := funcs.Must(btcec.NewPrivateKey())
				signature := ec.Sign(sk, multiSig.PayloadHash).Serialize()

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
		signatures     map[string]types.Signature
	)

	givenNewSignSession := Given("new signing session", func() {
		validators = slices.Expand(func(int) sdk.ValAddress { return rand.ValAddr() }, 3)
		participants := map[string]snapshot.Participant{
			validators[0].String(): snapshot.NewParticipant(validators[0], sdk.NewUint(1)),
			validators[1].String(): snapshot.NewParticipant(validators[1], sdk.NewUint(2)),
			validators[2].String(): snapshot.NewParticipant(validators[2], sdk.NewUint(3)),
		}
		publicKeys := make(map[string]exported.PublicKey, len(validators))
		privateKeys = make(map[string]*btcec.PrivateKey, len(validators))
		for _, v := range validators {
			privateKey := funcs.Must(btcec.NewPrivateKey())
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
		payloadHash := rand.Bytes(exported.HashLength)
		expiresAt := rand.I64Between(2, 100) // starts from 2 so CompletedAt can be 1
		gracePeriod := int64(3)
		module := rand.NormalizedStr(5)

		signingSession = types.NewSigningSession(id, key, payloadHash, expiresAt, gracePeriod, module)
	})

	whenSignaturesAreCreated := When("signatures are created", func() {
		signatures = make(map[string]types.Signature, len(privateKeys))
		for p, sk := range privateKeys {
			signatures[p] = ec.Sign(sk, signingSession.MultiSig.PayloadHash).Serialize()
		}
	})

	t.Run("ValidateBasic", func(t *testing.T) {
		whenIsCompleted := When("is completed", func() {
			signingSession.State = exported.Completed
			signingSession.CompletedAt = signingSession.ExpiresAt - 1
			signingSession.MultiSig.Sigs = make(map[string]types.Signature)

			for _, v := range validators {
				signingSession.MultiSig.Sigs[v.String()] = signatures[v.String()]
			}
		})

		givenNewSignSession.
			When("", func() {}).
			Then("should return nil", func(t *testing.T) {
				assert.NoError(t, signingSession.ValidateBasic())
			}).
			Run(t)

		givenNewSignSession.
			When("module is invalid", func() {
				signingSession.Module = "___---"
			}).
			Then("should return error", func(t *testing.T) {
				assert.Error(t, signingSession.ValidateBasic())
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
			When("completed at is not set", func() {
				signingSession.CompletedAt = rand.I64Between(-2, 1)
				signingSession.State = exported.Completed
			}).
			Then("should return error", func(t *testing.T) {
				assert.Error(t, signingSession.ValidateBasic())
			}).
			Run(t, 2)

		givenNewSignSession.
			When("completed at is set", func() {
				signingSession.CompletedAt = 10
				signingSession.State = exported.Pending
			}).
			Then("should return error", func(t *testing.T) {
				assert.Error(t, signingSession.ValidateBasic())
			}).
			Run(t)

		givenNewSignSession.
			When("completed at is greater than expires at", func() {
				signingSession.CompletedAt = 10
				signingSession.ExpiresAt = 9
			}).
			Then("should return error", func(t *testing.T) {
				assert.Error(t, signingSession.ValidateBasic())
			}).
			Run(t)

		givenNewSignSession.
			When("some participant in multi sig is not found in the key", func() {
				sk := funcs.Must(btcec.NewPrivateKey())

				signingSession.MultiSig.Sigs = make(map[string]types.Signature)
				signingSession.MultiSig.Sigs[rand.ValAddr().String()] = ec.Sign(sk, signingSession.MultiSig.PayloadHash).Serialize()
			}).
			Then("should return error", func(t *testing.T) {
				assert.Error(t, signingSession.ValidateBasic())
			}).
			Run(t)

		givenNewSignSession.
			When("some signature in multi sig mismatches with the corresponding public key", func() {
				sk := funcs.Must(btcec.NewPrivateKey())

				signingSession.MultiSig.Sigs = make(map[string]types.Signature)
				signingSession.MultiSig.Sigs[validators[0].String()] = ec.Sign(sk, signingSession.MultiSig.PayloadHash).Serialize()
			}).
			Then("should return error", func(t *testing.T) {
				assert.Error(t, signingSession.ValidateBasic())
			}).
			Run(t)

		givenNewSignSession.
			When2(whenSignaturesAreCreated).
			When2(whenIsCompleted).
			Then("should return nil", func(t *testing.T) {
				assert.NoError(t, signingSession.ValidateBasic())
			}).
			Run(t)

		givenNewSignSession.
			When2(whenSignaturesAreCreated).
			When2(whenIsCompleted).
			When("completed at is not set", func() {
				signingSession.CompletedAt = 0
			}).
			Then("should return error", func(t *testing.T) {
				assert.Error(t, signingSession.ValidateBasic())
			}).
			Run(t)

		givenNewSignSession.
			When2(whenSignaturesAreCreated).
			When2(whenIsCompleted).
			When("signing threshold is not met", func() {
				delete(signingSession.MultiSig.Sigs, validators[2].String())
			}).
			Then("should return error", func(t *testing.T) {
				assert.Error(t, signingSession.ValidateBasic())
			}).
			Run(t)
	})

	t.Run("AddSig", func(t *testing.T) {
		var (
			blockHeight int64
			participant sdk.ValAddress
			signature   types.Signature
		)

		whenIsNotExpired := When("is not expired", func() { blockHeight = signingSession.ExpiresAt - signingSession.GracePeriod - 2 })
		whenParticipantIsValid := When("participant is valid", func() {
			participant = rand.Of(validators...)
			signature = signatures[participant.String()]
		})

		givenNewSignSession.
			When2(whenSignaturesAreCreated).
			When("is expired", func() {
				blockHeight = signingSession.ExpiresAt
			}).
			When2(whenParticipantIsValid).
			Then("should return error", func(t *testing.T) {
				assert.Error(t, signingSession.AddSig(blockHeight, participant, signature))
			}).
			Run(t)

		givenNewSignSession.
			When2(whenSignaturesAreCreated).
			When2(whenIsNotExpired).
			When("participant is invalid", func() {
				participant = rand.ValAddr()
				signature = signatures[rand.Of(validators...).String()]
			}).
			Then("should return error", func(t *testing.T) {
				assert.Error(t, signingSession.AddSig(blockHeight, participant, signature))
			}).
			Run(t)

		givenNewSignSession.
			When2(whenSignaturesAreCreated).
			When2(whenIsNotExpired).
			When2(whenParticipantIsValid).
			When("participant has already submitted its signature", func() {
				funcs.MustNoErr(signingSession.AddSig(blockHeight, participant, signature))
			}).
			Then("should return error", func(t *testing.T) {
				assert.Error(t, signingSession.AddSig(blockHeight, participant, signature))
			}).
			Run(t)

		givenNewSignSession.
			When2(whenIsNotExpired).
			When2(whenParticipantIsValid).
			When("signature is invalid", func() {
				sk := funcs.Must(btcec.NewPrivateKey())
				signature = ec.Sign(sk, signingSession.MultiSig.PayloadHash).Serialize()
			}).
			Then("should return error", func(t *testing.T) {
				assert.Error(t, signingSession.AddSig(blockHeight, participant, signature))
			}).
			Run(t)

		givenNewSignSession.
			When2(whenSignaturesAreCreated).
			When2(whenIsNotExpired).
			When2(whenParticipantIsValid).
			When("is completed", func() {
				funcs.MustNoErr(signingSession.AddSig(blockHeight, validators[2], signatures[validators[2].String()]))
				funcs.MustNoErr(signingSession.AddSig(blockHeight, validators[1], signatures[validators[1].String()]))
			}).
			When("is outside the grace period", func() {
				blockHeight = signingSession.CompletedAt + signingSession.GracePeriod + 1
			}).
			Then("should return error", func(t *testing.T) {
				assert.ErrorContains(t, signingSession.AddSig(blockHeight, validators[0], signatures[validators[0].String()]), "closed")
			}).
			Run(t)

		givenNewSignSession.
			When2(whenSignaturesAreCreated).
			When2(whenIsNotExpired).
			Then("should be able to complete the signing session", func(t *testing.T) {
				startBlockHeight := blockHeight

				for i := len(validators) - 1; i >= 0; i-- {
					p := validators[i]
					assert.NoError(t, signingSession.AddSig(blockHeight, p, signatures[p.String()]))
					blockHeight += 1
				}

				assert.Equal(t, exported.Completed, signingSession.State)
				assert.Equal(t, startBlockHeight+1, signingSession.CompletedAt)
			}).
			Run(t)
	})

	t.Run("GetMissingParticipants", func(t *testing.T) {
		var (
			participant sdk.ValAddress
		)

		givenNewSignSession.
			When2(whenSignaturesAreCreated).
			When("some participant submitted its signature", func() {
				blockHeight := signingSession.ExpiresAt - 1
				participant = rand.Of(validators...)

				funcs.MustNoErr(signingSession.AddSig(blockHeight, participant, signatures[participant.String()]))
			}).
			Then("should return the correct missing participants", func(t *testing.T) {
				actual := signingSession.GetMissingParticipants()

				assert.Len(t, actual, 2)
				assert.ElementsMatch(t, slices.Filter(validators, func(v sdk.ValAddress) bool { return !participant.Equals(v) }), actual)
			}).
			Run(t, 5)
	})

	t.Run("Result", func(t *testing.T) {
		givenNewSignSession.
			When("", func() {}).
			Then("should return error", func(t *testing.T) {
				_, err := signingSession.Result()

				assert.Error(t, err)
			}).
			Run(t)

		givenNewSignSession.
			When2(whenSignaturesAreCreated).
			When("is completed", func() {
				blockHeight := signingSession.ExpiresAt - 1

				funcs.MustNoErr(signingSession.AddSig(blockHeight, validators[2], signatures[validators[2].String()]))
				funcs.MustNoErr(signingSession.AddSig(blockHeight, validators[1], signatures[validators[1].String()]))
			}).
			Then("should get valid multi sig", func(t *testing.T) {
				actual, err := signingSession.Result()

				assert.NoError(t, err)
				assert.NoError(t, actual.ValidateBasic())
				assert.Equal(t, signingSession.Key.ID, actual.KeyID)

				participantsWeight := sdk.ZeroUint()
				for p, sig := range actual.Sigs {
					participantsWeight = participantsWeight.Add(signingSession.Key.Snapshot.GetParticipantWeight(funcs.Must(sdk.ValAddressFromBech32(p))))

					pubKey, ok := signingSession.Key.PubKeys[p]

					assert.True(t, ok)
					assert.True(t, sig.Verify(actual.PayloadHash, pubKey))
				}
				assert.True(t, participantsWeight.GTE(signingSession.Key.GetMinPassingWeight()))
			}).
			Run(t)
	})
}
