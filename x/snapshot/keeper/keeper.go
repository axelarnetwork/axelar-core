package keeper

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/snapshot/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
)

const (
	proxyCountKey  = "proxyCount"
	lastCounterKey = "lastcounter"

	counterPrefix = "counter_"
	proxyPrefix   = "vald_"
)

// Make sure the keeper implements the Snapshotter interface
var _ exported.Snapshotter = Keeper{}

// Keeper represents the snapshot keeper
type Keeper struct {
	storeKey sdk.StoreKey
	staking  types.StakingKeeper
	slasher  exported.Slasher
	tss      exported.Tss
	cdc      codec.BinaryCodec
	params   params.Subspace
}

// NewKeeper creates a new keeper for the staking module
func NewKeeper(cdc codec.BinaryCodec, key sdk.StoreKey, paramSpace params.Subspace, staking types.StakingKeeper, slasher exported.Slasher, tss exported.Tss) Keeper {
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

// SetProxyReady establishes that the specified proxy is ready to be registered
func (k Keeper) SetProxyReady(ctx sdk.Context, operator sdk.ValAddress, proxy sdk.AccAddress) {
	key := []byte(proxyPrefix + operator.String())
	ctx.KVStore(k.storeKey).Set(key, proxy.Bytes())
}

// IsProxyReady returns true if a proxy has issued a readiness message for the given operator address
func (k Keeper) IsProxyReady(ctx sdk.Context, operator sdk.ValAddress) bool {
	key := []byte(proxyPrefix + operator.String())
	return ctx.KVStore(k.storeKey).Has(key)
}

// TakeSnapshot attempts to create a new snapshot based on the given key requirment
func (k Keeper) TakeSnapshot(ctx sdk.Context, keyRequirement tss.KeyRequirement) (exported.Snapshot, error) {
	s, ok := k.GetLatestSnapshot(ctx)

	if !ok {
		k.setLatestCounter(ctx, 0)
		return k.executeSnapshot(ctx, 0, keyRequirement)
	}

	lockingPeriod := k.getLockingPeriod(ctx)
	if s.Timestamp.Add(lockingPeriod).After(ctx.BlockTime()) {
		return exported.Snapshot{}, fmt.Errorf("not enough time has passed since last snapshot, need to wait %s longer",
			s.Timestamp.Add(lockingPeriod).Sub(ctx.BlockTime()).String())
	}

	k.setLatestCounter(ctx, s.Counter+1)
	return k.executeSnapshot(ctx, s.Counter+1, keyRequirement)
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
	k.cdc.MustUnmarshalLengthPrefixed(bz, &snapshot)

	return snapshot, true
}

// GetLatestCounter returns the latest snapshot counter
func (k Keeper) GetLatestCounter(ctx sdk.Context) int64 {
	bz := ctx.KVStore(k.storeKey).Get([]byte(lastCounterKey))
	if bz == nil {
		return -1
	}

	return int64(binary.LittleEndian.Uint64(bz))
}

func (k Keeper) executeSnapshot(ctx sdk.Context, counter int64, keyRequirement tss.KeyRequirement) (exported.Snapshot, error) {
	var validators []exported.SDKValidator
	var participants []exported.Validator
	var nonParticipants []exported.Validator
	snapshotConsensusPower, totalConsensusPower := sdk.ZeroInt(), sdk.ZeroInt()

	var requiredValidatorsCount int
	switch keyRequirement.KeyShareDistributionPolicy {
	case tss.WeightedByStake:
		requiredValidatorsCount = 0
	case tss.OnePerValidator:
		requiredValidatorsCount = int(keyRequirement.MaxTotalShareCount)
	default:
		return exported.Snapshot{}, fmt.Errorf("invalid key share distribution policy %d", keyRequirement.KeyShareDistributionPolicy)
	}

	validatorIter := func(_ int64, validator stakingtypes.ValidatorI) (stop bool) {
		totalConsensusPower = totalConsensusPower.AddRaw(validator.GetConsensusPower(k.staking.PowerReduction(ctx)))

		// this explicit type cast is necessary, because snapshot needs to call UnpackInterfaces() on the validator
		// and it is not exposed in the ValidatorI interface
		v, ok := validator.(stakingtypes.Validator)
		if !ok {
			k.Logger(ctx).Error(fmt.Sprintf("unexpected validator type: expected %T, got %T", stakingtypes.Validator{}, validator))
			nonParticipants = append(nonParticipants, exported.NewValidator(&v, 0))
			return false
		}

		illegibility, err := k.GetValidatorIllegibility(ctx, &v)
		if err != nil {
			k.Logger(ctx).Error(err.Error())
			nonParticipants = append(nonParticipants, exported.NewValidator(&v, 0))
			return false
		}

		if illegibility = illegibility.FilterIllegibilityForNewKey(); !illegibility.Is(exported.None) {
			k.Logger(ctx).Error(fmt.Sprintf("excluding validator %s from snapshot %d due to [%s]",
				validator.GetOperator().String(),
				counter,
				illegibility.String(),
			))
			nonParticipants = append(nonParticipants, exported.NewValidator(&v, 0))
			return false
		}

		if !k.tss.IsOperatorAvailable(ctx, validator.GetOperator()) {
			k.Logger(ctx).Error(fmt.Sprintf("excluding validator %s from snapshot %d due to [not-available]",
				validator.GetOperator().String(),
				counter,
			))
			nonParticipants = append(nonParticipants, exported.NewValidator(&v, 0))
			return false
		}

		validators = append(validators, &v)

		// if requiredValidatorsCount equals 0, we will iterate through all validators and potentially put them all into the snapshot
		return len(validators) == requiredValidatorsCount
	}
	// IterateBondedValidatorsByPower(https://github.com/cosmos/cosmos-sdk/blob/7fc7b3f6ff82eb5ede52881778114f6b38bd7dfa/x/staking/keeper/alias_functions.go#L33) iterates validators by power in descending order
	k.staking.IterateBondedValidatorsByPower(ctx, validatorIter)

	// minBondFractionPerShare is only relavant if KeyShareDistributionPolicy is set to WeightedByStake
	minBondFractionPerShare := utils.Threshold{Numerator: 1, Denominator: keyRequirement.MaxTotalShareCount}
	for _, validator := range validators {
		if keyRequirement.KeyShareDistributionPolicy == tss.WeightedByStake && !minBondFractionPerShare.IsMet(sdk.NewInt(validator.GetConsensusPower(k.staking.PowerReduction(ctx))), totalConsensusPower) {
			nonParticipants = append(nonParticipants, exported.NewValidator(validator, 0))
			continue
		}

		snapshotConsensusPower = snapshotConsensusPower.AddRaw(validator.GetConsensusPower(k.staking.PowerReduction(ctx)))
		participants = append(participants, exported.NewValidator(validator, 0))
	}

	// build event-related data
	participantsAddr := make([]string, 0, len(participants))
	participantsStake := make([]uint32, 0, len(participants))
	for _, participant := range participants {
		participantsAddr = append(participantsAddr, participant.GetSDKValidator().GetOperator().String())
		participantsStake = append(participantsStake, uint32(participant.GetSDKValidator().GetConsensusPower(k.staking.PowerReduction(ctx))))
	}

	nonParticipantsAddr := make([]string, 0, len(nonParticipants))
	nonParticipantsStake := make([]uint32, 0, len(nonParticipants))
	for _, nonParticipant := range nonParticipants {
		nonParticipantsAddr = append(nonParticipantsAddr, nonParticipant.GetSDKValidator().GetOperator().String())
		nonParticipantsStake = append(nonParticipantsStake, uint32(nonParticipant.GetSDKValidator().GetConsensusPower(k.staking.PowerReduction(ctx))))
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeCreateSnapshot,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeParticipants, string(types.ModuleCdc.LegacyAmino.MustMarshalJSON(participantsAddr))),
			sdk.NewAttribute(types.AttributeParticipantsStake, string(types.ModuleCdc.LegacyAmino.MustMarshalJSON(participantsStake))),
			sdk.NewAttribute(types.AttributeNonParticipants, string(types.ModuleCdc.LegacyAmino.MustMarshalJSON(nonParticipantsAddr))),
			sdk.NewAttribute(types.AttributeNonParticipantsStake, string(types.ModuleCdc.LegacyAmino.MustMarshalJSON(nonParticipantsStake))),
		),
	)

	k.Logger(ctx).Debug(fmt.Sprintf("Snapshot %d has participating validators %v with stake %v and non-participating validators %v with stake %v",
		counter,
		participantsAddr,
		participantsStake,
		nonParticipantsAddr,
		nonParticipantsStake,
	))

	if len(participants) == 0 {
		return exported.Snapshot{}, fmt.Errorf("no validator is eligible for keygen")
	}

	if !keyRequirement.MinKeygenThreshold.IsMet(snapshotConsensusPower, totalConsensusPower) {
		return exported.Snapshot{}, fmt.Errorf(fmt.Sprintf(
			"Unable to meet min stake threshold required for keygen: active %s out of %s total",
			snapshotConsensusPower.String(),
			totalConsensusPower.String(),
		))
	}

	// Since IterateBondedValidatorsByPower iterates validators by power in descending order, the last participant is
	// the one with least amount of bond among all participants
	bondPerShare := participants[len(participants)-1].GetSDKValidator().GetConsensusPower(k.staking.PowerReduction(ctx))
	totalShareCount := sdk.ZeroInt()
	for i := range participants {
		switch keyRequirement.KeyShareDistributionPolicy {
		case tss.WeightedByStake:
			participants[i].ShareCount = participants[i].GetSDKValidator().GetConsensusPower(k.staking.PowerReduction(ctx)) / bondPerShare
		case tss.OnePerValidator:
			participants[i].ShareCount = 1
		default:
			return exported.Snapshot{}, fmt.Errorf("invalid key share distribution policy %d", keyRequirement.KeyShareDistributionPolicy)
		}

		totalShareCount = totalShareCount.AddRaw(participants[i].ShareCount)
	}

	if totalShareCount.LT(sdk.NewInt(keyRequirement.MinTotalShareCount)) {
		return exported.Snapshot{}, fmt.Errorf("invalid total share count [%d], must be greater than or equal to [%d]",
			totalShareCount.Int64(),
			keyRequirement.MinTotalShareCount,
		)
	}

	corruptionThreshold := tss.ComputeAbsCorruptionThreshold(keyRequirement.SafetyThreshold, totalShareCount)
	if corruptionThreshold < 0 ||
		corruptionThreshold >= totalShareCount.Int64() {
		return exported.Snapshot{}, fmt.Errorf("invalid corruption threshold: %d, total share count: %d", corruptionThreshold, totalShareCount.Int64())
	}

	snapshot := exported.Snapshot{
		Validators:                 participants,
		Timestamp:                  ctx.BlockTime(),
		Height:                     ctx.BlockHeight(),
		TotalShareCount:            totalShareCount,
		Counter:                    counter,
		KeyShareDistributionPolicy: keyRequirement.KeyShareDistributionPolicy,
		CorruptionThreshold:        corruptionThreshold,
	}
	ctx.KVStore(k.storeKey).Set(counterKey(counter), k.cdc.MustMarshalLengthPrefixed(&snapshot))

	return snapshot, nil
}

func (k Keeper) setLatestCounter(ctx sdk.Context, counter int64) {
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, uint64(counter))

	ctx.KVStore(k.storeKey).Set([]byte(lastCounterKey), bz)
}

func counterKey(counter int64) []byte {
	return []byte(fmt.Sprintf("%s%d", counterPrefix, counter))
}

// RegisterProxy registers a proxy address for a given operator, which can broadcast messages in the principal's name
// The proxy will be marked as active and to be included in the next snapshot by default
func (k Keeper) RegisterProxy(ctx sdk.Context, operator sdk.ValAddress, proxy sdk.AccAddress) error {
	if val := k.staking.Validator(ctx, operator); val == nil {
		return fmt.Errorf("validator %s is unknown", operator.String())
	}

	key := []byte(proxyPrefix + operator.String())
	bz := ctx.KVStore(k.storeKey).Get(key)
	if bz == nil {
		return fmt.Errorf("no readiness notification found addressed to operator %s", operator.String())
	}
	if !bytes.Equal(bz, proxy.Bytes()) {
		return fmt.Errorf("proxy address mismatch (expected %s, actual %s)", proxy.String(), sdk.AccAddress(bz))
	}

	k.Logger(ctx).Debug("getting proxy count")
	count := k.getProxyCount(ctx)

	storedProxy := ctx.KVStore(k.storeKey).Get(operator)
	if storedProxy != nil {
		ctx.KVStore(k.storeKey).Delete(storedProxy)
		count--
	}
	k.Logger(ctx).Debug("setting proxy")
	ctx.KVStore(k.storeKey).Set(proxy, operator)
	// Creating a reverse lookup
	bz = append([]byte{1}, proxy...)
	ctx.KVStore(k.storeKey).Set(operator, bz)
	count++
	k.Logger(ctx).Debug("setting proxy count")
	k.setProxyCount(ctx, count)
	k.Logger(ctx).Debug("done")
	return nil
}

// DeactivateProxy deactivates the proxy address for a given operator
func (k Keeper) DeactivateProxy(ctx sdk.Context, operator sdk.ValAddress) error {
	val := k.staking.Validator(ctx, operator)
	if val == nil {
		return fmt.Errorf("validator %s is unknown", operator.String())
	}

	storedProxy := ctx.KVStore(k.storeKey).Get(operator)
	if storedProxy == nil {
		return fmt.Errorf("validator %s has no proxy registered", operator.String())
	}

	k.Logger(ctx).Debug(fmt.Sprintf("deactivating proxy %s", sdk.AccAddress(storedProxy[1:]).String()))
	bz := append([]byte{0}, storedProxy[1:]...)
	ctx.KVStore(k.storeKey).Set(operator, bz)

	return nil
}

// GetOperator returns the proxy address for a given principal address. Returns nil if not set.
func (k Keeper) GetOperator(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress {
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
	active = bz[0] == 1
	return addr, active
}

func (k Keeper) setProxyCount(ctx sdk.Context, count int) {
	k.Logger(ctx).Debug(fmt.Sprintf("number of known proxies: %v", count))
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, uint64(count))

	ctx.KVStore(k.storeKey).Set([]byte(proxyCountKey), bz)

}

func (k Keeper) getProxyCount(ctx sdk.Context) int {
	bz := ctx.KVStore(k.storeKey).Get([]byte(proxyCountKey))
	if bz == nil {
		return 0
	}

	return int(binary.LittleEndian.Uint64(bz))
}

// GetValidatorIllegibility returns the illegibility of the given validator
func (k Keeper) GetValidatorIllegibility(ctx sdk.Context, validator exported.SDKValidator) (exported.ValidatorIllegibility, error) {
	consAddr, err := validator.GetConsAddr()
	if err != nil {
		return exported.None, err
	}

	signedBlocksWindow := k.slasher.SignedBlocksWindow(ctx)
	signingInfo, signingInfoFound := k.slasher.GetValidatorSigningInfo(ctx, consAddr)
	_, hasProxyRegistered := k.GetProxy(ctx, validator.GetOperator())

	illegibility := exported.None

	if signingInfoFound && signingInfo.GetTombstoned() {
		illegibility |= exported.Tombstoned
	}

	if validator.IsJailed() {
		illegibility |= exported.Jailed
	}

	missedBlocks := utils.Threshold{Numerator: signingInfo.MissedBlocksCounter, Denominator: signedBlocksWindow}
	if missedBlocks.GTE(k.tss.GetMaxMissedBlocksPerWindow(ctx)) {
		illegibility |= exported.MissedTooManyBlocks
	}

	if !hasProxyRegistered {
		illegibility |= exported.NoProxyRegistered
	}

	if k.tss.GetTssSuspendedUntil(ctx, validator.GetOperator()) > ctx.BlockHeight() {
		illegibility |= exported.TssSuspended
	}

	return illegibility, nil
}
