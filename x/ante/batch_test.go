package ante_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/stretchr/testify/assert"

	appParams "github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/ante"
	"github.com/axelarnetwork/axelar-core/x/ante/types/mock"
	auxiliarytypes "github.com/axelarnetwork/axelar-core/x/auxiliary/types"
	evmTypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	rewardtypes "github.com/axelarnetwork/axelar-core/x/reward/types"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
	votetypes "github.com/axelarnetwork/axelar-core/x/vote/types"
	. "github.com/axelarnetwork/utils/test"
)

func TestBatch(t *testing.T) {
	var (
		handler       sdk.AnteDecorator
		tx            *mock.FeeTxMock
		batchMsg      *auxiliarytypes.BatchRequest
		unwrappedMsgs []sdk.Msg
	)

	sender := rand.AccAddr()

	givenBatchAnteHandler := Given("the batch ante handler", func() {
		encCfg := appParams.MakeEncodingConfig()
		handler = ante.NewBatchDecorator(encCfg.Codec)
	})

	givenBatchAnteHandler.
		When("messages do not contain batch", func() {
			tx = &mock.FeeTxMock{
				GetMsgsFunc: func() []sdk.Msg {
					return []sdk.Msg{
						votetypes.NewVoteRequest(sender, vote.PollID(rand.PosI64()), evmTypes.NewVoteEvents(nexus.ChainName(rand.NormalizedStr(3)))),
						votetypes.NewVoteRequest(sender, vote.PollID(rand.PosI64()), evmTypes.NewVoteEvents(nexus.ChainName(rand.NormalizedStr(3)))),
						votetypes.NewVoteRequest(sender, vote.PollID(rand.PosI64()), evmTypes.NewVoteEvents(nexus.ChainName(rand.NormalizedStr(3)))),
					}
				},
			}
		}).
		Then("should pass messages as it", func(t *testing.T) {
			_, err := handler.AnteHandle(sdk.Context{}, tx, false,
				func(_ sdk.Context, tx sdk.Tx, _ bool) (sdk.Context, error) {
					unwrappedMsgs = tx.GetMsgs()
					return sdk.Context{}, nil
				})

			assert.NoError(t, err)
			assert.Equal(t, 3, len(unwrappedMsgs))
		}).
		Run(t)

	givenBatchAnteHandler.
		When("a Batch Request is valid", func() {
			batchMsg = auxiliarytypes.NewBatchRequest(sender, []sdk.Msg{
				votetypes.NewVoteRequest(sender, vote.PollID(rand.PosI64()), evmTypes.NewVoteEvents(nexus.ChainName(rand.NormalizedStr(3)))),
				votetypes.NewVoteRequest(sender, vote.PollID(rand.PosI64()), evmTypes.NewVoteEvents(nexus.ChainName(rand.NormalizedStr(3)))),
			})
		}).
		Then("should unwrap inner message", func(t *testing.T) {
			tx = &mock.FeeTxMock{
				GetMsgsFunc: func() []sdk.Msg {
					return []sdk.Msg{
						votetypes.NewVoteRequest(sender, vote.PollID(rand.PosI64()), evmTypes.NewVoteEvents(nexus.ChainName(rand.NormalizedStr(3)))),
						batchMsg,
						votetypes.NewVoteRequest(sender, vote.PollID(rand.PosI64()), evmTypes.NewVoteEvents(nexus.ChainName(rand.NormalizedStr(3)))),
						batchMsg,
						votetypes.NewVoteRequest(sender, vote.PollID(rand.PosI64()), evmTypes.NewVoteEvents(nexus.ChainName(rand.NormalizedStr(3)))),
					}
				},
			}

			_, err := handler.AnteHandle(sdk.Context{}, tx, false,
				func(_ sdk.Context, tx sdk.Tx, _ bool) (sdk.Context, error) {
					unwrappedMsgs = tx.GetMsgs()
					return sdk.Context{}, nil
				})

			assert.NoError(t, err)
			assert.Equal(t, 9, len(unwrappedMsgs))
		}).
		Run(t)

	givenBatchAnteHandler.
		When("a tx contains an authz MsgExec wrapping messages", func() {
			exec := authz.NewMsgExec(sender, []sdk.Msg{
				votetypes.NewVoteRequest(sender, vote.PollID(rand.PosI64()), evmTypes.NewVoteEvents(nexus.ChainName(rand.NormalizedStr(3)))),
			})
			tx = &mock.FeeTxMock{
				GetMsgsFunc: func() []sdk.Msg {
					return []sdk.Msg{&exec}
				},
			}
		}).
		Then("should unwrap the inner messages", func(t *testing.T) {
			_, err := handler.AnteHandle(sdk.Context{}, tx, false,
				func(_ sdk.Context, tx sdk.Tx, _ bool) (sdk.Context, error) {
					unwrappedMsgs = tx.GetMsgs()
					return sdk.Context{}, nil
				})

			assert.NoError(t, err)
			assert.Equal(t, 2, len(unwrappedMsgs))
		}).
		Run(t)
}

func TestBatchRejectsDisallowedNesting(t *testing.T) {
	encCfg := appParams.MakeEncodingConfig()
	handler := ante.NewBatchDecorator(encCfg.Codec)
	sender := rand.AccAddr()

	voteMsg := votetypes.NewVoteRequest(sender, vote.PollID(rand.PosI64()), evmTypes.NewVoteEvents(nexus.ChainName(rand.NormalizedStr(3))))

	run := func(msgs ...sdk.Msg) error {
		tx := &mock.FeeTxMock{GetMsgsFunc: func() []sdk.Msg { return msgs }}
		_, err := handler.AnteHandle(sdk.Context{}, tx, false,
			func(sdk.Context, sdk.Tx, bool) (sdk.Context, error) { return sdk.Context{}, nil })
		return err
	}

	innerExec := authz.NewMsgExec(sender, []sdk.Msg{voteMsg})
	execInExec := authz.NewMsgExec(sender, []sdk.Msg{&innerExec})
	assert.ErrorIs(t, run(&execInExec), sdkerrors.ErrUnauthorized)

	batchInExec := authz.NewMsgExec(sender, []sdk.Msg{auxiliarytypes.NewBatchRequest(sender, []sdk.Msg{voteMsg})})
	assert.ErrorIs(t, run(&batchInExec), sdkerrors.ErrUnauthorized)

	leafExec := authz.NewMsgExec(sender, []sdk.Msg{voteMsg})
	execInBatch := auxiliarytypes.NewBatchRequest(sender, []sdk.Msg{&leafExec})
	assert.ErrorIs(t, run(execInBatch), sdkerrors.ErrUnauthorized)

	roleGatedInExec := authz.NewMsgExec(sender, []sdk.Msg{&evmTypes.CreateDeployTokenRequest{Sender: sender.String()}})
	assert.ErrorIs(t, run(&roleGatedInExec), sdkerrors.ErrUnauthorized)

	refundInExec := authz.NewMsgExec(sender, []sdk.Msg{rewardtypes.NewRefundMsgRequest(sender, voteMsg)})
	assert.ErrorIs(t, run(&refundInExec), sdkerrors.ErrUnauthorized)

	flatExec := authz.NewMsgExec(sender, []sdk.Msg{voteMsg})
	assert.NoError(t, run(&flatExec))
	assert.NoError(t, run(auxiliarytypes.NewBatchRequest(sender, []sdk.Msg{voteMsg})))
	assert.NoError(t, run(auxiliarytypes.NewBatchRequest(sender, []sdk.Msg{rewardtypes.NewRefundMsgRequest(sender, voteMsg)})))
}
