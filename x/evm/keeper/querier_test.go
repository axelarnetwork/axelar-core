package keeper_test

import (
	"fmt"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/evm/exported"
	evmKeeper "github.com/axelarnetwork/axelar-core/x/evm/keeper"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/evm/types/mock"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

func TestQueryTokenAddress(t *testing.T) {

	var (
		ethKeeper       *mock.ChainKeeperMock
		nexusKeeper     *mock.NexusMock
		ctx             sdk.Context
		evmChain        string
		symbol          string
		expectedAddress common.Address
	)

	setup := func() {
		evmChain = rand.StrBetween(5, 10)
		expectedAddress = randomAddress()

		ethKeeper = &mock.ChainKeeperMock{
			GetNameFunc:           func() string { return evmChain },
			GetGatewayAddressFunc: func(sdk.Context) (common.Address, bool) { return randomAddress(), true },
			GetTokenAddressFunc: func(sdk.Context, string, common.Address) (common.Address, error) {
				return expectedAddress, nil
			},
		}
		nexusKeeper = &mock.NexusMock{
			GetChainFunc: func(_ sdk.Context, chain string) (nexus.Chain, bool) {
				if strings.ToLower(chain) == strings.ToLower(evmChain) {
					return nexus.Chain{
						Name:                  chain,
						NativeAsset:           rand.StrBetween(5, 20),
						SupportsForeignAssets: true,
					}, true
				}
				return nexus.Chain{}, false
			},
		}
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())
		evmChain = exported.Ethereum.Name
	}

	repeatCount := 20

	t.Run("happy path", testutils.Func(func(t *testing.T) {
		setup()

		res, err := evmKeeper.QueryTokenAddress(ctx, ethKeeper, nexusKeeper, symbol)

		assert := assert.New(t)
		assert.NoError(err)
		assert.Len(ethKeeper.GetTokenAddressCalls(), 1)
		assert.Equal(expectedAddress.Bytes(), res)

	}).Repeat(repeatCount))

	t.Run("gateway not set", testutils.Func(func(t *testing.T) {
		setup()
		ethKeeper.GetGatewayAddressFunc = func(sdk.Context) (common.Address, bool) { return common.Address{}, false }

		_, err := evmKeeper.QueryTokenAddress(ctx, ethKeeper, nexusKeeper, symbol)

		assert := assert.New(t)
		assert.Error(err)

	}).Repeat(repeatCount))

	t.Run("token not deployed", testutils.Func(func(t *testing.T) {
		setup()
		ethKeeper.GetTokenAddressFunc = func(sdk.Context, string, common.Address) (common.Address, error) {
			return common.Address{}, fmt.Errorf("could not find token address")
		}

		_, err := evmKeeper.QueryTokenAddress(ctx, ethKeeper, nexusKeeper, symbol)

		assert := assert.New(t)
		assert.Error(err)

	}).Repeat(repeatCount))

}

func TestQueryDepositAddress(t *testing.T) {

	var (
		ethKeeper       *mock.ChainKeeperMock
		nexusKeeper     *mock.NexusMock
		ctx             sdk.Context
		evmChain        string
		data            []byte
		expectedAddress common.Address
	)

	setup := func() {
		evmChain = rand.StrBetween(5, 10)
		expectedAddress = randomAddress()

		ethKeeper = &mock.ChainKeeperMock{
			GetNameFunc:           func() string { return evmChain },
			GetGatewayAddressFunc: func(sdk.Context) (common.Address, bool) { return randomAddress(), true },
			GetTokenAddressFunc: func(sdk.Context, string, common.Address) (common.Address, error) {
				return randomAddress(), nil
			},
			GetBurnerAddressAndSaltFunc: func(sdk.Context, common.Address, string, common.Address) (common.Address, common.Hash, error) {
				return expectedAddress, randomHash(), nil
			},
		}
		nexusKeeper = &mock.NexusMock{
			GetChainFunc: func(_ sdk.Context, chain string) (nexus.Chain, bool) {
				if strings.ToLower(chain) == strings.ToLower(evmChain) {
					return nexus.Chain{
						Name:                  chain,
						NativeAsset:           rand.StrBetween(5, 20),
						SupportsForeignAssets: true,
					}, true
				}
				return nexus.Chain{}, false
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

		res, err := evmKeeper.QueryDepositAddress(ctx, ethKeeper, nexusKeeper, data)

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

		res, err := evmKeeper.QueryDepositAddress(ctx, ethKeeper, nexusKeeper, data)

		assert := assert.New(t)
		assert.NoError(err)
		assert.Len(ethKeeper.GetBurnerAddressAndSaltCalls(), 1)
		assert.Len(nexusKeeper.GetRecipientCalls(), 1)
		assert.Equal(expectedAddress.Bytes(), res)

	}).Repeat(repeatCount))

	t.Run("gateway not deployed", testutils.Func(func(t *testing.T) {
		setup()
		ethKeeper.GetGatewayAddressFunc = func(sdk.Context) (common.Address, bool) { return common.Address{}, false }

		_, err := evmKeeper.QueryDepositAddress(ctx, ethKeeper, nexusKeeper, data)

		assert := assert.New(t)
		assert.Error(err)

	}).Repeat(repeatCount))

	t.Run("token contract not deployed", testutils.Func(func(t *testing.T) {
		setup()
		ethKeeper.GetTokenAddressFunc = func(sdk.Context, string, common.Address) (common.Address, error) {
			return common.Address{}, fmt.Errorf("could not find token address")
		}

		_, err := evmKeeper.QueryDepositAddress(ctx, ethKeeper, nexusKeeper, data)

		assert := assert.New(t)
		assert.Error(err)

	}).Repeat(repeatCount))

	t.Run("cannot get deposit address", testutils.Func(func(t *testing.T) {
		setup()
		ethKeeper.GetBurnerAddressAndSaltFunc = func(sdk.Context, common.Address, string, common.Address) (common.Address, common.Hash, error) {
			return common.Address{}, common.Hash{}, fmt.Errorf("could not find deposit address")
		}

		_, err := evmKeeper.QueryDepositAddress(ctx, ethKeeper, nexusKeeper, data)

		assert := assert.New(t)
		assert.Error(err)

	}).Repeat(repeatCount))

	t.Run("Ethereum chain not registered", testutils.Func(func(t *testing.T) {
		setup()
		nexusKeeper.GetChainFunc = func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
			return nexus.Chain{}, false
		}
		_, err := evmKeeper.QueryDepositAddress(ctx, ethKeeper, nexusKeeper, data)

		assert := assert.New(t)
		assert.Error(err)

	}).Repeat(repeatCount))

	t.Run("deposit address not linked", testutils.Func(func(t *testing.T) {
		setup()
		nexusKeeper.GetRecipientFunc = func(sdk.Context, nexus.CrossChainAddress) (nexus.CrossChainAddress, bool) {
			return nexus.CrossChainAddress{}, false
		}
		_, err := evmKeeper.QueryDepositAddress(ctx, ethKeeper, nexusKeeper, data)

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
