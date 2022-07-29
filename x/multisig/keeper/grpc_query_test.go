package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	multisig "github.com/axelarnetwork/axelar-core/x/multisig/exported"
	multisigTestutils "github.com/axelarnetwork/axelar-core/x/multisig/exported/testutils"
	"github.com/axelarnetwork/axelar-core/x/multisig/keeper"
	"github.com/axelarnetwork/axelar-core/x/multisig/types"
	"github.com/axelarnetwork/axelar-core/x/multisig/types/mock"
	typesTestutils "github.com/axelarnetwork/axelar-core/x/multisig/types/testutils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	. "github.com/axelarnetwork/utils/test"
)

func TestKeyID(t *testing.T) {
	var (
		multisigKeeper *mock.KeeperMock
		stakingKeeper  *mock.StakerMock
		ctx            sdk.Context
		grpcQuerier    *keeper.Querier
		existingChain  nexus.ChainName
		existingKeyID  multisig.KeyID
	)

	setup := func() {
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())

		existingChain = "existing"
		existingKeyID = multisig.KeyID("keyID")

		multisigKeeper = &mock.KeeperMock{
			GetCurrentKeyIDFunc: func(ctx sdk.Context, chain nexus.ChainName) (multisig.KeyID, bool) {
				if chain == existingChain {
					return existingKeyID, true
				}
				return "", false
			},
		}

		stakingKeeper = &mock.StakerMock{}

		q := keeper.NewGRPCQuerier(multisigKeeper, stakingKeeper)
		grpcQuerier = &q
	}

	setup()

	repeatCount := 1

	t.Run("when chain exists get the keyID", testutils.Func(func(t *testing.T) {
		expectedRes := types.KeyIDResponse{
			KeyID: existingKeyID,
		}

		res, err := grpcQuerier.KeyID(sdk.WrapSDKContext(ctx), &types.KeyIDRequest{
			Chain: existingChain.String(),
		})

		assert := assert.New(t)
		assert.NoError(err)

		assert.Equal(expectedRes, *res)
	}).Repeat(repeatCount))

	t.Run("if chain does not exist, should get NotFound grpc code", testutils.Func(func(t *testing.T) {
		chain := "non-existing-chain"
		res, err := grpcQuerier.KeyID(sdk.WrapSDKContext(ctx), &types.KeyIDRequest{
			Chain: chain,
		})

		assert := assert.New(t)
		assert.Nil(res)
		s, ok := status.FromError(err)
		assert.Equal(codes.NotFound, s.Code())
		assert.Equal(true, ok)
	}).Repeat(repeatCount))
}

func TestNextKeyID(t *testing.T) {
	var (
		multisigKeeper *mock.KeeperMock
		stakingKeeper  *mock.StakerMock
		ctx            sdk.Context
		grpcQuerier    *keeper.Querier
		existingChain  nexus.ChainName
		existingKeyID  multisig.KeyID
	)

	setup := func() {
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())

		existingChain = "existing"
		existingKeyID = multisig.KeyID("keyID")

		multisigKeeper = &mock.KeeperMock{
			GetNextKeyIDFunc: func(ctx sdk.Context, chain nexus.ChainName) (multisig.KeyID, bool) {
				if chain == existingChain {
					return existingKeyID, true
				}
				return "", false
			},
		}

		stakingKeeper = &mock.StakerMock{}

		q := keeper.NewGRPCQuerier(multisigKeeper, stakingKeeper)
		grpcQuerier = &q
	}

	setup()

	repeatCount := 1

	t.Run("when chain exists get the keyID", testutils.Func(func(t *testing.T) {
		expectedRes := types.NextKeyIDResponse{
			KeyID: existingKeyID,
		}

		res, err := grpcQuerier.NextKeyID(sdk.WrapSDKContext(ctx), &types.NextKeyIDRequest{
			Chain: existingChain.String(),
		})

		assert := assert.New(t)
		assert.NoError(err)

		assert.Equal(expectedRes, *res)
	}).Repeat(repeatCount))

	t.Run("if chain does not exist, should get NotFound grpc code", testutils.Func(func(t *testing.T) {
		chain := "non-existing-chain"
		res, err := grpcQuerier.NextKeyID(sdk.WrapSDKContext(ctx), &types.NextKeyIDRequest{
			Chain: chain,
		})

		assert := assert.New(t)
		assert.Nil(res)
		s, ok := status.FromError(err)
		assert.Equal(codes.NotFound, s.Code())
		assert.Equal(true, ok)
	}).Repeat(repeatCount))
}

func TestKey(t *testing.T) {
	var (
		multisigKeeper *mock.KeeperMock
		stakingKeeper  *mock.StakerMock
		ctx            sdk.Context
		querier        keeper.Querier
		key            types.Key
	)

	givenQuerier := Given("multisig querier", func() {
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())
		multisigKeeper = &mock.KeeperMock{}
		stakingKeeper = &mock.StakerMock{}

		querier = keeper.NewGRPCQuerier(multisigKeeper, stakingKeeper)
	})

	givenQuerier.
		When("key is not found", func() {
			multisigKeeper.GetKeyFunc = func(sdk.Context, multisig.KeyID) (multisig.Key, bool) { return nil, false }
		}).
		Then("should return error NotFound", func(t *testing.T) {
			res, err := querier.Key(sdk.WrapSDKContext(ctx), &types.KeyRequest{KeyID: multisigTestutils.KeyID()})

			assert.Nil(t, res)
			s, ok := status.FromError(err)
			assert.Equal(t, codes.NotFound, s.Code())
			assert.True(t, ok)
		}).
		Run(t)

	givenQuerier.
		When("key is found", func() {
			key = typesTestutils.Key()
			multisigKeeper.GetKeyFunc = func(sdk.Context, multisig.KeyID) (multisig.Key, bool) { return &key, true }
		}).
		Then("should return error NotFound", func(t *testing.T) {
			res, err := querier.Key(sdk.WrapSDKContext(ctx), &types.KeyRequest{KeyID: key.ID})

			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, key.ID, res.KeyID)
			assert.Equal(t, key.State, res.State)
			assert.Equal(t, key.GetHeight(), res.Height)
			assert.Equal(t, key.GetTimestamp(), res.Timestamp)
			assert.Equal(t, key.GetMinPassingWeight(), res.ThresholdWeight)
			assert.Equal(t, key.GetBondedWeight(), res.BondedWeight)
			assert.Len(t, res.Participants, len(key.GetParticipants()))

			for i, p := range res.Participants {
				if i < len(res.Participants)-1 {
					assert.True(t, p.Weight.GTE(res.Participants[i+1].Weight))
				}
			}
		}).
		Run(t)
}
