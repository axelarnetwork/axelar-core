package keeper_test

import (
	"bytes"
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
	bitcoinKeeper "github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
)

func TestKeeper_GetAddress(t *testing.T) {
	var (
		ctx    sdk.Context
		keeper bitcoinKeeper.Keeper
	)
	setup := func() {
		encCfg := appParams.MakeEncodingConfig()
		btcSubspace := params.NewSubspace(encCfg.Marshaler, encCfg.Amino, sdk.NewKVStoreKey("params"), sdk.NewKVStoreKey("tparams"), "btc")
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
		keeper = bitcoinKeeper.NewKeeper(encCfg.Marshaler, sdk.NewKVStoreKey("btc"), btcSubspace)
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
		keeper bitcoinKeeper.Keeper
	)
	setup := func() {
		encCfg := appParams.MakeEncodingConfig()
		btcSubspace := params.NewSubspace(encCfg.Marshaler, encCfg.Amino, sdk.NewKVStoreKey("params"), sdk.NewKVStoreKey("tparams"), "btc")
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
		keeper = bitcoinKeeper.NewKeeper(encCfg.Marshaler, sdk.NewKVStoreKey("btc"), btcSubspace)
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
		keyID := rand.StrBetween(5, 100)

		keeper.SetConfirmedOutpointInfo(ctx, keyID, info)

		_, _, ok := keeper.GetOutPointInfo(ctx, *outpoint)
		assert.True(t, ok)
	}).Repeat(20))
}

func TestScheduleTxs(t *testing.T) {
	t.Run("scheduling unsigned txs", testutils.Func(func(t *testing.T) {
		encCfg := appParams.MakeEncodingConfig()
		btcSubspace := params.NewSubspace(encCfg.Marshaler, encCfg.Amino, sdk.NewKVStoreKey("params"), sdk.NewKVStoreKey("tparams"), "btc")
		ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
		keeper := bitcoinKeeper.NewKeeper(encCfg.Marshaler, sdk.NewKVStoreKey("btc"), btcSubspace)

		numTxs := int(rand.I64Between(10, 30))
		currentHeight := ctx.BlockHeight()
		scheduledHeight := currentHeight + rand.I64Between(10, 30)
		expectedTxs := make([]types.ScheduledUnsignedTx, numTxs)
		counter := rand.I64Between(1, 10)
		roles := []tss.KeyRole{tss.MasterKey, tss.SecondaryKey}

		// schedule txs
		for i := 0; i < numTxs; i++ {
			numOutpoints := int(rand.I64Between(1, 5))
			outpoints := make([]tss.SignInfo, numOutpoints)
			for j := 0; j < numOutpoints; j++ {
				outpoint := tss.SignInfo{
					KeyID:           rand.StrBetween(5, 10),
					SigID:           rand.StrBetween(5, 10),
					Msg:             rand.BytesBetween(50, 100),
					SnapshotCounter: counter + 1,
				}
				outpoints[j] = outpoint
			}

			tx := types.ScheduledUnsignedTx{
				UnsignedTx: types.UnsignedTx{
					Tx: rand.Bytes(20),
				},
				KeyRole:   roles[int(rand.I64Between(0, int64(len(roles))))],
				SignInfos: outpoints,
			}
			keeper.ScheduleUnsignedTx(ctx, scheduledHeight, tx)
			expectedTxs[i] = tx
		}

		// verify txs from above
		ctx = ctx.WithBlockHeight(scheduledHeight)
		txs := keeper.GetScheduledTxs(ctx)

		actualNumTxs := 0
		for _, expected := range expectedTxs {
			for _, actual := range txs {
				bz1, err := expected.Marshal()
				if err != nil {
					panic(err.Error())
				}
				bz2, err := actual.Marshal()
				if err != nil {
					panic(err.Error())
				}
				if bytes.Equal(bz1, bz2) {
					actualNumTxs++
					break
				}
			}
		}

		assert.Len(t, expectedTxs, actualNumTxs)
		assert.Equal(t, numTxs, actualNumTxs)

		// delete scheduled txs
		keeper.DeleteScheduledTxs(ctx)
		assert.Len(t, keeper.GetScheduledTxs(ctx), 0)
	}).Repeat(20))
}
