package keeper

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/axelarnetwork/axelar-core/x/nexus/types"
	"github.com/axelarnetwork/utils/funcs"
)

// Hooks defines the nexus hooks for the gov module.
type Hooks struct {
	k   types.Nexus
	gov types.GovKeeper
}

var _ govtypes.GovHooks = Hooks{}

func (k Keeper) Hooks(gov types.GovKeeper) Hooks {
	return Hooks{k, gov}
}

// AfterProposalDeposit implements govtypes.GovHooks.
func (h Hooks) AfterProposalDeposit(ctx sdk.Context, proposalID uint64, depositorAddr sdk.AccAddress) {
	proposal := funcs.MustOk(h.gov.GetProposal(ctx, proposalID))

	switch c := proposal.GetContent().(type) {
	case *types.CallContractsProposal:
		minDeposits := h.k.GetParams(ctx).CallContractsProposalMinDeposits

		for _, contractCall := range c.ContractCalls {
			contractAddress := strings.ToLower(contractCall.ContractAddress)

			minDeposit, ok := minDeposits[contractAddress]
			if !ok {
				continue
			}

			if !proposal.TotalDeposit.IsAllGTE(minDeposit.Coins) {
				panic(fmt.Errorf("proposal %d does not have enough deposits for calling contract %s", proposalID, contractAddress))
			}
		}
	default:
		return
	}
}

// AfterProposalFailedMinDeposit implements govtypes.GovHooks.
func (Hooks) AfterProposalFailedMinDeposit(ctx sdk.Context, proposalID uint64) {}

// AfterProposalSubmission implements govtypes.GovHooks.
func (Hooks) AfterProposalSubmission(ctx sdk.Context, proposalID uint64) {}

// AfterProposalVote implements govtypes.GovHooks.
func (Hooks) AfterProposalVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress) {}

// AfterProposalVotingPeriodEnded implements govtypes.GovHooks.
func (Hooks) AfterProposalVotingPeriodEnded(ctx sdk.Context, proposalID uint64) {}
