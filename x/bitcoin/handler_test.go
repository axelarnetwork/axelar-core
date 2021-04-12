package bitcoin

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
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/go-amino"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types/mock"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

func TestHandleMsgLink(t *testing.T) {
	var (
		btcKeeper   *mock.BTCKeeperMock
		signer      *mock.SignerMock
		nexusKeeper *mock.NexusMock
		ctx         sdk.Context
		msg         types.MsgLink
	)
	setup := func() {
		btcKeeper = &mock.BTCKeeperMock{
			GetNetworkFunc: func(ctx sdk.Context) types.Network { return types.Mainnet },
			SetAddressFunc: func(sdk.Context, types.AddressInfo) {},
			LoggerFunc:     func(sdk.Context) log.Logger { return log.TestingLogger() },
		}
		signer = &mock.SignerMock{GetCurrentKeyFunc: func(sdk.Context, nexus.Chain, tss.KeyRole) (tss.Key, bool) {
			sk, _ := ecdsa.GenerateKey(btcec.S256(), cryptoRand.Reader)
			return tss.Key{Value: sk.PublicKey, ID: rand.StrBetween(5, 20)}, true
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
		ctx = sdk.NewContext(nil, abci.Header{Height: rand.PosI64()}, false, log.TestingLogger())
		msg = randomMsgLink()
	}
	repeatCount := 20

	t.Run("happy path", testutils.Func(func(t *testing.T) {
		setup()
		res, err := HandleMsgLink(ctx, btcKeeper, signer, nexusKeeper, msg)
		assert.NoError(t, err)
		assert.Len(t, btcKeeper.SetAddressCalls(), 1)
		assert.Len(t, nexusKeeper.LinkAddressesCalls(), 1)
		assert.Equal(t, exported.Bitcoin, signer.GetCurrentKeyCalls()[0].Chain)
		assert.Equal(t, msg.RecipientChain, nexusKeeper.GetChainCalls()[0].Chain)
		assert.Equal(t, btcKeeper.SetAddressCalls()[0].Address.Address.EncodeAddress(), string(res.Data))
	}).Repeat(repeatCount))

	t.Run("no master key", testutils.Func(func(t *testing.T) {
		setup()
		signer.GetCurrentKeyFunc = func(sdk.Context, nexus.Chain, tss.KeyRole) (tss.Key, bool) { return tss.Key{}, false }
		_, err := HandleMsgLink(ctx, btcKeeper, signer, nexusKeeper, msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

	t.Run("unknown chain", testutils.Func(func(t *testing.T) {
		setup()
		nexusKeeper.GetChainFunc = func(sdk.Context, string) (nexus.Chain, bool) { return nexus.Chain{}, false }
		_, err := HandleMsgLink(ctx, btcKeeper, signer, nexusKeeper, msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

	t.Run("asset not registered", testutils.Func(func(t *testing.T) {
		setup()
		nexusKeeper.IsAssetRegisteredFunc = func(sdk.Context, string, string) bool { return false }
		_, err := HandleMsgLink(ctx, btcKeeper, signer, nexusKeeper, msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))
}

func TestHandleMsgConfirmOutpoint(t *testing.T) {
	var (
		btcKeeper *mock.BTCKeeperMock
		voter     *mock.VoterMock
		signer    *mock.SignerMock
		ctx       sdk.Context
		msg       types.MsgConfirmOutpoint
	)
	setup := func() {
		address := randomAddress()
		btcKeeper = &mock.BTCKeeperMock{
			GetOutPointInfoFunc: func(sdk.Context, wire.OutPoint) (types.OutPointInfo, types.OutPointState, bool) {
				return types.OutPointInfo{}, 0, false
			},
			GetAddressFunc: func(sdk.Context, string) (types.AddressInfo, bool) {
				return types.AddressInfo{
					Address:      address,
					RedeemScript: rand.Bytes(200),
					Key: tss.Key{
						ID:    rand.StrBetween(5, 20),
						Value: ecdsa.PublicKey{},
					},
				}, true
			},
			GetRevoteLockingPeriodFunc:        func(sdk.Context) int64 { return int64(mathRand.Uint32()) },
			GetRequiredConfirmationHeightFunc: func(sdk.Context) uint64 { return mathRand.Uint64() },
			SetPendingOutpointInfoFunc:        func(sdk.Context, vote.PollMeta, types.OutPointInfo) {},
			CodecFunc:                         func() *amino.Codec { return testutils.Codec() },
		}
		voter = &mock.VoterMock{
			InitPollFunc: func(sdk.Context, vote.PollMeta, int64) error { return nil },
		}

		signer = &mock.SignerMock{
			GetCurrentKeyIDFunc: func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (string, bool) {
				return rand.StrBetween(5, 20), true
			},
			GetSnapshotCounterForKeyIDFunc: func(sdk.Context, string) (int64, bool) {
				return rand.PosI64(), true
			},
		}

		ctx = sdk.NewContext(nil, abci.Header{Height: rand.PosI64()}, false, log.TestingLogger())
		msg = randomMsgConfirmOutpoint()
		msg.OutPointInfo.Address = address.EncodeAddress()
	}

	repeatCount := 20
	t.Run("happy path outpoint", testutils.Func(func(t *testing.T) {
		setup()
		res, err := HandleMsgConfirmOutpoint(ctx, btcKeeper, voter, signer, msg)
		assert.NoError(t, err)
		assert.Len(t, testutils.Events(res.Events).Filter(func(event sdk.Event) bool { return event.Type == types.EventTypeOutpointConfirmation }), 1)
		assert.Equal(t, msg.OutPointInfo, btcKeeper.SetPendingOutpointInfoCalls()[0].Info)
		assert.Equal(t, voter.InitPollCalls()[0].Poll, btcKeeper.SetPendingOutpointInfoCalls()[0].Poll)
	}).Repeat(repeatCount))

	t.Run("already confirmed", testutils.Func(func(t *testing.T) {
		setup()
		btcKeeper.GetOutPointInfoFunc = func(sdk.Context, wire.OutPoint) (types.OutPointInfo, types.OutPointState, bool) {
			return msg.OutPointInfo, types.CONFIRMED, true
		}
		_, err := HandleMsgConfirmOutpoint(ctx, btcKeeper, voter, signer, msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

	t.Run("already spent", testutils.Func(func(t *testing.T) {
		setup()
		btcKeeper.GetOutPointInfoFunc = func(sdk.Context, wire.OutPoint) (types.OutPointInfo, types.OutPointState, bool) {
			return msg.OutPointInfo, types.SPENT, true
		}
		_, err := HandleMsgConfirmOutpoint(ctx, btcKeeper, voter, signer, msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

	t.Run("address unknown", testutils.Func(func(t *testing.T) {
		setup()
		btcKeeper.GetAddressFunc = func(sdk.Context, string) (types.AddressInfo, bool) { return types.AddressInfo{}, false }
		_, err := HandleMsgConfirmOutpoint(ctx, btcKeeper, voter, signer, msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

	t.Run("init poll failed", testutils.Func(func(t *testing.T) {
		setup()
		voter.InitPollFunc = func(sdk.Context, vote.PollMeta, int64) error { return fmt.Errorf("poll setup failed") }
		_, err := HandleMsgConfirmOutpoint(ctx, btcKeeper, voter, signer, msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))
}

func TestHandleMsgVoteConfirmOutpoint(t *testing.T) {
	var (
		btcKeeper   *mock.BTCKeeperMock
		voter       *mock.VoterMock
		nexusKeeper *mock.NexusMock
		ctx         sdk.Context
		msg         types.MsgVoteConfirmOutpoint
		info        types.OutPointInfo
	)
	setup := func() {
		info = randomOutpointInfo()
		msg = randomMsgVoteConfirmOutpoint()
		msg.OutPoint = *info.OutPoint
		btcKeeper = &mock.BTCKeeperMock{
			GetOutPointInfoFunc: func(sdk.Context, wire.OutPoint) (types.OutPointInfo, types.OutPointState, bool) {
				return types.OutPointInfo{}, 0, false
			},
			SetOutpointInfoFunc:           func(sdk.Context, types.OutPointInfo, types.OutPointState) {},
			GetPendingOutPointInfoFunc:    func(sdk.Context, vote.PollMeta) (types.OutPointInfo, bool) { return info, true },
			DeletePendingOutPointInfoFunc: func(sdk.Context, vote.PollMeta) {},
			CodecFunc:                     func() *amino.Codec { return testutils.Codec() },
			GetSignedTxFunc:               func(sdk.Context) (*wire.MsgTx, bool) { return nil, false },
		}
		voter = &mock.VoterMock{
			TallyVoteFunc:  func(sdk.Context, sdk.AccAddress, vote.PollMeta, vote.VotingData) error { return nil },
			ResultFunc:     func(sdk.Context, vote.PollMeta) vote.VotingData { return true },
			DeletePollFunc: func(sdk.Context, vote.PollMeta) {},
		}
		nexusKeeper = &mock.NexusMock{
			EnqueueForTransferFunc: func(sdk.Context, nexus.CrossChainAddress, sdk.Coin) error { return nil },
		}

		ctx = sdk.NewContext(nil, abci.Header{Height: rand.PosI64()}, false, log.TestingLogger())
	}

	repeats := 20
	t.Run("happy path confirm deposit", testutils.Func(func(t *testing.T) {
		setup()

		_, err := HandleMsgVoteConfirmOutpoint(ctx, btcKeeper, voter, nexusKeeper, msg)
		assert.NoError(t, err)
		assert.Len(t, voter.DeletePollCalls(), 1)
		assert.Len(t, btcKeeper.DeletePendingOutPointInfoCalls(), 1)
		assert.Equal(t, info, btcKeeper.SetOutpointInfoCalls()[0].Info)
		assert.Equal(t, types.CONFIRMED, btcKeeper.SetOutpointInfoCalls()[0].State)
		assert.Len(t, btcKeeper.DeleteSignedTxCalls(), 0)
		assert.Equal(t, info.Address, nexusKeeper.EnqueueForTransferCalls()[0].Sender.Address)
		assert.Equal(t, int64(info.Amount), nexusKeeper.EnqueueForTransferCalls()[0].Amount.Amount.Int64())
	}).Repeat(repeats))

	t.Run("happy path confirm consolidation", testutils.Func(func(t *testing.T) {
		setup()
		tx := wire.NewMsgTx(wire.TxVersion)
		info.OutPoint.Hash = tx.TxHash()
		msg.OutPoint.Hash = tx.TxHash()
		btcKeeper.GetSignedTxFunc = func(sdk.Context) (*wire.MsgTx, bool) { return tx, true }
		btcKeeper.DeleteSignedTxFunc = func(sdk.Context) {}

		_, err := HandleMsgVoteConfirmOutpoint(ctx, btcKeeper, voter, nexusKeeper, msg)
		assert.NoError(t, err)
		assert.Len(t, voter.DeletePollCalls(), 1)
		assert.Len(t, btcKeeper.DeletePendingOutPointInfoCalls(), 1)
		assert.Equal(t, info, btcKeeper.SetOutpointInfoCalls()[0].Info)
		assert.Equal(t, types.CONFIRMED, btcKeeper.SetOutpointInfoCalls()[0].State)
		assert.Len(t, btcKeeper.DeleteSignedTxCalls(), 1)
		assert.Len(t, nexusKeeper.EnqueueForTransferCalls(), 0)
	}).Repeat(repeats))

	t.Run("happy path reject", testutils.Func(func(t *testing.T) {
		setup()
		voter.ResultFunc = func(sdk.Context, vote.PollMeta) vote.VotingData { return false }

		_, err := HandleMsgVoteConfirmOutpoint(ctx, btcKeeper, voter, nexusKeeper, msg)
		assert.NoError(t, err)
		assert.Len(t, voter.DeletePollCalls(), 1)
		assert.Len(t, btcKeeper.DeletePendingOutPointInfoCalls(), 1)
		assert.Len(t, btcKeeper.SetOutpointInfoCalls(), 0)
		assert.Len(t, nexusKeeper.EnqueueForTransferCalls(), 0)
		assert.Len(t, btcKeeper.DeleteSignedTxCalls(), 0)
	}).Repeat(repeats))

	t.Run("happy path no result yet", testutils.Func(func(t *testing.T) {
		setup()
		voter.ResultFunc = func(sdk.Context, vote.PollMeta) vote.VotingData { return nil }

		_, err := HandleMsgVoteConfirmOutpoint(ctx, btcKeeper, voter, nexusKeeper, msg)
		assert.NoError(t, err)
		assert.Len(t, voter.DeletePollCalls(), 0)
		assert.Len(t, btcKeeper.DeletePendingOutPointInfoCalls(), 0)
		assert.Len(t, btcKeeper.SetOutpointInfoCalls(), 0)
		assert.Len(t, nexusKeeper.EnqueueForTransferCalls(), 0)
		assert.Len(t, btcKeeper.DeleteSignedTxCalls(), 0)
	}).Repeat(repeats))

	t.Run("happy path poll already completed", testutils.Func(func(t *testing.T) {
		setup()
		btcKeeper.GetPendingOutPointInfoFunc = func(sdk.Context, vote.PollMeta) (types.OutPointInfo, bool) {
			return types.OutPointInfo{}, false
		}
		btcKeeper.GetOutPointInfoFunc = func(sdk.Context, wire.OutPoint) (types.OutPointInfo, types.OutPointState, bool) {
			return info, types.CONFIRMED, true
		}

		_, err := HandleMsgVoteConfirmOutpoint(ctx, btcKeeper, voter, nexusKeeper, msg)
		assert.NoError(t, err)
		assert.Len(t, voter.DeletePollCalls(), 0)
		assert.Len(t, btcKeeper.DeletePendingOutPointInfoCalls(), 0)
		assert.Len(t, btcKeeper.SetOutpointInfoCalls(), 0)
		assert.Len(t, nexusKeeper.EnqueueForTransferCalls(), 0)
		assert.Len(t, btcKeeper.DeleteSignedTxCalls(), 0)
	}).Repeat(repeats))

	t.Run("happy path second poll (outpoint already confirmed)", testutils.Func(func(t *testing.T) {
		setup()
		btcKeeper.GetOutPointInfoFunc = func(sdk.Context, wire.OutPoint) (types.OutPointInfo, types.OutPointState, bool) {
			return info, types.CONFIRMED, true
		}

		_, err := HandleMsgVoteConfirmOutpoint(ctx, btcKeeper, voter, nexusKeeper, msg)
		assert.NoError(t, err)
		assert.Len(t, voter.DeletePollCalls(), 1)
		assert.Len(t, btcKeeper.DeletePendingOutPointInfoCalls(), 1)
		assert.Len(t, btcKeeper.SetOutpointInfoCalls(), 0)
		assert.Len(t, nexusKeeper.EnqueueForTransferCalls(), 0)
		assert.Len(t, btcKeeper.DeleteSignedTxCalls(), 0)
	}).Repeat(repeats))

	t.Run("happy path already spent", testutils.Func(func(t *testing.T) {
		setup()
		btcKeeper.GetOutPointInfoFunc = func(sdk.Context, wire.OutPoint) (types.OutPointInfo, types.OutPointState, bool) {
			return info, types.SPENT, true
		}

		_, err := HandleMsgVoteConfirmOutpoint(ctx, btcKeeper, voter, nexusKeeper, msg)
		assert.NoError(t, err)
		assert.Len(t, voter.DeletePollCalls(), 1)
		assert.Len(t, btcKeeper.DeletePendingOutPointInfoCalls(), 1)
		assert.Len(t, btcKeeper.SetOutpointInfoCalls(), 0)
		assert.Len(t, nexusKeeper.EnqueueForTransferCalls(), 0)
		assert.Len(t, btcKeeper.DeleteSignedTxCalls(), 0)
	}).Repeat(repeats))

	t.Run("unknown outpoint", testutils.Func(func(t *testing.T) {
		setup()
		btcKeeper.GetPendingOutPointInfoFunc =
			func(sdk.Context, vote.PollMeta) (types.OutPointInfo, bool) { return types.OutPointInfo{}, false }

		_, err := HandleMsgVoteConfirmOutpoint(ctx, btcKeeper, voter, nexusKeeper, msg)
		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("tally failed", testutils.Func(func(t *testing.T) {
		setup()
		voter.TallyVoteFunc = func(sdk.Context, sdk.AccAddress, vote.PollMeta, vote.VotingData) error {
			return fmt.Errorf("failed")
		}

		_, err := HandleMsgVoteConfirmOutpoint(ctx, btcKeeper, voter, nexusKeeper, msg)
		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("enqueue transfer failed", testutils.Func(func(t *testing.T) {
		setup()
		nexusKeeper.EnqueueForTransferFunc = func(sdk.Context, nexus.CrossChainAddress, sdk.Coin) error {
			return fmt.Errorf("failed")
		}

		_, err := HandleMsgVoteConfirmOutpoint(ctx, btcKeeper, voter, nexusKeeper, msg)
		assert.Error(t, err)
	}).Repeat(repeats))
	t.Run("outpoint does not match poll", testutils.Func(func(t *testing.T) {
		setup()
		info = randomOutpointInfo()

		_, err := HandleMsgVoteConfirmOutpoint(ctx, btcKeeper, voter, nexusKeeper, msg)
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
		msg         types.MsgSignPendingTransfers

		transfers      []nexus.CrossChainTransfer
		transferAmount int64
		deposits       []types.OutPointInfo
		depositAmount  int64
	)

	setup := func() {
		ctx = sdk.NewContext(nil, abci.Header{Height: rand.PosI64()}, false, log.TestingLogger())
		msg = types.MsgSignPendingTransfers{
			Fee: btcutil.Amount(rand.I64Between(0, 1000000)),
		}

		transferAmount = 0
		transfers = []nexus.CrossChainTransfer{}
		for i := int64(0); i < rand.I64Between(0, 50); i++ {
			transfers = append(transfers, randomTransfer())
			transferAmount += transfers[i].Asset.Amount.Int64()
		}
		depositAmount = 0
		deposits = []types.OutPointInfo{}
		for depositAmount <= transferAmount+int64(msg.Fee) {
			deposit := randomOutpointInfo()
			deposits = append(deposits, deposit)
			depositAmount += int64(deposit.Amount)
		}

		masterPrivateKey, _ := ecdsa.GenerateKey(btcec.S256(), cryptoRand.Reader)
		masterKey := tss.Key{ID: rand.StrBetween(5, 20), Value: masterPrivateKey.PublicKey}
		secondaryPrivateKey, _ := ecdsa.GenerateKey(btcec.S256(), cryptoRand.Reader)
		secondaryKey := tss.Key{ID: rand.StrBetween(5, 20), Value: secondaryPrivateKey.PublicKey}

		btcKeeper = &mock.BTCKeeperMock{
			GetUnsignedTxFunc:             func(sdk.Context) (*wire.MsgTx, bool) { return nil, false },
			GetSignedTxFunc:               func(sdk.Context) (*wire.MsgTx, bool) { return nil, false },
			GetNetworkFunc:                func(sdk.Context) types.Network { return types.Mainnet },
			LoggerFunc:                    func(sdk.Context) log.Logger { return log.TestingLogger() },
			GetConfirmedOutPointInfosFunc: func(sdk.Context) []types.OutPointInfo { return deposits },
			DeleteOutpointInfoFunc:        func(sdk.Context, wire.OutPoint) {},
			SetOutpointInfoFunc:           func(sdk.Context, types.OutPointInfo, types.OutPointState) {},
			DoesMasterKeyUtxoExistFunc:    func(sdk.Context) bool { return false },
			SetMasterKeyUtxoExistsFunc:    func(sdk.Context) {},
			GetAddressFunc: func(_ sdk.Context, encodedAddress string) (types.AddressInfo, bool) {
				sk, _ := ecdsa.GenerateKey(btcec.S256(), cryptoRand.Reader)
				return types.AddressInfo{
					Address:      nil,
					RedeemScript: nil,
					Key: tss.Key{
						ID:    secondaryKey.ID,
						Value: sk.PublicKey,
					},
				}, true
			},
			SetAddressFunc:    func(sdk.Context, types.AddressInfo) {},
			SetUnsignedTxFunc: func(sdk.Context, *wire.MsgTx) {},
		}
		nexusKeeper = &mock.NexusMock{
			GetPendingTransfersForChainFunc: func(sdk.Context, nexus.Chain) []nexus.CrossChainTransfer { return transfers },
			ArchivePendingTransferFunc:      func(sdk.Context, nexus.CrossChainTransfer) {},
		}
		signer = &mock.SignerMock{
			GetNextKeyFunc: func(sdk.Context, nexus.Chain, tss.KeyRole) (tss.Key, bool) {
				return tss.Key{}, false
			},
			GetKeyRoleFunc: func(ctx sdk.Context, keyID string) (tss.KeyRole, bool) {
				switch keyID {
				case masterKey.ID:
					return tss.MasterKey, true
				case secondaryKey.ID:
					return tss.SecondaryKey, true
				default:
					return -1, false
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
		}
		snapshotter = &mock.SnapshotterMock{
			GetSnapshotFunc: func(_ sdk.Context, counter int64) (snapshot.Snapshot, bool) {
				return snapshot.Snapshot{
					Validators: []snapshot.Validator{},
					Timestamp:  time.Now(),
					Height:     rand.PosI64(),
					TotalPower: sdk.NewInt(rand.PosI64()),
					Counter:    counter,
				}, true
			},
		}

	}

	repeatCount := 20
	t.Run("happy path more deposits than transfers", testutils.Func(func(t *testing.T) {
		setup()

		_, err := HandleMsgSignPendingTransfers(ctx, btcKeeper, signer, nexusKeeper, snapshotter, voter, msg)
		assert.NoError(t, err)
		assert.Len(t, btcKeeper.SetUnsignedTxCalls()[0].Tx.TxIn, len(deposits))
		assert.Len(t, btcKeeper.SetUnsignedTxCalls()[0].Tx.TxOut, len(transfers)+1) // + consolidation outpoint
		assert.Len(t, nexusKeeper.ArchivePendingTransferCalls(), len(transfers))
		assert.Len(t, btcKeeper.DeleteOutpointInfoCalls(), len(deposits))
		assert.Len(t, btcKeeper.SetOutpointInfoCalls(), len(deposits))
		assert.Len(t, btcKeeper.SetMasterKeyUtxoExistsCalls(), 1)
		mapi(len(btcKeeper.SetOutpointInfoCalls()), func(i int) { assert.Equal(t, types.SPENT, btcKeeper.SetOutpointInfoCalls()[i].State) })
		assert.Len(t, signer.StartSignCalls(), len(deposits))

	}).Repeat(repeatCount))

	t.Run("happy path consolidation to next master key", testutils.Func(func(t *testing.T) {
		setup()
		signer.GetNextKeyFunc = signer.GetCurrentKeyFunc

		_, err := HandleMsgSignPendingTransfers(ctx, btcKeeper, signer, nexusKeeper, snapshotter, voter, msg)
		assert.NoError(t, err)
		assert.Len(t, signer.GetCurrentKeyCalls(), 0)
		assert.Len(t, btcKeeper.SetUnsignedTxCalls()[0].Tx.TxIn, len(deposits))
		assert.Len(t, btcKeeper.SetUnsignedTxCalls()[0].Tx.TxOut, len(transfers)+1) // + 1 consolidation outpoint
		assert.Len(t, nexusKeeper.ArchivePendingTransferCalls(), len(transfers))
		assert.Len(t, btcKeeper.DeleteOutpointInfoCalls(), len(deposits))
		assert.Len(t, btcKeeper.SetOutpointInfoCalls(), len(deposits))
		assert.Len(t, btcKeeper.SetMasterKeyUtxoExistsCalls(), 1)
		mapi(len(btcKeeper.SetOutpointInfoCalls()), func(i int) { assert.Equal(t, types.SPENT, btcKeeper.SetOutpointInfoCalls()[i].State) })
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

		_, err := HandleMsgSignPendingTransfers(ctx, btcKeeper, signer, nexusKeeper, snapshotter, voter, msg)
		assert.NoError(t, err)
		assert.Len(t, btcKeeper.SetUnsignedTxCalls()[0].Tx.TxIn, len(deposits))
		assert.Len(t, btcKeeper.SetUnsignedTxCalls()[0].Tx.TxOut, len(transfers)-wrongAddressCount+1) // + 1 consolidation outpoint
		assert.Len(t, nexusKeeper.ArchivePendingTransferCalls(), len(transfers)-wrongAddressCount)
		assert.Len(t, btcKeeper.DeleteOutpointInfoCalls(), len(deposits))
		assert.Len(t, btcKeeper.SetOutpointInfoCalls(), len(deposits))
		assert.Len(t, btcKeeper.SetMasterKeyUtxoExistsCalls(), 1)
		mapi(len(btcKeeper.SetOutpointInfoCalls()), func(i int) { assert.Equal(t, types.SPENT, btcKeeper.SetOutpointInfoCalls()[i].State) })
		assert.Len(t, signer.StartSignCalls(), len(deposits))
	}).Repeat(repeatCount))

	t.Run("deposits == transfers", testutils.Func(func(t *testing.T) {
		setup()
		// equalize deposits and transfers
		transfer := randomTransfer()
		transfer.Asset.Amount = sdk.NewInt(depositAmount - transferAmount - int64(msg.Fee))
		transfers = append(transfers, transfer)
		transferAmount += transfer.Asset.Amount.Int64()

		_, err := HandleMsgSignPendingTransfers(ctx, btcKeeper, signer, nexusKeeper, snapshotter, voter, msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

	t.Run("signing already in progress", testutils.Func(func(t *testing.T) {
		setup()
		btcKeeper.GetUnsignedTxFunc = func(sdk.Context) (*wire.MsgTx, bool) { return wire.NewMsgTx(wire.TxVersion), true }

		_, err := HandleMsgSignPendingTransfers(ctx, btcKeeper, signer, nexusKeeper, snapshotter, voter, msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

	t.Run("previous tx not confirmed", testutils.Func(func(t *testing.T) {
		setup()
		btcKeeper.GetSignedTxFunc = func(sdk.Context) (*wire.MsgTx, bool) { return wire.NewMsgTx(wire.TxVersion), true }

		_, err := HandleMsgSignPendingTransfers(ctx, btcKeeper, signer, nexusKeeper, snapshotter, voter, msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

	t.Run("unknown outpoint address", testutils.Func(func(t *testing.T) {
		setup()
		btcKeeper.GetAddressFunc = func(sdk.Context, string) (types.AddressInfo, bool) { return types.AddressInfo{}, false }

		_, err := HandleMsgSignPendingTransfers(ctx, btcKeeper, signer, nexusKeeper, snapshotter, voter, msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

	t.Run("not enough deposits", testutils.Func(func(t *testing.T) {
		setup()
		deposits = deposits[:len(deposits)-1]

		_, err := HandleMsgSignPendingTransfers(ctx, btcKeeper, signer, nexusKeeper, snapshotter, voter, msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

	t.Run("no master keys", testutils.Func(func(t *testing.T) {
		setup()
		signer.GetNextKeyFunc = func(sdk.Context, nexus.Chain, tss.KeyRole) (tss.Key, bool) { return tss.Key{}, false }
		signer.GetCurrentKeyFunc = func(sdk.Context, nexus.Chain, tss.KeyRole) (tss.Key, bool) { return tss.Key{}, false }

		_, err := HandleMsgSignPendingTransfers(ctx, btcKeeper, signer, nexusKeeper, snapshotter, voter, msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

	t.Run("no snapshot counter", testutils.Func(func(t *testing.T) {
		setup()
		signer.GetSnapshotCounterForKeyIDFunc = func(sdk.Context, string) (int64, bool) { return 0, false }

		_, err := HandleMsgSignPendingTransfers(ctx, btcKeeper, signer, nexusKeeper, snapshotter, voter, msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

	t.Run("no snapshot", testutils.Func(func(t *testing.T) {
		setup()
		snapshotter.GetSnapshotFunc = func(sdk.Context, int64) (snapshot.Snapshot, bool) { return snapshot.Snapshot{}, false }

		_, err := HandleMsgSignPendingTransfers(ctx, btcKeeper, signer, nexusKeeper, snapshotter, voter, msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

	t.Run("sign fails", testutils.Func(func(t *testing.T) {
		setup()
		signer.StartSignFunc = func(sdk.Context, types.InitPoller, string, string, []byte, snapshot.Snapshot) error {
			return fmt.Errorf("failed")
		}

		_, err := HandleMsgSignPendingTransfers(ctx, btcKeeper, signer, nexusKeeper, snapshotter, voter, msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

}

func mapi(n int, f func(i int)) {
	for i := 0; i < n; i++ {
		f(i)
	}
}

func randomMsgLink() types.MsgLink {
	return types.MsgLink{
		Sender:         sdk.AccAddress(rand.StrBetween(5, 20)),
		RecipientAddr:  rand.StrBetween(5, 100),
		RecipientChain: rand.StrBetween(5, 100),
	}
}

func randomMsgConfirmOutpoint() types.MsgConfirmOutpoint {
	return types.NewMsgConfirmOutpoint(sdk.AccAddress(rand.StrBetween(5, 20)), randomOutpointInfo())
}

func randomMsgVoteConfirmOutpoint() types.MsgVoteConfirmOutpoint {
	return types.MsgVoteConfirmOutpoint{
		Sender: sdk.AccAddress(rand.StrBetween(5, 20)),
		Poll: vote.PollMeta{
			Module: types.ModuleName,
			ID:     rand.StrBetween(5, 20),
		},
		OutPoint:  *randomOutpointInfo().OutPoint,
		Confirmed: rand.Bools(0.5).Next(),
	}
}

func randomOutpointInfo() types.OutPointInfo {
	txHash, err := chainhash.NewHash(rand.Bytes(chainhash.HashSize))
	if err != nil {
		panic(err)
	}
	return types.OutPointInfo{
		OutPoint: wire.NewOutPoint(txHash, mathRand.Uint32()),
		Amount:   btcutil.Amount(rand.I64Between(1, 10000000000)),
		Address:  randomAddress().EncodeAddress(),
	}
}

func randomTransfer() nexus.CrossChainTransfer {
	return nexus.CrossChainTransfer{
		Recipient: nexus.CrossChainAddress{Chain: exported.Bitcoin, Address: randomAddress().EncodeAddress()},
		Asset:     sdk.NewInt64Coin(exported.Bitcoin.NativeAsset, rand.I64Between(1, 100000000)),
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
