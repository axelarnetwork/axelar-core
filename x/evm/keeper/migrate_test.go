package keeper_test

import (
	"github.com/axelarnetwork/axelar-core/x/evm/exported"
	"github.com/axelarnetwork/axelar-core/x/evm/keeper"
	"github.com/axelarnetwork/axelar-core/x/evm/types/mock"
	"github.com/axelarnetwork/axelar-core/x/evm/types/testutils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/stretchr/testify/assert"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramsKeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	. "github.com/axelarnetwork/utils/test"
)

func TestGetMigrationHandler(t *testing.T) {
	var (
		ctx     sdk.Context
		k       *keeper.BaseKeeper
		handler func(ctx sdk.Context) error
		err         error
		expectedIDs []types.CommandID
	)

	cmdTypes := map[types.CommandType]string{
		types.COMMAND_TYPE_MINT_TOKEN:                      "mintToken",
		types.COMMAND_TYPE_DEPLOY_TOKEN:                    "deployToken",
		types.COMMAND_TYPE_BURN_TOKEN:                      "burnToken",
		types.COMMAND_TYPE_TRANSFER_OPERATORSHIP:           "transferOperatorship",
		types.COMMAND_TYPE_APPROVE_CONTRACT_CALL_WITH_MINT: "approveContractCallWithMint",
		types.COMMAND_TYPE_APPROVE_CONTRACT_CALL:           "approveContractCall",
	}

	Given("a context", func() {
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	}).
		Given("a keeper", func() {
			encCfg := params.MakeEncodingConfig()
			pk := paramsKeeper.NewKeeper(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("params"), sdk.NewKVStoreKey("tparams"))
			k = keeper.NewKeeper(encCfg.Codec, sdk.NewKVStoreKey(types.StoreKey), pk)
			k.InitChains(ctx)
			funcs.MustNoErr(k.CreateChain(ctx, types.DefaultParams()[0]))
		}).
		Given("a migration handler", func() {
			n := &mock.NexusMock{
				GetChainsFunc: func(sdk.Context) []nexus.Chain { return []nexus.Chain{exported.Ethereum} },
			}
			handler = keeper.Migrate6To7(k, n)
		}).
		Given("there are old commands", func() {
			ck := funcs.Must(k.ForChain(ctx, exported.Ethereum.Name))

			expectedIDs = []types.CommandID{}
			for i := 0; i < 3; i++ {
				cmd := testutils.RandomCommand()
				cmd.Command = cmdTypes[cmd.Type]
				cmd.Type = types.COMMAND_TYPE_UNSPECIFIED
				expectedIDs = append(expectedIDs, cmd.ID)
				funcs.MustNoErr(ck.EnqueueCommand(ctx, cmd))
			}

			_ = funcs.Must(ck.CreateNewBatchToSign(ctx))
		}).
		Given("there are commands in the queue", func() {
			ck := funcs.Must(k.ForChain(ctx, exported.Ethereum.Name))

			for i := 0; i < 2; i++ {
				cmd := testutils.RandomCommand()
				cmd.Command = cmdTypes[cmd.Type]
				cmd.Type = types.COMMAND_TYPE_UNSPECIFIED
				expectedIDs = append(expectedIDs, cmd.ID)
				funcs.MustNoErr(ck.EnqueueCommand(ctx, cmd))
			}
		}).
		When("calling migration", func() {
			err = handler(ctx)
		}).
		Then("it succeeds", func(t *testing.T) {
			assert.NoError(t, err)
		}).
		Then("all commands have been migrated", func(t *testing.T) {
			ck := funcs.Must(k.ForChain(ctx, exported.Ethereum.Name))

			for _, id := range expectedIDs {
				cmd, ok := ck.GetCommand(ctx, id)
				assert.True(t, ok)

				assert.NotEqual(t, types.COMMAND_TYPE_UNSPECIFIED, cmd.Type)
			}
		}).Run(t, 20)
}
