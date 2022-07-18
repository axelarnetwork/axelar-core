package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	multisig "github.com/axelarnetwork/axelar-core/x/multisig/exported"
	"github.com/axelarnetwork/axelar-core/x/multisig/keeper"
	"github.com/axelarnetwork/axelar-core/x/multisig/types"
	"github.com/axelarnetwork/axelar-core/x/multisig/types/mock"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
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

	t.Run("if chain does not exist, should get empty response", testutils.Func(func(t *testing.T) {
		chain := "non-existing-chain"
		res, err := grpcQuerier.KeyID(sdk.WrapSDKContext(ctx), &types.KeyIDRequest{
			Chain: chain,
		})

		assert := assert.New(t)
		assert.Nil(err)
		assert.Equal(res.KeyID, multisig.KeyID(""))
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

	t.Run("if chain does not exist, should get empty response", testutils.Func(func(t *testing.T) {
		chain := "non-existing-chain"
		res, err := grpcQuerier.NextKeyID(sdk.WrapSDKContext(ctx), &types.NextKeyIDRequest{
			Chain: chain,
		})

		assert := assert.New(t)
		assert.Nil(err)
		assert.Equal(res.KeyID, multisig.KeyID(""))
	}).Repeat(repeatCount))
}
