package reward_test

import (
	"context"
	"errors"
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	rand2 "github.com/axelarnetwork/axelar-core/testutils/rand"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/reward"
	"github.com/axelarnetwork/axelar-core/x/reward/exported"
	exportedmock "github.com/axelarnetwork/axelar-core/x/reward/exported/mock"
	"github.com/axelarnetwork/axelar-core/x/reward/types"
	"github.com/axelarnetwork/axelar-core/x/reward/types/mock"
	"github.com/axelarnetwork/utils/funcs"
)

// TestEndBlocker_ExternalChainVotingInflation tests the reward distribution for external chain maintainers.
func TestEndBlocker_ExternalChainVotingInflation(t *testing.T) {
	t.Run("rewards chain maintainers when validator lookup succeeds", func(t *testing.T) {
		s := newEndBlockerTestSetup(t)

		consKey := ed25519.GenPrivKey().PubKey()
		validator := funcs.Must(stakingtypes.NewValidator(s.maintainer.String(), consKey, stakingtypes.Description{}))
		validator.Status = stakingtypes.Bonded
		validator.Tokens = math.NewInt(1000000000000)

		s.staker.ValidatorFunc = func(ctx context.Context, addr sdk.ValAddress) (stakingtypes.ValidatorI, error) {
			return validator, nil
		}

		err := s.runEndBlocker()

		assert.NoError(t, err)
		assert.NotEmpty(t, s.rewardPool.AddRewardCalls(), "bonded chain maintainers should receive external chain voting inflation rewards")
	})

	t.Run("skips maintainer when validator lookup fails", func(t *testing.T) {
		s := newEndBlockerTestSetup(t)

		s.staker.ValidatorFunc = func(ctx context.Context, addr sdk.ValAddress) (stakingtypes.ValidatorI, error) {
			return nil, errors.New("validator not found")
		}

		err := s.runEndBlocker()

		assert.NoError(t, err)
		assert.Empty(t, s.rewardPool.AddRewardCalls(), "maintainers with failed validator lookups should be skipped without panic")
	})
}

type endBlockerTestSetup struct {
	ctx         sdk.Context
	mintK       mintkeeper.Keeper
	staker      *mock.StakerMock
	rewardPool  *exportedmock.RewardPoolMock
	rewarder    *mock.RewarderMock
	nexusKeeper *mock.NexusMock
	slasher     *mock.SlasherMock
	msig        *mock.MultiSigMock
	snapshotter *mock.SnapshotterMock
	maintainer  sdk.ValAddress
}

func newEndBlockerTestSetup(t *testing.T) *endBlockerTestSetup {
	encCfg := app.MakeEncodingConfig()
	store := fake.NewMultiStore()
	ctx := sdk.NewContext(store, tmproto.Header{Height: 100, ChainID: "axelar-dojo-1"}, false, log.NewTestLogger(t))

	maintainer := rand2.ValAddr()
	chain := nexus.Chain{Name: nexus.ChainName("ethereum")}

	staker := &mock.StakerMock{
		PowerReductionFunc: func(ctx context.Context) math.Int {
			return math.NewInt(1000000)
		},
		IterateBondedValidatorsByPowerFunc: func(ctx context.Context, fn func(index int64, validator stakingtypes.ValidatorI) (stop bool)) error {
			return nil
		},
		StakingTokenSupplyFunc: func(ctx context.Context) (math.Int, error) {
			return math.NewInt(1000000000000), nil
		},
		BondedRatioFunc: func(ctx context.Context) (math.LegacyDec, error) {
			return math.LegacyMustNewDecFromStr("0.5"), nil
		},
	}

	accK := &mock.AccountKeeperMock{
		GetModuleAddressFunc: func(string) sdk.AccAddress { return authtypes.NewModuleAddress(minttypes.ModuleName) },
	}
	mintK := mintkeeper.NewKeeper(encCfg.Codec, runtime.NewKVStoreService(storetypes.NewKVStoreKey("mint")), staker, accK, nil, authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String())

	funcs.MustNoErr(mintK.Minter.Set(ctx, minttypes.Minter{Inflation: math.LegacyMustNewDecFromStr("0.1")}))
	funcs.MustNoErr(mintK.Params.Set(ctx, minttypes.Params{
		MintDenom:           "uaxl",
		BlocksPerYear:       6311520,
		InflationRateChange: math.LegacyMustNewDecFromStr("0.13"),
		InflationMax:        math.LegacyMustNewDecFromStr("0.20"),
		InflationMin:        math.LegacyMustNewDecFromStr("0.07"),
		GoalBonded:          math.LegacyMustNewDecFromStr("0.67"),
	}))

	rewardPool := &exportedmock.RewardPoolMock{
		AddRewardFunc: func(valAddress sdk.ValAddress, coin sdk.Coin) {},
	}

	rewarder := &mock.RewarderMock{
		GetParamsFunc: func(ctx sdk.Context) types.Params {
			return types.Params{
				KeyMgmtRelativeInflationRate:     math.LegacyMustNewDecFromStr("0.01"),
				ExternalChainVotingInflationRate: math.LegacyMustNewDecFromStr("0.01"),
			}
		},
		GetPoolFunc: func(ctx sdk.Context, name string) exported.RewardPool {
			return rewardPool
		},
	}

	nexusKeeper := &mock.NexusMock{
		GetChainsFunc: func(ctx sdk.Context) []nexus.Chain {
			return []nexus.Chain{chain}
		},
		IsChainActivatedFunc: func(ctx sdk.Context, c nexus.Chain) bool {
			return true
		},
		GetChainMaintainersFunc: func(ctx sdk.Context, c nexus.Chain) []sdk.ValAddress {
			return []sdk.ValAddress{maintainer}
		},
	}

	slasher := &mock.SlasherMock{
		IsTombstonedFunc: func(ctx context.Context, consAddr sdk.ConsAddress) bool {
			return false
		},
	}

	msig := &mock.MultiSigMock{
		HasOptedOutFunc: func(ctx sdk.Context, participant sdk.AccAddress) bool {
			return false
		},
	}

	snapshotter := &mock.SnapshotterMock{
		GetProxyFunc: func(ctx sdk.Context, operator sdk.ValAddress) (sdk.AccAddress, bool) {
			return rand2.AccAddr(), true
		},
	}

	return &endBlockerTestSetup{
		ctx:         ctx,
		mintK:       mintK,
		staker:      staker,
		rewardPool:  rewardPool,
		rewarder:    rewarder,
		nexusKeeper: nexusKeeper,
		slasher:     slasher,
		msig:        msig,
		snapshotter: snapshotter,
		maintainer:  maintainer,
	}
}

func (s *endBlockerTestSetup) runEndBlocker() error {
	_, err := reward.EndBlocker(s.ctx, s.rewarder, s.nexusKeeper, s.mintK, s.staker, s.slasher, s.msig, s.snapshotter)
	return err
}
