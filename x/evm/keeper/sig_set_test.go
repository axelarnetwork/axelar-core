package keeper

import (
	"bytes"
	"sort"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/maps"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	multisig "github.com/axelarnetwork/axelar-core/x/multisig/exported"
	multisigtypes "github.com/axelarnetwork/axelar-core/x/multisig/types"
	multisigTestutils2 "github.com/axelarnetwork/axelar-core/x/multisig/types/testutils"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
	. "github.com/axelarnetwork/utils/test"
)

func TestOptimizeSignatureSet(t *testing.T) {
	var (
		key       multisig.Key
		signature *multisigtypes.MultiSig

		operators []types.Operator

		// For testing
		addressWeights map[string]sdk.Uint
		sigAddresses   map[string]string
		participants   []sdk.ValAddress
	)

	getOperator := func(val sdk.ValAddress) types.Operator {
		return types.Operator{
			Address:   crypto.PubkeyToAddress(funcs.MustOk(key.GetPubKey(val)).ToECDSAPubKey()),
			Signature: signature.Sigs[val.String()],
			Weight:    key.GetWeight(val),
		}
	}

	// Calculate cumulative weights from sig set
	f := func(c sdk.Uint, sig []byte) sdk.Uint {
		addr, ok := sigAddresses[common.Bytes2Hex(sig)]
		if !ok {
			panic("failed to get address")
		}
		return c.Add(addressWeights[addr])
	}

	givenWeightsAndSigs := Given("a multisig key", func() {
		k := multisigTestutils2.Key()
		key = multisig.Key(&k)
		participants = key.GetParticipants()

		addressWeights = make(map[string]sdk.Uint, len(participants))
		sigAddresses = make(map[string]string, len(participants))

		signature = &multisigtypes.MultiSig{
			Sigs: make(map[string]multisigtypes.Signature, len(participants)),
		}
		for _, p := range participants {
			// Build address, weight, sig map for testing
			sig := rand.Bytes(crypto.SignatureLength)
			address := crypto.PubkeyToAddress(funcs.MustOk(key.GetPubKey(p)).ToECDSAPubKey())
			addressWeights[address.Hex()] = key.GetWeight(p)
			sigAddresses[common.Bytes2Hex(sig)] = address.Hex()
			signature.Sigs[p.String()] = sig
		}
	}).
		Given("operator list", func() {
			operators = slices.Map(participants, getOperator)
		})

	shouldOptimizeSigSet := Then("should optimize signature set", func(t *testing.T) {
		optimizedSigs := optimizeSignatureSet(operators, key.GetMinPassingWeight())

		// Optimized sigs should pass threshold
		if !slices.Reduce(optimizedSigs, sdk.ZeroUint(), f).GTE(key.GetMinPassingWeight()) {
		}

		assert.True(t, slices.Reduce(optimizedSigs, sdk.ZeroUint(), f).GTE(key.GetMinPassingWeight()))

		// Optimized sig set follow it's evm addresses in ascending order
		sortedAddresses := slices.Map(maps.Keys(addressWeights), common.HexToAddress)
		sort.SliceStable(sortedAddresses, func(i, j int) bool {
			return bytes.Compare(sortedAddresses[i].Bytes(), sortedAddresses[j].Bytes()) < 0
		})
		optimizedAddresses := slices.Map(optimizedSigs, func(sig []byte) common.Address {
			return common.HexToAddress(sigAddresses[common.Bytes2Hex(sig)])
		})
		assert.True(t, isSigSetOrdered(optimizedAddresses, sortedAddresses))

		// Should not pass threshold without min weighted sig
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
			currWeight := slices.Reduce(participants, sdk.ZeroUint(),
				func(c sdk.Uint, p sdk.ValAddress) sdk.Uint {
					return c.Add(key.GetWeight(p))
				},
			)

			slices.Reduce(participants, currWeight,
				func(c sdk.Uint, p sdk.ValAddress) sdk.Uint {
					if currWeight.Sub(key.GetWeight(p)).GTE(key.GetMinPassingWeight()) {
						delete(signature.Sigs, p.String())
						currWeight = currWeight.Sub(key.GetWeight(p))
					}
					return currWeight
				},
			)
			operators = slices.Map(slices.Map(maps.Keys(signature.Sigs), func(s string) sdk.ValAddress {
				return funcs.Must(sdk.ValAddressFromBech32(s))
			}), getOperator)
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
