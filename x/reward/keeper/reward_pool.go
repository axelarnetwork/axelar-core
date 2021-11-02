package keeper

import (
	"github.com/axelarnetwork/axelar-core/x/reward/exported"
	"github.com/axelarnetwork/axelar-core/x/reward/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

var _ exported.RewardPool = &rewardPool{}

type rewardPool struct {
	types.Pool
	ctx         sdk.Context
	k           Keeper
	banker      types.Banker
	distributor types.Distributor
	staker      types.Staker
}

func newPool(ctx sdk.Context, k Keeper, banker types.Banker, distributor types.Distributor, staker types.Staker, p types.Pool) *rewardPool {
	return &rewardPool{
		ctx:         ctx,
		k:           k,
		banker:      banker,
		distributor: distributor,
		staker:      staker,
		Pool:        p,
	}
}

func (p rewardPool) getRewards(validator sdk.ValAddress) (sdk.Coins, bool) {
	for _, reward := range p.Rewards {
		if reward.Validator.Equals(validator) {
			return reward.Coins, true
		}
	}

	return sdk.Coins{}, false
}

func (p *rewardPool) AddReward(validator sdk.ValAddress, coin sdk.Coin) {
	defer func() {
		p.k.Logger(p.ctx).Debug("adding rewards in pool", "pool", p.Name, "validator", validator.String(), "coin", coin.String())

		p.k.setPool(p.ctx, p.Pool)
	}()

	for i, reward := range p.Rewards {
		if reward.Validator.Equals(validator) {
			p.Rewards[i].Coins = reward.Coins.Add(coin)

			return
		}
	}

	p.Rewards = append(p.Rewards, &types.Pool_Reward{
		Validator: validator,
		Coins:     sdk.NewCoins(coin),
	})
}

func (p *rewardPool) ReleaseRewards(validator sdk.ValAddress) error {
	rewards, ok := p.getRewards(validator)
	if !ok {
		return nil
	}

	if err := p.banker.MintCoins(p.ctx, types.ModuleName, rewards); err != nil {
		return err
	}

	if err := p.banker.SendCoinsFromModuleToModule(p.ctx, types.ModuleName, distrtypes.ModuleName, rewards); err != nil {
		return err
	}

	p.k.Logger(p.ctx).Info("releasing rewards in pool", "pool", p.Name, "validator", validator.String())

	p.distributor.AllocateTokensToValidator(
		p.ctx,
		p.staker.Validator(p.ctx, validator),
		sdk.NewDecCoinsFromCoins(rewards...),
	)
	p.ClearRewards(validator)

	return nil
}

func (p *rewardPool) ClearRewards(validator sdk.ValAddress) {
	for i, reward := range p.Rewards {
		if reward.Validator.Equals(validator) {
			p.k.Logger(p.ctx).Info("clearing rewards in pool", "pool", p.Name, "validator", validator.String())

			p.Rewards = append(p.Rewards[:i], p.Rewards[i+1:]...)
			p.k.setPool(p.ctx, p.Pool)

			return
		}
	}
}
