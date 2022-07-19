package reward

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	abci "github.com/tendermint/tendermint/abci/types"

	multisigTypes "github.com/axelarnetwork/axelar-core/x/multisig/types"
	"github.com/axelarnetwork/axelar-core/x/reward/exported"
	"github.com/axelarnetwork/axelar-core/x/reward/types"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
)

// BeginBlocker is called at the beginning of every block
func BeginBlocker(ctx sdk.Context, _ abci.RequestBeginBlock, _ types.Rewarder) {}

// EndBlocker is called at the end of every block, process external chain voting inflation
func EndBlocker(ctx sdk.Context, _ abci.RequestEndBlock, k types.Rewarder, n types.Nexus, m types.Minter, s types.Staker, slasher types.Slasher, msig types.MultiSig, ss types.Snapshotter) []abci.ValidatorUpdate {
	handleExternalChainVotingInflation(ctx, k, n, m, s)
	handleKeyMgmtInflation(ctx, k, m, s, slasher, msig, ss)

	return nil
}

func addRewardsByConsensusPower(ctx sdk.Context, s types.Staker, rewardPool exported.RewardPool, validators []stakingtypes.Validator, totalReward sdk.DecCoin) {
	totalAmount := totalReward.Amount
	denom := totalReward.Denom

	validatorsWithConsensusPower := slices.Filter(validators, func(v stakingtypes.Validator) bool {
		return v.GetConsensusPower(s.PowerReduction(ctx)) > 0
	})
	totalConsensusPower := slices.Reduce(validatorsWithConsensusPower, sdk.ZeroInt(), func(total sdk.Int, v stakingtypes.Validator) sdk.Int {
		return total.AddRaw(v.GetConsensusPower(s.PowerReduction(ctx)))
	})

	slices.ForEach(validatorsWithConsensusPower, func(v stakingtypes.Validator) {
		// Each validator receives reward weighted by consensus power
		amount := totalAmount.MulInt64(v.GetConsensusPower(s.PowerReduction(ctx))).QuoInt(totalConsensusPower).RoundInt()
		rewardPool.AddReward(
			v.GetOperator(),
			sdk.NewCoin(denom, amount),
		)
	})

}

func handleKeyMgmtInflation(ctx sdk.Context, k types.Rewarder, m types.Minter, s types.Staker, slasher types.Slasher, mSig types.MultiSig, ss types.Snapshotter) {
	rewardPool := k.GetPool(ctx, multisigTypes.ModuleName)
	minter := m.GetMinter(ctx)
	mintParams := m.GetParams(ctx)
	totalAmount := minter.BlockProvision(mintParams).Amount.ToDec().Mul(k.GetParams(ctx).KeyMgmtRelativeInflationRate)

	var validators []stakingtypes.Validator

	validatorIterFn := func(_ int64, v stakingtypes.ValidatorI) bool {
		validator, ok := v.(stakingtypes.Validator)
		if !ok {
			return false
		}

		filter := funcs.And(
			funcs.Not(snapshot.ValidatorI.IsJailed),
			funcs.Not(snapshot.IsTombstoned(ctx, slasher)),
			snapshot.IsProxyActive(ctx, ss.GetProxy),
		)

		if !filter(validator) {
			return false
		}

		validators = append(validators, validator)

		return false
	}
	s.IterateBondedValidatorsByPower(ctx, validatorIterFn)

	addRewardsByConsensusPower(ctx, s, rewardPool, validators, sdk.NewDecCoinFromDec(mintParams.MintDenom, totalAmount))
}

func handleExternalChainVotingInflation(ctx sdk.Context, k types.Rewarder, n types.Nexus, m types.Minter, s types.Staker) {
	totalStakingSupply := m.StakingTokenSupply(ctx)
	blocksPerYear := m.GetParams(ctx).BlocksPerYear
	inflationRate := k.GetParams(ctx).ExternalChainVotingInflationRate
	denom := m.GetParams(ctx).MintDenom
	amountPerChain := totalStakingSupply.ToDec().Mul(inflationRate).QuoInt64(int64(blocksPerYear))

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

		var validators []stakingtypes.Validator
		for _, maintainer := range maintainers {
			v := s.Validator(ctx, maintainer)
			if v == nil {
				continue
			}

			validators = append(validators, v.(stakingtypes.Validator))
		}

		addRewardsByConsensusPower(ctx, s, rewardPool, validators, sdk.NewDecCoinFromDec(denom, amountPerChain))
	}
}
