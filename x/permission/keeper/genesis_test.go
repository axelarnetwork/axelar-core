package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	rand2 "github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/permission/exported"
	"github.com/axelarnetwork/axelar-core/x/permission/keeper"
	"github.com/axelarnetwork/axelar-core/x/permission/types"
	. "github.com/axelarnetwork/utils/test"
	"github.com/axelarnetwork/utils/test/rand"
)

func TestGenesis(t *testing.T) {
	cfg := app.MakeEncodingConfig()
	var (
		k              keeper.Keeper
		ctx            sdk.Context
		initialGenesis *types.GenesisState
	)

	Given("a keeper",
		func() {
			subspace := paramstypes.NewSubspace(cfg.Codec, cfg.Amino, sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "permission")
			k = keeper.NewKeeper(cfg.Codec, sdk.NewKVStoreKey(types.StoreKey), subspace)

		}).
		When("the state is initialized from a genesis state",
			func() {
				initialGenesis = types.NewGenesisState(types.Params{}, randomMultisigGovernanceKey(), randomGovAccounts())
				assert.NoError(t, initialGenesis.Validate())

				ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
				k.InitGenesis(ctx, initialGenesis)
			}).
		Then("export the identical state",
			func(t *testing.T) {
				exportedGenesis := k.ExportGenesis(ctx)
				assert.NoError(t, exportedGenesis.Validate())
				assert.Equal(t, initialGenesis.GovernanceKey, exportedGenesis.GovernanceKey)
				assert.ElementsMatch(t, initialGenesis.GovAccounts, exportedGenesis.GovAccounts)
				assert.Equal(t, len(initialGenesis.GovAccounts), len(exportedGenesis.GovAccounts))
			}).Run(t, 10)
}

func randomMultisigGovernanceKey() *multisig.LegacyAminoPubKey {
	return multisig.NewLegacyAminoPubKey(3,
		[]crypto.PubKey{
			secp256k1.GenPrivKey().PubKey(),
			secp256k1.GenPrivKey().PubKey(),
			secp256k1.GenPrivKey().PubKey(),
			secp256k1.GenPrivKey().PubKey(),
			secp256k1.GenPrivKey().PubKey(),
			secp256k1.GenPrivKey().PubKey(),
		},
	)
}

func randomGovAccounts() []types.GovAccount {
	count := rand.I64Between(0, 10)
	var accounts []types.GovAccount
	for i := int64(0); i < count; i++ {
		accounts = append(accounts, randomGovAccount())
	}
	return accounts
}

func randomGovAccount() types.GovAccount {
	return types.GovAccount{
		Address: rand2.AccAddr(),
		Role:    exported.ROLE_CHAIN_MANAGEMENT,
	}
}
