package bitcoin

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// NewHandler creates an sdk.Handler for all bitcoin type messages
func NewHandler(k types.BTCKeeper, v types.Voter, signer types.Signer, n types.Nexus, snapshotter types.Snapshotter) sdk.Handler {
	h := func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case types.MsgLink:
			return HandleMsgLink(ctx, k, signer, n, msg)
		case types.MsgConfirmOutpoint:
			return HandleMsgConfirmOutpoint(ctx, k, v, signer, msg)
		case types.MsgVoteConfirmOutpoint:
			return HandleMsgVoteConfirmOutpoint(ctx, k, v, n, msg)
		case types.MsgSignPendingTransfers:
			return HandleMsgSignPendingTransfers(ctx, k, signer, n, snapshotter, v, msg)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest,
				fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg))
		}
	}

	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		res, err := h(ctx, msg)
		if err != nil {
			k.Logger(ctx).Debug(err.Error())
			return nil, sdkerrors.Wrap(types.ErrBitcoin, err.Error())
		}
		k.Logger(ctx).Debug(res.Log)
		return res, nil
	}
}

// HandleMsgLink handles address linking
func HandleMsgLink(ctx sdk.Context, k types.BTCKeeper, s types.Signer, n types.Nexus, msg types.MsgLink) (*sdk.Result, error) {
	key, recipientChain, err := checkLinkRequisites(ctx, s, n, msg.RecipientChain)
	if err != nil {
		return nil, err
	}

	recipient := nexus.CrossChainAddress{Chain: recipientChain, Address: msg.RecipientAddr}
	depositAddr := types.NewLinkedAddress(key, k.GetNetwork(ctx), recipient)
	n.LinkAddresses(ctx, depositAddr.ToCrossChainAddr(), recipient)
	k.SetAddress(ctx, depositAddr)

	return &sdk.Result{
		Data:   []byte(depositAddr.EncodeAddress()),
		Log:    fmt.Sprintf("successfully linked {%s} and {%s}", depositAddr.ToCrossChainAddr().String(), recipient.String()),
		Events: ctx.EventManager().Events(),
	}, nil
}

// HandleMsgConfirmOutpoint handles the confirmation of a Bitcoin outpoint
func HandleMsgConfirmOutpoint(ctx sdk.Context, k types.BTCKeeper, voter types.InitPoller, signer types.Signer, msg types.MsgConfirmOutpoint) (*sdk.Result, error) {
	_, state, ok := k.GetOutPointInfo(ctx, *msg.OutPointInfo.OutPoint)
	switch {
	case !ok:
		break
	case state == types.CONFIRMED:
		return nil, fmt.Errorf("already confirmed")
	case state == types.SPENT:
		return nil, fmt.Errorf("already spent")
	}

	if _, ok := k.GetAddress(ctx, msg.OutPointInfo.Address); !ok {
		return nil, fmt.Errorf("outpoint address unknown, aborting deposit confirmation")
	}

	keyID, ok := signer.GetCurrentKeyID(ctx, exported.Bitcoin, tss.MasterKey)
	if !ok {
		return nil, fmt.Errorf("no master key for chain %s found", exported.Bitcoin.Name)
	}

	counter, ok := signer.GetSnapshotCounterForKeyID(ctx, keyID)
	if !ok {
		return nil, fmt.Errorf("no snapshot counter for key ID %s registered", keyID)
	}

	poll := vote.NewPollMetaWithNonce(types.ModuleName, msg.Type(), msg.OutPointInfo.OutPoint.String(), ctx.BlockHeight(), k.GetRevoteLockingPeriod(ctx))
	if err := voter.InitPoll(ctx, poll, counter); err != nil {
		return nil, err
	}
	k.SetPendingOutpointInfo(ctx, poll, msg.OutPointInfo)

	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeOutpointConfirmation,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueStart),
		sdk.NewAttribute(types.AttributeKeyConfHeight, strconv.FormatUint(k.GetRequiredConfirmationHeight(ctx), 10)),
		sdk.NewAttribute(types.AttributeKeyOutPointInfo, string(k.Codec().MustMarshalJSON(msg.OutPointInfo))),
		sdk.NewAttribute(types.AttributeKeyPoll, string(k.Codec().MustMarshalJSON(poll))),
	))

	return &sdk.Result{
		Log:    fmt.Sprintf("votes on confirmation of %s started", msg.OutPointInfo.OutPoint.String()),
		Events: ctx.EventManager().Events(),
	}, nil
}

// HandleMsgVoteConfirmOutpoint handles the votes on an outpoint confirmation
func HandleMsgVoteConfirmOutpoint(ctx sdk.Context, k types.BTCKeeper, v types.Voter, n types.Nexus, msg types.MsgVoteConfirmOutpoint) (*sdk.Result, error) {
	// has the outpoint been confirmed before?
	confirmedOutPointInfo, state, confirmedBefore := k.GetOutPointInfo(ctx, msg.OutPoint)
	// is there an ongoing poll?
	pendingOutPointInfo, pollFound := k.GetPendingOutPointInfo(ctx, msg.Poll)

	switch {
	// a malicious user could try to delete an ongoing poll by providing an already confirmed outpoint,
	// so we need to check that it matches the poll before deleting
	case confirmedBefore && pollFound && pendingOutPointInfo.OutPoint.String() == confirmedOutPointInfo.OutPoint.String():
		v.DeletePoll(ctx, msg.Poll)
		k.DeletePendingOutPointInfo(ctx, msg.Poll)
		fallthrough
	// If the voting threshold has been met and additional votes are received they should not return an error
	case confirmedBefore:
		switch state {
		case types.CONFIRMED:
			return &sdk.Result{Log: fmt.Sprintf("outpoint %s already confirmed", msg.OutPoint.String())}, nil
		case types.SPENT:
			return &sdk.Result{Log: fmt.Sprintf("outpoint %s already spent", msg.OutPoint.String())}, nil
		default:
			panic(fmt.Sprintf("invalid outpoint state %v", state))
		}
	case !pollFound:
		return nil, fmt.Errorf("no outpoint found for poll %s", msg.Poll.String())
	case pendingOutPointInfo.OutPoint.String() != msg.OutPoint.String():
		return nil, fmt.Errorf("outpoint %s does not match poll %s", msg.OutPoint.String(), msg.Poll.String())
	default:
		// assert: the outpoint is known and has not been confirmed before
	}

	if err := v.TallyVote(ctx, msg.Sender, msg.Poll, msg.Confirmed); err != nil {
		return nil, err
	}

	result := v.Result(ctx, msg.Poll)
	if result == nil {
		return &sdk.Result{Log: fmt.Sprintf("not enough votes to confirm outpoint %s yet", msg.OutPoint.String())}, nil
	}

	// assert: the poll has completed
	confirmed, ok := result.(bool)
	if !ok {
		return nil, fmt.Errorf("result of poll %s has wrong type, expected bool, got %T", msg.Poll.String(), result)
	}

	v.DeletePoll(ctx, msg.Poll)
	k.DeletePendingOutPointInfo(ctx, msg.Poll)

	// handle poll result
	event := sdk.NewEvent(types.EventTypeOutpointConfirmation,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(types.AttributeKeyPoll, string(k.Codec().MustMarshalJSON(msg.Poll))),
		sdk.NewAttribute(types.AttributeKeyOutPointInfo, string(k.Codec().MustMarshalJSON(pendingOutPointInfo))))

	if !confirmed {
		ctx.EventManager().EmitEvent(
			event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueReject)))
		return &sdk.Result{
			Log:    fmt.Sprintf("outpoint %s was discarded ", msg.OutPoint.String()),
			Events: ctx.EventManager().Events(),
		}, nil
	}
	ctx.EventManager().EmitEvent(
		event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueConfirm)))

	k.SetOutpointInfo(ctx, pendingOutPointInfo, types.CONFIRMED)

	// TODO: handle withdrawals to deposit or consolidation addresses (this is currently undefined behaviour),
	//  i.e. multiple outpoints in the SignedTx need to be confirmed

	// if this is the consolidation outpoint it means the latest consolidation transaction is confirmed on Bitcoin
	if tx, ok := k.GetSignedTx(ctx); ok && tx.TxHash() == pendingOutPointInfo.OutPoint.Hash {
		k.DeleteSignedTx(ctx)
		return &sdk.Result{
			Events: ctx.EventManager().Events(),
			Log:    "confirmed consolidation transaction"}, nil
	}

	// handle cross-chain transfer
	depositAddr := nexus.CrossChainAddress{Address: pendingOutPointInfo.Address, Chain: exported.Bitcoin}
	amount := sdk.NewInt64Coin(exported.Bitcoin.NativeAsset, int64(pendingOutPointInfo.Amount))
	if err := n.EnqueueForTransfer(ctx, depositAddr, amount); err != nil {
		return nil, sdkerrors.Wrap(err, "cross-chain transfer failed")
	}

	return &sdk.Result{
		Events: ctx.EventManager().Events(),
		Log:    fmt.Sprintf("transfer of %s from {%s} successfully prepared", amount.Amount.String(), depositAddr.String()),
	}, nil
}

// HandleMsgSignPendingTransfers handles the signing of a consolidation transaction (consolidate confirmed outpoints and pay out transfers)
func HandleMsgSignPendingTransfers(ctx sdk.Context, k types.BTCKeeper, signer types.Signer, n types.Nexus, snapshotter types.Snapshotter, v types.Voter, msg types.MsgSignPendingTransfers) (*sdk.Result, error) {
	if _, ok := k.GetUnsignedTx(ctx); ok {
		return nil, fmt.Errorf("consolidation in progress")
	}
	if _, ok := k.GetSignedTx(ctx); ok {
		return nil, fmt.Errorf("previous consolidation transaction must be confirmed first")
	}

	outPuts, totalWithdrawals := prepareOutputs(ctx, k, n)
	prevOuts, totalDeposits, err := prepareInputs(ctx, k)
	if err != nil {
		return nil, err
	}

	change := totalDeposits.Sub(totalWithdrawals).SubRaw(int64(msg.Fee))
	switch change.Sign() {
	case -1:
		return nil, fmt.Errorf("not enough deposits (%s) to make all withdrawals (%s) with a transaction fee of %s",
			totalDeposits.String(), totalWithdrawals.String(), msg.Fee.String(),
		)
	case 0:
		k.Logger(ctx).Info("creating a transaction without change")
	case 1:
		changeOutput, err := prepareChange(ctx, k, signer, change)
		if err != nil {
			return nil, err
		}
		outPuts = append(outPuts, changeOutput)
	default:
		return nil, fmt.Errorf("sign value of change for consolidation transaction unexpected: %d", change.Sign())
	}

	tx, err := types.CreateTx(prevOuts, outPuts)
	if err != nil {
		return nil, err
	}
	k.SetUnsignedTx(ctx, tx)

	err = startSignInputs(ctx, signer, snapshotter, v, tx, prevOuts)
	if err != nil {
		return nil, err
	}

	return &sdk.Result{
		Events: ctx.EventManager().Events(),
		Log:    fmt.Sprintf("successfully started signing protocols to consolidate pending transfers"),
	}, nil
}

func prepareOutputs(ctx sdk.Context, k types.BTCKeeper, n types.Nexus) ([]types.Output, sdk.Int) {
	pendingTransfers := n.GetPendingTransfersForChain(ctx, exported.Bitcoin)
	var outPuts []types.Output
	totalOut := sdk.ZeroInt()
	for _, transfer := range pendingTransfers {
		recipient, err := btcutil.DecodeAddress(transfer.Recipient.Address, k.GetNetwork(ctx).Params())
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("%s is not a valid address", transfer.Recipient))
			continue
		}

		outPuts = append(outPuts,
			types.Output{Amount: btcutil.Amount(transfer.Asset.Amount.Int64()), Recipient: recipient})
		totalOut = totalOut.Add(transfer.Asset.Amount)
		n.ArchivePendingTransfer(ctx, transfer)
	}
	return outPuts, totalOut
}

func prepareInputs(ctx sdk.Context, k types.BTCKeeper) ([]types.OutPointToSign, sdk.Int, error) {
	var prevOuts []types.OutPointToSign
	totalDeposits := sdk.ZeroInt()
	for _, info := range k.GetConfirmedOutPointInfos(ctx) {
		addr, ok := k.GetAddress(ctx, info.Address)
		if !ok {
			return nil, sdk.ZeroInt(), fmt.Errorf("address for confirmed outpoint %s must be known", info.OutPoint.String())
		}
		prevOuts = append(prevOuts, types.OutPointToSign{OutPointInfo: info, AddressInfo: addr})
		totalDeposits = totalDeposits.AddRaw(int64(info.Amount))
		k.DeleteOutpointInfo(ctx, *info.OutPoint)
		k.SetOutpointInfo(ctx, info, types.SPENT)
	}
	return prevOuts, totalDeposits, nil
}

func prepareChange(ctx sdk.Context, k types.BTCKeeper, signer types.Signer, change sdk.Int) (types.Output, error) {
	if !change.IsInt64() {
		return types.Output{}, fmt.Errorf("the calculated change is too large for a single transaction")
	}

	// if a new master key has been assigned for rotation spend change to that address, otherwise use the current one
	key, ok := signer.GetNextKey(ctx, exported.Bitcoin, tss.MasterKey)
	if !ok {
		key, ok = signer.GetCurrentKey(ctx, exported.Bitcoin, tss.MasterKey)
		if !ok {
			return types.Output{}, fmt.Errorf("key not found")
		}
	}

	addr := types.NewConsolidationAddress(key, k.GetNetwork(ctx))
	k.SetAddress(ctx, addr)

	return types.Output{Amount: btcutil.Amount(change.Int64()), Recipient: addr.Address}, nil
}

func startSignInputs(ctx sdk.Context, signer types.Signer, snapshotter types.Snapshotter, v types.Voter, tx *wire.MsgTx, outpointsToSign []types.OutPointToSign) error {
	for i, in := range outpointsToSign {
		hash, err := txscript.CalcWitnessSigHash(in.RedeemScript, txscript.NewTxSigHashes(tx), txscript.SigHashAll, tx, i, int64(in.Amount))
		if err != nil {
			return err
		}

		counter, ok := signer.GetSnapshotCounterForKeyID(ctx, in.Key.ID)
		if !ok {
			return fmt.Errorf("no snapshot counter for key ID %s registered", in.Key.ID)
		}

		snapshot, ok := snapshotter.GetSnapshot(ctx, counter)
		if !ok {
			return fmt.Errorf("no snapshot found for counter num %d", counter)
		}

		sigID := hex.EncodeToString(hash)
		err = signer.StartSign(ctx, v, in.Key.ID, sigID, hash, snapshot)
		if err != nil {
			return err
		}
	}
	return nil
}

func checkLinkRequisites(ctx sdk.Context, s types.Signer, n types.Nexus, recipientChainName string) (tss.Key, nexus.Chain, error) {
	key, ok := s.GetCurrentKey(ctx, exported.Bitcoin, tss.MasterKey)
	if !ok {
		return tss.Key{}, nexus.Chain{}, fmt.Errorf("master key not set")
	}

	recipientChain, ok := n.GetChain(ctx, recipientChainName)
	if !ok {
		return tss.Key{}, nexus.Chain{}, fmt.Errorf("unknown recipient chain")
	}

	found := n.IsAssetRegistered(ctx, recipientChain.Name, exported.Bitcoin.NativeAsset)
	if !found {
		return tss.Key{}, nexus.Chain{}, fmt.Errorf("asset '%s' not registered for chain '%s'", exported.Bitcoin.NativeAsset, recipientChain.Name)
	}
	return key, recipientChain, nil
}
