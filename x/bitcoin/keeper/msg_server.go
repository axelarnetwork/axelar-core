package keeper

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	gogoprototypes "github.com/gogo/protobuf/types"

	"github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

var _ types.MsgServiceServer = msgServer{}

type msgServer struct {
	types.BTCKeeper
	signer      types.Signer
	nexus       types.Nexus
	voter       types.Voter
	snapshotter types.Snapshotter
}

// NewMsgServerImpl returns an implementation of the bitcoin MsgServiceServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper types.BTCKeeper, s types.Signer, n types.Nexus, v types.Voter, snap types.Snapshotter) types.MsgServiceServer {
	return msgServer{
		BTCKeeper:   keeper,
		signer:      s,
		nexus:       n,
		voter:       v,
		snapshotter: snap,
	}
}

// Link handles address linking
func (s msgServer) Link(c context.Context, req *types.LinkRequest) (*types.LinkResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	masterKey, ok := s.signer.GetCurrentKey(ctx, exported.Bitcoin, tss.MasterKey)
	if !ok {
		return nil, fmt.Errorf("master key not set")
	}

	secondaryKey, ok := s.signer.GetCurrentKey(ctx, exported.Bitcoin, tss.SecondaryKey)
	if !ok {
		return nil, fmt.Errorf("secondary key not set")
	}

	recipientChain, ok := s.nexus.GetChain(ctx, req.RecipientChain)
	if !ok {
		return nil, fmt.Errorf("unknown recipient chain")
	}

	if !s.nexus.IsAssetRegistered(ctx, recipientChain.Name, exported.Bitcoin.NativeAsset) {
		return nil, fmt.Errorf("asset '%s' not registered for chain '%s'", exported.Bitcoin.NativeAsset, recipientChain.Name)
	}

	recipient := nexus.CrossChainAddress{Chain: recipientChain, Address: req.RecipientAddr}
	depositAddressInfo := types.NewLinkedAddress(masterKey, secondaryKey, s.GetNetwork(ctx), recipient)
	s.nexus.LinkAddresses(ctx, depositAddressInfo.ToCrossChainAddr(), recipient)
	s.SetAddress(ctx, depositAddressInfo)

	return &types.LinkResponse{DepositAddr: depositAddressInfo.Address}, nil
}

// ConfirmOutpoint handles the confirmation of a Bitcoin outpoint
func (s msgServer) ConfirmOutpoint(c context.Context, req *types.ConfirmOutpointRequest) (*types.ConfirmOutpointResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	_, state, ok := s.GetOutPointInfo(ctx, req.OutPointInfo.GetOutPoint())
	switch {
	case !ok:
		break
	case state == types.CONFIRMED:
		return nil, fmt.Errorf("already confirmed")
	case state == types.SPENT:
		return nil, fmt.Errorf("already spent")
	}

	if _, ok := s.GetAddress(ctx, req.OutPointInfo.Address); !ok {
		return nil, fmt.Errorf("outpoint address unknown, aborting deposit confirmation")
	}

	keyID, ok := s.signer.GetCurrentKeyID(ctx, exported.Bitcoin, tss.MasterKey)
	if !ok {
		return nil, fmt.Errorf("no master key for chain %s found", exported.Bitcoin.Name)
	}

	counter, ok := s.signer.GetSnapshotCounterForKeyID(ctx, keyID)
	if !ok {
		return nil, fmt.Errorf("no snapshot counter for key ID %s registered", keyID)
	}

	pollKey := vote.NewPollKey(types.ModuleName, req.OutPointInfo.OutPoint)
	if err := s.voter.InitializePoll(ctx, pollKey, counter, vote.ExpiryAt(ctx.BlockHeight()+s.BTCKeeper.GetRevoteLockingPeriod(ctx))); err != nil {
		return nil, err
	}
	s.SetPendingOutpointInfo(ctx, pollKey, req.OutPointInfo)

	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeOutpointConfirmation,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueStart),
		sdk.NewAttribute(types.AttributeKeyConfHeight, strconv.FormatUint(s.GetRequiredConfirmationHeight(ctx), 10)),
		sdk.NewAttribute(types.AttributeKeyOutPointInfo, string(types.ModuleCdc.MustMarshalJSON(&req.OutPointInfo))),
		sdk.NewAttribute(types.AttributeKeyPoll, string(types.ModuleCdc.MustMarshalJSON(&pollKey))),
	))

	return &types.ConfirmOutpointResponse{}, nil
}

// VoteConfirmOutpoint handles the votes on an outpoint confirmation
func (s msgServer) VoteConfirmOutpoint(c context.Context, req *types.VoteConfirmOutpointRequest) (*types.VoteConfirmOutpointResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	// has the outpoint been confirmed before?
	confirmedOutPointInfo, state, confirmedBefore := s.GetOutPointInfo(ctx, *types.MustConvertOutPointFromStr(req.OutPoint))
	// is there an ongoing poll?
	pendingOutPointInfo, pollFound := s.GetPendingOutPointInfo(ctx, req.PollKey)

	switch {
	// a malicious user could try to delete an ongoing poll by providing an already confirmed outpoint,
	// so we need to check that it matches the poll before deleting
	case confirmedBefore && pollFound && pendingOutPointInfo.OutPoint == confirmedOutPointInfo.OutPoint:
		s.DeletePendingOutPointInfo(ctx, req.PollKey)
		fallthrough
	// If the voting threshold has been met and additional votes are received they should not return an error
	case confirmedBefore:
		switch state {
		case types.CONFIRMED:
			return &types.VoteConfirmOutpointResponse{Status: fmt.Sprintf("outpoint %s already confirmed", req.OutPoint)}, nil
		case types.SPENT:
			return &types.VoteConfirmOutpointResponse{Status: fmt.Sprintf("outpoint %s already spent", req.OutPoint)}, nil
		default:
			panic(fmt.Sprintf("invalid outpoint state %v", state))
		}
	case !pollFound:
		return nil, fmt.Errorf("no outpoint found for poll %s", req.PollKey.String())
	case pendingOutPointInfo.OutPoint != req.OutPoint:
		return nil, fmt.Errorf("outpoint %s does not match poll %s", req.OutPoint, req.PollKey.String())
	default:
		// assert: the outpoint is known and has not been confirmed before
	}

	voter := s.snapshotter.GetOperator(ctx, req.Sender)
	if voter == nil {
		return nil, fmt.Errorf("account %v is not registered as a validator proxy", req.Sender.String())
	}

	poll := s.voter.GetPoll(ctx, req.PollKey)
	if err := poll.Vote(voter, &gogoprototypes.BoolValue{Value: req.Confirmed}); err != nil {
		return nil, err
	}

	if poll.Is(vote.Pending) {
		return &types.VoteConfirmOutpointResponse{Status: fmt.Sprintf("not enough votes to confirm outpoint %s yet", req.OutPoint)}, nil
	}

	if poll.Is(vote.Failed) {
		s.DeletePendingOutPointInfo(ctx, req.PollKey)
		return &types.VoteConfirmOutpointResponse{Status: fmt.Sprintf("poll %s failed", poll.GetKey())}, nil
	}

	confirmed, ok := poll.GetResult().(*gogoprototypes.BoolValue)
	if !ok {
		return nil, fmt.Errorf("result of poll %s has wrong type, expected bool, got %T", req.PollKey.String(), poll.GetResult())
	}

	logger := ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
	logger.Info(fmt.Sprintf("bitcoin outpoint confirmation result is %s", poll.GetResult()))
	s.DeletePendingOutPointInfo(ctx, req.PollKey)

	// handle poll result
	event := sdk.NewEvent(types.EventTypeOutpointConfirmation,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(types.AttributeKeyPoll, string(types.ModuleCdc.MustMarshalJSON(&req.PollKey))),
		sdk.NewAttribute(types.AttributeKeyOutPointInfo, string(types.ModuleCdc.MustMarshalJSON(&pendingOutPointInfo))))

	if !confirmed.Value {
		ctx.EventManager().EmitEvent(
			event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueReject)))

		return &types.VoteConfirmOutpointResponse{
			Status: fmt.Sprintf("outpoint %s was discarded ", req.OutPoint),
		}, nil
	}

	ctx.EventManager().EmitEvent(
		event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueConfirm)))

	addr, ok := s.GetAddress(ctx, pendingOutPointInfo.Address)
	if !ok {
		return nil, fmt.Errorf("cannot confirm outpoint of unknown address")
	}

	s.SetConfirmedOutpointInfo(ctx, addr.KeyID, pendingOutPointInfo)

	switch addr.Role {
	case types.Deposit:
		// handle cross-chain transfer
		depositAddr := nexus.CrossChainAddress{Address: pendingOutPointInfo.Address, Chain: exported.Bitcoin}
		amount := sdk.NewInt64Coin(exported.Bitcoin.NativeAsset, int64(pendingOutPointInfo.Amount))
		if err := s.nexus.EnqueueForTransfer(ctx, depositAddr, amount); err != nil {
			return nil, sdkerrors.Wrap(err, "cross-chain transfer failed")
		}

		return &types.VoteConfirmOutpointResponse{
			Status: fmt.Sprintf("transfer of %s from {%s} successfully prepared", amount.Amount.String(), depositAddr.String()),
		}, nil
	case types.Consolidation:
		// the outpoint simply deposits funds into a consolidation address. Simply confirm
		return &types.VoteConfirmOutpointResponse{
			Status: "confirmed top up of consolidation balance"}, nil
	default:
		return nil, fmt.Errorf("outpoint sends funds to address with unrecognized role")
	}
}

// SignPendingTransfers handles the signing of a consolidation transaction (consolidate confirmed outpoints and pay out transfers)
func (s msgServer) SignPendingTransfers(c context.Context, req *types.SignPendingTransfersRequest) (*types.SignPendingTransfersResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	if _, ok := s.GetUnsignedTx(ctx); ok {
		return nil, fmt.Errorf("consolidation in progress")
	}

	consolidationKey, ok := s.signer.GetKey(ctx, req.KeyID)
	if !ok {
		return nil, fmt.Errorf("unkown key %s", req.KeyID)
	}

	currSecondaryKey, ok := s.signer.GetCurrentKey(ctx, exported.Bitcoin, tss.SecondaryKey)
	if !ok {
		return nil, fmt.Errorf("current %s key is not set", tss.SecondaryKey.SimpleString())
	}

	nextSecondaryKey, nextSecondaryKeyAssigned := s.signer.GetNextKey(ctx, exported.Bitcoin, tss.SecondaryKey)
	if nextSecondaryKeyAssigned {
		return nil, fmt.Errorf("key %s is already assigned as the next %s key, rotate keys first", nextSecondaryKey.ID, tss.SecondaryKey.SimpleString())
	}

	outputs, totalOut := prepareOutputs(ctx, s, s.nexus)
	if len(outputs) == 0 {
		s.Logger(ctx).Info("creating consolidation transaction without any withdrawals")
	}

	anyoneCanSpendOutput := types.Output{Amount: s.BTCKeeper.GetMinimumWithdrawalAmount(ctx), Recipient: s.BTCKeeper.GetAnyoneCanSpendAddress(ctx).GetAddress()}
	totalOut = totalOut.AddRaw(int64(anyoneCanSpendOutput.Amount))

	inputs, totalDeposits, err := prepareInputs(ctx, s, s.signer, currSecondaryKey.ID)
	if err != nil {
		return nil, err
	}

	unsignedTx := types.Transaction{}

	// If consolidating to a new key, that key has to be eligible for the role
	if currSecondaryKey.ID != consolidationKey.ID {
		if err := s.signer.AssertMatchesRequirements(ctx, s.snapshotter, exported.Bitcoin, consolidationKey.ID, tss.SecondaryKey); err != nil {
			return nil, sdkerrors.Wrapf(err, "key %s does not match requirements for role %s", consolidationKey.ID, tss.SecondaryKey.SimpleString())
		}

		// TODO: think about the solution how to make sure the queue is never empty
		if !s.GetConfirmedOutpointInfoQueueForKey(ctx, currSecondaryKey.ID).IsEmpty() {
			return nil, sdkerrors.Wrapf(err, "key %s still has outpoints to be signed and therefore it cannot be rotated out yet", currSecondaryKey.ID)
		}

		unsignedTx.AssignNextKey = true
		unsignedTx.NextKeyRole = tss.SecondaryKey
		unsignedTx.NextKeyID = consolidationKey.ID
	}

	consolidationAddress := types.NewConsolidationAddress(consolidationKey, s.GetNetwork(ctx))
	txSizeUpperBound, err := estimateTxSizeWithZeroChange(ctx, s, consolidationAddress, inputs, append([]types.Output{anyoneCanSpendOutput}, outputs...))
	if err != nil {
		return nil, err
	}

	// consolidation transactions always pay 1 satoshi/byte, which is the default minimum relay fee rate bitcoin-core sets
	fee := sdk.NewInt(txSizeUpperBound).MulRaw(types.MinRelayTxFeeSatoshiPerByte)
	change := totalDeposits.Sub(totalOut).Sub(fee)

	if change.Sign() <= 0 {
		return nil, fmt.Errorf("not enough deposits (%s) to make all withdrawals (%s) with a transaction fee of %s",
			totalDeposits.String(), totalOut.String(), btcutil.Amount(fee.Int64()).String(),
		)
	}

	changeOutput, err := prepareChange(ctx, s, consolidationAddress, change)
	if err != nil {
		return nil, err
	}
	// vout 0 is always the change, and vout 1 is always anyone-can-spend for consolidation transaction
	outputs = append([]types.Output{changeOutput, anyoneCanSpendOutput}, outputs...)

	tx, err := types.CreateTx(inputs, outputs)
	if err != nil {
		return nil, err
	}

	unsignedTx.SetTx(tx)
	s.SetUnsignedTx(ctx, &unsignedTx)

	err = startSignInputs(ctx, s.signer, s.snapshotter, s.voter, tx, inputs)
	if err != nil {
		return nil, err
	}

	return &types.SignPendingTransfersResponse{}, nil
}

func estimateTxSizeWithZeroChange(ctx sdk.Context, k types.BTCKeeper, address types.AddressInfo, inputs []types.OutPointToSign, outputs []types.Output) (int64, error) {
	zeroChangeOutput, err := prepareChange(ctx, k, address, sdk.ZeroInt())
	if err != nil {
		return 0, err
	}

	tx, err := types.CreateTx(inputs, append(outputs, zeroChangeOutput))
	if err != nil {
		return 0, err
	}

	return types.EstimateTxSize(*tx, inputs), nil
}

func prepareOutputs(ctx sdk.Context, k types.BTCKeeper, n types.Nexus) ([]types.Output, sdk.Int) {
	minAmount := sdk.NewInt(int64(k.GetMinimumWithdrawalAmount(ctx)))
	pendingTransfers := n.GetTransfersForChain(ctx, exported.Bitcoin, nexus.Pending)
	outputs := []types.Output{}
	total := sdk.ZeroInt()
	network := k.GetNetwork(ctx).Params()

	// Combine output to same destination address
	for _, transfer := range pendingTransfers {
		if _, err := btcutil.DecodeAddress(transfer.Recipient.Address, network); err == nil {
			n.ArchivePendingTransfer(ctx, transfer)
		}
	}

	getRecipient := func(transfer nexus.CrossChainTransfer) string {
		return transfer.Recipient.Address
	}

	for _, transfer := range nexus.MergeTransfersBy(pendingTransfers, getRecipient) {
		recipient, err := btcutil.DecodeAddress(transfer.Recipient.Address, network)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("%s is not a valid address", transfer.Recipient.Address))
			continue
		}

		encodedAddress := recipient.EncodeAddress()
		amount := transfer.Asset.Amount

		// Check if the recipient has unsent dust amount
		unsentDust := k.GetDustAmount(ctx, encodedAddress)
		k.DeleteDustAmount(ctx, encodedAddress)

		amount = amount.Add(sdk.NewInt(int64(unsentDust)))
		if amount.LT(minAmount) {
			// Set and continue
			k.SetDustAmount(ctx, encodedAddress, btcutil.Amount(amount.Int64()))
			event := sdk.NewEvent(types.EventTypeWithdrawal,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
				sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueFailed),
				sdk.NewAttribute(types.AttributeKeyDestinationAddress, encodedAddress),
				sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
				sdk.NewAttribute(sdk.EventTypeMessage, fmt.Sprintf("Withdrawal below minmum amount %s", minAmount)),
			)
			ctx.EventManager().EmitEvent(event)
			continue
		}

		outputs = append(outputs,
			types.Output{Amount: btcutil.Amount(amount.Int64()), Recipient: recipient})
		total = total.Add(amount)
	}

	return outputs, total
}

func prepareInputs(ctx sdk.Context, k types.BTCKeeper, signer types.Signer, keyID string) ([]types.OutPointToSign, sdk.Int, error) {
	var inputs []types.OutPointToSign
	total := sdk.ZeroInt()

	// TODO: the confirmed outpoint info queue should by ordered by value desc instead of block height asc
	confirmedOutpointInfoQueue := k.GetConfirmedOutpointInfoQueueForKey(ctx, keyID)
	maxInputCount := k.GetMaxInputCount(ctx)

	var info types.OutPointInfo
	for len(inputs) < int(maxInputCount) && confirmedOutpointInfoQueue.Dequeue(&info) {
		addressInfo, ok := k.GetAddress(ctx, info.Address)
		if !ok {
			return nil, sdk.ZeroInt(), fmt.Errorf("address for confirmed outpoint %s must be known", info.OutPoint)
		}

		inputs = append(inputs, types.OutPointToSign{OutPointInfo: info, AddressInfo: addressInfo})
		total = total.AddRaw(int64(info.Amount))
		k.DeleteOutpointInfo(ctx, info.GetOutPoint())
		k.SetSpentOutpointInfo(ctx, info)
	}

	return inputs, total, nil
}

func prepareChange(ctx sdk.Context, k types.BTCKeeper, consolidationAddress types.AddressInfo, change sdk.Int) (types.Output, error) {
	if !change.IsInt64() {
		return types.Output{}, fmt.Errorf("the calculated change is too large for a single transaction")
	}

	k.SetAddress(ctx, consolidationAddress)

	telemetry.NewLabel("btc_secondary_addr", consolidationAddress.Address)
	telemetry.SetGauge(float32(change.Int64()), "btc_secondary_addr_balance")

	return types.Output{Amount: btcutil.Amount(change.Int64()), Recipient: consolidationAddress.GetAddress()}, nil
}

func startSignInputs(ctx sdk.Context, signer types.Signer, snapshotter types.Snapshotter, v types.Voter, tx *wire.MsgTx, outpointsToSign []types.OutPointToSign) error {
	for i, in := range outpointsToSign {
		hash, err := txscript.CalcWitnessSigHash(in.RedeemScript, txscript.NewTxSigHashes(tx), txscript.SigHashAll, tx, i, int64(in.Amount))
		if err != nil {
			return err
		}

		counter, ok := signer.GetSnapshotCounterForKeyID(ctx, in.KeyID)
		if !ok {
			return fmt.Errorf("no snapshot counter for key ID %s registered", in.KeyID)
		}

		snapshot, ok := snapshotter.GetSnapshot(ctx, counter)
		if !ok {
			return fmt.Errorf("no snapshot found for counter num %d", counter)
		}

		sigID := hex.EncodeToString(hash)
		err = signer.StartSign(ctx, v, in.KeyID, sigID, hash, snapshot)
		if err != nil {
			return err
		}
	}
	return nil
}
