package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	multisig "github.com/axelarnetwork/axelar-core/x/multisig/exported"
	multisigtypes "github.com/axelarnetwork/axelar-core/x/multisig/types"
	multisigTestutils2 "github.com/axelarnetwork/axelar-core/x/multisig/types/testutils"
	"github.com/axelarnetwork/utils/slices"
	. "github.com/axelarnetwork/utils/test"
)

func TestOptimizeSignatureSet(t *testing.T) {
	var key multisig.Key
	var addressWeights map[string]sdk.Uint
	var addressSigs map[string]multisigtypes.Signature
	var sigAddresses map[string]string // for testing

	// calculate cumulative weights from sig set
	f := func(c sdk.Uint, sig []byte) sdk.Uint {
		addr, ok := sigAddresses[common.Bytes2Hex(sig)]
		if !ok {
			panic("failed to get address")
		}
		return c.Add(addressWeights[addr])
	}

	givenWeightsAndSigs := Given("weights and signatures", func() {
		key = multisig.Key(multisigTestutils2.Key())
		participants := key.GetParticipants()

		addressWeights = make(map[string]sdk.Uint, len(participants))
		addressSigs = make(map[string]multisigtypes.Signature, len(participants))
		sigAddresses = make(map[string]string, len(participants))

		for _, p := range participants {
			weight := key.GetWeight(p)
			addressWeights[p.String()] = weight
			randSig := rand.Bytes(crypto.SignatureLength)
			addressSigs[p.String()] = randSig
			sigAddresses[common.Bytes2Hex(randSig)] = p.String()
		}

	})

	givenWeightsAndSigs.
		When("", func() {}).
		Then("should optimize signature set", func(t *testing.T) {
			optimizedSigs := optimizeSignatureSet(addressSigs, addressWeights, key.GetMinPassingWeight())

			// optimized sigs should pass threshold
			assert.True(t, slices.Reduce(optimizedSigs, sdk.ZeroUint(), f).GTE(key.GetMinPassingWeight()))

			// should not pass threshold without last sig
			assert.False(t, slices.Reduce(optimizedSigs[:len(optimizedSigs)-1], sdk.ZeroUint(), f).GTE(key.GetMinPassingWeight()))
		}).
		Run(t, 20)

	givenWeightsAndSigs.
		When("missing some signatures", func() {
			totalWeight := slices.Reduce(key.GetParticipants(), sdk.ZeroUint(),
				func(c sdk.Uint, p sdk.ValAddress) sdk.Uint {
					return c.Add(key.GetWeight(p))
				})

			currWeight := totalWeight
			for _, p := range key.GetParticipants() {
				if currWeight.Sub(key.GetWeight(p)).GTE(key.GetMinPassingWeight()) {
					delete(addressSigs, p.String())
					currWeight = currWeight.Sub(key.GetWeight(p))
				}
			}

		}).
		Then("should optimize signature set", func(t *testing.T) {
			f := func(c sdk.Uint, sig []byte) sdk.Uint {
				addr := sigAddresses[common.Bytes2Hex(sig)]
				return c.Add(addressWeights[addr])
			}

			optimizedSigs := optimizeSignatureSet(addressSigs, addressWeights, key.GetMinPassingWeight())

			// optimized sigs should pass threshold
			assert.True(t, slices.Reduce(optimizedSigs, sdk.ZeroUint(), f).GTE(key.GetMinPassingWeight()))

			// should not pass threshold without last sig
			assert.False(t, slices.Reduce(optimizedSigs[:len(optimizedSigs)-1], sdk.ZeroUint(), f).GTE(key.GetMinPassingWeight()))
		}).
		Run(t, 20)

	//givenOperators.
	//	Then("should optimize signature set", func(t *testing.T) {
	//		optimizedSigs := optimizeSignatureSet(addressSigs, addressWeights, key.GetMinPassingWeight())
	//
	//		// includes all operators
	//		assert.Equal(t, len(operators), len(optimized))
	//
	//		// sorted by address
	//		assert.True(t, sort.SliceIsSorted(optimized, func(i, j int) bool {
	//			return bytes.Compare(operators[i].Address.Bytes(), operators[j].Address.Bytes()) < 0
	//		}))
	//
	//		// optimized sigs should be above threshold
	//		operatorsWithSig := slices.Filter(optimized, func(o types.Operator) bool { return o.Signature != nil })
	//		assert.True(t, slices.Reduce(operatorsWithSig, sdk.ZeroUint(), func(c sdk.Uint, o types.Operator) sdk.Uint {
	//			return c.Add(o.Weight)
	//		}).GTE(key.GetMinPassingWeight()))
	//
	//		// sig should be nil after pass the threshold
	//		operatorsWithoutSig := slices.Filter(optimized, func(o types.Operator) bool { return o.Signature == nil })
	//		assert.Equal(t, len(operators), len(operatorsWithSig)+len(operatorsWithoutSig))
	//	}).
	//	Run(t, 20)
}
