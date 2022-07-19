package keeper

import (
	"bytes"
	"github.com/axelarnetwork/utils/funcs"
	"golang.org/x/exp/maps"
	"sort"
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
	var multisigKey *multisigtypes.MultiSig

	// For testing
	var addressWeights map[string]sdk.Uint
	var participantSigs map[string]multisigtypes.Signature
	var sigAddresses map[string]string

	// Calculate cumulative weights from sig set
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
		participantSigs = make(map[string]multisigtypes.Signature, len(participants))
		sigAddresses = make(map[string]string, len(participants))

		for _, p := range participants {
			randSig := rand.Bytes(crypto.SignatureLength)

			pubKey := funcs.MustOk(key.GetPubKey(p))
			address := crypto.PubkeyToAddress(pubKey.ToECDSAPubKey())
			sigAddresses[common.Bytes2Hex(randSig)] = address.Hex()

			weight := key.GetWeight(p)
			addressWeights[address.Hex()] = weight
			participantSigs[p.String()] = randSig
		}
		multisigKey = &multisigtypes.MultiSig{
			Sigs: participantSigs,
		}
	})
	shouldOptimizeSigSet := Then("should optimize signature set", func(t *testing.T) {
		optimizedSigs := optimizeSignatureSet(multisigKey, key)

		// Optimized sigs should pass threshold
		assert.True(t, slices.Reduce(optimizedSigs, sdk.ZeroUint(), f).GTE(key.GetMinPassingWeight()))

		// Sig set addresses should be in ascending order
		sortedAddresses := slices.Map(maps.Keys(addressWeights), common.HexToAddress)
		sort.SliceStable(sortedAddresses, func(i, j int) bool {
			return bytes.Compare(sortedAddresses[i].Bytes(), sortedAddresses[j].Bytes()) < 0
		})

		optimizedAddresses := slices.Map(optimizedSigs, func(sig []byte) common.Address {
			return common.HexToAddress(sigAddresses[common.Bytes2Hex(sig)])
		})

		assert.True(t, isSigSetOrdered(optimizedAddresses, sortedAddresses))

		minWeightedSig := slices.Reduce(optimizedSigs[1:], optimizedSigs[0], func(c []byte, s []byte) []byte {
			minWeight := addressWeights[common.Bytes2Hex(c)]
			next := addressWeights[common.Bytes2Hex(s)]
			if next.LT(minWeight) {
				c = s
			}

			return c
		})
		assert.False(t, slices.Reduce(optimizedSigs, sdk.ZeroUint(), f).Sub(addressWeights[sigAddresses[common.Bytes2Hex(minWeightedSig)]]).GTE(key.GetMinPassingWeight()))

	})
	givenWeightsAndSigs.
		When("", func() {}).
		Then2(shouldOptimizeSigSet).
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
					delete(participantSigs, p.String())
					currWeight = currWeight.Sub(key.GetWeight(p))
				}
			}
			multisigKey.Sigs = participantSigs
		}).
		Then2(shouldOptimizeSigSet).
		Run(t, 20)
}

func isSigSetOrdered(optimizedAddresses, sortedAddresses []common.Address) bool {
	i := 0
	j := 0
	for i < len(optimizedAddresses) {
		if j >= len(sortedAddresses) {
			return false
		}
		switch bytes.Compare(optimizedAddresses[i].Bytes(), sortedAddresses[j].Bytes()) {
		case 0:
			i++
			j++
		case 1:
			j++
		default:
			return false
		}

	}
	return true
}
