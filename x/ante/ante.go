package ante

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/x/ante/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
)

func logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// HandlerDecorator is an ante decorator wrapper for an ante handler
type HandlerDecorator struct {
	handler sdk.AnteHandler
}

// NewAnteHandlerDecorator constructor for HandlerDecorator
func NewAnteHandlerDecorator(handler sdk.AnteHandler) HandlerDecorator {
	return HandlerDecorator{handler}
}

// AnteHandle wraps the next AnteHandler to perform custom pre- and post-processing
func (decorator HandlerDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
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
		case *stakingtypes.MsgUndelegate:
			valAddress, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
			if err != nil {
				return sdk.Context{}, err
			}
			chains := d.nexus.GetChains(ctx)

			for _, chain := range chains {
				for _, keyRole := range tss.GetKeyRoles() {
					currentKeyID, found := d.tss.GetCurrentKeyID(ctx, chain, keyRole)
					if found && isValidatorHoldingTssShareOf(ctx, d.tss, d.snapshotter, valAddress, currentKeyID) {
						return ctx, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "validator %s cannot unbond while holding tss share of %s's current %s key ", valAddress, chain.Name, keyRole.SimpleString())
					}

					nextKeyID, found := d.tss.GetNextKeyID(ctx, chain, keyRole)
					if found && isValidatorHoldingTssShareOf(ctx, d.tss, d.snapshotter, valAddress, nextKeyID) {
						return ctx, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "validator %s cannot unbond while holding tss share of %s's current %s key ", valAddress, chain.Name, keyRole.SimpleString())
					}

					rotationCount := d.tss.GetRotationCount(ctx, chain, keyRole)
					unbondingLockingKeyRotationCount := d.tss.GetKeyUnbondingLockingKeyRotationCount(ctx)
					for i := 1; i <= int(unbondingLockingKeyRotationCount); i++ {
						key, ok := d.tss.GetKeyByRotationCount(ctx, chain, keyRole, rotationCount-int64(i))
						if !ok {
							logger(ctx).Debug(fmt.Sprintf("only %d previous %s keys exist for chain", i, keyRole.SimpleString()))

							break
						}

						if isValidatorHoldingTssShareOf(ctx, d.tss, d.snapshotter, valAddress, key.ID) {
							return ctx, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "validator %s cannot unbond while holding tss share of %s's %s key %s", valAddress, chain.Name, keyRole.SimpleString(), key.ID)
						}
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

	for _, validator := range snapshot.Validators {
		if validator.GetSDKValidator().GetOperator().Equals(valAddress) {
			return true
		}
	}

	return false
}
