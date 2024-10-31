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
	"github.com/axelarnetwork/axelar-core/x/axelarnet/keeper"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types/mock"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	nexusmock "github.com/axelarnetwork/axelar-core/x/nexus/exported/mock"
	nexustypes "github.com/axelarnetwork/axelar-core/x/nexus/types"
	. "github.com/axelarnetwork/utils/test"
)

func TestMigrate6to7(t *testing.T) {
	var (
		bank          *mock.BankKeeperMock
		account       *mock.AccountKeeperMock
		nexusK        *mock.NexusMock
		lockableAsset *nexusmock.LockableAssetMock
	)

	encCfg := app.MakeEncodingConfig()
	subspace := params.NewSubspace(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("axelarnetKey"), sdk.NewKVStoreKey("tAxelarnetKey"), "axelarnet")
	k := keeper.NewKeeper(encCfg.Codec, sdk.NewKVStoreKey("axelarnet"), subspace, &mock.ChannelKeeperMock{}, &mock.FeegrantKeeperMock{})
	ibcK := keeper.NewIBCKeeper(k, &mock.IBCTransferKeeperMock{})
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())

	Given("keeper setup before migration", func() {
		bank = &mock.BankKeeperMock{}
		account = &mock.AccountKeeperMock{
			GetModuleAddressFunc: func(_ string) sdk.AccAddress {
				return rand.AccAddr()
			},
		}
		lockableAsset = &nexusmock.LockableAssetMock{
			LockFromFunc: func(ctx sdk.Context, fromAddr sdk.AccAddress) error {
				return nil
			},
			GetAssetFunc: func() sdk.Coin {
				return rand.Coin()
			},
		}
		nexusK = &mock.NexusMock{
			NewLockableAssetFunc: func(ctx sdk.Context, ibc nexustypes.IBCKeeper, bank nexustypes.BankKeeper, coin sdk.Coin) (nexus.LockableAsset, error) {
				return lockableAsset, nil
			},
		}
	}).
		When("Axelarnet module account has balance for failed cross chain transfers", func() {
			bank.SpendableCoinsFunc = func(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
				return sdk.NewCoins(rand.Coin(), rand.Coin(), rand.Coin())
			}
		}).
		Then("the migration should lock back to escrow account", func(t *testing.T) {
			err := keeper.Migrate6to7(k, bank, account, nexusK, ibcK)(ctx)
			assert.NoError(t, err)
			assert.Len(t, lockableAsset.LockFromCalls(), 3)
		}).
		Run(t)
}
