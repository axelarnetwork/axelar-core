package keeper

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// NewVoteHandler returns the handler for processing vote delivered by the vote module
func NewVoteHandler(cdc codec.Codec, keeper types.BaseKeeper, nexus types.Nexus) vote.VoteHandler {
	return func(ctx sdk.Context, chainName string, result *vote.Vote) error {

		chain, ok := nexus.GetChain(ctx, chainName)
		if !ok {
			return fmt.Errorf("%s is not a registered chain", chainName)
		}

		if err := validateChainActivated(ctx, nexus, chain); err != nil {
			return err
		}

		events, err := types.UnpackEvents(cdc, result.Results)
		if err != nil {
			return err
		}
		var errors []error
		for _, event := range events {
			var err error

			// validate event
			err = event.Validate()
			if err != nil {
				errors = append(errors, fmt.Errorf("event %d: %s", event.Index, err.Error()))
				continue
			}

			// check if event confirmed before
			eventID := event.GetID()
			if _, ok := keeper.ForChain(chainName).GetEvent(ctx, eventID); ok {
				errors = append(errors, fmt.Errorf("event %s is already confirmed", eventID))
				continue
			}

			switch e := event.GetEvent().(type) {
			case *types.Event_Transfer:
				err = handleVoteConfirmDeposit(ctx, keeper, nexus, chain, e)
			case *types.Event_TokenDeployed:
				err = handleVoteConfirmToken(ctx, keeper, chain, e)
			default:
				err = fmt.Errorf("event %d: unsupported event type %T", event.Index, event)
			}

			if err != nil {
				errors = append(errors, fmt.Errorf("event %d: %s", event.Index, err.Error()))
				continue
			}

			keeper.ForChain(chain.Name).SetConfirmedEvent(ctx, event)
		}

		if len(errors) != 0 {
			return fmt.Errorf("failed to process events: %s", errors)
		}

		return nil
	}
}

func handleVoteConfirmDeposit(ctx sdk.Context, k types.BaseKeeper, n types.Nexus, chain nexus.Chain, event *types.Event_Transfer) error {

	keeper := k.ForChain(chain.Name)

	k.Logger(ctx).Info(fmt.Sprintf("deposit confirmation result is %s", event.Transfer.String()), "chain", chain.Name)

	// get deposit address
	burnerInfo := keeper.GetBurnerInfo(ctx, event.Transfer.To)

	depositAddr := nexus.CrossChainAddress{Chain: chain, Address: event.Transfer.To.Hex()}
	recipient, ok := n.GetRecipient(ctx, depositAddr)
	if !ok {
		return fmt.Errorf("cross-chain sender has no recipient")
	}

	amount := sdk.NewCoin(burnerInfo.Asset, sdk.NewIntFromBigInt(event.Transfer.Amount.BigInt()))
	transferID, err := n.EnqueueForTransfer(ctx, depositAddr, amount)
	if err != nil {
		return err
	}

	k.Logger(ctx).Info(fmt.Sprintf("%s deposit confirmation result to %s", chain.Name, burnerInfo.BurnerAddress))

	height, ok := keeper.GetRequiredConfirmationHeight(ctx)
	if !ok {
		return fmt.Errorf("could not find confirmation height")
	}

	// handle poll result
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeDepositConfirmation,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyChain, chain.Name),
			sdk.NewAttribute(types.AttributeKeySourceChain, chain.Name),
			sdk.NewAttribute(types.AttributeKeyConfHeight, strconv.FormatUint(height, 10)),
			sdk.NewAttribute(types.AttributeKeyDestinationChain, recipient.Chain.Name),
			sdk.NewAttribute(types.AttributeKeyDestinationAddress, recipient.Address),
			sdk.NewAttribute(types.AttributeKeyAmount, event.Transfer.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyAsset, burnerInfo.Asset),
			sdk.NewAttribute(types.AttributeKeyDepositAddress, depositAddr.Address),
			sdk.NewAttribute(types.AttributeKeyTokenAddress, burnerInfo.TokenAddress.Hex()),
			sdk.NewAttribute(types.AttributeKeyTransferID, transferID.String()),
		))

	return nil
}

func handleVoteConfirmToken(ctx sdk.Context, k types.BaseKeeper, chain nexus.Chain, event *types.Event_TokenDeployed) error {
	keeper := k.ForChain(chain.Name)
	token := keeper.GetERC20TokenBySymbol(ctx, event.TokenDeployed.Symbol)
	if token.Is(types.NonExistent) {
		return fmt.Errorf("token %s does not exist", event.TokenDeployed.Symbol)
	}

	if token.GetAddress() != event.TokenDeployed.TokenAddress {
		return fmt.Errorf("token address %s does not match expected %s", event.TokenDeployed.TokenAddress.Hex(), token.GetAddress().Hex())
	}

	if err := token.ConfirmDeployment(); err != nil {
		return err
	}

	k.Logger(ctx).Info(fmt.Sprintf("token %s deployment confirmed on chain %s", event.TokenDeployed.Symbol, chain.Name))

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeTokenConfirmation,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyChain, chain.Name),
			sdk.NewAttribute(types.AttributeKeyAsset, token.GetAsset()),
			sdk.NewAttribute(types.AttributeKeySymbol, token.GetDetails().Symbol),
			sdk.NewAttribute(types.AttributeKeyTokenAddress, token.GetAddress().Hex()),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueConfirm),
		))

	return nil
}
