package keeper

import (
	"context"
	"strconv"
	"testing"

	tssd "github.com/axelarnetwork/tssd/pb"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/axelarnetwork/axelar-core/testutils/mock"
	stExported "github.com/axelarnetwork/axelar-core/x/staking/exported"
)

func newStaker() mock.TestStaker {
	val1 := stExported.Validator{Address: sdk.ValAddress("validator1"), Power: 100}
	val2 := stExported.Validator{Address: sdk.ValAddress("validator2"), Power: 100}
	val3 := stExported.Validator{Address: sdk.ValAddress("validator3"), Power: 100}
	val4 := stExported.Validator{Address: sdk.ValAddress("validator4"), Power: 100}
	staker := mock.NewTestStaker(val1, val2, val3, val4)
	return staker
}

func prepareBroadcaster(t *testing.T, ctx sdk.Context, cdc *codec.Codec, validators []stExported.Validator, msgIn chan sdk.Msg) mock.Broadcaster {
	broadcaster := mock.NewBroadcaster(cdc, sdk.AccAddress("proxy0"), validators[0].Address, msgIn)

	for i, v := range validators {
		assert.NoError(t, broadcaster.RegisterProxy(ctx, v.Address, sdk.AccAddress("proxy"+strconv.Itoa(i))))
	}

	return broadcaster
}

type mockTssClient struct {
}

func (tc mockTssClient) Keygen(_ context.Context, _ ...grpc.CallOption) (tssd.GG18_KeygenClient, error) {
	return mockKeyGenClient{}, nil
}

func (tc mockTssClient) Sign(ctx context.Context, opts ...grpc.CallOption) (tssd.GG18_SignClient, error) {
	return mockSignClient{}, nil
}

func (tc mockTssClient) GetKey(ctx context.Context, in *tssd.Uid, opts ...grpc.CallOption) (*tssd.Bytes, error) {
	panic("implement me")
}

func (tc mockTssClient) GetSig(ctx context.Context, in *tssd.Uid, opts ...grpc.CallOption) (*tssd.Bytes, error) {
	panic("implement me")
}

type mockKeyGenClient struct {
	recv chan *tssd.MessageOut
}

func (kc mockKeyGenClient) Send(in *tssd.MessageIn) error {
	return nil
}

func (kc mockKeyGenClient) Recv() (*tssd.MessageOut, error) {
	return <-kc.recv, nil
}

func (kc mockKeyGenClient) Header() (metadata.MD, error) {
	panic("implement me")
}

func (kc mockKeyGenClient) Trailer() metadata.MD {
	panic("implement me")
}

func (kc mockKeyGenClient) CloseSend() error {
	panic("implement me")
}

func (kc mockKeyGenClient) Context() context.Context {
	panic("implement me")
}

func (kc mockKeyGenClient) SendMsg(msg interface{}) error {
	panic("implement me")
}

func (kc mockKeyGenClient) RecvMsg(msg interface{}) error {
	panic("implement me")
}

type mockSignClient struct {
	recv chan *tssd.MessageOut
}

func (sc mockSignClient) Send(in *tssd.MessageIn) error {
	return nil
}

func (sc mockSignClient) Recv() (*tssd.MessageOut, error) {
	return <-sc.recv, nil
}

func (sc mockSignClient) Header() (metadata.MD, error) {
	panic("implement me")
}

func (sc mockSignClient) Trailer() metadata.MD {
	panic("implement me")
}

func (sc mockSignClient) CloseSend() error {
	panic("implement me")
}

func (sc mockSignClient) Context() context.Context {
	panic("implement me")
}

func (sc mockSignClient) SendMsg(msg interface{}) error {
	panic("implement me")
}

func (sc mockSignClient) RecvMsg(msg interface{}) error {
	panic("implement me")
}
