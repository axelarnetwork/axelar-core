package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

func (k Keeper) setValidatorStatus(ctx sdk.Context, validatorStatus types.ValidatorStatus) {
	k.getStore(ctx).Set(validatorStatusPrefix.AppendStr(validatorStatus.Validator.String()), &validatorStatus)
}

func (k Keeper) getValidatorStatus(ctx sdk.Context, validator sdk.ValAddress) (validatorStatus types.ValidatorStatus, ok bool) {
	return validatorStatus, k.getStore(ctx).Get(validatorStatusPrefix.AppendStr(validator.String()), &validatorStatus)
}

func (k Keeper) getValidatorStatuses(ctx sdk.Context) (validatorStatuses []types.ValidatorStatus) {
	iter := k.getStore(ctx).Iterator(validatorStatusPrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var validatorStatus types.ValidatorStatus
		iter.UnmarshalValue(&validatorStatus)

		validatorStatuses = append(validatorStatuses, validatorStatus)
	}

	return validatorStatuses
}

func (k Keeper) setSuspendedUntil(ctx sdk.Context, validator sdk.ValAddress, suspendedUntilBlockNumber int64) {
	validatorStatus, _ := k.getValidatorStatus(ctx, validator)
	validatorStatus.Validator = validator
	validatorStatus.SuspendedUntil = uint64(suspendedUntilBlockNumber)

	k.setValidatorStatus(ctx, validatorStatus)
}

// GetSuspendedUntil returns the block number at which a validator is released from TSS suspension
func (k Keeper) GetSuspendedUntil(ctx sdk.Context, validator sdk.ValAddress) int64 {
	validatorStatus, _ := k.getValidatorStatus(ctx, validator)

	return int64(validatorStatus.SuspendedUntil)
}
