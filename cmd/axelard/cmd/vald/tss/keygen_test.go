package tss

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	"google.golang.org/grpc"

	"github.com/axelarnetwork/axelar-core/app"
	mock2 "github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/broadcaster/types/mock"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/tss/rpc/mock"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	tss "github.com/axelarnetwork/axelar-core/x/tss/types"
	mock3 "github.com/axelarnetwork/axelar-core/x/tss/types/mock"
)

func TestMgr_ProcessKeygenStart(t *testing.T) {
	var (
		mgr          *Mgr
		attributes   []sdk.Attribute
		keygenClient *mock3.TofndKeyGenClientMock
	)
	setup := func() {
		cdc := app.MakeEncodingConfig().Amino
		principalAddr := rand.StrBetween(5, 20)
		keygenClient = &mock3.TofndKeyGenClientMock{
			SendFunc:      func(*tofnd.MessageIn) error { return nil },
			CloseSendFunc: func() error { return nil },
		}
		client := &mock.ClientMock{
			KeygenFunc: func(context.Context, ...grpc.CallOption) (tofnd.GG20_KeygenClient, error) {
				return keygenClient, nil
			},
		}

		mgr = NewMgr(
			client,
			1*time.Second,
			principalAddr,
			&mock2.BroadcasterMock{},
			rand.Bytes(sdk.AddrLen),
			log.TestingLogger(),
			cdc,
		)

		attributes = []sdk.Attribute{
			{Key: tss.AttributeKeyKeyID, Value: rand.StrBetween(5, 20)},
			{Key: tss.AttributeKeyThreshold, Value: strconv.FormatInt(rand.I64Between(1, 100), 10)},
			{Key: tss.AttributeKeyParticipants, Value: string(cdc.MustMarshalJSON([]string{principalAddr}))},
			{Key: tss.AttributeKeyParticipantShareCounts, Value: string(cdc.MustMarshalJSON([]uint32{uint32(rand.I64Between(1, 20))}))},
		}

	}
	repeats := 20
	t.Run("server response io.EOF", testutils.Func(func(t *testing.T) {
		setup()
		keygenClient.RecvFunc = func() (*tofnd.MessageOut, error) { return nil, io.EOF }

		assert.Error(t, mgr.ProcessKeygenStart(rand.PosI64(), attributes))
	}).Repeat(repeats))

	t.Run("server response error", testutils.Func(func(t *testing.T) {
		setup()
		keygenClient.RecvFunc = func() (*tofnd.MessageOut, error) { return nil, fmt.Errorf("some error") }

		assert.Error(t, mgr.ProcessKeygenStart(rand.PosI64(), attributes))
	}).Repeat(repeats))
}
