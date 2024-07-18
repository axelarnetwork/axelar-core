package evm_test

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	geth "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/app"
	broadcastmock "github.com/axelarnetwork/axelar-core/sdk-utils/broadcast/mock"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/vald/evm"
	evmmock "github.com/axelarnetwork/axelar-core/vald/evm/mock"
	evmrpc "github.com/axelarnetwork/axelar-core/vald/evm/rpc"
	"github.com/axelarnetwork/axelar-core/vald/evm/rpc/mock"
	"github.com/axelarnetwork/axelar-core/x/evm/exported"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
	votetypes "github.com/axelarnetwork/axelar-core/x/vote/types"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/monads/results"
	"github.com/axelarnetwork/utils/slices"
	. "github.com/axelarnetwork/utils/test"
)

func TestMgr_ProcessTransferKeyConfirmation(t *testing.T) {
	var (
		mgr            *evm.Mgr
		event          *types.ConfirmKeyTransferStarted
		rpc            *mock.ClientMock
		broadcaster    *broadcastmock.BroadcasterMock
		txID           types.Hash
		gatewayAddress types.Address
		pollID         vote.PollID
		txReceipt      *geth.Receipt
		valAddr        sdk.ValAddress
	)

	givenEvmMgr := Given("EVM mgr", func() {
		rpc = &mock.ClientMock{}
		broadcaster = &broadcastmock.BroadcasterMock{
			BroadcastFunc: func(ctx context.Context, msgs ...sdk.Msg) (*sdk.TxResponse, error) { return nil, nil },
		}
		evmMap := make(map[string]evmrpc.Client)
		evmMap["ethereum"] = rpc
		valAddr = rand.ValAddr()
		mgr = evm.NewMgr(evmMap, broadcaster, valAddr, rand.AccAddr(), "", &evmmock.LatestFinalizedBlockCacheMock{
			GetFunc: func(_ nexus.ChainName) *big.Int { return big.NewInt(0) },
			SetFunc: func(_ nexus.ChainName, _ *big.Int) {},
		})
	})

	givenTxReceiptAndBlockAreFound := Given("tx receipt and block can be found", func() {
		tx := geth.NewTransaction(0, common.BytesToAddress(rand.Bytes(common.HashLength)), big.NewInt(0), 21000, big.NewInt(1), []byte{})
		blockNumber := uint64(rand.I64Between(1, 1000))

		txID = types.Hash(tx.Hash())
		txReceipt = &geth.Receipt{
			TxHash:      common.Hash(txID),
			BlockNumber: big.NewInt(rand.I64Between(0, int64(blockNumber-types.DefaultParams()[0].ConfirmationHeight+2))),
			Logs:        []*geth.Log{},
			Status:      1,
		}

		rpc.TransactionReceiptsFunc = func(ctx context.Context, txHashes []common.Hash) ([]evmrpc.TxReceiptResult, error) {
			return slices.Map(txHashes, func(hash common.Hash) evmrpc.TxReceiptResult {
				if bytes.Equal(hash.Bytes(), txID.Bytes()) {
					return evmrpc.TxReceiptResult(results.FromOk(*txReceipt))
				}

				return evmrpc.TxReceiptResult(results.FromErr[geth.Receipt](ethereum.NotFound))
			}), nil
		}
		rpc.HeaderByNumberFunc = func(ctx context.Context, number *big.Int) (*evmrpc.Header, error) {
			if number.Cmp(txReceipt.BlockNumber) == 0 {
				number := hexutil.Big(*big.NewInt(int64(blockNumber)))
				return &evmrpc.Header{Number: &number, Transactions: []common.Hash{txReceipt.TxHash}}, nil
			}

			return nil, fmt.Errorf("not found")
		}
		rpc.LatestFinalizedBlockNumberFunc = func(ctx context.Context, confirmations uint64) (*big.Int, error) {
			return txReceipt.BlockNumber, nil
		}
	})

	givenEventConfirmKeyTransfer := Given("event confirm key transfer", func() {
		gatewayAddress = types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength)))
		pollID = vote.PollID(rand.PosI64())
		event = types.NewConfirmKeyTransferStarted(
			exported.Ethereum.Name,
			txID,
			gatewayAddress,
			types.DefaultParams()[0].ConfirmationHeight,
			vote.PollParticipants{
				PollID:       pollID,
				Participants: []sdk.ValAddress{valAddr},
			},
		)
	})

	assertAndGetVoteEvents := func(t *testing.T, isEmpty bool) *types.VoteEvents {
		assert.Len(t, broadcaster.BroadcastCalls(), 1)
		assert.Len(t, broadcaster.BroadcastCalls()[0].Msgs, 1)

		voteEvents := broadcaster.BroadcastCalls()[0].Msgs[0].(*votetypes.VoteRequest).Vote.GetCachedValue().(*types.VoteEvents)
		if isEmpty {
			assert.Empty(t, voteEvents.Events)
		} else {
			assert.Len(t, voteEvents.Events, 1)
		}

		return voteEvents
	}

	thenShouldVoteNoEvent := Then("should vote no event", func(t *testing.T) {
		err := mgr.ProcessTransferKeyConfirmation(event)
		assert.NoError(t, err)

		assertAndGetVoteEvents(t, true)
	})

	givenEvmMgr.
		Given2(givenTxReceiptAndBlockAreFound).
		Given2(givenEventConfirmKeyTransfer).
		Branch(
			When("is not operatorship transferred event", func() {
				txReceipt.Logs = append(txReceipt.Logs, &geth.Log{
					Address: common.Address(gatewayAddress),
					Topics:  []common.Hash{common.BytesToHash(rand.Bytes(common.HashLength))},
				})
			}).
				Then2(thenShouldVoteNoEvent),

			When("is not emitted from the gateway", func() {
				txReceipt.Logs = append(txReceipt.Logs, &geth.Log{
					Address: common.BytesToAddress(rand.Bytes(common.AddressLength)),
					Topics:  []common.Hash{evm.MultisigTransferOperatorshipSig},
				})
			}).
				Then2(thenShouldVoteNoEvent),

			When("is invalid operatorship transferred event", func() {
				txReceipt.Logs = append(txReceipt.Logs, &geth.Log{
					Address: common.Address(gatewayAddress),
					Topics:  []common.Hash{evm.MultisigTransferOperatorshipSig},
					Data:    rand.Bytes(int(rand.I64Between(0, 1000))),
				})
			}).
				Then2(thenShouldVoteNoEvent),

			When("is valid operatorship transferred event", func() {
				newOperatorsData := common.Hex2Bytes("00000000000000000000000000000000000000000000000000000000000000600000000000000000000000000000000000000000000000000000000000000180000000000000000000000000000000000000000000000000000000000000001e000000000000000000000000000000000000000000000000000000000000000800000000000000000000000019cc2044857d23129a29f763d0338da837ce35f60000000000000000000000002ab6fa7de5e9e9423125a4246e4de1b9c755607400000000000000000000000037cc4b7e8f9f505ca8126db8a9d070566ed5dae70000000000000000000000003e56f0d4497ac44993d9ea272d4707f8be6b42a6000000000000000000000000462b96f617d5d92f63f9949c6f4626623ea73fa400000000000000000000000068b93045fe7d8794a7caf327e7f855cd6cd03bb80000000000000000000000009e77c30badbbc412a0c20c6ce43b671c6f103434000000000000000000000000c1c0c8d2131cc866834c6382096eadfef1af2f52000000000000000000000000000000000000000000000000000000000000000800000000000000000000000000000000000000000000000000000000000000030000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000070000000000000000000000000000000000000000000000000000000000000005")
				txReceipt.Logs = append(txReceipt.Logs, &geth.Log{})
				txReceipt.Logs = append(txReceipt.Logs, &geth.Log{
					Address: common.Address(gatewayAddress),
					Topics:  []common.Hash{evm.MultisigTransferOperatorshipSig},
					Data:    funcs.Must(abi.Arguments{{Type: funcs.Must(abi.NewType("bytes", "bytes", nil))}}.Pack(newOperatorsData)),
				})
			}).
				Then("should vote for the correct event", func(t *testing.T) {
					err := mgr.ProcessTransferKeyConfirmation(event)
					assert.NoError(t, err)

					actual := assertAndGetVoteEvents(t, false)
					assert.Equal(t, exported.Ethereum.Name, actual.Chain)
					assert.Equal(t, exported.Ethereum.Name, actual.Events[0].Chain)
					assert.Equal(t, txID, actual.Events[0].TxID)
					assert.EqualValues(t, 1, actual.Events[0].Index)
					assert.IsType(t, &types.Event_MultisigOperatorshipTransferred{}, actual.Events[0].Event)

					actualEvent := actual.Events[0].Event.(*types.Event_MultisigOperatorshipTransferred)
					assert.Len(t, actualEvent.MultisigOperatorshipTransferred.NewOperators, 8)
					assert.Len(t, actualEvent.MultisigOperatorshipTransferred.NewWeights, 8)
					assert.EqualValues(t, 30, actualEvent.MultisigOperatorshipTransferred.NewThreshold.BigInt().Int64())
				}),
		).
		Run(t, 5)
}

func TestMgr_ProcessTransferKeyConfirmation_FileCoinTestnetRescue(t *testing.T) {
	var (
		mgr         *evm.Mgr
		event       *types.ConfirmKeyTransferStarted
		rpc         *mock.ClientMock
		broadcaster *broadcastmock.BroadcasterMock
	)

	app.SetConfig()
	encoding := app.MakeEncodingConfig()

	chain := nexus.ChainName("filecoin-2")
	var participants vote.PollParticipants
	funcs.MustNoErr(encoding.Codec.UnmarshalJSON([]byte(`{"poll_id":"1342934","participants":["axelarvaloper1q8g8dmuc7x2uz9kkhf0tw364rxx96mntvp2zts","axelarvaloper1qn6e260hnjhl8ufqppq5ppymx7e6ek03z7sl9w","axelarvaloper1ql2cpr7sq52kvzzqyltahnm25qnj6fu9eegdk7","axelarvaloper1pvmzv87l6jwvuff7e96ty6xhtm3ryagd0047xx","axelarvaloper1z755tuclthwkmzc6aq70q4njjha7wh07gn8vpl","axelarvaloper1r6n590p42qzj8mfvmqm965s3hcksw9vny805dh","axelarvaloper1yrgmkhufaq5658z5ludx33mvtjgn6e4psqgs0n","axelarvaloper1y9q0v4sjlnf6d7n4vewp6f8fnnfg8z6glfsnae","axelarvaloper1ymq2mtjcgy7nh2qy8rcnyfd95kuwayxtwrczqy","axelarvaloper19ze2qz8p3nv7ayvawnspkttcnk6yjafaw9edx8","axelarvaloper1x5wgh6vwye60wv3dtshs9dmqggwfx2ldh0v54p","axelarvaloper1xu9d223797jud23u53rkk5zy9gwy730d62rvd8","axelarvaloper1243yj8nwd4c6dcqxtg7lhltsslv58dhpkjjxdf","axelarvaloper1duae8kuzne6neuqkttxa7w335enn4anjsl2sse","axelarvaloper1spwcc5856py7styddn45synlawqyv5fdxj0jee","axelarvaloper14zzgt08fp4e4rwdtdfgv57x6hcdan6vjzcjx8u","axelarvaloper1agj8h9cuzsxyclam2lma80at065mhxtm4xazh2","axelarvaloper1letwg3pgtqwcl7jfuxaplsvglw7h55233sn3x4"]}`), &participants))

	givenEvmMgr := Given("EVM mgr", func() {
		rpc = &mock.ClientMock{}
		broadcaster = &broadcastmock.BroadcasterMock{
			BroadcastFunc: func(ctx context.Context, msgs ...sdk.Msg) (*sdk.TxResponse, error) { return nil, nil },
		}
		evmMap := make(map[string]evmrpc.Client)
		evmMap[string(chain)] = rpc
		mgr = evm.NewMgr(evmMap, broadcaster, participants.Participants[0], rand.AccAddr(), "axelar-testnet-lisbon-3", &evmmock.LatestFinalizedBlockCacheMock{
			GetFunc: func(_ nexus.ChainName) *big.Int { return big.NewInt(0) },
			SetFunc: func(_ nexus.ChainName, _ *big.Int) {},
		})
	})

	givenTxReceiptIsNotFound := Given("tx receipt and block can not be found", func() {
		rpc.TransactionReceiptsFunc = func(ctx context.Context, txHashes []common.Hash) ([]evmrpc.TxReceiptResult, error) {
			return nil, fmt.Errorf("not found")
		}
		rpc.HeaderByNumberFunc = func(ctx context.Context, number *big.Int) (*evmrpc.Header, error) {
			return nil, fmt.Errorf("not found")
		}
		rpc.LatestFinalizedBlockNumberFunc = func(ctx context.Context, confirmations uint64) (*big.Int, error) {
			return big.NewInt(1_796_888), nil
		}
	})

	assertAndGetVoteEvents := func(t *testing.T) *types.VoteEvents {
		assert.Len(t, broadcaster.BroadcastCalls(), 1)
		assert.Len(t, broadcaster.BroadcastCalls()[0].Msgs, 1)

		voteEvents := broadcaster.BroadcastCalls()[0].Msgs[0].(*votetypes.VoteRequest).Vote.GetCachedValue().(*types.VoteEvents)
		assert.Len(t, voteEvents.Events, 1)

		return voteEvents
	}

	givenEvmMgr.
		Given2(givenTxReceiptIsNotFound).
		When("when confirming specific missed filecoin transfer event", func() {
			var abciEvent abci.Event
			funcs.MustNoErr(encoding.Codec.UnmarshalJSON([]byte(confirmKeyTransferStartedEvent), &abciEvent))
			event = funcs.Must(sdk.ParseTypedEvent(abciEvent)).(*types.ConfirmKeyTransferStarted)
		}).
		Then("should vote yes for event", func(t *testing.T) {
			err := mgr.ProcessTransferKeyConfirmation(event)
			assert.NoError(t, err)

			actual := assertAndGetVoteEvents(t)

			var expected types.VoteEvents
			funcs.MustNoErr(encoding.Codec.UnmarshalJSON([]byte(voteEvent), &expected))

			assert.Equal(t, &expected, actual)
		}).
		Run(t)
}

// received from 'axelard q tx C3E946DFC220318101F2748A0014138A5E18B847685CFA15866A8447F97F1FF7 --output json' on a testnet node
var confirmKeyTransferStartedEvent = "{ \"type\": \"axelar.evm.v1beta1.ConfirmKeyTransferStarted\", \"attributes\": [ { \"key\": \"Y2hhaW4=\", \"value\": \"ImZpbGVjb2luLTIi\", \"index\": true }, { \"key\": \"Y29uZmlybWF0aW9uX2hlaWdodA==\", \"value\": \"IjEwIg==\", \"index\": true }, { \"key\": \"Z2F0ZXdheV9hZGRyZXNz\", \"value\": \"WzE1MywxNDUsMjMsMjEyLDY2LDMyLDI0Myw2Miw0LDY1LDI1MSwxNzEsNDIsOTAsMjE5LDE0MywyNDQsMTMzLDE5Nyw3N10=\", \"index\": true }, { \"key\": \"cGFydGljaXBhbnRz\", \"value\": \"eyJwb2xsX2lkIjoiMTM0MjkzNCIsInBhcnRpY2lwYW50cyI6WyJheGVsYXJ2YWxvcGVyMXE4ZzhkbXVjN3gydXo5a2toZjB0dzM2NHJ4eDk2bW50dnAyenRzIiwiYXhlbGFydmFsb3BlcjFxbjZlMjYwaG5qaGw4dWZxcHBxNXBweW14N2U2ZWswM3o3c2w5dyIsImF4ZWxhcnZhbG9wZXIxcWwyY3ByN3NxNTJrdnp6cXlsdGFobm0yNXFuajZmdTllZWdkazciLCJheGVsYXJ2YWxvcGVyMXB2bXp2ODdsNmp3dnVmZjdlOTZ0eTZ4aHRtM3J5YWdkMDA0N3h4IiwiYXhlbGFydmFsb3BlcjF6NzU1dHVjbHRod2ttemM2YXE3MHE0bmpqaGE3d2gwN2duOHZwbCIsImF4ZWxhcnZhbG9wZXIxcjZuNTkwcDQycXpqOG1mdm1xbTk2NXMzaGNrc3c5dm55ODA1ZGgiLCJheGVsYXJ2YWxvcGVyMXlyZ21raHVmYXE1NjU4ejVsdWR4MzNtdnRqZ242ZTRwc3FnczBuIiwiYXhlbGFydmFsb3BlcjF5OXEwdjRzamxuZjZkN240dmV3cDZmOGZubmZnOHo2Z2xmc25hZSIsImF4ZWxhcnZhbG9wZXIxeW1xMm10amNneTduaDJxeThyY255ZmQ5NWt1d2F5eHR3cmN6cXkiLCJheGVsYXJ2YWxvcGVyMTl6ZTJxejhwM252N2F5dmF3bnNwa3R0Y25rNnlqYWZhdzllZHg4IiwiYXhlbGFydmFsb3BlcjF4NXdnaDZ2d3llNjB3djNkdHNoczlkbXFnZ3dmeDJsZGgwdjU0cCIsImF4ZWxhcnZhbG9wZXIxeHU5ZDIyMzc5N2p1ZDIzdTUzcmtrNXp5OWd3eTczMGQ2MnJ2ZDgiLCJheGVsYXJ2YWxvcGVyMTI0M3lqOG53ZDRjNmRjcXh0ZzdsaGx0c3NsdjU4ZGhwa2pqeGRmIiwiYXhlbGFydmFsb3BlcjFkdWFlOGt1em5lNm5ldXFrdHR4YTd3MzM1ZW5uNGFuanNsMnNzZSIsImF4ZWxhcnZhbG9wZXIxc3B3Y2M1ODU2cHk3c3R5ZGRuNDVzeW5sYXdxeXY1ZmR4ajBqZWUiLCJheGVsYXJ2YWxvcGVyMTR6emd0MDhmcDRlNHJ3ZHRkZmd2NTd4NmhjZGFuNnZqemNqeDh1IiwiYXhlbGFydmFsb3BlcjFhZ2o4aDljdXpzeHljbGFtMmxtYTgwYXQwNjVtaHh0bTR4YXpoMiIsImF4ZWxhcnZhbG9wZXIxbGV0d2czcGd0cXdjbDdqZnV4YXBsc3ZnbHc3aDU1MjMzc24zeDQiXX0=\", \"index\": true }, { \"key\": \"dHhfaWQ=\", \"value\": \"WzIxNywyMDQsMTQxLDk1LDIsNTYsMjIzLDIyNyw5NiwxMTksMTEwLDIxNywyLDE1Miw1MywyMTYsMTY0LDExMiwyMTIsMjExLDE0LDcwLDE2NCw2MSwyMzUsMTE4LDI0MywyMiwxOTIsMTg5LDg3LDY0XQ==\", \"index\": true } ] }"

// received from 'axelard q tx D09FABF26ED165F7ACEF738DB3CADD01595C26857B9EB63C7DDB2A38B56FB1FF --output json' on a testnet node
var voteEvent = "{\"chain\":\"filecoin-2\",\"events\":[{\"chain\":\"filecoin-2\",\"tx_id\":[217,204,141,95,2,56,223,227,96,119,110,217,2,152,53,216,164,112,212,211,14,70,164,61,235,118,243,22,192,189,87,64],\"index\":\"0\",\"status\":\"STATUS_UNSPECIFIED\",\"multisig_operatorship_transferred\":{\"new_operators\":[[11,7,202,144,174,251,96,155,200,110,108,37,173,200,49,73,50,1,216,205],[13,228,121,87,86,165,184,211,114,100,160,237,98,109,235,149,8,23,101,240],[15,123,242,71,190,51,50,167,197,148,199,73,93,22,70,50,56,165,46,48],[15,237,3,37,192,97,245,31,75,51,167,134,228,99,87,45,116,169,244,93],[16,142,244,80,139,53,144,17,134,103,160,116,238,76,211,237,52,61,116,91],[17,36,65,252,8,185,236,166,98,176,218,133,242,23,1,69,76,87,204,7],[21,254,39,85,138,239,117,161,240,183,220,250,128,18,46,34,236,221,54,49],[26,69,110,16,179,96,209,130,228,138,67,242,18,223,60,226,45,213,54,157],[27,134,71,206,138,155,197,45,57,81,176,79,96,188,107,118,64,28,226,137],[34,89,108,239,113,123,92,11,85,171,131,210,161,35,62,115,115,62,129,147],[44,199,206,41,20,193,122,159,143,228,61,135,156,210,132,130,207,21,191,171],[52,2,173,254,109,230,217,47,197,29,12,194,37,223,45,14,55,216,199,2],[59,189,165,2,14,88,6,215,190,5,100,37,207,70,1,29,96,105,111,23],[69,230,253,120,247,53,6,83,211,92,145,122,45,142,20,76,144,10,6,193],[86,199,65,231,54,236,211,239,52,177,183,36,126,93,143,217,166,38,23,38],[90,153,54,110,73,26,77,77,28,254,195,33,97,175,115,141,110,112,158,144],[95,183,67,11,80,252,228,77,22,121,123,32,139,162,223,36,86,190,254,235],[97,17,201,201,191,64,109,95,224,32,196,238,196,20,109,163,29,155,129,9],[100,129,91,34,41,54,166,63,76,28,14,208,161,170,71,217,251,16,179,219],[108,70,40,49,250,203,163,144,79,107,167,153,56,26,197,10,136,147,40,212],[111,225,42,171,35,206,109,31,238,72,213,99,89,95,239,3,153,52,206,190],[112,14,40,85,188,24,29,209,97,45,152,174,126,219,31,137,17,6,135,187],[113,224,206,124,76,70,120,114,156,228,238,226,205,244,178,164,105,26,78,194],[117,192,161,167,79,99,113,8,207,10,46,216,245,171,54,214,203,15,27,219],[118,179,69,8,178,77,197,140,246,110,173,224,81,40,107,134,78,208,85,52],[125,86,226,40,90,96,14,113,48,103,53,19,77,32,173,64,21,142,121,150],[136,148,206,76,66,98,101,0,165,25,9,31,5,50,55,2,202,220,30,189],[137,8,103,211,214,163,18,37,141,118,61,2,11,138,176,148,104,100,40,221],[140,24,101,84,242,226,205,188,74,46,90,233,22,66,0,126,215,178,234,163],[142,8,158,246,198,113,163,153,243,252,165,169,183,224,216,85,80,37,169,206],[142,172,122,15,169,135,192,182,167,56,15,133,107,17,11,1,68,134,140,141],[143,181,155,148,9,79,99,230,118,233,200,74,4,224,68,219,11,121,63,143],[145,3,229,33,0,56,123,221,214,214,166,174,225,127,202,171,154,246,195,78],[151,144,2,5,81,61,159,130,83,233,145,72,247,49,89,241,231,211,89,6],[163,24,157,254,205,234,20,199,233,111,4,121,25,117,8,63,177,174,18,75],[167,107,20,183,176,225,168,26,230,23,168,204,255,34,128,5,104,221,190,55],[170,187,31,5,55,125,1,65,54,206,92,251,17,34,119,106,118,18,149,206],[176,189,117,157,248,203,253,203,237,145,154,94,127,219,238,29,243,86,165,107],[181,206,32,74,213,152,189,206,14,48,143,142,73,47,146,136,119,151,143,99],[194,198,63,125,69,177,71,95,50,6,38,62,158,31,181,74,88,120,173,192],[198,34,36,1,123,50,38,73,158,175,14,221,57,151,204,109,28,27,142,110],[202,153,170,54,150,71,135,176,117,242,157,187,185,223,85,47,101,126,210,20],[207,41,42,249,114,20,58,213,89,251,177,181,86,54,171,213,13,180,34,29],[215,1,131,227,23,116,193,251,47,81,165,101,92,226,76,188,37,210,199,70],[221,85,67,12,233,26,182,99,29,206,235,148,200,101,149,105,194,47,244,143],[221,126,89,205,71,17,173,41,120,55,78,198,195,124,232,184,90,4,82,8],[236,40,187,140,57,147,247,255,150,114,224,183,249,52,7,56,43,79,96,29],[240,169,209,98,226,137,90,223,38,166,66,13,14,134,24,109,154,176,32,239],[248,156,104,80,228,217,172,17,225,212,41,156,145,170,65,198,122,9,202,28],[254,87,113,178,104,164,217,36,253,22,220,153,210,147,162,202,84,231,92,158]],\"new_threshold\":\"26642\",\"new_weights\":[\"338\",\"321\",\"1000\",\"142\",\"145\",\"452\",\"8623\",\"148\",\"318\",\"731\",\"158\",\"141\",\"19\",\"362\",\"171\",\"150\",\"20\",\"17\",\"145\",\"148\",\"155\",\"181\",\"164\",\"8\",\"150\",\"712\",\"244\",\"424\",\"231\",\"1002\",\"153\",\"141\",\"141\",\"161\",\"787\",\"727\",\"483\",\"219\",\"614\",\"8741\",\"8821\",\"177\",\"461\",\"149\",\"707\",\"949\",\"142\",\"566\",\"205\",\"697\"]}}]}"
