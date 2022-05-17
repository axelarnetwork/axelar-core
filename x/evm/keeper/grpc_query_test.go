package keeper_test

import (
	"fmt"
	"math/big"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	btc "github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	evmKeeper "github.com/axelarnetwork/axelar-core/x/evm/keeper"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/evm/types/mock"
	evmTest "github.com/axelarnetwork/axelar-core/x/evm/types/testutils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	tssTestUtils "github.com/axelarnetwork/axelar-core/x/tss/exported/testutils"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

func TestQueryPendingCommands(t *testing.T) {
	var (
		chainKeeper *mock.ChainKeeperMock
		baseKeeper  *mock.BaseKeeperMock
		signer      *mock.SignerMock
		nexusKeeper *mock.NexusMock
		ctx         sdk.Context
		evmChain    string
		asset       string
		symbol      string
		chainID     sdk.Int
		keyID       tss.KeyID
		cmds        []types.Command
	)

	setup := func() {
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())
		evmChain = rand.StrBetween(5, 10)
		asset = btc.NativeAsset
		symbol = "axelarBTC"
		chainID = sdk.NewInt(1)
		keyID = tssTestUtils.RandKeyID()
		cmdDeploy, _ := types.CreateDeployTokenCommand(chainID, keyID, asset, createDetails(asset, symbol), types.ZeroAddress)
		cmdMint, _ := types.CreateMintTokenCommand(keyID, types.NewCommandID(rand.Bytes(10), chainID), symbol, common.BytesToAddress(rand.Bytes(common.AddressLength)), big.NewInt(rand.I64Between(1000, 100000)))
		cmdBurn, _ := types.CreateBurnTokenCommand(chainID, keyID, ctx.BlockHeight(), types.BurnerInfo{
			BurnerAddress: types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength))),
			TokenAddress:  types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength))),
			Symbol:        symbol,
			Salt:          types.Hash(common.BytesToHash(rand.Bytes(common.HashLength))),
		}, false)
		cmds = append(cmds, cmdDeploy, cmdMint, cmdBurn)

		chainKeeper = &mock.ChainKeeperMock{
			GetNameFunc: func() string { return evmChain },
			GetPendingCommandsFunc: func(sdk.Context) []types.Command {
				return cmds
			},
		}

		nexusKeeper = &mock.NexusMock{
			GetChainFunc: func(_ sdk.Context, chain string) (nexus.Chain, bool) {
				if strings.EqualFold(chain, evmChain) {
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
			ForChainFunc: func(chain string) types.ChainKeeper {
				return chainKeeper
			},
		}
	}

	repeatCount := 20

	t.Run("happy path", testutils.Func(func(t *testing.T) {
		setup()

		q := evmKeeper.NewGRPCQuerier(baseKeeper, nexusKeeper, signer)

		res, err := q.PendingCommands(sdk.WrapSDKContext(ctx), &types.PendingCommandsRequest{Chain: evmChain})
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
		signer          *mock.SignerMock
		ctx             sdk.Context
		evmChain        string
		expectedDeposit types.ERC20Deposit
		chainKeeper     *mock.ChainKeeperMock
		nexusKeeper     *mock.NexusMock
		grpcQuerier     *evmKeeper.Querier
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
				if strings.EqualFold(chain, evmChain) {
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
			ForChainFunc: func(chain string) types.ChainKeeper {
				return chainKeeper
			},
		}

		q := evmKeeper.NewGRPCQuerier(baseKeeper, nexusKeeper, signer)
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
				Amount:        expectedDeposit.Amount.String(),
			},
		})

		assert := assert.New(t)
		assert.NoError(err)
		assert.Len(chainKeeper.GetNameCalls(), 1)
		assert.Len(chainKeeper.GetPendingDepositCalls(), 1)
		assert.Len(chainKeeper.GetDepositCalls(), 1)
		assert.Len(nexusKeeper.GetChainCalls(), 1)

		assert.Equal(types.DepositStatus_None, res.Status)
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

		res, err := grpcQuerier.DepositState(sdk.WrapSDKContext(ctx), &types.DepositStateRequest{
			Chain: evmChain,
			Params: &types.QueryDepositStateParams{
				TxID:          expectedDeposit.TxID,
				BurnerAddress: expectedDeposit.BurnerAddress,
				Amount:        expectedDeposit.Amount.String(),
			},
		})

		assert := assert.New(t)
		assert.NoError(err)
		assert.Len(chainKeeper.GetNameCalls(), 1)
		assert.Len(chainKeeper.GetPendingDepositCalls(), 1)
		assert.Len(chainKeeper.GetDepositCalls(), 1)
		assert.Len(nexusKeeper.GetChainCalls(), 1)

		assert.Equal(types.DepositStatus_Pending, res.Status)

	}).Repeat(repeatCount))

	t.Run("deposit confirmed", testutils.Func(func(t *testing.T) {
		setup()
		chainKeeper.GetDepositFunc = func(_ sdk.Context, txID common.Hash, burnerAddr common.Address) (types.ERC20Deposit, types.DepositStatus, bool) {
			if types.Hash(txID) == expectedDeposit.TxID && types.Address(burnerAddr) == expectedDeposit.BurnerAddress {
				return expectedDeposit, types.DepositStatus_Confirmed, true
			}
			return types.ERC20Deposit{}, 0, false
		}

		res, err := grpcQuerier.DepositState(sdk.WrapSDKContext(ctx), &types.DepositStateRequest{
			Chain: evmChain,
			Params: &types.QueryDepositStateParams{
				TxID:          expectedDeposit.TxID,
				BurnerAddress: expectedDeposit.BurnerAddress,
				Amount:        expectedDeposit.Amount.String(),
			},
		})

		assert := assert.New(t)
		assert.NoError(err)
		assert.Len(chainKeeper.GetNameCalls(), 1)
		assert.Len(chainKeeper.GetPendingDepositCalls(), 1)
		assert.Len(chainKeeper.GetDepositCalls(), 1)
		assert.Len(nexusKeeper.GetChainCalls(), 1)

		assert.Equal(types.DepositStatus_Confirmed, res.Status)

	}).Repeat(repeatCount))

	t.Run("deposit burned", testutils.Func(func(t *testing.T) {
		setup()
		chainKeeper.GetDepositFunc = func(_ sdk.Context, txID common.Hash, burnerAddr common.Address) (types.ERC20Deposit, types.DepositStatus, bool) {
			if types.Hash(txID) == expectedDeposit.TxID && types.Address(burnerAddr) == expectedDeposit.BurnerAddress {
				return expectedDeposit, types.DepositStatus_Burned, true
			}
			return types.ERC20Deposit{}, 0, false
		}

		res, err := grpcQuerier.DepositState(sdk.WrapSDKContext(ctx), &types.DepositStateRequest{
			Chain: evmChain,
			Params: &types.QueryDepositStateParams{
				TxID:          expectedDeposit.TxID,
				BurnerAddress: expectedDeposit.BurnerAddress,
				Amount:        expectedDeposit.Amount.String(),
			},
		})

		assert := assert.New(t)
		assert.NoError(err)
		assert.Len(chainKeeper.GetNameCalls(), 1)
		assert.Len(chainKeeper.GetPendingDepositCalls(), 1)
		assert.Len(chainKeeper.GetDepositCalls(), 1)
		assert.Len(nexusKeeper.GetChainCalls(), 1)

		assert.Equal(types.DepositStatus_Burned, res.Status)

	}).Repeat(repeatCount))

	t.Run("chain not registered", testutils.Func(func(t *testing.T) {
		setup()
		nexusKeeper.GetChainFunc = func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
			return nexus.Chain{}, false
		}
		_, err := grpcQuerier.DepositState(sdk.WrapSDKContext(ctx), &types.DepositStateRequest{
			Chain: evmChain,
			Params: &types.QueryDepositStateParams{
				TxID:          expectedDeposit.TxID,
				BurnerAddress: expectedDeposit.BurnerAddress,
				Amount:        expectedDeposit.Amount.String(),
			},
		})

		assert := assert.New(t)
		assert.EqualError(err, fmt.Sprintf("rpc error: code = NotFound desc = %s is not a registered chain", evmChain))

	}).Repeat(repeatCount))
}

func TestChains(t *testing.T) {
	var (
		baseKeeper  *mock.BaseKeeperMock
		signer      *mock.SignerMock
		nexusKeeper *mock.NexusMock
		ctx         sdk.Context
		evmChain    string
		nonEvmChain string
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

		expectedRes = types.ChainsResponse{Chains: []string{evmChain}}
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

		q := evmKeeper.NewGRPCQuerier(baseKeeper, nexusKeeper, signer)
		grpcQuerier = &q
		res, err := grpcQuerier.Chains(sdk.WrapSDKContext(ctx), &types.ChainsRequest{})

		assert := assert.New(t)
		assert.NoError(err)

		assert.Equal(expectedRes, *res)
	}).Repeat(repeatCount))

	t.Run("evm chain doesn't exist", testutils.Func(func(t *testing.T) {
		setup()

		expectedRes = types.ChainsResponse{Chains: []string{}}
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

		q := evmKeeper.NewGRPCQuerier(baseKeeper, nexusKeeper, signer)
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
		signer        *mock.SignerMock
		nexusKeeper   *mock.NexusMock
		chainKeeper   *mock.ChainKeeperMock
		ctx           sdk.Context
		expectedRes   types.GatewayAddressResponse
		grpcQuerier   *evmKeeper.Querier
		address       common.Address
		existingChain string
	)

	setup := func() {
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())
		address = common.BytesToAddress([]byte{0})

		chainKeeper = &mock.ChainKeeperMock{
			GetGatewayAddressFunc: func(ctx sdk.Context) (common.Address, bool) {
				return address, true
			},
		}

		existingChain = "existing"
		baseKeeper = &mock.BaseKeeperMock{
			HasChainFunc: func(ctx sdk.Context, chain string) bool {
				return chain == existingChain
			},
			ForChainFunc: func(chain string) types.ChainKeeper {
				return chainKeeper
			},
		}

		q := evmKeeper.NewGRPCQuerier(baseKeeper, nexusKeeper, signer)
		grpcQuerier = &q
	}

	repeatCount := 1

	t.Run("gateway exists", testutils.Func(func(t *testing.T) {
		setup()

		expectedRes = types.GatewayAddressResponse{
			Address: address.Hex(),
		}

		res, err := grpcQuerier.GatewayAddress(sdk.WrapSDKContext(ctx), &types.GatewayAddressRequest{
			Chain: existingChain,
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
			GetGatewayAddressFunc: func(ctx sdk.Context) (common.Address, bool) {
				return address, false
			},
		}

		_, err := grpcQuerier.GatewayAddress(sdk.WrapSDKContext(ctx), &types.GatewayAddressRequest{
			Chain: existingChain,
		})

		assert := assert.New(t)
		assert.Error(err)
	}).Repeat(repeatCount))
}

func TestBytecode(t *testing.T) {
	var (
		baseKeeper     *mock.BaseKeeperMock
		signer         *mock.SignerMock
		nexusKeeper    *mock.NexusMock
		chainKeeper    *mock.ChainKeeperMock
		ctx            sdk.Context
		expectedRes    types.BytecodeResponse
		grpcQuerier    *evmKeeper.Querier
		existingChain  string
		contracts      []string
		bytecodesExist bool
	)

	setup := func() {
		existingChain = "existing"
		contracts = []string{"token", "burner"}

		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())

		nexusKeeper = &mock.NexusMock{
			GetChainFunc: func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
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
			ForChainFunc: func(chain string) types.ChainKeeper {
				return chainKeeper
			},
		}

		q := evmKeeper.NewGRPCQuerier(baseKeeper, nexusKeeper, signer)
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
				Chain:    existingChain,
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
		signer             *mock.SignerMock
		chainKeeper        *mock.ChainKeeperMock
		nexusKeeper        *mock.NexusMock
		ctx                sdk.Context
		expectedResp       types.EventResponse
		grpcQuerier        *evmKeeper.Querier
		existingChain      string
		nonExistingChain   string
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
						TxId:   types.Hash(common.HexToHash(existingTxID)),
						Index:  0,
						Status: existingStatus,
						Event:  nil,
					}, true
				}
				return types.Event{}, false
			},
		}

		baseKeeper = &mock.BaseKeeperMock{
			HasChainFunc: func(_ sdk.Context, chain string) bool {
				return chain == existingChain
			},
			ForChainFunc: func(chain string) types.ChainKeeper {
				return chainKeeper
			},
		}

		q := evmKeeper.NewGRPCQuerier(baseKeeper, nexusKeeper, signer)
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
					TxId:   types.Hash(common.HexToHash(existingTxID)),
					Index:  0,
					Status: existingStatus,
					Event:  nil,
				},
			}

			res, err := grpcQuerier.Event(sdk.WrapSDKContext(ctx), &types.EventRequest{
				Chain:   existingChain,
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
			Chain:   nonExistingChain,
			EventId: existingEventID,
		})

		assert := assert.New(t)
		assert.Error(err)

		assert.Equal(err.Error(), fmt.Sprintf("rpc error: code = NotFound desc = [%s] is not a registered chain: bridge error", nonExistingChain))
	}).Repeat(repeatCount))

	t.Run("event doesn't exist", testutils.Func(func(t *testing.T) {
		setup()
		_, err := grpcQuerier.Event(sdk.WrapSDKContext(ctx), &types.EventRequest{
			Chain:   existingChain,
			EventId: nonExistingEventID,
		})

		assert := assert.New(t)
		assert.Error(err)

		assert.Equal(err.Error(), fmt.Sprintf("rpc error: code = NotFound desc = no event with ID [%s] was found: bridge error", nonExistingEventID))
	}).Repeat(repeatCount))
}
