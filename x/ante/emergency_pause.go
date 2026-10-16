package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/ante/types"
	permission "github.com/axelarnetwork/axelar-core/x/permission/exported"
	permissiontypes "github.com/axelarnetwork/axelar-core/x/permission/types"
	rewardtypes "github.com/axelarnetwork/axelar-core/x/reward/types"
)

// EmergencyPause rejects messages that are not allowed at the current emergency pause level.
// It runs in CheckTx and in block execution. Messages executed by other messages are additionally
// checked by the permission module's circuit breaker on the msg service router.
type EmergencyPause struct {
	permission types.Permission
}

// NewEmergencyPause constructor for EmergencyPause
func NewEmergencyPause(permission types.Permission) EmergencyPause {
	return EmergencyPause{permission}
}

// AnteHandle fails if any message is not allowed at the current pause level.
// BatchRequest and authz MsgExec are already unwrapped by the BatchDecorator, RefundMsgRequest is unwrapped here.
func (d EmergencyPause) AnteHandle(ctx sdk.Context, msgs []sdk.Msg, simulate bool, next MessageAnteHandler) (sdk.Context, error) {
	level := d.permission.GetPauseLevel(ctx)
	if !level.IsPaused() || d.permission.IsPauseSkipped(ctx) {
		return next(ctx, msgs, simulate)
	}

	for _, msg := range msgs {
		if err := checkAllowedWhilePaused(level, msg); err != nil {
			return ctx, err
		}

		if refund, ok := msg.(*rewardtypes.RefundMsgRequest); ok {
			var inner sdk.Msg
			if refund.InnerMessage != nil {
				inner = refund.GetInnerMessage()
			}
			if inner == nil {
				return ctx, permissiontypes.ErrChainPaused.Wrap("cannot unwrap the inner message of RefundMsgRequest")
			}

			if err := checkAllowedWhilePaused(level, inner); err != nil {
				return ctx, err
			}
		}
	}

	return next(ctx, msgs, simulate)
}

func checkAllowedWhilePaused(level permission.PauseLevel, msg sdk.Msg) error {
	typeURL := sdk.MsgTypeURL(msg)
	if !permission.IsAllowedWhilePaused(level, typeURL) {
		return permissiontypes.ErrChainPaused.Wrapf("message %s is not allowed at pause level %s", typeURL, level)
	}

	return nil
}
