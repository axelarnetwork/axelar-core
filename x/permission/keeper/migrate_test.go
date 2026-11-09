package keeper_test

import (
	"testing"

	"cosmossdk.io/log"
	store "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	rand2 "github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/permission/exported"
	"github.com/axelarnetwork/axelar-core/x/permission/keeper"
	"github.com/axelarnetwork/axelar-core/x/permission/types"
	. "github.com/axelarnetwork/utils/test"
)

func TestMigrate2to3(t *testing.T) {
	cfg := app.MakeEncodingConfig()
	var (
		k           keeper.Keeper
		ctx         sdk.Context
		govKey      *multisig.LegacyAminoPubKey
		obsolete    []sdk.AccAddress
		controllers []types.GovAccount
	)

	givenKeeper := Given("a keeper", func() {
		subspace := paramstypes.NewSubspace(cfg.Codec, cfg.Amino, store.NewKVStoreKey("paramsKey"), store.NewKVStoreKey("tparamsKey"), "permission")
		k = keeper.NewKeeper(cfg.Codec, store.NewKVStoreKey(types.StoreKey), subspace)
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.NewTestLogger(t))
	})

	withAccounts := func(withGovKey bool) func() {
		return func() {
			obsolete = []sdk.AccAddress{rand2.AccAddr(), rand2.AccAddr()}
			controllers = randomGovAccounts()

			accounts := append([]types.GovAccount{}, controllers...)
			for _, addr := range obsolete {
				accounts = append(accounts, types.NewGovAccount(addr, exported.ROLE_ACCESS_CONTROL))
			}

			govKey = nil
			if withGovKey {
				govKey = randomMultisigGovernanceKey()
				accounts = append(accounts, types.NewGovAccount(govKey.Address().Bytes(), exported.ROLE_ACCESS_CONTROL))
			}

			k.InitGenesis(ctx, types.NewGenesisState(types.DefaultParams(), govKey, accounts))
		}
	}

	migrate := When("migrating", func() {
		assert.NoError(t, keeper.Migrate2to3(k)(ctx))
	})

	thenControllersAreKept := Then("controllers keep their role", func(t *testing.T) {
		for _, controller := range controllers {
			assert.Equal(t, exported.ROLE_CHAIN_MANAGEMENT, k.GetRole(ctx, controller.Address))
		}
	})

	givenKeeper.
		Branch(
			Given("a governance key and obsolete access control accounts", withAccounts(true)).
				When2(migrate).
				Then("only the governance key has access control", func(t *testing.T) {
					assert.Equal(t, exported.ROLE_ACCESS_CONTROL, k.GetRole(ctx, govKey.Address().Bytes()))
					for _, addr := range obsolete {
						assert.Equal(t, exported.ROLE_UNRESTRICTED, k.GetRole(ctx, addr))
					}

					assert.NoError(t, k.ExportGenesis(ctx).Validate())
				}).
				Then2(thenControllersAreKept),

			Given("no governance key and access control accounts", withAccounts(false)).
				When2(migrate).
				Then("the access control accounts are kept", func(t *testing.T) {
					for _, addr := range obsolete {
						assert.Equal(t, exported.ROLE_ACCESS_CONTROL, k.GetRole(ctx, addr))
					}
				}).
				Then2(thenControllersAreKept),
		).Run(t, 10)
}
