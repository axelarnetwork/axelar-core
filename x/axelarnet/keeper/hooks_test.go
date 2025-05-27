package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/cosmos/gogoproto/proto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

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
		proposal govv1beta1.Proposal
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
		govK.GetProposalFunc = func(ctx sdk.Context, proposalID uint64) (govv1.Proposal, bool) {
			return funcs.Must(convertToNewProposal(proposal)), proposalID == proposal.ProposalId
		}
	}).
		Branch(
			When("the proposal is not a nexus call contracts proposal", func() {
				proposal = funcs.Must(govv1beta1.NewProposal(
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
				proposal = funcs.Must(govv1beta1.NewProposal(
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
		proposal govv1beta1.Proposal
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
		govK.GetProposalFunc = func(ctx sdk.Context, proposalID uint64) (govv1.Proposal, bool) {
			return funcs.Must(convertToNewProposal(proposal)), proposalID == proposal.ProposalId
		}
	}).
		Branch(
			When("the proposal is not a nexus call contracts proposal", func() {
				proposal = funcs.Must(govv1beta1.NewProposal(
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
				proposal = funcs.Must(govv1beta1.NewProposal(
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

func convertToNewProposal(oldProp govv1beta1.Proposal) (govv1.Proposal, error) {
	msg, err := govv1.NewLegacyContent(oldProp.GetContent(), authtypes.NewModuleAddress(types.ModuleName).String())
	if err != nil {
		return govv1.Proposal{}, err
	}
	msgAny, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return govv1.Proposal{}, err
	}

	return govv1.Proposal{
		Id:       oldProp.ProposalId,
		Messages: []*codectypes.Any{msgAny},
		Status:   govv1.ProposalStatus(oldProp.Status),
		FinalTallyResult: &govv1.TallyResult{
			YesCount:        oldProp.FinalTallyResult.Yes.String(),
			NoCount:         oldProp.FinalTallyResult.No.String(),
			AbstainCount:    oldProp.FinalTallyResult.Abstain.String(),
			NoWithVetoCount: oldProp.FinalTallyResult.NoWithVeto.String(),
		},
		SubmitTime:      &oldProp.SubmitTime,
		DepositEndTime:  &oldProp.DepositEndTime,
		TotalDeposit:    oldProp.TotalDeposit,
		VotingStartTime: &oldProp.VotingStartTime,
		VotingEndTime:   &oldProp.VotingEndTime,
		Title:           oldProp.GetContent().GetTitle(),
		Summary:         oldProp.GetContent().GetDescription(),
	}, nil
}
