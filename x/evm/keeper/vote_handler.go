package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/utils/slices"

	"github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// NewVoteHandler returns the handler for processing vote delivered by the vote module
func NewVoteHandler(cdc codec.Codec, keeper types.BaseKeeper, nexus types.Nexus) vote.VoteHandler {
	return func(ctx sdk.Context, result *vote.Vote) error {
		events, err := types.UnpackEvents(cdc, result.Results)
		if err != nil {
			return err
		}

		if len(events) == 0 {
			return nil
		}

		chainName := events[0].Chain
		if slices.Any(events, func(event types.Event) bool { return event.Chain != chainName }) {
			return fmt.Errorf("events are not from the same source chain")
		}

		chain, ok := nexus.GetChain(ctx, chainName)
		if !ok {
			return fmt.Errorf("%s is not a registered chain", chainName)
		}

		var errors []error
		for _, event := range events {
			var err error

			// validate event
			err = event.ValidateBasic()
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

			switch event.GetEvent().(type) {
			case *types.Event_Transfer:
				err = handleVoteConfirmDeposit(ctx, keeper, nexus, chain, event)
			case *types.Event_TokenDeployed:
				err = handleVoteConfirmToken(ctx, keeper, chain, event)
			default:
				err = fmt.Errorf("event %s: unsupported event type %T", eventID, event)
			}

			if err != nil {
				errors = append(errors, fmt.Errorf("event %s: %s", eventID, err.Error()))
				continue
			}
			// set event complete
			keeper.ForChain(chainName).SetConfirmedEvent(ctx, event)
			keeper.ForChain(chainName).SetEventCompleted(ctx, eventID)
		}

		if len(errors) != 0 {
			for _, err := range errors {
				keeper.Logger(ctx).Error(fmt.Sprintf("failed to process event: %s", err.Error()))
			}
			return fmt.Errorf("failed to process events: %s", errors)
		}

		return nil
	}
}

func handleVoteConfirmDeposit(ctx sdk.Context, k types.BaseKeeper, n types.Nexus, chain nexus.Chain, event types.Event) error {

	keeper := k.ForChain(chain.Name)
	transferEvent := event.GetEvent().(*types.Event_Transfer)

	k.Logger(ctx).Info(fmt.Sprintf("deposit confirmation result is %s", transferEvent.Transfer.String()), "chain", chain.Name)

	// get deposit address
	burnerInfo := keeper.GetBurnerInfo(ctx, transferEvent.Transfer.To)
	if burnerInfo == nil {
		return fmt.Errorf("no burner info found for address %s", transferEvent.Transfer.To.Hex())
	}

	depositAddr := nexus.CrossChainAddress{Chain: chain, Address: transferEvent.Transfer.To.Hex()}
	recipient, ok := n.GetRecipient(ctx, depositAddr)
	if !ok {
		return fmt.Errorf("cross-chain sender has no recipient")
	}

	amount := sdk.NewCoin(burnerInfo.Asset, sdk.NewIntFromBigInt(transferEvent.Transfer.Amount.BigInt()))
	transferID, err := n.EnqueueForTransfer(ctx, depositAddr, amount)
	if err != nil {
		return err
	}

	k.Logger(ctx).Info(fmt.Sprintf("%s deposit confirmation result to %s", chain.Name, burnerInfo.BurnerAddress))

	// handle poll result
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeDepositConfirmation,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyChain, chain.Name),
			sdk.NewAttribute(types.AttributeKeySourceChain, chain.Name),
			sdk.NewAttribute(types.AttributeKeyDestinationChain, recipient.Chain.Name),
			sdk.NewAttribute(types.AttributeKeyDestinationAddress, recipient.Address),
			sdk.NewAttribute(types.AttributeKeyAmount, transferEvent.Transfer.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyAsset, burnerInfo.Asset),
			sdk.NewAttribute(types.AttributeKeyDepositAddress, depositAddr.Address),
			sdk.NewAttribute(types.AttributeKeyTokenAddress, burnerInfo.TokenAddress.Hex()),
			sdk.NewAttribute(types.AttributeKeyTxID, event.TxId.Hex()),
			sdk.NewAttribute(types.AttributeKeyTransferID, transferID.String()),
		))

	return nil
}

func handleVoteConfirmToken(ctx sdk.Context, k types.BaseKeeper, chain nexus.Chain, event types.Event) error {
	keeper := k.ForChain(chain.Name)
	tokenDeployedEvent := event.GetEvent().(*types.Event_TokenDeployed)

	token := keeper.GetERC20TokenBySymbol(ctx, tokenDeployedEvent.TokenDeployed.Symbol)
	if token.Is(types.NonExistent) {
		return fmt.Errorf("token %s does not exist", tokenDeployedEvent.TokenDeployed.Symbol)
	}

	if token.GetAddress() != tokenDeployedEvent.TokenDeployed.TokenAddress {
		return fmt.Errorf("token address %s does not match expected %s", tokenDeployedEvent.TokenDeployed.TokenAddress.Hex(), token.GetAddress().Hex())
	}

	if err := token.ConfirmDeployment(); err != nil {
		return err
	}

	k.Logger(ctx).Info(fmt.Sprintf("token %s deployment confirmed on chain %s", tokenDeployedEvent.TokenDeployed.Symbol, chain.Name))

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeTokenConfirmation,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyChain, chain.Name),
			sdk.NewAttribute(types.AttributeKeyAsset, token.GetAsset()),
			sdk.NewAttribute(types.AttributeKeySymbol, token.GetDetails().Symbol),
			sdk.NewAttribute(types.AttributeKeyTokenAddress, token.GetAddress().Hex()),
			sdk.NewAttribute(types.AttributeKeyTxID, event.TxId.Hex()),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueConfirm),
		))

	return nil
}
