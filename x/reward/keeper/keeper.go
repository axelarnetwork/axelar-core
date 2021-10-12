package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/reward/exported"
	"github.com/axelarnetwork/axelar-core/x/reward/types"
)

var (
	poolNamePrefix = utils.KeyFromStr("pool")
)

var _ types.Rewarder = Keeper{}

// Keeper provides access to all state changes regarding the reward module
type Keeper struct {
	storeKey    sdk.StoreKey
	cdc         codec.BinaryCodec
	paramSpace  paramtypes.Subspace
	banker      types.Banker
	distributor types.Distributor
	staker      types.Staker
}

// NewKeeper returns a new reward keeper
func NewKeeper(cdc codec.BinaryCodec, storeKey sdk.StoreKey, paramSpace paramtypes.Subspace, banker types.Banker, distributor types.Distributor, staker types.Staker) Keeper {
	return Keeper{
		cdc:         cdc,
		storeKey:    storeKey,
		paramSpace:  paramSpace.WithKeyTable(types.KeyTable()),
		banker:      banker,
		distributor: distributor,
		staker:      staker,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetParams returns the total set of reward parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)

	return params
}

// SetParams sets the total set of reward parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

// GetPool returns the reward pool of the given name, or returns an empty reward pool if not found
func (k Keeper) GetPool(ctx sdk.Context, name string) exported.RewardPool {
	var pool types.Pool
	key := poolNamePrefix.Append(utils.LowerCaseKey(name))
	ok := k.getStore(ctx).Get(key, &pool)
	if !ok {
		return newPool(ctx, k, k.banker, k.distributor, k.staker, types.NewPool(name))
	}

	return newPool(ctx, k, k.banker, k.distributor, k.staker, pool)
}

func (k Keeper) setPool(ctx sdk.Context, pool types.Pool) {
	key := poolNamePrefix.Append(utils.LowerCaseKey(pool.Name))
	k.getStore(ctx).Set(key, &pool)
}

func (k Keeper) getStore(ctx sdk.Context) utils.KVStore {
	return utils.NewNormalizedStore(ctx.KVStore(k.storeKey), k.cdc)
}

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

	p.k.Logger(p.ctx).Info(fmt.Sprintf("releasing rewards in pool %s for validator %s", p.Name, validator.String()))

	if err := p.banker.MintCoins(p.ctx, types.ModuleName, rewards); err != nil {
		return err
	}

	if err := p.banker.SendCoinsFromModuleToModule(p.ctx, types.ModuleName, distrtypes.ModuleName, rewards); err != nil {
		return err
	}

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
			p.Rewards = append(p.Rewards[:i], p.Rewards[i+1:]...)
			p.k.setPool(p.ctx, p.Pool)

			return
		}
	}
}
