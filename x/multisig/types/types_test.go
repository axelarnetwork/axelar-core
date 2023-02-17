package types_test

import (
	"testing"

	"github.com/btcsuite/btcd/btcec/v2"
	ec "github.com/btcsuite/btcd/btcec/v2/ecdsa"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	utilstestutils "github.com/axelarnetwork/axelar-core/utils/testutils"
	"github.com/axelarnetwork/axelar-core/x/multisig/exported"
	multisigtestutils "github.com/axelarnetwork/axelar-core/x/multisig/exported/testutils"
	"github.com/axelarnetwork/axelar-core/x/multisig/types"
	typestestutils "github.com/axelarnetwork/axelar-core/x/multisig/types/testutils"
	snapshottestutils "github.com/axelarnetwork/axelar-core/x/snapshot/exported/testutils"
	"github.com/axelarnetwork/utils/funcs"
	. "github.com/axelarnetwork/utils/test"
)

func TestKeygenSession(t *testing.T) {
	var (
		keygenSession types.KeygenSession
		blockHeight   int64
		participant   sdk.ValAddress
		pubKey        exported.PublicKey
	)

	givenNewKeygenSession := Given("new keygen session", func() {
		threshold := utilstestutils.RandThreshold()
		snapshot := snapshottestutils.Snapshot(uint64(rand.I64Between(10, 20)), threshold)

		keygenSession = types.NewKeygenSession(multisigtestutils.KeyID(), threshold, threshold, snapshot, rand.I64Between(10, 100), types.DefaultParams().KeygenGracePeriod)
	})

	t.Run("ValidateBasic", func(t *testing.T) {
		givenNewKeygenSession.
			When("", func() {}).
			Then("should return nil", func(t *testing.T) {
				assert.NoError(t, keygenSession.ValidateBasic())
			}).
			Run(t, 5)

		givenNewKeygenSession.
			When("key is invalid", func() {
				keygenSession.Key.ID = ""
			}).
			Then("should return error", func(t *testing.T) {
				assert.Error(t, keygenSession.ValidateBasic())
			}).
			Run(t, 5)

		givenNewKeygenSession.
			When("keygen threshold is less than signing threshold", func() {
				keygenSession.KeygenThreshold = utils.NewThreshold(5, 10)
				keygenSession.Key.SigningThreshold = utils.NewThreshold(6, 10)
			}).
			Then("should return error", func(t *testing.T) {
				assert.Error(t, keygenSession.ValidateBasic())
			}).
			Run(t, 5)

		givenNewKeygenSession.
			When("expires at is not set", func() {
				keygenSession.ExpiresAt = 0
			}).
			Then("should return error", func(t *testing.T) {
				assert.Error(t, keygenSession.ValidateBasic())
			}).
			Run(t, 5)

		givenNewKeygenSession.
			When("is completed but compelete at is not set", func() {
				keygenSession.State = exported.Completed
				keygenSession.Key = typestestutils.Key()
			}).
			Then("should return error", func(t *testing.T) {
				assert.Error(t, keygenSession.ValidateBasic())
			}).
			Run(t, 5)

		givenNewKeygenSession.
			When("is pending but compelete at is set", func() {
				keygenSession.CompletedAt = 100
			}).
			Then("should return error", func(t *testing.T) {
				assert.Error(t, keygenSession.ValidateBasic())
			}).
			Run(t, 5)
	})

	t.Run("AddKey", func(t *testing.T) {
		givenNewKeygenSession.
			When("", func() {}).
			Then("should complete after all participants have submitted their public key", func(t *testing.T) {
				blockHeight := keygenSession.ExpiresAt - 1

				for _, p := range keygenSession.Key.Snapshot.GetParticipantAddresses() {
					err := keygenSession.AddKey(blockHeight, p, typestestutils.PublicKey())
					assert.NoError(t, err)
				}

				assert.Equal(t, exported.Completed, keygenSession.GetState())
				assert.Equal(t, blockHeight, keygenSession.GetCompletedAt())
			}).
			Run(t)

		givenNewKeygenSession.
			When("block height is after expiry", func() {
				blockHeight = keygenSession.GetExpiresAt() + rand.I64Between(0, 10)
				participant = keygenSession.GetKey().Snapshot.GetParticipantAddresses()[0]
				pubKey = typestestutils.PublicKey()
			}).
			Then("should return error", func(t *testing.T) {
				err := keygenSession.AddKey(blockHeight, participant, typestestutils.PublicKey())
				assert.ErrorContains(t, err, "expired")
			}).
			Run(t)

		givenNewKeygenSession.
			When("participant is not in the snapshot", func() {
				blockHeight = keygenSession.GetExpiresAt() - 1
				participant = rand.ValAddr()
				pubKey = typestestutils.PublicKey()
			}).
			Then("should return error", func(t *testing.T) {
				err := keygenSession.AddKey(blockHeight, participant, pubKey)
				assert.ErrorContains(t, err, "not a participant")
			}).
			Run(t)

		givenNewKeygenSession.
			When("participant has already submitted its public key", func() {
				blockHeight = keygenSession.GetExpiresAt() - 1
				participant = keygenSession.GetKey().Snapshot.GetParticipantAddresses()[0]
				pubKey = typestestutils.PublicKey()

				funcs.MustNoErr(keygenSession.AddKey(blockHeight, participant, pubKey))
			}).
			Then("should return error", func(t *testing.T) {
				err := keygenSession.AddKey(blockHeight, participant, pubKey)
				assert.ErrorContains(t, err, "already submitted")
			}).
			Run(t)

		givenNewKeygenSession.
			When("duplicate public key is received", func() {
				blockHeight = keygenSession.GetExpiresAt() - 1
				participant = keygenSession.GetKey().Snapshot.GetParticipantAddresses()[0]
				pubKey = typestestutils.PublicKey()

				funcs.MustNoErr(keygenSession.AddKey(blockHeight, participant, pubKey))
				participant = keygenSession.GetKey().Snapshot.GetParticipantAddresses()[1]
			}).
			Then("should return error", func(t *testing.T) {
				err := keygenSession.AddKey(blockHeight, participant, pubKey)
				assert.ErrorContains(t, err, "duplicate")
			}).
			Run(t)

		givenNewKeygenSession.
			When("keygen is already completed", func() {
				keygenSession.State = exported.Completed
				keygenSession.CompletedAt = keygenSession.ExpiresAt - keygenSession.GracePeriod - 2
			}).
			Then("should fail if past the grace period", func(t *testing.T) {
				blockHeight := keygenSession.CompletedAt + keygenSession.GracePeriod + 1
				err := keygenSession.AddKey(blockHeight, keygenSession.Key.Snapshot.GetParticipantAddresses()[0], typestestutils.PublicKey())
				assert.ErrorContains(t, err, "closed")
			}).
			Run(t)
	})

	t.Run("GetMissingParticipants", func(t *testing.T) {
		givenNewKeygenSession.
			When("", func() {}).
			Then("should return all participants", func(t *testing.T) {
				actual := keygenSession.GetMissingParticipants()

				assert.Equal(t, keygenSession.GetKey().Snapshot.GetParticipantAddresses(), actual)
			}).
			Run(t)

		givenNewKeygenSession.
			When("some participant has submitted its public key", func() {
				blockHeight = keygenSession.GetExpiresAt() - 1
				participant = keygenSession.GetKey().Snapshot.GetParticipantAddresses()[0]
				pubKey = typestestutils.PublicKey()

				keygenSession.AddKey(blockHeight, participant, pubKey)
			}).
			Then("should exclude those have submitted already", func(t *testing.T) {
				actual := keygenSession.GetMissingParticipants()

				assert.Subset(t, keygenSession.GetKey().Snapshot.GetParticipantAddresses(), actual)
				assert.NotContains(t, actual, participant)
			}).
			Run(t)
	})

	t.Run("Result", func(t *testing.T) {
		givenNewKeygenSession.
			When("", func() {}).
			Then("should return error", func(t *testing.T) {
				_, err := keygenSession.Result()

				assert.Error(t, err)
			}).
			Run(t)

		givenNewKeygenSession.
			When("enough participants have submitted their public keys", func() {
				for _, p := range keygenSession.Key.Snapshot.GetParticipantAddresses() {
					keygenSession.AddKey(keygenSession.ExpiresAt-1, p, typestestutils.PublicKey())
				}
			}).
			Then("should return a valid key", func(t *testing.T) {
				key, err := keygenSession.Result()

				assert.NoError(t, err)
				assert.NoError(t, key.ValidateBasic())
			}).
			Run(t)
	})
}

func TestKey(t *testing.T) {
	var (
		key types.Key
	)

	givenRandomKey := Given("random valid key", func() {
		key = typestestutils.Key()
	})

	t.Run("GetParticipantsWeight", func(t *testing.T) {
		givenRandomKey.
			When("all participants have submitted their public key", func() {}).
			Then("should return correct participants weight", func(t *testing.T) {
				assert.Equal(t, key.Snapshot.GetParticipantsWeight(), key.GetParticipantsWeight())
			}).
			Run(t, 5)
	})

	t.Run("ValidateBasic", func(t *testing.T) {
		givenRandomKey.
			When("is valid", func() {}).
			Then("should return nil", func(t *testing.T) {
				assert.NoError(t, key.ValidateBasic())
			}).
			Run(t, 5)

		givenRandomKey.
			When("id is invalid", func() {
				key.ID = ""
			}).
			Then("should return error", func(t *testing.T) {
				assert.Error(t, key.ValidateBasic())
			}).
			Run(t, 5)

		givenRandomKey.
			When("snapshot is invalid", func() {
				key.Snapshot.Participants = nil
			}).
			Then("should return error", func(t *testing.T) {
				assert.Error(t, key.ValidateBasic())
			}).
			Run(t, 5)

		givenRandomKey.
			When("contains duplicate public key", func() {
				publicKey := typestestutils.PublicKey()
				for address := range key.PubKeys {
					key.PubKeys[address] = publicKey
				}
			}).
			Then("should return error", func(t *testing.T) {
				assert.Error(t, key.ValidateBasic())
			}).
			Run(t, 5)

		givenRandomKey.
			When("contains invalid participant address", func() {
				key.PubKeys[rand.StrBetween(1, 100)] = typestestutils.PublicKey()
			}).
			Then("should return error", func(t *testing.T) {
				assert.Error(t, key.ValidateBasic())
			}).
			Run(t, 5)

		givenRandomKey.
			When("contains invalid public key", func() {
				for address := range key.PubKeys {
					key.PubKeys[address] = rand.Bytes(100)
				}
			}).
			Then("should return error", func(t *testing.T) {
				assert.Error(t, key.ValidateBasic())
			}).
			Run(t, 5)

		givenRandomKey.
			When("contains invalid participant", func() {
				key.PubKeys[rand.ValAddr().String()] = typestestutils.PublicKey()
			}).
			Then("should return error", func(t *testing.T) {
				assert.Error(t, key.ValidateBasic())
			}).
			Run(t, 5)

		givenRandomKey.
			When("contains invalid signing threshold", func() {
				key.SigningThreshold = utils.OneThreshold
				key.SigningThreshold.Numerator += 1
			}).
			Then("should return error", func(t *testing.T) {
				assert.Error(t, key.ValidateBasic())
			}).
			Run(t, 5)
	})
}

func TestSignature_Verify(t *testing.T) {
	var (
		sk      *btcec.PrivateKey
		payload []byte
		sig     types.Signature
	)
	Given("a private key", func() {
		sk = funcs.Must(btcec.NewPrivateKey())
	}).
		Given("a payload", func() {
			payload = rand.Bytes(30)
		}).
		Branch(
			When("a signature is created", func() {
				s := ec.Sign(sk, payload)
				sig = s.Serialize()
			}).
				Then("signature verification succeeds", func(t *testing.T) {
					assert.True(t, sig.Verify(payload, sk.PubKey().SerializeCompressed()))
				}),
			When("a an invalid signature is created", func() {
				wrongKey := funcs.Must(btcec.NewPrivateKey())
				s := ec.Sign(wrongKey, payload)
				sig = s.Serialize()
			}).
				Then("signature verification fails", func(t *testing.T) {
					assert.False(t, sig.Verify(payload, sk.PubKey().SerializeCompressed()))
				}),
		).Run(t)
}
