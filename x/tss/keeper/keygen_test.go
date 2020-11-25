package keeper

import (
	"testing"

	tssd "github.com/axelarnetwork/tssd/pb"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/mock"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

func TestKeeper_StartKeygen_IdAlreadyInUse_ReturnError(t *testing.T) {
	ctx := sdk.NewContext(mock.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	staker := newStaker()
	broadcaster := prepareBroadcaster(t, ctx, testutils.Codec(), staker.GetAllValidators(ctx), nil)

	subspace := params.NewSubspace(testutils.Codec(), sdk.NewKVStoreKey("storeKey"), sdk.NewKVStoreKey("tstorekey"), "tss")
	k := NewKeeper(mockTssClient{}, subspace, broadcaster, staker)

	err := k.StartKeygen(ctx, "keyID", 1, staker.GetAllValidators(ctx))

	assert.NoError(t, err)

	err = k.StartKeygen(ctx, "keyID", 4, staker.GetAllValidators(ctx))
	assert.Error(t, err)
}

// Even if no session exists the keeper must not return an error, because we need to keep validators and
// non-participating nodes consistent (for non-participating nodes there should be no session)
func TestKeeper_KeygenMsg_NoSessionWithGivenID_Return(t *testing.T) {
	ctx := sdk.NewContext(mock.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	staker := newStaker()
	broadcaster := prepareBroadcaster(t, ctx, testutils.Codec(), staker.GetAllValidators(ctx), nil)
	subspace := params.NewSubspace(testutils.Codec(), sdk.NewKVStoreKey("storeKey"), sdk.NewKVStoreKey("tstorekey"), "tss")
	k := NewKeeper(mockTssClient{}, subspace, broadcaster, staker)

	assert.NoError(t, k.KeygenMsg(ctx, types.MsgKeygenTraffic{
		Sender:    broadcaster.Proxy,
		SessionID: "keyID",
		Payload:   &tssd.TrafficOut{},
	}))
}
