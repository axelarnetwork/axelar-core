package utils

import (
	"fmt"

	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	snapTypes "github.com/axelarnetwork/axelar-core/x/snapshot/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// FilterActiveValidators returns the subset of all validators that bonded and should be declared active
// and their aggregate staking power
func FilterActiveValidators(ctx sdk.Context, slasher snapTypes.Slasher, validators []snapshot.Validator) ([]snapshot.Validator, error) {
	var activeValidators []snapshot.Validator

	for _, validator := range validators {

		addr := validator.GetConsAddr()
		signingInfo, found := slasher.GetValidatorSigningInfo(ctx, addr)
		if !found {
			return nil, fmt.Errorf("snapshot: couldn't retrieve signing info for a validator")
		}

		// check if for any reason the validator should be declared as inactive
		// e.g., the validator missed to vote on blocks
		if signingInfo.Tombstoned || signingInfo.MissedBlocksCounter > 0 || validator.IsJailed() {
			continue
		}
		activeValidators = append(activeValidators, validator)
	}

	return activeValidators, nil
}

// FilterProxies selects only validators that have registered broadcast proxies
func FilterProxies(ctx sdk.Context, broadcaster snapTypes.Broadcaster, validators []snapshot.Validator) []snapshot.Validator {
	var withProxies []snapshot.Validator
	for _, v := range validators {
		proxy := broadcaster.GetProxy(ctx, v.GetOperator())
		if proxy != nil {
			withProxies = append(withProxies, v)
		}
	}

	return withProxies
}
