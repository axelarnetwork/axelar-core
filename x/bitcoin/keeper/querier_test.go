package keeper

import (
	"crypto/ecdsa"
	cryptoRand "crypto/rand"
	"encoding/binary"
	// "encoding/hex"
	// "fmt"
	// "math"
	"testing"

	"github.com/btcsuite/btcd/btcec"
	// "github.com/btcsuite/btcd/btcjson"
	// "github.com/btcsuite/btcd/chaincfg/chainhash"
	// "github.com/btcsuite/btcd/mempool"
	// "github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	// abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	// sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types/mock"
	ethereum "github.com/axelarnetwork/axelar-core/x/ethereum/exported"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
)

func TestQueryMasterAddress(t *testing.T) {

	var (
		btcKeeper   *mock.BTCKeeperMock
		signer      *mock.SignerMock
		ctx         sdk.Context
	)


	setup := func() {
		btcKeeper = &mock.BTCKeeperMock{
			GetNetworkFunc: func(ctx sdk.Context) types.Network { return types.Mainnet },
			GetAddressFunc: func(sdk.Context, string) ( types.AddressInfo, bool) { return types.AddressInfo {}, true },
		}
		signer = &mock.SignerMock{
			GetCurrentKeyFunc: func(_ sdk.Context, _ nexus.Chain, keyRole tss.KeyRole) (tss.Key, bool) {
				sk, _ := ecdsa.GenerateKey(btcec.S256(), cryptoRand.Reader)
				return tss.Key{Value: sk.PublicKey, ID: rand.StrBetween(5, 20), Role: keyRole}, true
			},
		}
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())
	}


	repeatCount := 20

	t.Run("happy path", testutils.Func(func(t *testing.T) {
		setup()

		res, err := queryMasterAddress(ctx, btcKeeper, signer)

		assert := assert.New(t)
		assert.NoError(err)
		assert.Len(signer.GetCurrentKeyCalls(), 1)
		assert.Len(btcKeeper.GetNetworkCalls(), 1)
		assert.Len(btcKeeper.GetAddressCalls(), 1)

		assert.Equal(string(res), btcKeeper.GetAddressCalls()[0].EncodedAddress)

	}).Repeat(repeatCount))

	t.Run("no master key", testutils.Func(func(t *testing.T) {
		setup()
		signer.GetCurrentKeyFunc = func(sdk.Context, nexus.Chain, tss.KeyRole) (tss.Key, bool) { return tss.Key{}, false }

		res, err := queryMasterAddress(ctx, btcKeeper, signer)

		assert := assert.New(t)
		assert.Error(err)
		assert.Nil(res)

	}).Repeat(repeatCount))

	t.Run("master key has no address", testutils.Func(func(t *testing.T) {
		setup()
		btcKeeper.GetAddressFunc = func(sdk.Context, string) ( types.AddressInfo, bool) { return types.AddressInfo {}, false }

		res, err := queryMasterAddress(ctx, btcKeeper, signer)

		assert := assert.New(t)
		assert.Error(err)
		assert.Nil(res)

	}).Repeat(repeatCount))

}


func TestQueryMinimumWithdrawAmount(t *testing.T) {

	var (
		btcKeeper   *mock.BTCKeeperMock
		ctx         sdk.Context
	)


	setup := func() {

		btcKeeper = &mock.BTCKeeperMock{
			GetMinimumWithdrawalAmountFunc: func(ctx sdk.Context) btcutil.Amount {
				var result btcutil.Amount
				// btcKeeper.params.Get(ctx, types.KeyMinimumWithdrawalAmount, &result)
			
				return result
			},
		}
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())
	}


	repeatCount := 20

	t.Run("happy path", testutils.Func(func(t *testing.T) {
		setup()

		kvstore := fake.NewTestKVStore()
		amount := make([]byte, 8)
		binary.LittleEndian.PutUint64(amount, uint64(btcutil.Amount(rand.I64Between(1, 5000))))
		kvstore.Set(types.KeyMinimumWithdrawalAmount, amount)

		_ = queryMinimumWithdrawAmount(ctx, btcKeeper)
		res := kvstore.Get(types.KeyMinimumWithdrawalAmount)

		assert := assert.New(t)
		assert.Len(btcKeeper.GetMinimumWithdrawalAmountCalls(), 1)
		assert.Equal(amount, res)

	}).Repeat(repeatCount))

}


func TestQueryDepositAddress(t *testing.T) {

	var (
		btcKeeper   *mock.BTCKeeperMock
		signer      *mock.SignerMock
		nexusKeeper *mock.NexusMock
		ctx         sdk.Context
		data		[]byte
	)


	setup := func() {

		btcKeeper = &mock.BTCKeeperMock{
			GetNetworkFunc: func(ctx sdk.Context) types.Network { return types.Mainnet },
		}
		signer = &mock.SignerMock{
			GetCurrentKeyFunc: func(_ sdk.Context, _ nexus.Chain, keyRole tss.KeyRole) (tss.Key, bool) {
				sk, _ := ecdsa.GenerateKey(btcec.S256(), cryptoRand.Reader)
				return tss.Key{Value: sk.PublicKey, ID: rand.StrBetween(5, 20), Role: keyRole}, true
			},
		}
		nexusKeeper = &mock.NexusMock{
			GetChainFunc: func(_ sdk.Context, chain string) (nexus.Chain, bool) {
				return nexus.Chain{
					Name:                  chain,
					NativeAsset:           rand.StrBetween(5, 20),
					SupportsForeignAssets: true,
				}, true
			},
			GetRecipientFunc: func(sdk.Context, nexus.CrossChainAddress) (nexus.CrossChainAddress, bool) {
				return nexus.CrossChainAddress{
					Chain:		ethereum.Ethereum,
					Address:	randomAddress().EncodeAddress(),
				}, true
			},
		}
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())
		data = types.ModuleCdc.MustMarshalJSON(&types.DepositQueryParams{Chain: "ethereum", Address: "0xf2151de34BbFb22f799243FFBeFf18FD5D701147"})
	}


	repeatCount := 20

	t.Run("happy path hard coded", testutils.Func(func(t *testing.T) {
		setup()

		res, err := queryDepositAddress(ctx, btcKeeper, signer, nexusKeeper, data)

		assert := assert.New(t)
		assert.NoError(err)
		assert.Len(btcKeeper.GetNetworkCalls(), 1)
		assert.Len(signer.GetCurrentKeyCalls(), 2)
		assert.Len(nexusKeeper.GetChainCalls(), 1)
		assert.Len(nexusKeeper.GetRecipientCalls(), 1)

		assert.Equal("ethereum", nexusKeeper.GetChainCalls()[0].Chain)

		assert.Equal(exported.Bitcoin, signer.GetCurrentKeyCalls()[0].Chain)
		assert.Equal(tss.MasterKey, signer.GetCurrentKeyCalls()[0].KeyRole)

		assert.Equal(exported.Bitcoin, signer.GetCurrentKeyCalls()[1].Chain)
		assert.Equal(tss.SecondaryKey, signer.GetCurrentKeyCalls()[1].KeyRole)

		assert.Equal(string(res), nexusKeeper.GetRecipientCalls()[0].Sender.Address)

	}).Repeat(repeatCount))

	t.Run("happy path", testutils.Func(func(t *testing.T) {
		setup()
		dataStr := &types.DepositQueryParams{Chain: rand.StrBetween(5, 20), Address: "0x" + rand.HexStr(40)}
		data = types.ModuleCdc.MustMarshalJSON(dataStr)

		res, err := queryDepositAddress(ctx, btcKeeper, signer, nexusKeeper, data)

		assert := assert.New(t)
		assert.NoError(err)
		assert.Len(btcKeeper.GetNetworkCalls(), 1)
		assert.Len(signer.GetCurrentKeyCalls(), 2)
		assert.Len(nexusKeeper.GetChainCalls(), 1)
		assert.Len(nexusKeeper.GetRecipientCalls(), 1)

		assert.Equal(dataStr.Chain, nexusKeeper.GetChainCalls()[0].Chain)

		assert.Equal(exported.Bitcoin, signer.GetCurrentKeyCalls()[0].Chain)
		assert.Equal(tss.MasterKey, signer.GetCurrentKeyCalls()[0].KeyRole)

		assert.Equal(exported.Bitcoin, signer.GetCurrentKeyCalls()[1].Chain)
		assert.Equal(tss.SecondaryKey, signer.GetCurrentKeyCalls()[1].KeyRole)

		assert.Equal(string(res), nexusKeeper.GetRecipientCalls()[0].Sender.Address)

	}).Repeat(repeatCount))

	t.Run("cannot parse recipient", testutils.Func(func(t *testing.T) {
		setup()
		data = nil

		_, err := queryDepositAddress(ctx, btcKeeper, signer, nexusKeeper, data)

		assert := assert.New(t)
		assert.Error(err)

	}).Repeat(repeatCount))

	t.Run("recipient chain not found", testutils.Func(func(t *testing.T) {
		setup()
		nexusKeeper.GetChainFunc = func(_ sdk.Context, chain string) (nexus.Chain, bool) {
			return exported.Bitcoin, false
		}

		_, err := queryDepositAddress(ctx, btcKeeper, signer, nexusKeeper, data)

		assert := assert.New(t)
		assert.Error(err)

	}).Repeat(repeatCount))


	t.Run("no master/secondary key", testutils.Func(func(t *testing.T) {
		setup()
		signer.GetCurrentKeyFunc = func(sdk.Context, nexus.Chain, tss.KeyRole) (tss.Key, bool) { return tss.Key{}, false }

		_, err := queryDepositAddress(ctx, btcKeeper, signer, nexusKeeper, data)

		assert := assert.New(t)
		assert.Error(err)

	}).Repeat(repeatCount))

	t.Run("deposit address not linked", testutils.Func(func(t *testing.T) {
		setup()
		nexusKeeper.GetRecipientFunc = func(sdk.Context, nexus.CrossChainAddress) (nexus.CrossChainAddress, bool) {
			return nexus.CrossChainAddress{}, false
		}

		_, err := queryDepositAddress(ctx, btcKeeper, signer, nexusKeeper, data)

		assert := assert.New(t)
		assert.Error(err)

	}).Repeat(repeatCount))
	
}
