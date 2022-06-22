package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/keeper"
	"github.com/axelarnetwork/axelar-core/x/vote/types"
	"github.com/axelarnetwork/axelar-core/x/vote/types/mock"
	. "github.com/axelarnetwork/utils/test"
)

func setup() (sdk.Context, keeper.Keeper, *mock.SnapshotterMock, *mock.StakingKeeperMock, *mock.RewarderMock) {
	snapshotter := mock.SnapshotterMock{}
	staking := mock.StakingKeeperMock{}
	rewarder := mock.RewarderMock{}

	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	encodingConfig := params.MakeEncodingConfig()
	types.RegisterLegacyAminoCodec(encodingConfig.Amino)
	types.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	subspace := paramstypes.NewSubspace(encodingConfig.Codec, encodingConfig.Amino, sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "vote")

	keeper := keeper.NewKeeper(
		encodingConfig.Codec,
		sdk.NewKVStoreKey(types.StoreKey),
		subspace,
		&snapshotter,
		&staking,
		&rewarder,
	)
	keeper.SetParams(ctx, types.DefaultParams())

	return ctx, keeper, &snapshotter, &staking, &rewarder
}

func TestKeeper(t *testing.T) {
	var (
		ctx     sdk.Context
		k       keeper.Keeper
		staking *mock.StakingKeeperMock
	)

	givenKeeper := Given("vote keeper", func() {
		ctx, k, _, staking, _ = setup()
	})

	repeats := 20

	t.Run("InitializePoll", testutils.Func(func(t *testing.T) {
		var (
			voters []sdk.ValAddress
		)

		givenVotersWith := func(voterCount int) GivenStatement {
			return Given("having voters", func() {
				voters = make([]sdk.ValAddress, voterCount)
				for i := 0; i < voterCount; i++ {
					voters[i] = rand.ValAddr()
				}
			})
		}

		givenKeeper.
			Given2(givenVotersWith(20)).
			When("voters are not validators", func() {
				staking.ValidatorFunc = func(ctx sdk.Context, addr sdk.ValAddress) stakingtypes.ValidatorI { return nil }
			}).
			Then("should return error", func(t *testing.T) {
				_, err := k.InitializePoll(
					ctx,
					voters,
					exported.ExpiryAt(rand.PosI64()),
					exported.ModuleMetadata(rand.NormalizedStr(5)),
				)

				assert.ErrorContains(t, err, "no voters set")
			}).
			Run(t)

		givenKeeper.
			Given2(givenVotersWith(20)).
			When("voters do not have consensus power", func() {
				staking.ValidatorFunc = func(ctx sdk.Context, addr sdk.ValAddress) stakingtypes.ValidatorI {
					if rand.Bools(0.5).Next() {
						return stakingtypes.Validator{Status: rand.Of(stakingtypes.Unbonded, stakingtypes.Unbonding), Tokens: sdk.NewInt(rand.PosI64())}
					}

					return stakingtypes.Validator{Status: stakingtypes.Bonded, Tokens: sdk.ZeroInt()}
				}
				staking.PowerReductionFunc = func(context sdk.Context) sdk.Int { return sdk.NewInt(10) }
			}).
			Then("should return error", func(t *testing.T) {
				_, err := k.InitializePoll(
					ctx,
					voters,
					exported.ExpiryAt(rand.PosI64()),
					exported.ModuleMetadata(rand.NormalizedStr(5)),
				)

				assert.ErrorContains(t, err, "no voters set")
			}).
			Run(t)
	}).Repeat(repeats))
}
