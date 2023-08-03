package keeper_test

import (
	"fmt"
	"testing"
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/keeper"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types/mock"
	evm "github.com/axelarnetwork/axelar-core/x/evm/exported"
	evmkeeper "github.com/axelarnetwork/axelar-core/x/evm/keeper"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/funcs"
	. "github.com/axelarnetwork/utils/test"
)

func TestAfterProposalDeposit(t *testing.T) {
	var (
		proposal govtypes.Proposal
	)

	encCfg := app.MakeEncodingConfig()
	subspace := params.NewSubspace(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("nexusKey"), sdk.NewKVStoreKey("tNexusKey"), "nexus")
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	nexusK := &mock.NexusMock{}
	govK := &mock.GovKeeperMock{}
	contractCall := types.ContractCall{
		Chain:           nexus.ChainName(rand.NormalizedStr(5)),
		ContractAddress: common.BytesToAddress(rand.Bytes(common.AddressLength)).Hex(),
		Payload:         rand.Bytes(100),
	}
	minDeposit := sdk.NewCoin("TEST", sdk.NewInt(rand.PosI64()))

	keeper := keeper.NewKeeper(encCfg.Codec, sdk.NewKVStoreKey("nexus"), subspace, &mock.ChannelKeeperMock{}, &mock.FeegrantKeeperMock{})
	keeper.SetParams(ctx, types.DefaultParams())

	nexusK.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
		return evm.Ethereum, true
	}

	Given("a proposal is created", func() {
		govK.GetProposalFunc = func(ctx sdk.Context, proposalID uint64) (govtypes.Proposal, bool) {
			return proposal, proposalID == proposal.ProposalId
		}
	}).
		Branch(
			When("the proposal is not a nexus call contracts proposal", func() {
				proposal = funcs.Must(govtypes.NewProposal(
					paramstypes.NewParameterChangeProposal("title", "description", []paramstypes.ParamChange{}),
					uint64(rand.I64Between(1, 100)),
					time.Now(),
					time.Now().AddDate(0, 0, 1),
				))
			}).
				Then("should not panic", func(t *testing.T) {
					assert.NotPanics(t, func() {
						keeper.Hooks(nexusK, govK).AfterProposalDeposit(ctx, proposal.ProposalId, rand.AccAddr())
					})
				}),

			When("the proposal is a nexus call contracts proposal", func() {
				proposal = funcs.Must(govtypes.NewProposal(
					types.NewCallContractsProposal("title", "description", []types.ContractCall{contractCall}),
					1,
					time.Now(),
					time.Now().AddDate(0, 0, 1),
				))
			}).
				Branch(
					When("keeper is setup with the default params", func() {
						keeper.SetParams(ctx, types.DefaultParams())
					}).
						Then("should not panic", func(t *testing.T) {
							assert.NotPanics(t, func() {
								keeper.Hooks(nexusK, govK).AfterProposalDeposit(ctx, proposal.ProposalId, rand.AccAddr())
							})
						}),

					When("keeper is setup with params that sets no min deposit for the contract call", func() {
						params := types.DefaultParams()
						params.CallContractsProposalMinDeposits = []types.CallContractProposalMinDeposit{
							{Chain: contractCall.Chain, ContractAddress: common.BytesToAddress(rand.Bytes(common.AddressLength)).Hex(), MinDeposits: sdk.NewCoins(minDeposit)},
							{Chain: nexus.ChainName(rand.NormalizedStr(5)), ContractAddress: contractCall.ContractAddress, MinDeposits: sdk.NewCoins(minDeposit)},
						}

						keeper.SetParams(ctx, params)
					}).
						Then("should not panic", func(t *testing.T) {
							assert.NotPanics(t, func() {
								keeper.Hooks(nexusK, govK).AfterProposalDeposit(ctx, proposal.ProposalId, rand.AccAddr())
							})
						}),

					When("keeper is setup with params that sets some min deposit for the contract call", func() {
						params := types.DefaultParams()
						params.CallContractsProposalMinDeposits = []types.CallContractProposalMinDeposit{
							{Chain: contractCall.Chain, ContractAddress: contractCall.ContractAddress, MinDeposits: sdk.NewCoins(minDeposit)},
						}

						keeper.SetParams(ctx, params)
					}).
						Branch(
							When("min deposit is met", func() {
								proposal.TotalDeposit = proposal.TotalDeposit.Add(minDeposit)
							}).
								Then("should not panic", func(t *testing.T) {
									assert.NotPanics(t, func() {
										keeper.Hooks(nexusK, govK).AfterProposalDeposit(ctx, proposal.ProposalId, rand.AccAddr())
									})
								}),

							When("min deposit is not met", func() {
								proposal.TotalDeposit = proposal.TotalDeposit.Add(minDeposit.SubAmount(sdk.NewInt(1)))
							}).
								Then("should panic", func(t *testing.T) {
									assert.PanicsWithError(t, fmt.Sprintf("proposal %d does not have enough deposits for calling contract %s on chain %s (required: %s, provided: %s)",
										proposal.ProposalId, contractCall.ContractAddress, contractCall.Chain, minDeposit.String(), proposal.TotalDeposit.String()),
										func() {
											keeper.Hooks(nexusK, govK).AfterProposalDeposit(ctx, proposal.ProposalId, rand.AccAddr())
										})
								}),
						),
				),
		).
		Run(t)
}

func TestAfterProposalSubmission(t *testing.T) {
	var (
		proposal govtypes.Proposal
	)

	encCfg := app.MakeEncodingConfig()
	subspace := params.NewSubspace(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("nexusKey"), sdk.NewKVStoreKey("tNexusKey"), "nexus")
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	nexusK := &mock.NexusMock{
		GetChainFunc: func(ctx sdk.Context, chainName nexus.ChainName) (nexus.Chain, bool) {
			return evm.Ethereum, chainName == evm.Ethereum.GetName()
		},
		ValidateAddressFunc: func(ctx sdk.Context, address nexus.CrossChainAddress) error {
			return evmkeeper.NewAddressValidator()(ctx, address)
		},
	}
	govK := &mock.GovKeeperMock{}
	nonRegisteredChain := nexus.ChainName(rand.NormalizedStr(5))

	keeper := keeper.NewKeeper(encCfg.Codec, sdk.NewKVStoreKey("nexus"), subspace, &mock.ChannelKeeperMock{}, &mock.FeegrantKeeperMock{})
	keeper.SetParams(ctx, types.DefaultParams())

	Given("a proposal is created", func() {
		govK.GetProposalFunc = func(ctx sdk.Context, proposalID uint64) (govtypes.Proposal, bool) {
			return proposal, proposalID == proposal.ProposalId
		}
	}).
		Branch(
			When("the proposal is not a nexus call contracts proposal", func() {
				proposal = funcs.Must(govtypes.NewProposal(
					paramstypes.NewParameterChangeProposal("title", "description", []paramstypes.ParamChange{}),
					uint64(rand.I64Between(1, 100)),
					time.Now(),
					time.Now().AddDate(0, 0, 1),
				))
			}).
				Then("should not panic", func(t *testing.T) {
					assert.NotPanics(t, func() {
						keeper.Hooks(nexusK, govK).AfterProposalSubmission(ctx, proposal.ProposalId)
					})
				}),

			When("the proposal is a nexus call contracts proposal", func() {
				proposal = funcs.Must(govtypes.NewProposal(
					types.NewCallContractsProposal("title", "description", []types.ContractCall{}),
					1,
					time.Now(),
					time.Now().AddDate(0, 0, 1),
				))
			}).
				Branch(
					When("contract call has non-registered chain", func() {
						proposal.Content = funcs.Must(codectypes.NewAnyWithValue(types.NewCallContractsProposal("title", "description", []types.ContractCall{
							{
								Chain:           nonRegisteredChain,
								ContractAddress: common.BytesToAddress(rand.Bytes(common.AddressLength)).Hex(),
								Payload:         rand.Bytes(100),
							},
						}).(proto.Message)))
					}).
						Then("should panic", func(t *testing.T) {
							assert.PanicsWithError(t, fmt.Sprintf("%s is not a registered chain", nonRegisteredChain), func() {
								keeper.Hooks(nexusK, govK).AfterProposalSubmission(ctx, proposal.ProposalId)
							})
						}),

					When("contract call has registered chain but invalid contract address", func() {
						proposal.Content = funcs.Must(codectypes.NewAnyWithValue(types.NewCallContractsProposal("title", "description", []types.ContractCall{
							{
								Chain:           evm.Ethereum.Name,
								ContractAddress: rand.NormalizedStr(42),
								Payload:         rand.Bytes(100),
							},
						}).(proto.Message)))
					}).
						Then("should panic", func(t *testing.T) {
							assert.PanicsWithError(t, "not an hex address", func() {
								keeper.Hooks(nexusK, govK).AfterProposalSubmission(ctx, proposal.ProposalId)
							})
						}),

					When("contract call has registered chain and valid contract address", func() {
						proposal.Content = funcs.Must(codectypes.NewAnyWithValue(types.NewCallContractsProposal("title", "description", []types.ContractCall{
							{
								Chain:           evm.Ethereum.Name,
								ContractAddress: common.BytesToAddress(rand.Bytes(common.AddressLength)).Hex(),
								Payload:         rand.Bytes(100),
							},
						}).(proto.Message)))
					}).
						Then("should not panic", func(t *testing.T) {
							assert.NotPanics(t, func() {
								keeper.Hooks(nexusK, govK).AfterProposalSubmission(ctx, proposal.ProposalId)
							})
						}),
				),
		).
		Run(t)
}
