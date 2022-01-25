package keeper_test

import (
	"bytes"
	"fmt"
	mathRand "math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogoprototypes "github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

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
	tssTestUtils "github.com/axelarnetwork/axelar-core/x/tss/exported/testutils"
	tsstypes "github.com/axelarnetwork/axelar-core/x/tss/types"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
	voteMock "github.com/axelarnetwork/axelar-core/x/vote/exported/mock"
)

func TestHandleMsgLink(t *testing.T) {
	var (
		server      types.MsgServiceServer
		btcKeeper   *mock.BTCKeeperMock
		signer      *mock.SignerMock
		nexusKeeper *mock.NexusMock
		ctx         sdk.Context
		msg         *types.LinkRequest

		externalKeys []tss.Key
	)
	setup := func() {
		externalKeyCount := tsstypes.DefaultParams().ExternalMultisigThreshold.Denominator
		externalKeys = make([]tss.Key, externalKeyCount)
		for i := 0; i < int(externalKeyCount); i++ {
			externalKeys[i] = createRandomKey(tss.ExternalKey)
		}

		btcKeeper = &mock.BTCKeeperMock{
			GetNetworkFunc:        func(ctx sdk.Context) types.Network { return types.Mainnet },
			SetAddressInfoFunc:    func(sdk.Context, types.AddressInfo) {},
			SetDepositAddressFunc: func(sdk.Context, nexus.CrossChainAddress, btcutil.Address) {},
			LoggerFunc:            func(sdk.Context) log.Logger { return log.TestingLogger() },
			GetMasterAddressExternalKeyLockDurationFunc: func(ctx sdk.Context) time.Duration {
				return types.DefaultParams().MasterAddressExternalKeyLockDuration
			},
		}
		signer = &mock.SignerMock{
			GetExternalMultisigThresholdFunc: func(ctx sdk.Context) utils.Threshold { return tsstypes.DefaultParams().ExternalMultisigThreshold },
			GetCurrentKeyFunc: func(_ sdk.Context, _ nexus.Chain, keyRole tss.KeyRole) (tss.Key, bool) {
				return createRandomKey(tss.SecondaryKey, time.Now()), true
			},
			GetKeyFunc: func(ctx sdk.Context, keyID tss.KeyID) (tss.Key, bool) {
				for _, externalKey := range externalKeys {
					if keyID == externalKey.ID {
						return externalKey, true
					}
				}

				return tss.Key{}, false
			},
			GetExternalKeyIDsFunc: func(ctx sdk.Context, chain nexus.Chain) ([]tss.KeyID, bool) {
				externalKeyIDs := make([]tss.KeyID, len(externalKeys))
				for i := 0; i < len(externalKeyIDs); i++ {
					externalKeyIDs[i] = externalKeys[i].ID
				}

				return externalKeyIDs, true
			},
		}
		nexusKeeper = &mock.NexusMock{
			IsChainActivatedFunc: func(ctx sdk.Context, chain nexus.Chain) bool {
				return chain == exported.Bitcoin
			},
			GetChainFunc: func(_ sdk.Context, chain string) (nexus.Chain, bool) {
				return nexus.Chain{
					Name:                  chain,
					SupportsForeignAssets: true,
					Module:                rand.Str(10),
				}, true
			},
			LinkAddressesFunc:     func(sdk.Context, nexus.CrossChainAddress, nexus.CrossChainAddress) error { return nil },
			IsAssetRegisteredFunc: func(sdk.Context, nexus.Chain, string) bool { return true },
		}
		ctx = rand.Context(nil)
		msg = randomMsgLink()
		server = bitcoinKeeper.NewMsgServerImpl(btcKeeper, signer, nexusKeeper, &mock.VoterMock{}, &mock.SnapshotterMock{})
	}
	repeatCount := 20

	t.Run("happy path", testutils.Func(func(t *testing.T) {
		setup()
		res, err := server.Link(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)
		assert.Len(t, btcKeeper.SetAddressInfoCalls(), 1)
		assert.Len(t, nexusKeeper.LinkAddressesCalls(), 1)
		assert.Equal(t, exported.Bitcoin, signer.GetCurrentKeyCalls()[0].Chain)
		assert.Equal(t, msg.RecipientChain, nexusKeeper.GetChainCalls()[0].Chain)
		assert.Equal(t, btcKeeper.SetAddressInfoCalls()[0].Address.Address, res.DepositAddr)
		assert.Equal(t, types.Deposit, btcKeeper.SetAddressInfoCalls()[0].Address.Role)
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
		nexusKeeper.IsAssetRegisteredFunc = func(sdk.Context, nexus.Chain, string) bool { return false }
		_, err := server.Link(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))
}

func TestHandleMsgConfirmOutpoint(t *testing.T) {
	var (
		btcKeeper *mock.BTCKeeperMock
		voter     *mock.VoterMock
		signer    *mock.SignerMock
		nexusMock *mock.NexusMock
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
			GetAddressInfoFunc: func(sdk.Context, string) (types.AddressInfo, bool) {
				return types.AddressInfo{
					Address:      address.EncodeAddress(),
					RedeemScript: rand.Bytes(200),
					Role:         types.Deposit,
					KeyID:        tssTestUtils.RandKeyID(),
				}, true
			},
			GetRevoteLockingPeriodFunc:        func(sdk.Context) int64 { return int64(mathRand.Uint32()) },
			GetRequiredConfirmationHeightFunc: func(sdk.Context) uint64 { return mathRand.Uint64() },
			SetPendingOutpointInfoFunc:        func(sdk.Context, vote.PollKey, types.OutPointInfo) {},
			GetVotingThresholdFunc:            func(ctx sdk.Context) utils.Threshold { return types.DefaultParams().VotingThreshold },
			GetMinVoterCountFunc:              func(ctx sdk.Context) int64 { return types.DefaultParams().MinVoterCount },
		}
		voter = &mock.VoterMock{
			InitializePollFunc: func(sdk.Context, vote.PollKey, []sdk.ValAddress, ...vote.PollProperty) error { return nil },
		}

		signer = &mock.SignerMock{
			GetCurrentKeyIDFunc: func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.KeyID, bool) {
				return tssTestUtils.RandKeyID(), true
			},
			GetSnapshotCounterForKeyIDFunc: func(sdk.Context, tss.KeyID) (int64, bool) {
				return rand.PosI64(), true
			},
		}

		nexusMock = &mock.NexusMock{
			IsChainActivatedFunc: func(ctx sdk.Context, chain nexus.Chain) bool {
				return chain == exported.Bitcoin
			},
			GetChainMaintainersFunc: func(ctx sdk.Context, chain nexus.Chain) []sdk.ValAddress {
				return []sdk.ValAddress{}
			},
		}

		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())
		msg = randomMsgConfirmOutpoint()
		msg.OutPointInfo.Address = address.EncodeAddress()
		server = bitcoinKeeper.NewMsgServerImpl(btcKeeper, signer, nexusMock, voter, &mock.SnapshotterMock{})
	}

	repeatCount := 20
	t.Run("happy path deposit", testutils.Func(func(t *testing.T) {
		setup()
		_, err := server.ConfirmOutpoint(sdk.WrapSDKContext(ctx), msg)
		events := ctx.EventManager().ABCIEvents()
		assert.NoError(t, err)
		assert.Len(t, testutils.Events(events).Filter(func(event abci.Event) bool { return event.Type == types.EventTypeOutpointConfirmation }), 1)
		assert.Equal(t, msg.OutPointInfo, btcKeeper.SetPendingOutpointInfoCalls()[0].Info)
		assert.Equal(t, voter.InitializePollCalls()[0].Key, btcKeeper.SetPendingOutpointInfoCalls()[0].Key)
	}).Repeat(repeatCount))
	t.Run("happy path consolidation", testutils.Func(func(t *testing.T) {
		setup()
		addr, _ := btcKeeper.GetAddressInfo(ctx, msg.OutPointInfo.Address)
		addr.Role = types.Consolidation
		btcKeeper.GetAddressInfoFunc = func(sdk.Context, string) (types.AddressInfo, bool) {
			return addr, true
		}

		_, err := server.ConfirmOutpoint(sdk.WrapSDKContext(ctx), msg)
		events := sdk.UnwrapSDKContext(sdk.WrapSDKContext(ctx)).EventManager().ABCIEvents()
		assert.NoError(t, err)
		assert.Len(t, testutils.Events(events).Filter(func(event abci.Event) bool { return event.Type == types.EventTypeOutpointConfirmation }), 1)
		assert.Equal(t, msg.OutPointInfo, btcKeeper.SetPendingOutpointInfoCalls()[0].Info)
		assert.Equal(t, voter.InitializePollCalls()[0].Key, btcKeeper.SetPendingOutpointInfoCalls()[0].Key)
	}).Repeat(repeatCount))
	t.Run("already confirmed", testutils.Func(func(t *testing.T) {
		setup()
		btcKeeper.GetOutPointInfoFunc = func(sdk.Context, wire.OutPoint) (types.OutPointInfo, types.OutPointState, bool) {
			return msg.OutPointInfo, types.OutPointState_Confirmed, true
		}
		_, err := server.ConfirmOutpoint(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

	t.Run("already spent", testutils.Func(func(t *testing.T) {
		setup()
		btcKeeper.GetOutPointInfoFunc = func(sdk.Context, wire.OutPoint) (types.OutPointInfo, types.OutPointState, bool) {
			return msg.OutPointInfo, types.OutPointState_Spent, true
		}
		_, err := server.ConfirmOutpoint(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

	t.Run("address unknown", testutils.Func(func(t *testing.T) {
		setup()
		btcKeeper.GetAddressInfoFunc = func(sdk.Context, string) (types.AddressInfo, bool) { return types.AddressInfo{}, false }
		_, err := server.ConfirmOutpoint(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

	t.Run("init poll failed", testutils.Func(func(t *testing.T) {
		setup()

		voter.InitializePollFunc = func(sdk.Context, vote.PollKey, []sdk.ValAddress, ...vote.PollProperty) error {
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
		rotationCount := rand.I64Between(100, 1000)
		address := randomAddress()
		info = randomOutpointInfo()
		msg = randomMsgVoteConfirmOutpoint()
		msg.OutPoint = info.OutPoint
		depositAddressInfo = types.AddressInfo{
			Address:      address.EncodeAddress(),
			RedeemScript: rand.Bytes(200),
			Role:         types.Deposit,
			KeyID:        tssTestUtils.RandKeyID(),
		}
		btcKeeper = &mock.BTCKeeperMock{
			GetTransactionFeeRateFunc: func(sdk.Context) sdk.Dec { return sdk.NewDecWithPrec(25, 5) },
			GetOutPointInfoFunc: func(sdk.Context, wire.OutPoint) (types.OutPointInfo, types.OutPointState, bool) {
				return types.OutPointInfo{}, 0, false
			},
			SetConfirmedOutpointInfoFunc:  func(sdk.Context, tss.KeyID, types.OutPointInfo) {},
			GetPendingOutPointInfoFunc:    func(sdk.Context, vote.PollKey) (types.OutPointInfo, bool) { return info, true },
			DeletePendingOutPointInfoFunc: func(sdk.Context, vote.PollKey) {},
			GetSignedTxFunc:               func(sdk.Context, chainhash.Hash) (types.SignedTx, bool) { return types.SignedTx{}, false },
			GetAddressInfoFunc: func(sdk.Context, string) (types.AddressInfo, bool) {
				return depositAddressInfo, true
			},
			GetUnconfirmedAmountFunc: func(sdk.Context, tss.KeyID) btcutil.Amount { return 0 },
			SetUnconfirmedAmountFunc: func(sdk.Context, tss.KeyID, btcutil.Amount) {},
			LoggerFunc:               func(sdk.Context) log.Logger { return log.TestingLogger() },
		}
		voter = &mock.VoterMock{
			GetPollFunc: func(sdk.Context, vote.PollKey) vote.Poll {
				return &voteMock.PollMock{
					VoteFunc:      func(sdk.ValAddress, codec.ProtoMarshaler) error { return nil },
					GetResultFunc: func() codec.ProtoMarshaler { return &gogoprototypes.BoolValue{Value: true} },
					IsFunc: func(state vote.PollState) bool {
						return state == vote.Completed
					},
					AllowOverrideFunc: func() {},
				}
			},
		}

		nexusKeeper = &mock.NexusMock{
			IsChainActivatedFunc: func(ctx sdk.Context, chain nexus.Chain) bool {
				return chain == exported.Bitcoin
			},
			EnqueueForTransferFunc: func(sdk.Context, nexus.CrossChainAddress, sdk.Coin, sdk.Dec) error { return nil },
			GetRecipientFunc: func(ctx sdk.Context, sender nexus.CrossChainAddress) (nexus.CrossChainAddress, bool) {
				return nexus.CrossChainAddress{Chain: nexus.Chain{}, Address: ""}, true
			},
		}
		currentSecondaryKey = createRandomKey(tss.SecondaryKey, time.Now())
		signerKeeper := &mock.SignerMock{
			GetNextKeyFunc:    func(sdk.Context, nexus.Chain, tss.KeyRole) (tss.Key, bool) { return tss.Key{}, false },
			GetCurrentKeyFunc: func(sdk.Context, nexus.Chain, tss.KeyRole) (tss.Key, bool) { return currentSecondaryKey, true },
			AssignNextKeyFunc: func(sdk.Context, nexus.Chain, tss.KeyRole, tss.KeyID) error { return nil },
			GetKeyFunc: func(ctx sdk.Context, keyID tss.KeyID) (tss.Key, bool) {
				if keyID == currentSecondaryKey.ID {
					return currentSecondaryKey, true
				}

				switch keyID {
				case currentSecondaryKey.ID:
					return currentSecondaryKey, true
				case depositAddressInfo.KeyID:
					return createRandomKey(tss.SecondaryKey), true
				}

				return tss.Key{}, false
			},
			GetRotationCountOfKeyIDFunc: func(ctx sdk.Context, keyID tss.KeyID) (int64, bool) {
				return rotationCount - tsstypes.DefaultParams().UnbondingLockingKeyRotationCount + 1, true
			},
			GetRotationCountFunc: func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) int64 {
				return rotationCount
			},
			GetKeyUnbondingLockingKeyRotationCountFunc: func(ctx sdk.Context) int64 { return tsstypes.DefaultParams().UnbondingLockingKeyRotationCount },
		}
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())
		snapshotter := &mock.SnapshotterMock{GetOperatorFunc: func(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress {
			return rand.ValAddr()
		}}
		server = bitcoinKeeper.NewMsgServerImpl(btcKeeper, signerKeeper, nexusKeeper, voter, snapshotter)
	}

	repeats := 20

	t.Run("happy path confirm deposit to deposit address", testutils.Func(func(t *testing.T) {
		setup()

		_, err := server.VoteConfirmOutpoint(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)
		assert.Len(t, btcKeeper.DeletePendingOutPointInfoCalls(), 1)
		assert.Equal(t, info, btcKeeper.SetConfirmedOutpointInfoCalls()[0].Info)
		assert.Equal(t, depositAddressInfo.KeyID, btcKeeper.SetConfirmedOutpointInfoCalls()[0].KeyID)
		assert.Equal(t, info.Address, nexusKeeper.EnqueueForTransferCalls()[0].Sender.Address)
		assert.Equal(t, int64(info.Amount), nexusKeeper.EnqueueForTransferCalls()[0].Amount.Amount.Int64())

		// GIVEN a valid vote WHEN voting THEN event is emitted that captures vote value
		assert.Len(t, testutils.Events(ctx.EventManager().ABCIEvents()).Filter(func(event abci.Event) bool {
			isValidType := event.GetType() == types.EventTypeOutpointConfirmation
			if !isValidType {
				return false
			}
			isVoteAction := len(testutils.Attributes(event.GetAttributes()).Filter(func(attribute abci.EventAttribute) bool {
				return string(attribute.GetKey()) == sdk.AttributeKeyAction &&
					string(attribute.GetValue()) == types.AttributeValueVoted
			})) == 1
			if !isVoteAction {
				return false
			}
			hasCorrectValue := len(testutils.Attributes(event.GetAttributes()).Filter(func(attribute abci.EventAttribute) bool {
				if string(attribute.GetKey()) != types.AttributeKeyValue {
					return false
				}
				return string(attribute.GetValue()) == strconv.FormatBool(msg.Confirmed)
			})) == 1
			return hasCorrectValue
		}), 1)
	}).Repeat(repeats))

	t.Run("happy path confirm deposit to consolidation address", testutils.Func(func(t *testing.T) {
		setup()
		addr, _ := btcKeeper.GetAddressInfo(ctx, info.Address)
		addr.Role = types.Consolidation
		btcKeeper.GetAddressInfoFunc = func(sdk.Context, string) (types.AddressInfo, bool) {
			return addr, true
		}

		_, err := server.VoteConfirmOutpoint(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)
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
		addr, _ := btcKeeper.GetAddressInfo(ctx, info.Address)
		addr.Role = types.Consolidation
		btcKeeper.GetAddressInfoFunc = func(sdk.Context, string) (types.AddressInfo, bool) {
			return addr, true
		}

		_, err := server.VoteConfirmOutpoint(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)
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
		assert.Len(t, btcKeeper.DeletePendingOutPointInfoCalls(), 1)
		assert.Equal(t, info, btcKeeper.SetConfirmedOutpointInfoCalls()[0].Info)
		assert.Equal(t, depositAddressInfo.KeyID, btcKeeper.SetConfirmedOutpointInfoCalls()[0].KeyID)
		assert.Len(t, nexusKeeper.EnqueueForTransferCalls(), 1)
	}).Repeat(repeats))

	t.Run("happy path reject", testutils.Func(func(t *testing.T) {
		setup()
		voter.GetPollFunc = func(sdk.Context, vote.PollKey) vote.Poll {
			return &voteMock.PollMock{
				VoteFunc:      func(sdk.ValAddress, codec.ProtoMarshaler) error { return nil },
				GetResultFunc: func() codec.ProtoMarshaler { return &gogoprototypes.BoolValue{Value: false} },
				IsFunc: func(state vote.PollState) bool {
					return state == vote.Completed
				},
				AllowOverrideFunc: func() {},
			}
		}

		_, err := server.VoteConfirmOutpoint(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)
		assert.Len(t, btcKeeper.DeletePendingOutPointInfoCalls(), 1)
		assert.Len(t, btcKeeper.SetConfirmedOutpointInfoCalls(), 0)
		assert.Len(t, nexusKeeper.EnqueueForTransferCalls(), 0)
	}).Repeat(repeats))

	t.Run("happy path no result yet", testutils.Func(func(t *testing.T) {
		setup()
		voter.GetPollFunc = func(sdk.Context, vote.PollKey) vote.Poll {
			return &voteMock.PollMock{
				VoteFunc:      func(sdk.ValAddress, codec.ProtoMarshaler) error { return nil },
				GetResultFunc: func() codec.ProtoMarshaler { return nil },
				IsFunc: func(state vote.PollState) bool {
					return state == vote.Pending
				},
			}
		}

		_, err := server.VoteConfirmOutpoint(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)
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
			return info, types.OutPointState_Confirmed, true
		}

		_, err := server.VoteConfirmOutpoint(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)
		assert.Len(t, btcKeeper.DeletePendingOutPointInfoCalls(), 0)
		assert.Len(t, btcKeeper.SetConfirmedOutpointInfoCalls(), 0)
		assert.Len(t, nexusKeeper.EnqueueForTransferCalls(), 0)
	}).Repeat(repeats))

	t.Run("happy path second poll (outpoint already confirmed)", testutils.Func(func(t *testing.T) {
		setup()
		btcKeeper.GetOutPointInfoFunc = func(sdk.Context, wire.OutPoint) (types.OutPointInfo, types.OutPointState, bool) {
			return info, types.OutPointState_Confirmed, true
		}

		_, err := server.VoteConfirmOutpoint(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)
		assert.Len(t, btcKeeper.DeletePendingOutPointInfoCalls(), 1)
		assert.Len(t, btcKeeper.SetConfirmedOutpointInfoCalls(), 0)
		assert.Len(t, nexusKeeper.EnqueueForTransferCalls(), 0)
	}).Repeat(repeats))

	t.Run("happy path already spent", testutils.Func(func(t *testing.T) {
		setup()
		btcKeeper.GetOutPointInfoFunc = func(sdk.Context, wire.OutPoint) (types.OutPointInfo, types.OutPointState, bool) {
			return info, types.OutPointState_Spent, true
		}

		_, err := server.VoteConfirmOutpoint(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)
		assert.Len(t, btcKeeper.DeletePendingOutPointInfoCalls(), 1)

		// voting events should not be emitted if vote cannot proceed
		assert.Len(t, testutils.Events(ctx.EventManager().ABCIEvents()).Filter(func(event abci.Event) bool {
			isValidType := event.GetType() == types.EventTypeOutpointConfirmation
			if !isValidType {
				return false
			}
			isVoteAction := len(testutils.Attributes(event.GetAttributes()).Filter(func(attribute abci.EventAttribute) bool {
				return string(attribute.GetKey()) == sdk.AttributeKeyAction &&
					string(attribute.GetValue()) == types.AttributeValueVoted
			})) == 1
			return isVoteAction
		}), 0)
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
		voter.GetPollFunc = func(sdk.Context, vote.PollKey) vote.Poll {
			return &voteMock.PollMock{
				VoteFunc: func(sdk.ValAddress, codec.ProtoMarshaler) error { return fmt.Errorf("some error") },
			}
		}

		_, err := server.VoteConfirmOutpoint(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("enqueue transfer failed", testutils.Func(func(t *testing.T) {
		setup()
		nexusKeeper.EnqueueForTransferFunc = func(sdk.Context, nexus.CrossChainAddress, sdk.Coin, sdk.Dec) error {
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

func TestCreateRescueTx(t *testing.T) {
	var (
		btcKeeper    *mock.BTCKeeperMock
		signerKeeper *mock.SignerMock
		server       types.MsgServiceServer

		ctx              sdk.Context
		secondaryKey     tss.Key
		nextSecondaryKey tss.Key
		oldSecondaryKey  tss.Key
		oldMasterKey     tss.Key
	)

	repeat := 100

	setup := func() {
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())
		secondaryKey = createRandomKey(tss.SecondaryKey)
		nextSecondaryKey = createRandomKey(tss.SecondaryKey)
		oldSecondaryKey = createRandomKey(tss.SecondaryKey)
		oldMasterKey = createRandomKey(tss.MasterKey)

		btcKeeper = &mock.BTCKeeperMock{
			LoggerFunc: func(ctx sdk.Context) log.Logger { return log.TestingLogger() },
			GetUnsignedTxFunc: func(ctx sdk.Context, txType types.TxType) (types.UnsignedTx, bool) {
				return types.UnsignedTx{}, false
			},
			GetConfirmedOutpointInfoQueueForKeyFunc: func(ctx sdk.Context, keyID tss.KeyID) utils.KVQueue {
				return &utilsmock.KVQueueMock{
					IsEmptyFunc: func() bool { return true },
					DequeueFunc: func(value codec.ProtoMarshaler, filter ...func(value codec.ProtoMarshaler) bool) bool {
						return false
					},
				}
			},
			GetMaxInputCountFunc: func(ctx sdk.Context) int64 {
				return types.DefaultParams().MaxInputCount
			},
			DeleteOutpointInfoFunc:   func(ctx sdk.Context, outPoint wire.OutPoint) {},
			SetSpentOutpointInfoFunc: func(ctx sdk.Context, info types.OutPointInfo) {},
			GetAnyoneCanSpendAddressFunc: func(ctx sdk.Context) types.AddressInfo {
				return types.NewAnyoneCanSpendAddress(types.DefaultParams().Network)
			},
			SetAddressInfoFunc: func(ctx sdk.Context, address types.AddressInfo) {},
			SetUnsignedTxFunc:  func(ctx sdk.Context, tx types.UnsignedTx) {},
			GetNetworkFunc: func(ctx sdk.Context) types.Network {
				return types.DefaultParams().Network
			},
			GetMinOutputAmountFunc: func(ctx sdk.Context) btcutil.Amount {
				satoshi, err := types.ToSatoshiCoin(types.DefaultParams().MinOutputAmount)
				if err != nil {
					panic(err)
				}

				return btcutil.Amount(satoshi.Amount.Int64())
			},
		}
		signerKeeper = &mock.SignerMock{
			GetOldActiveKeysFunc: func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) ([]tss.Key, error) {
				switch keyRole {
				case tss.MasterKey:
					return []tss.Key{oldMasterKey}, nil
				case tss.SecondaryKey:
					return []tss.Key{oldSecondaryKey}, nil
				}

				return []tss.Key{}, nil
			},

			GetNextKeyFunc: func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.Key, bool) {
				return tss.Key{}, false
			},
			GetCurrentKeyFunc: func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.Key, bool) {
				if chain == exported.Bitcoin && keyRole == tss.SecondaryKey {
					return secondaryKey, true
				}

				return tss.Key{}, false
			},
		}

		voter := &mock.VoterMock{}
		nexusKeeper := &mock.NexusMock{
			IsChainActivatedFunc: func(ctx sdk.Context, chain nexus.Chain) bool {
				return chain == exported.Bitcoin
			},
		}
		snapshotter := &mock.SnapshotterMock{}
		server = bitcoinKeeper.NewMsgServerImpl(btcKeeper, signerKeeper, nexusKeeper, voter, snapshotter)
	}

	t.Run("shoud return error when no UTXO require", testutils.Func(func(t *testing.T) {
		setup()

		req := types.NewCreateRescueTxRequest(rand.AccAddr())
		_, err := server.CreateRescueTx(sdk.WrapSDKContext(ctx), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no rescue needed")
	}).Repeat(repeat))

	t.Run("should rescue UTXOs of an old master key", testutils.Func(func(t *testing.T) {
		setup()

		var inputs []types.OutPointInfo
		inputsTotal := sdk.ZeroInt()
		for i := 0; i < int(types.DefaultParams().MaxInputCount); i++ {
			input := randomOutpointInfo()
			inputs = append(inputs, input)
			inputsTotal = inputsTotal.AddRaw(int64(input.Amount))
		}

		btcKeeper.GetConfirmedOutpointInfoQueueForKeyFunc = func(ctx sdk.Context, keyID tss.KeyID) utils.KVQueue {
			if keyID == oldMasterKey.ID {
				dequeueCount := 0

				return &utilsmock.KVQueueMock{
					IsEmptyFunc: func() bool { return true },
					DequeueFunc: func(value codec.ProtoMarshaler, filter ...func(value codec.ProtoMarshaler) bool) bool {
						if dequeueCount >= len(inputs) {
							return false
						}

						types.ModuleCdc.MustUnmarshalLengthPrefixed(
							types.ModuleCdc.MustMarshalLengthPrefixed(&inputs[dequeueCount]),
							value,
						)

						dequeueCount++
						return true
					},
				}
			}

			return &utilsmock.KVQueueMock{
				IsEmptyFunc: func() bool { return true },
				DequeueFunc: func(value codec.ProtoMarshaler, filter ...func(value codec.ProtoMarshaler) bool) bool {
					return false
				},
			}
		}
		btcKeeper.GetOutPointInfoFunc = func(ctx sdk.Context, outPoint wire.OutPoint) (types.OutPointInfo, types.OutPointState, bool) {
			for _, input := range inputs {
				if input.OutPoint == outPoint.String() {
					return input, types.OutPointState_Spent, true
				}
			}

			return types.OutPointInfo{}, types.OutPointState_None, false
		}
		btcKeeper.GetAddressInfoFunc = func(_ sdk.Context, encodedAddress string) (types.AddressInfo, bool) {
			return types.AddressInfo{
				Address:      encodedAddress,
				RedeemScript: nil,
				KeyID:        oldMasterKey.ID,
			}, true
		}

		req := types.NewCreateRescueTxRequest(rand.AccAddr())
		_, err := server.CreateRescueTx(sdk.WrapSDKContext(ctx), req)
		assert.NoError(t, err)

		network := types.DefaultParams().Network
		expectedAnyoneCanSpendAddress := types.NewAnyoneCanSpendAddress(network).Address

		expectedSecondaryConsolidationAddress, err := types.NewSecondaryConsolidationAddress(secondaryKey, network)
		assert.NoError(t, err)

		minOutputAmount, err := types.ToSatoshiCoin(types.DefaultParams().MinOutputAmount)
		if err != nil {
			panic(err)
		}

		assert.Len(t, btcKeeper.SetUnsignedTxCalls(), 1)
		assert.Len(t, btcKeeper.DeleteOutpointInfoCalls(), len(inputs))
		assert.Len(t, btcKeeper.SetSpentOutpointInfoCalls(), len(inputs))
		assert.Len(t, btcKeeper.SetAddressInfoCalls(), 1)
		assert.Equal(t, expectedSecondaryConsolidationAddress.Address, btcKeeper.SetAddressInfoCalls()[0].Address.Address)
		actualUnsignedTx := btcKeeper.SetUnsignedTxCalls()[0].Tx
		assert.Equal(t, types.Rescue, actualUnsignedTx.Type)
		assert.Len(t, actualUnsignedTx.GetTx().TxIn, len(inputs))
		for i, txIn := range actualUnsignedTx.GetTx().TxIn {
			assert.Equal(t, txIn.Sequence, wire.MaxTxInSequenceNum)
			assert.Equal(t, txIn.PreviousOutPoint.String(), inputs[i].OutPoint)
		}
		var expectedOutputs []types.Output
		expectedOutputs = append(expectedOutputs, types.Output{
			Recipient: types.MustDecodeAddress(expectedAnyoneCanSpendAddress, network),
			Amount:    btcutil.Amount(minOutputAmount.Amount.Int64()),
		})
		expectedOutputs = append(expectedOutputs, types.Output{
			Recipient: types.MustDecodeAddress(expectedSecondaryConsolidationAddress.Address, network),
		})
		assertTxOutputs(t, actualUnsignedTx.GetTx(), expectedOutputs...)
		assert.Equal(t, uint32(0), actualUnsignedTx.GetTx().LockTime)
		assert.Greater(t, btcutil.Amount(inputsTotal.Int64()), actualUnsignedTx.InternalTransferAmount)
		assert.Greater(t, actualUnsignedTx.InternalTransferAmount, btcutil.Amount(0))
	}).Repeat(repeat))

	t.Run("should rescue UTXOs of an old secondary key", testutils.Func(func(t *testing.T) {
		setup()

		var inputs []types.OutPointInfo
		inputsTotal := sdk.ZeroInt()
		for i := 0; i < int(types.DefaultParams().MaxInputCount); i++ {
			input := randomOutpointInfo()
			inputs = append(inputs, input)
			inputsTotal = inputsTotal.AddRaw(int64(input.Amount))
		}

		signerKeeper.GetNextKeyFunc = func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.Key, bool) {
			return nextSecondaryKey, true
		}
		btcKeeper.GetConfirmedOutpointInfoQueueForKeyFunc = func(ctx sdk.Context, keyID tss.KeyID) utils.KVQueue {
			if keyID == oldSecondaryKey.ID {
				dequeueCount := 0

				return &utilsmock.KVQueueMock{
					IsEmptyFunc: func() bool { return true },
					DequeueFunc: func(value codec.ProtoMarshaler, filter ...func(value codec.ProtoMarshaler) bool) bool {
						if dequeueCount >= len(inputs) {
							return false
						}

						types.ModuleCdc.MustUnmarshalLengthPrefixed(
							types.ModuleCdc.MustMarshalLengthPrefixed(&inputs[dequeueCount]),
							value,
						)

						dequeueCount++
						return true
					},
				}
			}

			return &utilsmock.KVQueueMock{
				IsEmptyFunc: func() bool { return true },
				DequeueFunc: func(value codec.ProtoMarshaler, filter ...func(value codec.ProtoMarshaler) bool) bool {
					return false
				},
			}
		}
		btcKeeper.GetOutPointInfoFunc = func(ctx sdk.Context, outPoint wire.OutPoint) (types.OutPointInfo, types.OutPointState, bool) {
			for _, input := range inputs {
				if input.OutPoint == outPoint.String() {
					return input, types.OutPointState_Spent, true
				}
			}

			return types.OutPointInfo{}, types.OutPointState_None, false
		}
		btcKeeper.GetAddressInfoFunc = func(_ sdk.Context, encodedAddress string) (types.AddressInfo, bool) {
			return types.AddressInfo{
				Address:      encodedAddress,
				RedeemScript: nil,
				KeyID:        oldSecondaryKey.ID,
			}, true
		}

		req := types.NewCreateRescueTxRequest(rand.AccAddr())
		_, err := server.CreateRescueTx(sdk.WrapSDKContext(ctx), req)
		assert.NoError(t, err)

		network := types.DefaultParams().Network
		expectedAnyoneCanSpendAddress := types.NewAnyoneCanSpendAddress(network).Address

		expectedSecondaryConsolidationAddress, err := types.NewSecondaryConsolidationAddress(nextSecondaryKey, network)
		assert.NoError(t, err)

		minOutputAmount, err := types.ToSatoshiCoin(types.DefaultParams().MinOutputAmount)
		if err != nil {
			panic(err)
		}

		assert.Len(t, btcKeeper.SetUnsignedTxCalls(), 1)
		assert.Len(t, btcKeeper.DeleteOutpointInfoCalls(), len(inputs))
		assert.Len(t, btcKeeper.SetSpentOutpointInfoCalls(), len(inputs))
		assert.Len(t, btcKeeper.SetAddressInfoCalls(), 1)
		assert.Equal(t, expectedSecondaryConsolidationAddress.Address, btcKeeper.SetAddressInfoCalls()[0].Address.Address)
		actualUnsignedTx := btcKeeper.SetUnsignedTxCalls()[0].Tx
		assert.Equal(t, types.Rescue, actualUnsignedTx.Type)
		assert.Len(t, actualUnsignedTx.GetTx().TxIn, len(inputs))
		for i, txIn := range actualUnsignedTx.GetTx().TxIn {
			assert.Equal(t, txIn.Sequence, wire.MaxTxInSequenceNum)
			assert.Equal(t, txIn.PreviousOutPoint.String(), inputs[i].OutPoint)
		}
		var expectedOutputs []types.Output
		expectedOutputs = append(expectedOutputs, types.Output{
			Recipient: types.MustDecodeAddress(expectedAnyoneCanSpendAddress, network),
			Amount:    btcutil.Amount(minOutputAmount.Amount.Int64()),
		})
		expectedOutputs = append(expectedOutputs, types.Output{
			Recipient: types.MustDecodeAddress(expectedSecondaryConsolidationAddress.Address, network),
		})
		assertTxOutputs(t, actualUnsignedTx.GetTx(), expectedOutputs...)
		assert.Equal(t, uint32(0), actualUnsignedTx.GetTx().LockTime)
		assert.Greater(t, btcutil.Amount(inputsTotal.Int64()), actualUnsignedTx.InternalTransferAmount)
		assert.Greater(t, actualUnsignedTx.InternalTransferAmount, btcutil.Amount(0))
	}).Repeat(repeat))
}

func TestCreateMasterTx(t *testing.T) {
	var (
		btcKeeper    *mock.BTCKeeperMock
		voter        *mock.VoterMock
		nexusKeeper  *mock.NexusMock
		signerKeeper *mock.SignerMock
		server       types.MsgServiceServer

		ctx                    sdk.Context
		masterKey              tss.Key
		oldMasterKey           tss.Key
		secondaryKey           tss.Key
		nextSecondaryKey       tss.Key
		consolidationKey       tss.Key
		externalKeys           []tss.Key
		masterKeyRotationCount int64
		inputs                 []types.OutPointInfo
		inputTotal             btcutil.Amount
	)
	setup := func() {
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())
		masterKey = createRandomKey(tss.MasterKey, time.Now())
		oldMasterKey = createRandomKey(tss.MasterKey)
		secondaryKey = createRandomKey(tss.SecondaryKey)
		nextSecondaryKey = createRandomKey(tss.SecondaryKey)
		consolidationKey = createRandomKey(tss.MasterKey)

		externalKeyCount := tsstypes.DefaultParams().ExternalMultisigThreshold.Denominator
		externalKeys = make([]tss.Key, externalKeyCount)
		for i := 0; i < int(externalKeyCount); i++ {
			externalKeys[i] = createRandomKey(tss.ExternalKey)
		}

		masterKeyRotationCount = rand.I64Between(100, 1000)
		oldMasterKeyRotationCount := masterKeyRotationCount - types.DefaultParams().MasterKeyRetentionPeriod

		inputCount := int(types.DefaultParams().MaxInputCount)
		inputs = make([]types.OutPointInfo, inputCount)
		inputTotal = 0
		for i := 0; i < inputCount; i++ {
			inputs[i] = randomOutpointInfo()
			inputTotal += inputs[i].Amount
		}

		btcKeeper = &mock.BTCKeeperMock{
			GetOutPointInfoFunc: func(ctx sdk.Context, outPoint wire.OutPoint) (types.OutPointInfo, types.OutPointState, bool) {
				for _, input := range inputs {
					if input.OutPoint == outPoint.String() {
						return input, types.OutPointState_Spent, true
					}
				}

				return types.OutPointInfo{}, types.OutPointState_None, false
			},
			GetConfirmedOutpointInfoQueueForKeyFunc: func(ctx sdk.Context, keyID tss.KeyID) utils.KVQueue {
				if keyID == masterKey.ID {
					dequeueCount := 0

					return &utilsmock.KVQueueMock{
						IsEmptyFunc: func() bool { return true },
						DequeueFunc: func(value codec.ProtoMarshaler, filter ...func(value codec.ProtoMarshaler) bool) bool {
							if dequeueCount >= len(inputs) {
								return false
							}

							types.ModuleCdc.MustUnmarshalLengthPrefixed(
								types.ModuleCdc.MustMarshalLengthPrefixed(&inputs[dequeueCount]),
								value,
							)

							dequeueCount++
							return true
						},
					}
				}

				return &utilsmock.KVQueueMock{}
			},
			GetMinOutputAmountFunc: func(ctx sdk.Context) btcutil.Amount {
				satoshi, err := types.ToSatoshiCoin(types.DefaultParams().MinOutputAmount)
				if err != nil {
					panic(err)
				}

				return btcutil.Amount(satoshi.Amount.Int64())
			},
			GetMaxSecondaryOutputAmountFunc: func(ctx sdk.Context) btcutil.Amount {
				satoshi, err := types.ToSatoshiCoin(types.DefaultParams().MaxSecondaryOutputAmount)
				if err != nil {
					panic(err)
				}

				return btcutil.Amount(satoshi.Amount.Int64())
			},
			GetMaxInputCountFunc: func(ctx sdk.Context) int64 {
				return types.DefaultParams().MaxInputCount
			},
			GetAnyoneCanSpendAddressFunc: func(ctx sdk.Context) types.AddressInfo {
				return types.NewAnyoneCanSpendAddress(types.DefaultParams().Network)
			},
			GetUnsignedTxFunc: func(ctx sdk.Context, txType types.TxType) (types.UnsignedTx, bool) {
				return types.UnsignedTx{}, false
			},
			GetMasterKeyRetentionPeriodFunc: func(ctx sdk.Context) int64 {
				return types.DefaultParams().MasterKeyRetentionPeriod
			},
			GetMasterAddressInternalKeyLockDurationFunc: func(ctx sdk.Context) time.Duration {
				return types.DefaultParams().MasterAddressInternalKeyLockDuration
			},
			GetMasterAddressExternalKeyLockDurationFunc: func(ctx sdk.Context) time.Duration {
				return types.DefaultParams().MasterAddressExternalKeyLockDuration
			},
			GetNetworkFunc: func(ctx sdk.Context) types.Network {
				return types.DefaultParams().Network
			},
			GetMaxTxSizeFunc: func(ctx sdk.Context) int64 { return types.DefaultParams().MaxTxSize },
			GetAddressInfoFunc: func(_ sdk.Context, encodedAddress string) (types.AddressInfo, bool) {
				return types.AddressInfo{
					Address:      encodedAddress,
					RedeemScript: nil,
					KeyID:        masterKey.ID,
				}, true
			},
			GetUnconfirmedAmountFunc: func(ctx sdk.Context, keyID tss.KeyID) btcutil.Amount { return 0 },
			DeleteOutpointInfoFunc:   func(ctx sdk.Context, outPoint wire.OutPoint) {},
			SetSpentOutpointInfoFunc: func(ctx sdk.Context, info types.OutPointInfo) {},
			SetAddressInfoFunc:       func(ctx sdk.Context, address types.AddressInfo) {},
			SetUnsignedTxFunc:        func(ctx sdk.Context, tx types.UnsignedTx) {},
		}
		voter = &mock.VoterMock{}
		nexusKeeper = &mock.NexusMock{
			IsChainActivatedFunc: func(ctx sdk.Context, chain nexus.Chain) bool {
				return chain == exported.Bitcoin
			},
		}
		signerKeeper = &mock.SignerMock{
			GetExternalKeyIDsFunc: func(ctx sdk.Context, chain nexus.Chain) ([]tss.KeyID, bool) {
				externalKeyIDs := make([]tss.KeyID, len(externalKeys))
				for i := 0; i < len(externalKeyIDs); i++ {
					externalKeyIDs[i] = externalKeys[i].ID
				}

				return externalKeyIDs, true
			},
			GetExternalMultisigThresholdFunc: func(ctx sdk.Context) utils.Threshold {
				return tsstypes.DefaultParams().ExternalMultisigThreshold
			},
			GetCurrentKeyFunc: func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.Key, bool) {
				switch keyRole {
				case tss.MasterKey:
					return masterKey, true
				case tss.SecondaryKey:
					return secondaryKey, true
				default:
					return tss.Key{}, false
				}
			},
			GetKeyFunc: func(ctx sdk.Context, keyID tss.KeyID) (tss.Key, bool) {
				switch keyID {
				case masterKey.ID:
					return masterKey, true
				case oldMasterKey.ID:
					return masterKey, true
				case secondaryKey.ID:
					return secondaryKey, true
				case consolidationKey.ID:
					return consolidationKey, true
				default:
					for _, externalKey := range externalKeys {
						if keyID == externalKey.ID {
							return externalKey, true
						}
					}

					return tss.Key{}, false
				}
			},
			GetRotationCountFunc: func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) int64 {
				if keyRole == tss.MasterKey {
					return masterKeyRotationCount
				}

				return 0
			},
			GetKeyByRotationCountFunc: func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole, rotationCount int64) (tss.Key, bool) {
				if keyRole == tss.MasterKey && rotationCount == oldMasterKeyRotationCount {
					return oldMasterKey, true
				}

				return tss.Key{}, false
			},
			GetNextKeyFunc: func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.Key, bool) {
				return tss.Key{}, false
			},
			AssertMatchesRequirementsFunc: func(ctx sdk.Context, snapshotter snapshot.Snapshotter, chain nexus.Chain, keyID tss.KeyID, keyRole tss.KeyRole) error {
				return nil
			},
			AssignNextKeyFunc: func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole, keyID tss.KeyID) error {
				return nil
			},
		}
		snapshotter := &mock.SnapshotterMock{}
		server = bitcoinKeeper.NewMsgServerImpl(btcKeeper, signerKeeper, nexusKeeper, voter, snapshotter)
	}

	t.Run("shoud create master consolidation transaction without key assignment when the consolidation key is the current master key", testutils.Func(func(t *testing.T) {
		setup()

		req := types.NewCreateMasterTxRequest(rand.AccAddr(), string(masterKey.ID), 0)
		_, err := server.CreateMasterTx(sdk.WrapSDKContext(ctx), req)
		assert.NoError(t, err)

		network := types.DefaultParams().Network
		expectedAnyoneCanSpendAddress := types.NewAnyoneCanSpendAddress(network).Address

		expectedMasterConsolidationAddress, err := types.NewMasterConsolidationAddress(masterKey, oldMasterKey, tsstypes.DefaultParams().ExternalMultisigThreshold.Numerator, externalKeys, masterKey.RotatedAt.Add(types.DefaultParams().MasterAddressInternalKeyLockDuration), masterKey.RotatedAt.Add(types.DefaultParams().MasterAddressExternalKeyLockDuration), network)
		assert.NoError(t, err)

		minOutputAmount, err := types.ToSatoshiCoin(types.DefaultParams().MinOutputAmount)
		if err != nil {
			panic(err)
		}

		assert.Len(t, btcKeeper.SetUnsignedTxCalls(), 1)
		assert.Len(t, btcKeeper.DeleteOutpointInfoCalls(), len(inputs))
		assert.Len(t, btcKeeper.SetSpentOutpointInfoCalls(), len(inputs))
		assert.Len(t, btcKeeper.SetAddressInfoCalls(), 1)
		assert.Equal(t, expectedMasterConsolidationAddress.Address, btcKeeper.SetAddressInfoCalls()[0].Address.Address)
		actualUnsignedTx := btcKeeper.SetUnsignedTxCalls()[0].Tx
		assert.Equal(t, types.MasterConsolidation, actualUnsignedTx.Type)
		assert.Len(t, actualUnsignedTx.GetTx().TxIn, len(inputs))
		for i, txIn := range actualUnsignedTx.GetTx().TxIn {
			assert.Equal(t, txIn.Sequence, wire.MaxTxInSequenceNum)
			assert.Equal(t, txIn.PreviousOutPoint.String(), inputs[i].OutPoint)
		}
		assertTxOutputs(t, actualUnsignedTx.GetTx(),
			types.Output{
				Recipient: types.MustDecodeAddress(expectedAnyoneCanSpendAddress, network),
				Amount:    btcutil.Amount(minOutputAmount.Amount.Int64()),
			},
			types.Output{
				Recipient: types.MustDecodeAddress(expectedMasterConsolidationAddress.Address, network),
			},
		)
		assert.Equal(t, uint32(0), actualUnsignedTx.GetTx().LockTime)
		assert.Equal(t, btcutil.Amount(0), actualUnsignedTx.InternalTransferAmount)

		assert.Len(t, signerKeeper.AssignNextKeyCalls(), 0)
	}))

	t.Run("should create master consolidation transaction sending no coin to the secondary key when the amount is not set", testutils.Func(func(t *testing.T) {
		setup()

		req := types.NewCreateMasterTxRequest(rand.AccAddr(), string(consolidationKey.ID), 0)
		_, err := server.CreateMasterTx(sdk.WrapSDKContext(ctx), req)
		assert.NoError(t, err)

		network := types.DefaultParams().Network
		expectedAnyoneCanSpendAddress := types.NewAnyoneCanSpendAddress(network).Address

		expectedMasterConsolidationAddress, err := types.NewMasterConsolidationAddress(consolidationKey, oldMasterKey, tsstypes.DefaultParams().ExternalMultisigThreshold.Numerator, externalKeys, masterKey.RotatedAt.Add(types.DefaultParams().MasterAddressInternalKeyLockDuration), masterKey.RotatedAt.Add(types.DefaultParams().MasterAddressExternalKeyLockDuration), network)
		assert.NoError(t, err)

		minOutputAmount, err := types.ToSatoshiCoin(types.DefaultParams().MinOutputAmount)
		if err != nil {
			panic(err)
		}

		assert.Len(t, btcKeeper.SetUnsignedTxCalls(), 1)
		assert.Len(t, btcKeeper.DeleteOutpointInfoCalls(), len(inputs))
		assert.Len(t, btcKeeper.SetSpentOutpointInfoCalls(), len(inputs))
		assert.Len(t, btcKeeper.SetAddressInfoCalls(), 1)
		assert.Equal(t, expectedMasterConsolidationAddress.Address, btcKeeper.SetAddressInfoCalls()[0].Address.Address)
		actualUnsignedTx := btcKeeper.SetUnsignedTxCalls()[0].Tx
		assert.Equal(t, types.MasterConsolidation, actualUnsignedTx.Type)
		assert.Len(t, actualUnsignedTx.GetTx().TxIn, len(inputs))
		for i, txIn := range actualUnsignedTx.GetTx().TxIn {
			assert.Equal(t, txIn.Sequence, wire.MaxTxInSequenceNum)
			assert.Equal(t, txIn.PreviousOutPoint.String(), inputs[i].OutPoint)
		}
		assertTxOutputs(t, actualUnsignedTx.GetTx(),
			types.Output{
				Recipient: types.MustDecodeAddress(expectedAnyoneCanSpendAddress, network),
				Amount:    btcutil.Amount(minOutputAmount.Amount.Int64()),
			},
			types.Output{
				Recipient: types.MustDecodeAddress(expectedMasterConsolidationAddress.Address, network),
			},
		)
		assert.Equal(t, uint32(0), actualUnsignedTx.GetTx().LockTime)
		assert.Equal(t, btcutil.Amount(0), actualUnsignedTx.InternalTransferAmount)

		assert.Len(t, signerKeeper.AssignNextKeyCalls(), 1)
		actualAssignNextKeyCall := signerKeeper.AssignNextKeyCalls()[0]
		assert.Equal(t, exported.Bitcoin, actualAssignNextKeyCall.Chain)
		assert.Equal(t, tss.MasterKey, actualAssignNextKeyCall.KeyRole)
		assert.Equal(t, consolidationKey.ID, actualAssignNextKeyCall.KeyID)
	}))

	t.Run("should create master consolidation transaction sending coins to the next secondary key when the amount is set and the next secondary key is already assigned", testutils.Func(func(t *testing.T) {
		setup()

		signerKeeper.GetNextKeyFunc = func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.Key, bool) {
			if keyRole == tss.SecondaryKey {
				return nextSecondaryKey, true
			}

			return tss.Key{}, false
		}

		secondaryKeyAmount := btcutil.Amount(rand.I64Between(1000, 10000))
		req := types.NewCreateMasterTxRequest(rand.AccAddr(), string(consolidationKey.ID), secondaryKeyAmount)
		_, err := server.CreateMasterTx(sdk.WrapSDKContext(ctx), req)
		assert.NoError(t, err)

		network := types.DefaultParams().Network
		expectedAnyoneCanSpendAddress := types.NewAnyoneCanSpendAddress(network).Address

		expectedSecondaryConsolidationAddress, err := types.NewSecondaryConsolidationAddress(nextSecondaryKey, network)
		assert.NoError(t, err)

		expectedMasterConsolidationAddress, err := types.NewMasterConsolidationAddress(consolidationKey, oldMasterKey, tsstypes.DefaultParams().ExternalMultisigThreshold.Numerator, externalKeys, masterKey.RotatedAt.Add(types.DefaultParams().MasterAddressInternalKeyLockDuration), masterKey.RotatedAt.Add(types.DefaultParams().MasterAddressExternalKeyLockDuration), network)
		assert.NoError(t, err)

		minOutputAmount, err := types.ToSatoshiCoin(types.DefaultParams().MinOutputAmount)
		if err != nil {
			panic(err)
		}

		assert.Len(t, btcKeeper.SetUnsignedTxCalls(), 1)
		assert.Len(t, btcKeeper.DeleteOutpointInfoCalls(), len(inputs))
		assert.Len(t, btcKeeper.SetSpentOutpointInfoCalls(), len(inputs))
		assert.Len(t, btcKeeper.SetAddressInfoCalls(), 2)
		assert.Equal(t, expectedSecondaryConsolidationAddress.Address, btcKeeper.SetAddressInfoCalls()[0].Address.Address)
		assert.Equal(t, expectedMasterConsolidationAddress.Address, btcKeeper.SetAddressInfoCalls()[1].Address.Address)
		actualUnsignedTx := btcKeeper.SetUnsignedTxCalls()[0].Tx
		assert.Equal(t, types.MasterConsolidation, actualUnsignedTx.Type)
		assert.Len(t, actualUnsignedTx.GetTx().TxIn, len(inputs))
		for i, txIn := range actualUnsignedTx.GetTx().TxIn {
			assert.Equal(t, txIn.Sequence, wire.MaxTxInSequenceNum)
			assert.Equal(t, txIn.PreviousOutPoint.String(), inputs[i].OutPoint)
		}
		assertTxOutputs(t, actualUnsignedTx.GetTx(),
			types.Output{
				Recipient: types.MustDecodeAddress(expectedSecondaryConsolidationAddress.Address, network),
				Amount:    secondaryKeyAmount,
			},
			types.Output{
				Recipient: types.MustDecodeAddress(expectedAnyoneCanSpendAddress, network),
				Amount:    btcutil.Amount(minOutputAmount.Amount.Int64()),
			},
			types.Output{
				Recipient: types.MustDecodeAddress(expectedMasterConsolidationAddress.Address, network),
			},
		)
		assert.Equal(t, uint32(0), actualUnsignedTx.GetTx().LockTime)
		assert.Equal(t, secondaryKeyAmount, actualUnsignedTx.InternalTransferAmount)

		assert.Len(t, signerKeeper.AssignNextKeyCalls(), 1)
		actualAssignNextKeyCall := signerKeeper.AssignNextKeyCalls()[0]
		assert.Equal(t, exported.Bitcoin, actualAssignNextKeyCall.Chain)
		assert.Equal(t, tss.MasterKey, actualAssignNextKeyCall.KeyRole)
		assert.Equal(t, consolidationKey.ID, actualAssignNextKeyCall.KeyID)
	}))

	t.Run("should create master consolidation transaction sending coins to the secondary key when the amount is set", func(t *testing.T) {
		setup()

		secondaryKeyAmount := btcutil.Amount(rand.I64Between(1000, 10000))
		req := types.NewCreateMasterTxRequest(rand.AccAddr(), string(consolidationKey.ID), secondaryKeyAmount)
		_, err := server.CreateMasterTx(sdk.WrapSDKContext(ctx), req)
		assert.NoError(t, err)

		network := types.DefaultParams().Network
		expectedAnyoneCanSpendAddress := types.NewAnyoneCanSpendAddress(network).Address
		expectedSecondaryConsolidationAddress, err := types.NewSecondaryConsolidationAddress(secondaryKey, network)
		assert.NoError(t, err)

		expectedMasterConsolidationAddress, err := types.NewMasterConsolidationAddress(consolidationKey, oldMasterKey, tsstypes.DefaultParams().ExternalMultisigThreshold.Numerator, externalKeys, masterKey.RotatedAt.Add(types.DefaultParams().MasterAddressInternalKeyLockDuration), masterKey.RotatedAt.Add(types.DefaultParams().MasterAddressExternalKeyLockDuration), network)
		assert.NoError(t, err)

		minOutputAmount, err := types.ToSatoshiCoin(types.DefaultParams().MinOutputAmount)
		if err != nil {
			panic(err)
		}

		assert.Len(t, btcKeeper.SetUnsignedTxCalls(), 1)
		assert.Len(t, btcKeeper.DeleteOutpointInfoCalls(), len(inputs))
		assert.Len(t, btcKeeper.SetSpentOutpointInfoCalls(), len(inputs))
		assert.Len(t, btcKeeper.SetAddressInfoCalls(), 2)
		assert.Equal(t, expectedSecondaryConsolidationAddress.Address, btcKeeper.SetAddressInfoCalls()[0].Address.Address)
		assert.Equal(t, expectedMasterConsolidationAddress.Address, btcKeeper.SetAddressInfoCalls()[1].Address.Address)
		actualUnsignedTx := btcKeeper.SetUnsignedTxCalls()[0].Tx
		assert.Equal(t, types.MasterConsolidation, actualUnsignedTx.Type)
		assert.Len(t, actualUnsignedTx.GetTx().TxIn, len(inputs))
		for i, txIn := range actualUnsignedTx.GetTx().TxIn {
			assert.Equal(t, txIn.Sequence, wire.MaxTxInSequenceNum)
			assert.Equal(t, txIn.PreviousOutPoint.String(), inputs[i].OutPoint)
		}
		assertTxOutputs(t, actualUnsignedTx.GetTx(),
			types.Output{
				Recipient: types.MustDecodeAddress(expectedSecondaryConsolidationAddress.Address, network),
				Amount:    secondaryKeyAmount,
			},
			types.Output{
				Recipient: types.MustDecodeAddress(expectedAnyoneCanSpendAddress, network),
				Amount:    btcutil.Amount(minOutputAmount.Amount.Int64()),
			},
			types.Output{
				Recipient: types.MustDecodeAddress(expectedMasterConsolidationAddress.Address, network),
			},
		)
		assert.Equal(t, uint32(0), actualUnsignedTx.GetTx().LockTime)
		assert.Equal(t, secondaryKeyAmount, actualUnsignedTx.InternalTransferAmount)

		assert.Len(t, signerKeeper.AssignNextKeyCalls(), 1)
		actualAssignNextKeyCall := signerKeeper.AssignNextKeyCalls()[0]
		assert.Equal(t, exported.Bitcoin, actualAssignNextKeyCall.Chain)
		assert.Equal(t, tss.MasterKey, actualAssignNextKeyCall.KeyRole)
		assert.Equal(t, consolidationKey.ID, actualAssignNextKeyCall.KeyID)
	})

	t.Run("should return error if consolidating to a new key while the current key still has UTXO", testutils.Func(func(t *testing.T) {
		setup()

		btcKeeper.GetConfirmedOutpointInfoQueueForKeyFunc = func(ctx sdk.Context, keyID tss.KeyID) utils.KVQueue {
			if keyID == masterKey.ID {
				dequeueCount := 0

				return &utilsmock.KVQueueMock{
					IsEmptyFunc: func() bool { return false },
					DequeueFunc: func(value codec.ProtoMarshaler, filter ...func(value codec.ProtoMarshaler) bool) bool {
						if dequeueCount >= len(inputs) {
							return false
						}

						types.ModuleCdc.MustUnmarshalLengthPrefixed(
							types.ModuleCdc.MustMarshalLengthPrefixed(&inputs[dequeueCount]),
							value,
						)

						dequeueCount++
						return true
					},
				}
			}

			return &utilsmock.KVQueueMock{}
		}

		req := types.NewCreateMasterTxRequest(rand.AccAddr(), string(consolidationKey.ID), 0)
		_, err := server.CreateMasterTx(sdk.WrapSDKContext(ctx), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "still has confirmed outpoints to spend, and spend is required before key rotation is allowed")
	}))

	t.Run("should return error if consolidating to a new key while the current key still has unconfirmed amount", testutils.Func(func(t *testing.T) {
		setup()

		btcKeeper.GetUnconfirmedAmountFunc = func(ctx sdk.Context, keyID tss.KeyID) btcutil.Amount {
			return btcutil.Amount(rand.I64Between(1, 100))
		}

		req := types.NewCreateMasterTxRequest(rand.AccAddr(), string(consolidationKey.ID), 0)
		_, err := server.CreateMasterTx(sdk.WrapSDKContext(ctx), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "still has unconfirmed outpoints to confirm, and confirm and spend is required before key rotation is allowed")
	}))

	t.Run("should return error if consolidating to a new key while the secondary key is sending coin to the current master key", func(t *testing.T) {
		setup()

		btcKeeper.GetUnsignedTxFunc = func(ctx sdk.Context, txType types.TxType) (types.UnsignedTx, bool) {
			if txType == types.SecondaryConsolidation {
				return types.UnsignedTx{InternalTransferAmount: btcutil.Amount(rand.I64Between(10, 100))}, true
			}

			return types.UnsignedTx{}, false
		}

		req := types.NewCreateMasterTxRequest(rand.AccAddr(), string(consolidationKey.ID), 0)
		_, err := server.CreateMasterTx(sdk.WrapSDKContext(ctx), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot assign the next master key while a secondary transaction is sending coin to the current master address")
	})
}

func TestCreatePendingTransfersTx(t *testing.T) {
	var (
		btcKeeper    *mock.BTCKeeperMock
		voter        *mock.VoterMock
		nexusKeeper  *mock.NexusMock
		signerKeeper *mock.SignerMock
		server       types.MsgServiceServer

		ctx                    sdk.Context
		masterKey              tss.Key
		nextMasterKey          tss.Key
		oldMasterKey           tss.Key
		secondaryKey           tss.Key
		consolidationKey       tss.Key
		externalKeys           []tss.Key
		masterKeyRotationCount int64
		inputs                 []types.OutPointInfo
		inputTotal             btcutil.Amount
		transfers              []nexus.CrossChainTransfer
	)
	setup := func() {
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())
		masterKey = createRandomKey(tss.MasterKey, time.Now())
		nextMasterKey = createRandomKey(tss.MasterKey, time.Now())
		oldMasterKey = createRandomKey(tss.MasterKey)
		secondaryKey = createRandomKey(tss.SecondaryKey)
		consolidationKey = createRandomKey(tss.SecondaryKey)

		externalKeyCount := tsstypes.DefaultParams().ExternalMultisigThreshold.Denominator
		externalKeys = make([]tss.Key, externalKeyCount)
		for i := 0; i < int(externalKeyCount); i++ {
			externalKeys[i] = createRandomKey(tss.ExternalKey)
		}

		masterKeyRotationCount = rand.I64Between(100, 1000)
		oldMasterKeyRotationCount := masterKeyRotationCount - types.DefaultParams().MasterKeyRetentionPeriod

		inputCount := int(types.DefaultParams().MaxInputCount)
		inputs = make([]types.OutPointInfo, inputCount)
		inputTotal = 0
		for i := 0; i < inputCount; i++ {
			inputs[i] = randomOutpointInfo()
			inputTotal += inputs[i].Amount
		}

		transfers = []nexus.CrossChainTransfer{}
		outputTotal := btcutil.Amount(0)
		for {
			transfer := randomCrossChainTransfer(int64(inputTotal))
			if transfer.Asset.Amount.AddRaw(int64(outputTotal)).Int64() > int64(inputTotal) {
				break
			}

			transfers = append(transfers, transfer)
			outputTotal += btcutil.Amount(transfer.Asset.Amount.Int64())
		}

		btcKeeper = &mock.BTCKeeperMock{
			LoggerFunc: func(ctx sdk.Context) log.Logger { return log.TestingLogger() },
			GetOutPointInfoFunc: func(ctx sdk.Context, outPoint wire.OutPoint) (types.OutPointInfo, types.OutPointState, bool) {
				for _, input := range inputs {
					if input.OutPoint == outPoint.String() {
						return input, types.OutPointState_Spent, true
					}
				}

				return types.OutPointInfo{}, types.OutPointState_None, false
			},
			GetConfirmedOutpointInfoQueueForKeyFunc: func(ctx sdk.Context, keyID tss.KeyID) utils.KVQueue {
				if keyID == secondaryKey.ID {
					dequeueCount := 0

					return &utilsmock.KVQueueMock{
						IsEmptyFunc: func() bool { return true },
						DequeueFunc: func(value codec.ProtoMarshaler, filter ...func(value codec.ProtoMarshaler) bool) bool {
							if dequeueCount >= len(inputs) {
								return false
							}

							types.ModuleCdc.MustUnmarshalLengthPrefixed(
								types.ModuleCdc.MustMarshalLengthPrefixed(&inputs[dequeueCount]),
								value,
							)

							dequeueCount++
							return true
						},
					}
				}

				return &utilsmock.KVQueueMock{}
			},
			GetMinOutputAmountFunc: func(ctx sdk.Context) btcutil.Amount {
				satoshi, err := types.ToSatoshiCoin(types.DefaultParams().MinOutputAmount)
				if err != nil {
					panic(err)
				}

				return btcutil.Amount(satoshi.Amount.Int64())
			},
			GetMaxInputCountFunc: func(ctx sdk.Context) int64 {
				return types.DefaultParams().MaxInputCount
			},
			GetAnyoneCanSpendAddressFunc: func(ctx sdk.Context) types.AddressInfo {
				return types.NewAnyoneCanSpendAddress(types.DefaultParams().Network)
			},
			GetUnsignedTxFunc: func(ctx sdk.Context, txType types.TxType) (types.UnsignedTx, bool) {
				return types.UnsignedTx{}, false
			},
			GetMasterKeyRetentionPeriodFunc: func(ctx sdk.Context) int64 {
				return types.DefaultParams().MasterKeyRetentionPeriod
			},
			GetMasterAddressInternalKeyLockDurationFunc: func(ctx sdk.Context) time.Duration {
				return types.DefaultParams().MasterAddressInternalKeyLockDuration
			},
			GetMasterAddressExternalKeyLockDurationFunc: func(ctx sdk.Context) time.Duration {
				return types.DefaultParams().MasterAddressExternalKeyLockDuration
			},
			GetNetworkFunc: func(ctx sdk.Context) types.Network {
				return types.DefaultParams().Network
			},
			GetMaxTxSizeFunc: func(ctx sdk.Context) int64 { return types.DefaultParams().MaxTxSize },
			GetAddressInfoFunc: func(_ sdk.Context, encodedAddress string) (types.AddressInfo, bool) {
				return types.AddressInfo{
					Address:      encodedAddress,
					RedeemScript: nil,
					KeyID:        masterKey.ID,
				}, true
			},
			GetUnconfirmedAmountFunc: func(ctx sdk.Context, keyID tss.KeyID) btcutil.Amount { return 0 },
			DeleteOutpointInfoFunc:   func(ctx sdk.Context, outPoint wire.OutPoint) {},
			SetSpentOutpointInfoFunc: func(ctx sdk.Context, info types.OutPointInfo) {},
			SetAddressInfoFunc:       func(ctx sdk.Context, address types.AddressInfo) {},
			SetUnsignedTxFunc:        func(ctx sdk.Context, tx types.UnsignedTx) {},
		}
		voter = &mock.VoterMock{}
		nexusKeeper = &mock.NexusMock{
			IsChainActivatedFunc: func(ctx sdk.Context, chain nexus.Chain) bool {
				return chain == exported.Bitcoin
			},
			GetTransfersForChainFunc: func(ctx sdk.Context, chain nexus.Chain, state nexus.TransferState) []nexus.CrossChainTransfer {
				if chain == exported.Bitcoin && state == nexus.Pending {
					return transfers
				}

				return []nexus.CrossChainTransfer{}
			},
			ArchivePendingTransferFunc: func(ctx sdk.Context, transfer nexus.CrossChainTransfer) {},
		}
		signerKeeper = &mock.SignerMock{
			GetExternalMultisigThresholdFunc: func(ctx sdk.Context) utils.Threshold {
				return tsstypes.DefaultParams().ExternalMultisigThreshold
			},
			GetExternalKeyIDsFunc: func(ctx sdk.Context, chain nexus.Chain) ([]tss.KeyID, bool) {
				externalKeyIDs := make([]tss.KeyID, len(externalKeys))
				for i := 0; i < len(externalKeyIDs); i++ {
					externalKeyIDs[i] = externalKeys[i].ID
				}

				return externalKeyIDs, true
			},
			GetCurrentKeyFunc: func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.Key, bool) {
				switch keyRole {
				case tss.MasterKey:
					return masterKey, true
				case tss.SecondaryKey:
					return secondaryKey, true
				default:
					return tss.Key{}, false
				}
			},
			GetKeyFunc: func(ctx sdk.Context, keyID tss.KeyID) (tss.Key, bool) {
				switch keyID {
				case masterKey.ID:
					return masterKey, true
				case oldMasterKey.ID:
					return masterKey, true
				case secondaryKey.ID:
					return secondaryKey, true
				case consolidationKey.ID:
					return consolidationKey, true
				default:
					for _, externalKey := range externalKeys {
						if keyID == externalKey.ID {
							return externalKey, true
						}
					}

					return tss.Key{}, false
				}
			},
			GetRotationCountFunc: func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) int64 {
				if keyRole == tss.MasterKey {
					return masterKeyRotationCount
				}

				return 0
			},
			GetKeyByRotationCountFunc: func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole, rotationCount int64) (tss.Key, bool) {
				if keyRole == tss.MasterKey && rotationCount == oldMasterKeyRotationCount {
					return oldMasterKey, true
				}

				return tss.Key{}, false
			},
			GetNextKeyFunc: func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.Key, bool) {
				return tss.Key{}, false
			},
			AssertMatchesRequirementsFunc: func(ctx sdk.Context, snapshotter snapshot.Snapshotter, chain nexus.Chain, keyID tss.KeyID, keyRole tss.KeyRole) error {
				return nil
			},
			AssignNextKeyFunc: func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole, keyID tss.KeyID) error {
				return nil
			},
			GetKeyUnbondingLockingKeyRotationCountFunc: func(ctx sdk.Context) int64 { return tsstypes.DefaultParams().UnbondingLockingKeyRotationCount },
		}
		snapshotter := &mock.SnapshotterMock{}
		server = bitcoinKeeper.NewMsgServerImpl(btcKeeper, signerKeeper, nexusKeeper, voter, snapshotter)
	}

	t.Run("shoud create secondary consolidation transaction without key assignment when the consolidation key is the current secondary key", testutils.Func(func(t *testing.T) {
		setup()

		req := types.NewCreatePendingTransfersTxRequest(rand.AccAddr(), string(secondaryKey.ID), 0)
		_, err := server.CreatePendingTransfersTx(sdk.WrapSDKContext(ctx), req)
		assert.NoError(t, err)

		network := types.DefaultParams().Network
		expectedAnyoneCanSpendAddress := types.NewAnyoneCanSpendAddress(network).Address

		expectedSecondaryConsolidationAddress, err := types.NewSecondaryConsolidationAddress(secondaryKey, network)
		assert.NoError(t, err)

		minOutputAmount, err := types.ToSatoshiCoin(types.DefaultParams().MinOutputAmount)
		if err != nil {
			panic(err)
		}

		assert.Len(t, btcKeeper.SetUnsignedTxCalls(), 1)
		assert.Len(t, btcKeeper.DeleteOutpointInfoCalls(), len(inputs))
		assert.Len(t, btcKeeper.SetSpentOutpointInfoCalls(), len(inputs))
		assert.Len(t, btcKeeper.SetAddressInfoCalls(), 1)
		assert.Equal(t, expectedSecondaryConsolidationAddress.Address, btcKeeper.SetAddressInfoCalls()[0].Address.Address)
		actualUnsignedTx := btcKeeper.SetUnsignedTxCalls()[0].Tx
		assert.Equal(t, types.SecondaryConsolidation, actualUnsignedTx.Type)
		assert.Len(t, actualUnsignedTx.GetTx().TxIn, len(inputs))
		for i, txIn := range actualUnsignedTx.GetTx().TxIn {
			assert.Equal(t, txIn.Sequence, wire.MaxTxInSequenceNum)
			assert.Equal(t, txIn.PreviousOutPoint.String(), inputs[i].OutPoint)
		}
		var expectedOutputs []types.Output
		for _, transfer := range transfers {
			expectedOutputs = append(expectedOutputs, types.Output{
				Recipient: types.MustDecodeAddress(transfer.Recipient.Address, network),
				Amount:    btcutil.Amount(transfer.Asset.Amount.Int64()),
			})
		}
		expectedOutputs = append(expectedOutputs, types.Output{
			Recipient: types.MustDecodeAddress(expectedAnyoneCanSpendAddress, network),
			Amount:    btcutil.Amount(minOutputAmount.Amount.Int64()),
		})
		expectedOutputs = append(expectedOutputs, types.Output{
			Recipient: types.MustDecodeAddress(expectedSecondaryConsolidationAddress.Address, network),
		})
		assertTxOutputs(t, actualUnsignedTx.GetTx(), expectedOutputs...)
		assert.Equal(t, uint32(0), actualUnsignedTx.GetTx().LockTime)
		assert.Equal(t, btcutil.Amount(0), actualUnsignedTx.InternalTransferAmount)

		assert.Len(t, signerKeeper.AssignNextKeyCalls(), 0)
	}))

	t.Run("should create secondary consolidation transaction sending no coin to the master key when the amount is not set", func(t *testing.T) {
		setup()

		req := types.NewCreatePendingTransfersTxRequest(rand.AccAddr(), string(consolidationKey.ID), 0)
		_, err := server.CreatePendingTransfersTx(sdk.WrapSDKContext(ctx), req)
		assert.NoError(t, err)

		network := types.DefaultParams().Network
		expectedAnyoneCanSpendAddress := types.NewAnyoneCanSpendAddress(network).Address

		expectedSecondaryConsolidationAddress, err := types.NewSecondaryConsolidationAddress(consolidationKey, network)
		assert.NoError(t, err)

		minOutputAmount, err := types.ToSatoshiCoin(types.DefaultParams().MinOutputAmount)
		if err != nil {
			panic(err)
		}

		assert.Len(t, btcKeeper.SetUnsignedTxCalls(), 1)
		assert.Len(t, btcKeeper.DeleteOutpointInfoCalls(), len(inputs))
		assert.Len(t, btcKeeper.SetSpentOutpointInfoCalls(), len(inputs))
		assert.Len(t, btcKeeper.SetAddressInfoCalls(), 1)
		assert.Equal(t, expectedSecondaryConsolidationAddress.Address, btcKeeper.SetAddressInfoCalls()[0].Address.Address)
		assert.Len(t, nexusKeeper.ArchivePendingTransferCalls(), len(transfers))
		actualUnsignedTx := btcKeeper.SetUnsignedTxCalls()[0].Tx
		assert.Equal(t, types.SecondaryConsolidation, actualUnsignedTx.Type)
		assert.Len(t, actualUnsignedTx.GetTx().TxIn, len(inputs))
		for i, txIn := range actualUnsignedTx.GetTx().TxIn {
			assert.Equal(t, txIn.Sequence, wire.MaxTxInSequenceNum)
			assert.Equal(t, txIn.PreviousOutPoint.String(), inputs[i].OutPoint)
		}
		var expectedOutputs []types.Output
		for _, transfer := range transfers {
			expectedOutputs = append(expectedOutputs, types.Output{
				Recipient: types.MustDecodeAddress(transfer.Recipient.Address, network),
				Amount:    btcutil.Amount(transfer.Asset.Amount.Int64()),
			})
		}
		expectedOutputs = append(expectedOutputs, types.Output{
			Recipient: types.MustDecodeAddress(expectedAnyoneCanSpendAddress, network),
			Amount:    btcutil.Amount(minOutputAmount.Amount.Int64()),
		})
		expectedOutputs = append(expectedOutputs, types.Output{
			Recipient: types.MustDecodeAddress(expectedSecondaryConsolidationAddress.Address, network),
		})
		assertTxOutputs(t, actualUnsignedTx.GetTx(), expectedOutputs...)
		assert.Equal(t, uint32(0), actualUnsignedTx.GetTx().LockTime)
		assert.Equal(t, btcutil.Amount(0), actualUnsignedTx.InternalTransferAmount)

		assert.Len(t, signerKeeper.AssignNextKeyCalls(), 1)
		actualAssignNextKeyCall := signerKeeper.AssignNextKeyCalls()[0]
		assert.Equal(t, exported.Bitcoin, actualAssignNextKeyCall.Chain)
		assert.Equal(t, tss.SecondaryKey, actualAssignNextKeyCall.KeyRole)
		assert.Equal(t, consolidationKey.ID, actualAssignNextKeyCall.KeyID)
	})

	t.Run("should create secondary consolidation transaction sending coin to the next master key when the amount is set and the next master key is already assigned", testutils.Func(func(t *testing.T) {
		setup()

		signerKeeper.GetNextKeyFunc = func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.Key, bool) {
			if keyRole == tss.MasterKey {
				return nextMasterKey, true
			}

			return tss.Key{}, false
		}

		masterKeyAmount := btcutil.Amount(transfers[len(transfers)-1].Asset.Amount.Int64())
		transfers = transfers[:len(transfers)-1]
		req := types.NewCreatePendingTransfersTxRequest(rand.AccAddr(), string(consolidationKey.ID), masterKeyAmount)
		_, err := server.CreatePendingTransfersTx(sdk.WrapSDKContext(ctx), req)
		assert.NoError(t, err)

		network := types.DefaultParams().Network
		expectedAnyoneCanSpendAddress := types.NewAnyoneCanSpendAddress(network).Address

		expectedSecondaryConsolidationAddress, err := types.NewSecondaryConsolidationAddress(consolidationKey, network)
		assert.NoError(t, err)

		expectedMasterConsolidationAddress, err := types.NewMasterConsolidationAddress(nextMasterKey, oldMasterKey, tsstypes.DefaultParams().ExternalMultisigThreshold.Numerator, externalKeys, masterKey.RotatedAt.Add(types.DefaultParams().MasterAddressInternalKeyLockDuration), masterKey.RotatedAt.Add(types.DefaultParams().MasterAddressExternalKeyLockDuration), network)
		assert.NoError(t, err)

		minOutputAmount, err := types.ToSatoshiCoin(types.DefaultParams().MinOutputAmount)
		if err != nil {
			panic(err)
		}

		assert.Len(t, btcKeeper.SetUnsignedTxCalls(), 1)
		assert.Len(t, btcKeeper.DeleteOutpointInfoCalls(), len(inputs))
		assert.Len(t, btcKeeper.SetSpentOutpointInfoCalls(), len(inputs))
		assert.Len(t, btcKeeper.SetAddressInfoCalls(), 2)
		assert.Equal(t, expectedMasterConsolidationAddress.Address, btcKeeper.SetAddressInfoCalls()[0].Address.Address)
		assert.Equal(t, expectedSecondaryConsolidationAddress.Address, btcKeeper.SetAddressInfoCalls()[1].Address.Address)
		actualUnsignedTx := btcKeeper.SetUnsignedTxCalls()[0].Tx
		assert.Equal(t, types.SecondaryConsolidation, actualUnsignedTx.Type)
		assert.Len(t, actualUnsignedTx.GetTx().TxIn, len(inputs))
		for i, txIn := range actualUnsignedTx.GetTx().TxIn {
			assert.Equal(t, txIn.Sequence, wire.MaxTxInSequenceNum)
			assert.Equal(t, txIn.PreviousOutPoint.String(), inputs[i].OutPoint)
		}
		var expectedOutputs []types.Output
		for _, transfer := range transfers {
			expectedOutputs = append(expectedOutputs, types.Output{
				Recipient: types.MustDecodeAddress(transfer.Recipient.Address, network),
				Amount:    btcutil.Amount(transfer.Asset.Amount.Int64()),
			})
		}
		expectedOutputs = append(expectedOutputs, types.Output{
			Recipient: types.MustDecodeAddress(expectedAnyoneCanSpendAddress, network),
			Amount:    btcutil.Amount(minOutputAmount.Amount.Int64()),
		})
		expectedOutputs = append(expectedOutputs, types.Output{
			Recipient: types.MustDecodeAddress(expectedMasterConsolidationAddress.Address, network),
			Amount:    masterKeyAmount,
		})
		expectedOutputs = append(expectedOutputs, types.Output{
			Recipient: types.MustDecodeAddress(expectedSecondaryConsolidationAddress.Address, network),
		})
		assertTxOutputs(t, actualUnsignedTx.GetTx(), expectedOutputs...)
		assert.Equal(t, uint32(0), actualUnsignedTx.GetTx().LockTime)
		assert.Equal(t, masterKeyAmount, actualUnsignedTx.InternalTransferAmount)

		assert.Len(t, signerKeeper.AssignNextKeyCalls(), 1)
		actualAssignNextKeyCall := signerKeeper.AssignNextKeyCalls()[0]
		assert.Equal(t, exported.Bitcoin, actualAssignNextKeyCall.Chain)
		assert.Equal(t, tss.SecondaryKey, actualAssignNextKeyCall.KeyRole)
		assert.Equal(t, consolidationKey.ID, actualAssignNextKeyCall.KeyID)
	}))

	t.Run("should create secondary consolidation transaction sending coin to the master key when the amount is set", testutils.Func(func(t *testing.T) {
		setup()

		masterKeyAmount := btcutil.Amount(transfers[len(transfers)-1].Asset.Amount.Int64())
		transfers = transfers[:len(transfers)-1]
		req := types.NewCreatePendingTransfersTxRequest(rand.AccAddr(), string(consolidationKey.ID), masterKeyAmount)
		_, err := server.CreatePendingTransfersTx(sdk.WrapSDKContext(ctx), req)
		assert.NoError(t, err)

		network := types.DefaultParams().Network
		expectedAnyoneCanSpendAddress := types.NewAnyoneCanSpendAddress(network).Address

		expectedSecondaryConsolidationAddress, err := types.NewSecondaryConsolidationAddress(consolidationKey, network)
		assert.NoError(t, err)

		expectedMasterConsolidationAddress, err := types.NewMasterConsolidationAddress(masterKey, oldMasterKey, tsstypes.DefaultParams().ExternalMultisigThreshold.Numerator, externalKeys, masterKey.RotatedAt.Add(types.DefaultParams().MasterAddressInternalKeyLockDuration), masterKey.RotatedAt.Add(types.DefaultParams().MasterAddressExternalKeyLockDuration), network)
		assert.NoError(t, err)

		minOutputAmount, err := types.ToSatoshiCoin(types.DefaultParams().MinOutputAmount)
		if err != nil {
			panic(err)
		}

		assert.Len(t, btcKeeper.SetUnsignedTxCalls(), 1)
		assert.Len(t, btcKeeper.DeleteOutpointInfoCalls(), len(inputs))
		assert.Len(t, btcKeeper.SetSpentOutpointInfoCalls(), len(inputs))
		assert.Len(t, btcKeeper.SetAddressInfoCalls(), 2)
		assert.Equal(t, expectedMasterConsolidationAddress.Address, btcKeeper.SetAddressInfoCalls()[0].Address.Address)
		assert.Equal(t, expectedSecondaryConsolidationAddress.Address, btcKeeper.SetAddressInfoCalls()[1].Address.Address)
		assert.Len(t, nexusKeeper.ArchivePendingTransferCalls(), len(transfers))
		actualUnsignedTx := btcKeeper.SetUnsignedTxCalls()[0].Tx
		assert.Equal(t, types.SecondaryConsolidation, actualUnsignedTx.Type)
		assert.Len(t, actualUnsignedTx.GetTx().TxIn, len(inputs))
		for i, txIn := range actualUnsignedTx.GetTx().TxIn {
			assert.Equal(t, txIn.Sequence, wire.MaxTxInSequenceNum)
			assert.Equal(t, txIn.PreviousOutPoint.String(), inputs[i].OutPoint)
		}
		var expectedOutputs []types.Output
		for _, transfer := range transfers {
			expectedOutputs = append(expectedOutputs, types.Output{
				Recipient: types.MustDecodeAddress(transfer.Recipient.Address, network),
				Amount:    btcutil.Amount(transfer.Asset.Amount.Int64()),
			})
		}
		expectedOutputs = append(expectedOutputs, types.Output{
			Recipient: types.MustDecodeAddress(expectedAnyoneCanSpendAddress, network),
			Amount:    btcutil.Amount(minOutputAmount.Amount.Int64()),
		})
		expectedOutputs = append(expectedOutputs, types.Output{
			Recipient: types.MustDecodeAddress(expectedMasterConsolidationAddress.Address, network),
			Amount:    masterKeyAmount,
		})
		expectedOutputs = append(expectedOutputs, types.Output{
			Recipient: types.MustDecodeAddress(expectedSecondaryConsolidationAddress.Address, network),
		})
		assertTxOutputs(t, actualUnsignedTx.GetTx(), expectedOutputs...)
		assert.Equal(t, uint32(0), actualUnsignedTx.GetTx().LockTime)
		assert.Equal(t, masterKeyAmount, actualUnsignedTx.InternalTransferAmount)

		assert.Len(t, signerKeeper.AssignNextKeyCalls(), 1)
		actualAssignNextKeyCall := signerKeeper.AssignNextKeyCalls()[0]
		assert.Equal(t, exported.Bitcoin, actualAssignNextKeyCall.Chain)
		assert.Equal(t, tss.SecondaryKey, actualAssignNextKeyCall.KeyRole)
		assert.Equal(t, consolidationKey.ID, actualAssignNextKeyCall.KeyID)
	}))

	t.Run("should return error if consolidating to a new key while the current key still has UTXO", func(t *testing.T) {
		setup()

		btcKeeper.GetConfirmedOutpointInfoQueueForKeyFunc = func(ctx sdk.Context, keyID tss.KeyID) utils.KVQueue {
			if keyID == secondaryKey.ID {
				dequeueCount := 0

				return &utilsmock.KVQueueMock{
					IsEmptyFunc: func() bool { return false },
					DequeueFunc: func(value codec.ProtoMarshaler, filter ...func(value codec.ProtoMarshaler) bool) bool {
						if dequeueCount >= len(inputs) {
							return false
						}

						types.ModuleCdc.MustUnmarshalLengthPrefixed(
							types.ModuleCdc.MustMarshalLengthPrefixed(&inputs[dequeueCount]),
							value,
						)

						dequeueCount++
						return true
					},
				}
			}

			return &utilsmock.KVQueueMock{}
		}

		req := types.NewCreatePendingTransfersTxRequest(rand.AccAddr(), string(consolidationKey.ID), 0)
		_, err := server.CreatePendingTransfersTx(sdk.WrapSDKContext(ctx), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "still has confirmed outpoints to spend, and spend is required before key rotation is allowed")
	})

	t.Run("should return error if consolidating to a new key while the current key still has unconfirmed amount", func(t *testing.T) {
		setup()

		btcKeeper.GetUnconfirmedAmountFunc = func(ctx sdk.Context, keyID tss.KeyID) btcutil.Amount {
			return btcutil.Amount(rand.I64Between(1, 100))
		}

		req := types.NewCreatePendingTransfersTxRequest(rand.AccAddr(), string(consolidationKey.ID), 0)
		_, err := server.CreatePendingTransfersTx(sdk.WrapSDKContext(ctx), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "still has unconfirmed outpoints to confirm, and confirm and spend is required before key rotation is allowed")
	})

	t.Run("should return error if consolidating to a new key while the master key is sending coin to the current secondary key", testutils.Func(func(t *testing.T) {
		setup()

		btcKeeper.GetUnsignedTxFunc = func(ctx sdk.Context, txType types.TxType) (types.UnsignedTx, bool) {
			if txType == types.MasterConsolidation {
				return types.UnsignedTx{InternalTransferAmount: btcutil.Amount(rand.I64Between(10, 100))}, true
			}

			return types.UnsignedTx{}, false
		}

		req := types.NewCreatePendingTransfersTxRequest(rand.AccAddr(), string(consolidationKey.ID), 0)
		_, err := server.CreatePendingTransfersTx(sdk.WrapSDKContext(ctx), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot assign the next secondary key while a master transaction is sending coin to the current secondary address")
	}))
}

func assertTxOutputs(t *testing.T, tx *wire.MsgTx, outputs ...types.Output) {
	assert.Len(t, tx.TxOut, len(outputs))

	for _, expected := range outputs {
		found := false
		pkScript, err := txscript.PayToAddrScript(expected.Recipient)
		if err != nil {
			panic(err)
		}

		for _, actual := range tx.TxOut {
			if !bytes.Equal(pkScript, actual.PkScript) {
				continue
			}

			// ignore if amount is 0
			if expected.Amount != 0 && expected.Amount != btcutil.Amount(actual.Value) {
				continue
			}

			found = true
			break
		}

		assert.True(t, found, fmt.Sprintf("expected output %s-%d not found", expected.Recipient, expected.Amount))
	}
}

func createRandomKey(keyRole tss.KeyRole, rotatedAt ...time.Time) tss.Key {
	privKey, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		panic(err)
	}

	key := tss.Key{
		ID: tssTestUtils.RandKeyID(),
		PublicKey: &tss.Key_ECDSAKey_{
			ECDSAKey: &tss.Key_ECDSAKey{
				Value: privKey.PubKey().SerializeCompressed(),
			},
		},
		Role:      keyRole,
		RotatedAt: nil,
	}

	if len(rotatedAt) > 0 {
		key.RotatedAt = &rotatedAt[0]
	}

	return key
}

func randomMsgLink() *types.LinkRequest {
	return types.NewLinkRequest(
		rand.AccAddr(),
		rand.StrBetween(5, 100),
		rand.StrBetween(5, 100))
}

func randomMsgConfirmOutpoint() *types.ConfirmOutpointRequest {
	return types.NewConfirmOutpointRequest(rand.AccAddr(), randomOutpointInfo())
}

func randomMsgVoteConfirmOutpoint() *types.VoteConfirmOutpointRequest {
	return types.NewVoteConfirmOutpointRequest(
		rand.AccAddr(),
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
	minOutputAmount, err := types.ToSatoshiCoin(types.DefaultParams().MinOutputAmount)
	if err != nil {
		panic(err)
	}

	return types.OutPointInfo{
		OutPoint: wire.NewOutPoint(txHash, vout).String(),
		Amount:   btcutil.Amount(rand.I64Between(minOutputAmount.Amount.Int64(), 10000000000)),
		Address:  randomAddress().EncodeAddress(),
	}
}

func randomCrossChainTransfer(maxAmount int64) nexus.CrossChainTransfer {
	minOutputAmount, err := types.ToSatoshiCoin(types.DefaultParams().MinOutputAmount)
	if err != nil {
		panic(err)
	}

	asset := types.DefaultParams().MinOutputAmount
	asset.Amount = asset.Amount.Add(sdk.NewDec(rand.I64Between(minOutputAmount.Amount.Int64(), maxAmount)))

	secondaryConsolidationAddress, _ := types.NewSecondaryConsolidationAddress(createRandomKey(tss.SecondaryKey), types.DefaultParams().Network)

	return nexus.NewPendingCrossChainTransfer(
		mathRand.Uint64(),
		nexus.CrossChainAddress{Chain: exported.Bitcoin, Address: secondaryConsolidationAddress.Address},
		sdk.NewCoin(asset.Denom, asset.Amount.TruncateInt()),
	)
}

func randomAddress() *btcutil.AddressWitnessScriptHash {
	addr, err := btcutil.NewAddressWitnessScriptHash(rand.Bytes(32), types.DefaultParams().Network.Params())
	if err != nil {
		panic(err)
	}
	return addr
}
