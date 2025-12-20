package evm

import (
	"fmt"

	"cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/CosmWasm/wasmd/x/wasm"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/utils/events"
	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	multisig "github.com/axelarnetwork/axelar-core/x/multisig/exported"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
)

// EndBlocker handles cross-chain message flow between EVM chains and the Axelar network.
// It performs two main operations each block:
//
// 1. ROUTING (EVM → Nexus): Process confirmed events from EVM chains
//
// Events are confirmed by validators through the vote handler and queued for processing.
// Routing is intentionally deferred to the EndBlocker rather than being performed when
// the poll completes, because the last voter to complete a poll should not bear the gas
// cost of expensive routing operations. This ensures voting remains cheap and predictable.
//
// Each confirmed event is routed based on its type:
//
//   - ContractCall/ContractCallWithToken → Creates a GeneralMessage in nexus and enqueues
//     it for routing to the destination chain.
//
//   - TokenDeployed → Confirms the token deployment on the source chain, enabling
//     cross-chain transfers for that token.
//
//   - MultisigOperatorshipTransferred → Confirms key rotation on the chain, updating
//     the active key used for signing outbound transactions.
//
// Unsupported event types (TokenSent, Transfer) are marked as failed.
//
// 2. DELIVERY (Nexus → EVM): Deliver pending messages to destination EVM chains
//
// Messages routed through nexus to EVM destinations are delivered by creating
// gateway commands:
//
//   - GeneralMessage (no asset) → ApproveContractCall command
//   - GeneralMessage (with asset) → ApproveContractCallWithMint command
//
// Commands are signed by the validator set and batched for execution on the
// destination chain's gateway contract.
//
// Error Handling:
//   - Each event/message is processed in a cached context (RunCached). If processing
//     fails, the cached writes are discarded but the failure is recorded.
//   - Failed events emit EVMEventFailed; failed messages emit ContractCallFailed.
//   - Failures don't block processing of subsequent events/messages.
//
// Bounded Computation:
//   - Processing is limited by EndBlockerLimit per chain per block.
//   - Remaining events/messages are processed in subsequent blocks.
//
// Event Emission:
//   - Success events (ContractCallApproved, etc.) are emitted inside RunCached,
//     so they roll back with state if something fails.
//   - Failure events (EVMEventFailed, ContractCallFailed) are emitted outside
//     RunCached, so they persist for observability even if the event/message
//     is retried later.
func EndBlocker(ctx sdk.Context, bk types.BaseKeeper, n types.Nexus, m types.MultisigKeeper) ([]abci.ValidatorUpdate, error) {
	processConfirmedEvents(ctx, bk, n, m)
	deliverPendingMessages(ctx, bk, n, m)

	return nil, nil
}

// ============================================================================
// ROUTING: EVM → Nexus
// ============================================================================

func processConfirmedEvents(ctx sdk.Context, bk types.BaseKeeper, n types.Nexus, m types.MultisigKeeper) {
	for _, chain := range slices.Filter(n.GetChains(ctx), types.IsEVMChain) {
		processConfirmedEventsForChain(ctx, chain, bk, n, m)
	}
}

func processConfirmedEventsForChain(ctx sdk.Context, chain nexus.Chain, bk types.BaseKeeper, n types.Nexus, m types.MultisigKeeper) {
	ck := funcs.Must(bk.ForChain(ctx, chain.Name))
	queue := ck.GetConfirmedEventQueue(ctx)
	endBlockerLimit := ck.GetParams(ctx).EndBlockerLimit

	var events []types.Event
	var event types.Event
	for int64(len(events)) < endBlockerLimit && queue.Dequeue(&event) {
		events = append(events, event)
	}

	for _, event := range events {
		success := utils.RunCached(ctx, bk, func(ctx sdk.Context) (bool, error) {
			if err := processConfirmedEvent(ctx, event, bk, n, m); err != nil {
				ck.Logger(ctx).Debug(fmt.Sprintf("failed processing event: %s", err.Error()),
					"chain", chain.Name.String(),
					"eventID", event.GetID(),
				)

				return false, err
			}

			ck.Logger(ctx).Debug("completed processing event",
				"chain", chain.Name.String(),
				"eventID", event.GetID(),
			)

			return true, nil
		})

		if !success {
			funcs.MustNoErr(ck.SetEventFailed(ctx, event.GetID()))
			continue
		}

		funcs.MustNoErr(ck.SetEventCompleted(ctx, event.GetID()))
	}
}

func processConfirmedEvent(ctx sdk.Context, event types.Event, bk types.BaseKeeper, n types.Nexus, m types.MultisigKeeper) error {
	if err := validateEvent(ctx, event, bk, n); err != nil {
		return err
	}

	switch event.GetEvent().(type) {
	case *types.Event_ContractCall:
		return routeContractCall(ctx, event, n)
	case *types.Event_ContractCallWithToken:
		return routeContractCallWithToken(ctx, event, bk, n)
	case *types.Event_TokenDeployed:
		return applyTokenDeployment(ctx, event, bk, n)
	case *types.Event_MultisigOperatorshipTransferred:
		return applyKeyRotation(ctx, event, bk, n, m)
	default:
		panic(fmt.Errorf("unsupported event type %T", event))
	}
}

func validateEvent(ctx sdk.Context, event types.Event, bk types.BaseKeeper, n types.Nexus) error {
	var destinationChainName nexus.ChainName
	var contractAddress string
	switch event := event.GetEvent().(type) {
	case *types.Event_ContractCall:
		destinationChainName = event.ContractCall.DestinationChain
		contractAddress = event.ContractCall.ContractAddress
	case *types.Event_ContractCallWithToken:
		destinationChainName = event.ContractCallWithToken.DestinationChain
		contractAddress = event.ContractCallWithToken.ContractAddress
	case *types.Event_TokenDeployed, *types.Event_MultisigOperatorshipTransferred:
		// skip checks for non-gateway tx event
		return nil
	default:
		panic(fmt.Errorf("unsupported event type %T", event))
	}

	destinationChain, ok := n.GetChain(ctx, destinationChainName)
	if !ok {
		return fmt.Errorf("destination chain not found")
	}

	if !n.IsChainActivated(ctx, destinationChain) {
		return fmt.Errorf("destination chain de-activated")
	}

	if !destinationChain.IsFrom(types.ModuleName) {
		return nil
	}

	if len(contractAddress) != 0 && !common.IsHexAddress(contractAddress) {
		return fmt.Errorf("invalid contract address")
	}

	destinationCk, err := bk.ForChain(ctx, destinationChainName)
	if err != nil {
		return fmt.Errorf("destination chain not EVM-compatible")
	}

	if _, ok := destinationCk.GetGatewayAddress(ctx); !ok {
		return fmt.Errorf("destination chain gateway not deployed yet")
	}

	return nil
}

func routeContractCall(ctx sdk.Context, event types.Event, n types.Nexus) error {
	return routeEventToNexus(ctx, n, event, nil)
}

func routeContractCallWithToken(ctx sdk.Context, event types.Event, bk types.BaseKeeper, n types.Nexus) error {
	e := event.GetContractCallWithToken()
	if e == nil {
		panic(fmt.Errorf("event is nil"))
	}

	sourceChain := funcs.MustOk(n.GetChain(ctx, event.Chain))
	sourceCk := funcs.Must(bk.ForChain(ctx, sourceChain.Name))

	token := sourceCk.GetERC20TokenBySymbol(ctx, e.Symbol)
	if !token.Is(types.Confirmed) {
		return fmt.Errorf("token with symbol %s not confirmed on source chain", e.Symbol)
	}

	coin := sdk.NewCoin(token.GetAsset(), math.Int(e.Amount))
	return routeEventToNexus(ctx, n, event, &coin)
}

func routeEventToNexus(ctx sdk.Context, n types.Nexus, event types.Event, asset *sdk.Coin) error {
	sourceChain := funcs.MustOk(n.GetChain(ctx, event.Chain))

	var message nexus.GeneralMessage
	switch e := event.GetEvent().(type) {
	case *types.Event_ContractCall:
		sender := nexus.CrossChainAddress{
			Chain:   sourceChain,
			Address: e.ContractCall.Sender.Hex(),
		}

		recipient := nexus.CrossChainAddress{
			Chain:   funcs.MustOk(n.GetChain(ctx, e.ContractCall.DestinationChain)),
			Address: e.ContractCall.ContractAddress,
		}

		message = nexus.NewGeneralMessage(
			string(event.GetID()),
			sender,
			recipient,
			e.ContractCall.PayloadHash.Bytes(),
			event.TxID.Bytes(),
			event.Index,
			nil,
		)

	case *types.Event_ContractCallWithToken:
		if asset == nil {
			return fmt.Errorf("expect asset for ContractCallWithToken")
		}

		sender := nexus.CrossChainAddress{
			Chain:   sourceChain,
			Address: e.ContractCallWithToken.Sender.Hex(),
		}

		recipient := nexus.CrossChainAddress{
			Chain:   funcs.MustOk(n.GetChain(ctx, e.ContractCallWithToken.DestinationChain)),
			Address: e.ContractCallWithToken.ContractAddress,
		}

		message = nexus.NewGeneralMessage(
			string(event.GetID()),
			sender,
			recipient,
			e.ContractCallWithToken.PayloadHash.Bytes(),
			event.TxID.Bytes(),
			event.Index,
			asset,
		)
	default:
		return fmt.Errorf("unsupported event type %T", event)
	}

	if message.Recipient.Chain.Name.Equals(axelarnet.Axelarnet.Name) {
		return fmt.Errorf("%s is not a supported recipient", axelarnet.Axelarnet.Name)
	}

	if message.Asset != nil && message.Recipient.Chain.IsFrom(wasm.ModuleName) {
		return fmt.Errorf("%s is not a supported recipient for calls with asset transfer", wasm.ModuleName)
	}

	if err := n.SetNewMessage(ctx, message); err != nil {
		return err
	}

	return n.EnqueueRouteMessage(ctx, message.ID)
}

func applyTokenDeployment(ctx sdk.Context, event types.Event, bk types.BaseKeeper, n types.Nexus) error {
	e := event.GetEvent().(*types.Event_TokenDeployed).TokenDeployed
	if e == nil {
		panic(fmt.Errorf("event is nil"))
	}

	chain := funcs.MustOk(n.GetChain(ctx, event.Chain))
	ck := funcs.Must(bk.ForChain(ctx, event.Chain))

	token := ck.GetERC20TokenBySymbol(ctx, e.Symbol)
	if token.Is(types.NonExistent) {
		return fmt.Errorf("token %s does not exist", e.Symbol)
	}

	if token.GetAddress() != e.TokenAddress {
		return fmt.Errorf("token address %s does not match expected %s", e.TokenAddress.Hex(), token.GetAddress().Hex())
	}

	if err := token.ConfirmDeployment(); err != nil {
		return err
	}

	ck.Logger(ctx).Info(fmt.Sprintf("token %s deployment confirmed on chain %s", e.Symbol, chain.Name),
		"chain", chain.Name,
		"asset", token.GetAsset(),
		"eventID", event.GetID(),
		"txID", event.TxID.Hex(),
	)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeTokenConfirmation,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyChain, chain.Name.String()),
			sdk.NewAttribute(types.AttributeKeyAsset, token.GetAsset()),
			sdk.NewAttribute(types.AttributeKeySymbol, token.GetDetails().Symbol),
			sdk.NewAttribute(types.AttributeKeyTokenAddress, token.GetAddress().Hex()),
			sdk.NewAttribute(types.AttributeKeyTxID, event.TxID.Hex()),
			sdk.NewAttribute(types.AttributeKeyEventID, string(event.GetID())),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueConfirm),
		))

	return nil
}

func applyKeyRotation(ctx sdk.Context, event types.Event, bk types.BaseKeeper, n types.Nexus, multisig types.MultisigKeeper) error {
	e := event.GetEvent().(*types.Event_MultisigOperatorshipTransferred).MultisigOperatorshipTransferred
	if e == nil {
		panic(fmt.Errorf("event is nil"))
	}

	chain := funcs.MustOk(n.GetChain(ctx, event.Chain))
	ck := funcs.Must(bk.ForChain(ctx, event.Chain))
	newAddresses := e.NewOperators
	newWeights := e.NewWeights
	newThreshold := e.NewThreshold

	nextKeyID, ok := multisig.GetNextKeyID(ctx, chain.Name)
	if !ok {
		return fmt.Errorf("next key for chain %s not found", chain.Name)
	}

	nextKey, ok := multisig.GetKey(ctx, nextKeyID)
	if !ok {
		return fmt.Errorf("key %s not found", nextKeyID)
	}

	expectedAddressWeights, expectedThreshold := types.ParseMultisigKey(nextKey)

	if len(newAddresses) != len(expectedAddressWeights) {
		return fmt.Errorf("new addresses length does not match, expected %d got %d", len(expectedAddressWeights), len(newAddresses))
	}

	addressSeen := make(map[string]bool)
	for i, newAddress := range newAddresses {
		newAddressHex := newAddress.Hex()
		if addressSeen[newAddressHex] {
			return fmt.Errorf("duplicate address in new addresses")
		}
		addressSeen[newAddressHex] = true

		expectedWeight, ok := expectedAddressWeights[newAddressHex]
		if !ok {
			return fmt.Errorf("new addresses do not match")
		}

		if !expectedWeight.Equal(newWeights[i]) {
			return fmt.Errorf("new weights do not match")
		}
	}

	if !newThreshold.Equal(expectedThreshold) {
		return fmt.Errorf("new threshold does not match, expected %s got %s", expectedThreshold.String(), newThreshold.String())
	}

	if err := multisig.RotateKey(ctx, chain.Name); err != nil {
		return err
	}

	ck.Logger(ctx).Info(fmt.Sprintf("successfully confirmed key transfer for chain %s", chain.Name),
		"chain", chain.Name,
		"txID", event.TxID.Hex(),
		"eventID", event.GetID(),
		"keyID", nextKeyID,
	)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeTransferKeyConfirmation,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(types.AttributeKeyChain, chain.Name.String()),
		sdk.NewAttribute(types.AttributeKeyTxID, event.TxID.Hex()),
		sdk.NewAttribute(types.AttributeKeyEventID, string(event.GetID())),
		sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueConfirm),
	))

	return nil
}

// ============================================================================
// DELIVERY: Nexus → EVM
// ============================================================================

func deliverPendingMessages(ctx sdk.Context, bk types.BaseKeeper, n types.Nexus, m types.MultisigKeeper) {
	for _, chain := range slices.Filter(n.GetChains(ctx), types.IsEVMChain) {
		destCk := funcs.Must(bk.ForChain(ctx, chain.Name))
		endBlockerLimit := destCk.GetParams(ctx).EndBlockerLimit
		msgs := n.GetProcessingMessages(ctx, chain.Name, endBlockerLimit)

		bk.Logger(ctx).Debug(fmt.Sprintf("delivering %d messages", len(msgs)), types.AttributeKeyChain, chain.Name)

		for _, msg := range msgs {
			success := false
			_ = utils.RunCached(ctx, bk, func(ctx sdk.Context) (bool, error) {
				if err := validateMessageForDelivery(ctx, destCk, n, m, chain, msg); err != nil {
					bk.Logger(ctx).Info(fmt.Sprintf("failed validating message for delivery: %s", err.Error()),
						types.AttributeKeyChain, msg.GetDestinationChain(),
						types.AttributeKeyMessageID, msg.ID,
					)
					return false, err
				}

				chainID := funcs.MustOk(destCk.GetChainID(ctx))
				keyID := funcs.MustOk(m.GetCurrentKeyID(ctx, chain.Name))

				switch msg.Type() {
				case nexus.TypeGeneralMessage:
					deliverMessage(ctx, destCk, chainID, keyID, msg)
				case nexus.TypeGeneralMessageWithToken:
					deliverMessageWithToken(ctx, destCk, chainID, keyID, msg)
				default:
					panic(fmt.Sprintf("unrecognized message type %d", msg.Type()))
				}

				success = true
				return true, nil
			})

			if !success {
				destCk.Logger(ctx).Error("failed delivering message",
					types.AttributeKeyChain, chain.Name.String(),
					types.AttributeKeyMessageID, msg.ID,
				)

				events.Emit(ctx, &types.ContractCallFailed{
					Chain:     chain.Name,
					MessageID: msg.ID,
				})

				funcs.MustNoErr(n.SetMessageFailed(ctx, msg.ID))

				continue
			}

			funcs.MustNoErr(n.SetMessageExecuted(ctx, msg.ID))
		}
	}
}

func validateMessageForDelivery(ctx sdk.Context, ck types.ChainKeeper, n types.Nexus, m types.MultisigKeeper, chain nexus.Chain, msg nexus.GeneralMessage) error {
	_, ok := m.GetCurrentKeyID(ctx, chain.Name)
	if !ok {
		return fmt.Errorf("current key not set")
	}

	if !n.IsChainActivated(ctx, chain) {
		return fmt.Errorf("destination chain de-activated")
	}

	if _, ok := ck.GetGatewayAddress(ctx); !ok {
		return fmt.Errorf("destination chain gateway not deployed yet")
	}

	if !common.IsHexAddress(msg.GetDestinationAddress()) {
		return fmt.Errorf("invalid contract address")
	}

	switch msg.Type() {
	case nexus.TypeGeneralMessage:
		return nil
	case nexus.TypeGeneralMessageWithToken:
		token := ck.GetERC20TokenByAsset(ctx, msg.Asset.GetDenom())
		if !token.Is(types.Confirmed) {
			return fmt.Errorf("asset %s not confirmed on destination chain", msg.Asset.GetDenom())
		}
		return nil
	default:
		return fmt.Errorf("unrecognized message type %d", msg.Type())
	}
}

func deliverMessage(ctx sdk.Context, ck types.ChainKeeper, chainID math.Int, keyID multisig.KeyID, msg nexus.GeneralMessage) {
	cmd := types.NewApproveContractCallCommandGeneric(chainID, keyID, common.HexToAddress(msg.GetDestinationAddress()), common.BytesToHash(msg.PayloadHash), common.BytesToHash(msg.SourceTxID), msg.GetSourceChain(), msg.GetSourceAddress(), msg.SourceTxIndex, msg.ID)
	funcs.MustNoErr(ck.EnqueueCommand(ctx, cmd))

	events.Emit(ctx, &types.ContractCallApproved{
		Chain:            msg.GetSourceChain(),
		EventID:          types.EventID(msg.ID),
		CommandID:        cmd.ID,
		Sender:           msg.GetSourceAddress(),
		DestinationChain: msg.GetDestinationChain(),
		ContractAddress:  msg.GetDestinationAddress(),
		PayloadHash:      types.Hash(common.BytesToHash(msg.PayloadHash)),
	})

	ck.Logger(ctx).Debug("delivered message",
		types.AttributeKeyChain, msg.GetDestinationChain(),
		types.AttributeKeyMessageID, msg.ID,
		types.AttributeKeyCommandsID, cmd.ID,
	)
}

func deliverMessageWithToken(ctx sdk.Context, ck types.ChainKeeper, chainID math.Int, keyID multisig.KeyID, msg nexus.GeneralMessage) {
	token := ck.GetERC20TokenByAsset(ctx, msg.Asset.GetDenom())
	cmd := types.NewApproveContractCallWithMintGeneric(chainID, keyID, common.BytesToHash(msg.SourceTxID), msg.SourceTxIndex, msg, token.GetDetails().Symbol)
	funcs.MustNoErr(ck.EnqueueCommand(ctx, cmd))

	events.Emit(ctx, &types.ContractCallWithMintApproved{
		Chain:            msg.GetSourceChain(),
		EventID:          types.EventID(msg.ID),
		CommandID:        cmd.ID,
		Sender:           msg.GetSourceAddress(),
		DestinationChain: msg.GetDestinationChain(),
		ContractAddress:  msg.GetDestinationAddress(),
		PayloadHash:      types.Hash(common.BytesToHash(msg.PayloadHash)),
		Asset:            *msg.Asset,
	})

	ck.Logger(ctx).Debug("delivered message with token",
		types.AttributeKeyChain, msg.GetDestinationChain(),
		types.AttributeKeyMessageID, msg.ID,
		types.AttributeKeyCommandsID, cmd.ID,
	)
}
