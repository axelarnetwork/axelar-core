package multisig_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	ec "github.com/btcsuite/btcd/btcec/v2/ecdsa"
	sdkclient "github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	broadcastmock "github.com/axelarnetwork/axelar-core/sdk-utils/broadcast/mock"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/vald/multisig"
	"github.com/axelarnetwork/axelar-core/vald/multisig/mock"
	"github.com/axelarnetwork/axelar-core/x/multisig/exported"
	"github.com/axelarnetwork/axelar-core/x/multisig/types"
	typestestutils "github.com/axelarnetwork/axelar-core/x/multisig/types/testutils"
	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	"github.com/axelarnetwork/utils/funcs"
	. "github.com/axelarnetwork/utils/test"
)

func TestMgr_ProcessSigningStarted(t *testing.T) {
	var (
		mgr         *multisig.Mgr
		participant sdk.ValAddress
		client      *mock.ClientMock
		broadcaster *broadcastmock.BroadcasterMock
		privateKey  *btcec.PrivateKey

		event *types.SigningStarted
	)

	givenMgr := Given("the multisig manager", func() {
		client = &mock.ClientMock{}
		broadcaster = &broadcastmock.BroadcasterMock{}
		participant = rand.ValAddr()

		mgr = multisig.NewMgr(
			client,
			sdkclient.Context{FromAddress: rand.AccAddr()},
			participant,
			broadcaster,
			time.Second,
		)
	})

	givenMgr.
		When("is not part of the listed participants", func() {
			key := typestestutils.Key()
			event = types.NewSigningStarted(uint64(rand.PosI64()), key, rand.Bytes(exported.HashLength), rand.NormalizedStr(3))
		}).
		Then("should ignore", func(t *testing.T) {
			err := mgr.ProcessSigningStarted(event)

			assert.NoError(t, err)
		}).
		Run(t)

	givenMgr.
		When("is part of the listed participants", func() {
			key := typestestutils.Key()
			privateKey = funcs.Must(btcec.NewPrivateKey())
			key.PubKeys[participant.String()] = privateKey.PubKey().SerializeCompressed()

			event = types.NewSigningStarted(uint64(rand.PosI64()), key, rand.Bytes(exported.HashLength), rand.NormalizedStr(3))
		}).
		Then("should handle", func(t *testing.T) {
			client.SignFunc = func(_ context.Context, in *tofnd.SignRequest, _ ...grpc.CallOption) (*tofnd.SignResponse, error) {
				return &tofnd.SignResponse{SignResponse: &tofnd.SignResponse_Signature{Signature: ec.Sign(privateKey, in.MsgToSign).Serialize()}}, nil
			}
			broadcaster.BroadcastFunc = func(ctx context.Context, msgs ...sdk.Msg) (*sdk.TxResponse, error) {
				for _, msg := range msgs {
					if _, ok := msg.(*types.SubmitSignatureRequest); !ok {
						return nil, fmt.Errorf("unexpected type of msg %T received", msg)
					}

					if err := msg.ValidateBasic(); err != nil {
						return nil, err
					}
				}

				return &sdk.TxResponse{}, nil
			}

			err := mgr.ProcessSigningStarted(event)
			assert.NoError(t, err)

			assert.Len(t, broadcaster.BroadcastCalls(), 1)
			assert.Len(t, broadcaster.BroadcastCalls()[0].Msgs, 1)
		}).
		Run(t)
}
