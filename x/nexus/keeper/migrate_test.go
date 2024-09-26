package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	axelarnettypes "github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	"github.com/axelarnetwork/axelar-core/x/nexus/keeper"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
	"github.com/axelarnetwork/axelar-core/x/nexus/types/mock"
)

func TestMigrate7to8(t *testing.T) {
	encCfg := app.MakeEncodingConfig()
	subspace := params.NewSubspace(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("nexusKey"), sdk.NewKVStoreKey("tNexusKey"), "nexus")
	k := keeper.NewKeeper(encCfg.Codec, sdk.NewKVStoreKey("nexus"), subspace)
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())

	t.Run("sendCoinsFromAxelarnetToNexus", func(t *testing.T) {
		bank := &mock.BankKeeperMock{}
		account := &mock.AccountKeeperMock{}
		balances := sdk.NewCoins(rand.Coin(), rand.Coin(), rand.Coin())
		axelarnetModuleAccount := rand.AccAddr()
		nexusModuleAccount := rand.AccAddr()

		account.GetModuleAddressFunc = func(name string) sdk.AccAddress {
			switch name {
			case axelarnettypes.ModuleName:
				return axelarnetModuleAccount
			case types.ModuleName:
				return nexusModuleAccount
			default:
				return sdk.AccAddress{}
			}
		}
		bank.GetAllBalancesFunc = func(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
			if addr.Equals(axelarnetModuleAccount) {
				return balances
			}
			return sdk.NewCoins()
		}
		bank.SendCoinsFromModuleToModuleFunc = func(ctx sdk.Context, sender, recipient string, coins sdk.Coins) error { return nil }

		err := keeper.Migrate7to8(k, bank, account)(ctx)

		assert.NoError(t, err)
		assert.Len(t, bank.SendCoinsFromModuleToModuleCalls(), 1)
		assert.Equal(t, axelarnettypes.ModuleName, bank.SendCoinsFromModuleToModuleCalls()[0].SenderModule)
		assert.Equal(t, types.ModuleName, bank.SendCoinsFromModuleToModuleCalls()[0].RecipientModule)
		assert.Equal(t, balances, bank.SendCoinsFromModuleToModuleCalls()[0].Amt)
	})
}
