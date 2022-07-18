package evm

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
)

// TODO: make this a param when we can easily switch between different kinds of keys and different settings
var keyRole = tss.SecondaryKey

func validateChains(ctx sdk.Context, sourceChainName nexus.ChainName, destinationChainName nexus.ChainName, bk types.BaseKeeper, n types.Nexus) (nexus.Chain, nexus.Chain, error) {
	sourceChain, ok := n.GetChain(ctx, sourceChainName)
	if !ok {
		panic(fmt.Errorf("%s is not a registered chain", sourceChainName))
	}

	destinationChain, ok := n.GetChain(ctx, destinationChainName)
	if !ok {
		return nexus.Chain{}, nexus.Chain{}, fmt.Errorf("%s is not a registered chain", destinationChainName)
	}

	return sourceChain, destinationChain, nil
}

func handleTokenSent(ctx sdk.Context, event types.Event, bk types.BaseKeeper, n types.Nexus) bool {
	e := event.GetEvent().(*types.Event_TokenSent).TokenSent
	if e == nil {
		panic(fmt.Errorf("event is nil"))
	}

	sourceChain, destinationChain, err := validateChains(ctx, event.Chain, e.DestinationChain, bk, n)
	if err != nil {
		bk.Logger(ctx).Info(err.Error())
		return false
	}

	sourceCk := bk.ForChain(sourceChain.Name)

	token := sourceCk.GetERC20TokenBySymbol(ctx, e.Symbol)
	if !token.Is(types.Confirmed) {
		bk.Logger(ctx).Info(fmt.Sprintf("%s token %s is not confirmed yet", event.Chain, e.Symbol))
		return false
	}

	asset := token.GetAsset()

	// check erc20 token status if destination is an evm chain
	if bk.HasChain(ctx, destinationChain.Name) {
		destinationCk := bk.ForChain(destinationChain.Name)

		if token := destinationCk.GetERC20TokenByAsset(ctx, asset); !token.Is(types.Confirmed) {
			bk.Logger(ctx).Info(fmt.Sprintf("%s token with asset %s is not confirmed yet", e.DestinationChain, asset))
			return false
		}
	}

	recipient := nexus.CrossChainAddress{Chain: destinationChain, Address: e.DestinationAddress}
	amount := sdk.NewCoin(asset, sdk.Int(e.Amount))
	transferID, err := n.EnqueueTransfer(ctx, sourceChain, recipient, amount)
	if err != nil {
		bk.Logger(ctx).Info(fmt.Sprintf("failed enqueuing transfer for event %s due to error %s", event.GetID(), err.Error()))
		return false
	}

	bk.Logger(ctx).Debug(fmt.Sprintf("enqueued transfer for event from chain %s", sourceChain.Name),
		"chain", destinationChain.Name,
		"eventID", event.GetID(),
		"transferID", transferID.String(),
	)

	return true
}

func handleContractCall(ctx sdk.Context, event types.Event, bk types.BaseKeeper, n types.Nexus, multisig types.MultisigKeeper) bool {
	e := event.GetEvent().(*types.Event_ContractCall).ContractCall
	if e == nil {
		panic(fmt.Errorf("event is nil"))
	}

	sourceChain, destinationChain, err := validateChains(ctx, event.Chain, e.DestinationChain, bk, n)
	if err != nil {
		bk.Logger(ctx).Info(err.Error())
		return false
	}

	if !bk.HasChain(ctx, destinationChain.Name) {
		bk.Logger(ctx).Info(fmt.Sprintf("destination chain %s is not an evm chain", destinationChain.Name))
		return false
	}

	destinationCk := bk.ForChain(destinationChain.Name)

	destinationChainID, ok := destinationCk.GetChainID(ctx)
	if !ok {
		panic(fmt.Errorf("could not find chain ID for '%s'", destinationChain.Name))
	}

	keyID, ok := multisig.GetCurrentKeyID(ctx, destinationChain.Name)
	if !ok {
		panic(fmt.Errorf("no key for chain %s found", destinationChain.Name))
	}

	cmd, err := types.CreateApproveContractCallCommand(
		destinationChainID,
		keyID,
		sourceChain.Name,
		event.TxId,
		event.Index,
		*e,
	)
	if err != nil {
		panic(err)
	}

	if err := destinationCk.EnqueueCommand(ctx, cmd); err != nil {
		panic(err)
	}

	bk.Logger(ctx).Debug(fmt.Sprintf("created %s command for event", cmd.Command),
		"chain", destinationChain.Name,
		"eventID", event.GetID(),
		"commandID", cmd.ID.Hex(),
	)

	return true
}

func handleContractCallWithToken(ctx sdk.Context, event types.Event, bk types.BaseKeeper, n types.Nexus, multisig types.MultisigKeeper) bool {
	e := event.GetEvent().(*types.Event_ContractCallWithToken).ContractCallWithToken
	if e == nil {
		panic(fmt.Errorf("event is nil"))
	}

	sourceChain, destinationChain, err := validateChains(ctx, event.Chain, e.DestinationChain, bk, n)
	if err != nil {
		bk.Logger(ctx).Info(err.Error())
		return false
	}

	if !bk.HasChain(ctx, destinationChain.Name) {
		bk.Logger(ctx).Info(fmt.Sprintf("destination chain %s is not an evm chain", destinationChain.Name))
		return false
	}

	sourceCk := bk.ForChain(sourceChain.Name)
	destinationCk := bk.ForChain(destinationChain.Name)

	token := sourceCk.GetERC20TokenBySymbol(ctx, e.Symbol)
	if !token.Is(types.Confirmed) {
		bk.Logger(ctx).Info(fmt.Sprintf("%s token %s is not confirmed yet", event.Chain, e.Symbol))
		return false
	}

	asset := token.GetAsset()
	destinationToken := destinationCk.GetERC20TokenByAsset(ctx, asset)
	if !destinationToken.Is(types.Confirmed) {
		bk.Logger(ctx).Info(fmt.Sprintf("%s token with asset %s is not confirmed yet", e.DestinationChain, asset))
		return false
	}

	if !common.IsHexAddress(e.ContractAddress) {
		bk.Logger(ctx).Info(fmt.Sprintf("invalid contract address %s for chain %s", e.ContractAddress, e.DestinationChain))
		return false
	}

	destinationChainID, ok := destinationCk.GetChainID(ctx)
	if !ok {
		panic(fmt.Errorf("could not find chain ID for '%s'", destinationChain.Name))
	}

	keyID, ok := multisig.GetCurrentKeyID(ctx, destinationChain.Name)
	if !ok {
		panic(fmt.Errorf("no key for chain %s found", destinationChain.Name))
	}

	cmd, err := types.CreateApproveContractCallWithMintCommand(
		destinationChainID,
		keyID,
		sourceChain.Name,
		event.TxId,
		event.Index,
		*e,
		e.Amount,
		destinationToken.GetDetails().Symbol,
	)
	if err != nil {
		panic(err)
	}

	if err := destinationCk.EnqueueCommand(ctx, cmd); err != nil {
		panic(err)
	}

	bk.Logger(ctx).Debug(fmt.Sprintf("created %s command for event", cmd.Command),
		"chain", destinationChain.Name,
		"eventID", event.GetID(),
		"commandID", cmd.ID.Hex(),
	)

	return true
}

func handleConfirmDeposit(ctx sdk.Context, event types.Event, ck types.ChainKeeper, n types.Nexus, chain nexus.Chain) bool {
	e := event.GetEvent().(*types.Event_Transfer).Transfer
	if e == nil {
		panic(fmt.Errorf("event is nil"))
	}

	// get deposit address
	burnerInfo := ck.GetBurnerInfo(ctx, e.To)
	if burnerInfo == nil {
		ck.Logger(ctx).Info(fmt.Sprintf("no burner info found for address %s", e.To.Hex()))
		return false
	}

	depositAddr := nexus.CrossChainAddress{Chain: chain, Address: e.To.Hex()}
	recipient, ok := n.GetRecipient(ctx, depositAddr)
	if !ok {
		ck.Logger(ctx).Info(fmt.Sprintf("cross-chain sender has no recipient %s", e.To.Hex()))
		return false
	}

	amount := sdk.NewCoin(burnerInfo.Asset, sdk.NewIntFromBigInt(e.Amount.BigInt()))
	transferID, err := n.EnqueueForTransfer(ctx, depositAddr, amount)
	if err != nil {
		ck.Logger(ctx).Info(err.Error())
		return false
	}

	// set confirmed deposit
	erc20Deposit := types.ERC20Deposit{
		TxID:             event.TxId,
		Amount:           e.Amount,
		Asset:            burnerInfo.Asset,
		DestinationChain: burnerInfo.DestinationChain,
		BurnerAddress:    burnerInfo.BurnerAddress,
	}

	if _, _, ok := ck.GetDeposit(ctx, common.Hash(event.TxId), common.Address(burnerInfo.BurnerAddress)); ok {
		ck.Logger(ctx).Info(fmt.Sprintf("%s deposit %s-%s already exists", chain.Name.String(), event.TxId.Hex(), burnerInfo.BurnerAddress.Hex()))
		return false
	}
	ck.SetDeposit(ctx, erc20Deposit, types.DepositStatus_Confirmed)

	ck.Logger(ctx).Info(fmt.Sprintf("deposit confirmation result to %s %s", e.To.Hex(), e.Amount), "chain", chain.Name)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeDepositConfirmation,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyChain, chain.Name.String()),
			sdk.NewAttribute(types.AttributeKeySourceChain, chain.Name.String()),
			sdk.NewAttribute(types.AttributeKeyDestinationChain, recipient.Chain.Name.String()),
			sdk.NewAttribute(types.AttributeKeyDestinationAddress, recipient.Address),
			sdk.NewAttribute(types.AttributeKeyAmount, e.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyAsset, burnerInfo.Asset),
			sdk.NewAttribute(types.AttributeKeyDepositAddress, depositAddr.Address),
			sdk.NewAttribute(types.AttributeKeyTokenAddress, burnerInfo.TokenAddress.Hex()),
			sdk.NewAttribute(types.AttributeKeyTxID, event.TxId.Hex()),
			sdk.NewAttribute(types.AttributeKeyTransferID, transferID.String()),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueConfirm),
		))

	return true
}

func handleTokenDeployed(ctx sdk.Context, event types.Event, ck types.ChainKeeper, chain nexus.Chain) bool {
	e := event.GetEvent().(*types.Event_TokenDeployed).TokenDeployed
	if e == nil {
		panic(fmt.Errorf("event is nil"))
	}

	token := ck.GetERC20TokenBySymbol(ctx, e.Symbol)
	if token.Is(types.NonExistent) {
		ck.Logger(ctx).Info(fmt.Sprintf("token %s does not exist", e.Symbol))
		return false
	}

	if token.GetAddress() != e.TokenAddress {
		ck.Logger(ctx).Info(fmt.Sprintf("token address %s does not match expected %s", e.TokenAddress.Hex(), token.GetAddress().Hex()))
		return false
	}

	if err := token.ConfirmDeployment(); err != nil {
		ck.Logger(ctx).Info(err.Error())
		return false
	}

	ck.Logger(ctx).Info(fmt.Sprintf("token %s deployment confirmed on chain %s", e.Symbol, chain.Name))

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeTokenConfirmation,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyChain, chain.Name.String()),
			sdk.NewAttribute(types.AttributeKeyAsset, token.GetAsset()),
			sdk.NewAttribute(types.AttributeKeySymbol, token.GetDetails().Symbol),
			sdk.NewAttribute(types.AttributeKeyTokenAddress, token.GetAddress().Hex()),
			sdk.NewAttribute(types.AttributeKeyTxID, event.TxId.Hex()),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueConfirm),
		))

	return true
}

func handleMultisigTransferKey(ctx sdk.Context, event types.Event, ck types.ChainKeeper, multisig types.MultisigKeeper, chain nexus.Chain) bool {
	e := event.GetEvent().(*types.Event_MultisigOperatorshipTransferred).MultisigOperatorshipTransferred
	if e == nil {
		panic(fmt.Errorf("event is nil"))
	}

	// TODO: add weights to the event and also check
	newAddresses := e.NewOperators
	newThreshold := e.NewThreshold

	nextKeyID, ok := multisig.GetNextKeyID(ctx, chain.Name)
	if !ok {
		ck.Logger(ctx).Info(fmt.Sprintf("next %s key for chain %s not found", keyRole.SimpleString(), chain.Name))
		return false
	}

	nextKey, found := multisig.GetKey(ctx, nextKeyID)
	if !found {
		ck.Logger(ctx).Info(fmt.Sprintf("key %s not found", nextKeyID))
		return false
	}

	expectedAddressWeights, expectedThreshold := types.ParseMultisigKey(nextKey)

	if len(newAddresses) != len(expectedAddressWeights) {
		ck.Logger(ctx).Info(fmt.Sprintf("new addresses length does not match, expected %d got %d", len(expectedAddressWeights), len(newAddresses)))
		return false
	}

	addressSeen := make(map[string]bool)
	for _, newAddress := range newAddresses {
		newAddressHex := newAddress.Hex()
		if addressSeen[newAddressHex] {
			ck.Logger(ctx).Info("duplicate address in new addresses")
			return false
		}
		addressSeen[newAddressHex] = true

		_, ok := expectedAddressWeights[newAddressHex]
		if !ok {
			ck.Logger(ctx).Info("new addresses do not match")
			return false
		}
	}

	if !newThreshold.Equal(expectedThreshold) {
		ck.Logger(ctx).Info(fmt.Sprintf("new threshold does not match, expected %s got %s", expectedThreshold.String(), newThreshold.String()))
		return false
	}

	if err := multisig.RotateKey(ctx, chain.Name); err != nil {
		ck.Logger(ctx).Info(err.Error())
		return false
	}

	ck.Logger(ctx).Info(fmt.Sprintf("successfully confirmed key transfer for chain %s", chain.Name),
		"txID", event.TxId.Hex(),
		"keyID", nextKeyID,
	)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeTransferKeyConfirmation,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(types.AttributeKeyChain, chain.Name.String()),
		sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueConfirm),
	))

	return true
}

func handleConfirmedEvents(ctx sdk.Context, bk types.BaseKeeper, n types.Nexus, multisig types.MultisigKeeper) error {
	shouldHandleEvent := func(e codec.ProtoMarshaler) bool {
		event := e.(*types.Event)

		var destinationChainName nexus.ChainName
		switch event := event.GetEvent().(type) {
		case *types.Event_ContractCall:
			destinationChainName = event.ContractCall.DestinationChain
		case *types.Event_ContractCallWithToken:
			destinationChainName = event.ContractCallWithToken.DestinationChain
		case *types.Event_TokenSent:
			destinationChainName = event.TokenSent.DestinationChain
		case *types.Event_Transfer, *types.Event_TokenDeployed,
			*types.Event_MultisigOwnershipTransferred, *types.Event_MultisigOperatorshipTransferred:
			// skip checks for non-gateway tx event
			return true
		default:
			panic(fmt.Errorf("unsupported event type %T", event))
		}

		// would handle event as failure if destination chain is not registered
		destinationChain, ok := n.GetChain(ctx, destinationChainName)
		if !ok {
			return true
		}

		// skip if destination chain is not activated
		if !n.IsChainActivated(ctx, destinationChain) {
			bk.Logger(ctx).Debug(fmt.Sprintf("skipping confirmed event %s due to destination chain being inactive", event.GetID()),
				"chain", event.Chain.String(),
				"destination_chain", destinationChainName.String(),
				"eventID", event.GetID(),
			)

			return false
		}

		// skip further checks and handle event if destination is not an evm chain
		if !bk.HasChain(ctx, destinationChainName) {
			return true
		}

		// skip if destination chain has not got gateway set yet
		if _, ok := bk.ForChain(destinationChainName).GetGatewayAddress(ctx); !ok {
			bk.Logger(ctx).Debug(fmt.Sprintf("skipping confirmed event %s due to destination chain not having gateway set", event.GetID()),
				"chain", event.Chain.String(),
				"destination_chain", destinationChainName.String(),
				"eventID", event.GetID(),
			)

			return false
		}

		return true
	}

	for _, chain := range n.GetChains(ctx) {
		ck := bk.ForChain(chain.Name)
		queue := ck.GetConfirmedEventQueue(ctx)
		// skip if confirmed event queue is empty
		if queue.IsEmpty() {
			continue
		}

		var event types.Event
		for queue.DequeueUntil(&event, shouldHandleEvent) {
			bk.Logger(ctx).Debug("handling confirmed event",
				"chain", chain.Name.String(),
				"eventID", event.GetID(),
			)

			var ok bool

			switch event.GetEvent().(type) {
			case *types.Event_ContractCall:
				ok = handleContractCall(ctx, event, bk, n, multisig)
			case *types.Event_ContractCallWithToken:
				ok = handleContractCallWithToken(ctx, event, bk, n, multisig)
			case *types.Event_TokenSent:
				ok = handleTokenSent(ctx, event, bk, n)
			case *types.Event_Transfer:
				ok = handleConfirmDeposit(ctx, event, ck, n, chain)
			case *types.Event_TokenDeployed:
				ok = handleTokenDeployed(ctx, event, ck, chain)
			case *types.Event_MultisigOperatorshipTransferred:
				ok = handleMultisigTransferKey(ctx, event, ck, multisig, chain)
			default:
				bk.Logger(ctx).Debug("unsupported event type %T", event,
					"chain", chain.Name.String(),
					"eventID", event.GetID(),
				)

				ok = false
			}

			if !ok {
				if err := ck.SetEventFailed(ctx, event.GetID()); err != nil {
					return err
				}

				ck.Logger(ctx).Debug("failed handling event",
					"chain", chain.Name,
					"eventID", event.GetID(),
				)

				continue
			}

			bk.Logger(ctx).Debug("completed handling event",
				"chain", chain.Name.String(),
				"eventID", event.GetID(),
			)

			if err := ck.SetEventCompleted(ctx, event.GetID()); err != nil {
				return err
			}
		}
	}

	return nil
}

// BeginBlocker check for infraction evidence or downtime of validators
// on every begin block
func BeginBlocker(sdk.Context, abci.RequestBeginBlock, types.BaseKeeper) {}

// EndBlocker called every block, process inflation, update validator set.
func EndBlocker(ctx sdk.Context, _ abci.RequestEndBlock, bk types.BaseKeeper, n types.Nexus, multisig types.MultisigKeeper) ([]abci.ValidatorUpdate, error) {
	if err := handleConfirmedEvents(ctx, bk, n, multisig); err != nil {
		return nil, err
	}

	return nil, nil
}
