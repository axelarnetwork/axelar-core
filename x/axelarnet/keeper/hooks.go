package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govv3 "github.com/cosmos/cosmos-sdk/x/gov/migrations/v3"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

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

	legacyProposal, err := govv3.ConvertToLegacyProposal(proposal)
	if err != nil {
		return err
	}

	switch c := legacyProposal.GetContent().(type) {
	case *types.CallContractsProposal:
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		minDepositsMap := h.k.GetParams(sdkCtx).CallContractsProposalMinDeposits.ToMap(sdkCtx, h.nexus)

		for _, contractCall := range c.ContractCalls {
			minDeposit := minDepositsMap.Get(contractCall.Chain, contractCall.ContractAddress)
			if !legacyProposal.TotalDeposit.IsAllGTE(minDeposit) {
				return fmt.Errorf("proposal %d does not have enough deposits for calling contract %s on chain %s (required: %s, provided: %s)",
					proposalID, contractCall.ContractAddress, contractCall.Chain, minDeposit.String(), legacyProposal.TotalDeposit.String())
			}
		}
		return nil
	default:
		return nil
	}
}

// AfterProposalFailedMinDeposit implements govtypes.GovHooks.
func (Hooks) AfterProposalFailedMinDeposit(context.Context, uint64) error { return nil }

// AfterProposalSubmission implements govtypes.GovHooks.
func (h Hooks) AfterProposalSubmission(ctx context.Context, proposalID uint64) error {
	proposal, err := h.gov.Proposals.Get(ctx, proposalID)
	if err != nil {
		return err
	}

	legacyProposal, err := govv3.ConvertToLegacyProposal(proposal)
	if err != nil {
		return err
	}

	switch c := legacyProposal.GetContent().(type) {
	case *types.CallContractsProposal:
		// perform stateful validations of the proposal
		for _, contractCall := range c.ContractCalls {
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
	default:
		return nil
	}
}

// AfterProposalVote implements govtypes.GovHooks.
func (Hooks) AfterProposalVote(context.Context, uint64, sdk.AccAddress) error {
	return nil
}

// AfterProposalVotingPeriodEnded implements govtypes.GovHooks.
func (Hooks) AfterProposalVotingPeriodEnded(context.Context, uint64) error {
	return nil
}
