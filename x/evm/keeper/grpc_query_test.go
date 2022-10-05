package keeper_test

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	evmKeeper "github.com/axelarnetwork/axelar-core/x/evm/keeper"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/evm/types/mock"
	evmTest "github.com/axelarnetwork/axelar-core/x/evm/types/testutils"
	multisig "github.com/axelarnetwork/axelar-core/x/multisig/exported"
	multisigTestutils "github.com/axelarnetwork/axelar-core/x/multisig/exported/testutils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

func TestQueryPendingCommands(t *testing.T) {
	var (
		chainKeeper    *mock.ChainKeeperMock
		baseKeeper     *mock.BaseKeeperMock
		multisigKeeper *mock.MultisigKeeperMock
		nexusKeeper    *mock.NexusMock
		ctx            sdk.Context
		evmChain       nexus.ChainName
		asset          string
		symbol         string
		chainID        sdk.Int
		keyID          multisig.KeyID
		cmds           []types.Command
	)

	setup := func() {
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())
		evmChain = nexus.ChainName(rand.StrBetween(5, 10))
		asset = rand.Str(5)
		symbol = rand.Str(5)
		chainID = sdk.NewInt(1)
		keyID = multisigTestutils.KeyID()
		dailyMintLimit := sdk.NewUint(uint64(rand.PosI64()))
		cmdDeploy := types.NewDeployTokenCommand(chainID, keyID, asset, createDetails(asset, symbol), types.ZeroAddress, dailyMintLimit)
		cmdMint := types.NewMintTokenCommand(keyID, types.NewCommandID(rand.Bytes(10), chainID), symbol, common.BytesToAddress(rand.Bytes(common.AddressLength)), big.NewInt(rand.I64Between(1000, 100000)))
		cmdBurn := types.NewBurnTokenCommand(chainID, keyID, ctx.BlockHeight(), types.BurnerInfo{
			BurnerAddress: types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength))),
			TokenAddress:  types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength))),
			Symbol:        symbol,
			Salt:          types.Hash(common.BytesToHash(rand.Bytes(common.HashLength))),
		}, false)
		cmds = append(cmds, cmdDeploy, cmdMint, cmdBurn)

		chainKeeper = &mock.ChainKeeperMock{
			GetNameFunc: func() string { return evmChain.String() },
			GetPendingCommandsFunc: func(sdk.Context) []types.Command {
				return cmds
			},
		}

		nexusKeeper = &mock.NexusMock{
			GetChainFunc: func(_ sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
				if chain.Equals(evmChain) {
					return nexus.Chain{
						Name:                  evmChain,
						SupportsForeignAssets: true,
						Module:                rand.Str(10),
					}, true
				}
				return nexus.Chain{}, false
			},
		}

		baseKeeper = &mock.BaseKeeperMock{
			ForChainFunc: func(chain nexus.ChainName) types.ChainKeeper {
				return chainKeeper
			},
		}
	}

	repeatCount := 20

	t.Run("happy path", testutils.Func(func(t *testing.T) {
		setup()

		q := evmKeeper.NewGRPCQuerier(baseKeeper, nexusKeeper, multisigKeeper)

		res, err := q.PendingCommands(sdk.WrapSDKContext(ctx), &types.PendingCommandsRequest{Chain: evmChain.String()})
		assert.NoError(t, err)

		var cmdResp []types.QueryCommandResponse
		for _, cmd := range cmds {
			resp, err := evmKeeper.GetCommandResponse(cmd)
			assert.NoError(t, err)
			cmdResp = append(cmdResp, resp)
		}

		assert.ElementsMatch(t, cmdResp, res.Commands)

	}).Repeat(repeatCount))
}

func TestQueryDepositState(t *testing.T) {
	var (
		baseKeeper      *mock.BaseKeeperMock
		multisig        *mock.MultisigKeeperMock
		ctx             sdk.Context
		evmChain        nexus.ChainName
		expectedDeposit types.ERC20Deposit
		chainKeeper     *mock.ChainKeeperMock
		nexusKeeper     *mock.NexusMock
		grpcQuerier     *evmKeeper.Querier
	)

	setup := func() {
		evmChain = nexus.ChainName(rand.StrBetween(5, 10))
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())

		expectedDeposit = types.ERC20Deposit{
			DestinationChain: nexus.ChainName(rand.StrBetween(5, 10)),
			Amount:           sdk.NewUint(uint64(rand.I64Between(100, 10000))),
			BurnerAddress:    evmTest.RandomAddress(),
			TxID:             evmTest.RandomHash(),
			Asset:            rand.StrBetween(5, 10),
		}

		chainKeeper = &mock.ChainKeeperMock{
			GetNameFunc: func() string { return evmChain.String() },
			GetDepositFunc: func(_ sdk.Context, txID types.Hash, burnerAddr types.Address) (types.ERC20Deposit, types.DepositStatus, bool) {
				return types.ERC20Deposit{}, 0, false
			},
		}
		nexusKeeper = &mock.NexusMock{
			GetChainFunc: func(_ sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
				if chain.Equals(evmChain) {
					return nexus.Chain{
						Name:                  evmChain,
						SupportsForeignAssets: true,
						Module:                rand.Str(10),
					}, true
				}
				return nexus.Chain{}, false
			},
		}
		baseKeeper = &mock.BaseKeeperMock{
			ForChainFunc: func(chain nexus.ChainName) types.ChainKeeper {
				return chainKeeper
			},
		}

		q := evmKeeper.NewGRPCQuerier(baseKeeper, nexusKeeper, multisig)
		grpcQuerier = &q
	}
	repeatCount := 20
	t.Run("no deposit", testutils.Func(func(t *testing.T) {
		setup()
		res, err := grpcQuerier.DepositState(sdk.WrapSDKContext(ctx), &types.DepositStateRequest{
			Chain: evmChain,
			Params: &types.QueryDepositStateParams{
				TxID:          expectedDeposit.TxID,
				BurnerAddress: expectedDeposit.BurnerAddress,
			},
		})

		assert := assert.New(t)
		assert.NoError(err)
		assert.Len(chainKeeper.GetNameCalls(), 1)
		assert.Len(chainKeeper.GetDepositCalls(), 1)
		assert.Len(nexusKeeper.GetChainCalls(), 1)

		assert.Equal(types.DepositStatus_None, res.Status)
	}).Repeat(repeatCount))

	t.Run("deposit confirmed", testutils.Func(func(t *testing.T) {
		setup()
		chainKeeper.GetDepositFunc = func(_ sdk.Context, txID types.Hash, burnerAddr types.Address) (types.ERC20Deposit, types.DepositStatus, bool) {
			if txID == expectedDeposit.TxID && burnerAddr == expectedDeposit.BurnerAddress {
				return expectedDeposit, types.DepositStatus_Confirmed, true
			}
			return types.ERC20Deposit{}, 0, false
		}

		res, err := grpcQuerier.DepositState(sdk.WrapSDKContext(ctx), &types.DepositStateRequest{
			Chain: evmChain,
			Params: &types.QueryDepositStateParams{
				TxID:          expectedDeposit.TxID,
				BurnerAddress: expectedDeposit.BurnerAddress,
			},
		})

		assert := assert.New(t)
		assert.NoError(err)
		assert.Len(chainKeeper.GetNameCalls(), 1)
		assert.Len(chainKeeper.GetDepositCalls(), 1)
		assert.Len(nexusKeeper.GetChainCalls(), 1)

		assert.Equal(types.DepositStatus_Confirmed, res.Status)

	}).Repeat(repeatCount))

	t.Run("deposit burned", testutils.Func(func(t *testing.T) {
		setup()
		chainKeeper.GetDepositFunc = func(_ sdk.Context, txID types.Hash, burnerAddr types.Address) (types.ERC20Deposit, types.DepositStatus, bool) {
			if txID == expectedDeposit.TxID && burnerAddr == expectedDeposit.BurnerAddress {
				return expectedDeposit, types.DepositStatus_Burned, true
			}
			return types.ERC20Deposit{}, 0, false
		}

		res, err := grpcQuerier.DepositState(sdk.WrapSDKContext(ctx), &types.DepositStateRequest{
			Chain: evmChain,
			Params: &types.QueryDepositStateParams{
				TxID:          expectedDeposit.TxID,
				BurnerAddress: expectedDeposit.BurnerAddress,
			},
		})

		assert := assert.New(t)
		assert.NoError(err)
		assert.Len(chainKeeper.GetNameCalls(), 1)
		assert.Len(chainKeeper.GetDepositCalls(), 1)
		assert.Len(nexusKeeper.GetChainCalls(), 1)

		assert.Equal(types.DepositStatus_Burned, res.Status)

	}).Repeat(repeatCount))

	t.Run("chain not registered", testutils.Func(func(t *testing.T) {
		setup()
		nexusKeeper.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
			return nexus.Chain{}, false
		}
		_, err := grpcQuerier.DepositState(sdk.WrapSDKContext(ctx), &types.DepositStateRequest{
			Chain: evmChain,
			Params: &types.QueryDepositStateParams{
				TxID:          expectedDeposit.TxID,
				BurnerAddress: expectedDeposit.BurnerAddress,
			},
		})

		assert := assert.New(t)
		assert.EqualError(err, fmt.Sprintf("rpc error: code = NotFound desc = %s is not a registered chain", evmChain))

	}).Repeat(repeatCount))
}

func TestChains(t *testing.T) {
	var (
		baseKeeper  *mock.BaseKeeperMock
		multisig    *mock.MultisigKeeperMock
		nexusKeeper *mock.NexusMock
		ctx         sdk.Context
		evmChain    nexus.ChainName
		nonEvmChain nexus.ChainName
		expectedRes types.ChainsResponse
		grpcQuerier *evmKeeper.Querier
	)

	setup := func() {
		evmChain = "evm-chain"
		nonEvmChain = "non-evm-chain"
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())
	}

	repeatCount := 1

	t.Run("evm chain exists", testutils.Func(func(t *testing.T) {
		setup()

		expectedRes = types.ChainsResponse{Chains: []nexus.ChainName{evmChain}}
		nexusKeeper = &mock.NexusMock{
			GetChainsFunc: func(ctx sdk.Context) []nexus.Chain {
				return []nexus.Chain{
					{
						Name:                  evmChain,
						SupportsForeignAssets: true,
						Module:                types.ModuleName,
					},
					{
						Name:                  nonEvmChain,
						SupportsForeignAssets: true,
						Module:                "non-evm",
					}}
			},
		}

		q := evmKeeper.NewGRPCQuerier(baseKeeper, nexusKeeper, multisig)
		grpcQuerier = &q
		res, err := grpcQuerier.Chains(sdk.WrapSDKContext(ctx), &types.ChainsRequest{})

		assert := assert.New(t)
		assert.NoError(err)

		assert.Equal(expectedRes, *res)
	}).Repeat(repeatCount))

	t.Run("evm chain doesn't exist", testutils.Func(func(t *testing.T) {
		setup()

		expectedRes = types.ChainsResponse{Chains: []nexus.ChainName{}}
		nexusKeeper = &mock.NexusMock{
			GetChainsFunc: func(ctx sdk.Context) []nexus.Chain {
				return []nexus.Chain{
					{
						Name:                  nonEvmChain,
						SupportsForeignAssets: true,
						Module:                "non-evm",
					}}
			},
		}

		q := evmKeeper.NewGRPCQuerier(baseKeeper, nexusKeeper, multisig)
		grpcQuerier = &q
		res, err := grpcQuerier.Chains(sdk.WrapSDKContext(ctx), &types.ChainsRequest{})

		assert := assert.New(t)
		assert.NoError(err)

		assert.Equal(expectedRes, *res)
	}).Repeat(repeatCount))
}

func TestGateway(t *testing.T) {
	var (
		baseKeeper    *mock.BaseKeeperMock
		multisig      *mock.MultisigKeeperMock
		nexusKeeper   *mock.NexusMock
		chainKeeper   *mock.ChainKeeperMock
		ctx           sdk.Context
		expectedRes   types.GatewayAddressResponse
		grpcQuerier   *evmKeeper.Querier
		address       types.Address
		existingChain nexus.ChainName
	)

	setup := func() {
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())
		address = types.Address(common.BytesToAddress([]byte{0}))

		chainKeeper = &mock.ChainKeeperMock{
			GetGatewayAddressFunc: func(ctx sdk.Context) (types.Address, bool) {
				return address, true
			},
		}

		existingChain = "existing"
		baseKeeper = &mock.BaseKeeperMock{
			HasChainFunc: func(ctx sdk.Context, chain nexus.ChainName) bool {
				return chain == existingChain
			},
			ForChainFunc: func(chain nexus.ChainName) types.ChainKeeper {
				return chainKeeper
			},
		}

		q := evmKeeper.NewGRPCQuerier(baseKeeper, nexusKeeper, multisig)
		grpcQuerier = &q
	}

	repeatCount := 1

	t.Run("gateway exists", testutils.Func(func(t *testing.T) {
		setup()

		expectedRes = types.GatewayAddressResponse{
			Address: address.Hex(),
		}

		res, err := grpcQuerier.GatewayAddress(sdk.WrapSDKContext(ctx), &types.GatewayAddressRequest{
			Chain: existingChain.String(),
		})

		assert := assert.New(t)
		assert.NoError(err)

		assert.Equal(expectedRes, *res)
	}).Repeat(repeatCount))

	t.Run("chain does not exist", testutils.Func(func(t *testing.T) {
		setup()

		_, err := grpcQuerier.GatewayAddress(sdk.WrapSDKContext(ctx), &types.GatewayAddressRequest{
			Chain: "non-existing-chain",
		})

		assert := assert.New(t)
		assert.Error(err)
	}).Repeat(repeatCount))

	t.Run("gateway does not exist", testutils.Func(func(t *testing.T) {
		setup()

		chainKeeper = &mock.ChainKeeperMock{
			GetGatewayAddressFunc: func(ctx sdk.Context) (types.Address, bool) {
				return address, false
			},
		}

		_, err := grpcQuerier.GatewayAddress(sdk.WrapSDKContext(ctx), &types.GatewayAddressRequest{
			Chain: existingChain.String(),
		})

		assert := assert.New(t)
		assert.Error(err)
	}).Repeat(repeatCount))
}

func TestBytecode(t *testing.T) {
	var (
		baseKeeper     *mock.BaseKeeperMock
		multisig       *mock.MultisigKeeperMock
		nexusKeeper    *mock.NexusMock
		chainKeeper    *mock.ChainKeeperMock
		ctx            sdk.Context
		expectedRes    types.BytecodeResponse
		grpcQuerier    *evmKeeper.Querier
		existingChain  nexus.ChainName
		contracts      []string
		bytecodesExist bool
	)

	setup := func() {
		existingChain = "existing"
		contracts = []string{"token", "burner"}

		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())

		nexusKeeper = &mock.NexusMock{
			GetChainFunc: func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
				if chain == existingChain {
					return nexus.Chain{
						Name:                  chain,
						SupportsForeignAssets: false,
						KeyType:               0,
						Module:                "evm",
					}, true
				}
				return nexus.Chain{}, false
			},
		}

		chainKeeper = &mock.ChainKeeperMock{
			GetTokenByteCodeFunc: func(ctx sdk.Context) ([]byte, bool) {
				if bytecodesExist {
					return []byte(contracts[0]), true
				}
				return nil, false
			},
			GetBurnerByteCodeFunc: func(ctx sdk.Context) ([]byte, bool) {
				if bytecodesExist {
					return []byte(contracts[1]), true
				}
				return nil, false
			},
		}

		baseKeeper = &mock.BaseKeeperMock{
			ForChainFunc: func(chain nexus.ChainName) types.ChainKeeper {
				return chainKeeper
			},
		}

		q := evmKeeper.NewGRPCQuerier(baseKeeper, nexusKeeper, multisig)
		grpcQuerier = &q
	}

	repeatCount := 1

	t.Run("bytecode exists", testutils.Func(func(t *testing.T) {
		setup()
		for _, bytecode := range contracts {
			hexBytecode := fmt.Sprintf("0x" + common.Bytes2Hex([]byte(bytecode)))
			expectedRes = types.BytecodeResponse{
				Bytecode: hexBytecode,
			}

			bytecodesExist = true

			res, err := grpcQuerier.Bytecode(sdk.WrapSDKContext(ctx), &types.BytecodeRequest{
				Chain:    existingChain.String(),
				Contract: bytecode,
			})

			assert := assert.New(t)
			assert.NoError(err)

			assert.Equal(expectedRes, *res)
		}
	}).Repeat(repeatCount))
}

func TestEvent(t *testing.T) {
	var (
		baseKeeper         *mock.BaseKeeperMock
		multisig           *mock.MultisigKeeperMock
		chainKeeper        *mock.ChainKeeperMock
		nexusKeeper        *mock.NexusMock
		ctx                sdk.Context
		expectedResp       types.EventResponse
		grpcQuerier        *evmKeeper.Querier
		existingChain      nexus.ChainName
		nonExistingChain   nexus.ChainName
		existingTxID       string
		existingEventID    string
		nonExistingEventID string
		existingStatus     types.Event_Status
	)

	setup := func() {
		existingChain = "existing-chain"
		nonExistingChain = "non-existing-chain"
		existingTxID = evmTest.RandomHash().Hex()
		existingEventID = fmt.Sprintf("%s-%d", existingTxID, rand.PosI64())
		nonExistingEventID = fmt.Sprintf("%s-%d", evmTest.RandomHash().Hex(), rand.PosI64())

		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())

		chainKeeper = &mock.ChainKeeperMock{
			GetEventFunc: func(ctx sdk.Context, eventID types.EventID) (types.Event, bool) {
				if eventID == types.EventID(existingEventID) {
					return types.Event{
						Chain:  existingChain,
						TxID:   types.Hash(common.HexToHash(existingTxID)),
						Index:  0,
						Status: existingStatus,
						Event:  nil,
					}, true
				}
				return types.Event{}, false
			},
		}

		baseKeeper = &mock.BaseKeeperMock{
			HasChainFunc: func(_ sdk.Context, chain nexus.ChainName) bool {
				return chain == existingChain
			},
			ForChainFunc: func(chain nexus.ChainName) types.ChainKeeper {
				return chainKeeper
			},
		}

		q := evmKeeper.NewGRPCQuerier(baseKeeper, nexusKeeper, multisig)
		grpcQuerier = &q
	}

	repeatCount := 10

	statuses := []types.Event_Status{types.EventCompleted, types.EventConfirmed, types.EventNonExistent}

	t.Run("chain and event exist", testutils.Func(func(t *testing.T) {
		setup()
		for _, status := range statuses {
			existingStatus = status
			expectedResp = types.EventResponse{
				Event: &types.Event{
					Chain:  existingChain,
					TxID:   types.Hash(common.HexToHash(existingTxID)),
					Index:  0,
					Status: existingStatus,
					Event:  nil,
				},
			}

			res, err := grpcQuerier.Event(sdk.WrapSDKContext(ctx), &types.EventRequest{
				Chain:   existingChain.String(),
				EventId: existingEventID,
			})

			assert := assert.New(t)
			assert.NoError(err)

			assert.Equal(expectedResp, *res)
		}
	}).Repeat(repeatCount))

	t.Run("chain doesn't exist", testutils.Func(func(t *testing.T) {
		setup()
		_, err := grpcQuerier.Event(sdk.WrapSDKContext(ctx), &types.EventRequest{
			Chain:   nonExistingChain.String(),
			EventId: existingEventID,
		})

		assert := assert.New(t)
		assert.Error(err)

		assert.Equal(err.Error(), fmt.Sprintf("rpc error: code = NotFound desc = [%s] is not a registered chain: bridge error", nonExistingChain))
	}).Repeat(repeatCount))

	t.Run("event doesn't exist", testutils.Func(func(t *testing.T) {
		setup()
		_, err := grpcQuerier.Event(sdk.WrapSDKContext(ctx), &types.EventRequest{
			Chain:   existingChain.String(),
			EventId: nonExistingEventID,
		})

		assert := assert.New(t)
		assert.Error(err)

		assert.Equal(err.Error(), fmt.Sprintf("rpc error: code = NotFound desc = no event with ID [%s] was found: bridge error", nonExistingEventID))
	}).Repeat(repeatCount))
}

func TestERC20Tokens(t *testing.T) {
	var (
		baseKeeper    *mock.BaseKeeperMock
		nexusKeeper   *mock.NexusMock
		chainKeeper   *mock.ChainKeeperMock
		ctx           sdk.Context
		existingChain nexus.ChainName
		expectedRes   types.ERC20TokensResponse
		grpcQuerier   *evmKeeper.Querier
	)

	external := types.CreateERC20Token(func(meta types.ERC20TokenMetadata) {}, types.ERC20TokenMetadata{Asset: "external", IsExternal: true})
	internal := types.CreateERC20Token(func(meta types.ERC20TokenMetadata) {}, types.ERC20TokenMetadata{Asset: "internal", IsExternal: false})

	setup := func() {
		existingChain = nexus.ChainName("chain")

		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())

		nexusKeeper = &mock.NexusMock{
			GetChainFunc: func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
				if chain == existingChain {
					return nexus.Chain{
						Name:                  chain,
						SupportsForeignAssets: true,
						Module:                types.ModuleName,
					}, true
				}
				return nexus.Chain{}, false
			},
		}
		chainKeeper = &mock.ChainKeeperMock{
			GetTokensFunc: func(ctx sdk.Context) []types.ERC20Token {
				return []types.ERC20Token{external, internal}
			},
		}
		baseKeeper = &mock.BaseKeeperMock{
			ForChainFunc: func(chain nexus.ChainName) types.ChainKeeper {
				return chainKeeper
			},
		}

		q := evmKeeper.NewGRPCQuerier(baseKeeper, nexusKeeper, nil)
		grpcQuerier = &q
	}

	repeatCount := 1

	t.Run("all erc20 tokens", testutils.Func(func(t *testing.T) {
		setup()

		expectedTokens := []types.ERC20TokensResponse_Token{
			{
				Asset:  external.GetAsset(),
				Symbol: external.GetDetails().Symbol,
			},
			{
				Asset:  internal.GetAsset(),
				Symbol: internal.GetDetails().Symbol,
			},
		}
		expectedRes = types.ERC20TokensResponse{Tokens: expectedTokens}

		res, err := grpcQuerier.ERC20Tokens(sdk.WrapSDKContext(ctx), &types.ERC20TokensRequest{Chain: existingChain.String()})
		assert := assert.New(t)
		assert.NoError(err)

		assert.Equal(expectedRes, *res)
	}).Repeat(repeatCount))

	t.Run("internal erc20 tokens only", testutils.Func(func(t *testing.T) {
		setup()

		expectedTokens := []types.ERC20TokensResponse_Token{{
			Asset:  internal.GetAsset(),
			Symbol: internal.GetDetails().Symbol,
		}}
		expectedRes = types.ERC20TokensResponse{Tokens: expectedTokens}

		res, err := grpcQuerier.ERC20Tokens(sdk.WrapSDKContext(ctx), &types.ERC20TokensRequest{Chain: existingChain.String(), Type: types.Internal})
		assert := assert.New(t)
		assert.NoError(err)
		assert.Equal(expectedRes, *res)
	}).Repeat(repeatCount))

	t.Run("external erc20 tokens only", testutils.Func(func(t *testing.T) {
		setup()

		expectedTokens := []types.ERC20TokensResponse_Token{{
			Asset:  external.GetAsset(),
			Symbol: external.GetDetails().Symbol,
		}}
		expectedRes = types.ERC20TokensResponse{Tokens: expectedTokens}

		res, err := grpcQuerier.ERC20Tokens(sdk.WrapSDKContext(ctx), &types.ERC20TokensRequest{Chain: existingChain.String(), Type: types.External})
		assert := assert.New(t)
		assert.NoError(err)
		assert.Equal(expectedRes, *res)
	}).Repeat(repeatCount))

	t.Run("non evm chain", testutils.Func(func(t *testing.T) {
		setup()

		res, err := grpcQuerier.ERC20Tokens(sdk.WrapSDKContext(ctx), &types.ERC20TokensRequest{Chain: "non-existing-chain", Type: types.Unspecified})
		assert := assert.New(t)
		assert.Error(err)
		assert.Nil(res)
	}).Repeat(repeatCount))
}

func TestTokenInfo(t *testing.T) {
	var (
		baseKeeper    *mock.BaseKeeperMock
		multisig      *mock.MultisigKeeperMock
		nexusKeeper   *mock.NexusMock
		chainKeeper   *mock.ChainKeeperMock
		existingChain nexus.ChainName
		ctx           sdk.Context
		grpcQuerier   *evmKeeper.Querier
	)

	burnerCode, err := hex.DecodeString(rand.HexStr(200))
	if err != nil {
		panic(err)
	}
	burnerCodeHash := types.Hash(crypto.Keccak256Hash(burnerCode)).Hex()
	token := types.CreateERC20Token(func(meta types.ERC20TokenMetadata) {}, types.ERC20TokenMetadata{
		Asset:        "token",
		Details:      types.NewTokenDetails("Token", "TOKEN", 10, sdk.NewInt(0)),
		TokenAddress: types.ZeroAddress,
		Status:       types.Confirmed,
		IsExternal:   true,
		BurnerCode:   burnerCode,
	})

	setup := func() {
		existingChain = nexus.ChainName("chain")

		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())

		nexusKeeper = &mock.NexusMock{
			GetChainFunc: func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
				if chain == existingChain {
					return nexus.Chain{
						Name:                  existingChain,
						SupportsForeignAssets: true,
						Module:                types.ModuleName,
					}, true
				}
				return nexus.Chain{}, false
			},
		}
		chainKeeper = &mock.ChainKeeperMock{
			GetERC20TokenByAssetFunc: func(ctx sdk.Context, asset string) types.ERC20Token {
				if asset == token.GetAsset() {
					return token
				}

				return types.NilToken
			},
			GetERC20TokenBySymbolFunc: func(ctx sdk.Context, symbol string) types.ERC20Token {
				if symbol == token.GetDetails().Symbol {
					return token
				}

				return types.NilToken
			},
		}
		baseKeeper = &mock.BaseKeeperMock{
			ForChainFunc: func(chain nexus.ChainName) types.ChainKeeper {
				return chainKeeper
			},
		}

		q := evmKeeper.NewGRPCQuerier(baseKeeper, nexusKeeper, multisig)
		grpcQuerier = &q
	}

	repeatCount := 1
	expectedRes := types.TokenInfoResponse{
		Asset:          token.GetAsset(),
		Details:        token.GetDetails(),
		Address:        token.GetAddress().Hex(),
		Confirmed:      token.Is(types.Confirmed),
		IsExternal:     token.IsExternal(),
		BurnerCodeHash: burnerCodeHash,
	}

	t.Run("token detail by asset", testutils.Func(func(t *testing.T) {
		setup()

		res, err := grpcQuerier.TokenInfo(sdk.WrapSDKContext(ctx), &types.TokenInfoRequest{
			Chain:  existingChain.String(),
			FindBy: &types.TokenInfoRequest_Asset{Asset: token.GetAsset()},
		})
		assert := assert.New(t)
		assert.NoError(err)
		assert.Equal(expectedRes, *res)
	}).Repeat(repeatCount))

	t.Run("token detail by symbol", testutils.Func(func(t *testing.T) {
		setup()

		res, err := grpcQuerier.TokenInfo(sdk.WrapSDKContext(ctx), &types.TokenInfoRequest{
			Chain:  existingChain.String(),
			FindBy: &types.TokenInfoRequest_Symbol{Symbol: token.GetDetails().Symbol},
		})
		assert := assert.New(t)
		assert.NoError(err)
		assert.Equal(expectedRes, *res)
	}).Repeat(repeatCount))

	t.Run("unknown token by asset", testutils.Func(func(t *testing.T) {
		setup()

		res, err := grpcQuerier.TokenInfo(sdk.WrapSDKContext(ctx), &types.TokenInfoRequest{
			Chain:  existingChain.String(),
			FindBy: &types.TokenInfoRequest_Asset{Asset: "unknown-token"},
		})
		assert := assert.New(t)
		assert.Nil(res)
		assert.Error(err)
	}).Repeat(repeatCount))

	t.Run("unknown token by symbol", testutils.Func(func(t *testing.T) {
		setup()

		res, err := grpcQuerier.TokenInfo(sdk.WrapSDKContext(ctx), &types.TokenInfoRequest{
			Chain:  existingChain.String(),
			FindBy: &types.TokenInfoRequest_Symbol{Symbol: "UTOKEN"},
		})
		assert := assert.New(t)
		assert.Nil(res)
		assert.Error(err)
	}).Repeat(repeatCount))
}
