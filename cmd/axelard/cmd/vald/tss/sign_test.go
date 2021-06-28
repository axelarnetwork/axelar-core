package tss

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	"google.golang.org/grpc"

	mock2 "github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/broadcaster/types/mock"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/tss/rpc/mock"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	tss "github.com/axelarnetwork/axelar-core/x/tss/types"
	mock3 "github.com/axelarnetwork/axelar-core/x/tss/types/mock"
)

func TestMgr_ProcessSignStart(t *testing.T) {
	var (
		mgr        *Mgr
		attributes []sdk.Attribute
		signClient *mock3.TofndSignClientMock
	)
	setup := func() {
		cdc := testutils.MakeEncodingConfig().Amino
		principalAddr := rand.StrBetween(5, 20)
		signClient = &mock3.TofndSignClientMock{
			SendFunc:      func(*tofnd.MessageIn) error { return nil },
			CloseSendFunc: func() error { return nil },
		}
		client := &mock.ClientMock{
			SignFunc: func(context.Context, ...grpc.CallOption) (tofnd.GG20_SignClient, error) {
				return signClient, nil
			},
		}

		mgr = NewMgr(
			client,
			1*time.Second,
			principalAddr,
			&mock2.BroadcasterMock{},
			rand.Bytes(sdk.AddrLen),
			rand.PosI64(),
			log.TestingLogger(),
			cdc,
		)

		attributes = []sdk.Attribute{
			{Key: tss.AttributeKeyKeyID, Value: rand.StrBetween(5, 20)},
			{Key: tss.AttributeKeySigID, Value: rand.StrBetween(5, 20)},
			{Key: tss.AttributeKeyParticipants, Value: string(cdc.MustMarshalJSON([]string{principalAddr}))},
			{Key: tss.AttributeKeyPayload, Value: string(rand.BytesBetween(100, 300))},
		}
	}
	repeats := 20
	t.Run("server response io.EOF", testutils.Func(func(t *testing.T) {
		setup()
		signClient.RecvFunc = func() (*tofnd.MessageOut, error) { return nil, io.EOF }

		assert.Error(t, mgr.ProcessSignStart(rand.PosI64(), attributes))
	}).Repeat(repeats))

	t.Run("server response error", testutils.Func(func(t *testing.T) {
		setup()
		signClient.RecvFunc = func() (*tofnd.MessageOut, error) { return nil, fmt.Errorf("some error") }

		assert.Error(t, mgr.ProcessSignStart(rand.PosI64(), attributes))
	}).Repeat(repeats))
}
