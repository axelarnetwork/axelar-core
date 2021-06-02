package keeper

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
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

	poll := vote.NewPollMeta(types.ModuleName, req.OutPointInfo.OutPoint)
	if err := s.voter.InitPoll(ctx, poll, counter, ctx.BlockHeight()+s.BTCKeeper.GetRevoteLockingPeriod(ctx)); err != nil {
		return nil, err
	}
	s.SetPendingOutpointInfo(ctx, poll, req.OutPointInfo)

	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeOutpointConfirmation,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueStart),
		sdk.NewAttribute(types.AttributeKeyConfHeight, strconv.FormatUint(s.GetRequiredConfirmationHeight(ctx), 10)),
		sdk.NewAttribute(types.AttributeKeyOutPointInfo, string(types.ModuleCdc.MustMarshalJSON(&req.OutPointInfo))),
		sdk.NewAttribute(types.AttributeKeyPoll, string(types.ModuleCdc.MustMarshalJSON(&poll))),
	))

	return &types.ConfirmOutpointResponse{}, nil
}

// VoteConfirmOutpoint handles the votes on an outpoint confirmation
func (s msgServer) VoteConfirmOutpoint(c context.Context, req *types.VoteConfirmOutpointRequest) (*types.VoteConfirmOutpointResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	// has the outpoint been confirmed before?
	confirmedOutPointInfo, state, confirmedBefore := s.GetOutPointInfo(ctx, *types.MustConvertOutPointFromStr(req.OutPoint))
	// is there an ongoing poll?
	pendingOutPointInfo, pollFound := s.GetPendingOutPointInfo(ctx, req.Poll)

	switch {
	// a malicious user could try to delete an ongoing poll by providing an already confirmed outpoint,
	// so we need to check that it matches the poll before deleting
	case confirmedBefore && pollFound && pendingOutPointInfo.OutPoint == confirmedOutPointInfo.OutPoint:
		s.voter.DeletePoll(ctx, req.Poll)
		s.DeletePendingOutPointInfo(ctx, req.Poll)
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
		return nil, fmt.Errorf("no outpoint found for poll %s", req.Poll.String())
	case pendingOutPointInfo.OutPoint != req.OutPoint:
		return nil, fmt.Errorf("outpoint %s does not match poll %s", req.OutPoint, req.Poll.String())
	default:
		// assert: the outpoint is known and has not been confirmed before
	}

	if err := s.voter.TallyVote(ctx, req.Sender, req.Poll, &gogoprototypes.BoolValue{Value: req.Confirmed}); err != nil {
		return nil, err
	}

	result := s.voter.Result(ctx, req.Poll)
	if result == nil {
		return &types.VoteConfirmOutpointResponse{Status: fmt.Sprintf("not enough votes to confirm outpoint %s yet", req.OutPoint)}, nil
	}

	// assert: the poll has completed
	confirmed, ok := result.(*gogoprototypes.BoolValue)
	if !ok {
		return nil, fmt.Errorf("result of poll %s has wrong type, expected bool, got %T", req.Poll.String(), result)
	}

	logger := ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
	logger.Info(fmt.Sprintf("bitcoin outpoint confirmation result is %s", result))
	s.voter.DeletePoll(ctx, req.Poll)
	s.DeletePendingOutPointInfo(ctx, req.Poll)

	// handle poll result
	event := sdk.NewEvent(types.EventTypeOutpointConfirmation,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(types.AttributeKeyPoll, string(types.ModuleCdc.MustMarshalJSON(&req.Poll))),
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

	s.SetOutpointInfo(ctx, pendingOutPointInfo, types.CONFIRMED)
	addr, ok := s.GetAddress(ctx, pendingOutPointInfo.Address)
	if !ok {
		return nil, fmt.Errorf("cannot confirm outpoint of unknown address")
	}

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
		tx, txExist := s.GetSignedTx(ctx)
		vout, voutExist := s.GetMasterKeyVout(ctx)
		// TODO: both booleans should always have the same value, we might be able to make use of cosmos invariant checks to enforce it
		//  without the need to check it every call
		if txExist && voutExist {
			txHash := tx.TxHash()

			// if this is the consolidation outpoint it means the latest consolidation transaction is confirmed on Bitcoin
			if wire.NewOutPoint(&txHash, vout).String() == pendingOutPointInfo.OutPoint {
				s.DeleteSignedTx(ctx)
				return &types.VoteConfirmOutpointResponse{
					Status: "confirmed consolidation transaction"}, nil
			}
		}

		// the outpoint simply deposits funds into a consolidation address. Simply confirm
		return &types.VoteConfirmOutpointResponse{
			Status: "confirmed top up of consolidation balance"}, nil
	default:
		return nil, fmt.Errorf("outpoint sends funds to address with unrecognized role")
	}
}

// SignPendingTransfers handles the signing of a consolidation transaction (consolidate confirmed outpoints and pay out transfers)
func (s msgServer) SignPendingTransfers(c context.Context, _ *types.SignPendingTransfersRequest) (*types.SignPendingTransfersResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	if _, ok := s.GetUnsignedTx(ctx); ok {
		return nil, fmt.Errorf("consolidation in progress")
	}
	if tx, ok := s.GetSignedTx(ctx); ok {
		vout, _ := s.GetMasterKeyVout(ctx)
		return nil, fmt.Errorf("previous consolidation transaction %s:%d must be confirmed first", tx.TxHash().String(), vout)
	}

	outputs, totalOut := prepareOutputs(ctx, s, s.nexus)
	if len(outputs) == 0 {
		s.Logger(ctx).Info("creating consolidation transaction without any withdrawals")
	}
	inputs, totalDeposits, err := prepareInputs(ctx, s, s.signer)
	if err != nil {
		return nil, err
	}

	txSizeUpperBound, err := estimateTxSizeWithZeroChange(ctx, s, s.signer, inputs, outputs)
	if err != nil {
		return nil, err
	}

	// consolidation transactions always pay 1 satoshi/byte, which is the default minimum relay fee rate bitcoin-core sets
	fee := sdk.NewInt(txSizeUpperBound).MulRaw(types.MinRelayTxFeeSatoshiPerByte)
	change := totalDeposits.Sub(totalOut).Sub(fee)

	switch change.Sign() {
	case -1, 0:
		return nil, fmt.Errorf("not enough deposits (%s) to make all withdrawals (%s) with a transaction fee of %s",
			totalDeposits.String(), totalOut.String(), btcutil.Amount(fee.Int64()).String(),
		)
	case 1:
		changeOutput, err := prepareChange(ctx, s, s.signer, change)
		if err != nil {
			return nil, err
		}
		// vout 0 is always the change, and vout 1 is always anyone-can-spend
		outputs = append([]types.Output{changeOutput}, outputs...)
		s.SetMasterKeyVout(ctx, 0)
	default:
		return nil, fmt.Errorf("sign value of change for consolidation transaction unexpected: %d", change.Sign())
	}

	tx, err := types.CreateTx(inputs, outputs)
	if err != nil {
		return nil, err
	}
	s.SetUnsignedTx(ctx, tx)

	err = startSignInputs(ctx, s.signer, s.snapshotter, s.voter, tx, inputs)
	if err != nil {
		return nil, err
	}

	return &types.SignPendingTransfersResponse{}, nil
}

func estimateTxSizeWithZeroChange(ctx sdk.Context, k types.BTCKeeper, signer types.Signer, inputs []types.OutPointToSign, outputs []types.Output) (int64, error) {
	zeroChangeOutput, err := prepareChange(ctx, k, signer, sdk.ZeroInt())
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
	// first output in consolidation transaction is always for our anyone-can-spend address for the
	// sake of child-pay-for-parent so that anyone can pay
	anyoneCanSpendOutput := types.Output{Amount: k.GetMinimumWithdrawalAmount(ctx), Recipient: k.GetAnyoneCanSpendAddress(ctx).GetAddress()}
	outputs := []types.Output{anyoneCanSpendOutput}
	totalOut := sdk.NewInt(int64(anyoneCanSpendOutput.Amount))

	addrWithdrawal := make(map[string]sdk.Int)
	var recipients []btcutil.Address

	// Combine output to same destination address
	for _, transfer := range pendingTransfers {
		recipient, err := btcutil.DecodeAddress(transfer.Recipient.Address, k.GetNetwork(ctx).Params())
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("%s is not a valid address", transfer.Recipient.Address))
			continue
		}
		recipients = append(recipients, recipient)
		encodedAddress := recipient.EncodeAddress()

		if _, ok := addrWithdrawal[encodedAddress]; !ok {
			addrWithdrawal[encodedAddress] = sdk.ZeroInt()
		}
		addrWithdrawal[encodedAddress] = addrWithdrawal[encodedAddress].Add(transfer.Asset.Amount)

		n.ArchivePendingTransfer(ctx, transfer)
	}

	for _, recipient := range recipients {
		encodedAddress := recipient.EncodeAddress()
		amount, ok := addrWithdrawal[encodedAddress]
		if !ok {
			continue
		}

		// delete from map to prevent recounting
		delete(addrWithdrawal, encodedAddress)

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

		totalOut = totalOut.Add(amount)
	}

	return outputs, totalOut
}

func prepareInputs(ctx sdk.Context, k types.BTCKeeper, signer types.Signer) ([]types.OutPointToSign, sdk.Int, error) {
	var prevOuts []types.OutPointToSign
	totalDeposits := sdk.ZeroInt()

	_, masterKeyUtxoExists := k.GetMasterKeyVout(ctx)
	masterKeyUtxoFound := false

	for _, info := range k.GetConfirmedOutPointInfos(ctx) {
		addr, ok := k.GetAddress(ctx, info.Address)
		if !ok {
			return nil, sdk.ZeroInt(), fmt.Errorf("address for confirmed outpoint %s must be known", info.OutPoint)
		}

		key, found := signer.GetKey(ctx, addr.KeyID)
		if !found {
			return nil, sdk.ZeroInt(), fmt.Errorf("key %s cannot be found", addr.KeyID)
		}

		if key.Role == tss.Unknown {
			return nil, sdk.ZeroInt(), fmt.Errorf("key role not set for key %s", addr.KeyID)
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
		return nil, sdk.ZeroInt(), fmt.Errorf("previous consolidation outpoint must be confirmed first")
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

	addressInfo := types.NewConsolidationAddress(key, k.GetNetwork(ctx))
	k.SetAddress(ctx, addressInfo)

	return types.Output{Amount: btcutil.Amount(change.Int64()), Recipient: addressInfo.GetAddress()}, nil
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
