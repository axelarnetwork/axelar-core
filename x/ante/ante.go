package ante

import (
	"fmt"

	"github.com/axelarnetwork/axelar-core/x/ante/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	sdkStaking "github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/tendermint/tendermint/libs/log"
)

func logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// AnteHandlerDecorator is an ante decorator wrapper for an ante handler
type AnteHandlerDecorator struct {
	handler sdk.AnteHandler
}

// NewAnteHandlerDecorator constructor for AnteHandlerDecorator
func NewAnteHandlerDecorator(handler sdk.AnteHandler) AnteHandlerDecorator {
	return AnteHandlerDecorator{handler}
}

// AnteHandle wraps the next AnteHandler to perform custom pre- and post-processing
func (decorator AnteHandlerDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	if newCtx, err = decorator.handler(ctx, tx, simulate); err != nil {
		return newCtx, err
	}

	return next(newCtx, tx, simulate)
}

// ValidateValidatorDeregisteredTssDecorator checks if the unbonding validator holds any tss share of active crypto keys
type ValidateValidatorDeregisteredTssDecorator struct {
	tss         types.Tss
	nexus       types.Nexus
	snapshotter types.Snapshotter
}

// NewValidateValidatorDeregisteredTssDecorator constructor for ValidateValidatorDeregisteredTssDecorator
func NewValidateValidatorDeregisteredTssDecorator(tss types.Tss, nexus types.Nexus, snapshotter types.Snapshotter) ValidateValidatorDeregisteredTssDecorator {
	return ValidateValidatorDeregisteredTssDecorator{
		tss,
		nexus,
		snapshotter,
	}
}

// AnteHandle fails the transaction if it finds any validator holding tss share of active keys is trying to unbond
func (d ValidateValidatorDeregisteredTssDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	msgs := tx.GetMsgs()

	for _, msg := range msgs {
		switch msg := msg.(type) {
		case sdkStaking.MsgUndelegate:
			valAddress := msg.ValidatorAddress
			chains := d.nexus.GetChains(ctx)

			for _, chain := range chains {
				for _, keyRole := range tss.GetKeyRoles() {
					currentKeyId, found := d.tss.GetCurrentKeyID(ctx, chain, keyRole)
					if found && isValidatorHoldingTssShareOf(ctx, d.tss, d.snapshotter, valAddress, currentKeyId) {
						return ctx, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "validator %s cannot unbond while holding tss share of %s's current %s key ", valAddress.String(), chain.Name, keyRole.String())
					}

					nextKeyId, found := d.tss.GetNextKeyID(ctx, chain, keyRole)
					if found && isValidatorHoldingTssShareOf(ctx, d.tss, d.snapshotter, valAddress, nextKeyId) {
						return ctx, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "validator %s cannot unbond while holding tss share of %s's current %s key ", valAddress.String(), chain.Name, keyRole.String())
					}
				}
			}
		default:
			continue
		}
	}

	return next(ctx, tx, simulate)
}

func isValidatorHoldingTssShareOf(ctx sdk.Context, tss types.Tss, snapshotter types.Snapshotter, valAddress sdk.ValAddress, keyID string) bool {
	counter, ok := tss.GetSnapshotCounterForKeyID(ctx, keyID)
	if !ok {
		logger(ctx).Error(fmt.Sprintf("no snapshot counter for key ID %s registered", keyID))

		return false
	}

	snapshot, ok := snapshotter.GetSnapshot(ctx, counter)
	if !ok {
		logger(ctx).Error(fmt.Sprintf("no snapshot found for counter num %d", counter))

		return false
	}

	for _, validators := range snapshot.Validators {
		if validators.GetOperator().Equals(valAddress) {
			return true
		}
	}

	return false
}
