package ante

import (
	"github.com/axelarnetwork/axelar-core/x/ante/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// CheckProxyReadyness checks if the proxy already sent its readyness message
type CheckProxyReadyness struct {
	snapshotter types.Snapshotter
}

// NewCheckProxyReadyness constructor for CheckProxyReadyness
func NewCheckProxyReadyness(snapshotter types.Snapshotter) CheckProxyReadyness {
	return CheckProxyReadyness{
		snapshotter,
	}
}

// AnteHandle fails the transaction if it finds any validator holding tss share of active keys is trying to unbond
func (d CheckProxyReadyness) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	// exempt genesis validator(s) from this check
	if ctx.BlockHeight() == 0 {
		return next(ctx, tx, simulate)
	}

	msgs := tx.GetMsgs()
	for _, msg := range msgs {
		switch msg := msg.(type) {
		case *stakingtypes.MsgCreateValidator:
			valAddress, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
			if err != nil {
				return ctx, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, err.Error())
			}
			if !d.snapshotter.IsProxyReady(ctx, valAddress) {
				return ctx, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "no readyness message found for operator %s", valAddress.String())
			}
		default:
			continue
		}
	}

	return next(ctx, tx, simulate)
}
