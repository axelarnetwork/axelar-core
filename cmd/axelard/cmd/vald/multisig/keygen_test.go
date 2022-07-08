package multisig_test

import (
	"context"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"
	sdkclient "github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	"google.golang.org/grpc"

	broadcastmock "github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/broadcast/mock"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/multisig"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/tss/rpc/mock"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/multisig/exported/testutils"
	"github.com/axelarnetwork/axelar-core/x/multisig/types"
	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
	. "github.com/axelarnetwork/utils/test"
)

func TestMgr_ProcessKeygenStarted(t *testing.T) {
	var (
		mgr         *multisig.Mgr
		participant sdk.ValAddress
		client      *mock.MultiSigClientMock
		broadcaster *broadcastmock.BroadcasterMock

		e abci.Event
	)

	givenMgr := Given("the multisig manager", func() {
		client = &mock.MultiSigClientMock{}
		broadcaster = &broadcastmock.BroadcasterMock{}
		participant = rand.ValAddr()

		mgr = multisig.NewMgr(
			client,
			sdkclient.Context{FromAddress: rand.AccAddr()},
			participant,
			log.TestingLogger(),
			broadcaster,
			time.Second,
		)
	})

	givenMgr.
		When("is not part of the listed participants", func() {
			event := funcs.Must(
				sdk.TypedEventToEvent(types.NewKeygenStarted(testutils.KeyID(), slices.Expand(func(int) sdk.ValAddress { return rand.ValAddr() }, 10))),
			)
			e = abci.Event{Type: event.Type, Attributes: event.Attributes}
		}).
		Then("should ignore", func(t *testing.T) {
			err := mgr.ProcessKeygenStarted(e)

			assert.NoError(t, err)
		}).
		Run(t)

	givenMgr.
		When("is part of the listed participants", func() {
			event := funcs.Must(
				sdk.TypedEventToEvent(types.NewKeygenStarted(testutils.KeyID(), []sdk.ValAddress{rand.ValAddr(), participant, rand.ValAddr()})),
			)
			e = abci.Event{Type: event.Type, Attributes: event.Attributes}
		}).
		Then("should ignore", func(t *testing.T) {
			sk := funcs.Must(btcec.NewPrivateKey(btcec.S256()))
			client.KeygenFunc = func(_ context.Context, in *tofnd.KeygenRequest, _ ...grpc.CallOption) (*tofnd.KeygenResponse, error) {
				return &tofnd.KeygenResponse{KeygenResponse: &tofnd.KeygenResponse_PubKey{PubKey: sk.PubKey().SerializeCompressed()}}, nil
			}
			client.SignFunc = func(_ context.Context, in *tofnd.SignRequest, _ ...grpc.CallOption) (*tofnd.SignResponse, error) {
				return &tofnd.SignResponse{SignResponse: &tofnd.SignResponse_Signature{Signature: funcs.Must(sk.Sign(in.MsgToSign)).Serialize()}}, nil
			}
			broadcaster.BroadcastFunc = func(ctx context.Context, msgs ...sdk.Msg) (*sdk.TxResponse, error) { return &sdk.TxResponse{}, nil }

			err := mgr.ProcessKeygenStarted(e)
			assert.NoError(t, err)

			assert.Len(t, broadcaster.BroadcastCalls(), 1)
			assert.Len(t, broadcaster.BroadcastCalls()[0].Msgs, 1)
			assert.IsType(t, broadcaster.BroadcastCalls()[0].Msgs[0], &types.SubmitPubKeyRequest{})
			assert.NoError(t, broadcaster.BroadcastCalls()[0].Msgs[0].(*types.SubmitPubKeyRequest).ValidateBasic())
		}).
		Run(t)
}
