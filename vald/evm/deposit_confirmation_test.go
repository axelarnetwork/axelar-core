package evm_test

import (
	"bytes"
	"context"
	"math/big"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	geth "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"

	broadcastmock "github.com/axelarnetwork/axelar-core/sdk-utils/broadcast/mock"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/vald/evm"
	evmmock "github.com/axelarnetwork/axelar-core/vald/evm/mock"
	evmRpc "github.com/axelarnetwork/axelar-core/vald/evm/rpc"
	"github.com/axelarnetwork/axelar-core/vald/evm/rpc/mock"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/evm/types/testutils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
	votetypes "github.com/axelarnetwork/axelar-core/x/vote/types"
	"github.com/axelarnetwork/utils/monads/results"
	"github.com/axelarnetwork/utils/slices"
	. "github.com/axelarnetwork/utils/test"
)

func TestMgr_ProccessDepositConfirmation(t *testing.T) {
	var (
		mgr         *evm.Mgr
		receipt     *geth.Receipt
		tokenAddr   types.Address
		depositAddr types.Address
		amount      sdk.Uint
		evmMap      map[string]evmRpc.Client
		rpc         *mock.ClientMock
		valAddr     sdk.ValAddress

		votes []*types.VoteEvents
		err   error
	)

	givenDeposit := Given("a deposit has been made", func() {
		amount = sdk.NewUint(uint64(rand.PosI64()))
		tokenAddr = testutils.RandomAddress()
		depositAddr = testutils.RandomAddress()
		randomTokenDeposit := &geth.Log{
			Address: common.Address(testutils.RandomAddress()),
			Topics: []common.Hash{
				evm.ERC20TransferSig,
				padToHash(testutils.RandomAddress()),
				padToHash(depositAddr),
			},
			Data: padToHash(big.NewInt(rand.PosI64())).Bytes(),
		}
		randomEvent := &geth.Log{
			Address: common.Address(tokenAddr),
			Topics: []common.Hash{
				common.Hash(testutils.RandomHash()),
				padToHash(testutils.RandomAddress()),
				padToHash(depositAddr),
			},
			Data: padToHash(big.NewInt(rand.PosI64())).Bytes(),
		}

		invalidDeposit := &geth.Log{
			Address: common.Address(tokenAddr),
			Topics: []common.Hash{
				evm.ERC20TransferSig,
				padToHash(testutils.RandomAddress()),
			},
			Data: padToHash(big.NewInt(rand.PosI64())).Bytes(),
		}

		zeroAmountDeposit := &geth.Log{
			Address: common.Address(tokenAddr),
			Topics: []common.Hash{
				evm.ERC20TransferSig,
				padToHash(testutils.RandomAddress()),
				padToHash(depositAddr),
			},
			Data: padToHash(big.NewInt(0)).Bytes(),
		}

		validDeposit := &geth.Log{
			Address: common.Address(tokenAddr),
			Topics: []common.Hash{
				evm.ERC20TransferSig,
				padToHash(testutils.RandomAddress()),
				padToHash(depositAddr),
			},
			Data: padToHash(amount.BigInt()).Bytes(),
		}

		receipt = &geth.Receipt{
			TxHash:      common.Hash(testutils.RandomHash()),
			BlockNumber: big.NewInt(rand.PosI64()),
			Logs:        []*geth.Log{randomTokenDeposit, validDeposit, randomEvent, invalidDeposit, zeroAmountDeposit},
			Status:      1,
		}
	})

	confirmingDeposit := When("confirming the existing deposit", func() {
		event := testutils.RandomConfirmDepositStarted()
		event.TxID = types.Hash(receipt.TxHash)
		evmMap[strings.ToLower(event.Chain.String())] = rpc
		event.DepositAddress = depositAddr
		event.TokenAddress = tokenAddr
		event.Participants = append(event.Participants, valAddr)

		err = mgr.ProcessDepositConfirmation(&event)
	})

	reject := Then("reject the confirmation", func(t *testing.T) {
		assert.Len(t, votes, 1)
		assert.Len(t, votes[0].Events, 0)
	})

	noError := Then("return no error", func(t *testing.T) {
		assert.NoError(t, err)
	})

	Given("an evm manager", func() {
		votes = []*types.VoteEvents{}
		evmMap = make(map[string]evmRpc.Client, 1)

		broadcaster := &broadcastmock.BroadcasterMock{
			BroadcastFunc: func(_ context.Context, msgs ...sdk.Msg) (*sdk.TxResponse, error) {
				for _, msg := range msgs {
					votes = append(votes, msg.(*votetypes.VoteRequest).Vote.GetCachedValue().(*types.VoteEvents))
				}
				return nil, nil
			},
		}

		valAddr = rand.ValAddr()
		mgr = evm.NewMgr(evmMap, broadcaster, valAddr, rand.AccAddr(), &evmmock.LatestFinalizedBlockCacheMock{
			GetFunc: func(_ nexus.ChainName) *big.Int { return big.NewInt(0) },
			SetFunc: func(_ nexus.ChainName, _ *big.Int) {},
		})
	}).
		Given("an evm rpc client", func() {
			rpc = &mock.ClientMock{
				HeaderByNumberFunc: func(context.Context, *big.Int) (*evmRpc.Header, error) {
					return &evmRpc.Header{Transactions: []common.Hash{receipt.TxHash}}, nil
				},
				TransactionReceiptsFunc: func(ctx context.Context, txHashes []common.Hash) ([]evmRpc.TxReceiptResult, error) {
					return slices.Map(txHashes, func(txHash common.Hash) evmRpc.TxReceiptResult {
						if bytes.Equal(txHash.Bytes(), receipt.TxHash.Bytes()) {
							return evmRpc.TxReceiptResult(results.FromOk(*receipt))
						}

						return evmRpc.TxReceiptResult(results.FromErr[geth.Receipt](ethereum.NotFound))
					}), nil
				},
				LatestFinalizedBlockNumberFunc: func(ctx context.Context, confirmations uint64) (*big.Int, error) {
					return receipt.BlockNumber, nil
				},
			}
		}).
		Branch(
			Given("no deposit has been made", func() {
				rpc.TransactionReceiptsFunc = func(ctx context.Context, txHashes []common.Hash) ([]evmRpc.TxReceiptResult, error) {
					return slices.Map(txHashes, func(hash common.Hash) evmRpc.TxReceiptResult {
						return evmRpc.TxReceiptResult(results.FromErr[geth.Receipt](ethereum.NotFound))
					}), nil
				}
			}).
				When("confirming a random deposit on the correct chain", func() {
					event := testutils.RandomConfirmDepositStarted()
					event.Participants = append(event.Participants, valAddr)

					evmMap[strings.ToLower(event.Chain.String())] = rpc
					err = mgr.ProcessDepositConfirmation(&event)
				}).
				Then2(noError).
				Then2(reject),

			givenDeposit.
				When("confirming the deposit on unsupported chain", func() {
					event := testutils.RandomConfirmDepositStarted()
					event.TxID = types.Hash(receipt.TxHash)
					event.DepositAddress = depositAddr
					event.TokenAddress = tokenAddr
					event.Participants = append(event.Participants, valAddr)

					err = mgr.ProcessDepositConfirmation(&event)
				}).
				Then("return error", func(t *testing.T) {
					assert.Error(t, err)
				}),

			givenDeposit.
				Given("confirmation height is not reached yet", func() {
					rpc.LatestFinalizedBlockNumberFunc = func(ctx context.Context, confirmations uint64) (*big.Int, error) {
						return sdk.NewIntFromBigInt(receipt.BlockNumber).SubRaw(int64(confirmations)).BigInt(), nil
					}
				}).
				When2(confirmingDeposit).
				Then2(noError).
				Then2(reject),

			givenDeposit.
				When2(confirmingDeposit).
				Then2(noError).
				Then("accept the confirmation", func(t *testing.T) {
					assert.Len(t, votes, 1)
					assert.Len(t, votes[0].Events, 1)
					transferEvent, ok := votes[0].Events[0].Event.(*types.Event_Transfer)
					assert.True(t, ok)

					assert.Equal(t, depositAddr, transferEvent.Transfer.To)
					assert.True(t, transferEvent.Transfer.Amount.Equal(amount))
				}),

			givenDeposit.
				When("confirming event with wrong tx ID", func() {
					event := testutils.RandomConfirmDepositStarted()
					evmMap[strings.ToLower(event.Chain.String())] = rpc
					event.DepositAddress = depositAddr
					event.TokenAddress = tokenAddr
					event.Participants = append(event.Participants, valAddr)

					err = mgr.ProcessDepositConfirmation(&event)
				}).
				Then2(noError).
				Then2(reject),

			givenDeposit.
				When("confirming event with wrong token address", func() {
					event := testutils.RandomConfirmDepositStarted()
					event.TxID = types.Hash(receipt.TxHash)
					evmMap[strings.ToLower(event.Chain.String())] = rpc
					event.DepositAddress = depositAddr
					event.Participants = append(event.Participants, valAddr)

					err = mgr.ProcessDepositConfirmation(&event)
				}).
				Then2(noError).
				Then2(reject),

			givenDeposit.
				When("confirming event with wrong deposit address", func() {
					event := testutils.RandomConfirmDepositStarted()
					event.TxID = types.Hash(receipt.TxHash)
					evmMap[strings.ToLower(event.Chain.String())] = rpc
					event.TokenAddress = tokenAddr
					event.Participants = append(event.Participants, valAddr)

					err = mgr.ProcessDepositConfirmation(&event)
				}).
				Then2(noError).
				Then2(reject),

			givenDeposit.
				When("confirming a deposit without being a participant", func() {
					event := testutils.RandomConfirmDepositStarted()
					event.TxID = types.Hash(receipt.TxHash)
					evmMap[strings.ToLower(event.Chain.String())] = rpc
					event.DepositAddress = depositAddr
					event.TokenAddress = tokenAddr

					err = mgr.ProcessDepositConfirmation(&event)
				}).
				Then2(noError).
				Then("do nothing", func(t *testing.T) {
					assert.Len(t, votes, 0)
				}),

			givenDeposit.
				Given("multiple deposits in a single tx", func() {
					additionalAmount := sdk.NewUint(uint64(rand.PosI64()))
					amount = amount.Add(additionalAmount)
					additionalDeposit := &geth.Log{
						Address: common.Address(tokenAddr),
						Topics: []common.Hash{
							evm.ERC20TransferSig,
							padToHash(testutils.RandomAddress()),
							padToHash(depositAddr),
						},
						Data: padToHash(additionalAmount.BigInt()).Bytes(),
					}
					receipt.Logs = append(receipt.Logs, additionalDeposit)
				}).
				When("confirming the deposits", func() {
					event := testutils.RandomConfirmDepositStarted()
					event.TxID = types.Hash(receipt.TxHash)
					evmMap[strings.ToLower(event.Chain.String())] = rpc
					event.DepositAddress = depositAddr
					event.TokenAddress = tokenAddr
					event.Participants = append(event.Participants, valAddr)

					err = mgr.ProcessDepositConfirmation(&event)
				}).
				Then2(noError).
				Then("vote for all deposits in the same tx", func(t *testing.T) {
					assert.Len(t, votes, 1)
					assert.Len(t, votes[0].Events, 2)

					actualAmount := sdk.ZeroUint()
					for _, event := range votes[0].Events {
						transferEvent, ok := event.Event.(*types.Event_Transfer)
						assert.True(t, ok)
						assert.Equal(t, depositAddr, transferEvent.Transfer.To)

						actualAmount = actualAmount.Add(transferEvent.Transfer.Amount)
					}

					assert.True(t, actualAmount.Equal(amount))
				}),
		).
		Run(t, 20)
}

func TestMgr_ProccessDepositConfirmationNoTopicsNotPanics(t *testing.T) {
	chain := nexus.ChainName(strings.ToLower(rand.NormalizedStr(5)))
	receipt := geth.Receipt{
		Logs:        []*geth.Log{{Topics: make([]common.Hash, 0)}},
		BlockNumber: big.NewInt(1),
		Status:      geth.ReceiptStatusSuccessful,
	}
	rpcClient := &mock.ClientMock{TransactionReceiptsFunc: func(_ context.Context, _ []common.Hash) ([]evmRpc.TxReceiptResult, error) {
		return []evmRpc.TxReceiptResult{evmRpc.TxReceiptResult(results.FromOk(receipt))}, nil
	}}
	cache := &evmmock.LatestFinalizedBlockCacheMock{GetFunc: func(chain nexus.ChainName) *big.Int {
		return big.NewInt(100)
	}}

	broadcaster := &broadcastmock.BroadcasterMock{BroadcastFunc: func(_ context.Context, _ ...sdk.Msg) (*sdk.TxResponse, error) {
		return nil, nil
	}}

	valAddr := rand.ValAddr()
	mgr := evm.NewMgr(map[string]evmRpc.Client{chain.String(): rpcClient}, broadcaster, valAddr, rand.AccAddr(), cache)

	assert.NotPanics(t, func() {
		mgr.ProcessDepositConfirmation(&types.ConfirmDepositStarted{TxID: types.Hash{1},
			PollParticipants: exported.PollParticipants{PollID: 10, Participants: []sdk.ValAddress{valAddr}},
			Chain:            chain,
		})
	})
}

type byter interface {
	Bytes() []byte
}

func padToHash[T byter](x T) common.Hash {
	return common.BytesToHash(common.LeftPadBytes(x.Bytes(), common.HashLength))
}
