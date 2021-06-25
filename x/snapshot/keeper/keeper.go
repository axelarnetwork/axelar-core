package keeper

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/snapshot/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
)

const (
	proxyCountKey  = "proxyCount"
	lastCounterKey = "lastcounter"

	counterPrefix = "counter_"
)

// Make sure the keeper implements the Snapshotter interface
var _ exported.Snapshotter = Keeper{}

// Keeper represents the snapshot keeper
type Keeper struct {
	storeKey sdk.StoreKey
	staking  types.StakingKeeper
	slasher  exported.Slasher
	tss      exported.Tss
	cdc      *codec.LegacyAmino
	params   params.Subspace
}

// NewKeeper creates a new keeper for the staking module
func NewKeeper(cdc *codec.LegacyAmino, key sdk.StoreKey, paramSpace params.Subspace, staking types.StakingKeeper, slasher exported.Slasher, tss exported.Tss) Keeper {
	return Keeper{
		storeKey: key,
		cdc:      cdc,
		staking:  staking,
		params:   paramSpace.WithKeyTable(types.KeyTable()),
		slasher:  slasher,
		tss:      tss,
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

// TakeSnapshot attempts to create a new snapshot; if subsetSize equals 0, snapshot will be created with all validators
func (k Keeper) TakeSnapshot(ctx sdk.Context, subsetSize int64, keyShareDistributionPolicy tss.KeyShareDistributionPolicy) (snapshotConsensusPower sdk.Int, totalConsensusPower sdk.Int, err error) {
	s, ok := k.GetLatestSnapshot(ctx)

	if !ok {
		k.setLatestCounter(ctx, 0)
		return k.executeSnapshot(ctx, 0, subsetSize, keyShareDistributionPolicy)
	}

	lockingPeriod := k.getLockingPeriod(ctx)
	if s.Timestamp.Add(lockingPeriod).After(ctx.BlockTime()) {
		return sdk.ZeroInt(), sdk.ZeroInt(), fmt.Errorf("not enough time has passed since last snapshot, need to wait %s longer",
			s.Timestamp.Add(lockingPeriod).Sub(ctx.BlockTime()).String())
	}

	k.setLatestCounter(ctx, s.Counter+1)
	return k.executeSnapshot(ctx, s.Counter+1, subsetSize, keyShareDistributionPolicy)
}

func (k Keeper) getLockingPeriod(ctx sdk.Context) time.Duration {
	var lockingPeriod time.Duration
	k.params.Get(ctx, types.KeyLockingPeriod, &lockingPeriod)
	return lockingPeriod
}

// GetLatestSnapshot retrieves the last created snapshot
func (k Keeper) GetLatestSnapshot(ctx sdk.Context) (exported.Snapshot, bool) {
	r := k.GetLatestCounter(ctx)
	if r == -1 {
		return exported.Snapshot{}, false
	}

	return k.GetSnapshot(ctx, r)
}

// GetSnapshot retrieves a snapshot by counter, if it exists
func (k Keeper) GetSnapshot(ctx sdk.Context, counter int64) (exported.Snapshot, bool) {
	bz := ctx.KVStore(k.storeKey).Get(counterKey(counter))
	if bz == nil {

		return exported.Snapshot{}, false
	}

	var snapshot exported.Snapshot
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &snapshot)

	return snapshot, true
}

// GetLatestCounter returns the latest snapshot counter
func (k Keeper) GetLatestCounter(ctx sdk.Context) int64 {
	bz := ctx.KVStore(k.storeKey).Get([]byte(lastCounterKey))

	if bz == nil {

		return -1
	}

	var i int64
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &i)
	return i
}

func (k Keeper) executeSnapshot(ctx sdk.Context, counter int64, subsetSize int64, keyShareDistributionPolicy tss.KeyShareDistributionPolicy) (snapshotConsensusPower sdk.Int, totalConsensusPower sdk.Int, err error) {
	var validators []exported.SDKValidator
	snapshotConsensusPower, totalConsensusPower = sdk.ZeroInt(), sdk.ZeroInt()

	validatorIter := func(_ int64, validator stakingtypes.ValidatorI) (stop bool) {
		totalConsensusPower = totalConsensusPower.AddRaw(validator.GetConsensusPower())

		// this explicit type cast is necessary, because snapshot needs to call UnpackInterfaces() on the validator
		// and it is not exposed in the ValidatorI interface
		v, ok := validator.(exported.SDKValidator)
		if !ok {
			k.Logger(ctx).Error(fmt.Sprintf("unexpected validator type: expected %T, got %T", stakingtypes.Validator{}, validator))
			return false
		}

		if !exported.IsValidatorActive(ctx, k.slasher, v) ||
			!exported.HasProxyRegistered(ctx, k, v) ||
			exported.IsValidatorTssSuspended(ctx, k.tss, v) {
			return false
		}

		validators = append(validators, v)

		// if subsetSize equals 0, we will iterate through all validators and potentially put them all into the snapshot
		return len(validators) == int(subsetSize)
	}
	// IterateBondedValidatorsByPower(https://github.com/cosmos/cosmos-sdk/blob/7fc7b3f6ff82eb5ede52881778114f6b38bd7dfa/x/staking/keeper/alias_functions.go#L33) iterates validators by power in descending order
	k.staking.IterateBondedValidatorsByPower(ctx, validatorIter)

	minBondFractionPerShare := k.tss.GetMinBondFractionPerShare(ctx)
	var participants []exported.Validator

	for _, validator := range validators {
		if !minBondFractionPerShare.IsMet(sdk.NewInt(validator.GetConsensusPower()), totalConsensusPower) {
			// Since IterateBondedValidatorsByPower iterates validators by power in descending order, once
			// we find a validator with consensus power below minimum, we don't have to continue anymore
			break
		}

		snapshotConsensusPower = snapshotConsensusPower.AddRaw(validator.GetConsensusPower())
		participants = append(participants, exported.NewValidator(validator, 0))
	}

	if len(participants) == 0 {
		return sdk.ZeroInt(), sdk.ZeroInt(), fmt.Errorf("no validator is eligible for keygen")
	}

	if subsetSize > 0 && len(participants) != int(subsetSize) {
		return sdk.ZeroInt(), sdk.ZeroInt(), fmt.Errorf("only %d validators are eligible for keygen which is less than desired subset size %d", len(participants), subsetSize)
	}

	// Since IterateBondedValidatorsByPower iterates validators by power in descending order, the last participant is
	// the one with least amount of bond among all participants
	bondPerShare := participants[len(participants)-1].GetConsensusPower()
	totalShareCount := sdk.ZeroInt()
	for i := range participants {
		switch keyShareDistributionPolicy {
		case tss.WeightedByStake:
			participants[i].ShareCount = participants[i].GetConsensusPower() / bondPerShare
		case tss.OnePerValidator:
			participants[i].ShareCount = 1
		default:
			return sdk.ZeroInt(), sdk.ZeroInt(), fmt.Errorf("invalid key share distribution policy %d", keyShareDistributionPolicy)
		}

		totalShareCount = totalShareCount.AddRaw(participants[i].ShareCount)
	}

	snapshot := exported.Snapshot{
		Validators:                 participants,
		Timestamp:                  ctx.BlockTime(),
		Height:                     ctx.BlockHeight(),
		TotalShareCount:            totalShareCount,
		Counter:                    counter,
		KeyShareDistributionPolicy: keyShareDistributionPolicy,
	}

	ctx.KVStore(k.storeKey).Set(counterKey(counter), k.cdc.MustMarshalBinaryLengthPrefixed(snapshot))

	return snapshotConsensusPower, totalConsensusPower, nil
}

func (k Keeper) setLatestCounter(ctx sdk.Context, counter int64) {
	ctx.KVStore(k.storeKey).Set([]byte(lastCounterKey), k.cdc.MustMarshalBinaryLengthPrefixed(counter))
}

func counterKey(counter int64) []byte {
	return []byte(fmt.Sprintf("%s%d", counterPrefix, counter))
}

// RegisterProxy registers a proxy address for a given principal, which can broadcast messages in the principal's name
// The proxy will be marked as active and to be included in the next snapshot by default
func (k Keeper) RegisterProxy(ctx sdk.Context, principal sdk.ValAddress, proxy sdk.AccAddress) error {
	val := k.staking.Validator(ctx, principal)
	if val == nil {
		return fmt.Errorf("validator %s is unknown", principal.String())
	}
	k.Logger(ctx).Debug("getting proxy count")
	count := k.getProxyCount(ctx)

	storedProxy := ctx.KVStore(k.storeKey).Get(principal)
	if storedProxy != nil {
		ctx.KVStore(k.storeKey).Delete(storedProxy)
		count--
	}
	k.Logger(ctx).Debug("setting proxy")
	ctx.KVStore(k.storeKey).Set(proxy, principal)
	// Creating a reverse lookup
	bz := append([]byte{1}, proxy...)
	ctx.KVStore(k.storeKey).Set(principal, bz)
	count++
	k.Logger(ctx).Debug("setting proxy count")
	k.setProxyCount(ctx, count)
	k.Logger(ctx).Debug("done")
	return nil
}

// DeactivateProxy deactivates the proxy address for a given principal
func (k Keeper) DeactivateProxy(ctx sdk.Context, principal sdk.ValAddress) error {
	val := k.staking.Validator(ctx, principal)
	if val == nil {
		return fmt.Errorf("validator %s is unknown", principal.String())
	}
	k.Logger(ctx).Debug("getting proxy count")

	storedProxy := ctx.KVStore(k.storeKey).Get(principal)
	if storedProxy == nil {
		return fmt.Errorf("validator %s has no proxy registered", principal.String())
	}

	k.Logger(ctx).Debug(fmt.Sprintf("deactivating proxy %s", sdk.AccAddress(storedProxy[1:]).String()))
	bz := append([]byte{0}, storedProxy[1:]...)
	ctx.KVStore(k.storeKey).Set(principal, bz)

	return nil
}

// GetPrincipal returns the proxy address for a given principal address. Returns nil if not set.
func (k Keeper) GetPrincipal(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress {
	if proxy == nil {
		return nil
	}
	return ctx.KVStore(k.storeKey).Get(proxy)
}

// GetProxy returns the proxy address for a given principal address. Returns nil if not set.
// The bool value denotes wether or not the proxy is active and to be included in the next snapshot
func (k Keeper) GetProxy(ctx sdk.Context, principal sdk.ValAddress) (addr sdk.AccAddress, active bool) {

	bz := ctx.KVStore(k.storeKey).Get(principal)
	if bz == nil {
		return nil, active
	}
	addr = bz[1:]

	if bz[0] == 1 {
		active = true
	}

	return addr, active
}

func (k Keeper) setProxyCount(ctx sdk.Context, count int) {
	k.Logger(ctx).Debug(fmt.Sprintf("number of known proxies: %v", count))
	ctx.KVStore(k.storeKey).Set([]byte(proxyCountKey), k.cdc.MustMarshalBinaryLengthPrefixed(count))
}

func (k Keeper) getProxyCount(ctx sdk.Context) int {
	bz := ctx.KVStore(k.storeKey).Get([]byte(proxyCountKey))
	if bz == nil {
		return 0
	}
	var count int
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &count)
	return count
}
