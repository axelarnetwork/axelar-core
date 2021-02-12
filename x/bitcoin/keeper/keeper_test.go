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
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
)

func TestKeeper_GetVerifiedOutpoints(t *testing.T) {
	init := func() (Keeper, sdk.Context) {
		cdc := testutils.Codec()
		btcSubspace := params.NewSubspace(cdc, sdk.NewKVStoreKey("params"), sdk.NewKVStoreKey("tparams"), "btc")
		ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
		return NewKeeper(cdc, sdk.NewKVStoreKey("btc"), btcSubspace), ctx
	}

	testCases := []struct {
		label   string
		prepare func(k Keeper, ctx sdk.Context, infoCount int) (expected []types.OutPointInfo)
	}{
		{"no outpoints", func(Keeper, sdk.Context, int) []types.OutPointInfo { return nil }},
		{"only unverified outpoints", prepareUnverifiedOutPoints},
		{"only verified outpoints", prepareVerifiedOutPoints},
		{"only spent outpoints", prepareSpentOutPoints},
		{"random assortment of outpoint states", prepareRandomOutPointStates},
	}

	repetitions := 10
	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			for i := 0; i < repetitions; i++ {
				k, ctx := init()
				infoCount := int(testutils.RandIntBetween(1, 200))
				expectedOuts := testCase.prepare(k, ctx, infoCount)
				actualOuts := k.GetVerifiedOutPointInfos(ctx)
				assert.ElementsMatch(t, expectedOuts, actualOuts, "expected: %d elements, got: %d elements", len(expectedOuts), len(actualOuts))
			}
		})
	}
}

func prepareUnverifiedOutPoints(k Keeper, ctx sdk.Context, infoCount int) []types.OutPointInfo {
	for i := 0; i < infoCount; i++ {
		info := randOutPointInfo()
		k.SetUnverifiedOutpointInfo(ctx, info)
	}
	return nil
}

func prepareVerifiedOutPoints(k Keeper, ctx sdk.Context, infoCount int) []types.OutPointInfo {
	var outs []types.OutPointInfo
	for i := 0; i < infoCount; i++ {
		info := randOutPointInfo()
		k.SetUnverifiedOutpointInfo(ctx, info)
		k.ProcessVerificationResult(ctx, info.OutPoint.String(), true)
		outs = append(outs, info)
	}
	return outs
}

func prepareSpentOutPoints(k Keeper, ctx sdk.Context, infoCount int) []types.OutPointInfo {
	for i := 0; i < infoCount; i++ {
		info := randOutPointInfo()
		k.SetUnverifiedOutpointInfo(ctx, info)
		k.ProcessVerificationResult(ctx, info.OutPoint.String(), true)
		k.SpendVerifiedOutPoint(ctx, info.OutPoint.String())
	}
	return nil
}

func prepareRandomOutPointStates(k Keeper, ctx sdk.Context, infoCount int) (expected []types.OutPointInfo) {
	var unverifiedCount, verifiedCount, spentCount int
	for _, state := range testutils.RandDistr(3).Samples(infoCount) {
		switch state {
		case 0: // unverified
			unverifiedCount++
		case 1: // verified
			verifiedCount++
		case 2: // spent
			spentCount++
		}
	}
	prepareUnverifiedOutPoints(k, ctx, unverifiedCount)
	prepareSpentOutPoints(k, ctx, spentCount)
	return prepareVerifiedOutPoints(k, ctx, verifiedCount)
}

func randOutPointInfo() types.OutPointInfo {
	txHash, err := chainhash.NewHash(testutils.RandBytes(chainhash.HashSize))
	if err != nil {
		panic(err)
	}
	blockHash, err := chainhash.NewHash(testutils.RandBytes(chainhash.HashSize))
	if err != nil {
		panic(err)
	}
	info := types.OutPointInfo{
		OutPoint:      wire.NewOutPoint(txHash, rand.Uint32()),
		Amount:        btcutil.Amount(testutils.RandPosInt()),
		BlockHash:     blockHash,
		Address:       testutils.RandStringBetween(20, 60),
		Confirmations: rand.Uint64(),
	}
	return info
}
