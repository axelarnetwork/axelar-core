package keeper

import (
	"crypto/sha256"
	"testing"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	tssTestUtils "github.com/axelarnetwork/axelar-core/x/tss/exported/testutils"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
	"github.com/axelarnetwork/axelar-core/x/tss/types/mock"
)

func TestMsgServer_RotateKey(t *testing.T) {
	var (
		server    types.MsgServiceServer
		ctx       sdk.Context
		tssKeeper *mock.TSSKeeperMock
	)
	setup := func() {
		tssKeeper = &mock.TSSKeeperMock{
			RotateKeyFunc:     func(sdk.Context, nexus.Chain, exported.KeyRole) error { return nil },
			LoggerFunc:        func(ctx sdk.Context) log.Logger { return ctx.Logger() },
			AssignNextKeyFunc: func(sdk.Context, nexus.Chain, exported.KeyRole, exported.KeyID) error { return nil },
			AssertMatchesRequirementsFunc: func(ctx sdk.Context, snapshotter snapshot.Snapshotter, chain nexus.Chain, keyID exported.KeyID, keyRole exported.KeyRole) error {
				return nil
			},
		}
		snapshotter := &mock.SnapshotterMock{}
		staker := &mock.StakingKeeperMock{}
		voter := &mock.VoterMock{}
		nexusKeeper := &mock.NexusMock{
			GetChainFunc: func(_ sdk.Context, chain string) (nexus.Chain, bool) {
				return nexus.Chain{
					Name:                  chain,
					NativeAsset:           rand.StrBetween(5, 10),
					SupportsForeignAssets: rand.Bools(0.5).Next(),
				}, true
			},
		}
		rewarder := &mock.RewarderMock{}
		server = NewMsgServerImpl(tssKeeper, snapshotter, staker, voter, nexusKeeper, rewarder)
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())
	}
	repeats := 20
	t.Run("first key rotation", testutils.Func(func(t *testing.T) {
		setup()
		tssKeeper.GetCurrentKeyIDFunc = func(sdk.Context, nexus.Chain, exported.KeyRole) (exported.KeyID, bool) { return "", false }
		tssKeeper.GetNextKeyIDFunc = func(sdk.Context, nexus.Chain, exported.KeyRole) (exported.KeyID, bool) { return "", true }

		_, err := server.RotateKey(sdk.WrapSDKContext(ctx), &types.RotateKeyRequest{
			Sender:  rand.AccAddr(),
			Chain:   rand.StrBetween(5, 20),
			KeyRole: exported.KeyRole(rand.I64Between(1, 3)),
			KeyID:   tssTestUtils.RandKeyID(),
		})

		assert.NoError(t, err)
		assert.Len(t, tssKeeper.AssignNextKeyCalls(), 1)
		assert.Len(t, tssKeeper.RotateKeyCalls(), 1)
	}).Repeat(repeats))

	t.Run("next key is assigned", testutils.Func(func(t *testing.T) {
		setup()
		keyID := tssTestUtils.RandKeyID()
		tssKeeper.GetCurrentKeyIDFunc = func(sdk.Context, nexus.Chain, exported.KeyRole) (exported.KeyID, bool) {
			return tssTestUtils.RandKeyID(), true
		}
		tssKeeper.GetNextKeyIDFunc = func(sdk.Context, nexus.Chain, exported.KeyRole) (exported.KeyID, bool) { return keyID, true }

		_, err := server.RotateKey(sdk.WrapSDKContext(ctx), &types.RotateKeyRequest{
			Sender:  rand.AccAddr(),
			Chain:   rand.StrBetween(5, 20),
			KeyRole: exported.KeyRole(rand.I64Between(1, 3)),
			KeyID:   keyID,
		})

		assert.Error(t, err)
		assert.Len(t, tssKeeper.AssignNextKeyCalls(), 0)
		assert.Len(t, tssKeeper.RotateKeyCalls(), 0)
	}).Repeat(repeats))

	t.Run("no next key is assigned", testutils.Func(func(t *testing.T) {
		setup()
		tssKeeper.GetCurrentKeyIDFunc = func(sdk.Context, nexus.Chain, exported.KeyRole) (exported.KeyID, bool) {
			return tssTestUtils.RandKeyID(), true
		}
		tssKeeper.GetNextKeyIDFunc = func(sdk.Context, nexus.Chain, exported.KeyRole) (exported.KeyID, bool) { return "", false }

		_, err := server.RotateKey(sdk.WrapSDKContext(ctx), &types.RotateKeyRequest{
			Sender:  rand.AccAddr(),
			Chain:   rand.StrBetween(5, 20),
			KeyRole: exported.KeyRole(rand.I64Between(1, 3)),
			KeyID:   tssTestUtils.RandKeyID(),
		})

		assert.Error(t, err)
	}).Repeat(repeats))
}

func TestMsgServer_SubmitMultisigPubKey(t *testing.T) {
	var (
		server    types.MsgServiceServer
		ctx       sdk.Context
		tssKeeper *mock.TSSKeeperMock
		randSnap  snapshot.Snapshot
	)
	setup := func() {
		randSnap = randSnapshot()
		tssKeeper = &mock.TSSKeeperMock{
			LoggerFunc:                     func(ctx sdk.Context) log.Logger { return ctx.Logger() },
			GetSnapshotCounterForKeyIDFunc: func(sdk.Context, exported.KeyID) (int64, bool) { return rand.PosI64(), true },
			SubmitPubKeysFunc:              func(sdk.Context, exported.KeyID, sdk.ValAddress, ...[]byte) bool { return true },
			GetMultisigPubKeyCountFunc:     func(sdk.Context, exported.KeyID) int64 { return randSnap.TotalShareCount.Int64() },
			IsMultisigKeygenCompletedFunc:  func(sdk.Context, exported.KeyID) bool { return false },
		}
		snapshotter := &mock.SnapshotterMock{
			GetOperatorFunc: func(sdk.Context, sdk.AccAddress) sdk.ValAddress { return randSnap.Validators[0].GetSDKValidator().GetOperator() },
			GetSnapshotFunc: func(sdk.Context, int64) (snapshot.Snapshot, bool) { return randSnap, true },
		}
		staker := &mock.StakingKeeperMock{}
		voter := &mock.VoterMock{}
		nexusKeeper := &mock.NexusMock{}
		rewarder := &mock.RewarderMock{}
		server = NewMsgServerImpl(tssKeeper, snapshotter, staker, voter, nexusKeeper, rewarder)
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())
	}
	repeats := 20
	t.Run("should return error when a validator submits for a completed multisig keygen", testutils.Func(func(t *testing.T) {
		setup()
		tssKeeper.IsMultisigKeygenCompletedFunc = func(sdk.Context, exported.KeyID) bool { return true }

		pubKeyInfos := newPubKeyInfo(randSnap.Validators[0])

		_, err := server.SubmitMultisigPubKeys(sdk.WrapSDKContext(ctx), &types.SubmitMultisigPubKeysRequest{
			Sender:      rand.AccAddr(),
			KeyID:       tssTestUtils.RandKeyID(),
			PubKeyInfos: pubKeyInfos,
		})

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("should return error when a validator submits incorrect number of pub keys", testutils.Func(func(t *testing.T) {
		setup()

		pubKeyInfos := newPubKeyInfo(randSnap.Validators[0])
		pubKeyInfos = append(pubKeyInfos, pubKeyInfos[0])
		_, err := server.SubmitMultisigPubKeys(sdk.WrapSDKContext(ctx), &types.SubmitMultisigPubKeysRequest{
			Sender:      rand.AccAddr(),
			KeyID:       tssTestUtils.RandKeyID(),
			PubKeyInfos: pubKeyInfos,
		})

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("should return error when a validator submits a duplicate pub key", testutils.Func(func(t *testing.T) {
		setup()
		tssKeeper.SubmitPubKeysFunc = func(sdk.Context, exported.KeyID, sdk.ValAddress, ...[]byte) bool { return false }

		pubKeyInfos := newPubKeyInfo(randSnap.Validators[0])
		_, err := server.SubmitMultisigPubKeys(sdk.WrapSDKContext(ctx), &types.SubmitMultisigPubKeysRequest{
			Sender:      rand.AccAddr(),
			KeyID:       tssTestUtils.RandKeyID(),
			PubKeyInfos: pubKeyInfos,
		})

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("should return error when a validator submits invalid pub key ", testutils.Func(func(t *testing.T) {
		setup()
		tssKeeper.SubmitPubKeysFunc = func(sdk.Context, exported.KeyID, sdk.ValAddress, ...[]byte) bool { return false }

		pubKeyInfos := newPubKeyInfo(randSnap.Validators[0])
		idx := rand.I64Between(0, int64(len(pubKeyInfos)))
		pubKeyInfos[int(idx)].PubKey = rand.Bytes(33)

		_, err := server.SubmitMultisigPubKeys(sdk.WrapSDKContext(ctx), &types.SubmitMultisigPubKeysRequest{
			Sender:      rand.AccAddr(),
			KeyID:       tssTestUtils.RandKeyID(),
			PubKeyInfos: pubKeyInfos,
		})

		assert.Error(t, err)
	}).Repeat(repeats))
}

func newPubKeyInfo(validator snapshot.Validator) []exported.PubKeyInfo {
	var pubKeyInfos []exported.PubKeyInfo
	for i := int64(0); i < validator.ShareCount; i++ {
		privKey, _ := btcec.NewPrivateKey(btcec.S256())
		pk := btcec.PublicKey(privKey.PublicKey)
		d := sha256.Sum256([]byte(snap.Validators[0].GetSDKValidator().GetOperator().String()))
		sig, _ := privKey.Sign(d[:])
		pubKeyInfos = append(pubKeyInfos, exported.PubKeyInfo{PubKey: pk.SerializeCompressed(), Signature: sig.Serialize()})

	}
	return pubKeyInfos
}
