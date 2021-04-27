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
		case *types.MsgLink:
			return HandleMsgLink(ctx, k, signer, n, msg)
		case *types.MsgConfirmOutpoint:
			return HandleMsgConfirmOutpoint(ctx, k, v, signer, msg)
		case *types.MsgVoteConfirmOutpoint:
			return HandleMsgVoteConfirmOutpoint(ctx, k, v, n, msg)
		case *types.MsgSignPendingTransfers:
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
func HandleMsgLink(ctx sdk.Context, k types.BTCKeeper, s types.Signer, n types.Nexus, msg *types.MsgLink) (*sdk.Result, error) {
	masterKey, ok := s.GetCurrentKey(ctx, exported.Bitcoin, tss.MasterKey)
	if !ok {
		return nil, fmt.Errorf("master key not set")
	}

	secondaryKey, ok := s.GetCurrentKey(ctx, exported.Bitcoin, tss.SecondaryKey)
	if !ok {
		return nil, fmt.Errorf("secondary key not set")
	}

	recipientChain, ok := n.GetChain(ctx, msg.RecipientChain)
	if !ok {
		return nil, fmt.Errorf("unknown recipient chain")
	}

	if !n.IsAssetRegistered(ctx, recipientChain.Name, exported.Bitcoin.NativeAsset) {
		return nil, fmt.Errorf("asset '%s' not registered for chain '%s'", exported.Bitcoin.NativeAsset, recipientChain.Name)
	}

	recipient := nexus.CrossChainAddress{Chain: recipientChain, Address: msg.RecipientAddr}
	depositAddr := types.NewLinkedAddress(masterKey, secondaryKey, k.GetNetwork(ctx), recipient)
	n.LinkAddresses(ctx, depositAddr.ToCrossChainAddr(), recipient)
	k.SetAddress(ctx, depositAddr)

	return &sdk.Result{
		Data:   []byte(depositAddr.EncodeAddress()),
		Log:    fmt.Sprintf("successfully linked {%s} and {%s}", depositAddr.ToCrossChainAddr().String(), recipient.String()),
		Events: ctx.EventManager().ABCIEvents(),
	}, nil
}

// HandleMsgConfirmOutpoint handles the confirmation of a Bitcoin outpoint
func HandleMsgConfirmOutpoint(ctx sdk.Context, k types.BTCKeeper, voter types.InitPoller, signer types.Signer, msg *types.MsgConfirmOutpoint) (*sdk.Result, error) {
	_, state, ok := k.GetOutPointInfo(ctx, msg.OutPointInfo.GetOutPoint())
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

	poll := vote.NewPollMetaWithNonce(types.ModuleName, msg.OutPointInfo.OutPoint, ctx.BlockHeight(), k.GetRevoteLockingPeriod(ctx))
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
		Log:    fmt.Sprintf("votes on confirmation of %s started", msg.OutPointInfo.OutPoint),
		Events: ctx.EventManager().ABCIEvents(),
	}, nil
}

// HandleMsgVoteConfirmOutpoint handles the votes on an outpoint confirmation
func HandleMsgVoteConfirmOutpoint(ctx sdk.Context, k types.BTCKeeper, v types.Voter, n types.Nexus, msg *types.MsgVoteConfirmOutpoint) (*sdk.Result, error) {
	// has the outpoint been confirmed before?
	confirmedOutPointInfo, state, confirmedBefore := k.GetOutPointInfo(ctx, *types.MustConvertOutPointFromStr(msg.OutPoint))
	// is there an ongoing poll?
	pendingOutPointInfo, pollFound := k.GetPendingOutPointInfo(ctx, msg.Poll)

	switch {
	// a malicious user could try to delete an ongoing poll by providing an already confirmed outpoint,
	// so we need to check that it matches the poll before deleting
	case confirmedBefore && pollFound && pendingOutPointInfo.OutPoint == confirmedOutPointInfo.OutPoint:
		v.DeletePoll(ctx, msg.Poll)
		k.DeletePendingOutPointInfo(ctx, msg.Poll)
		fallthrough
	// If the voting threshold has been met and additional votes are received they should not return an error
	case confirmedBefore:
		switch state {
		case types.CONFIRMED:
			return &sdk.Result{Log: fmt.Sprintf("outpoint %s already confirmed", msg.OutPoint)}, nil
		case types.SPENT:
			return &sdk.Result{Log: fmt.Sprintf("outpoint %s already spent", msg.OutPoint)}, nil
		default:
			panic(fmt.Sprintf("invalid outpoint state %v", state))
		}
	case !pollFound:
		return nil, fmt.Errorf("no outpoint found for poll %s", msg.Poll.String())
	case pendingOutPointInfo.OutPoint != msg.OutPoint:
		return nil, fmt.Errorf("outpoint %s does not match poll %s", msg.OutPoint, msg.Poll.String())
	default:
		// assert: the outpoint is known and has not been confirmed before
	}

	if err := v.TallyVote(ctx, msg.Sender, msg.Poll, msg.Confirmed); err != nil {
		return nil, err
	}

	result := v.Result(ctx, msg.Poll)
	if result == nil {
		return &sdk.Result{Log: fmt.Sprintf("not enough votes to confirm outpoint %s yet", msg.OutPoint)}, nil
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
			Log:    fmt.Sprintf("outpoint %s was discarded ", msg.OutPoint),
			Events: ctx.EventManager().ABCIEvents(),
		}, nil
	}
	ctx.EventManager().EmitEvent(
		event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueConfirm)))

	k.SetOutpointInfo(ctx, pendingOutPointInfo, types.CONFIRMED)

	// TODO: handle withdrawals to deposit or consolidation addresses (this is currently undefined behaviour),
	//  i.e. multiple outpoints in the SignedTx need to be confirmed

	// if this is the consolidation outpoint it means the latest consolidation transaction is confirmed on Bitcoin
	if tx, ok := k.GetSignedTx(ctx); ok && tx.TxHash() == pendingOutPointInfo.GetOutPoint().Hash {
		k.DeleteSignedTx(ctx)
		return &sdk.Result{
			Events: ctx.EventManager().ABCIEvents(),
			Log:    "confirmed consolidation transaction"}, nil
	}

	// handle cross-chain transfer
	depositAddr := nexus.CrossChainAddress{Address: pendingOutPointInfo.Address, Chain: exported.Bitcoin}
	amount := sdk.NewInt64Coin(exported.Bitcoin.NativeAsset, int64(pendingOutPointInfo.Amount))
	if err := n.EnqueueForTransfer(ctx, depositAddr, amount); err != nil {
		return nil, sdkerrors.Wrap(err, "cross-chain transfer failed")
	}

	return &sdk.Result{
		Events: ctx.EventManager().ABCIEvents(),
		Log:    fmt.Sprintf("transfer of %s from {%s} successfully prepared", amount.Amount.String(), depositAddr.String()),
	}, nil
}

// HandleMsgSignPendingTransfers handles the signing of a consolidation transaction (consolidate confirmed outpoints and pay out transfers)
func HandleMsgSignPendingTransfers(ctx sdk.Context, k types.BTCKeeper, signer types.Signer, n types.Nexus, snapshotter types.Snapshotter, v types.Voter, msg *types.MsgSignPendingTransfers) (*sdk.Result, error) {
	if _, ok := k.GetUnsignedTx(ctx); ok {
		return nil, fmt.Errorf("consolidation in progress")
	}
	if _, ok := k.GetSignedTx(ctx); ok {
		return nil, fmt.Errorf("previous consolidation transaction must be confirmed first")
	}

	outputs, totalWithdrawals := prepareOutputs(ctx, k, n)
	inputs, totalDeposits, err := prepareInputs(ctx, k, signer)
	if err != nil {
		return nil, err
	}

	change := totalDeposits.Sub(totalWithdrawals).SubRaw(msg.Fee)
	switch change.Sign() {
	case -1, 0:
		return nil, fmt.Errorf("not enough deposits (%s) to make all withdrawals (%s) with a transaction fee of %s",
			totalDeposits.String(), totalWithdrawals.String(), btcutil.Amount(msg.Fee).String(),
		)
	case 1:
		changeOutput, err := prepareChange(ctx, k, signer, change)
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, changeOutput)
		k.SetMasterKeyOutpointExists(ctx)
	default:
		return nil, fmt.Errorf("sign value of change for consolidation transaction unexpected: %d", change.Sign())
	}

	tx, err := types.CreateTx(inputs, outputs)
	if err != nil {
		return nil, err
	}
	k.SetUnsignedTx(ctx, tx)

	err = startSignInputs(ctx, signer, snapshotter, v, tx, inputs)
	if err != nil {
		return nil, err
	}

	return &sdk.Result{
		Events: ctx.EventManager().ABCIEvents(),
		Log:    "successfully started signing protocols to consolidate pending transfers",
	}, nil
}

func prepareOutputs(ctx sdk.Context, k types.BTCKeeper, n types.Nexus) ([]types.Output, sdk.Int) {
	minAmount := sdk.NewInt(int64(k.GetMinimumWithdrawalAmount(ctx)))

	pendingTransfers := n.GetPendingTransfersForChain(ctx, exported.Bitcoin)
	var outputs []types.Output
	totalOut := sdk.ZeroInt()

	addrWithdrawal := make(map[string]sdk.Int)
	// Combine output to same destination address
	for _, transfer := range pendingTransfers {
		recipient, err := btcutil.DecodeAddress(transfer.Recipient.Address, k.GetNetwork(ctx).Params())
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("%s is not a valid address", transfer.Recipient))
			continue
		}
		encodeAddr := recipient.EncodeAddress()
		if _, ok := addrWithdrawal[encodeAddr]; !ok {
			addrWithdrawal[encodeAddr] = sdk.ZeroInt()
		}
		addrWithdrawal[encodeAddr] = addrWithdrawal[encodeAddr].Add(transfer.Asset.Amount)

		n.ArchivePendingTransfer(ctx, transfer)
	}

	// Loop over pendingTransfer again for deterministic operation
	for _, transfer := range pendingTransfers {
		recipient, err := btcutil.DecodeAddress(transfer.Recipient.Address, k.GetNetwork(ctx).Params())
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("%s is not a valid address", transfer.Recipient))
			continue
		}

		encodeAddr := recipient.EncodeAddress()
		amount, ok := addrWithdrawal[encodeAddr]
		if !ok {
			continue
		}

		// delete from map to prevent recounting
		delete(addrWithdrawal, encodeAddr)

		// Check if the recipient has unsent dust amount
		unsentDust := k.GetDustAmount(ctx, encodeAddr)
		amount = amount.Add(sdk.NewInt(int64(unsentDust)))
		if amount.LT(minAmount) {
			// Set and continue
			k.SetDustAmount(ctx, recipient.EncodeAddress(), btcutil.Amount(amount.Int64()))
			event := sdk.NewEvent(types.EventTypeWithdrawalFailed,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
				sdk.NewAttribute(types.AttributeKeyDestinationAddress, recipient.EncodeAddress()),
				sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
				sdk.NewAttribute(sdk.EventTypeMessage, fmt.Sprintf("Withdrawal below minmum amount %s", minAmount)),
			)
			ctx.EventManager().EmitEvent(event)
			continue
		}

		if unsentDust > 0 {
			k.DeleteDustAmount(ctx, encodeAddr)
		}

		outputs = append(outputs,
			types.Output{Amount: btcutil.Amount(amount.Int64()), Recipient: recipient})

		totalOut = totalOut.Add(amount)
	}

	return outputs, totalOut
}

func prepareInputs(ctx sdk.Context, k types.BTCKeeper, signer types.Signer) ([]types.OutPointToSign, sdk.Int, error) {
	var prevOuts []types.OutPointToSign
	totalDeposits := sdk.ZeroInt()

	masterKeyUtxoExists := k.DoesMasterKeyOutpointExist(ctx)
	masterKeyUtxoFound := false

	for _, info := range k.GetConfirmedOutPointInfos(ctx) {
		addr, ok := k.GetAddress(ctx, info.Address)
		if !ok {
			return nil, sdk.ZeroInt(), fmt.Errorf("address for confirmed outpoint %s must be known", info.OutPoint)
		}

		key, found := signer.GetKey(ctx, addr.Key.ID)
		if !found {
			return nil, sdk.ZeroInt(), fmt.Errorf("key %s cannot be found", addr.Key.ID)
		}

		if key.Role == tss.Unknown {
			return nil, sdk.ZeroInt(), fmt.Errorf("key role not set for key %s", addr.Key.ID)
		}

		if key.Role == tss.MasterKey {
			masterKeyUtxoFound = true
		}

		prevOuts = append(prevOuts, types.OutPointToSign{OutPointInfo: info, AddressInfo: addr})
		totalDeposits = totalDeposits.AddRaw(int64(info.Amount))
		k.DeleteOutpointInfo(ctx, info.GetOutPoint())
		k.SetOutpointInfo(ctx, info, types.SPENT)
	}

	if masterKeyUtxoExists != masterKeyUtxoFound {
		return nil, sdk.ZeroInt(), fmt.Errorf("expect to spend UTXO of master key but not found")
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
