package keeper

import (
	"testing"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/tss/types/mock"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

func TestGetMigrationHandler(t *testing.T) {
	encCfg := app.MakeEncodingConfig()
	storeKey := sdk.NewKVStoreKey(paramstypes.StoreKey)
	tstoreKey := sdk.NewKVStoreKey(paramstypes.TStoreKey)
	subspace := paramstypes.NewSubspace(encCfg.Codec, encCfg.Amino, storeKey, tstoreKey, types.ModuleName)
	k := NewKeeper(encCfg.Codec, storeKey, subspace, &mock.SlasherMock{}, &mock.RewarderMock{})
	GetMigrationHandler(k, &mock.MultiSigKeeperMock{}, &mock.NexusMock{}, &mock.SnapshotterMock{})

	// TODO: finish this test
	panic("implement me")
}
