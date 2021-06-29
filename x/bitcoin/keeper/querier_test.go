package keeper

import (
	"crypto/ecdsa"
	cryptoRand "crypto/rand"
	mathRand "math/rand"
	"testing"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types/mock"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
)

func TestQueryMasterAddress(t *testing.T) {

	var (
		btcKeeper *mock.BTCKeeperMock
		signer    *mock.SignerMock
		ctx       sdk.Context
	)

	setup := func() {
		btcKeeper = &mock.BTCKeeperMock{
			GetNetworkFunc: func(ctx sdk.Context) types.Network { return types.Mainnet },
			GetAddressFunc: func(sdk.Context, string) (types.AddressInfo, bool) {
				return types.AddressInfo{
					Address:      randomAddress().EncodeAddress(),
					RedeemScript: rand.Bytes(200),
					Role:         types.Deposit,
					KeyID:        rand.StrBetween(5, 20),
				}, true
			},
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
		assert := assert.New(t)

		var key tss.Key
		signer = &mock.SignerMock{
			GetCurrentKeyFunc: func(_ sdk.Context, _ nexus.Chain, keyRole tss.KeyRole) (tss.Key, bool) {
				sk, _ := ecdsa.GenerateKey(btcec.S256(), cryptoRand.Reader)
				key = tss.Key{Value: sk.PublicKey, ID: rand.StrBetween(5, 20), Role: keyRole}
				return key, true
			},
		}

		res, err := queryMasterAddress(ctx, btcKeeper, signer)
		assert.NoError(err)

		var resp types.QueryMasterAddressResponse
		err = resp.Unmarshal(res)
		assert.NoError(err)

		assert.Len(btcKeeper.GetAddressCalls(), 1)
		assert.Len(signer.GetCurrentKeyCalls(), 1)

		assert.Equal(btcKeeper.GetAddressCalls()[0].EncodedAddress, resp.MasterAddress)
		assert.Equal(key.ID, resp.MasterKeyId)

	}).Repeat(repeatCount))

	t.Run("no master key", testutils.Func(func(t *testing.T) {
		setup()
		signer.GetCurrentKeyFunc = func(sdk.Context, nexus.Chain, tss.KeyRole) (tss.Key, bool) { return tss.Key{}, false }

		_, err := queryMasterAddress(ctx, btcKeeper, signer)

		assert := assert.New(t)
		assert.Error(err)

	}).Repeat(repeatCount))

	t.Run("master key has no address", testutils.Func(func(t *testing.T) {
		setup()
		btcKeeper.GetAddressFunc = func(sdk.Context, string) (types.AddressInfo, bool) { return types.AddressInfo{}, false }

		_, err := queryMasterAddress(ctx, btcKeeper, signer)

		assert := assert.New(t)
		assert.Error(err)

	}).Repeat(repeatCount))

}

func TestQueryDepositAddress(t *testing.T) {

	var (
		btcKeeper   *mock.BTCKeeperMock
		signer      *mock.SignerMock
		nexusKeeper *mock.NexusMock
		ctx         sdk.Context
		data        []byte
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
					Chain:   exported.Bitcoin,
					Address: randomAddress().EncodeAddress(),
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
		assert.Len(nexusKeeper.GetRecipientCalls(), 1)

		assert.Equal(nexusKeeper.GetChainCalls()[0].Chain, "ethereum")
		assert.Equal(string(res), nexusKeeper.GetRecipientCalls()[0].Sender.Address)

	}).Repeat(repeatCount))

	t.Run("happy path", testutils.Func(func(t *testing.T) {
		setup()
		dataStr := &types.DepositQueryParams{Chain: rand.StrBetween(5, 20), Address: "0x" + rand.HexStr(40)}
		data = types.ModuleCdc.MustMarshalJSON(dataStr)

		res, err := queryDepositAddress(ctx, btcKeeper, signer, nexusKeeper, data)

		assert := assert.New(t)
		assert.NoError(err)
		assert.Len(nexusKeeper.GetRecipientCalls(), 1)

		assert.Equal(nexusKeeper.GetChainCalls()[0].Chain, dataStr.Chain)
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

func TestQueryTxState(t *testing.T) {

	var (
		btcKeeper *mock.BTCKeeperMock
		ctx       sdk.Context
		data      []byte
	)

	setup := func() {
		btcKeeper = &mock.BTCKeeperMock{
			GetOutPointInfoFunc: func(ctx sdk.Context, outpoint wire.OutPoint) (types.OutPointInfo, types.OutPointState, bool) {
				return randomOutpointInfo(), types.CONFIRMED, true
			},
		}
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())

		txHash, err := chainhash.NewHash(rand.Bytes(chainhash.HashSize))
		if err != nil {
			panic(err)
		}
		vout := mathRand.Uint32()
		if vout == 0 {
			vout++
		}
		data = []byte(wire.NewOutPoint(txHash, vout).String())
	}

	repeatCount := 20

	t.Run("happy path", testutils.Func(func(t *testing.T) {
		setup()

		_, err := queryTxState(ctx, btcKeeper, data)

		assert := assert.New(t)
		assert.NoError(err)
		assert.Len(btcKeeper.GetOutPointInfoCalls(), 1)

	}).Repeat(repeatCount))

	t.Run("transaction not found", testutils.Func(func(t *testing.T) {
		setup()
		btcKeeper.GetOutPointInfoFunc = func(ctx sdk.Context, outpoint wire.OutPoint) (types.OutPointInfo, types.OutPointState, bool) {
			return types.OutPointInfo{}, 0, false
		}

		_, err := queryTxState(ctx, btcKeeper, data)

		assert := assert.New(t)
		assert.Error(err)

	}).Repeat(repeatCount))
}

func TestGetRawConsolidationTx(t *testing.T) {

	var (
		btcKeeper *mock.BTCKeeperMock
		ctx       sdk.Context
	)

	setup := func() {
		btcKeeper = &mock.BTCKeeperMock{
			GetUnsignedTxFunc: func(sdk.Context) (*wire.MsgTx, bool) { return nil, false },
			GetSignedTxFunc:   func(sdk.Context) (*wire.MsgTx, bool) { return wire.NewMsgTx(wire.TxVersion), true },
		}
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())
	}

	repeatCount := 20

	t.Run("happy path", testutils.Func(func(t *testing.T) {
		setup()

		_, err := getRawConsolidationTx(ctx, btcKeeper)

		assert := assert.New(t)
		assert.NoError(err)
		assert.Len(btcKeeper.GetUnsignedTxCalls(), 1)
		assert.Len(btcKeeper.GetSignedTxCalls(), 1)

	}).Repeat(repeatCount))

	t.Run("consolidation transaction unsigned", testutils.Func(func(t *testing.T) {
		setup()

		_, err := getRawConsolidationTx(ctx, btcKeeper)
		btcKeeper.GetUnsignedTxFunc = func(sdk.Context) (*wire.MsgTx, bool) { return wire.NewMsgTx(wire.TxVersion), true }
		btcKeeper.GetSignedTxFunc = func(sdk.Context) (*wire.MsgTx, bool) { return nil, false }

		assert := assert.New(t)
		assert.NoError(err)
		assert.Len(btcKeeper.GetUnsignedTxCalls(), 1)

	}).Repeat(repeatCount))

	t.Run("no consolidation transaction", testutils.Func(func(t *testing.T) {
		setup()

		_, err := getRawConsolidationTx(ctx, btcKeeper)
		btcKeeper.GetUnsignedTxFunc = func(sdk.Context) (*wire.MsgTx, bool) { return nil, false }
		btcKeeper.GetSignedTxFunc = func(sdk.Context) (*wire.MsgTx, bool) { return nil, false }

		assert := assert.New(t)
		assert.NoError(err)
		assert.Len(btcKeeper.GetUnsignedTxCalls(), 1)
		assert.Len(btcKeeper.GetSignedTxCalls(), 1)

	}).Repeat(repeatCount))
}

func TestGetConsolidationTxState(t *testing.T) {

	var (
		btcKeeper *mock.BTCKeeperMock
		ctx       sdk.Context
	)

	setup := func() {
		btcKeeper = &mock.BTCKeeperMock{
			GetSignedTxFunc: func(sdk.Context) (*wire.MsgTx, bool) { return wire.NewMsgTx(wire.TxVersion), true },
			GetOutPointInfoFunc: func(ctx sdk.Context, outpoint wire.OutPoint) (types.OutPointInfo, types.OutPointState, bool) {
				return randomOutpointInfo(), types.CONFIRMED, true
			},
			GetMasterKeyVoutFunc: func(sdk.Context) (uint32, bool) {
				vout := mathRand.Uint32()
				if vout == 0 {
					vout++
				}
				return vout, true
			},
		}
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())
	}

	repeatCount := 20

	t.Run("happy path", testutils.Func(func(t *testing.T) {
		setup()

		res, err := getConsolidationTxState(ctx, btcKeeper)

		assert := assert.New(t)
		assert.NoError(err)
		assert.Len(btcKeeper.GetSignedTxCalls(), 1)
		assert.Len(btcKeeper.GetOutPointInfoCalls(), 1)
		assert.Equal(string(res), "bitcoin transaction state is confirmed")

	}).Repeat(repeatCount))

	t.Run("no signed consolidation transaction", testutils.Func(func(t *testing.T) {
		setup()
		btcKeeper.GetSignedTxFunc = func(sdk.Context) (*wire.MsgTx, bool) { return nil, false }

		_, err := getConsolidationTxState(ctx, btcKeeper)

		assert := assert.New(t)
		assert.Error(err)

	}).Repeat(repeatCount))

	t.Run("master key vout not set", testutils.Func(func(t *testing.T) {
		setup()
		btcKeeper.GetMasterKeyVoutFunc = func(sdk.Context) (uint32, bool) { return 0, false }

		_, err := getConsolidationTxState(ctx, btcKeeper)

		assert := assert.New(t)
		assert.Error(err)

	}).Repeat(repeatCount))

	t.Run("consolidation transaction not tracked", testutils.Func(func(t *testing.T) {
		setup()
		btcKeeper.GetOutPointInfoFunc = func(ctx sdk.Context, outpoint wire.OutPoint) (types.OutPointInfo, types.OutPointState, bool) {
			return types.OutPointInfo{}, 0, false
		}

		_, err := getConsolidationTxState(ctx, btcKeeper)

		assert := assert.New(t)
		assert.Error(err)

	}).Repeat(repeatCount))
}
