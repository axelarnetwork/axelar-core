package keeper

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/evm/exported"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/evm/types/mock"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

func TestQueryTokenAddress(t *testing.T) {

	var (
		ethKeeper       *mock.EVMKeeperMock
		nexusKeeper     *mock.NexusMock
		ctx             sdk.Context
		evmChain        string
		symbol          string
		expectedAddress common.Address
	)

	setup := func() {
		expectedAddress = randomAddress()

		ethKeeper = &mock.EVMKeeperMock{
			GetGatewayAddressFunc: func(ctx sdk.Context, evmChain string) (common.Address, bool) { return randomAddress(), true },
			GetTokenAddressFunc: func(ctx sdk.Context, evmChain string, symbol string, gatewayAddr common.Address) (common.Address, error) {
				return expectedAddress, nil
			},
		}
		nexusKeeper = &mock.NexusMock{
			GetChainFunc: func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
				return nexus.Chain{
					Name:                  chain,
					NativeAsset:           rand.StrBetween(5, 20),
					SupportsForeignAssets: true,
				}, true
			},
		}
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())
		evmChain = exported.Ethereum.Name
	}

	repeatCount := 20

	t.Run("happy path", testutils.Func(func(t *testing.T) {
		setup()

		res, err := queryTokenAddress(ctx, ethKeeper, nexusKeeper, evmChain, symbol)

		assert := assert.New(t)
		assert.NoError(err)
		assert.Len(ethKeeper.GetTokenAddressCalls(), 1)
		assert.Equal(expectedAddress.Bytes(), res)

	}).Repeat(repeatCount))

	t.Run("gateway not set", testutils.Func(func(t *testing.T) {
		setup()
		ethKeeper.GetGatewayAddressFunc = func(ctx sdk.Context, evmChain string) (common.Address, bool) { return common.Address{}, false }

		_, err := queryTokenAddress(ctx, ethKeeper, nexusKeeper, evmChain, symbol)

		assert := assert.New(t)
		assert.Error(err)

	}).Repeat(repeatCount))

	t.Run("token not deployed", testutils.Func(func(t *testing.T) {
		setup()
		ethKeeper.GetTokenAddressFunc = func(ctx sdk.Context, evmChain string, symbol string, gatewayAddr common.Address) (common.Address, error) {
			return common.Address{}, fmt.Errorf("could not find token address")
		}

		_, err := queryTokenAddress(ctx, ethKeeper, nexusKeeper, evmChain, symbol)

		assert := assert.New(t)
		assert.Error(err)

	}).Repeat(repeatCount))

}

func TestQueryDepositAddress(t *testing.T) {

	var (
		ethKeeper       *mock.EVMKeeperMock
		nexusKeeper     *mock.NexusMock
		ctx             sdk.Context
		evmChain        string
		data            []byte
		expectedAddress common.Address
	)

	setup := func() {
		expectedAddress = randomAddress()

		ethKeeper = &mock.EVMKeeperMock{
			GetGatewayAddressFunc: func(ctx sdk.Context, evmChain string) (common.Address, bool) { return randomAddress(), true },
			GetTokenAddressFunc: func(ctx sdk.Context, evmChain string, symbol string, gatewayAddr common.Address) (common.Address, error) {
				return randomAddress(), nil
			},
			GetBurnerAddressAndSaltFunc: func(ctx sdk.Context, evmChain string, tokenAddr common.Address, recipient string, gatewayAddr common.Address) (common.Address, common.Hash, error) {
				return expectedAddress, randomHash(), nil
			},
		}
		nexusKeeper = &mock.NexusMock{
			GetChainFunc: func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
				return nexus.Chain{
					Name:                  chain,
					NativeAsset:           rand.StrBetween(5, 20),
					SupportsForeignAssets: true,
				}, true
			},
			GetRecipientFunc: func(sdk.Context, nexus.CrossChainAddress) (nexus.CrossChainAddress, bool) {
				return nexus.CrossChainAddress{
					Chain:   exported.Ethereum,
					Address: randomAddress().String(),
				}, true
			},
		}
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())
		evmChain = exported.Ethereum.Name
		data = types.ModuleCdc.MustMarshalJSON(&types.DepositQueryParams{
			Chain:   "bitcoin",
			Address: "tb1qg2z5jatp22zg7wyhpthhgwvn0un05mdwmqgjln",
			Symbol:  "satoshi",
		})

	}

	repeatCount := 20

	t.Run("happy path hard coded", testutils.Func(func(t *testing.T) {
		setup()

		res, err := queryDepositAddress(ctx, ethKeeper, nexusKeeper, evmChain, data)

		assert := assert.New(t)
		assert.NoError(err)
		assert.Len(ethKeeper.GetBurnerAddressAndSaltCalls(), 1)
		assert.Len(nexusKeeper.GetRecipientCalls(), 1)
		assert.Equal(expectedAddress.Bytes(), res)

	}).Repeat(repeatCount))

	t.Run("happy path", testutils.Func(func(t *testing.T) {
		setup()
		dataStr := &types.DepositQueryParams{
			Chain:   rand.StrBetween(5, 20),
			Address: "tb" + rand.HexStr(40),
			Symbol:  rand.StrBetween(3, 8),
		}
		data = types.ModuleCdc.MustMarshalJSON(dataStr)

		res, err := queryDepositAddress(ctx, ethKeeper, nexusKeeper, evmChain, data)

		assert := assert.New(t)
		assert.NoError(err)
		assert.Len(ethKeeper.GetBurnerAddressAndSaltCalls(), 1)
		assert.Len(nexusKeeper.GetRecipientCalls(), 1)
		assert.Equal(expectedAddress.Bytes(), res)

	}).Repeat(repeatCount))

	t.Run("gateway not deployed", testutils.Func(func(t *testing.T) {
		setup()
		ethKeeper.GetGatewayAddressFunc = func(ctx sdk.Context, evmChain string) (common.Address, bool) { return common.Address{}, false }

		_, err := queryDepositAddress(ctx, ethKeeper, nexusKeeper, evmChain, data)

		assert := assert.New(t)
		assert.Error(err)

	}).Repeat(repeatCount))

	t.Run("token contract not deployed", testutils.Func(func(t *testing.T) {
		setup()
		ethKeeper.GetTokenAddressFunc = func(ctx sdk.Context, evmChain string, symbol string, gatewayAddr common.Address) (common.Address, error) {
			return common.Address{}, fmt.Errorf("could not find token address")
		}

		_, err := queryDepositAddress(ctx, ethKeeper, nexusKeeper, evmChain, data)

		assert := assert.New(t)
		assert.Error(err)

	}).Repeat(repeatCount))

	t.Run("cannot get deposit address", testutils.Func(func(t *testing.T) {
		setup()
		ethKeeper.GetBurnerAddressAndSaltFunc = func(ctx sdk.Context, evmChain string, tokenAddr common.Address, recipient string, gatewayAddr common.Address) (common.Address, common.Hash, error) {
			return common.Address{}, common.Hash{}, fmt.Errorf("could not find deposit address")
		}

		_, err := queryDepositAddress(ctx, ethKeeper, nexusKeeper, evmChain, data)

		assert := assert.New(t)
		assert.Error(err)

	}).Repeat(repeatCount))

	t.Run("Ethereum chain not registered", testutils.Func(func(t *testing.T) {
		setup()
		nexusKeeper.GetChainFunc = func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
			return nexus.Chain{}, false
		}
		_, err := queryDepositAddress(ctx, ethKeeper, nexusKeeper, evmChain, data)

		assert := assert.New(t)
		assert.Error(err)

	}).Repeat(repeatCount))

	t.Run("deposit address not linked", testutils.Func(func(t *testing.T) {
		setup()
		nexusKeeper.GetRecipientFunc = func(sdk.Context, nexus.CrossChainAddress) (nexus.CrossChainAddress, bool) {
			return nexus.CrossChainAddress{}, false
		}
		_, err := queryDepositAddress(ctx, ethKeeper, nexusKeeper, evmChain, data)

		assert := assert.New(t)
		assert.Error(err)

	}).Repeat(repeatCount))

}

func randomAddress() common.Address {
	return common.BytesToAddress(rand.Bytes(common.AddressLength))
}

func randomHash() common.Hash {
	return common.BytesToHash(rand.Bytes(common.HashLength))
}
