package keeper

import (
	"testing"

	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	evm "github.com/axelarnetwork/axelar-core/x/evm/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/types/mock"
	"github.com/axelarnetwork/utils/test/rand"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
)

func TestGetMigrationHandler_activateCosmosChains(t *testing.T) {
	ctx, keeper := setup()
	axelarnetKeeper := mock.AxelarnetKeeperMock{
		IsCosmosChainFunc:   func(ctx sdk.Context, chain string) bool { return chain == axelarnet.Axelarnet.Name },
		GetFeeCollectorFunc: func(ctx sdk.Context) (sdk.AccAddress, bool) { return nil, false },
	}

	keeper.SetChain(ctx, axelarnet.Axelarnet)
	keeper.SetChain(ctx, evm.Ethereum)

	handler := GetMigrationHandler(keeper, &axelarnetKeeper)
	handler(ctx)

	chains := keeper.GetChains(ctx)
	assert.Len(t, chains, 2)
	for _, chain := range chains {
		assert.True(t, keeper.IsChainActivated(ctx, chain) == (chain.Name == axelarnet.Axelarnet.Name))
	}
}

func TestGetMigrationHandler_addTransferFee(t *testing.T) {
	feeCollector := rand.AccAddr()
	amount := sdk.NewCoin(axelarnet.Uaxl, sdk.NewInt(rand.PosI64()))
	feeCollectorAddress := exported.CrossChainAddress{Chain: axelarnet.Axelarnet, Address: feeCollector.String()}
	nonFeeCollectorAddress := exported.CrossChainAddress{Chain: axelarnet.Axelarnet, Address: rand.AccAddr().String()}
	nonAxelarnetAddress := exported.CrossChainAddress{Chain: evm.Ethereum, Address: feeCollector.String()}

	ctx, keeper := setup()
	axelarnetKeeper := mock.AxelarnetKeeperMock{
		GetFeeCollectorFunc: func(ctx sdk.Context) (sdk.AccAddress, bool) { return feeCollector, true },
	}

	// archived
	keeper.setNewPendingTransfer(ctx, feeCollectorAddress, amount)
	keeper.ArchivePendingTransfer(ctx, exported.NewPendingCrossChainTransfer(0, feeCollectorAddress, amount))
	// not to fee collector
	keeper.setNewPendingTransfer(ctx, nonFeeCollectorAddress, amount)
	// not on axelarnet
	keeper.setNewPendingTransfer(ctx, nonAxelarnetAddress, amount)
	// pending transfer to fee collector on axelarnet
	keeper.setNewPendingTransfer(ctx, feeCollectorAddress, amount)

	handler := GetMigrationHandler(keeper, &axelarnetKeeper)
	handler(ctx)

	assert.Len(t, keeper.getTransfers(ctx), 3)
	assert.Equal(t, keeper.getTransferFee(ctx), exported.TransferFee{Coins: sdk.NewCoins(amount)})
}
