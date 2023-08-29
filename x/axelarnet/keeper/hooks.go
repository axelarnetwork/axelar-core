package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/funcs"
)

// Hooks defines the nexus hooks for the gov module.
type Hooks struct {
	k     Keeper
	nexus types.Nexus
	gov   types.GovKeeper
}

var _ govtypes.GovHooks = Hooks{}

func (k Keeper) Hooks(nexus types.Nexus, gov types.GovKeeper) Hooks {
	return Hooks{k, nexus, gov}
}

// AfterProposalDeposit implements govtypes.GovHooks.
func (h Hooks) AfterProposalDeposit(ctx sdk.Context, proposalID uint64, _ sdk.AccAddress) {
	proposal := funcs.MustOk(h.gov.GetProposal(ctx, proposalID))

	switch c := proposal.GetContent().(type) {
	case *types.CallContractsProposal:
		minDepositsMap := h.k.GetParams(ctx).CallContractsProposalMinDeposits.ToMap(ctx, h.nexus)

		for _, contractCall := range c.ContractCalls {
			minDeposit := minDepositsMap.Get(contractCall.Chain, contractCall.ContractAddress)
			if !proposal.TotalDeposit.IsAllGTE(minDeposit) {
				panic(fmt.Errorf("proposal %d does not have enough deposits for calling contract %s on chain %s (required: %s, provided: %s)",
					proposalID, contractCall.ContractAddress, contractCall.Chain, minDeposit.String(), proposal.TotalDeposit.String()))
			}
		}
	default:
		return
	}
}

// AfterProposalFailedMinDeposit implements govtypes.GovHooks.
func (Hooks) AfterProposalFailedMinDeposit(ctx sdk.Context, proposalID uint64) {}

// AfterProposalSubmission implements govtypes.GovHooks.
func (h Hooks) AfterProposalSubmission(ctx sdk.Context, proposalID uint64) {
	proposal := funcs.MustOk(h.gov.GetProposal(ctx, proposalID))

	switch c := proposal.GetContent().(type) {
	case *types.CallContractsProposal:
		// perform stateful validations of the proposal
		for _, contractCall := range c.ContractCalls {
			chain, ok := h.nexus.GetChain(ctx, contractCall.Chain)
			if !ok {
				panic(fmt.Errorf("%s is not a registered chain", contractCall.Chain))
			}

			crossChainAddress := nexus.CrossChainAddress{Chain: chain, Address: contractCall.ContractAddress}
			if err := h.nexus.ValidateAddress(ctx, crossChainAddress); err != nil {
				panic(err)
			}
		}
	default:
		return
	}
}

// AfterProposalVote implements govtypes.GovHooks.
func (Hooks) AfterProposalVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress) {}

// AfterProposalVotingPeriodEnded implements govtypes.GovHooks.
func (Hooks) AfterProposalVotingPeriodEnded(ctx sdk.Context, proposalID uint64) {}
