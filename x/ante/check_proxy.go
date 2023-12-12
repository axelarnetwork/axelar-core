package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/axelarnetwork/axelar-core/x/ante/types"
)

// CheckProxy checks if the proxy already sent its readiness message
type CheckProxy struct {
	snapshotter types.Snapshotter
}

// NewCheckProxy constructor for CheckProxyReadiness
func NewCheckProxy(snapshotter types.Snapshotter) CheckProxy {
	return CheckProxy{
		snapshotter,
	}
}

// AnteHandle fails the transaction if it finds any validator holding multiSig share of active keys is trying to unbond
func (d CheckProxy) AnteHandle(ctx sdk.Context, msgs []sdk.Msg, simulate bool, next MessageAnteHandler) (sdk.Context, error) {
	// exempt genesis validator(s) from this check
	if ctx.BlockHeight() == 0 {
		return next(ctx, msgs, simulate)
	}

	for _, msg := range msgs {
		switch msg := msg.(type) {
		case *stakingtypes.MsgCreateValidator:
			valAddress, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
			if err != nil {
				return ctx, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, err.Error())
			}
			if proxy, active := d.snapshotter.GetProxy(ctx, valAddress); proxy.Empty() || !active {
				return ctx, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "no proxy found for operator %s", valAddress.String())
			}
		default:
			continue
		}
	}

	return next(ctx, msgs, simulate)
}
