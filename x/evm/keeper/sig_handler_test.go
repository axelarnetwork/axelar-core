package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	gogoprototypes "github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app/params"
	fakeMock "github.com/axelarnetwork/axelar-core/testutils/fake/interfaces/mock"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/evm/exported"
	"github.com/axelarnetwork/axelar-core/x/evm/keeper"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/evm/types/mock"
	multisig "github.com/axelarnetwork/axelar-core/x/multisig/exported"
	multisigtestutils "github.com/axelarnetwork/axelar-core/x/multisig/types/testutils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/funcs"
	. "github.com/axelarnetwork/utils/test"
)

func setup2() (sdk.Context, *mock.BaseKeeperMock, *mock.ChainKeeperMock, multisig.SigHandler) {
	ctx := sdk.NewContext(&fakeMock.MultiStoreMock{}, tmproto.Header{}, false, log.TestingLogger())
	chaink := &mock.ChainKeeperMock{
		LoggerFunc: func(ctx sdk.Context) log.Logger { return log.TestingLogger() },
	}
	basek := &mock.BaseKeeperMock{
		ForChainFunc: func(chain nexus.ChainName) types.ChainKeeper { return chaink },
		LoggerFunc:   func(ctx sdk.Context) log.Logger { return log.TestingLogger() },
		HasChainFunc: func(ctx sdk.Context, chain nexus.ChainName) bool { return true },
	}

	encCfg := params.MakeEncodingConfig()
	handler := keeper.NewSigHandler(encCfg.Codec, basek)
	return ctx, basek, chaink, handler
}

func TestHandleCompleted(t *testing.T) {
	var (
		ctx                  sdk.Context
		basek                *mock.BaseKeeperMock
		chaink               *mock.ChainKeeperMock
		sig                  utils.ValidatedProtoMarshaler
		moduleMetadata       codec.ProtoMarshaler
		handler              multisig.SigHandler
		commandBatchMetadata types.CommandBatchMetadata
	)

	repeat := 20

	givenSigsAndModuleMetadata := Given("sigs and module metadata", func() {
		ctx, basek, chaink, handler = setup2()

		multisig := multisigtestutils.MultiSig()
		sig = &multisig
		moduleMetadata = funcs.Must(codectypes.NewAnyWithValue(&types.SigMetadata{
			Type:           types.SigCommand,
			Chain:          exported.Ethereum.Name,
			CommandBatchID: rand.Bytes(common.HashLength),
		})).GetCachedValue().(codec.ProtoMarshaler)
	})

	givenSigsAndModuleMetadata.
		When("module metadata is invalid", func() {
			moduleMetadata = funcs.Must(codectypes.NewAnyWithValue(&gogoprototypes.StringValue{})).GetCachedValue().(codec.ProtoMarshaler)
		}).
		Then("handle completed should panic", func(t *testing.T) {
			assert.Panics(t, func() { _ = handler.HandleCompleted(ctx, sig, moduleMetadata) })
		}).
		Run(t)

	givenSigsAndModuleMetadata.
		When("chain not found", func() {
			basek.HasChainFunc = func(ctx sdk.Context, chain nexus.ChainName) bool { return false }
		}).
		Then("should return error", func(t *testing.T) {
			assert.Error(t, handler.HandleCompleted(ctx, sig, moduleMetadata))
		}).
		Run(t)

	givenSigsAndModuleMetadata.
		When("batch status is not signing", func() {
			chaink.GetBatchByIDFunc = func(ctx sdk.Context, id []byte) types.CommandBatch {
				return types.NewCommandBatch(
					types.CommandBatchMetadata{Status: rand.Of(types.BatchAborted, types.BatchNonExistent, types.BatchSigned)},
					func(batch types.CommandBatchMetadata) {})
			}
		}).
		Then("should return error", func(t *testing.T) {
			assert.Error(t, handler.HandleCompleted(ctx, sig, moduleMetadata))
		}).
		Run(t, repeat)

	givenSigsAndModuleMetadata.
		When("batch status is signing", func() {
			chaink.GetBatchByIDFunc = func(ctx sdk.Context, id []byte) types.CommandBatch {
				commandBatchMetadata = types.CommandBatchMetadata{Status: types.BatchSigning}
				return types.NewCommandBatch(
					commandBatchMetadata,
					func(batch types.CommandBatchMetadata) { commandBatchMetadata = batch })
			}
		}).
		Then("should set command status and signature", func(t *testing.T) {
			err := handler.HandleCompleted(ctx, sig, moduleMetadata)
			assert.Nil(t, err)
			assert.Equal(t, types.BatchSigned, commandBatchMetadata.Status)
			assert.Equal(t, funcs.Must(codectypes.NewAnyWithValue(sig)), commandBatchMetadata.Signature)
		}).
		Run(t)
}
