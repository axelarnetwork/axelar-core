package bitcoin

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils/denom"
	balance "github.com/axelarnetwork/axelar-core/x/balance/exported"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// NewHandler creates an sdk.Handler for all bitcoin type messages
func NewHandler(k keeper.Keeper, v types.Voter, rpc types.RPCClient, signer types.Signer, snap types.Snapshotter, b types.Balancer) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		var res *sdk.Result
		var err error
		switch msg := msg.(type) {
		case types.MsgVerifyTx:
			res, err = handleMsgVerifyTx(ctx, k, v, rpc, msg)
		case *types.MsgVoteVerifiedTx:
			res, err = handleMsgVoteVerifiedTx(ctx, k, v, b, msg)
		case types.MsgSignTx:
			res, err = handleMsgSignTx(ctx, k, signer, snap, msg)
		case types.MsgLink:
			res, err = handleMsgLink(ctx, k, signer, b, msg)
		case types.MsgSignPendingTransfers:
			res, err = handleMsgSignPendingTransfersTx(ctx, k, signer, snap, b, msg)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest,
				fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg))
		}
		if err != nil {
			return nil, sdkerrors.Wrap(types.ErrBitcoin, err.Error())
		}
		return res, nil
	}
}

func handleMsgLink(ctx sdk.Context, k keeper.Keeper, s types.Signer, b types.Balancer, msg types.MsgLink) (*sdk.Result, error) {
	key, ok := s.GetCurrentMasterKey(ctx, balance.Bitcoin)
	if !ok {
		return nil, fmt.Errorf("master key not set")
	}

	keyID, ok := s.GetCurrentMasterKeyID(ctx, balance.Bitcoin)
	if !ok {
		return nil, fmt.Errorf("master key not set")
	}

	btcAddr, err := k.CreateDepositAddress(ctx, msg.Recipient, keyID, btcec.PublicKey(key))
	if err != nil {
		return nil, err
	}

	err = b.LinkAddresses(ctx, balance.CrossChainAddress{Chain: balance.Bitcoin, Address: btcAddr.EncodeAddress()}, msg.Recipient)
	if err != nil {
		return nil, err
	}

	logMsg := fmt.Sprintf("successfully linked {%s} and {%s}", btcAddr.EncodeAddress(), msg.Recipient.String())
	k.Logger(ctx).Info(logMsg)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeAddress, btcAddr.String()),
			sdk.NewAttribute(types.AttributeAddress, msg.Recipient.String()),
		),
	)

	return &sdk.Result{
		Data:   []byte(btcAddr.EncodeAddress()),
		Log:    logMsg,
		Events: ctx.EventManager().Events(),
	}, nil
}

func handleMsgVerifyTx(ctx sdk.Context, k keeper.Keeper, v types.Voter, rpc types.RPCClient, msg types.MsgVerifyTx) (*sdk.Result, error) {
	k.Logger(ctx).Debug("verifying bitcoin transaction")

	txID := msg.OutPointInfo.OutPoint.Hash.String()
	poll := exported.PollMeta{Module: types.ModuleName, Type: msg.Type(), ID: msg.OutPointInfo.OutPoint.String()}
	if err := v.InitPoll(ctx, poll); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeTxID, txID),
			sdk.NewAttribute(types.AttributeAmount, msg.OutPointInfo.Amount.String()),
		),
	)

	// store outpoint for later reference
	k.SetUnverifiedOutpointInfo(ctx, msg.OutPointInfo)

	/*
	 Anyone not able to verify the transaction will automatically record a negative vote,
	 but only validators will later send out that vote.
	*/

	err := verifyTx(rpc, msg.OutPointInfo, k.GetRequiredConfirmationHeight(ctx))
	switch err {
	// verification successful
	case nil:
		if err := v.RecordVote(ctx, &types.MsgVoteVerifiedTx{PollMeta: poll, VotingData: true}); err != nil {
			k.Logger(ctx).Error(sdkerrors.Wrap(err, "voting failed").Error())
			return &sdk.Result{Log: err.Error(), Events: ctx.EventManager().Events()}, nil
		}

		k.Logger(ctx).Debug(fmt.Sprintf("transaction (%s) was verified", txID))
		return &sdk.Result{Log: "successfully verified transaction", Events: ctx.EventManager().Events()}, nil
	// verification unsuccessful
	default:
		if err := v.RecordVote(ctx, &types.MsgVoteVerifiedTx{PollMeta: poll, VotingData: false}); err != nil {
			k.Logger(ctx).Error(sdkerrors.Wrap(err, "voting failed").Error())
			return &sdk.Result{Log: err.Error(), Events: ctx.EventManager().Events()}, nil
		}

		k.Logger(ctx).Debug(sdkerrors.Wrapf(err, "expected transaction (%s) could not be verified", txID).Error())
		return &sdk.Result{Log: err.Error(), Events: ctx.EventManager().Events()}, nil
	}
}

// This can be used as a potential hook to immediately act on a poll being decided by the vote
func handleMsgVoteVerifiedTx(ctx sdk.Context, k keeper.Keeper, v types.Voter, b types.Balancer, msg *types.MsgVoteVerifiedTx) (*sdk.Result, error) {
	event := sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
		sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
		sdk.NewAttribute(types.AttributePoll, msg.PollMeta.String()),
		sdk.NewAttribute(types.AttributeVotingData, strconv.FormatBool(msg.VotingData)),
	)

	if err := v.TallyVote(ctx, msg); err != nil {
		return nil, err
	}

	if confirmed := v.Result(ctx, msg.Poll()); confirmed != nil {
		outPoint, err := types.OutPointFromStr(msg.PollMeta.ID)
		if err != nil {
			return nil, err
		}
		k.ProcessVerificationResult(ctx, msg.PollMeta.ID, confirmed.(bool))
		v.DeletePoll(ctx, msg.Poll())

		info, ok := k.GetVerifiedOutPointInfo(ctx, outPoint)
		if !ok {
			return nil, fmt.Errorf("outpoint not verified")
		}
		addr, err := btcutil.DecodeAddress(info.Address, k.GetNetwork(ctx).Params)
		if err != nil {
			return nil, err
		}
		id, ok := k.GetKeyIDByAddress(ctx, addr)
		if !ok {
			return nil, fmt.Errorf("key id not found")
		}
		k.SetKeyIDByOutpoint(ctx, outPoint, id)

		depositAddr := balance.CrossChainAddress{Address: info.Address, Chain: balance.Bitcoin}
		amount := sdk.NewInt64Coin(denom.Satoshi, int64(info.Amount))
		err = b.EnqueueForTransfer(ctx, depositAddr, amount)
		if err != nil {
			k.Logger(ctx).Info(fmt.Sprintf("prepared no transfer: %s", err))
		}

		event = event.AppendAttributes(sdk.NewAttribute(types.AttributePollConfirmed, strconv.FormatBool(confirmed.(bool))))
	}

	ctx.EventManager().EmitEvent(event)
	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgSignTx(ctx sdk.Context, k keeper.Keeper, signer types.Signer, snap types.Snapshotter, msg types.MsgSignTx) (*sdk.Result, error) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
		),
	)
	k.SetRawTx(ctx, msg.Outpoint, msg.RawTx)
	k.SpendVerifiedOutPoint(ctx, msg.Outpoint.String())
	hashes, err := signTx(ctx, k, signer, snap, msg.RawTx)
	if err != nil {
		return nil, err
	}
	return &sdk.Result{
		Data:   hashes[0],
		Log:    fmt.Sprintf("successfully started signing protocol for transaction that spends %s.", msg.Outpoint),
		Events: ctx.EventManager().Events(),
	}, nil
}

func handleMsgSignPendingTransfersTx(ctx sdk.Context, k keeper.Keeper, signer types.Signer, snap types.Snapshotter, balancer types.Balancer, msg types.MsgSignPendingTransfers) (*sdk.Result, error) {
	pendingTransfers := balancer.GetPendingTransfersForChain(ctx, balance.Bitcoin)

	if len(pendingTransfers) == 0 {
		return &sdk.Result{
			Log:    fmt.Sprintf("no pending transfer for chain %s found", balance.Ethereum),
			Events: ctx.EventManager().Events(),
		}, nil
	}

	var outPuts []types.Output
	totalWithdrawals := sdk.ZeroInt()
	for _, transfer := range pendingTransfers {
		if transfer.Asset.Denom != denom.Satoshi {
			return nil, fmt.Errorf("expected transfer currency to be %s, got %s", denom.Satoshi, transfer.Asset.Denom)
		}
		recipient, err := btcutil.DecodeAddress(transfer.Recipient.Address, k.GetNetwork(ctx).Params)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("%s is not a valid address", transfer.Recipient))
			continue
		}

		outPuts = append(outPuts, types.Output{
			Amount:    btcutil.Amount(transfer.Asset.Amount.Int64()),
			Recipient: recipient,
		})
		totalWithdrawals = totalWithdrawals.Add(transfer.Asset.Amount)
		balancer.ArchivePendingTransfer(ctx, transfer)
	}

	var prevOuts []*wire.OutPoint
	totalDeposits := sdk.ZeroInt()
	for _, info := range k.GetVerifiedOutPointInfos(ctx) {
		prevOuts = append(prevOuts, info.OutPoint)
		totalDeposits = totalDeposits.AddRaw(int64(info.Amount))
		k.SpendVerifiedOutPoint(ctx, info.OutPoint.String())
	}

	change := totalDeposits.Sub(totalWithdrawals).SubRaw(int64(msg.Fee))
	if change.IsNegative() {
		return nil, fmt.Errorf("not enough deposits (%s) to make all withdrawals (%s) with a transaction fee of %s",
			totalDeposits.String(),
			totalWithdrawals.String(),
			msg.Fee.String(),
		)
	}

	if change.IsPositive() {
		if !change.IsInt64() {
			return nil, fmt.Errorf("the calculated change is too large for a single transaction")
		}

		pk, ok := signer.GetCurrentMasterKey(ctx, balance.Bitcoin)
		if !ok {
			return nil, fmt.Errorf("key not found")
		}
		masterAddr, script, err := k.GenerateMasterAddressAndRedeemScript(ctx, btcec.PublicKey(pk))
		if err != nil {
			return nil, err
		}
		k.SetRedeemScript(ctx, masterAddr, script)
		outPuts = append(outPuts, types.Output{
			Amount:    btcutil.Amount(change.Int64()),
			Recipient: masterAddr,
		})
	}

	tx, err := types.CreateTx(prevOuts, outPuts)
	if err != nil {
		return nil, err
	}
	k.SetRawConsolidationTx(ctx, tx)

	_, err = signTx(ctx, k, signer, snap, tx)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
		),
	)

	return &sdk.Result{
		Log:    fmt.Sprintf("successfully started signing protocols to spend pending transfers"),
		Events: ctx.EventManager().Events(),
	}, nil
}

func signTx(ctx sdk.Context, k keeper.Keeper, signer types.Signer, snap types.Snapshotter, tx *wire.MsgTx) ([][]byte, error) {
	// Print out the hash that becomes the input for the threshold signing
	hashes, err := k.GetHashesToSign(ctx, tx)
	if err != nil {
		return nil, err
	}
	for i, hash := range hashes {
		serializedHash := hex.EncodeToString(hash)
		k.Logger(ctx).Info(fmt.Sprintf("hash to sign: %s", serializedHash))

		keyID, ok := k.GetKeyIDByOutpoint(ctx, &tx.TxIn[i].PreviousOutPoint)
		if !ok {
			return nil, fmt.Errorf("no key ID for chain %s found", balance.Bitcoin)
		}

		round, ok := signer.GetSnapshotRoundForKeyID(ctx, keyID)
		if !ok {
			return nil, fmt.Errorf("no snapshot round for key ID %s registered", keyID)
		}
		s, ok := snap.GetSnapshot(ctx, round)
		if !ok {
			return nil, fmt.Errorf("no snapshot found")
		}
		err = signer.StartSign(ctx, keyID, serializedHash, hash, s.Validators)
		if err != nil {
			if !ok {
				return nil, err
			}
		}
	}
	return hashes, nil
}

func verifyTx(rpc types.RPCClient, expectedInfo types.OutPointInfo, requiredConfirmations uint64) error {
	actualInfo, err := rpc.GetOutPointInfo(expectedInfo.BlockHash, expectedInfo.OutPoint)
	if err != nil {
		return sdkerrors.Wrap(err, "could not retrieve Bitcoin transaction")
	}

	if actualInfo.Address != expectedInfo.Address {
		return fmt.Errorf("expected destination address does not match actual destination address")
	}

	if actualInfo.Amount != expectedInfo.Amount {
		return fmt.Errorf("expected amount does not match actual amount")
	}

	if actualInfo.Confirmations < requiredConfirmations {
		return fmt.Errorf("not enough confirmations yet")
	}

	return nil
}
