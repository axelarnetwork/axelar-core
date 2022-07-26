package tss

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	"google.golang.org/grpc"

	"github.com/axelarnetwork/axelar-core/app"
	mock2 "github.com/axelarnetwork/axelar-core/sdk-utils/broadcast/mock"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/vald/tss/rpc/mock"
	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	tss "github.com/axelarnetwork/axelar-core/x/tss/types"
	mock3 "github.com/axelarnetwork/axelar-core/x/tss/types/mock"
	tmEvents "github.com/axelarnetwork/tm-events/events"
)

func TestMgr_ProcessSignStart(t *testing.T) {
	var (
		mgr        *Mgr
		attributes map[string]string
		signClient *mock3.TofndSignClientMock
	)
	setup := func() {
		cdc := app.MakeEncodingConfig().Amino
		principalAddr := rand.StrBetween(5, 20)
		signClient = &mock3.TofndSignClientMock{
			SendFunc:      func(*tofnd.MessageIn) error { return nil },
			CloseSendFunc: func() error { return nil },
		}
		cli := &mock.ClientMock{
			SignFunc: func(context.Context, ...grpc.CallOption) (tofnd.GG20_SignClient, error) {
				return signClient, nil
			},
		}
		multiSigCli := &mock.MultiSigClientMock{
			KeygenFunc: func(ctx context.Context, in *tofnd.KeygenRequest, opts ...grpc.CallOption) (*tofnd.KeygenResponse, error) {
				return &tofnd.KeygenResponse{KeygenResponse: &tofnd.KeygenResponse_PubKey{PubKey: rand.Bytes(33)}}, nil
			},
		}

		mgr = NewMgr(
			cli,
			multiSigCli,
			client.Context{},
			1*time.Second,
			principalAddr,
			&mock2.BroadcasterMock{},
			log.TestingLogger(),
			cdc,
		)

		attributes = map[string]string{
			tss.AttributeKeyKeyID:        rand.StrBetween(5, 20),
			tss.AttributeKeySigID:        rand.StrBetween(5, 20),
			tss.AttributeKeyParticipants: string(cdc.MustMarshalJSON([]string{principalAddr})),
			tss.AttributeKeyPayload:      string(rand.BytesBetween(100, 300)),
		}
	}
	repeats := 20
	t.Run("server response io.EOF", testutils.Func(func(t *testing.T) {
		setup()
		signClient.RecvFunc = func() (*tofnd.MessageOut, error) { return nil, io.EOF }

		assert.Error(t, mgr.ProcessSignStart(tmEvents.Event{Height: rand.PosI64(), Attributes: attributes}))
	}).Repeat(repeats))

	t.Run("server response error", testutils.Func(func(t *testing.T) {
		setup()
		signClient.RecvFunc = func() (*tofnd.MessageOut, error) { return nil, fmt.Errorf("some error") }

		assert.Error(t, mgr.ProcessSignStart(tmEvents.Event{Height: rand.PosI64(), Attributes: attributes}))
	}).Repeat(repeats))
}
