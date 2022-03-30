package evm

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/x/evm/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
)

func handleContractCallWithToken(ctx sdk.Context, event types.Event, bk types.BaseKeeper, n types.Nexus, s types.Signer) (types.Command, bool) {
	e := event.GetEvent().(*types.Event_ContractCallWithToken).ContractCallWithToken
	if e == nil {
		panic(fmt.Errorf("event is nil"))
	}

	sourceCk := bk.ForChain(event.Chain)
	destinationCk := bk.ForChain(e.DestinationChain)

	// TODO: cosmos sdk invariants may be a good solution to avoid the check here
	sourceChain, ok := n.GetChain(ctx, event.Chain)
	if !ok {
		panic(fmt.Errorf("%s is not a registered chain", event.Chain))
	}

	destinationChain, ok := n.GetChain(ctx, e.DestinationChain)
	if !ok {
		panic(fmt.Errorf("%s is not a registered chain", e.DestinationChain))
	}

	token := sourceCk.GetERC20TokenBySymbol(ctx, e.Symbol)
	if !token.Is(types.Confirmed) {
		bk.Logger(ctx).Info(fmt.Sprintf("%s token %s is not confirmed yet", event.Chain, e.Symbol))
		return types.Command{}, false
	}

	asset := token.GetAsset()
	if token := destinationCk.GetERC20TokenByAsset(ctx, asset); !token.Is(types.Confirmed) {
		bk.Logger(ctx).Info(fmt.Sprintf("%s token with asset %s is not confirmed yet", e.DestinationChain, asset))
		return types.Command{}, false
	}

	if !common.IsHexAddress(e.ContractAddress) {
		bk.Logger(ctx).Info(fmt.Sprintf("invalid contract address %s for chain %s", e.ContractAddress, e.DestinationChain))
		return types.Command{}, false
	}

	coin := sdk.NewCoin(asset, sdk.Int(e.Amount))
	fee, err := n.ComputeTransferFee(ctx, sourceChain, destinationChain, coin)
	if err != nil {
		bk.Logger(ctx).Info(fmt.Sprintf("failed computing transfer fee for event %s with error %s", event.GetID(), err.Error()))
		return types.Command{}, false
	}

	if coin.IsLT(fee) {
		bk.Logger(ctx).Info(fmt.Sprintf("amount %s less than fee %s", e.Amount.String(), fee.Amount.String()))
		return types.Command{}, false
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
	)
	if err != nil {
		panic(err)
	}

	n.AddTransferFee(ctx, fee)

	return cmd, true
}

func handleConfirmedEvents(ctx sdk.Context, bk types.BaseKeeper, n types.Nexus, s types.Signer) error {
	for _, chain := range n.GetChains(ctx) {
		// skip if not an evm chain
		if !bk.HasChain(ctx, chain.Name) {
			continue
		}

		// skip if not activated
		if !n.IsChainActivated(ctx, chain) {
			continue
		}

		ck := bk.ForChain(chain.Name)
		// skip if gateway not set yet
		if _, ok := ck.GetGatewayAddress(ctx); !ok {
			continue
		}

		queue := ck.GetContractCallQueue(ctx)
		// skip if confirmed event queue is empty
		if queue.IsEmpty() {
			continue
		}

		var event types.Event
		for queue.Dequeue(&event) {
			var cmd types.Command
			var ok bool

			switch event.GetEvent().(type) {
			case *types.Event_ContractCallWithToken:
				cmd, ok = handleContractCallWithToken(ctx, event, bk, n, s)
			default:
				return fmt.Errorf("unsupported event type %T", event)
			}

			switch ok {
			case true:
				if err := ck.EnqueueCommand(ctx, cmd); err != nil {
					return err
				}

				if err := ck.SetEventCompleted(ctx, event.GetID()); err != nil {
					return err
				}

				ck.Logger(ctx).Debug(fmt.Sprintf("created %s command for event", cmd.Command),
					"chain", chain.Name,
					"eventID", event.GetID(),
					"commandID", cmd.ID.Hex(),
				)
			default:
				if err := ck.SetEventFailed(ctx, event.GetID()); err != nil {
					return err
				}

				ck.Logger(ctx).Debug("failed creating command for event",
					"chain", chain.Name,
					"eventID", event.GetID(),
				)
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
