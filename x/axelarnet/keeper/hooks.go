package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// Hooks defines the nexus hooks for the gov module.
type Hooks struct {
	k     Keeper
	nexus types.Nexus
	gov   govkeeper.Keeper
}

var _ govtypes.GovHooks = Hooks{}

func (k Keeper) Hooks(nexus types.Nexus, gov govkeeper.Keeper) Hooks {
	return Hooks{k, nexus, gov}
}

// AfterProposalDeposit implements govtypes.GovHooks.
func (h Hooks) AfterProposalDeposit(ctx context.Context, proposalID uint64, _ sdk.AccAddress) error {
	proposal, err := h.gov.Proposals.Get(ctx, proposalID)
	if err != nil {
		return err
	}

	contractCalls, err := h.extractContractCalls(proposal)
	if err != nil {
		return err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	totalDeposit := sdk.NewCoins(proposal.TotalDeposit...)
	minDepositsMap := h.k.GetParams(sdkCtx).CallContractsProposalMinDeposits.ToMap(sdkCtx, h.nexus)
	for _, contractCall := range contractCalls {
		minDeposit := minDepositsMap.Get(contractCall.Chain, contractCall.ContractAddress)
		if !totalDeposit.IsAllGTE(minDeposit) {
			return fmt.Errorf("proposal %d does not have enough deposits for calling contract %s on chain %s (required: %s, provided: %s)",
				proposalID, contractCall.ContractAddress, contractCall.Chain, minDeposit.String(), totalDeposit.String())
		}
	}

	return nil
}

// AfterProposalFailedMinDeposit implements govtypes.GovHooks.
func (Hooks) AfterProposalFailedMinDeposit(context.Context, uint64) error { return nil }

// AfterProposalSubmission implements govtypes.GovHooks.
func (h Hooks) AfterProposalSubmission(ctx context.Context, proposalID uint64) error {
	proposal, err := h.gov.Proposals.Get(ctx, proposalID)
	if err != nil {
		return err
	}

	contractCalls, err := h.extractContractCalls(proposal)
	if err != nil {
		return err
	}

	// perform stateful validations of the proposal
	for _, contractCall := range contractCalls {
		sdkCtx := sdk.UnwrapSDKContext(ctx)

		chain, ok := h.nexus.GetChain(sdkCtx, contractCall.Chain)
		if !ok {
			return fmt.Errorf("%s is not a registered chain", contractCall.Chain)
		}

		crossChainAddress := nexus.CrossChainAddress{Chain: chain, Address: contractCall.ContractAddress}
		if err := h.nexus.ValidateAddress(sdkCtx, crossChainAddress); err != nil {
			return err
		}
	}
	return nil
}

// AfterProposalVote implements govtypes.GovHooks.
func (Hooks) AfterProposalVote(context.Context, uint64, sdk.AccAddress) error {
	return nil
}

// AfterProposalVotingPeriodEnded implements govtypes.GovHooks.
func (Hooks) AfterProposalVotingPeriodEnded(context.Context, uint64) error {
	return nil
}

// extractContractCalls walks through the proposal messages and returns all contract calls inside.
func (h Hooks) extractContractCalls(proposal govv1.Proposal) ([]types.ContractCall, error) {
	contractCalls := make([]types.ContractCall, 0)
	for _, message := range proposal.GetMessages() {
		var sdkMsg sdk.Msg
		err := h.k.cdc.UnpackAny(message, &sdkMsg)
		if err != nil {
			return nil, err
		}

		switch sdkMsg := sdkMsg.(type) {
		case *govv1.MsgExecLegacyContent:
			// handle legacy proposal
			var legacyProposal govv1beta1.Content
			err = h.k.cdc.UnpackAny(sdkMsg.GetContent(), &legacyProposal)
			if err != nil {
				return nil, err
			}

			switch c := legacyProposal.(type) {
			case *types.CallContractsProposal:
				contractCalls = append(contractCalls, c.ContractCalls...)
			}
		case *types.CallContractRequest:
			// handle direct call contract request message
			contractCalls = append(contractCalls, types.ContractCall{
				Chain:           sdkMsg.Chain,
				ContractAddress: sdkMsg.ContractAddress,
				Payload:         sdkMsg.Payload,
			})
		default:
			// do not care about other messages
		}
	}

	return contractCalls, nil
}
