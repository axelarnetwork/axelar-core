package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/core/address"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	store "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
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
		proposal govv1.Proposal
	)

	encCfg := app.MakeEncodingConfig()
	subspace := params.NewSubspace(encCfg.Codec, encCfg.Amino, store.NewKVStoreKey("nexusKey"), store.NewKVStoreKey("tNexusKey"), "nexus")
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.NewTestLogger(t))
	nexusK := &mock.NexusMock{}
	govK := govkeeper.NewKeeper(
		encCfg.Codec,
		runtime.NewKVStoreService(store.NewKVStoreKey(govtypes.StoreKey)),
		&mock.AccountKeeperMock{
			GetModuleAddressFunc: func(moduleName string) sdk.AccAddress {
				return authtypes.NewModuleAddress(moduleName)
			},
			AddressCodecFunc: func() address.Codec {
				return authcodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix())
			},
		},
		&mock.BankKeeperMock{},
		&mock.StakingKeeperMock{},
		&mock.DistributionKeeperMock{},

		nil,
		govtypes.DefaultConfig(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	contractCall := types.ContractCall{
		Chain:           nexus.ChainName(rand.NormalizedStr(5)),
		ContractAddress: common.BytesToAddress(rand.Bytes(common.AddressLength)).Hex(),
		Payload:         rand.Bytes(100),
	}
	callContractRequest := types.CallContractRequest{
		Sender:          rand.AccAddr().String(), // should be governance account in real request
		Chain:           contractCall.Chain,
		ContractAddress: contractCall.ContractAddress,
		Payload:         contractCall.Payload,
	}
	minDeposit := sdk.NewCoin("TEST", math.NewInt(rand.PosI64()))

	keeper := keeper.NewKeeper(encCfg.Codec, store.NewKVStoreKey("nexus"), subspace, &mock.ChannelKeeperMock{}, &mock.FeegrantKeeperMock{})
	keeper.SetParams(ctx, types.DefaultParams())

	nexusK.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
		return evm.Ethereum, true
	}

	Given("a legacy proposal is created", func() {}).
		Branch(
			When("the proposal is not a nexus call contracts proposal", func() {
				proposal = funcs.Must(convertToNewProposal(funcs.Must(govv1beta1.NewProposal(
					paramstypes.NewParameterChangeProposal("title", "description", []paramstypes.ParamChange{}),
					uint64(rand.I64Between(1, 100)),
					time.Now(),
					time.Now().AddDate(0, 0, 1),
				))))
				funcs.MustNoErr(govK.SetProposal(ctx, proposal))
			}).
				Then("should not error", func(t *testing.T) {
					assert.NoError(t, keeper.Hooks(nexusK, *govK).AfterProposalDeposit(ctx, proposal.Id, rand.AccAddr()))
				}),

			When("the proposal is a nexus call contracts proposal", func() {
				proposal = funcs.Must(convertToNewProposal(funcs.Must(govv1beta1.NewProposal(
					types.NewCallContractsProposal("title", "description", []types.ContractCall{contractCall}),
					1,
					time.Now(),
					time.Now().AddDate(0, 0, 1),
				))))
				funcs.MustNoErr(govK.SetProposal(ctx, proposal))
			}).
				Branch(
					When("keeper is setup with the default params", func() {
						keeper.SetParams(ctx, types.DefaultParams())
					}).
						Then("should return no error", func(t *testing.T) {
							assert.NoError(t, keeper.Hooks(nexusK, *govK).AfterProposalDeposit(ctx, proposal.Id, rand.AccAddr()))
						}),

					When("keeper is setup with params that sets no min deposit for the contract call", func() {
						params := types.DefaultParams()
						params.CallContractsProposalMinDeposits = []types.CallContractProposalMinDeposit{
							{Chain: contractCall.Chain, ContractAddress: common.BytesToAddress(rand.Bytes(common.AddressLength)).Hex(), MinDeposits: sdk.NewCoins(minDeposit)},
							{Chain: nexus.ChainName(rand.NormalizedStr(5)), ContractAddress: contractCall.ContractAddress, MinDeposits: sdk.NewCoins(minDeposit)},
						}

						keeper.SetParams(ctx, params)
					}).
						Then("should return no error", func(t *testing.T) {
							assert.NoError(t, keeper.Hooks(nexusK, *govK).AfterProposalDeposit(ctx, proposal.Id, rand.AccAddr()))
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
								proposal.TotalDeposit = sdk.NewCoins(proposal.TotalDeposit...).Add(minDeposit)
								funcs.MustNoErr(govK.SetProposal(ctx, proposal))
							}).
								Then("should return no error", func(t *testing.T) {
									assert.NoError(t, keeper.Hooks(nexusK, *govK).AfterProposalDeposit(ctx, proposal.Id, rand.AccAddr()))
								}),

							When("min deposit is not met", func() {
								proposal.TotalDeposit = sdk.NewCoins(proposal.TotalDeposit...).Add(minDeposit.SubAmount(math.NewInt(1)))
								funcs.MustNoErr(govK.SetProposal(ctx, proposal))
							}).
								Then("should return error", func(t *testing.T) {
									assert.ErrorContains(t,
										keeper.Hooks(nexusK, *govK).AfterProposalDeposit(ctx, proposal.Id, rand.AccAddr()),
										fmt.Sprintf("proposal %d does not have enough deposits for calling contract %s on chain %s (required: %s, provided: %s)",
											proposal.Id, contractCall.ContractAddress, contractCall.Chain, minDeposit.String(), sdk.NewCoins(proposal.TotalDeposit...).String(),
										),
									)
								}),
						),
				),
		).
		Run(t)

	Given("a v1 proposal is created", func() {}).
		Branch(
			When("the proposal does not contain a nexus contract call", func() {
				proposal = funcs.Must(govv1.NewProposal(
					[]sdk.Msg{}, // TODO: actually put msg inside
					1,
					time.Now(),
					time.Now().AddDate(0, 0, 1),
					"metadata", "title", "description",
					rand.AccAddr(), false,
				))
				funcs.MustNoErr(govK.SetProposal(ctx, proposal))
			}).
				Then("should not error", func(t *testing.T) {
					assert.NoError(t, keeper.Hooks(nexusK, *govK).AfterProposalDeposit(ctx, proposal.Id, rand.AccAddr()))
				}),

			When("the proposal contains a nexus contract call", func() {
				proposal = funcs.Must(govv1.NewProposal(
					[]sdk.Msg{&callContractRequest},
					1,
					time.Now(),
					time.Now().AddDate(0, 0, 1),
					"metadata", "title", "description",
					rand.AccAddr(), false,
				))
				funcs.MustNoErr(govK.SetProposal(ctx, proposal))
			}).
				Branch(
					When("keeper is setup with the default params", func() {
						keeper.SetParams(ctx, types.DefaultParams())
					}).
						Then("should return no error", func(t *testing.T) {
							assert.NoError(t, keeper.Hooks(nexusK, *govK).AfterProposalDeposit(ctx, proposal.Id, rand.AccAddr()))
						}),

					When("keeper is setup with params that sets no min deposit for the contract call", func() {
						params := types.DefaultParams()
						params.CallContractsProposalMinDeposits = []types.CallContractProposalMinDeposit{
							{Chain: callContractRequest.Chain, ContractAddress: common.BytesToAddress(rand.Bytes(common.AddressLength)).Hex(), MinDeposits: sdk.NewCoins(minDeposit)},
							{Chain: nexus.ChainName(rand.NormalizedStr(5)), ContractAddress: callContractRequest.ContractAddress, MinDeposits: sdk.NewCoins(minDeposit)},
						}

						keeper.SetParams(ctx, params)
					}).
						Then("should return no error", func(t *testing.T) {
							assert.NoError(t, keeper.Hooks(nexusK, *govK).AfterProposalDeposit(ctx, proposal.Id, rand.AccAddr()))
						}),

					When("keeper is setup with params that sets some min deposit for the contract call", func() {
						params := types.DefaultParams()
						params.CallContractsProposalMinDeposits = []types.CallContractProposalMinDeposit{
							{Chain: callContractRequest.Chain, ContractAddress: callContractRequest.ContractAddress, MinDeposits: sdk.NewCoins(minDeposit)},
						}

						keeper.SetParams(ctx, params)
					}).
						Branch(
							When("min deposit is met", func() {
								proposal.TotalDeposit = sdk.NewCoins(proposal.TotalDeposit...).Add(minDeposit)
								funcs.MustNoErr(govK.SetProposal(ctx, proposal))
							}).
								Then("should return no error", func(t *testing.T) {
									assert.NoError(t, keeper.Hooks(nexusK, *govK).AfterProposalDeposit(ctx, proposal.Id, rand.AccAddr()))
								}),

							When("min deposit is not met", func() {
								proposal.TotalDeposit = sdk.NewCoins(proposal.TotalDeposit...).Add(minDeposit.SubAmount(math.NewInt(1)))
								funcs.MustNoErr(govK.SetProposal(ctx, proposal))
							}).
								Then("should return error", func(t *testing.T) {
									assert.ErrorContains(t,
										keeper.Hooks(nexusK, *govK).AfterProposalDeposit(ctx, proposal.Id, rand.AccAddr()),
										fmt.Sprintf("proposal %d does not have enough deposits for calling contract %s on chain %s (required: %s, provided: %s)",
											proposal.Id, callContractRequest.ContractAddress, callContractRequest.Chain, minDeposit.String(), sdk.NewCoins(proposal.TotalDeposit...).String(),
										),
									)
								}),
						),
				),
		).Run(t)
}

func TestAfterProposalSubmission(t *testing.T) {
	var (
		legacyProposal govv1beta1.Proposal
		proposal       govv1.Proposal
	)

	encCfg := app.MakeEncodingConfig()
	subspace := params.NewSubspace(encCfg.Codec, encCfg.Amino, store.NewKVStoreKey("nexusKey"), store.NewKVStoreKey("tNexusKey"), "nexus")
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.NewTestLogger(t))
	nexusK := &mock.NexusMock{
		GetChainFunc: func(ctx sdk.Context, chainName nexus.ChainName) (nexus.Chain, bool) {
			return evm.Ethereum, chainName == evm.Ethereum.GetName()
		},
		ValidateAddressFunc: func(ctx sdk.Context, address nexus.CrossChainAddress) error {
			return evmkeeper.NewAddressValidator()(ctx, address)
		},
	}
	govK := govkeeper.NewKeeper(
		encCfg.Codec,
		runtime.NewKVStoreService(store.NewKVStoreKey(govtypes.StoreKey)),
		&mock.AccountKeeperMock{
			GetModuleAddressFunc: func(moduleName string) sdk.AccAddress {
				return authtypes.NewModuleAddress(moduleName)
			},
			AddressCodecFunc: func() address.Codec {
				return authcodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix())
			},
		},
		&mock.BankKeeperMock{},
		&mock.StakingKeeperMock{},
		&mock.DistributionKeeperMock{},

		nil,
		govtypes.DefaultConfig(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	nonRegisteredChain := nexus.ChainName(rand.NormalizedStr(5))

	keeper := keeper.NewKeeper(encCfg.Codec, store.NewKVStoreKey("nexus"), subspace, &mock.ChannelKeeperMock{}, &mock.FeegrantKeeperMock{})
	keeper.SetParams(ctx, types.DefaultParams())

	Given("a legacy proposal is created", func() {}).
		Branch(
			When("the proposal is not a nexus call contracts proposal", func() {
				legacyProposal = funcs.Must(govv1beta1.NewProposal(
					paramstypes.NewParameterChangeProposal("title", "description", []paramstypes.ParamChange{}),
					uint64(rand.I64Between(1, 100)),
					time.Now(),
					time.Now().AddDate(0, 0, 1),
				))
				funcs.MustNoErr(govK.SetProposal(ctx, funcs.Must(convertToNewProposal(legacyProposal))))
			}).
				Then("should return no error", func(t *testing.T) {
					assert.NoError(t, keeper.Hooks(nexusK, *govK).AfterProposalSubmission(ctx, legacyProposal.ProposalId))
				}),

			When("the proposal is a nexus call contracts proposal", func() {
				legacyProposal = funcs.Must(govv1beta1.NewProposal(
					types.NewCallContractsProposal("title", "description", []types.ContractCall{}),
					1,
					time.Now(),
					time.Now().AddDate(0, 0, 1),
				))
				funcs.MustNoErr(govK.SetProposal(ctx, funcs.Must(convertToNewProposal(legacyProposal))))
			}).
				Branch(
					When("contract call has non-registered chain", func() {
						legacyProposal.Content = funcs.Must(codectypes.NewAnyWithValue(types.NewCallContractsProposal("title", "description", []types.ContractCall{
							{
								Chain:           nonRegisteredChain,
								ContractAddress: common.BytesToAddress(rand.Bytes(common.AddressLength)).Hex(),
								Payload:         rand.Bytes(100),
							},
						}).(proto.Message)))
						funcs.MustNoErr(govK.SetProposal(ctx, funcs.Must(convertToNewProposal(legacyProposal))))
					}).
						Then("should return error", func(t *testing.T) {
							assert.ErrorContains(t,
								keeper.Hooks(nexusK, *govK).AfterProposalSubmission(ctx, legacyProposal.ProposalId),
								fmt.Sprintf("%s is not a registered chain", nonRegisteredChain),
							)
						}),

					When("contract call has registered chain but invalid contract address", func() {
						legacyProposal.Content = funcs.Must(codectypes.NewAnyWithValue(types.NewCallContractsProposal("title", "description", []types.ContractCall{
							{
								Chain:           evm.Ethereum.Name,
								ContractAddress: rand.NormalizedStr(42),
								Payload:         rand.Bytes(100),
							},
						}).(proto.Message)))
						funcs.MustNoErr(govK.SetProposal(ctx, funcs.Must(convertToNewProposal(legacyProposal))))
					}).
						Then("should return error", func(t *testing.T) {
							assert.ErrorContains(t, keeper.Hooks(nexusK, *govK).AfterProposalSubmission(ctx, legacyProposal.ProposalId),
								"not an hex address")
						}),

					When("contract call has registered chain and valid contract address", func() {
						legacyProposal.Content = funcs.Must(codectypes.NewAnyWithValue(types.NewCallContractsProposal("title", "description", []types.ContractCall{
							{
								Chain:           evm.Ethereum.Name,
								ContractAddress: common.BytesToAddress(rand.Bytes(common.AddressLength)).Hex(),
								Payload:         rand.Bytes(100),
							},
						}).(proto.Message)))
						funcs.MustNoErr(govK.SetProposal(ctx, funcs.Must(convertToNewProposal(legacyProposal))))
					}).
						Then("should return no error", func(t *testing.T) {
							assert.NoError(t, keeper.Hooks(nexusK, *govK).AfterProposalSubmission(ctx, legacyProposal.ProposalId))
						}),
				),
		).
		Run(t)

	Given("a v1 proposal is created", func() {}).
		Branch(
			When("the proposal does not contain a nexus contract call", func() {
				proposal = funcs.Must(govv1.NewProposal(
					[]sdk.Msg{}, // TODO: actually put msg inside
					uint64(rand.I64Between(1, 100)),
					time.Now(),
					time.Now().AddDate(0, 0, 1),
					"metadata", "title", "description",
					rand.AccAddr(), false,
				))
				funcs.MustNoErr(govK.SetProposal(ctx, proposal))
			}).
				Then("should return no error", func(t *testing.T) {
					assert.NoError(t, keeper.Hooks(nexusK, *govK).AfterProposalSubmission(ctx, proposal.Id))
				}),

			When("the proposal contains a nexus contract call", func() {
				proposal = funcs.Must(govv1.NewProposal(
					[]sdk.Msg{types.NewCallContractRequest(
						rand.AccAddr(),
						evm.Ethereum.Name.String(),
						common.BytesToAddress(rand.Bytes(common.AddressLength)).Hex(),
						rand.Bytes(100),
						nil,
					)},
					1,
					time.Now(),
					time.Now().AddDate(0, 0, 1),
					"metadata", "title", "description",
					rand.AccAddr(), false,
				))
				funcs.MustNoErr(govK.SetProposal(ctx, proposal))
			}).
				Branch(
					When("contract call has non-registered chain", func() {
						proposal.Messages[0] = funcs.Must(codectypes.NewAnyWithValue(types.NewCallContractRequest(
							rand.AccAddr(),
							nonRegisteredChain.String(),
							common.BytesToAddress(rand.Bytes(common.AddressLength)).Hex(),
							rand.Bytes(100),
							nil,
						)))
						funcs.MustNoErr(govK.SetProposal(ctx, proposal))
					}).
						Then("should return error", func(t *testing.T) {
							assert.ErrorContains(t,
								keeper.Hooks(nexusK, *govK).AfterProposalSubmission(ctx, proposal.Id),
								fmt.Sprintf("%s is not a registered chain", nonRegisteredChain),
							)
						}),

					When("contract call has registered chain but invalid contract address", func() {
						proposal.Messages[0] = funcs.Must(codectypes.NewAnyWithValue(types.NewCallContractRequest(
							rand.AccAddr(),
							evm.Ethereum.Name.String(),
							rand.NormalizedStr(42),
							rand.Bytes(100),
							nil,
						)))
						funcs.MustNoErr(govK.SetProposal(ctx, proposal))
					}).
						Then("should return error", func(t *testing.T) {
							assert.ErrorContains(t, keeper.Hooks(nexusK, *govK).AfterProposalSubmission(ctx, proposal.Id),
								"not an hex address")
						}),

					When("contract call has registered chain and valid contract address", func() {
						proposal.Messages[0] = funcs.Must(codectypes.NewAnyWithValue(types.NewCallContractRequest(
							rand.AccAddr(),
							evm.Ethereum.Name.String(),
							common.BytesToAddress(rand.Bytes(common.AddressLength)).Hex(),
							rand.Bytes(100),
							nil,
						)))
						funcs.MustNoErr(govK.SetProposal(ctx, proposal))
					}).
						Then("should return no error", func(t *testing.T) {
							assert.NoError(t, keeper.Hooks(nexusK, *govK).AfterProposalSubmission(ctx, proposal.Id))
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
