package keeper

import (
	"math/rand"
	"testing"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	rand2 "github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

func TestKeeper_GetConfirmedOutPointInfos(t *testing.T) {
	setup := func() (Keeper, sdk.Context) {
		cdc := testutils.Codec()
		btcSubspace := params.NewSubspace(cdc, sdk.NewKVStoreKey("params"), sdk.NewKVStoreKey("tparams"), "btc")
		ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
		return NewKeeper(cdc, sdk.NewKVStoreKey("btc"), btcSubspace), ctx
	}

	testCases := []struct {
		label   string
		prepare func(k Keeper, ctx sdk.Context, infoCount int) (expected []types.OutPointInfo)
	}{
		{"no outpoints", prepareNoOutpoints},
		{"only pending outpoints", preparePendingOutPoints},
		{"only confirmed outpoints", prepareConfirmedOutPoints},
		{"only spent outpoints", prepareSpentOutPoints},
		{"random assortment of outpoint states", prepareRandomOutPointStates},
	}

	repeatCount := 10
	for _, testCase := range testCases {
		t.Run(testCase.label, testutils.Func(func(t *testing.T) {

			k, ctx := setup()
			infoCount := int(rand2.I64Between(1, 200))
			expectedOuts := testCase.prepare(k, ctx, infoCount)
			actualConfirmedOuts := k.GetConfirmedOutPointInfos(ctx)
			assert.ElementsMatch(t, expectedOuts, actualConfirmedOuts,
				"expected: %d elements, got: %d elements", len(expectedOuts), len(actualConfirmedOuts))

		}).Repeat(repeatCount))
	}
}

func prepareNoOutpoints(Keeper, sdk.Context, int) []types.OutPointInfo {
	return nil
}

func preparePendingOutPoints(k Keeper, ctx sdk.Context, infoCount int) []types.OutPointInfo {
	var outs []types.OutPointInfo
	for i := 0; i < infoCount; i++ {
		info := randOutPointInfo()
		k.SetPendingOutpointInfo(ctx, exported.PollMeta{ID: rand2.StrBetween(5, 20)}, info)
		outs = append(outs, info)
	}
	return nil
}

func prepareConfirmedOutPoints(k Keeper, ctx sdk.Context, infoCount int) []types.OutPointInfo {
	return prepareOutPoints(k, ctx, infoCount, types.CONFIRMED)
}

func prepareSpentOutPoints(k Keeper, ctx sdk.Context, infoCount int) []types.OutPointInfo {
	_ = prepareOutPoints(k, ctx, infoCount, types.SPENT)
	return nil
}

func prepareOutPoints(k Keeper, ctx sdk.Context, infoCount int, state types.OutPointState) []types.OutPointInfo {
	var outs []types.OutPointInfo
	for i := 0; i < infoCount; i++ {
		info := randOutPointInfo()
		k.SetOutpointInfo(ctx, info, state)
		outs = append(outs, info)
	}
	return outs
}

func prepareRandomOutPointStates(k Keeper, ctx sdk.Context, infoCount int) []types.OutPointInfo {
	var pendingCount, confirmedCount, spentCount int
	for _, state := range rand2.Distr(3).Samples(infoCount) {
		switch types.OutPointState(state) {
		case 2: // pending
			pendingCount++
		case types.CONFIRMED:
			confirmedCount++
		case types.SPENT:
			spentCount++
		}
	}
	_ = preparePendingOutPoints(k, ctx, pendingCount)
	_ = prepareOutPoints(k, ctx, spentCount, types.SPENT)
	return prepareOutPoints(k, ctx, confirmedCount, types.CONFIRMED)
}

func randOutPointInfo() types.OutPointInfo {
	txHash, err := chainhash.NewHash(rand2.Bytes(chainhash.HashSize))
	if err != nil {
		panic(err)
	}
	info := types.OutPointInfo{
		OutPoint: wire.NewOutPoint(txHash, rand.Uint32()),
		Amount:   btcutil.Amount(rand2.PosI64()),
		Address:  rand2.StrBetween(20, 60),
	}
	return info
}
