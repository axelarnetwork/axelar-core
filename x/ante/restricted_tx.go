package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/ante/types"
	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/types"
)

// RestrictedTx checks if the signer is authorized to send restricted transactions
type RestrictedTx struct {
	tss types.Tss
}

// NewRestrictedTx constructor for RestrictedTx
func NewRestrictedTx(tss types.Tss) RestrictedTx {
	return RestrictedTx{
		tss,
	}
}

// AnteHandle fails if the signer is not authorized to send the transaction
func (d RestrictedTx) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {

	msgs := tx.GetMsgs()
	for _, msg := range msgs {
		switch msg := msg.(type) {
		case *tss.UpdateGovernanceKeyRequest, *axelarnet.RegisterFeeCollectorRequest:
			signer := msg.GetSigners()[0]

			governanceKey, ok := d.tss.GetGovernanceKey(ctx)
			if !ok {
				panic("governance key not found")
			}

			if !signer.Equals(sdk.AccAddress(governanceKey.Address().Bytes())) {
				return ctx, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "%s is not authorized to send transaction %T", signer, msg)
			}
		default:
			continue
		}
	}

	return next(ctx, tx, simulate)
}
