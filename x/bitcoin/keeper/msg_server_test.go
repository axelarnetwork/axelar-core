package keeper_test

import (
	"crypto/ecdsa"
	cryptoRand "crypto/rand"
	"fmt"
	mathRand "math/rand"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogoprototypes "github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	utilsmock "github.com/axelarnetwork/axelar-core/utils/mock"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	bitcoinKeeper "github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types/mock"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
	votetypes "github.com/axelarnetwork/axelar-core/x/vote/types"
)

func TestHandleMsgLink(t *testing.T) {
	var (
		server      types.MsgServiceServer
		btcKeeper   *mock.BTCKeeperMock
		signer      *mock.SignerMock
		nexusKeeper *mock.NexusMock
		ctx         sdk.Context
		msg         *types.LinkRequest
	)
	setup := func() {
		btcKeeper = &mock.BTCKeeperMock{
			GetNetworkFunc: func(ctx sdk.Context) types.Network { return types.Mainnet },
			SetAddressFunc: func(sdk.Context, types.AddressInfo) {},
			LoggerFunc:     func(sdk.Context) log.Logger { return log.TestingLogger() },
		}
		signer = &mock.SignerMock{GetCurrentKeyFunc: func(_ sdk.Context, _ nexus.Chain, keyRole tss.KeyRole) (tss.Key, bool) {
			sk, _ := ecdsa.GenerateKey(btcec.S256(), cryptoRand.Reader)
			return tss.Key{Value: sk.PublicKey, ID: rand.StrBetween(5, 20), Role: keyRole}, true
		}}
		nexusKeeper = &mock.NexusMock{
			GetChainFunc: func(_ sdk.Context, chain string) (nexus.Chain, bool) {
				return nexus.Chain{
					Name:                  chain,
					NativeAsset:           rand.StrBetween(5, 20),
					SupportsForeignAssets: true,
				}, true
			},
			IsAssetRegisteredFunc: func(sdk.Context, string, string) bool { return true },
			LinkAddressesFunc:     func(sdk.Context, nexus.CrossChainAddress, nexus.CrossChainAddress) {},
		}
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())
		msg = randomMsgLink()
		server = bitcoinKeeper.NewMsgServerImpl(btcKeeper, signer, nexusKeeper, &mock.VoterMock{}, &mock.SnapshotterMock{})
	}
	repeatCount := 20

	t.Run("happy path", testutils.Func(func(t *testing.T) {
		setup()
		res, err := server.Link(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)
		assert.Len(t, btcKeeper.SetAddressCalls(), 1)
		assert.Len(t, nexusKeeper.LinkAddressesCalls(), 1)
		assert.Equal(t, exported.Bitcoin, signer.GetCurrentKeyCalls()[0].Chain)
		assert.Equal(t, msg.RecipientChain, nexusKeeper.GetChainCalls()[0].Chain)
		assert.Equal(t, btcKeeper.SetAddressCalls()[0].Address.Address, res.DepositAddr)
		assert.Equal(t, types.Deposit, btcKeeper.SetAddressCalls()[0].Address.Role)
	}).Repeat(repeatCount))

	t.Run("no master key", testutils.Func(func(t *testing.T) {
		setup()
		signer.GetCurrentKeyFunc = func(sdk.Context, nexus.Chain, tss.KeyRole) (tss.Key, bool) { return tss.Key{}, false }
		_, err := server.Link(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

	t.Run("unknown chain", testutils.Func(func(t *testing.T) {
		setup()
		nexusKeeper.GetChainFunc = func(sdk.Context, string) (nexus.Chain, bool) { return nexus.Chain{}, false }
		_, err := server.Link(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

	t.Run("asset not registered", testutils.Func(func(t *testing.T) {
		setup()
		nexusKeeper.IsAssetRegisteredFunc = func(sdk.Context, string, string) bool { return false }
		_, err := server.Link(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))
}

func TestHandleMsgConfirmOutpoint(t *testing.T) {
	var (
		btcKeeper *mock.BTCKeeperMock
		voter     *mock.VoterMock
		signer    *mock.SignerMock
		ctx       sdk.Context
		msg       *types.ConfirmOutpointRequest
		server    types.MsgServiceServer
	)
	setup := func() {
		address := randomAddress()
		btcKeeper = &mock.BTCKeeperMock{
			GetOutPointInfoFunc: func(sdk.Context, wire.OutPoint) (types.OutPointInfo, types.OutPointState, bool) {
				return types.OutPointInfo{}, 0, false
			},
			GetAddressFunc: func(sdk.Context, string) (types.AddressInfo, bool) {
				return types.AddressInfo{
					Address:      address.EncodeAddress(),
					RedeemScript: rand.Bytes(200),
					Role:         types.Deposit,
					KeyID:        rand.StrBetween(5, 20),
				}, true
			},
			GetRevoteLockingPeriodFunc:        func(sdk.Context) int64 { return int64(mathRand.Uint32()) },
			GetRequiredConfirmationHeightFunc: func(sdk.Context) uint64 { return mathRand.Uint64() },
			SetPendingOutpointInfoFunc:        func(sdk.Context, vote.PollKey, types.OutPointInfo) {},
		}
		voter = &mock.VoterMock{
			InitPollFunc: func(sdk.Context, vote.PollKey, int64, int64, ...utils.Threshold) error { return nil },
		}

		signer = &mock.SignerMock{
			GetCurrentKeyIDFunc: func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (string, bool) {
				return rand.StrBetween(5, 20), true
			},
			GetSnapshotCounterForKeyIDFunc: func(sdk.Context, string) (int64, bool) {
				return rand.PosI64(), true
			},
		}

		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())
		msg = randomMsgConfirmOutpoint()
		msg.OutPointInfo.Address = address.EncodeAddress()
		server = bitcoinKeeper.NewMsgServerImpl(btcKeeper, signer, &mock.NexusMock{}, voter, &mock.SnapshotterMock{})
	}

	repeatCount := 20
	t.Run("happy path deposit", testutils.Func(func(t *testing.T) {
		setup()
		_, err := server.ConfirmOutpoint(sdk.WrapSDKContext(ctx), msg)
		events := ctx.EventManager().ABCIEvents()
		assert.NoError(t, err)
		assert.Len(t, testutils.Events(events).Filter(func(event abci.Event) bool { return event.Type == types.EventTypeOutpointConfirmation }), 1)
		assert.Equal(t, msg.OutPointInfo, btcKeeper.SetPendingOutpointInfoCalls()[0].Info)
		assert.Equal(t, voter.InitPollCalls()[0].Poll, btcKeeper.SetPendingOutpointInfoCalls()[0].Poll)
	}).Repeat(repeatCount))
	t.Run("happy path consolidation", testutils.Func(func(t *testing.T) {
		setup()
		addr, _ := btcKeeper.GetAddress(ctx, msg.OutPointInfo.Address)
		addr.Role = types.Consolidation
		btcKeeper.GetAddressFunc = func(sdk.Context, string) (types.AddressInfo, bool) {
			return addr, true
		}

		_, err := server.ConfirmOutpoint(sdk.WrapSDKContext(ctx), msg)
		events := sdk.UnwrapSDKContext(sdk.WrapSDKContext(ctx)).EventManager().ABCIEvents()
		assert.NoError(t, err)
		assert.Len(t, testutils.Events(events).Filter(func(event abci.Event) bool { return event.Type == types.EventTypeOutpointConfirmation }), 1)
		assert.Equal(t, msg.OutPointInfo, btcKeeper.SetPendingOutpointInfoCalls()[0].Info)
		assert.Equal(t, voter.InitPollCalls()[0].Poll, btcKeeper.SetPendingOutpointInfoCalls()[0].Poll)
	}).Repeat(repeatCount))
	t.Run("already confirmed", testutils.Func(func(t *testing.T) {
		setup()
		btcKeeper.GetOutPointInfoFunc = func(sdk.Context, wire.OutPoint) (types.OutPointInfo, types.OutPointState, bool) {
			return msg.OutPointInfo, types.CONFIRMED, true
		}
		_, err := server.ConfirmOutpoint(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

	t.Run("already spent", testutils.Func(func(t *testing.T) {
		setup()
		btcKeeper.GetOutPointInfoFunc = func(sdk.Context, wire.OutPoint) (types.OutPointInfo, types.OutPointState, bool) {
			return msg.OutPointInfo, types.SPENT, true
		}
		_, err := server.ConfirmOutpoint(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

	t.Run("address unknown", testutils.Func(func(t *testing.T) {
		setup()
		btcKeeper.GetAddressFunc = func(sdk.Context, string) (types.AddressInfo, bool) { return types.AddressInfo{}, false }
		_, err := server.ConfirmOutpoint(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

	t.Run("init poll failed", testutils.Func(func(t *testing.T) {
		setup()
		voter.InitPollFunc = func(sdk.Context, vote.PollKey, int64, int64, ...utils.Threshold) error {
			return fmt.Errorf("poll setup failed")
		}
		_, err := server.ConfirmOutpoint(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))
}

func TestHandleMsgVoteConfirmOutpoint(t *testing.T) {
	var (
		btcKeeper   *mock.BTCKeeperMock
		voter       *mock.VoterMock
		nexusKeeper *mock.NexusMock
		ctx         sdk.Context
		msg         *types.VoteConfirmOutpointRequest
		info        types.OutPointInfo
		server      types.MsgServiceServer

		currentSecondaryKey tss.Key
		depositAddressInfo  types.AddressInfo
	)
	setup := func() {
		address := randomAddress()
		info = randomOutpointInfo()
		msg = randomMsgVoteConfirmOutpoint()
		msg.OutPoint = info.OutPoint
		depositAddressInfo = types.AddressInfo{
			Address:      address.EncodeAddress(),
			RedeemScript: rand.Bytes(200),
			Role:         types.Deposit,
			KeyID:        rand.StrBetween(5, 20),
		}
		btcKeeper = &mock.BTCKeeperMock{
			GetOutPointInfoFunc: func(sdk.Context, wire.OutPoint) (types.OutPointInfo, types.OutPointState, bool) {
				return types.OutPointInfo{}, 0, false
			},
			SetConfirmedOutpointInfoFunc:  func(sdk.Context, string, types.OutPointInfo) {},
			GetPendingOutPointInfoFunc:    func(sdk.Context, vote.PollKey) (types.OutPointInfo, bool) { return info, true },
			DeletePendingOutPointInfoFunc: func(sdk.Context, vote.PollKey) {},
			GetSignedTxFunc:               func(sdk.Context, chainhash.Hash) (*wire.MsgTx, bool) { return nil, false },
			GetAddressFunc: func(sdk.Context, string) (types.AddressInfo, bool) {
				return depositAddressInfo, true
			},
		}
		voter = &mock.VoterMock{
			TallyVoteFunc: func(sdk.Context, sdk.AccAddress, vote.PollKey, codec.ProtoMarshaler) (*votetypes.Poll, error) {
				result, _ := codectypes.NewAnyWithValue(&gogoprototypes.BoolValue{Value: true})

				return &votetypes.Poll{Result: result}, nil
			},
			DeletePollFunc: func(sdk.Context, vote.PollKey) {},
		}
		nexusKeeper = &mock.NexusMock{
			EnqueueForTransferFunc: func(sdk.Context, nexus.CrossChainAddress, sdk.Coin) error { return nil },
		}
		privateKey, _ := ecdsa.GenerateKey(btcec.S256(), cryptoRand.Reader)
		currentSecondaryKey = tss.Key{ID: rand.StrBetween(5, 20), Value: privateKey.PublicKey, Role: tss.MasterKey}
		signerKeeper := &mock.SignerMock{
			GetNextKeyFunc:    func(sdk.Context, nexus.Chain, tss.KeyRole) (tss.Key, bool) { return tss.Key{}, false },
			GetCurrentKeyFunc: func(sdk.Context, nexus.Chain, tss.KeyRole) (tss.Key, bool) { return currentSecondaryKey, true },
			AssignNextKeyFunc: func(sdk.Context, nexus.Chain, tss.KeyRole, string) error { return nil },
		}
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())
		server = bitcoinKeeper.NewMsgServerImpl(btcKeeper, signerKeeper, nexusKeeper, voter, &mock.SnapshotterMock{})
	}

	repeats := 20

	t.Run("happy path confirm deposit to deposit address", testutils.Func(func(t *testing.T) {
		setup()

		_, err := server.VoteConfirmOutpoint(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)
		assert.Len(t, voter.DeletePollCalls(), 1)
		assert.Len(t, btcKeeper.DeletePendingOutPointInfoCalls(), 1)
		assert.Equal(t, info, btcKeeper.SetConfirmedOutpointInfoCalls()[0].Info)
		assert.Equal(t, depositAddressInfo.KeyID, btcKeeper.SetConfirmedOutpointInfoCalls()[0].KeyID)
		assert.Equal(t, info.Address, nexusKeeper.EnqueueForTransferCalls()[0].Sender.Address)
		assert.Equal(t, int64(info.Amount), nexusKeeper.EnqueueForTransferCalls()[0].Amount.Amount.Int64())
	}).Repeat(repeats))

	t.Run("happy path confirm deposit to consolidation address", testutils.Func(func(t *testing.T) {
		setup()
		addr, _ := btcKeeper.GetAddress(ctx, info.Address)
		addr.Role = types.Consolidation
		btcKeeper.GetAddressFunc = func(sdk.Context, string) (types.AddressInfo, bool) {
			return addr, true
		}

		_, err := server.VoteConfirmOutpoint(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)
		assert.Len(t, voter.DeletePollCalls(), 1)
		assert.Len(t, btcKeeper.DeletePendingOutPointInfoCalls(), 1)
		assert.Equal(t, info, btcKeeper.SetConfirmedOutpointInfoCalls()[0].Info)
		assert.Equal(t, depositAddressInfo.KeyID, btcKeeper.SetConfirmedOutpointInfoCalls()[0].KeyID)
		assert.Len(t, nexusKeeper.EnqueueForTransferCalls(), 0)
	}).Repeat(repeats))

	t.Run("happy path confirm deposit to consolidation address in consolidation tx", testutils.Func(func(t *testing.T) {
		setup()
		tx := wire.NewMsgTx(wire.TxVersion)
		hash := tx.TxHash()
		op := wire.NewOutPoint(&hash, info.GetOutPoint().Index)
		info.OutPoint = op.String()
		msg.OutPoint = op.String()
		addr, _ := btcKeeper.GetAddress(ctx, info.Address)
		addr.Role = types.Consolidation
		btcKeeper.GetAddressFunc = func(sdk.Context, string) (types.AddressInfo, bool) {
			return addr, true
		}

		_, err := server.VoteConfirmOutpoint(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)
		assert.Len(t, voter.DeletePollCalls(), 1)
		assert.Len(t, btcKeeper.DeletePendingOutPointInfoCalls(), 1)
		assert.Equal(t, info, btcKeeper.SetConfirmedOutpointInfoCalls()[0].Info)
		assert.Equal(t, depositAddressInfo.KeyID, btcKeeper.SetConfirmedOutpointInfoCalls()[0].KeyID)
		assert.Len(t, nexusKeeper.EnqueueForTransferCalls(), 0)
	}).Repeat(repeats))

	t.Run("happy path confirm deposit to deposit address in consolidation tx", testutils.Func(func(t *testing.T) {
		setup()
		tx := wire.NewMsgTx(wire.TxVersion)
		hash := tx.TxHash()
		op := wire.NewOutPoint(&hash, info.GetOutPoint().Index)
		info.OutPoint = op.String()
		msg.OutPoint = op.String()

		_, err := server.VoteConfirmOutpoint(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)
		assert.Len(t, voter.DeletePollCalls(), 1)
		assert.Len(t, btcKeeper.DeletePendingOutPointInfoCalls(), 1)
		assert.Equal(t, info, btcKeeper.SetConfirmedOutpointInfoCalls()[0].Info)
		assert.Equal(t, depositAddressInfo.KeyID, btcKeeper.SetConfirmedOutpointInfoCalls()[0].KeyID)
		assert.Len(t, nexusKeeper.EnqueueForTransferCalls(), 1)
	}).Repeat(repeats))

	t.Run("happy path reject", testutils.Func(func(t *testing.T) {
		setup()
		voter.TallyVoteFunc =
			func(sdk.Context, sdk.AccAddress, vote.PollKey, codec.ProtoMarshaler) (*votetypes.Poll, error) {
				result, _ := codectypes.NewAnyWithValue(&gogoprototypes.BoolValue{Value: false})

				return &votetypes.Poll{Result: result}, nil
			}

		_, err := server.VoteConfirmOutpoint(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)
		assert.Len(t, voter.DeletePollCalls(), 1)
		assert.Len(t, btcKeeper.DeletePendingOutPointInfoCalls(), 1)
		assert.Len(t, btcKeeper.SetConfirmedOutpointInfoCalls(), 0)
		assert.Len(t, nexusKeeper.EnqueueForTransferCalls(), 0)
	}).Repeat(repeats))

	t.Run("happy path no result yet", testutils.Func(func(t *testing.T) {
		setup()
		voter.TallyVoteFunc =
			func(sdk.Context, sdk.AccAddress, vote.PollKey, codec.ProtoMarshaler) (*votetypes.Poll, error) {
				return &votetypes.Poll{}, nil
			}

		_, err := server.VoteConfirmOutpoint(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)
		assert.Len(t, voter.DeletePollCalls(), 0)
		assert.Len(t, btcKeeper.DeletePendingOutPointInfoCalls(), 0)
		assert.Len(t, btcKeeper.SetConfirmedOutpointInfoCalls(), 0)
		assert.Len(t, nexusKeeper.EnqueueForTransferCalls(), 0)
	}).Repeat(repeats))

	t.Run("happy path poll already completed", testutils.Func(func(t *testing.T) {
		setup()
		btcKeeper.GetPendingOutPointInfoFunc = func(sdk.Context, vote.PollKey) (types.OutPointInfo, bool) {
			return types.OutPointInfo{}, false
		}
		btcKeeper.GetOutPointInfoFunc = func(sdk.Context, wire.OutPoint) (types.OutPointInfo, types.OutPointState, bool) {
			return info, types.CONFIRMED, true
		}

		_, err := server.VoteConfirmOutpoint(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)
		assert.Len(t, voter.DeletePollCalls(), 0)
		assert.Len(t, btcKeeper.DeletePendingOutPointInfoCalls(), 0)
		assert.Len(t, btcKeeper.SetConfirmedOutpointInfoCalls(), 0)
		assert.Len(t, nexusKeeper.EnqueueForTransferCalls(), 0)
	}).Repeat(repeats))

	t.Run("happy path second poll (outpoint already confirmed)", testutils.Func(func(t *testing.T) {
		setup()
		btcKeeper.GetOutPointInfoFunc = func(sdk.Context, wire.OutPoint) (types.OutPointInfo, types.OutPointState, bool) {
			return info, types.CONFIRMED, true
		}

		_, err := server.VoteConfirmOutpoint(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)
		assert.Len(t, voter.DeletePollCalls(), 1)
		assert.Len(t, btcKeeper.DeletePendingOutPointInfoCalls(), 1)
		assert.Len(t, btcKeeper.SetConfirmedOutpointInfoCalls(), 0)
		assert.Len(t, nexusKeeper.EnqueueForTransferCalls(), 0)
	}).Repeat(repeats))

	t.Run("happy path already spent", testutils.Func(func(t *testing.T) {
		setup()
		btcKeeper.GetOutPointInfoFunc = func(sdk.Context, wire.OutPoint) (types.OutPointInfo, types.OutPointState, bool) {
			return info, types.SPENT, true
		}

		_, err := server.VoteConfirmOutpoint(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)
		assert.Len(t, voter.DeletePollCalls(), 1)
		assert.Len(t, btcKeeper.DeletePendingOutPointInfoCalls(), 1)
		assert.Len(t, btcKeeper.SetConfirmedOutpointInfoCalls(), 0)
		assert.Len(t, nexusKeeper.EnqueueForTransferCalls(), 0)
	}).Repeat(repeats))

	t.Run("unknown outpoint", testutils.Func(func(t *testing.T) {
		setup()
		btcKeeper.GetPendingOutPointInfoFunc =
			func(sdk.Context, vote.PollKey) (types.OutPointInfo, bool) { return types.OutPointInfo{}, false }

		_, err := server.VoteConfirmOutpoint(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("tally failed", testutils.Func(func(t *testing.T) {
		setup()
		voter.TallyVoteFunc =
			func(sdk.Context, sdk.AccAddress, vote.PollKey, codec.ProtoMarshaler) (*votetypes.Poll, error) {
				return nil, fmt.Errorf("failed")
			}

		_, err := server.VoteConfirmOutpoint(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("enqueue transfer failed", testutils.Func(func(t *testing.T) {
		setup()
		nexusKeeper.EnqueueForTransferFunc = func(sdk.Context, nexus.CrossChainAddress, sdk.Coin) error {
			return fmt.Errorf("failed")
		}

		_, err := server.VoteConfirmOutpoint(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
	}).Repeat(repeats))
	t.Run("outpoint does not match poll", testutils.Func(func(t *testing.T) {
		setup()
		info = randomOutpointInfo()

		_, err := server.VoteConfirmOutpoint(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
	}).Repeat(repeats))
}

func TestHandleMsgSignPendingTransfers(t *testing.T) {
	var (
		btcKeeper   *mock.BTCKeeperMock
		signer      *mock.SignerMock
		voter       *mock.VoterMock
		nexusKeeper *mock.NexusMock
		snapshotter *mock.SnapshotterMock
		ctx         sdk.Context
		cdc         codec.Marshaler
		msg         *types.SignPendingTransfersRequest
		server      types.MsgServiceServer

		transfers               []nexus.CrossChainTransfer
		transferAmount          int64
		deposits                []types.OutPointInfo
		depositAmount           int64
		minimumWithdrawalAmount btcutil.Amount
	)

	setup := func() {
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())
		cdc = app.MakeEncodingConfig().Marshaler

		// let the minimum start at 2 so the dust limit can still go below
		minimumWithdrawalAmount = btcutil.Amount(rand.I64Between(2, 5000))
		depositAmount, transferAmount = 0, 0
		deposits, transfers = []types.OutPointInfo{}, []nexus.CrossChainTransfer{}

		for i := int64(0); i < rand.I64Between(1, types.DefaultParams().MaxInputCount); i++ {
			deposit := randomOutpointInfo()
			deposits = append(deposits, deposit)
			depositAmount += int64(deposit.Amount)
		}

		for {
			transfer := randomTransfer(int64(minimumWithdrawalAmount), depositAmount)

			if transferAmount+transfer.Asset.Amount.Int64() > depositAmount {
				break
			}

			transfers = append(transfers, transfer)
			transferAmount += transfer.Asset.Amount.Int64()
		}

		dustAmount := make(map[string]btcutil.Amount)
		dequeueCount := 0

		masterPrivateKey, _ := ecdsa.GenerateKey(btcec.S256(), cryptoRand.Reader)
		masterKey := tss.Key{ID: rand.StrBetween(5, 20), Value: masterPrivateKey.PublicKey, Role: tss.MasterKey}
		secondaryPrivateKey, _ := ecdsa.GenerateKey(btcec.S256(), cryptoRand.Reader)
		secondaryKey := tss.Key{ID: rand.StrBetween(5, 20), Value: secondaryPrivateKey.PublicKey, Role: tss.SecondaryKey}

		msg = types.NewSignPendingTransfersRequest(rand.Bytes(sdk.AddrLen), secondaryKey.ID)

		btcKeeper = &mock.BTCKeeperMock{
			GetUnsignedTxFunc: func(sdk.Context) (*types.Transaction, bool) { return nil, false },
			GetNetworkFunc:    func(sdk.Context) types.Network { return types.Mainnet },
			LoggerFunc:        func(sdk.Context) log.Logger { return log.TestingLogger() },
			GetConfirmedOutpointInfoQueueForKeyFunc: func(sdk.Context, string) utils.KVQueue {
				return &utilsmock.KVQueueMock{
					IsEmptyFunc: func() bool { return true },
					DequeueFunc: func(value codec.ProtoMarshaler) bool {
						if dequeueCount >= len(deposits) {
							return false
						}

						cdc.MustUnmarshalBinaryLengthPrefixed(
							cdc.MustMarshalBinaryLengthPrefixed(&deposits[dequeueCount]),
							value,
						)

						dequeueCount++
						return true
					},
				}
			},
			DeleteOutpointInfoFunc:       func(sdk.Context, wire.OutPoint) {},
			SetConfirmedOutpointInfoFunc: func(sdk.Context, string, types.OutPointInfo) {},
			SetSpentOutpointInfoFunc:     func(sdk.Context, types.OutPointInfo) {},
			GetAddressFunc: func(_ sdk.Context, encodedAddress string) (types.AddressInfo, bool) {
				return types.AddressInfo{
					Address:      "",
					RedeemScript: nil,
					KeyID:        secondaryKey.ID,
				}, true
			},
			SetAddressFunc:                 func(sdk.Context, types.AddressInfo) {},
			SetUnsignedTxFunc:              func(sdk.Context, *types.Transaction) {},
			GetMinimumWithdrawalAmountFunc: func(sdk.Context) btcutil.Amount { return minimumWithdrawalAmount },
			GetMaxInputCountFunc:           func(sdk.Context) int64 { return types.DefaultParams().MaxInputCount },
			GetDustAmountFunc: func(ctx sdk.Context, encodeAddr string) btcutil.Amount {
				amount, ok := dustAmount[encodeAddr]
				if !ok {
					return 0
				}
				return amount
			},
			SetDustAmountFunc: func(ctx sdk.Context, encodeAddr string, amount btcutil.Amount) {
				if _, ok := dustAmount[encodeAddr]; !ok {
					dustAmount[encodeAddr] = 0
				}
				dustAmount[encodeAddr] += amount
			},
			DeleteDustAmountFunc: func(ctx sdk.Context, encodeAddr string) {
				delete(dustAmount, encodeAddr)
			},
			GetAnyoneCanSpendAddressFunc: func(ctx sdk.Context) types.AddressInfo {
				return types.NewAnyoneCanSpendAddress(types.DefaultParams().Network)
			},
		}
		nexusKeeper = &mock.NexusMock{
			GetTransfersForChainFunc:   func(sdk.Context, nexus.Chain, nexus.TransferState) []nexus.CrossChainTransfer { return transfers },
			ArchivePendingTransferFunc: func(sdk.Context, nexus.CrossChainTransfer) {},
		}
		signer = &mock.SignerMock{
			GetNextKeyFunc: func(sdk.Context, nexus.Chain, tss.KeyRole) (tss.Key, bool) {
				return tss.Key{}, false
			},
			GetKeyFunc: func(ctx sdk.Context, keyID string) (tss.Key, bool) {
				switch keyID {
				case masterKey.ID:
					return masterKey, true
				case secondaryKey.ID:
					return secondaryKey, true
				default:
					return tss.Key{}, false
				}
			},
			GetCurrentKeyFunc: func(_ sdk.Context, _ nexus.Chain, keyRole tss.KeyRole) (tss.Key, bool) {
				if keyRole == tss.MasterKey {
					return masterKey, true
				}

				return secondaryKey, true
			},
			GetSnapshotCounterForKeyIDFunc: func(sdk.Context, string) (int64, bool) {
				return rand.PosI64(), true
			},
			StartSignFunc: func(sdk.Context, types.InitPoller, string, string, []byte, snapshot.Snapshot) error { return nil },
			AssertMatchesRequirementsFunc: func(sdk.Context, snapshot.Snapshotter, nexus.Chain, string, tss.KeyRole) error {
				return nil
			},
		}
		snapshotter = &mock.SnapshotterMock{
			GetSnapshotFunc: func(_ sdk.Context, counter int64) (snapshot.Snapshot, bool) {
				return snapshot.Snapshot{
					Validators:      []snapshot.Validator{},
					Timestamp:       time.Now(),
					Height:          rand.PosI64(),
					TotalShareCount: sdk.NewInt(rand.PosI64()),
					Counter:         counter,
				}, true
			},
		}
		server = bitcoinKeeper.NewMsgServerImpl(btcKeeper, signer, nexusKeeper, voter, snapshotter)
	}

	repeatCount := 20
	t.Run("happy path more deposits than transfers", testutils.Func(func(t *testing.T) {
		setup()

		_, err := server.SignPendingTransfers(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)
		assert.Len(t, btcKeeper.SetUnsignedTxCalls()[0].Tx.GetTx().TxIn, len(deposits))
		assert.Len(t, btcKeeper.SetUnsignedTxCalls()[0].Tx.GetTx().TxOut, len(transfers)+2) // + consolidation outpoint + anyone-can-spend outpoint
		assert.Len(t, nexusKeeper.ArchivePendingTransferCalls(), len(transfers))
		assert.Len(t, btcKeeper.DeleteOutpointInfoCalls(), len(deposits))
		assert.Len(t, btcKeeper.SetSpentOutpointInfoCalls(), len(deposits))
		assert.Len(t, signer.StartSignCalls(), len(deposits))

	}).Repeat(repeatCount))

	t.Run("happy path consolidation to next secondary key", testutils.Func(func(t *testing.T) {
		setup()
		nextSecondaryKeyID := rand.StrBetween(5, 20)
		msg = types.NewSignPendingTransfersRequest(rand.Bytes(sdk.AddrLen), nextSecondaryKeyID)
		prevGetKey := signer.GetKeyFunc
		pk, _ := ecdsa.GenerateKey(btcec.S256(), cryptoRand.Reader)
		signer.GetKeyFunc = func(ctx sdk.Context, keyID string) (tss.Key, bool) {
			key, ok := prevGetKey(ctx, keyID)
			if !ok {
				return tss.Key{
					ID:    keyID,
					Value: pk.PublicKey,
					Role:  tss.Unknown,
				}, true
			}
			return key, ok
		}

		_, err := server.SignPendingTransfers(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)
		assert.Len(t, btcKeeper.SetUnsignedTxCalls()[0].Tx.GetTx().TxIn, len(deposits))
		assert.Len(t, btcKeeper.SetUnsignedTxCalls()[0].Tx.GetTx().TxOut, len(transfers)+2) // + consolidation outpoint + anyone-can-spend outpoint
		assert.Len(t, nexusKeeper.ArchivePendingTransferCalls(), len(transfers))
		assert.Len(t, btcKeeper.DeleteOutpointInfoCalls(), len(deposits))
		assert.Len(t, btcKeeper.SetSpentOutpointInfoCalls(), len(deposits))
		assert.Len(t, signer.StartSignCalls(), len(deposits))

	}).Repeat(repeatCount))

	t.Run("happy path some wrong recipient addresses", testutils.Func(func(t *testing.T) {
		setup()
		var wrongAddressCount int
		if len(transfers) > 0 {
			wrongAddressCount = int(rand.I64Between(0, int64(len(transfers))))
			for i := 0; i < wrongAddressCount; i++ {
				transfers[i].Recipient.Address = rand.StrBetween(5, 100)
			}
		}

		_, err := server.SignPendingTransfers(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)
		assert.Len(t, btcKeeper.SetUnsignedTxCalls()[0].Tx.GetTx().TxIn, len(deposits))
		assert.Len(t, btcKeeper.SetUnsignedTxCalls()[0].Tx.GetTx().TxOut, len(transfers)-wrongAddressCount+2) // + consolidation outpoint + anyone-can-spend outpoint
		assert.Len(t, nexusKeeper.ArchivePendingTransferCalls(), len(transfers)-wrongAddressCount)
		assert.Len(t, btcKeeper.DeleteOutpointInfoCalls(), len(deposits))
		assert.Len(t, btcKeeper.SetSpentOutpointInfoCalls(), len(deposits))
		assert.Len(t, signer.StartSignCalls(), len(deposits))
	}).Repeat(repeatCount))

	t.Run("happy path transfer to same destination address", testutils.Func(func(t *testing.T) {
		setup()

		// this test case is not interested in less than 2 transfers
		if len(transfers) < 1 {
			return
		}

		var sameAddressCount int
		randAddress := randomAddress()

		sameAddressCount = int(rand.I64Between(1, int64(len(transfers)+1)))
		for i := 0; i < sameAddressCount; i++ {
			transfers[i].Recipient.Address = randAddress.EncodeAddress()
		}

		uniqueTransferCount := len(transfers) - sameAddressCount + 1

		_, err := server.SignPendingTransfers(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)
		assert.Len(t, btcKeeper.SetUnsignedTxCalls()[0].Tx.GetTx().TxIn, len(deposits))
		assert.Len(t, btcKeeper.SetUnsignedTxCalls()[0].Tx.GetTx().TxOut, uniqueTransferCount+2) // + consolidation outpoint + anyone-can-spend output
		assert.Len(t, nexusKeeper.ArchivePendingTransferCalls(), len(transfers))
		assert.Len(t, btcKeeper.DeleteOutpointInfoCalls(), len(deposits))
		assert.Len(t, btcKeeper.SetSpentOutpointInfoCalls(), len(deposits))
		assert.Len(t, signer.StartSignCalls(), len(deposits))
	}).Repeat(repeatCount))

	t.Run("happy path transfer below minimum amount", testutils.Func(func(t *testing.T) {
		setup()
		var belowMinimumCount int
		if len(transfers) > 0 {
			belowMinimumCount = int(rand.I64Between(1, int64(len(transfers)+1)))
			for i := 0; i < belowMinimumCount; i++ {
				transfers[i].Asset.Amount = sdk.NewInt(rand.I64Between(0, int64(minimumWithdrawalAmount)))
			}
		}

		_, err := server.SignPendingTransfers(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)
		assert.Len(t, btcKeeper.SetUnsignedTxCalls()[0].Tx.GetTx().TxIn, len(deposits))
		assert.Len(t, btcKeeper.SetUnsignedTxCalls()[0].Tx.GetTx().TxOut, len(transfers)-belowMinimumCount+2) // + consolidation outpoint + anyone-can-spend output
		assert.Len(t, nexusKeeper.ArchivePendingTransferCalls(), len(transfers))
		assert.Len(t, btcKeeper.DeleteOutpointInfoCalls(), len(deposits))
		assert.Len(t, btcKeeper.SetSpentOutpointInfoCalls(), len(deposits))
		assert.Len(t, signer.StartSignCalls(), len(deposits))
	}).Repeat(repeatCount))

	t.Run("happy path rescuing previously ignored output", testutils.Func(func(t *testing.T) {
		setup()

		dust := make(map[string]btcutil.Amount)
		for i := 0; i < len(transfers); i++ {
			encodeAddr := transfers[i].Recipient.Address
			dustAmount := btcutil.Amount(rand.I64Between(1, int64(minimumWithdrawalAmount)+1)) // exclusive limit
			btcKeeper.SetDustAmountFunc(ctx, encodeAddr, dustAmount)
			dust[encodeAddr] += dustAmount
		}

		_, err := server.SignPendingTransfers(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)
		assert.Len(t, btcKeeper.SetUnsignedTxCalls()[0].Tx.GetTx().TxIn, len(deposits))
		assert.Len(t, btcKeeper.SetUnsignedTxCalls()[0].Tx.GetTx().TxOut, len(transfers)+2) // + consolidation outpoint + anyone-can-spend output
		assert.Len(t, nexusKeeper.ArchivePendingTransferCalls(), len(transfers))
		assert.Len(t, btcKeeper.DeleteOutpointInfoCalls(), len(deposits))
		assert.Len(t, btcKeeper.SetSpentOutpointInfoCalls(), len(deposits))
		assert.Len(t, signer.StartSignCalls(), len(deposits))

		txOut := btcKeeper.SetUnsignedTxCalls()[0].Tx.GetTx().TxOut
		for i := 0; i < len(transfers); i++ {
			encodeAddr := transfers[i].Recipient.Address
			assert.Equal(t, btcKeeper.GetDustAmountFunc(ctx, encodeAddr), btcutil.Amount(0))
			assert.Equal(t, int64(dust[encodeAddr])+transfers[i].Asset.Amount.Int64(), txOut[i+2].Value)
		}

	}).Repeat(repeatCount))

	t.Run("it should include inputs less than or equal to maxInputCount", testutils.Func(func(t *testing.T) {
		setup()

		depositCount := rand.I64Between(types.DefaultParams().MaxInputCount+1, types.DefaultParams().MaxInputCount*100)
		for len(deposits) <= int(depositCount) {
			deposit := randomOutpointInfo()
			deposits = append(deposits, deposit)
			depositAmount += int64(deposit.Amount)
		}

		_, err := server.SignPendingTransfers(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)

		assert.Len(t, btcKeeper.SetUnsignedTxCalls()[0].Tx.GetTx().TxIn, int(types.DefaultParams().MaxInputCount))
		assert.Len(t, btcKeeper.DeleteOutpointInfoCalls(), int(types.DefaultParams().MaxInputCount))
		assert.Len(t, signer.StartSignCalls(), int(types.DefaultParams().MaxInputCount))
	}).Repeat(repeatCount))

	t.Run("deposits == transfers", testutils.Func(func(t *testing.T) {
		setup()
		// equalize deposits and transfers
		transfer := randomTransfer(int64(minimumWithdrawalAmount), 1000000)
		transfer.Asset.Amount = sdk.NewInt(depositAmount - transferAmount)
		transfers = append(transfers, transfer)
		transferAmount += transfer.Asset.Amount.Int64()

		_, err := server.SignPendingTransfers(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

	t.Run("signing already in progress", testutils.Func(func(t *testing.T) {
		setup()
		btcKeeper.GetUnsignedTxFunc = func(sdk.Context) (*types.Transaction, bool) { return &types.Transaction{}, true }

		_, err := server.SignPendingTransfers(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

	t.Run("unknown outpoint address", testutils.Func(func(t *testing.T) {
		setup()
		btcKeeper.GetAddressFunc = func(sdk.Context, string) (types.AddressInfo, bool) { return types.AddressInfo{}, false }

		_, err := server.SignPendingTransfers(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

	t.Run("not enough deposits", testutils.Func(func(t *testing.T) {
		setup()
		transfers = append(transfers, nexus.CrossChainTransfer{
			Recipient: nexus.CrossChainAddress{Chain: exported.Bitcoin, Address: randomAddress().EncodeAddress()},
			Asset:     sdk.NewInt64Coin(exported.Bitcoin.NativeAsset, depositAmount),
			ID:        mathRand.Uint64(),
		})

		_, err := server.SignPendingTransfers(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

	t.Run("no master keys", testutils.Func(func(t *testing.T) {
		setup()
		signer.GetNextKeyFunc = func(sdk.Context, nexus.Chain, tss.KeyRole) (tss.Key, bool) { return tss.Key{}, false }
		signer.GetCurrentKeyFunc = func(sdk.Context, nexus.Chain, tss.KeyRole) (tss.Key, bool) { return tss.Key{}, false }

		_, err := server.SignPendingTransfers(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

	t.Run("no snapshot counter", testutils.Func(func(t *testing.T) {
		setup()
		signer.GetSnapshotCounterForKeyIDFunc = func(sdk.Context, string) (int64, bool) { return 0, false }

		_, err := server.SignPendingTransfers(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

	t.Run("no snapshot", testutils.Func(func(t *testing.T) {
		setup()
		snapshotter.GetSnapshotFunc = func(sdk.Context, int64) (snapshot.Snapshot, bool) { return snapshot.Snapshot{}, false }

		_, err := server.SignPendingTransfers(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

	t.Run("sign fails", testutils.Func(func(t *testing.T) {
		setup()
		signer.StartSignFunc = func(sdk.Context, types.InitPoller, string, string, []byte, snapshot.Snapshot) error {
			return fmt.Errorf("failed")
		}

		_, err := server.SignPendingTransfers(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

}

func mapi(n int, f func(i int)) {
	for i := 0; i < n; i++ {
		f(i)
	}
}

func randomMsgLink() *types.LinkRequest {
	return types.NewLinkRequest(
		rand.Bytes(sdk.AddrLen),
		rand.StrBetween(5, 100),
		rand.StrBetween(5, 100))
}

func randomMsgConfirmOutpoint() *types.ConfirmOutpointRequest {
	return types.NewConfirmOutpointRequest(rand.Bytes(sdk.AddrLen), randomOutpointInfo())
}

func randomMsgVoteConfirmOutpoint() *types.VoteConfirmOutpointRequest {
	return types.NewVoteConfirmOutpointRequest(
		rand.Bytes(sdk.AddrLen),
		vote.PollKey{
			Module: types.ModuleName,
			ID:     rand.StrBetween(5, 20),
		},
		randomOutpointInfo().GetOutPoint(),
		rand.Bools(0.5).Next(),
	)
}

func randomOutpointInfo() types.OutPointInfo {
	txHash, err := chainhash.NewHash(rand.Bytes(chainhash.HashSize))
	if err != nil {
		panic(err)
	}
	vout := mathRand.Uint32()
	if vout == 0 {
		vout++
	}
	return types.OutPointInfo{
		OutPoint: wire.NewOutPoint(txHash, vout).String(),
		Amount:   btcutil.Amount(rand.I64Between(1, 10000000000)),
		Address:  randomAddress().EncodeAddress(),
	}
}

func randomTransfer(lowerAmount int64, upperAmount int64) nexus.CrossChainTransfer {
	return nexus.CrossChainTransfer{
		Recipient: nexus.CrossChainAddress{Chain: exported.Bitcoin, Address: randomAddress().EncodeAddress()},
		Asset:     sdk.NewInt64Coin(exported.Bitcoin.NativeAsset, rand.I64Between(lowerAmount, upperAmount)),
		ID:        mathRand.Uint64(),
	}
}

func randomAddress() *btcutil.AddressWitnessScriptHash {
	addr, err := btcutil.NewAddressWitnessScriptHash(rand.Bytes(32), types.DefaultParams().Network.Params())
	if err != nil {
		panic(err)
	}
	return addr
}
