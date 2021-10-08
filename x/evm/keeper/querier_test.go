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
	btc "github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	"github.com/axelarnetwork/axelar-core/x/evm/exported"
	evmKeeper "github.com/axelarnetwork/axelar-core/x/evm/keeper"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/evm/types/mock"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

func TestQueryTokenAddress(t *testing.T) {

	var (
		chainKeeper     *mock.ChainKeeperMock
		nexusKeeper     *mock.NexusMock
		ctx             sdk.Context
		evmChain        string
		asset           string
		expectedAddress common.Address
	)

	setup := func() {
		evmChain = rand.StrBetween(5, 10)
		expectedAddress = randomAddress()
		asset = btc.Bitcoin.NativeAsset

		chainKeeper = &mock.ChainKeeperMock{
			GetNameFunc:           func() string { return evmChain },
			GetGatewayAddressFunc: func(sdk.Context) (common.Address, bool) { return randomAddress(), true },
			GetERC20TokenFunc: func(ctx sdk.Context, a string) types.ERC20Token {
				if asset == a {
					return createMockConfirmedERC20Token(asset, types.Address(expectedAddress), createDetails())
				}
				return types.NilToken
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

		res, err := evmKeeper.QueryTokenAddress(ctx, chainKeeper, nexusKeeper, asset)

		assert := assert.New(t)
		assert.NoError(err)
		assert.Len(chainKeeper.GetERC20TokenCalls(), 1)
		assert.Equal(expectedAddress.Bytes(), res)

	}).Repeat(repeatCount))

	t.Run("token not deployed", testutils.Func(func(t *testing.T) {
		setup()
		chainKeeper.GetERC20TokenFunc = func(ctx sdk.Context, asset string) types.ERC20Token {
			return types.NilToken
		}

		_, err := evmKeeper.QueryTokenAddress(ctx, chainKeeper, nexusKeeper, asset)

		assert := assert.New(t)
		assert.Error(err)

	}).Repeat(repeatCount))

}

func TestQueryDepositState(t *testing.T) {
	var (
		ctx             sdk.Context
		evmChain        string
		expectedDeposit types.ERC20Deposit
		chainKeeper     *mock.ChainKeeperMock
		nexusKeeper     *mock.NexusMock
	)

	setup := func() {
		evmChain = rand.StrBetween(5, 10)
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())

		expectedDeposit = types.ERC20Deposit{
			DestinationChain: rand.StrBetween(5, 10),
			Amount:           sdk.NewUint(uint64(rand.I64Between(100, 10000))),
			BurnerAddress:    types.Address(randomAddress()),
			TxID:             types.Hash(randomHash()),
			Asset:            rand.StrBetween(5, 10),
		}

		chainKeeper = &mock.ChainKeeperMock{
			GetNameFunc: func() string { return evmChain },
			GetPendingDepositFunc: func(sdk.Context, vote.PollKey) (types.ERC20Deposit, bool) {
				return types.ERC20Deposit{}, false
			},
			GetDepositFunc: func(_ sdk.Context, txID common.Hash, burnerAddr common.Address) (types.ERC20Deposit, types.DepositState, bool) {
				return types.ERC20Deposit{}, 0, false
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
	}
	repeatCount := 20
	t.Run("no deposit", testutils.Func(func(t *testing.T) {
		setup()
		data := types.ModuleCdc.MustMarshalJSON(&expectedDeposit)
		res, err := evmKeeper.QueryDepositState(ctx, chainKeeper, nexusKeeper, data)

		assert := assert.New(t)
		assert.NoError(err)
		assert.Len(chainKeeper.GetNameCalls(), 1)
		assert.Len(chainKeeper.GetPendingDepositCalls(), 1)
		assert.Len(chainKeeper.GetDepositCalls(), 1)
		assert.Len(nexusKeeper.GetChainCalls(), 1)

		var depositState types.QueryDepositStateResponse
		types.ModuleCdc.MustUnmarshalLengthPrefixed(res, &depositState)
		assert.Equal(types.DepositStatus_None, depositState.Status)
	}).Repeat(repeatCount))

	t.Run("deposit pending", testutils.Func(func(t *testing.T) {
		setup()
		pollKey := vote.NewPollKey(types.ModuleName, fmt.Sprintf("%s_%s_%d", expectedDeposit.TxID.Hex(), expectedDeposit.BurnerAddress.Hex(), expectedDeposit.Amount.Uint64()))
		chainKeeper.GetPendingDepositFunc = func(_ sdk.Context, k vote.PollKey) (types.ERC20Deposit, bool) {
			if pollKey == k {
				return expectedDeposit, true
			}
			return types.ERC20Deposit{}, false
		}

		data := types.ModuleCdc.MustMarshalJSON(&expectedDeposit)
		res, err := evmKeeper.QueryDepositState(ctx, chainKeeper, nexusKeeper, data)

		assert := assert.New(t)
		assert.NoError(err)
		assert.Len(chainKeeper.GetNameCalls(), 1)
		assert.Len(chainKeeper.GetPendingDepositCalls(), 1)
		assert.Len(chainKeeper.GetDepositCalls(), 1)
		assert.Len(nexusKeeper.GetChainCalls(), 1)

		var depositState types.QueryDepositStateResponse
		types.ModuleCdc.MustUnmarshalLengthPrefixed(res, &depositState)
		assert.Equal(types.DepositStatus_Pending, depositState.Status)

	}).Repeat(repeatCount))

	t.Run("deposit confirmed", testutils.Func(func(t *testing.T) {
		setup()
		chainKeeper.GetDepositFunc = func(_ sdk.Context, txID common.Hash, burnerAddr common.Address) (types.ERC20Deposit, types.DepositState, bool) {
			if types.Hash(txID) == expectedDeposit.TxID && types.Address(burnerAddr) == expectedDeposit.BurnerAddress {
				return expectedDeposit, types.CONFIRMED, true
			}
			return types.ERC20Deposit{}, 0, false
		}

		data := types.ModuleCdc.MustMarshalJSON(&expectedDeposit)
		res, err := evmKeeper.QueryDepositState(ctx, chainKeeper, nexusKeeper, data)

		assert := assert.New(t)
		assert.NoError(err)
		assert.Len(chainKeeper.GetNameCalls(), 1)
		assert.Len(chainKeeper.GetPendingDepositCalls(), 1)
		assert.Len(chainKeeper.GetDepositCalls(), 1)
		assert.Len(nexusKeeper.GetChainCalls(), 1)

		var depositState types.QueryDepositStateResponse
		types.ModuleCdc.MustUnmarshalLengthPrefixed(res, &depositState)
		assert.Equal(types.DepositStatus_Confirmed, depositState.Status)

	}).Repeat(repeatCount))

	t.Run("deposit burned", testutils.Func(func(t *testing.T) {
		setup()
		chainKeeper.GetDepositFunc = func(_ sdk.Context, txID common.Hash, burnerAddr common.Address) (types.ERC20Deposit, types.DepositState, bool) {
			if types.Hash(txID) == expectedDeposit.TxID && types.Address(burnerAddr) == expectedDeposit.BurnerAddress {
				return expectedDeposit, types.BURNED, true
			}
			return types.ERC20Deposit{}, 0, false
		}

		data := types.ModuleCdc.MustMarshalJSON(&expectedDeposit)
		res, err := evmKeeper.QueryDepositState(ctx, chainKeeper, nexusKeeper, data)

		assert := assert.New(t)
		assert.NoError(err)
		assert.Len(chainKeeper.GetNameCalls(), 1)
		assert.Len(chainKeeper.GetPendingDepositCalls(), 1)
		assert.Len(chainKeeper.GetDepositCalls(), 1)
		assert.Len(nexusKeeper.GetChainCalls(), 1)

		var depositState types.QueryDepositStateResponse
		types.ModuleCdc.MustUnmarshalLengthPrefixed(res, &depositState)
		assert.Equal(types.DepositStatus_Burned, depositState.Status)

	}).Repeat(repeatCount))

	t.Run("unmarshaling error", testutils.Func(func(t *testing.T) {
		setup()
		data := rand.BytesBetween(10, 50)
		_, err := evmKeeper.QueryDepositState(ctx, chainKeeper, nexusKeeper, data)

		assert := assert.New(t)
		assert.EqualError(err, "could not unmarshal parameters: bridge error")

	}).Repeat(repeatCount))

	t.Run("chain not registered", testutils.Func(func(t *testing.T) {
		setup()
		data := rand.BytesBetween(10, 50)
		nexusKeeper.GetChainFunc = func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
			return nexus.Chain{}, false
		}
		_, err := evmKeeper.QueryDepositState(ctx, chainKeeper, nexusKeeper, data)

		assert := assert.New(t)
		assert.EqualError(err, fmt.Sprintf("%s is not a registered chain: bridge error", evmChain))

	}).Repeat(repeatCount))
}

func TestQueryDepositAddress(t *testing.T) {

	var (
		chainKeeper     *mock.ChainKeeperMock
		nexusKeeper     *mock.NexusMock
		ctx             sdk.Context
		evmChain        string
		data            []byte
		expectedAddress common.Address
	)

	setup := func() {
		evmChain = rand.StrBetween(5, 10)
		expectedAddress = randomAddress()

		chainKeeper = &mock.ChainKeeperMock{
			GetNameFunc:           func() string { return evmChain },
			GetGatewayAddressFunc: func(sdk.Context) (common.Address, bool) { return randomAddress(), true },
			GetERC20TokenFunc: func(ctx sdk.Context, a string) types.ERC20Token {
				if btc.Bitcoin.NativeAsset == a {
					return createMockConfirmedERC20Token(a, types.Address(expectedAddress), createDetails())
				}
				return types.NilToken
			},
			GetBurnerAddressAndSaltFunc: func(sdk.Context, types.Address, string, common.Address) (common.Address, common.Hash, error) {
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
			Chain:   btc.Bitcoin.Name,
			Address: "tb1qg2z5jatp22zg7wyhpthhgwvn0un05mdwmqgjln",
			Asset:   btc.Bitcoin.NativeAsset,
		})

	}

	repeatCount := 20

	t.Run("happy path hard coded", testutils.Func(func(t *testing.T) {
		setup()

		res, err := evmKeeper.QueryDepositAddress(ctx, chainKeeper, nexusKeeper, data)

		assert := assert.New(t)
		assert.NoError(err)
		assert.Len(chainKeeper.GetBurnerAddressAndSaltCalls(), 1)
		assert.Len(nexusKeeper.GetRecipientCalls(), 1)
		assert.Equal(expectedAddress.Bytes(), res)

	}).Repeat(repeatCount))

	t.Run("happy path", testutils.Func(func(t *testing.T) {
		setup()
		dataStr := &types.DepositQueryParams{
			Chain:   rand.StrBetween(5, 20),
			Address: "tb" + rand.HexStr(40),
			Asset:   rand.StrBetween(3, 8),
		}
		data = types.ModuleCdc.MustMarshalJSON(dataStr)
		chainKeeper.GetERC20TokenFunc = func(ctx sdk.Context, a string) types.ERC20Token {
			if dataStr.Asset == a {
				return createMockConfirmedERC20Token(a, types.Address(expectedAddress), createDetails())
			}
			return types.NilToken
		}

		res, err := evmKeeper.QueryDepositAddress(ctx, chainKeeper, nexusKeeper, data)

		assert := assert.New(t)
		assert.NoError(err)
		assert.Len(chainKeeper.GetBurnerAddressAndSaltCalls(), 1)
		assert.Len(nexusKeeper.GetRecipientCalls(), 1)
		assert.Equal(expectedAddress.Bytes(), res)

	}).Repeat(repeatCount))

	t.Run("gateway not deployed", testutils.Func(func(t *testing.T) {
		setup()
		chainKeeper.GetGatewayAddressFunc = func(sdk.Context) (common.Address, bool) { return common.Address{}, false }

		_, err := evmKeeper.QueryDepositAddress(ctx, chainKeeper, nexusKeeper, data)

		assert := assert.New(t)
		assert.Error(err)

	}).Repeat(repeatCount))

	t.Run("token contract not deployed", testutils.Func(func(t *testing.T) {
		setup()
		chainKeeper.GetERC20TokenFunc = func(ctx sdk.Context, a string) types.ERC20Token {
			return types.NilToken
		}

		_, err := evmKeeper.QueryDepositAddress(ctx, chainKeeper, nexusKeeper, data)

		assert := assert.New(t)
		assert.Error(err)

	}).Repeat(repeatCount))

	t.Run("cannot get deposit address", testutils.Func(func(t *testing.T) {
		setup()
		chainKeeper.GetBurnerAddressAndSaltFunc = func(sdk.Context, types.Address, string, common.Address) (common.Address, common.Hash, error) {
			return common.Address{}, common.Hash{}, fmt.Errorf("could not find deposit address")
		}

		_, err := evmKeeper.QueryDepositAddress(ctx, chainKeeper, nexusKeeper, data)

		assert := assert.New(t)
		assert.Error(err)

	}).Repeat(repeatCount))

	t.Run("chain not registered", testutils.Func(func(t *testing.T) {
		setup()
		nexusKeeper.GetChainFunc = func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
			return nexus.Chain{}, false
		}
		_, err := evmKeeper.QueryDepositAddress(ctx, chainKeeper, nexusKeeper, data)

		assert := assert.New(t)
		assert.Error(err)

	}).Repeat(repeatCount))

	t.Run("deposit address not linked", testutils.Func(func(t *testing.T) {
		setup()
		nexusKeeper.GetRecipientFunc = func(sdk.Context, nexus.CrossChainAddress) (nexus.CrossChainAddress, bool) {
			return nexus.CrossChainAddress{}, false
		}
		_, err := evmKeeper.QueryDepositAddress(ctx, chainKeeper, nexusKeeper, data)

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

func createMockConfirmedERC20Token(asset string, addr types.Address, details types.TokenDetails) types.ERC20Token {
	meta := types.ERC20TokenMetadata{
		Asset:        asset,
		Details:      details,
		Status:       types.Confirmed,
		TokenAddress: addr,
		ChainID:      sdk.NewIntFromUint64(uint64(rand.I64Between(1, 10))),
	}
	return types.CreateERC20Token(
		func(meta types.ERC20TokenMetadata) {},
		meta,
	)
}
