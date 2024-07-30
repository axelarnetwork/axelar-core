package axelarnet_test

import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/CosmWasm/wasmd/x/wasm"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/axelarnet"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/keeper"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types/mock"
	evmtypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	evmtestutils "github.com/axelarnetwork/axelar-core/x/evm/types/testutils"
	nexusTypes "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	nexustestutils "github.com/axelarnetwork/axelar-core/x/nexus/exported/testutils"
	. "github.com/axelarnetwork/utils/test"
)

func TestProposalHandler(t *testing.T) {
	var (
		ctx      sdk.Context
		k        keeper.Keeper
		n        *mock.NexusMock
		a        *mock.AccountKeeperMock
		handler  govtypes.Handler
		proposal *types.CallContractsProposal

		destChain         nexusTypes.Chain
		contractCall      types.ContractCall
		genMsg            nexusTypes.GeneralMessage
		governanceAccount sdk.AccAddress
	)

	givenProposal := Given("a CallContractsProposal", func() {
		ctx, k, _ = setup()

		destChain = nexustestutils.RandomChain()
		destChain.Module = evmtypes.ModuleName
		governanceAccount = rand.AccAddr()

		contractCall = types.ContractCall{
			Chain:           destChain.Name,
			ContractAddress: evmtestutils.RandomAddress().Hex(),
			Payload:         rand.BytesBetween(100, 500),
		}

		proposal = &types.CallContractsProposal{
			Title:         "Test Proposal",
			Description:   "This is a test proposal",
			ContractCalls: []types.ContractCall{contractCall},
		}

		n = &mock.NexusMock{
			SetNewMessageFunc: func(ctx sdk.Context, msg nexusTypes.GeneralMessage) error {
				genMsg = msg
				return nil
			},
			GenerateMessageIDFunc: func(ctx sdk.Context) (string, []byte, uint64) {
				hash := rand.Bytes(32)
				return fmt.Sprintf("%s-%d", hex.EncodeToString(hash[:]), 0), hash[:], 0
			},
		}

		a = &mock.AccountKeeperMock{
			GetModuleAddressFunc: func(name string) sdk.AccAddress {
				return governanceAccount
			},
		}

		handler = axelarnet.NewProposalHandler(k, n, a)
	})

	whenDestChainIsFound := givenProposal.
		When("destination chain is found in nexus", func() {
			n.GetChainFunc = func(ctx sdk.Context, chain nexusTypes.ChainName) (nexusTypes.Chain, bool) {
				return destChain, true
			}
		})

	whenDestChainIsNotFound := givenProposal.
		When("destination chain is not found in nexus", func() {
			n.GetChainFunc = func(ctx sdk.Context, chain nexusTypes.ChainName) (nexusTypes.Chain, bool) {
				if chain == exported.Axelarnet.Name {
					return exported.Axelarnet, true
				}
				return nexusTypes.Chain{}, false
			}
		})

	whenDestChainIsFound.
		Then("should set new message in nexus", func(t *testing.T) {
			err := handler(ctx, proposal)
			assert.NoError(t, err)

			assert.Equal(t, genMsg.Sender.Address, governanceAccount.String())
			assert.Equal(t, genMsg.Sender.Chain, exported.Axelarnet)
			assert.Equal(t, genMsg.Recipient.Chain, destChain)
			assert.Equal(t, genMsg.Recipient.Address, contractCall.ContractAddress)
			assert.Equal(t, genMsg.PayloadHash, crypto.Keccak256(contractCall.Payload))
		}).
		Run(t)

	whenDestChainIsNotFound.
		Then("should set new message in nexus with wasm chain", func(t *testing.T) {
			err := handler(ctx, proposal)
			assert.NoError(t, err)

			assert.Equal(t, genMsg.Sender.Address, governanceAccount.String())
			assert.Equal(t, genMsg.Sender.Chain, exported.Axelarnet)
			assert.Equal(t, genMsg.Recipient.Chain.Name, nexusTypes.ChainName(strings.ToLower(contractCall.Chain.String())))
			assert.Equal(t, genMsg.Recipient.Chain.Module, wasm.ModuleName)
			assert.Equal(t, genMsg.Recipient.Address, contractCall.ContractAddress)
			assert.Equal(t, genMsg.PayloadHash, crypto.Keccak256(contractCall.Payload))
		}).
		Run(t)
}
