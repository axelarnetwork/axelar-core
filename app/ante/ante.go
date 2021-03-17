package ante

import (
	"fmt"

	"github.com/axelarnetwork/axelar-core/app/ante/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	sdkStaking "github.com/cosmos/cosmos-sdk/x/staking"
)

type AnteHandlerDecorator struct {
	handler sdk.AnteHandler
}

func NewAnteHandlerDecorator(handler sdk.AnteHandler) AnteHandlerDecorator {
	return AnteHandlerDecorator{handler}
}

func (decorator AnteHandlerDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	if newCtx, err = decorator.handler(ctx, tx, simulate); err != nil {
		return newCtx, err
	}

	return next(newCtx, tx, simulate)
}

type ValidateValidatorDeregisteredTssDecorator struct {
	tss         types.Tss
	nexus       types.Nexus
	snapshotter types.Snapshotter
}

func NewValidateValidatorDeregisteredTssDecorator(tss types.Tss, nexus types.Nexus, snapshotter types.Snapshotter) ValidateValidatorDeregisteredTssDecorator {
	return ValidateValidatorDeregisteredTssDecorator{
		tss,
		nexus,
		snapshotter,
	}
}

func (d ValidateValidatorDeregisteredTssDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	msgs := tx.GetMsgs()

	for _, msg := range msgs {
		switch msg := msg.(type) {
		case sdkStaking.MsgUndelegate:
			valAddress := msg.ValidatorAddress
			chains := d.nexus.GetChains(ctx)

			for _, chain := range chains {
				if currentMasterKeyId, found := d.tss.GetCurrentMasterKeyID(ctx, chain); found {
					isHoldingTssShare, err := isValidatorHoldingTssShareOf(ctx, d.tss, d.snapshotter, valAddress, currentMasterKeyId)
					if err != nil {
						return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
					}

					if isHoldingTssShare {
						return ctx, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "validator %s cannot unbond while holding tss share of %s's current master key ", valAddress.String(), chain.Name)
					}
				}

				if nextMasterKeyId, found := d.tss.GetNextMasterKeyID(ctx, chain); found {
					isHoldingTssShare, err := isValidatorHoldingTssShareOf(ctx, d.tss, d.snapshotter, valAddress, nextMasterKeyId)
					if err != nil {
						return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
					}

					if isHoldingTssShare {
						return ctx, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "validator %s cannot unbond while holding tss share of %s's current master key ", valAddress.String(), chain.Name)
					}
				}
			}
		default:
			continue
		}
	}

	return next(ctx, tx, simulate)
}

func isValidatorHoldingTssShareOf(ctx sdk.Context, tss types.Tss, snapshotter types.Snapshotter, valAddress sdk.ValAddress, keyID string) (bool, error) {
	counter, ok := tss.GetSnapshotCounterForKeyID(ctx, keyID)
	if !ok {
		return false, fmt.Errorf("no snapshot counter for key ID %s registered", keyID)
	}

	snapshot, ok := snapshotter.GetSnapshot(ctx, counter)
	if !ok {
		return false, fmt.Errorf("no snapshot found for counter num %d", counter)
	}

	for _, validators := range snapshot.Validators {
		if validators.GetOperator().Equals(valAddress) {
			return true, nil
		}
	}

	return false, nil
}
