package evm

import (
	"fmt"
	"sort"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/utils/slices"
)

func validateChains(ctx sdk.Context, sourceChainName nexus.ChainName, destinationChainName nexus.ChainName, bk types.BaseKeeper, n types.Nexus) (nexus.Chain, nexus.Chain, error) {
	sourceChain, ok := n.GetChain(ctx, sourceChainName)
	if !ok {
		panic(fmt.Errorf("%s is not a registered chain", sourceChainName))
	}

	destinationChain, ok := n.GetChain(ctx, destinationChainName)
	if !ok {
		return nexus.Chain{}, nexus.Chain{}, fmt.Errorf("%s is not a registered chain", destinationChainName)
	}

	if !bk.HasChain(ctx, destinationChainName) {
		return nexus.Chain{}, nexus.Chain{}, fmt.Errorf("destination chain %s is not an evm chain", destinationChainName)
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
	destinationCk := bk.ForChain(destinationChain.Name)

	token := sourceCk.GetERC20TokenBySymbol(ctx, e.Symbol)
	if !token.Is(types.Confirmed) {
		bk.Logger(ctx).Info(fmt.Sprintf("%s token %s is not confirmed yet", event.Chain, e.Symbol))
		return false
	}

	asset := token.GetAsset()
	if token := destinationCk.GetERC20TokenByAsset(ctx, asset); !token.Is(types.Confirmed) {
		bk.Logger(ctx).Info(fmt.Sprintf("%s token with asset %s is not confirmed yet", e.DestinationChain, asset))
		return false
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

func handleContractCall(ctx sdk.Context, event types.Event, bk types.BaseKeeper, n types.Nexus, s types.Signer) bool {
	e := event.GetEvent().(*types.Event_ContractCall).ContractCall
	if e == nil {
		panic(fmt.Errorf("event is nil"))
	}

	sourceChain, destinationChain, err := validateChains(ctx, event.Chain, e.DestinationChain, bk, n)
	if err != nil {
		bk.Logger(ctx).Info(err.Error())
		return false
	}

	destinationCk := bk.ForChain(destinationChain.Name)

	destinationChainID, ok := destinationCk.GetChainID(ctx)
	if !ok {
		panic(fmt.Errorf("could not find chain ID for '%s'", destinationChain.Name))
	}

	keyID, ok := s.GetCurrentKeyID(ctx, destinationChain, tss.SecondaryKey)
	if !ok {
		panic(fmt.Errorf("no secondary key for chain %s found", destinationChain.Name))
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

func handleContractCallWithToken(ctx sdk.Context, event types.Event, bk types.BaseKeeper, n types.Nexus, s types.Signer) bool {
	e := event.GetEvent().(*types.Event_ContractCallWithToken).ContractCallWithToken
	if e == nil {
		panic(fmt.Errorf("event is nil"))
	}

	sourceChain, destinationChain, err := validateChains(ctx, event.Chain, e.DestinationChain, bk, n)
	if err != nil {
		bk.Logger(ctx).Info(err.Error())
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

	coin := sdk.NewCoin(asset, sdk.Int(e.Amount))
	fee, err := n.ComputeTransferFee(ctx, sourceChain, destinationChain, coin)
	if err != nil {
		bk.Logger(ctx).Info(fmt.Sprintf("failed computing transfer fee for event %s with error %s", event.GetID(), err.Error()))
		return false
	}

	if coin.IsLT(fee) {
		bk.Logger(ctx).Info(fmt.Sprintf("amount %s less than fee %s", e.Amount.String(), fee.Amount.String()))
		return false
	}

	destinationChainID, ok := destinationCk.GetChainID(ctx)
	if !ok {
		panic(fmt.Errorf("could not find chain ID for '%s'", destinationChain.Name))
	}

	keyID, ok := s.GetCurrentKeyID(ctx, destinationChain, tss.SecondaryKey)
	if !ok {
		panic(fmt.Errorf("no secondary key for chain %s found", destinationChain.Name))
	}

	amount := e.Amount.Sub(sdk.Uint(fee.Amount))
	cmd, err := types.CreateApproveContractCallWithMintCommand(
		destinationChainID,
		keyID,
		sourceChain.Name,
		event.TxId,
		event.Index,
		*e,
		amount,
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

	n.AddTransferFee(ctx, fee)

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

func handleMultisigTransferKey(ctx sdk.Context, event types.Event, ck types.ChainKeeper, s types.Signer, chain nexus.Chain) bool {
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
		panic(fmt.Errorf("event %s: unsupported event type %T", event.GetID(), event))
	}

	nextKeyID, ok := s.GetNextKeyID(ctx, chain, keyRole)
	if !ok {
		ck.Logger(ctx).Info(fmt.Sprintf("next %s key for chain %s not found", keyRole.SimpleString(), chain.Name))
		return false
	}

	nextKey, found := s.GetKey(ctx, nextKeyID)
	if !found {
		ck.Logger(ctx).Info(fmt.Sprintf("key %s not found", nextKeyID))
		return false
	}

	expectedAddress, expectedThreshold, err := types.GetMultisigAddresses(nextKey)
	if err != nil {
		ck.Logger(ctx).Info(err.Error())
		return false
	}

	newOwners := slices.Map(newAddresses, func(addr types.Address) common.Address { return common.Address(addr) })
	if !areAddressesEqual(expectedAddress, newOwners) {
		ck.Logger(ctx).Info(fmt.Sprintf("new adddress does not match, expected %v got %v", expectedAddress, newOwners))
		return false
	}

	if !sdk.NewUint(uint64(expectedThreshold)).Equal(newThreshold) {
		ck.Logger(ctx).Info(fmt.Sprintf("new threshold does not match, expected %d got %d", expectedThreshold, newThreshold.Uint64()))
		return false
	}

	if err := s.RotateKey(ctx, chain, keyRole); err != nil {
		ck.Logger(ctx).Info(err.Error())
		return false
	}

	ck.Logger(ctx).Info(fmt.Sprintf("successfully confirmed %s key transfer for chain %s",
		keyRole.SimpleString(), chain.Name), "txID", event.TxId.Hex(), "rotation count", s.GetRotationCount(ctx, chain, keyRole))

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeTransferKeyConfirmation,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(types.AttributeKeyChain, chain.Name.String()),
		sdk.NewAttribute(types.AttributeKeyTransferKeyType, keyRole.SimpleString()),
		sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueConfirm),
	))

	return true
}

func handleConfirmedEvents(ctx sdk.Context, bk types.BaseKeeper, n types.Nexus, s types.Signer) error {
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
		// would handle event as failure if destination chain is not an evm chain
		if !bk.HasChain(ctx, destinationChainName) {
			return true
		}
		// skip if destination chain is not activated
		if !n.IsChainActivated(ctx, destinationChain) {
			return false
		}
		// skip if destination chain has not got gateway set yet
		if _, ok := bk.ForChain(destinationChainName).GetGatewayAddress(ctx); !ok {
			return false
		}
		// skip if destination chain has the secondary key rotation in progress
		if _, nextSecondaryKeyAssigned := s.GetNextKeyID(ctx, destinationChain, tss.SecondaryKey); nextSecondaryKeyAssigned {
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
			var ok bool

			switch event.GetEvent().(type) {
			case *types.Event_ContractCall:
				ok = handleContractCall(ctx, event, bk, n, s)
			case *types.Event_ContractCallWithToken:
				ok = handleContractCallWithToken(ctx, event, bk, n, s)
			case *types.Event_TokenSent:
				ok = handleTokenSent(ctx, event, bk, n)
			case *types.Event_Transfer:
				ok = handleConfirmDeposit(ctx, event, ck, n, chain)
			case *types.Event_TokenDeployed:
				ok = handleTokenDeployed(ctx, event, ck, chain)
			case *types.Event_MultisigOwnershipTransferred, *types.Event_MultisigOperatorshipTransferred:
				ok = handleMultisigTransferKey(ctx, event, ck, s, chain)
			default:
				return fmt.Errorf("unsupported event type %T", event)
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
func EndBlocker(ctx sdk.Context, _ abci.RequestEndBlock, bk types.BaseKeeper, n types.Nexus, s types.Signer) ([]abci.ValidatorUpdate, error) {
	if err := handleConfirmedEvents(ctx, bk, n, s); err != nil {
		return nil, err
	}

	return nil, nil
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
