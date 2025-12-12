package reward

import (
	"os"
	"strings"
	"time"

	"cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	multisigTypes "github.com/axelarnetwork/axelar-core/x/multisig/types"
	"github.com/axelarnetwork/axelar-core/x/reward/exported"
	"github.com/axelarnetwork/axelar-core/x/reward/types"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
)

// Activation times for the validator reward fix per chain.
// Can be overridden via VALIDATOR_REWARD_FIX_ACTIVATION_TIME environment variable (RFC3339 format).
const (
	// MainnetFixActivationTime is the activation time for mainnet (axelar-dojo-1)
	MainnetFixActivationTime = "2025-12-17T14:00:00Z"
	// TestnetFixActivationTime is the activation time for testnet (axelar-testnet-lisbon-3)
	TestnetFixActivationTime  = "2025-12-16T14:00:00Z"
	StagenetFixActivationTime = "2025-12-12T18:00:00Z"
	DevnetFixActivationTime   = "2025-12-12T16:00:00Z"
)

func getValidatorRewardFixActivationTime(chainID string) string {
	if envVal := os.Getenv("VALIDATOR_REWARD_FIX_ACTIVATION_TIME"); envVal != "" {
		return envVal
	}

	if strings.Contains(chainID, "devnet") {
		return DevnetFixActivationTime
	}
	if strings.HasPrefix(chainID, "axelar-stagenet") {
		return StagenetFixActivationTime
	}
	if strings.HasPrefix(chainID, "axelar-testnet") {
		return TestnetFixActivationTime
	}
	// Default to mainnet activation time for axelar-dojo-1 and any unknown chain
	return MainnetFixActivationTime
}

func isValidatorRewardFixActive(chainID string, blockTime time.Time) bool {
	activationTime, err := time.Parse(time.RFC3339, getValidatorRewardFixActivationTime(chainID))
	if err != nil {
		return true // if parsing fails, activate the fix
	}
	return !blockTime.Before(activationTime)
}

// EndBlocker is called at the end of every block, process external chain voting inflation
func EndBlocker(ctx sdk.Context, k types.Rewarder, n types.Nexus, m mintkeeper.Keeper, s types.Staker, slasher types.Slasher, msig types.MultiSig, ss types.Snapshotter) ([]abci.ValidatorUpdate, error) {
	handleExternalChainVotingInflation(ctx, k, n, m, s, slasher, ss)
	err := handleKeyMgmtInflation(ctx, k, m, s, slasher, msig, ss)

	return nil, err
}

func addRewardsByConsensusPower(ctx sdk.Context, s types.Staker, rewardPool exported.RewardPool, validators []snapshot.ValidatorI, totalReward sdk.DecCoin) {
	totalAmount := totalReward.Amount
	denom := totalReward.Denom

	validatorsWithConsensusPower := slices.Filter(validators, func(v snapshot.ValidatorI) bool {
		return v.GetConsensusPower(s.PowerReduction(ctx)) > 0
	})
	totalConsensusPower := slices.Reduce(validatorsWithConsensusPower, math.ZeroInt(), func(total math.Int, v snapshot.ValidatorI) math.Int {
		return total.AddRaw(v.GetConsensusPower(s.PowerReduction(ctx)))
	})

	slices.ForEach(validatorsWithConsensusPower, func(v snapshot.ValidatorI) {
		// Each validator receives reward weighted by consensus power
		amount := totalAmount.MulInt64(v.GetConsensusPower(s.PowerReduction(ctx))).QuoInt(totalConsensusPower).RoundInt()
		rewardPool.AddReward(
			toValAddress(v.GetOperator()),
			sdk.NewCoin(denom, amount),
		)
	})
}

func excludeJailedOrTombstoned(ctx sdk.Context, slasher types.Slasher, v snapshot.ValidatorI) bool {
	isTombstoned := func(v snapshot.ValidatorI) bool {
		consAdd, err := v.GetConsAddr()
		if err != nil {
			return true
		}

		return slasher.IsTombstoned(ctx, consAdd)
	}

	filter := funcs.Or(
		snapshot.ValidatorI.IsJailed,
		isTombstoned,
	)

	return filter(v)
}

func handleKeyMgmtInflation(ctx sdk.Context, k types.Rewarder, m mintkeeper.Keeper, s types.Staker, slasher types.Slasher, mSig types.MultiSig, ss types.Snapshotter) error {
	rewardPool := k.GetPool(ctx, multisigTypes.ModuleName)
	minter := funcs.Must(m.Minter.Get(ctx))
	mintParams := funcs.Must(m.Params.Get(ctx))
	totalAmount := minter.BlockProvision(mintParams).Amount.ToLegacyDec().Mul(k.GetParams(ctx).KeyMgmtRelativeInflationRate)

	isKeygenParticipant := func(v snapshot.ValidatorI) bool {
		proxy, isActive := ss.GetProxy(ctx, toValAddress(v.GetOperator()))

		return isActive && !mSig.HasOptedOut(ctx, proxy)
	}

	var validators []snapshot.ValidatorI

	validatorIterFn := func(_ int64, validator stakingtypes.ValidatorI) bool {
		if excludeJailedOrTombstoned(ctx, slasher, validator) {
			return false
		}

		if !isKeygenParticipant(validator) {
			return false
		}

		validators = append(validators, validator)

		return false
	}

	err := s.IterateBondedValidatorsByPower(ctx, validatorIterFn)
	if err != nil {
		return err
	}

	addRewardsByConsensusPower(ctx, s, rewardPool, validators, sdk.NewDecCoinFromDec(mintParams.MintDenom, totalAmount))

	return nil
}

func handleExternalChainVotingInflation(ctx sdk.Context, k types.Rewarder, n types.Nexus, m mintkeeper.Keeper, s types.Staker, slasher types.Slasher, ss types.Snapshotter) {
	totalStakingSupply := funcs.Must(m.StakingTokenSupply(ctx))
	blocksPerYear := funcs.Must(m.Params.Get(ctx)).BlocksPerYear
	inflationRate := k.GetParams(ctx).ExternalChainVotingInflationRate
	denom := funcs.Must(m.Params.Get(ctx)).MintDenom
	amountPerChain := totalStakingSupply.ToLegacyDec().Mul(inflationRate).QuoInt64(int64(blocksPerYear))

	for _, chain := range n.GetChains(ctx) {
		// ignore inactive chain
		if !n.IsChainActivated(ctx, chain) {
			continue
		}

		rewardPool := k.GetPool(ctx, chain.Name.String())
		maintainers := n.GetChainMaintainers(ctx, chain)
		if len(maintainers) == 0 {
			continue
		}

		var validators []snapshot.ValidatorI
		for _, maintainer := range maintainers {
			v, err := s.Validator(ctx, maintainer)
			if isValidatorRewardFixActive(ctx.ChainID(), ctx.BlockTime()) {
				if err != nil {
					continue
				}
			} else {
				if err == nil {
					continue
				}
			}

			if !v.IsBonded() {
				continue
			}

			if excludeJailedOrTombstoned(ctx, slasher, v) {
				continue
			}

			if _, isProxyActive := ss.GetProxy(ctx, toValAddress(v.GetOperator())); !isProxyActive {
				continue
			}

			validators = append(validators, v)
		}

		addRewardsByConsensusPower(ctx, s, rewardPool, validators, sdk.NewDecCoinFromDec(denom, amountPerChain))
	}
}

func toValAddress(addr string) sdk.ValAddress {
	return funcs.Must(sdk.ValAddressFromBech32(addr))
}
