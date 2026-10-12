package ante_test

import (
	"crypto/sha256"
	"testing"

	"github.com/btcsuite/btcd/btcec/v2"
	ec "github.com/btcsuite/btcd/btcec/v2/ecdsa"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/ante"
	"github.com/axelarnetwork/axelar-core/x/ante/types/mock"
	multisigtypes "github.com/axelarnetwork/axelar-core/x/multisig/types"
	"github.com/axelarnetwork/axelar-core/x/permission/exported"
	permissiontypes "github.com/axelarnetwork/axelar-core/x/permission/types"
	rewardtypes "github.com/axelarnetwork/axelar-core/x/reward/types"
	"github.com/axelarnetwork/utils/funcs"
)

func TestEmergencyPause(t *testing.T) {
	encCfg := app.MakeEncodingConfig()
	sender := rand.AccAddr()

	handle := func(level exported.PauseLevel, msgs ...sdk.Msg) error {
		decorator := ante.NewEmergencyPause(&mock.PermissionMock{
			GetPauseLevelFunc: func(sdk.Context) exported.PauseLevel { return level },
		})
		_, err := ante.ChainMessageAnteDecorators(decorator).ToAnteHandler()(sdk.Context{}, &mockTx{msgs: msgs}, false)
		return err
	}

	submitSignature := func() sdk.Msg {
		priv := funcs.Must(btcec.NewPrivateKey())
		hash := sha256.Sum256([]byte("payload"))
		return multisigtypes.NewSubmitSignatureRequest(sender, 1, ec.Sign(priv, hash[:]).Serialize())
	}
	bankSend := banktypes.NewMsgSend(sender, rand.AccAddr(), sdk.NewCoins(sdk.NewInt64Coin("uaxl", 1)))

	t.Run("not paused: every message passes", func(t *testing.T) {
		assert.NoError(t, handle(exported.PAUSE_LEVEL_UNSPECIFIED, bankSend, rewardtypes.NewRefundMsgRequest(sender, submitSignature())))
	})

	t.Run("paused: non-allowlisted messages are rejected", func(t *testing.T) {
		for _, level := range []exported.PauseLevel{exported.PAUSE_LEVEL_ALL, exported.PAUSE_LEVEL_EXCEPT_VALIDATORS} {
			assert.ErrorIs(t, handle(level, bankSend), permissiontypes.ErrChainPaused)
			// a single disallowed message rejects the whole tx
			assert.ErrorIs(t, handle(level, permissiontypes.NewUnpauseRequest(sender, "resolved"), bankSend), permissiontypes.ErrChainPaused)
		}
	})

	t.Run("paused: every allowlisted message passes", func(t *testing.T) {
		for level, urls := range exported.EmergencyPauseAllowlist {
			for _, url := range urls {
				msg := funcs.Must(encCfg.InterfaceRegistry.Resolve(url)).(sdk.Msg)
				if _, ok := msg.(*rewardtypes.RefundMsgRequest); ok {
					msg = rewardtypes.NewRefundMsgRequest(sender, submitSignature())
				}

				assert.NoError(t, handle(level, msg), "%s at level %s", url, level)
			}
		}
	})

	t.Run("paused: the inner message of RefundMsgRequest is checked", func(t *testing.T) {
		refund := rewardtypes.NewRefundMsgRequest(sender, submitSignature())
		assert.NoError(t, handle(exported.PAUSE_LEVEL_EXCEPT_VALIDATORS, refund))
		assert.ErrorIs(t, handle(exported.PAUSE_LEVEL_ALL, refund), permissiontypes.ErrChainPaused)

		// refunds can only wrap refundable messages, so test the inner check with an empty wrapper
		assert.ErrorIs(t, handle(exported.PAUSE_LEVEL_EXCEPT_VALIDATORS, &rewardtypes.RefundMsgRequest{Sender: sender.String()}), permissiontypes.ErrChainPaused)
	})
}

type mockTx struct {
	sdk.Tx
	msgs []sdk.Msg
}

func (m *mockTx) GetMsgs() []sdk.Msg {
	return m.msgs
}
