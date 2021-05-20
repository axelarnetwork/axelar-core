package keeper

import (
	mathRand "math/rand"
	"strings"
	"testing"
	"unicode"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	appParams "github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

func TestKeeper_GetAddress(t *testing.T) {
	var (
		ctx    sdk.Context
		keeper Keeper
	)
	setup := func() {
		encCfg := appParams.MakeEncodingConfig()
		btcSubspace := params.NewSubspace(encCfg.Marshaler, encCfg.Amino, sdk.NewKVStoreKey("params"), sdk.NewKVStoreKey("tparams"), "btc")
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
		keeper = NewKeeper(encCfg.Marshaler, sdk.NewKVStoreKey("btc"), btcSubspace)
	}
	t.Run("case insensitive", testutils.Func(func(t *testing.T) {
		setup()
		addr, err := btcutil.NewAddressWitnessScriptHash(rand.Bytes(32), types.Mainnet.Params())
		assert.NoError(t, err)

		addrStr1 := strings.ToLower(addr.EncodeAddress())
		addrStr2 := strings.ToUpper(addrStr1)
		assert.NotEqual(t, addrStr1, addrStr2)

		info := types.AddressInfo{
			Address:      addrStr1,
			Role:         types.Deposit,
			RedeemScript: rand.Bytes(200),
			KeyID:        rand.StrBetween(5, 20),
		}
		keeper.SetAddress(ctx, info)
		result, ok := keeper.GetAddress(ctx, addrStr2)
		assert.True(t, ok)
		assert.Equal(t, info, result)
	}).Repeat(20))

}

func TestKeeper_GetOutPointInfo(t *testing.T) {
	var (
		ctx    sdk.Context
		keeper Keeper
	)
	setup := func() {
		encCfg := appParams.MakeEncodingConfig()
		btcSubspace := params.NewSubspace(encCfg.Marshaler, encCfg.Amino, sdk.NewKVStoreKey("params"), sdk.NewKVStoreKey("tparams"), "btc")
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
		keeper = NewKeeper(encCfg.Marshaler, sdk.NewKVStoreKey("btc"), btcSubspace)
	}

	t.Run("case insensitive", testutils.Func(func(t *testing.T) {
		setup()
		hash, _ := chainhash.NewHash(rand.Bytes(chainhash.HashSize))

		outpoint := wire.NewOutPoint(hash, mathRand.Uint32())
		outStr := outpoint.String()

		var runes []rune
		flipDistr := rand.Bools(0.5)

		for _, r := range outStr {
			if unicode.IsLetter(r) && flipDistr.Next() {
				runes = append(runes, unicode.ToUpper(r))
			} else {
				runes = append(runes, r)
			}
		}

		outStr = string(runes)
		info := types.OutPointInfo{
			OutPoint: outStr,
			Amount:   btcutil.Amount(rand.PosI64()),
			Address:  rand.StrBetween(5, 100),
		}

		keeper.SetOutpointInfo(ctx, info, types.CONFIRMED)

		_, _, ok := keeper.GetOutPointInfo(ctx, info.GetOutPoint())
		assert.True(t, ok)
	}).Repeat(20))
}

func TestKeeper_GetConfirmedOutPointInfos(t *testing.T) {
	setup := func() (Keeper, sdk.Context) {
		encCfg := appParams.MakeEncodingConfig()
		btcSubspace := params.NewSubspace(encCfg.Marshaler, encCfg.Amino, sdk.NewKVStoreKey("params"), sdk.NewKVStoreKey("tparams"), "btc")
		ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
		return NewKeeper(encCfg.Marshaler, sdk.NewKVStoreKey("btc"), btcSubspace), ctx
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
			infoCount := int(rand.I64Between(1, 200))
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
	for i := 0; i < infoCount; i++ {
		info := randOutPointInfo()
		k.SetPendingOutpointInfo(ctx, exported.PollMeta{ID: rand.StrBetween(5, 20)}, info)
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
	for _, state := range rand.Distr(3).Samples(infoCount) {
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
	txHash, err := chainhash.NewHash(rand.Bytes(chainhash.HashSize))
	if err != nil {
		panic(err)
	}
	info := types.OutPointInfo{
		OutPoint: wire.NewOutPoint(txHash, mathRand.Uint32()).String(),
		Amount:   btcutil.Amount(rand.PosI64()),
		Address:  rand.StrBetween(20, 60),
	}
	return info
}
