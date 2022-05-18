package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/keeper"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
	"github.com/axelarnetwork/axelar-core/x/tss/types/mock"
)

func TestNextKeyID(t *testing.T) {
	var (
		tssKeeper       *mock.TSSKeeperMock
		nexusKeeper     *mock.NexusMock
		stakingKeeper   *mock.StakingKeeperMock
		ctx             sdk.Context
		grpcQuerier     *keeper.Querier
		existingChain   nexus.ChainName
		existingKeyID   tss.KeyID
		existingKeyRole tss.KeyRole
	)

	setup := func() {
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())

		existingChain = "existing"
		existingKeyID = tss.KeyID("keyID")
		existingKeyRole = tss.MasterKey

		tssKeeper = &mock.TSSKeeperMock{
			GetNextKeyIDFunc: func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.KeyID, bool) {
				if chain.Name == existingChain && keyRole == existingKeyRole {
					return existingKeyID, true
				}
				return "", false
			},
		}

		nexusKeeper = &mock.NexusMock{
			GetChainFunc: func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
				if chain == existingChain {
					return nexus.Chain{
						Name:                  existingChain,
						SupportsForeignAssets: false,
						KeyType:               tss.Multisig,
						Module:                "evm",
					}, true
				}
				return nexus.Chain{}, false
			},
		}

		stakingKeeper = &mock.StakingKeeperMock{
			ValidatorFunc: func(ctx sdk.Context, addr sdk.ValAddress) stakingtypes.ValidatorI {
				return stakingtypes.Validator{}
			},
		}

		q := keeper.NewGRPCQuerier(tssKeeper, nexusKeeper, stakingKeeper)
		grpcQuerier = &q
	}

	setup()

	repeatCount := 1

	t.Run("when chain and key role exist get the keyID", testutils.Func(func(t *testing.T) {
		expectedRes := types.NextKeyIDResponse{
			KeyID: existingKeyID,
		}

		res, err := grpcQuerier.NextKeyID(sdk.WrapSDKContext(ctx), &types.NextKeyIDRequest{
			Chain:   existingChain.String(),
			KeyRole: existingKeyRole,
		})

		assert := assert.New(t)
		assert.NoError(err)

		assert.Equal(expectedRes, *res)
	}).Repeat(repeatCount))

	t.Run("if chain does not exist, get NotFound grpc code", testutils.Func(func(t *testing.T) {
		chain := "non-existing-chain"
		res, err := grpcQuerier.NextKeyID(sdk.WrapSDKContext(ctx), &types.NextKeyIDRequest{
			Chain:   chain,
			KeyRole: existingKeyRole,
		})

		assert := assert.New(t)
		assert.Nil(res)
		s, ok := status.FromError(err)
		assert.Equal(codes.NotFound, s.Code())
		assert.Equal(true, ok)
	}).Repeat(repeatCount))

	t.Run("if key role does not exist, get OK grpc code", testutils.Func(func(t *testing.T) {
		res, err := grpcQuerier.NextKeyID(sdk.WrapSDKContext(ctx), &types.NextKeyIDRequest{
			Chain:   existingChain.String(),
			KeyRole: tss.SecondaryKey,
		})

		assert := assert.New(t)
		assert.Nil(res)
		s, ok := status.FromError(err)
		assert.Equal(codes.OK, s.Code())
		assert.Equal(true, ok)
	}).Repeat(repeatCount))
}

func TestAssignbleKey(t *testing.T) {
	var (
		tssKeeper       *mock.TSSKeeperMock
		nexusKeeper     *mock.NexusMock
		stakingKeeper   *mock.StakingKeeperMock
		ctx             sdk.Context
		grpcQuerier     *keeper.Querier
		existingChain   nexus.ChainName
		existingKeyRole tss.KeyRole
	)

	setup := func() {
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())

		existingChain = "existing"
		existingKeyRole = tss.MasterKey

		tssKeeper = &mock.TSSKeeperMock{
			GetNextKeyIDFunc: func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.KeyID, bool) {
				if chain.Name == existingChain && keyRole == existingKeyRole {
					return "dummy-key-id", true
				}
				return "", false
			},
		}

		nexusKeeper = &mock.NexusMock{
			GetChainFunc: func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
				if chain == existingChain {
					return nexus.Chain{
						Name:                  existingChain,
						SupportsForeignAssets: false,
						KeyType:               tss.Multisig,
						Module:                "evm",
					}, true
				}
				return nexus.Chain{}, false
			},
		}

		stakingKeeper = &mock.StakingKeeperMock{
			ValidatorFunc: func(ctx sdk.Context, addr sdk.ValAddress) stakingtypes.ValidatorI {
				return stakingtypes.Validator{}
			},
		}

		q := keeper.NewGRPCQuerier(tssKeeper, nexusKeeper, stakingKeeper)
		grpcQuerier = &q
	}

	setup()

	repeatCount := 1

	t.Run("chain and key exist", testutils.Func(func(t *testing.T) {
		expectedRes := types.AssignableKeyResponse{
			Assignable: false,
		}

		res, err := grpcQuerier.AssignableKey(sdk.WrapSDKContext(ctx), &types.AssignableKeyRequest{
			Chain:   existingChain.String(),
			KeyRole: existingKeyRole,
		})

		assert := assert.New(t)
		assert.NoError(err)
		assert.Equal(expectedRes, *res)

		s, ok := status.FromError(err)
		assert.Equal(codes.OK, s.Code())
		assert.Equal(true, ok)
	}).Repeat(repeatCount))

	t.Run("only chain exist", testutils.Func(func(t *testing.T) {
		expectedRes := types.AssignableKeyResponse{
			Assignable: true,
		}

		res, err := grpcQuerier.AssignableKey(sdk.WrapSDKContext(ctx), &types.AssignableKeyRequest{
			Chain:   existingChain.String(),
			KeyRole: tss.SecondaryKey,
		})

		assert := assert.New(t)
		assert.NoError(err)

		assert.Equal(expectedRes, *res)

		s, ok := status.FromError(err)
		assert.Equal(codes.OK, s.Code())
		assert.Equal(true, ok)
	}).Repeat(repeatCount))

	t.Run("chain does not exist", testutils.Func(func(t *testing.T) {
		res, err := grpcQuerier.AssignableKey(sdk.WrapSDKContext(ctx), &types.AssignableKeyRequest{
			Chain:   "non-existing-chain",
			KeyRole: tss.MasterKey,
		})

		assert := assert.New(t)
		assert.Nil(res)
		s, ok := status.FromError(err)
		assert.Equal(codes.NotFound, s.Code())
		assert.Equal(true, ok)
	}).Repeat(repeatCount))
}
