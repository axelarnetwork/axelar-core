package keeper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/mock"
	"github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/snapshot/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	abci "github.com/tendermint/tendermint/abci/types"
	tmCrypto "github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/log"

	sdkExported "github.com/cosmos/cosmos-sdk/x/staking/exported"
	typesStaking "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	xParams "github.com/cosmos/cosmos-sdk/x/params"
	xStaking "github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/supply"
)

// Cases to test
var testCases = []struct {
	numValidators, totalPower int
}{
	{
		numValidators: 5,
		totalPower:    50,
	},
	{
		numValidators: 10,
		totalPower:    100,
	},
	{
		numValidators: 3,
		totalPower:    10,
	},
}

func init() {

	//Necessary if tests execute with the real sdk staking keeper
	cdc := testutils.Codec()
	cdc.RegisterConcrete(&mockPubKey{}, "mockPubKey", nil)
	cdc.RegisterInterface((*tmCrypto.PubKey)(nil), nil)
	cdc.RegisterConcrete("", "string", nil)
	typesStaking.RegisterCodec(cdc)

}

// Tests the snapshot functionality
func TestSnapshots(t *testing.T) {

	for i, params := range testCases {

		t.Run(fmt.Sprintf("Test-%d", i), makeTestKeeper(params.numValidators, params.totalPower))
	}

}

func makeTestKeeper(numValidators, totalPower int) func(t *testing.T) {

	return func(t *testing.T) {

		ctx := sdk.NewContext(mock.NewMultiStore(), abci.Header{}, false, log.TestingLogger())

		cdc := testutils.Codec()

		validators := genValidators(t, numValidators, totalPower)

		staking := NewTestKeeper(validators...)

		// If I use the real sdk staking keeper, the tests block and I have no idea why.
		// Therefore, the tests will use the mock keeper above, at least for now
		// TODO: figure it out
		/*staking := getStakingKeeper(cdc)
		staking.SetLastTotalPower(ctx, sdk.NewInt(int64(totalPower)))
		for _, v := range validators {
			staking.SetValidator(ctx, v)
		}*/

		assert.Equal(t, sdk.NewInt(int64(totalPower)), staking.GetLastTotalPower(ctx))

		keeper := NewKeeper(cdc, mock.NewKVStoreKey("staking"), staking)

		iterated := make([]exported.Validator, 0)

		keeper.IterateValidators(ctx, func(index int64, validator exported.Validator) (stop bool) {

			iterated = append(iterated, validator)

			return false
		})

		assertValidators(t, ctx, numValidators, staking, keeper, iterated)

		_, ok := keeper.GetSnapshot(ctx, 0)

		assert.False(t, ok)
		assert.Equal(t, keeper.GetLatestRound(ctx), int64(-1))

		_, ok = keeper.GetLatestSnapshot(ctx)

		assert.False(t, ok)

		err := keeper.TakeSnapshot(ctx)

		assert.NoError(t, err)

		snapshot, ok := keeper.GetSnapshot(ctx, 0)

		assert.True(t, ok)
		assert.Equal(t, keeper.GetLatestRound(ctx), int64(0))
		assertValidators(t, ctx, numValidators, staking, keeper, snapshot.Validators)

		err = keeper.TakeSnapshot(ctx)

		assert.Error(t, err)

		ctx = ctx.WithBlockTime(ctx.BlockTime().Add(interval + 100))

		err = keeper.TakeSnapshot(ctx)

		assert.NoError(t, err)

		snapshot, ok = keeper.GetSnapshot(ctx, 1)

		assert.True(t, ok)
		assert.Equal(t, keeper.GetLatestRound(ctx), int64(1))
		assertValidators(t, ctx, numValidators, staking, keeper, snapshot.Validators)

	}
}

// auxiliary function for the snapshot test unit
func assertValidators(t *testing.T, ctx sdk.Context, numValidators int, staking types.StakingKeeper, keeper Keeper, validators []exported.Validator) {

	assert.Equal(t, numValidators, len(validators))

	for _, val := range validators {

		v1, ok := staking.GetValidator(ctx, val.Address)

		assert.True(t, ok)
		assert.Equal(t, val.Address, v1.GetOperator())
		assert.Equal(t, val.Power, v1.GetConsensusPower())

		v2, ok := keeper.Validator(ctx, val.Address)

		assert.True(t, ok)
		assert.Equal(t, val.Address, v2.Address)
		assert.Equal(t, val.Power, v2.Power)
	}
}

// This function returns a set of validators whose voting power adds up to the specified total power
func genValidators(t *testing.T, numValidators, totalConsPower int) (validators []typesStaking.Validator) {

	t.Logf("Total Power: %v", totalConsPower)

	validators = make([]typesStaking.Validator, numValidators)

	quotient, remainder := totalConsPower/numValidators, totalConsPower%numValidators

	for i := 0; i < numValidators; i++ {

		desc := typesStaking.Description{
			Moniker: fmt.Sprintf("TestValidator-%d", i),
		}

		pk := newKeysPair()

		val := typesStaking.NewValidator(sdk.ValAddress(pk.PubKey().Bytes()), pk.PubKey(), desc)

		tokens := sdk.TokensFromConsensusPower(int64(quotient))

		if i == 0 {
			tokens = tokens.Add(sdk.TokensFromConsensusPower(int64(remainder)))

		}

		val, _ = val.AddTokensFromDel(tokens)

		val = val.UpdateStatus(sdk.Bonded)

		t.Logf("Tokens for %s: %v", val.GetMoniker(), val.GetTokens())
		t.Logf("Consensus Power for for %s: %v", val.GetMoniker(), val.GetConsensusPower())
		t.Logf("Is %s bonded?: %v", val.GetMoniker(), val.IsBonded())

		validators[i] = val

	}

	return
}

// Mocks for the crypto.PubKey and crypto PrivKey interfaces used
// when instantiating validators

const mockKeysSize = 32

type mockPrivKey struct {
	dummyPubKey  []byte
	dummyPrivKey []byte
}

func newKeysPair() tmCrypto.PrivKey {

	dummyPubKey := make([]byte, mockKeysSize)
	dummyPrivKey := make([]byte, mockKeysSize)

	rand.Read(dummyPubKey)
	rand.Read(dummyPrivKey)

	return &mockPrivKey{
		dummyPrivKey: dummyPrivKey,
		dummyPubKey:  dummyPubKey,
	}

}

func (p *mockPrivKey) Bytes() []byte {
	bz, _ := json.Marshal(p.dummyPrivKey)
	return bz

}
func (p *mockPrivKey) Sign(msg []byte) ([]byte, error) {

	hasher := fnv.New128()
	hasher.Write(p.dummyPubKey)
	return hasher.Sum(msg), nil

}

func (p *mockPrivKey) PubKey() tmCrypto.PubKey {

	return &mockPubKey{
		dummyPubKey: p.dummyPubKey,
	}

}
func (p *mockPrivKey) Equals(key tmCrypto.PrivKey) bool {
	return bytes.Equal(p.Bytes(), key.Bytes())
}

type mockPubKey struct {
	dummyPubKey []byte
}

func (p *mockPubKey) Address() tmCrypto.Address {
	hasher := fnv.New128()
	hasher.Write(p.Bytes())
	return hasher.Sum(nil)

}

func (p *mockPubKey) Bytes() []byte {

	bz, _ := json.Marshal(p.dummyPubKey)
	return bz

}
func (p *mockPubKey) VerifyBytes(msg []byte, sig []byte) bool {

	hasher := fnv.New128()
	hasher.Write(p.dummyPubKey)
	hash := hasher.Sum(msg)

	return bytes.Equal(sig, hash)

}
func (p *mockPubKey) Equals(key tmCrypto.PubKey) bool {

	return bytes.Equal(p.Bytes(), key.Bytes())

}

// This function is used to instantiate a real staking keeper,
// but the tests block if this is used. The code is based on
// what is used in app.go
func getStakingKeeper(cdc *codec.Codec) xStaking.Keeper {

	maccPerms := map[string][]string{
		auth.FeeCollectorName:      nil,
		distr.ModuleName:           nil,
		xStaking.BondedPoolName:    {supply.Burner, supply.Staking},
		xStaking.NotBondedPoolName: {supply.Burner, supply.Staking},
	}

	keys := sdk.NewKVStoreKeys(xStaking.StoreKey, supply.StoreKey, auth.StoreKey, xParams.StoreKey)

	tkeys := sdk.NewTransientStoreKeys(xStaking.TStoreKey, xParams.TStoreKey)

	paramsKeeper := xParams.NewKeeper(cdc, keys[xParams.StoreKey], tkeys[xParams.TStoreKey])

	authSubspace := paramsKeeper.Subspace(auth.DefaultParamspace)
	bankSubspace := paramsKeeper.Subspace(bank.DefaultParamspace)
	stakingSubspace := paramsKeeper.Subspace(xStaking.DefaultParamspace)

	accountKeeper := auth.NewAccountKeeper(
		cdc,
		keys[auth.StoreKey],
		authSubspace,
		auth.ProtoBaseAccount,
	)

	bankKeeper := bank.NewBaseKeeper(
		accountKeeper,
		bankSubspace,
		moduleAccountAddrs(maccPerms),
	)

	supplyKeeper := supply.NewKeeper(
		cdc,
		keys[supply.StoreKey],
		accountKeeper,
		bankKeeper,
		maccPerms,
	)

	return xStaking.NewKeeper(
		cdc,
		keys[xStaking.StoreKey],
		supplyKeeper,
		stakingSubspace,
	)
}

func moduleAccountAddrs(maccPerms map[string][]string) map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[supply.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// The code above blocks the tests, so for the moment we use
// a mock staking keeper, which works well with these test units

var _ types.StakingKeeper = mockKeeper{}

type mockKeeper struct {
	validators []typesStaking.Validator
	totalPower sdk.Int
}

func NewTestKeeper(validators ...typesStaking.Validator) types.StakingKeeper {

	keeper := mockKeeper{
		make([]typesStaking.Validator, 0),
		sdk.NewInt(0),
	}

	for _, val := range validators {

		keeper.validators = append(keeper.validators, val)
		keeper.totalPower = keeper.totalPower.AddRaw(val.GetConsensusPower())
	}

	return keeper

}

func (k mockKeeper) GetLastTotalPower(_ sdk.Context) (power sdk.Int) {

	return k.totalPower
}

func (k mockKeeper) IterateValidators(_ sdk.Context, fn func(index int64, validator sdkExported.ValidatorI) (stop bool)) {

	for i, val := range k.validators {

		fn(int64(i), val)

	}
}

func (k mockKeeper) GetValidator(_ sdk.Context, addr sdk.ValAddress) (validator typesStaking.Validator, found bool) {

	found = false

	for _, validator = range k.validators {

		if bytes.Equal(validator.GetOperator().Bytes(), addr.Bytes()) {

			found = true
			break

		}

	}

	return
}
