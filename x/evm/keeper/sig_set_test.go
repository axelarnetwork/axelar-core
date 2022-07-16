package keeper

import (
	"bytes"
	"sort"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	multisig "github.com/axelarnetwork/axelar-core/x/multisig/exported"
	multisigTestutils2 "github.com/axelarnetwork/axelar-core/x/multisig/types/testutils"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
	. "github.com/axelarnetwork/utils/test"
)

func TestOptimizeSignatureSet(t *testing.T) {
	var operators []types.Operator
	var key multisig.Key

	givenOperators := When("given a list of operators", func() {
		key = multisig.Key(multisigTestutils2.Key())

		for _, val := range key.GetParticipants() {
			pubKey := funcs.MustOk(key.GetPubKey(val))
			pk := pubKey.GetECDSAPubKey()

			operators = append(operators, types.Operator{
				Address:   types.Address(crypto.PubkeyToAddress(pk)),
				Weight:    key.GetWeight(val),
				Signature: rand.Bytes(crypto.SignatureLength),
			})
		}
	})

	givenOperators.
		Then("should optimize signature set", func(t *testing.T) {
			optimized := optimizeSignatureSet(key, operators)

			// includes all operators
			assert.Equal(t, len(operators), len(optimized))

			// sorted by address
			assert.True(t, sort.SliceIsSorted(optimized, func(i, j int) bool {
				return bytes.Compare(operators[i].Address.Bytes(), operators[j].Address.Bytes()) < 0
			}))

			// optimized sigs should be above threshold
			operatorsWithSig := slices.Filter(optimized, func(o types.Operator) bool { return o.Signature != nil })
			assert.True(t, slices.Reduce(operatorsWithSig, sdk.ZeroUint(), func(c sdk.Uint, o types.Operator) sdk.Uint {
				return c.Add(o.Weight)
			}).GTE(key.GetMinPassingWeight()))

			// sig should be nil after pass the threshold
			operatorsWithoutSig := slices.Filter(optimized, func(o types.Operator) bool { return o.Signature == nil })
			assert.Equal(t, len(operators), len(operatorsWithSig)+len(operatorsWithoutSig))
		}).
		Run(t, 20)

}
