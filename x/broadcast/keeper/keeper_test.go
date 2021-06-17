package keeper

import (
	"fmt"
	"testing"

	appParams "github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	broadcastMock "github.com/axelarnetwork/axelar-core/x/broadcast/types/mock"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestKeeper_RegisterProxy(t *testing.T) {
	var (
		ctx              sdk.Context
		keeper           Keeper
		principalAddress sdk.ValAddress
		staker           broadcastMock.StakerMock
	)

	setup := func() {
		encCfg := appParams.MakeEncodingConfig()
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
		principalAddress = sdk.ValAddress(rand.Bytes(sdk.AddrLen))

		staker = broadcastMock.StakerMock{
			ValidatorFunc: func(_ sdk.Context, addr sdk.ValAddress) stakingTypes.ValidatorI {

				if principalAddress.Equals(addr) {
					return stakingTypes.Validator{
						OperatorAddress: sdk.ValAddress(rand.Bytes(sdk.AddrLen)).String(),
						Tokens:          sdk.TokensFromConsensusPower(rand.I64Between(1, 100000)),
						Status:          stakingTypes.Bonded,
						ConsensusPubkey: nil,
					}
				}
				return nil
			},
		}

		keeper = NewKeeper(encCfg.Amino, sdk.NewKVStoreKey("broadcast"), &staker)
	}
	t.Run("happy path", testutils.Func(func(t *testing.T) {
		setup()

		proxy := sdk.AccAddress(rand.Bytes(sdk.AddrLen))
		err := keeper.RegisterProxy(ctx, principalAddress, proxy)

		assert.NoError(t, err)
		assert.Equal(t, 1, len(staker.ValidatorCalls()))
		assert.True(t, principalAddress.Equals(staker.ValidatorCalls()[0].Addr))

	}).Repeat(20))

	t.Run("unknown validator", testutils.Func(func(t *testing.T) {
		setup()

		address := sdk.ValAddress(rand.Bytes(sdk.AddrLen))
		proxy := sdk.AccAddress(rand.Bytes(sdk.AddrLen))
		err := keeper.RegisterProxy(ctx, address, proxy)

		assert.Error(t, err)
		assert.Equal(t, 1, len(staker.ValidatorCalls()))
		assert.True(t, address.Equals(staker.ValidatorCalls()[0].Addr))

	}).Repeat(20))
}

func TestKeeper_DeregisterProxy(t *testing.T) {
	var (
		ctx              sdk.Context
		keeper           Keeper
		principalAddress sdk.ValAddress
		proxy            sdk.AccAddress
		staker           broadcastMock.StakerMock
	)

	setup := func() {
		encCfg := appParams.MakeEncodingConfig()
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
		principalAddress = sdk.ValAddress(rand.Bytes(sdk.AddrLen))
		proxy = sdk.AccAddress(rand.Bytes(sdk.AddrLen))

		staker = broadcastMock.StakerMock{
			ValidatorFunc: func(_ sdk.Context, addr sdk.ValAddress) stakingTypes.ValidatorI {

				if principalAddress.Equals(addr) {
					return stakingTypes.Validator{
						OperatorAddress: sdk.ValAddress(rand.Bytes(sdk.AddrLen)).String(),
						Tokens:          sdk.TokensFromConsensusPower(rand.I64Between(1, 100000)),
						Status:          stakingTypes.Bonded,
						ConsensusPubkey: nil,
					}
				}
				return nil
			},
		}

		keeper = NewKeeper(encCfg.Amino, sdk.NewKVStoreKey("broadcast"), &staker)
		if err := keeper.RegisterProxy(ctx, principalAddress, proxy); err != nil {
			panic(fmt.Sprintf("setup failed for unit test: %v", err))
		}
	}
	t.Run("happy path", testutils.Func(func(t *testing.T) {
		setup()

		err := keeper.DeregisterProxy(ctx, principalAddress)

		assert.NoError(t, err)
		assert.Equal(t, 2, len(staker.ValidatorCalls()))
		assert.True(t, principalAddress.Equals(staker.ValidatorCalls()[1].Addr))

	}).Repeat(20))

	t.Run("unknown validator", testutils.Func(func(t *testing.T) {
		setup()

		address := sdk.ValAddress(rand.Bytes(sdk.AddrLen))
		err := keeper.DeregisterProxy(ctx, address)

		assert.Error(t, err)
		assert.Equal(t, 2, len(staker.ValidatorCalls()))
		assert.True(t, address.Equals(staker.ValidatorCalls()[1].Addr))

	}).Repeat(20))

	t.Run("no proxy", testutils.Func(func(t *testing.T) {
		setup()

		staker.ValidatorFunc = func(_ sdk.Context, addr sdk.ValAddress) stakingTypes.ValidatorI { return nil }

		err := keeper.DeregisterProxy(ctx, principalAddress)

		assert.Error(t, err)
		assert.Equal(t, 2, len(staker.ValidatorCalls()))
		assert.True(t, principalAddress.Equals(staker.ValidatorCalls()[1].Addr))

	}).Repeat(20))
}
