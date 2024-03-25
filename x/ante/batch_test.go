package ante_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	appParams "github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/ante"
	"github.com/axelarnetwork/axelar-core/x/ante/types/mock"
	batchtypes "github.com/axelarnetwork/axelar-core/x/batch/types"
	evmTypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
	votetypes "github.com/axelarnetwork/axelar-core/x/vote/types"
	. "github.com/axelarnetwork/utils/test"
)

func TestBatch(t *testing.T) {
	var (
		handler       sdk.AnteDecorator
		tx            *mock.FeeTxMock
		batchMsg      *batchtypes.BatchRequest
		unwrappedMsgs []sdk.Msg
	)

	sender := rand.AccAddr()

	givenBatchAnteHandler := Given("the batch ante handler", func() {
		encCfg := appParams.MakeEncodingConfig()
		handler = ante.NewBatchDecorator(encCfg.Codec)
	})

	givenBatchAnteHandler.
		When("a BatchRequest contains nested batch message", func() {
			req := batchtypes.NewBatchRequest(sender, []sdk.Msg{
				votetypes.NewVoteRequest(sender, vote.PollID(rand.PosI64()), evmTypes.NewVoteEvents(nexus.ChainName(rand.NormalizedStr(3)))),
			})
			batchMsg = batchtypes.NewBatchRequest(sender, []sdk.Msg{req})
		}).
		Then("ante handler should return an error", func(t *testing.T) {
			tx = &mock.FeeTxMock{
				GetMsgsFunc: func() []sdk.Msg {
					return []sdk.Msg{batchMsg}
				},
			}

			_, err := handler.AnteHandle(sdk.Context{}, tx, false,
				func(sdk.Context, sdk.Tx, bool) (sdk.Context, error) { return sdk.Context{}, nil })
			assert.ErrorContains(t, err, "nested batch")
		}).
		Run(t)

	givenBatchAnteHandler.
		When("a Batch Request is valid", func() {
			batchMsg = batchtypes.NewBatchRequest(sender, []sdk.Msg{
				votetypes.NewVoteRequest(sender, vote.PollID(rand.PosI64()), evmTypes.NewVoteEvents(nexus.ChainName(rand.NormalizedStr(3)))),
				votetypes.NewVoteRequest(sender, vote.PollID(rand.PosI64()), evmTypes.NewVoteEvents(nexus.ChainName(rand.NormalizedStr(3)))),
			})
		}).
		Then("should unwrap inner message", func(t *testing.T) {
			tx = &mock.FeeTxMock{
				GetMsgsFunc: func() []sdk.Msg {
					return []sdk.Msg{batchMsg, batchMsg}
				},
			}

			_, err := handler.AnteHandle(sdk.Context{}, tx, false,
				func(_ sdk.Context, tx sdk.Tx, _ bool) (sdk.Context, error) {
					unwrappedMsgs = tx.GetMsgs()
					return sdk.Context{}, nil
				})

			assert.NoError(t, err)
			assert.Equal(t, 6, len(unwrappedMsgs))
		}).
		Run(t)
}
