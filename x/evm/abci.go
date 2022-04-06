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

func validateChains(ctx sdk.Context, sourceChainName string, destinationChainName string, bk types.BaseKeeper, n types.Nexus) (nexus.Chain, nexus.Chain, error) {
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

func handleConfirmedEvents(ctx sdk.Context, bk types.BaseKeeper, n types.Nexus, s types.Signer) error {
	shouldHandleEvent := func(e codec.ProtoMarshaler) bool {
		event := e.(*types.Event)

		var destinationChainName string
		switch event := event.GetEvent().(type) {
		case *types.Event_ContractCall:
			destinationChainName = event.ContractCall.DestinationChain
		case *types.Event_ContractCallWithToken:
			destinationChainName = event.ContractCallWithToken.DestinationChain
		case *types.Event_TokenSent:
			destinationChainName = event.TokenSent.DestinationChain
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
		for queue.Dequeue(&event, shouldHandleEvent) {
			var ok bool

			switch event.GetEvent().(type) {
			case *types.Event_ContractCall:
				ok = handleContractCall(ctx, event, bk, n, s)
			case *types.Event_ContractCallWithToken:
				ok = handleContractCallWithToken(ctx, event, bk, n, s)
			case *types.Event_TokenSent:
				ok = handleTokenSent(ctx, event, bk, n)
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
