package keeper_test

import (
	"fmt"
	"math/big"
	"strings"
	"testing"

	evmTest "github.com/axelarnetwork/axelar-core/x/evm/types/testutils"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	tssTestUtils "github.com/axelarnetwork/axelar-core/x/tss/exported/testutils"
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

func TestQueryPendingCommands(t *testing.T) {
	var (
		chainKeeper *mock.ChainKeeperMock
		nexusKeeper *mock.NexusMock
		ctx         sdk.Context
		evmChain    string
		asset       string
		symbol      string
		chainID     *big.Int
		keyID       tss.KeyID
		cmds        []types.Command
	)

	setup := func() {
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())
		evmChain = rand.StrBetween(5, 10)
		asset = btc.Satoshi
		symbol = "axelarBTC"
		chainID = big.NewInt(1)
		keyID = tssTestUtils.RandKeyID()
		cmdDeploy, _ := types.CreateDeployTokenCommand(chainID, keyID, createDetails(asset, symbol), types.ZeroAddress)
		cmdMint, _ := types.CreateMintTokenCommand(keyID, types.NewCommandID(rand.Bytes(10), chainID), symbol, common.BytesToAddress(rand.Bytes(common.AddressLength)), big.NewInt(rand.I64Between(1000, 100000)))
		cmdBurn, _ := types.CreateBurnTokenCommand(chainID, keyID, ctx.BlockHeight(), types.BurnerInfo{
			BurnerAddress: types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength))),
			TokenAddress:  types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength))),
			Symbol:        symbol,
			Salt:          types.Hash(common.BytesToHash(rand.Bytes(common.HashLength))),
		})
		cmds = append(cmds, cmdDeploy, cmdMint, cmdBurn)

		chainKeeper = &mock.ChainKeeperMock{
			GetNameFunc: func() string { return evmChain },
			GetPendingCommandsFunc: func(sdk.Context) []types.Command {
				return cmds
			},
		}

		nexusKeeper = &mock.NexusMock{
			GetChainFunc: func(_ sdk.Context, chain string) (nexus.Chain, bool) {
				if strings.ToLower(chain) == strings.ToLower(evmChain) {
					return nexus.Chain{
						Name:                  evmChain,
						SupportsForeignAssets: true,
						Module:                rand.Str(10),
					}, true
				}
				return nexus.Chain{}, false
			},
		}
	}

	repeatCount := 20

	t.Run("happy path", testutils.Func(func(t *testing.T) {
		setup()

		var res types.QueryPendingCommandsResponse
		bz, err := evmKeeper.QueryPendingCommands(ctx, chainKeeper, nexusKeeper)
		assert.NoError(t, err)

		err = res.Unmarshal(bz)
		assert.NoError(t, err)

		var cmdResp []types.QueryCommandResponse
		for _, cmd := range cmds {
			resp, err := evmKeeper.GetCommandResponse(ctx, evmChain, nexusKeeper, cmd)
			assert.NoError(t, err)
			cmdResp = append(cmdResp, resp)
		}

		assert.ElementsMatch(t, cmdResp, res.Commands)

	}).Repeat(repeatCount))
}

func TestQueryTokenAddress(t *testing.T) {

	var (
		chainKeeper     *mock.ChainKeeperMock
		nexusKeeper     *mock.NexusMock
		ctx             sdk.Context
		evmChain        string
		asset           string
		symbol          string
		expectedAddress types.Address
	)

	setup := func() {
		evmChain = rand.StrBetween(5, 10)
		expectedAddress = evmTest.RandomAddress()
		asset = btc.Satoshi
		symbol = "axelarBTC"

		chainKeeper = &mock.ChainKeeperMock{
			GetNameFunc:           func() string { return evmChain },
			GetGatewayAddressFunc: func(sdk.Context) (common.Address, bool) { return common.Address(evmTest.RandomAddress()), true },
			GetERC20TokenBySymbolFunc: func(ctx sdk.Context, s string) types.ERC20Token {
				if symbol == s {
					return createMockConfirmedERC20Token(asset, expectedAddress, createDetails(asset, symbol))
				}
				return types.NilToken
			},
			GetERC20TokenByAssetFunc: func(ctx sdk.Context, a string) types.ERC20Token {
				if asset == a {
					return createMockConfirmedERC20Token(asset, types.Address(expectedAddress), createDetails(asset, symbol))
				}
				return types.NilToken
			},
		}
		nexusKeeper = &mock.NexusMock{
			GetChainFunc: func(_ sdk.Context, chain string) (nexus.Chain, bool) {
				if strings.ToLower(chain) == strings.ToLower(evmChain) {
					return nexus.Chain{
						Name:                  evmChain,
						SupportsForeignAssets: true,
						Module:                rand.Str(10),
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

		var res types.QueryTokenAddressResponse
		bz, err := evmKeeper.QueryTokenAddressByAsset(ctx, chainKeeper, nexusKeeper, asset)
		types.ModuleCdc.UnmarshalLengthPrefixed(bz, &res)

		assert := assert.New(t)
		assert.NoError(err)
		assert.Len(chainKeeper.GetERC20TokenByAssetCalls(), 1)
		assert.Equal(expectedAddress.Hex(), res.Address)

		bz, err = evmKeeper.QueryTokenAddressBySymbol(ctx, chainKeeper, nexusKeeper, symbol)
		types.ModuleCdc.UnmarshalLengthPrefixed(bz, &res)

		assert.NoError(err)
		assert.Len(chainKeeper.GetERC20TokenBySymbolCalls(), 1)
		assert.Equal(expectedAddress.Hex(), res.Address)

	}).Repeat(repeatCount))

	t.Run("token not found", testutils.Func(func(t *testing.T) {
		setup()

		_, err := evmKeeper.QueryTokenAddressByAsset(ctx, chainKeeper, nexusKeeper, rand.Str(10))

		assert := assert.New(t)
		assert.Error(err)

		_, err = evmKeeper.QueryTokenAddressBySymbol(ctx, chainKeeper, nexusKeeper, rand.Str(3))
		assert.Error(err)

	}).Repeat(repeatCount))

	t.Run("token not deployed", testutils.Func(func(t *testing.T) {
		setup()
		chainKeeper.GetERC20TokenBySymbolFunc = func(sdk.Context, string) types.ERC20Token {
			return types.NilToken
		}

		_, err := evmKeeper.QueryTokenAddressBySymbol(ctx, chainKeeper, nexusKeeper, symbol)

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
			BurnerAddress:    evmTest.RandomAddress(),
			TxID:             evmTest.RandomHash(),
			Asset:            rand.StrBetween(5, 10),
		}

		chainKeeper = &mock.ChainKeeperMock{
			GetNameFunc: func() string { return evmChain },
			GetPendingDepositFunc: func(sdk.Context, vote.PollKey) (types.ERC20Deposit, bool) {
				return types.ERC20Deposit{}, false
			},
			GetDepositFunc: func(_ sdk.Context, txID common.Hash, burnerAddr common.Address) (types.ERC20Deposit, types.DepositStatus, bool) {
				return types.ERC20Deposit{}, 0, false
			},
		}
		nexusKeeper = &mock.NexusMock{
			GetChainFunc: func(_ sdk.Context, chain string) (nexus.Chain, bool) {
				if strings.ToLower(chain) == strings.ToLower(evmChain) {
					return nexus.Chain{
						Name:                  evmChain,
						SupportsForeignAssets: true,
						Module:                rand.Str(10),
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
		chainKeeper.GetDepositFunc = func(_ sdk.Context, txID common.Hash, burnerAddr common.Address) (types.ERC20Deposit, types.DepositStatus, bool) {
			if types.Hash(txID) == expectedDeposit.TxID && types.Address(burnerAddr) == expectedDeposit.BurnerAddress {
				return expectedDeposit, types.DepositStatus_Confirmed, true
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
		chainKeeper.GetDepositFunc = func(_ sdk.Context, txID common.Hash, burnerAddr common.Address) (types.ERC20Deposit, types.DepositStatus, bool) {
			if types.Hash(txID) == expectedDeposit.TxID && types.Address(burnerAddr) == expectedDeposit.BurnerAddress {
				return expectedDeposit, types.DepositStatus_Burned, true
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
		data := types.ModuleCdc.MustMarshalJSON(&expectedDeposit)
		nexusKeeper.GetChainFunc = func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
			return nexus.Chain{}, false
		}
		_, err := evmKeeper.QueryDepositState(ctx, chainKeeper, nexusKeeper, data)

		assert := assert.New(t)
		assert.EqualError(err, fmt.Sprintf("%s is not a registered chain: bridge error", evmChain))

	}).Repeat(repeatCount))
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
