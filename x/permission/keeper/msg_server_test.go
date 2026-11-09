package keeper_test

import (
	"testing"

	"cosmossdk.io/log"
	store "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/x/permission/exported"
	"github.com/axelarnetwork/axelar-core/x/permission/keeper"
	"github.com/axelarnetwork/axelar-core/x/permission/types"
	. "github.com/axelarnetwork/utils/test"
)

func TestMsgServer_UpdateGovernanceKey(t *testing.T) {
	cfg := app.MakeEncodingConfig()
	var (
		k       keeper.Keeper
		ctx     sdk.Context
		server  types.MsgServer
		oldKey  *multisig.LegacyAminoPubKey
		newKey  *multisig.LegacyAminoPubKey
		oldAddr sdk.AccAddress
		newAddr sdk.AccAddress
	)

	givenGovernanceKey := Given("a governance key", func() {
		subspace := paramstypes.NewSubspace(cfg.Codec, cfg.Amino, store.NewKVStoreKey("paramsKey"), store.NewKVStoreKey("tparamsKey"), "permission")
		k = keeper.NewKeeper(cfg.Codec, store.NewKVStoreKey(types.StoreKey), subspace)
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.NewTestLogger(t))
		server = keeper.NewMsgServerImpl(k)

		oldKey = randomMultisigGovernanceKey()
		oldAddr = oldKey.Address().Bytes()
		k.InitGenesis(ctx, types.NewGenesisState(types.DefaultParams(), oldKey, []types.GovAccount{
			types.NewGovAccount(oldAddr, exported.ROLE_ACCESS_CONTROL),
		}))
	})

	updateFrom := func(sender func() sdk.AccAddress) func() {
		return func() {
			req := types.NewUpdateGovernanceKeyRequest(sender(), 1, secp256k1.GenPrivKey().PubKey())
			newKey = &req.GovernanceKey
			newAddr = newKey.Address().Bytes()

			_, err := server.UpdateGovernanceKey(ctx, req)
			assert.NoError(t, err)
		}
	}

	thenOnlyNewKeyHasAccessControl := Then("only the new governance key has access control", func(t *testing.T) {
		key, ok := k.GetGovernanceKey(ctx)
		assert.True(t, ok)
		assert.Equal(t, newKey.Address(), key.Address())

		assert.Equal(t, exported.ROLE_ACCESS_CONTROL, k.GetRole(ctx, newAddr))
		assert.Equal(t, exported.ROLE_UNRESTRICTED, k.GetRole(ctx, oldAddr))
	})

	givenGovernanceKey.
		Branch(
			When("the governance account updates the key", updateFrom(func() sdk.AccAddress { return oldAddr })).
				Then2(thenOnlyNewKeyHasAccessControl),

			When("a governance proposal updates the key", updateFrom(func() sdk.AccAddress { return authtypes.NewModuleAddress(govtypes.ModuleName) })).
				Then2(thenOnlyNewKeyHasAccessControl),
		).Run(t)
}
