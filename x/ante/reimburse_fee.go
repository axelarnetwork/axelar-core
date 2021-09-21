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
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// ReimburseFeeDecorator reimburse tss and vote txs
type ReimburseFeeDecorator struct {
	ak          antetypes.AccountKeeper
	bankKeeper  types.BankKeeper
	tss         types.Tss
	voter       types.Voter
	snapshotter types.Snapshotter
}

// NewReimburseFeeDecorator constructor for ReimburseFeeDecorator
func NewReimburseFeeDecorator(ak antetypes.AccountKeeper, bk types.BankKeeper, tss types.Tss, voter types.Voter, snapshotter types.Snapshotter) ReimburseFeeDecorator {
	return ReimburseFeeDecorator{
		ak,
		bk,
		tss,
		voter,
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

		err := reimburseFees(d.bankKeeper, ctx, feePayer, fee)
		if err != nil {
			return ctx, err
		}
	}

	return next(ctx, tx, simulate)
}

func (d ReimburseFeeDecorator) qualifyForReimburse(ctx sdk.Context, msgs []sdk.Msg) bool {
	for _, msg := range msgs {
		var pollKey vote.PollKey
		switch msg := msg.(type) {
		case *tsstypes.AckRequest, *tsstypes.ProcessKeygenTrafficRequest:
			// Validator must be registered for key gen
			validator, found := d.getValidator(ctx, msg)
			if !found {
				return false
			}
			_, hasProxyRegistered := d.snapshotter.GetProxy(ctx, validator)
			if !hasProxyRegistered {
				return false
			}
			continue
		case *tsstypes.ProcessSignTrafficRequest:
			// Validator must participate in signing for the given sig ID
			validator, found := d.getValidator(ctx, msg)
			if !found {
				return false
			}
			if !d.tss.DoesValidatorParticipateInSign(ctx, msg.SessionID, validator) {
				return false
			}
		case *tsstypes.VotePubKeyRequest:
			pollKey = msg.PollKey
		case *tsstypes.VoteSigRequest:
			pollKey = msg.PollKey
		case *btctypes.VoteConfirmOutpointRequest:
			pollKey = msg.PollKey
		case *evmtypes.VoteConfirmChainRequest:
			pollKey = msg.PollKey
		case *evmtypes.VoteConfirmDepositRequest:
			pollKey = msg.PollKey
		case *evmtypes.VoteConfirmTokenRequest:
			pollKey = msg.PollKey
		case *evmtypes.VoteConfirmTransferKeyRequest:
			pollKey = msg.PollKey
		default:
			return false
		}

		validator, found := d.getValidator(ctx, msg)
		if !found {
			return false
		}
		// Validator must be included for a poll
		poll := d.voter.GetPoll(ctx, pollKey)
		snapshot, ok := d.snapshotter.GetSnapshot(ctx, poll.GetSnapshotSeqNo())
		if !ok {
			return false
		}
		_, ok = snapshot.GetValidator(validator)
		if !ok {
			return false
		}
	}

	return true
}

// getValidator returns the validator address associated to the proxy address
func (d ReimburseFeeDecorator) getValidator(ctx sdk.Context, msg sdk.Msg) (sdk.ValAddress, bool) {
	sender := msg.GetSigners()[0]
	validator := d.snapshotter.GetOperator(ctx, sender)
	return validator, validator == nil
}

// reimburseFees reimburse fees to the given account.
func reimburseFees(bankKeeper types.BankKeeper, ctx sdk.Context, acc sdk.AccAddress, fees sdk.Coins) error {
	if !fees.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "invalid fee amount: %s", fees)
	}
	err := bankKeeper.SendCoinsFromModuleToAccount(ctx, authtypes.FeeCollectorName, acc, fees)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, err.Error())
	}

	return nil
}
