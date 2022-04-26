package keeper

import (
	"fmt"
	"sort"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/axelarnetwork/utils/slices"

	"github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// NewVoteHandler returns the handler for processing vote delivered by the vote module
func NewVoteHandler(cdc codec.Codec, keeper types.BaseKeeper, nexus types.Nexus, signer types.Signer) vote.VoteHandler {
	return func(ctx sdk.Context, result *vote.Vote) error {
		events, err := types.UnpackEvents(cdc, result.Results)
		if err != nil {
			return err
		}

		if len(events) == 0 {
			return fmt.Errorf("vote result has no events")
		}

		chainName := events[0].Chain
		if slices.Any(events, func(event types.Event) bool { return event.Chain != chainName }) {
			return fmt.Errorf("events are not from the same source chain")
		}

		chain, ok := nexus.GetChain(ctx, chainName)
		if !ok {
			return fmt.Errorf("%s is not a registered chain", chainName)
		}
		if !keeper.HasChain(ctx, chainName) {
			return fmt.Errorf("%s is not an evm chain", chainName)
		}

		chainK := keeper.ForChain(chain.Name)
		cacheCtx, writeCache := ctx.CacheContext()

		err = handleEvents(cacheCtx, chainK, nexus, signer, events, chain)
		if err != nil {
			// set events to failed, we will deal with later
			for _, e := range events {
				chainK.SetFailedEvent(ctx, e)
			}
			return err
		}

		writeCache()
		ctx.EventManager().EmitEvents(cacheCtx.EventManager().Events())
		return nil
	}
}

func handleEvents(ctx sdk.Context, ck types.ChainKeeper, nexus types.Nexus, signer types.Signer, events []types.Event, chain nexus.Chain) error {
	for _, event := range events {
		var err error
		// validate event
		err = event.ValidateBasic()
		if err != nil {
			return fmt.Errorf("event %s: %s", event.GetID(), err.Error())
		}

		// check if event confirmed before
		eventID := event.GetID()
		if _, ok := ck.GetEvent(ctx, eventID); ok {
			return fmt.Errorf("event %s is already confirmed", eventID)
		}
		ck.SetConfirmedEvent(ctx, event)

		switch event.GetEvent().(type) {
		case *types.Event_Transfer:
			err = handleVoteConfirmDeposit(ctx, ck, nexus, chain, event)
		case *types.Event_TokenDeployed:
			err = handleVoteConfirmToken(ctx, ck, chain, event)
		case *types.Event_MultisigOwnershipTransferred, *types.Event_MultisigOperatorshipTransferred:
			err = handleVoteConfirmMultisigTransferKey(ctx, ck, signer, chain, event)
		case *types.Event_ContractCall, *types.Event_ContractCallWithToken, *types.Event_TokenSent:
			err = handleVoteConfirmGatewayTx(ctx, ck, event)
		default:
			err = fmt.Errorf("event %s: unsupported event type %T", eventID, event)
		}

		if err != nil {
			return fmt.Errorf("event %s: %s", eventID, err.Error())
		}
	}

	return nil
}

func handleVoteConfirmDeposit(ctx sdk.Context, keeper types.ChainKeeper, n types.Nexus, chain nexus.Chain, event types.Event) error {
	transferEvent := event.GetEvent().(*types.Event_Transfer)

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

	// set confirmed deposit
	erc20Deposit := types.ERC20Deposit{
		TxID:             event.TxId,
		Amount:           transferEvent.Transfer.Amount,
		Asset:            burnerInfo.Asset,
		DestinationChain: burnerInfo.DestinationChain,
		BurnerAddress:    burnerInfo.BurnerAddress,
	}
	keeper.SetDeposit(ctx, erc20Deposit, types.DepositStatus_Confirmed)

	keeper.SetEventCompleted(ctx, event.GetID())
	keeper.Logger(ctx).Info(fmt.Sprintf("deposit confirmation result to %s %s", transferEvent.Transfer.To.Hex(), transferEvent.Transfer.Amount), "chain", chain.Name)

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
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueConfirm),
		))

	return nil
}

func handleVoteConfirmToken(ctx sdk.Context, keeper types.ChainKeeper, chain nexus.Chain, event types.Event) error {
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

	keeper.SetEventCompleted(ctx, event.GetID())
	keeper.Logger(ctx).Info(fmt.Sprintf("token %s deployment confirmed on chain %s", tokenDeployedEvent.TokenDeployed.Symbol, chain.Name))

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

func handleVoteConfirmMultisigTransferKey(ctx sdk.Context, keeper types.ChainKeeper, s types.Signer, chain nexus.Chain, event types.Event) error {
	var newAddresses []types.Address
	var newThreshold sdk.Uint
	var keyRole tss.KeyRole

	switch e := event.GetEvent().(type) {
	case *types.Event_MultisigOwnershipTransferred:
		newAddresses = e.MultisigOwnershipTransferred.NewOwners
		newThreshold = e.MultisigOwnershipTransferred.NewThreshold
		keyRole = tss.MasterKey
	case *types.Event_MultisigOperatorshipTransferred:
		newAddresses = e.MultisigOperatorshipTransferred.NewOperators
		newThreshold = e.MultisigOperatorshipTransferred.NewThreshold
		keyRole = tss.SecondaryKey
	default:
		return fmt.Errorf("event %s: unsupported event type %T", event.GetID(), event)
	}

	nextKeyID, ok := s.GetNextKeyID(ctx, chain, keyRole)
	if !ok {
		return fmt.Errorf("next %s key for chain %s not found", keyRole.SimpleString(), chain.Name)
	}

	nextKey, found := s.GetKey(ctx, nextKeyID)
	if !found {
		return fmt.Errorf("key %s not found", nextKeyID)
	}

	expectedAddress, expectedThreshold, err := getMultisigAddresses(nextKey)
	if err != nil {
		return err
	}

	newOwners := slices.Map(newAddresses, func(addr types.Address) common.Address { return common.Address(addr) })
	if !areAddressesEqual(expectedAddress, newOwners) {
		return fmt.Errorf("new adddress does not match, expected %v got %v", expectedAddress, newOwners)
	}

	if !sdk.NewUint(uint64(expectedThreshold)).Equal(newThreshold) {
		return fmt.Errorf("new threshold does not match, expected %d got %d", expectedThreshold, newThreshold.Uint64())
	}

	if err := s.RotateKey(ctx, chain, keyRole); err != nil {
		return err
	}

	keeper.SetEventCompleted(ctx, event.GetID())
	keeper.Logger(ctx).Info(fmt.Sprintf("successfully confirmed %s key transfer for chain %s",
		keyRole.SimpleString(), chain.Name), "txID", event.TxId.Hex(), "rotation count", s.GetRotationCount(ctx, chain, keyRole))

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeTransferKeyConfirmation,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(types.AttributeKeyChain, chain.Name),
		sdk.NewAttribute(types.AttributeKeyTransferKeyType, keyRole.SimpleString()),
	))

	return nil
}

func handleVoteConfirmGatewayTx(ctx sdk.Context, k types.ChainKeeper, event types.Event) error {
	k.Logger(ctx).Info(fmt.Sprintf("gateway transaction confirmation confirmation result is %s", event.String()))

	// emit gatewayTxConfirmation event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeGatewayTxConfirmation,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyChain, event.Chain),
			sdk.NewAttribute(types.AttributeKeyTxID, event.TxId.Hex()),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueConfirm)),
	)

	return nil
}

func areAddressesEqual(addressesA, addressesB []common.Address) bool {
	if len(addressesA) != len(addressesB) {
		return false
	}

	addressesToHex := func(addr common.Address) string { return addr.Hex() }

	hexesA := slices.Map(addressesA, addressesToHex)
	sort.Strings(hexesA)
	hexesB := slices.Map(addressesB, addressesToHex)
	sort.Strings(hexesB)

	for i, hexA := range hexesA {
		if hexA != hexesB[i] {
			return false
		}
	}

	return true
}
