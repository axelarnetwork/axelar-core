package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"

	"github.com/axelarnetwork/axelar-core/x/ante/types"
	permission "github.com/axelarnetwork/axelar-core/x/permission/exported"
)

// RestrictedTx checks if the signer is authorized to send restricted transactions
type RestrictedTx struct {
	permission types.Permission
}

// NewRestrictedTx constructor for RestrictedTx
func NewRestrictedTx(permission types.Permission) RestrictedTx {
	return RestrictedTx{
		permission,
	}
}

// AnteHandle fails if the signer is not authorized to send the transaction
func (d RestrictedTx) AnteHandle(ctx sdk.Context, msgs []sdk.Msg, simulate bool, next MessageAnteHandler) (sdk.Context, error) {
	for _, msg := range msgs {
		signers := msg.GetSigners()
		var signer sdk.AccAddress

		if len(signers) != 0 {
			signer = signers[0]
		}

		signerRole := d.permission.GetRole(ctx, signer)
		switch permission.GetPermissionRole((msg).(descriptor.Message)) {
		case permission.ROLE_ACCESS_CONTROL:
			if permission.ROLE_ACCESS_CONTROL != signerRole {
				return ctx, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "signer '%s' is not authorized to send transaction %T", signer, msg)
			}
		case permission.ROLE_CHAIN_MANAGEMENT:
			if permission.ROLE_CHAIN_MANAGEMENT != signerRole {
				return ctx, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "signer '%s' is not authorized to send transaction %T", signer, msg)
			}
		default:
			continue
		}
	}

	return next(ctx, msgs, simulate)
}
