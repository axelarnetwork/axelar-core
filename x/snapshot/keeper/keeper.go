package keeper

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/snapshot/types"
)

const (
	operatorPrefix = "operator_"
	proxyPrefix    = "proxy_"
)

// Keeper represents the snapshot keeper
type Keeper struct {
	storeKey sdk.StoreKey
	staking  types.StakingKeeper
	bank     types.BankKeeper
	slasher  types.Slasher
	cdc      codec.BinaryCodec
	params   params.Subspace
}

// NewKeeper creates a new keeper for the staking module
func NewKeeper(cdc codec.BinaryCodec, key sdk.StoreKey, paramSpace params.Subspace, staking types.StakingKeeper, bank types.BankKeeper, slasher types.Slasher) Keeper {
	return Keeper{
		storeKey: key,
		cdc:      cdc,
		staking:  staking,
		bank:     bank,
		params:   paramSpace.WithKeyTable(types.KeyTable()),
		slasher:  slasher,
	}
}

// Logger returns the logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// SetParams sets the module's parameters
func (k Keeper) SetParams(ctx sdk.Context, set types.Params) {
	k.params.SetParamSet(ctx, &set)
}

// GetParams gets the module's parameters
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.params.GetParamSet(ctx, &params)
	return
}

// GetMinProxyBalance returns the minimum balance proxies must hold
func (k Keeper) GetMinProxyBalance(ctx sdk.Context) sdk.Int {
	var minBalance int64
	k.params.Get(ctx, types.KeyMinProxyBalance, &minBalance)
	return sdk.NewInt(minBalance)
}

// ActivateProxy registers a proxy address for a given operator, which can broadcast messages in the principal's name
// The proxy will be marked as active and to be included in the next snapshot by default
func (k Keeper) ActivateProxy(ctx sdk.Context, operator sdk.ValAddress, proxy sdk.AccAddress) error {
	if bytes.Equal(operator, proxy) {
		return fmt.Errorf("proxy address cannot be the same as the operator address")
	}

	if existing, ok := k.getProxiedValidator(ctx, operator); ok && !existing.Proxy.Equals(proxy) {
		return fmt.Errorf(
			"proxy mismatch, expected %s, got %s",
			existing.Proxy.String(),
			proxy.String(),
		)
	}

	if existing, ok := k.getProxiedValidator(ctx, proxy); ok && !existing.Validator.Equals(operator) {
		return fmt.Errorf(
			"validator mismatch, expected %s, got %s",
			existing.Validator.String(),
			operator.String(),
		)
	}

	minBalance := k.GetMinProxyBalance(ctx)
	denom := k.staking.BondDenom(ctx)
	if balance := k.bank.GetBalance(ctx, proxy, denom); balance.Amount.LT(minBalance) {
		return fmt.Errorf("account %s does not have sufficient funds to become a proxy (minimum %s%s, actual %s)",
			proxy.String(), minBalance.String(), denom, balance.String())
	}

	k.setProxiedValidator(ctx, types.NewProxiedValidator(operator, proxy, true))

	return nil
}

// DeactivateProxy deactivates the proxy address for a given operator
func (k Keeper) DeactivateProxy(ctx sdk.Context, operator sdk.ValAddress) error {
	val := k.staking.Validator(ctx, operator)
	if val == nil {
		return fmt.Errorf("validator %s is unknown", operator.String())
	}

	proxiedValidator, ok := k.getProxiedValidator(ctx, operator)
	if !ok {
		return fmt.Errorf("validator %s has no proxy registered", operator.String())
	}

	proxiedValidator.Active = false
	k.setProxiedValidator(ctx, proxiedValidator)

	return nil
}

func (k Keeper) getProxiedValidator(ctx sdk.Context, addr sdk.Address) (types.ProxiedValidator, bool) {
	var proxiedValidator types.ProxiedValidator

	if bz := ctx.KVStore(k.storeKey).Get([]byte(proxyPrefix + addr.String())); bz != nil {
		k.cdc.MustUnmarshalLengthPrefixed(bz, &proxiedValidator)
		return proxiedValidator, true
	} else if bz := ctx.KVStore(k.storeKey).Get([]byte(operatorPrefix + addr.String())); bz != nil {
		k.cdc.MustUnmarshalLengthPrefixed(bz, &proxiedValidator)
		return proxiedValidator, true
	} else {
		return types.ProxiedValidator{}, false
	}
}

func (k Keeper) getProxiedValidators(ctx sdk.Context) []types.ProxiedValidator {
	var proxiedValidators []types.ProxiedValidator

	iter := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), []byte(proxyPrefix))
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var proxiedValidator types.ProxiedValidator
		k.cdc.MustUnmarshalLengthPrefixed(iter.Value(), &proxiedValidator)

		proxiedValidators = append(proxiedValidators, proxiedValidator)
	}

	return proxiedValidators
}

func (k Keeper) setProxiedValidator(ctx sdk.Context, proxiedValidator types.ProxiedValidator) {
	bz := k.cdc.MustMarshalLengthPrefixed(&proxiedValidator)

	ctx.KVStore(k.storeKey).Set([]byte(operatorPrefix+proxiedValidator.Validator.String()), bz)
	ctx.KVStore(k.storeKey).Set([]byte(proxyPrefix+proxiedValidator.Proxy.String()), bz)
}

// GetOperator returns the principal address for a given proxy address. Returns nil if not set.
func (k Keeper) GetOperator(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress {
	if proxiedValidator, ok := k.getProxiedValidator(ctx, proxy); ok && proxiedValidator.Active {
		return proxiedValidator.Validator
	}

	return nil
}

// GetProxy returns the proxy address for a given operator address. Returns nil if not set.
// The bool value denotes wether or not the proxy is active and to be included in the next snapshot
func (k Keeper) GetProxy(ctx sdk.Context, operator sdk.ValAddress) (addr sdk.AccAddress, active bool) {
	if proxiedValidator, ok := k.getProxiedValidator(ctx, operator); ok {
		return proxiedValidator.Proxy, proxiedValidator.Active
	}

	return nil, false
}

// CreateSnapshot returns a new snapshot giving each candidate its proper weight,
// or returns an error if the threshold cannot be met given the total weight of all
// validators in the system; candidates are excluded if the given filterFunc is
// evaluated to false or their weight is zero (NOTE: snapshot itself does not keep track of the threshold)
func (k Keeper) CreateSnapshot(
	ctx sdk.Context,
	candidates []sdk.ValAddress,
	filterFunc func(exported.ValidatorI) bool,
	weightFunc func(consensusPower sdk.Uint) sdk.Uint,
	threshold utils.Threshold,
) (exported.Snapshot, error) {
	powerReduction := k.staking.PowerReduction(ctx)
	participants := make([]exported.Participant, 0, len(candidates))

	for _, candidate := range candidates {
		validator := k.staking.Validator(ctx, candidate)
		if validator == nil || !filterFunc(validator) {
			continue
		}

		weight := weightFunc(sdk.NewUint(uint64(validator.GetConsensusPower(powerReduction))))
		// Participants with zero weight are useless for all intents and purposes.
		// We filter them out here so any process dealing with snapshots doesn't have to worry about them
		if weight.IsZero() {
			continue
		}
		participants = append(participants, exported.NewParticipant(validator.GetOperator(), weight))
	}

	bondedWeight := sdk.ZeroUint()
	k.staking.IterateBondedValidatorsByPower(ctx, func(_ int64, v stakingtypes.ValidatorI) (stop bool) {
		if v == nil {
			panic("nil bonded validator received")
		}

		weight := weightFunc(sdk.NewUint(uint64(v.GetConsensusPower(powerReduction))))
		bondedWeight = bondedWeight.Add(weight)

		// we do not stop until we've iterated through all bonded validators.
		// Due to the unknown nature of weightFunc, every validator might contribute
		// some weight
		return false
	})

	snapshot := exported.NewSnapshot(ctx.BlockTime(), ctx.BlockHeight(), participants, bondedWeight)

	participantsWeight := snapshot.GetParticipantsWeight()
	if participantsWeight.LT(snapshot.CalculateMinPassingWeight(threshold)) {
		return exported.Snapshot{}, fmt.Errorf("given threshold %s cannot be met (participants weight: %s, bonded weight: %s)", threshold.String(), participantsWeight, bondedWeight)
	}

	if err := snapshot.ValidateBasic(); err != nil {
		return exported.Snapshot{}, err
	}

	return snapshot, nil
}
