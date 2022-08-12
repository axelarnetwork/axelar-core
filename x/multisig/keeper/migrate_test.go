package keeper

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/multisig/types"
	"github.com/axelarnetwork/axelar-core/x/multisig/types/mock"
	"github.com/axelarnetwork/axelar-core/x/multisig/types/testutils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
	. "github.com/axelarnetwork/utils/test"
)

func TestGetMigrationHandler(t *testing.T) {
	encCfg := params.MakeEncodingConfig()
	chain := nexus.Chain{
		Name: nexus.ChainName(rand.NormalizedStr(5)),
	}
	currentKey := testutils.Key()
	nextKey := testutils.Key()

	var (
		k         Keeper
		ctx       sdk.Context
		tssMock   *mock.TssMock
		nexusMock *mock.NexusMock
	)

	givenKeepers := Given("keepers", func() {
		subspace := paramstypes.NewSubspace(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "multisig")
		k = NewKeeper(encCfg.Codec, sdk.NewKVStoreKey(types.StoreKey), subspace)
		ctx = rand.Context(fake.NewMultiStore())

		k.InitGenesis(ctx, types.DefaultGenesisState())

		tssMock = &mock.TssMock{}
		nexusMock = &mock.NexusMock{
			GetChainsFunc: func(ctx sdk.Context) []nexus.Chain {
				return []nexus.Chain{chain}
			},
		}
	})

	givenKeepers.
		When("chain does not have current key set", func() {}).
		Then("should skip", func(t *testing.T) {
			handler := GetMigrationHandler(k, tssMock, nexusMock)
			assert.NoError(t, handler(ctx))
		}).
		Run(t)

	givenKeepers.
		When("chain has current key set", func() {
			k.setKey(ctx, currentKey)
			k.AssignKey(ctx, chain.Name, currentKey.ID)
			k.RotateKey(ctx, chain.Name)
		}).
		When("tss does not have the key set", func() {
			tssMock.GetKeyFunc = func(sdk.Context, tss.KeyID) (tss.Key, bool) { return tss.Key{}, false }
		}).
		Then("should skip", func(t *testing.T) {
			handler := GetMigrationHandler(k, tssMock, nexusMock)
			assert.NoError(t, handler(ctx))

			assert.Equal(t, currentKey.SigningThreshold, funcs.MustOk(k.getKey(ctx, currentKey.ID)).SigningThreshold)
		}).
		Run(t)

	givenKeepers.
		When("chain has current key set", func() {
			k.setKey(ctx, currentKey)
			k.AssignKey(ctx, chain.Name, currentKey.ID)
			k.RotateKey(ctx, chain.Name)
		}).
		When("tss has the key set", func() {
			tssMock.GetKeyFunc = func(sdk.Context, tss.KeyID) (tss.Key, bool) {
				t := rand.I64Between(1, currentKey.GetMinPassingWeight().BigInt().Int64())

				return tss.Key{
					PublicKey: &tss.Key_MultisigKey_{
						MultisigKey: &tss.Key_MultisigKey{
							Values:    slices.Expand(func(int) []byte { return rand.Bytes(32) }, int(currentKey.GetBondedWeight().Uint64())),
							Threshold: t,
						},
					},
				}, true
			}
		}).
		Then("should migrate the signing threshold", func(t *testing.T) {
			handler := GetMigrationHandler(k, tssMock, nexusMock)
			assert.NoError(t, handler(ctx))

			actual := funcs.MustOk(k.getKey(ctx, currentKey.ID))
			assert.NotEqual(t, currentKey.SigningThreshold, actual.SigningThreshold)
			assert.NoError(t, actual.ValidateBasic())
		}).
		Run(t)

	givenKeepers.
		When("chain has next key set", func() {
			k.setKey(ctx, currentKey)
			k.AssignKey(ctx, chain.Name, currentKey.ID)
			k.RotateKey(ctx, chain.Name)

			k.setKey(ctx, nextKey)
			k.AssignKey(ctx, chain.Name, nextKey.ID)
		}).
		When("tss has the key set", func() {
			tssMock.GetKeyFunc = func(ctx sdk.Context, keyID tss.KeyID) (tss.Key, bool) {
				var key types.Key

				switch keyID {
				case tss.KeyID(currentKey.ID):
					key = currentKey
				case tss.KeyID(nextKey.ID):
					key = nextKey
				default:
					panic(fmt.Errorf("unexpected key ID %s", keyID))
				}

				t := rand.I64Between(1, key.GetMinPassingWeight().BigInt().Int64())

				return tss.Key{
					PublicKey: &tss.Key_MultisigKey_{
						MultisigKey: &tss.Key_MultisigKey{
							Values:    slices.Expand(func(int) []byte { return rand.Bytes(32) }, int(key.GetBondedWeight().Uint64())),
							Threshold: t,
						},
					},
				}, true
			}
		}).
		Then("should migrate the signing threshold", func(t *testing.T) {
			handler := GetMigrationHandler(k, tssMock, nexusMock)
			assert.NoError(t, handler(ctx))

			actual := funcs.MustOk(k.getKey(ctx, currentKey.ID))
			assert.NotEqual(t, currentKey.SigningThreshold, actual.SigningThreshold)
			assert.NoError(t, actual.ValidateBasic())

			actual = funcs.MustOk(k.getKey(ctx, nextKey.ID))
			assert.NotEqual(t, nextKey.SigningThreshold, actual.SigningThreshold)
			assert.NoError(t, actual.ValidateBasic())
		}).
		Run(t)
}
