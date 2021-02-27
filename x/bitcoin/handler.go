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

	"github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// NewHandler creates an sdk.Handler for all bitcoin type messages
func NewHandler(k keeper.Keeper, v types.Voter, rpc types.RPCClient, signer types.Signer, n types.Nexus) sdk.Handler {
	h := func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case types.MsgLink:
			return handleMsgLink(ctx, k, signer, n, msg)
		case types.MsgVerifyTx:
			return handleMsgVerifyTx(ctx, k, v, rpc, msg)
		case *types.MsgVoteVerifiedTx:
			return handleMsgVoteVerifiedTx(ctx, k, v, n, msg)
		case types.MsgSignPendingTransfers:
			return handleMsgSignPendingTransfers(ctx, k, signer, n, msg)
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

func handleMsgLink(ctx sdk.Context, k keeper.Keeper, s types.Signer, n types.Nexus, msg types.MsgLink) (*sdk.Result, error) {
	keyID, ok := s.GetCurrentMasterKeyID(ctx, exported.Bitcoin)
	if !ok {
		return nil, fmt.Errorf("master key not set")
	}

	masterKey, ok := s.GetKey(ctx, keyID)
	if !ok {
		return nil, fmt.Errorf("master key not set")
	}

	recipientChain, ok := n.GetChain(ctx, msg.RecipientChain)
	if !ok {
		return nil, sdkerrors.Wrap(types.ErrBitcoin, "unknown recipient chain")
	}

	recipient := nexus.CrossChainAddress{Chain: recipientChain, Address: msg.RecipientAddr}
	redeemScript, err := types.CreateCrossChainRedeemScript(btcec.PublicKey(masterKey), recipient)
	if err != nil {
		return nil, err
	}
	depositAddr, err := types.CreateDepositAddress(k.GetNetwork(ctx), redeemScript)
	if err != nil {
		return nil, err
	}

	encodedDepositAddr := depositAddr.EncodeAddress()
	deposit := nexus.CrossChainAddress{Chain: exported.Bitcoin, Address: encodedDepositAddr}

	n.LinkAddresses(ctx, deposit, recipient)

	// store script and key used to create the address for later reference
	k.SetRedeemScriptByAddress(ctx, depositAddr, redeemScript)
	k.SetKeyIDByAddress(ctx, depositAddr, keyID)

	return &sdk.Result{
		Data: []byte(encodedDepositAddr),
		Log:  fmt.Sprintf("successfully linked {%s} and {%s}", deposit.String(), recipient.String()),
	}, nil
}

func handleMsgVerifyTx(ctx sdk.Context, k keeper.Keeper, v types.Voter, rpc types.RPCClient, msg types.MsgVerifyTx) (*sdk.Result, error) {
	poll := vote.PollMeta{Module: types.ModuleName, Type: msg.Type(), ID: msg.OutPointInfo.OutPoint.String()}
	if err := v.InitPoll(ctx, poll); err != nil {
		return nil, err
	}

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
		v.RecordVote(&types.MsgVoteVerifiedTx{PollMeta: poll, VotingData: true})
		return &sdk.Result{Log: fmt.Sprintf("successfully verified outpoint %s", msg.OutPointInfo.OutPoint.String())}, nil
	// verification unsuccessful
	default:
		v.RecordVote(&types.MsgVoteVerifiedTx{PollMeta: poll, VotingData: false})
		return &sdk.Result{Log: sdkerrors.Wrapf(err, "outpoint %s not verified", msg.OutPointInfo.OutPoint.String()).Error()}, nil
	}
}

// This can be used as a potential hook to immediately act on a poll being decided by the vote
func handleMsgVoteVerifiedTx(ctx sdk.Context, k keeper.Keeper, v types.Voter, n types.Nexus, msg *types.MsgVoteVerifiedTx) (*sdk.Result, error) {
	// Check if the outpoint has been verified already.
	// If the voting threshold has been met and additional votes are received they should not return an error
	outPoint, err := types.OutPointFromStr(msg.PollMeta.ID)
	if err != nil {
		return nil, err
	}
	info, ok := k.GetVerifiedOutPointInfo(ctx, outPoint)
	if ok {
		return &sdk.Result{Log: fmt.Sprintf("outpoint %s already verified", info.OutPoint.String())}, nil
	}

	if err := v.TallyVote(ctx, msg); err != nil {
		return nil, err
	}

	result := v.Result(ctx, msg.Poll())
	if result == nil {
		return &sdk.Result{Log: fmt.Sprintf("not enough votes to verify outpoint %s yet", msg.PollMeta.ID)}, nil
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(sdk.EventTypeMessage, sdk.Attribute{
			Key:   types.AttributePollConfirmed,
			Value: strconv.FormatBool(result.(bool)),
		}))

	k.ProcessVerificationResult(ctx, outPoint, result.(bool))
	v.DeletePoll(ctx, msg.Poll())

	info, ok = k.GetVerifiedOutPointInfo(ctx, outPoint)
	if !ok {
		return &sdk.Result{Log: fmt.Sprintf("outpoint %s was discarded", msg.PollMeta.ID)}, nil
	}
	addr, err := btcutil.DecodeAddress(info.Address, k.GetNetwork(ctx).Params)
	if err != nil {
		return nil, err
	}
	keyID, ok := k.GetKeyIDByAddress(ctx, addr)
	if !ok {
		return nil, fmt.Errorf("key ID not found")
	}
	k.SetKeyIDByOutpoint(ctx, outPoint, keyID)

	depositAddr := nexus.CrossChainAddress{Address: info.Address, Chain: exported.Bitcoin}
	amount := sdk.NewInt64Coin(exported.Bitcoin.NativeAsset, int64(info.Amount))

	// outpoints that are not used as deposits for cross-chain transfers need to be verified as well (e.g. funds held by the master key).
	// Therefore, failing to enqueue for transfer is not an error
	if err = n.EnqueueForTransfer(ctx, depositAddr, amount); err != nil {
		return &sdk.Result{
			Events: ctx.EventManager().Events(),
			Log:    sdkerrors.Wrap(err, "prepared no transfer").Error()}, nil
	}

	return &sdk.Result{
		Events: ctx.EventManager().Events(),
		Log:    fmt.Sprintf("transfer of %s from {%s} successfully prepared", amount.Amount.String(), depositAddr.String()),
	}, nil
}

func handleMsgSignPendingTransfers(ctx sdk.Context, k keeper.Keeper, signer types.Signer, n types.Nexus, msg types.MsgSignPendingTransfers) (*sdk.Result, error) {
	outPuts, totalWithdrawals := prepareOutputs(ctx, k, n)
	prevOuts, totalDeposits := prepareInputs(ctx, k)

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
	k.SetRawTx(ctx, tx)

	err = startSignInputs(ctx, k, signer, tx)
	if err != nil {
		return nil, err
	}

	return &sdk.Result{Log: fmt.Sprintf("successfully started signing protocols to consolidate pending transfers")}, nil
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

func prepareChange(ctx sdk.Context, k keeper.Keeper, signer types.Signer, change sdk.Int) (types.Output, error) {
	if !change.IsInt64() {
		return types.Output{}, fmt.Errorf("the calculated change is too large for a single transaction")
	}

	// if a new master key has been assigned for rotation spend change to that address, otherwise use the current one
	keyID, ok := signer.GetNextMasterKeyID(ctx, exported.Bitcoin)
	if !ok {
		keyID, ok = signer.GetCurrentMasterKeyID(ctx, exported.Bitcoin)
		if !ok {
			return types.Output{}, fmt.Errorf("key not found")
		}
	}

	pk, ok := signer.GetKey(ctx, keyID)
	if !ok {
		return types.Output{}, fmt.Errorf("key not found")
	}

	redeemScript, err := types.CreateMasterRedeemScript(btcec.PublicKey(pk))
	if err != nil {
		return types.Output{}, err
	}
	masterAddr, err := types.CreateDepositAddress(k.GetNetwork(ctx), redeemScript)
	if err != nil {
		return types.Output{}, err
	}
	k.SetRedeemScriptByAddress(ctx, masterAddr, redeemScript)
	k.SetKeyIDByAddress(ctx, masterAddr, keyID)
	return types.Output{Amount: btcutil.Amount(change.Int64()), Recipient: masterAddr}, nil
}

func prepareInputs(ctx sdk.Context, k keeper.Keeper) ([]*wire.OutPoint, sdk.Int) {
	var prevOuts []*wire.OutPoint
	totalDeposits := sdk.ZeroInt()
	for _, info := range k.GetVerifiedOutPointInfos(ctx) {
		prevOuts = append(prevOuts, info.OutPoint)
		totalDeposits = totalDeposits.AddRaw(int64(info.Amount))
		k.SpendVerifiedOutPoint(ctx, info.OutPoint.String())
	}
	return prevOuts, totalDeposits
}

func prepareOutputs(ctx sdk.Context, k keeper.Keeper, n types.Nexus) ([]types.Output, sdk.Int) {
	pendingTransfers := n.GetPendingTransfersForChain(ctx, exported.Bitcoin)
	var outPuts []types.Output
	totalOut := sdk.ZeroInt()
	for _, transfer := range pendingTransfers {
		recipient, err := btcutil.DecodeAddress(transfer.Recipient.Address, k.GetNetwork(ctx).Params)
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

func startSignInputs(ctx sdk.Context, k keeper.Keeper, signer types.Signer, tx *wire.MsgTx) error {
	hashes, err := k.GetHashesToSign(ctx, tx)
	if err != nil {
		return err
	}
	for i, hash := range hashes {
		serializedHash := hex.EncodeToString(hash)

		keyID, ok := k.GetKeyIDByOutpoint(ctx, &tx.TxIn[i].PreviousOutPoint)
		if !ok {
			return fmt.Errorf("no key ID for chain %s found", exported.Bitcoin.Name)
		}

		err = signer.StartSign(ctx, keyID, serializedHash, hash)
		if err != nil {
			return err
		}
	}
	return nil
}
