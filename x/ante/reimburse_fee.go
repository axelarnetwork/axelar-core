package ante

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	antetypes "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/axelarnetwork/axelar-core/x/ante/types"
	btctypes "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	evmtypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	tsstypes "github.com/axelarnetwork/axelar-core/x/tss/types"
)

// ReimburseFeeDecorator reimburse tss and vote txs
type ReimburseFeeDecorator struct {
	ak          antetypes.AccountKeeper
	bankKeeper  types.BankKeeper
	staking     types.Staking
	snapshotter types.Snapshotter
}

// NewReimburseFeeDecorator constructor for ReimburseFeeDecorator
func NewReimburseFeeDecorator(ak antetypes.AccountKeeper, bk types.BankKeeper, staking types.Staking, snapshotter types.Snapshotter) ReimburseFeeDecorator {
	return ReimburseFeeDecorator{
		ak,
		bk,
		staking,
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

		feePayer := feeTx.FeePayer()
		if addr := d.ak.GetModuleAddress(authtypes.FeeCollectorName); addr == nil {
			panic(fmt.Sprintf("%s module account has not been set", authtypes.FeeCollectorName))
		}
		fee := feeTx.GetFee()

		err := reimburseFees(ctx, d.bankKeeper, feePayer, fee)
		if err != nil {
			return ctx, err
		}
	}

	return next(ctx, tx, simulate)
}

func (d ReimburseFeeDecorator) qualifyForReimburse(ctx sdk.Context, msgs []sdk.Msg) bool {
	for _, msg := range msgs {
		switch msg.(type) {
		case *tsstypes.ProcessKeygenTrafficRequest, *tsstypes.AckRequest:
			// Validator must be registered for key gen
			validator, found := getValidator(ctx, d.snapshotter, msg)
			if !found {
				return false
			}
			_, hasProxyRegistered := d.snapshotter.GetProxy(ctx, validator)
			if !hasProxyRegistered {
				return false
			}
			continue
		case *tsstypes.ProcessSignTrafficRequest, *tsstypes.VotePubKeyRequest, *tsstypes.VoteSigRequest,
			*btctypes.VoteConfirmOutpointRequest, *evmtypes.VoteConfirmChainRequest, *evmtypes.VoteConfirmDepositRequest,
			*evmtypes.VoteConfirmTokenRequest, *evmtypes.VoteConfirmTransferKeyRequest:
			// Validator must be bounded
			validatorAddr, found := getValidator(ctx, d.snapshotter, msg)
			if !found {
				return false
			}
			validator := d.staking.Validator(ctx, validatorAddr)
			if !validator.IsBonded() {
				return false
			}
			continue

		default:
			return false
		}
	}

	return true
}

// getValidator returns the validator address associated to the proxy address
func getValidator(ctx sdk.Context, snapshotter types.Snapshotter, msg sdk.Msg) (sdk.ValAddress, bool) {
	sender := msg.GetSigners()[0]
	validator := snapshotter.GetOperator(ctx, sender)
	return validator, validator != nil
}

// reimburseFees reimburse fees to the given account.
func reimburseFees(ctx sdk.Context, bankKeeper types.BankKeeper, acc sdk.AccAddress, fees sdk.Coins) error {
	if !fees.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "invalid fee amount: %s", fees)
	}
	err := bankKeeper.SendCoinsFromModuleToAccount(ctx, authtypes.FeeCollectorName, acc, fees)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, err.Error())
	}

	return nil
}
