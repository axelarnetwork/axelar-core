package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/ante/types"
	axelarnetTypes "github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	btctypes "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	evmtypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	tsstypes "github.com/axelarnetwork/axelar-core/x/tss/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	antetypes "github.com/cosmos/cosmos-sdk/x/auth/ante"
)

// ReimburseFeeDecorator reimburse tss and vote txs
type ReimburseFeeDecorator struct {
	ak          antetypes.AccountKeeper
	staking     types.Staking
	axelarnet   types.Axelarnet
	snapshotter types.Snapshotter
}

// NewReimburseFeeDecorator constructor for ReimburseFeeDecorator
func NewReimburseFeeDecorator(ak antetypes.AccountKeeper, staking types.Staking, snapshotter types.Snapshotter, axelarnet types.Axelarnet) ReimburseFeeDecorator {
	return ReimburseFeeDecorator{
		ak,
		staking,
		axelarnet,
		snapshotter,
	}
}

// AnteHandle reimburse the tss and vote transactions from proxy accounts
func (d ReimburseFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	msgs := tx.GetMsgs()

	if d.qualifyForReimburse(ctx, msgs) {
		feeTx, ok := tx.(sdk.FeeTx)
		if !ok {
			return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
		}
		fee := feeTx.GetFee()

		innerMsg := msgs[0].(*axelarnetTypes.RefundMessageRequest).GetInnerMessage()
		err := d.axelarnet.SetPotentialRefund(ctx, axelarnetTypes.GetMsgKey(innerMsg), fee[0])
		if err != nil {
			return ctx, err
		}

	}

	return next(ctx, tx, simulate)
}

func (d ReimburseFeeDecorator) qualifyForReimburse(ctx sdk.Context, msgs []sdk.Msg) bool {
	if len(msgs) != 1 {
		return false
	}

	switch msg := msgs[0].(type) {
	case *axelarnetTypes.RefundMessageRequest:
		innerMsg := msg.GetInnerMessage()
		switch innerMsg.(type) {
		case *tsstypes.ProcessKeygenTrafficRequest, *tsstypes.AckRequest:
			// Validator must be registered for key gen
			validator := getValidator(ctx, d.snapshotter, msgs[0])
			if validator == nil {
				return false
			}
			_, hasProxyRegistered := d.snapshotter.GetProxy(ctx, validator)
			if !hasProxyRegistered {
				return false
			}

		case *tsstypes.ProcessSignTrafficRequest, *tsstypes.VotePubKeyRequest, *tsstypes.VoteSigRequest,
			*btctypes.VoteConfirmOutpointRequest, *evmtypes.VoteConfirmChainRequest, *evmtypes.VoteConfirmDepositRequest,
			*evmtypes.VoteConfirmTokenRequest, *evmtypes.VoteConfirmTransferKeyRequest:
			// Validator must be bounded
			validatorAddr := getValidator(ctx, d.snapshotter, msgs[0])
			if validatorAddr == nil {
				return false
			}
			validator := d.staking.Validator(ctx, validatorAddr)
			if !validator.IsBonded() {
				return false
			}
		default:
			return false
		}
	default:
		return false
	}

	return true
}

// getValidator returns the validator address associated to the proxy address
func getValidator(ctx sdk.Context, snapshotter types.Snapshotter, msg sdk.Msg) sdk.ValAddress {
	sender := msg.GetSigners()[0]
	validator := snapshotter.GetOperator(ctx, sender)
	return validator
}
