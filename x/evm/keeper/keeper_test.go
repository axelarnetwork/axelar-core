package keeper

import (
	"testing"

	params "github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestSetBurnerInfoGetBurnerInfo(t *testing.T) {
	var (
		ctx    sdk.Context
		keeper Keeper
	)

	setup := func() {
		encCfg := params.MakeEncodingConfig()
		subspace := paramstypes.NewSubspace(encCfg.Marshaler, encCfg.Amino, sdk.NewKVStoreKey("params"), sdk.NewKVStoreKey("tparams"), "eth")
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
		keeper = NewKeeper(encCfg.Marshaler, sdk.NewKVStoreKey("eth"), subspace)
	}

	t.Run("should set and get the burner info", testutils.Func(func(t *testing.T) {
		setup()

		burnerInfo := types.BurnerInfo{
			TokenAddress: types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength))),
			Symbol:       rand.StrBetween(2, 5),
			Salt:         types.Hash(common.BytesToHash(rand.Bytes(common.HashLength))),
		}
		burnerAddress := common.BytesToAddress(rand.Bytes(common.AddressLength))

		keeper.SetBurnerInfo(ctx, burnerAddress, &burnerInfo)
		actual := keeper.GetBurnerInfo(ctx, burnerAddress)

		assert.NotNil(t, actual)
		assert.Equal(t, *actual, burnerInfo)
	}).Repeat(20))

}
