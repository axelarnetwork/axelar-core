package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/ante/types"
	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	evm "github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/types"
	permission "github.com/axelarnetwork/axelar-core/x/permission/exported"
	permissionTypes "github.com/axelarnetwork/axelar-core/x/permission/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/types"
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
func (d RestrictedTx) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {

	msgs := tx.GetMsgs()
	for _, msg := range msgs {
		switch msg := msg.(type) {
		case *permissionTypes.UpdateGovernanceKeyRequest, *permissionTypes.RegisterControllerRequest,
			*axelarnet.RegisterFeeCollectorRequest:

			signer := msg.GetSigners()[0]
			if permission.ROLE_ACCESS_CONTROL != d.permission.GetRole(ctx, signer) {
				return ctx, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "%s is not authorized to send transaction %T", signer, msg)
			}
		case *tss.RegisterExternalKeysRequest, *tss.StartKeygenRequest,
			*tss.RotateKeyRequest, *axelarnet.RegisterIBCPathRequest,
			*axelarnet.RegisterAssetRequest, *axelarnet.AddCosmosBasedChainRequest,
			*evm.AddChainRequest, *evm.ConfirmGatewayDeploymentRequest,
			*evm.CreateDeployTokenRequest, *evm.CreateTransferOwnershipRequest,
			*evm.CreateTransferOperatorshipRequest, *nexus.ActivateChainRequest:

			signer := msg.GetSigners()[0]
			if permission.ROLE_CHAIN_MANAGEMENT != d.permission.GetRole(ctx, signer) {
				return ctx, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "%s is not authorized to send transaction %T", signer, msg)
			}
		default:
			continue
		}
	}

	return next(ctx, tx, simulate)
}
