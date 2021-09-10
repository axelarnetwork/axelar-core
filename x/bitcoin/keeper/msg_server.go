package keeper

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/armon/go-metrics"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	gogoprototypes "github.com/gogo/protobuf/types"

	"github.com/axelarnetwork/axelar-core/utils"
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

func (s msgServer) SubmitExternalSignature(c context.Context, req *types.SubmitExternalSignatureRequest) (*types.SubmitExternalSignatureResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	externalKey, ok := s.signer.GetKey(ctx, req.KeyID)
	if !ok || externalKey.Role != tss.ExternalKey {
		return nil, fmt.Errorf("external key %s not found", req.KeyID)
	}

	sig, err := btcec.ParseDERSignature(req.Signature, btcec.S256())
	if err != nil {
		return nil, err
	}

	if !ecdsa.Verify(&externalKey.Value, req.SigHash, sig.R, sig.S) {
		return nil, fmt.Errorf("invalid signature for external key %s received", req.KeyID)
	}

	sigID := getSigID(req.SigHash, req.KeyID)
	s.signer.SetSig(ctx, sigID, req.Signature)
	s.signer.SetSigStatus(ctx, sigID, tss.SigStatus_Signed)

	info := tss.SignInfo{
		KeyID: req.KeyID,
		SigID: sigID,
	}
	s.signer.SetInfoForSig(ctx, sigID, info)

	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeExternalSignature,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueSubmitted),
		sdk.NewAttribute(types.AttributeKeyKeyID, req.KeyID),
		sdk.NewAttribute(types.AttributeKeySigID, sigID),
	))

	return &types.SubmitExternalSignatureResponse{}, nil
}

func (s msgServer) RegisterExternalKeys(c context.Context, req *types.RegisterExternalKeysRequest) (*types.RegisterExternalKeysResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	requiredExternalKeyCount := s.GetExternalMultisigThreshold(ctx).Denominator
	if len(req.ExternalKeys) != int(requiredExternalKeyCount) {
		return nil, fmt.Errorf("%d external keys are required", requiredExternalKeyCount)
	}

	keyIDs := make([]string, len(req.ExternalKeys))
	for i, externalKey := range req.ExternalKeys {
		if _, ok := s.signer.GetKey(ctx, externalKey.ID); ok {
			return nil, fmt.Errorf("external key ID %s is already used", externalKey.ID)
		}

		pubKey, err := btcec.ParsePubKey(externalKey.PubKey, btcec.S256())
		if err != nil {
			return nil, fmt.Errorf("invalid external key received")
		}

		s.signer.SetKey(ctx, externalKey.ID, *pubKey.ToECDSA())
		s.signer.SetKeyRole(ctx, externalKey.ID, tss.ExternalKey)
		keyIDs[i] = externalKey.ID

		ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeKey,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueAssigned),
			sdk.NewAttribute(types.AttributeKeyRole, tss.ExternalKey.SimpleString()),
			sdk.NewAttribute(types.AttributeKeyKeyID, externalKey.ID),
		))
	}

	s.SetExternalKeyIDs(ctx, keyIDs)

	return &types.RegisterExternalKeysResponse{}, nil
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
	depositAddressInfo := types.NewDepositAddress(masterKey, secondaryKey, s.GetNetwork(ctx), recipient)
	s.nexus.LinkAddresses(ctx, depositAddressInfo.ToCrossChainAddr(), recipient)
	s.SetAddress(ctx, depositAddressInfo)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeLink,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyMasterKeyID, masterKey.ID),
			sdk.NewAttribute(types.AttributeKeySecondaryKeyID, secondaryKey.ID),
			sdk.NewAttribute(types.AttributeKeyDepositAddress, depositAddressInfo.Address),
			sdk.NewAttribute(types.AttributeKeyDestinationChain, recipient.Chain.Name),
			sdk.NewAttribute(types.AttributeKeyDestinationAddress, recipient.Address),
		),
	)

	return &types.LinkResponse{DepositAddr: depositAddressInfo.Address}, nil
}

// ConfirmOutpoint handles the confirmation of a Bitcoin outpoint
func (s msgServer) ConfirmOutpoint(c context.Context, req *types.ConfirmOutpointRequest) (*types.ConfirmOutpointResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	_, state, ok := s.GetOutPointInfo(ctx, req.OutPointInfo.GetOutPoint())
	switch {
	case !ok:
		break
	case state == types.OutPointState_Confirmed:
		return nil, fmt.Errorf("already confirmed")
	case state == types.OutPointState_Spent:
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

	pollKey := vote.NewPollKey(types.ModuleName, fmt.Sprintf("%s_%s_%d", req.OutPointInfo.OutPoint, req.OutPointInfo.Address, req.OutPointInfo.Amount))
	if err := s.voter.InitializePoll(
		ctx,
		pollKey,
		counter,
		vote.ExpiryAt(ctx.BlockHeight()+s.BTCKeeper.GetRevoteLockingPeriod(ctx)),
		vote.Threshold(s.GetVotingThreshold(ctx)),
		vote.MinVoterCount(s.GetMinVoterCount(ctx)),
	); err != nil {
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
		case types.OutPointState_Confirmed:
			return &types.VoteConfirmOutpointResponse{Status: fmt.Sprintf("outpoint %s already confirmed", req.OutPoint)}, nil
		case types.OutPointState_Spent:
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
	voteValue := &gogoprototypes.BoolValue{Value: req.Confirmed}
	if err := poll.Vote(voter, voteValue); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeOutpointConfirmation,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueVoted),
		sdk.NewAttribute(types.AttributeKeyValue, strconv.FormatBool(voteValue.Value)),
	))

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

	// allow each return path to append unique attributes
	defer func() { ctx.EventManager().EmitEvent(event) }()

	if !confirmed.Value {
		poll.AllowOverride()
		event = event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueReject))

		return &types.VoteConfirmOutpointResponse{
			Status: fmt.Sprintf("outpoint %s was discarded ", req.OutPoint),
		}, nil
	}

	event = event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueConfirm))

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

		telemetry.IncrCounter(float32(pendingOutPointInfo.Amount), types.ModuleName, "total", "deposit")
		telemetry.IncrCounter(1, types.ModuleName, "total", "deposit", "count")

		recipient, ok := s.nexus.GetRecipient(ctx, depositAddr)
		if !ok {
			return nil, fmt.Errorf("cross-chain sender has no recipient")
		}
		event = event.AppendAttributes(
			sdk.NewAttribute(types.AttributeKeyDestinationChain, recipient.Chain.Name),
			sdk.NewAttribute(types.AttributeKeyDestinationAddress, recipient.Address),
		)

		return &types.VoteConfirmOutpointResponse{
			Status: fmt.Sprintf("transfer of %s from {%s} successfully prepared", amount.Amount.String(), depositAddr.String()),
		}, nil
	case types.Consolidation:
		unconfirmedAmount := s.BTCKeeper.GetUnconfirmedAmount(ctx, addr.KeyID)
		s.BTCKeeper.SetUnconfirmedAmount(ctx, addr.KeyID, unconfirmedAmount-pendingOutPointInfo.Amount)
		// the outpoint simply deposits funds into a consolidation address. Simply confirm
		return &types.VoteConfirmOutpointResponse{
			Status: "confirmed top up of consolidation balance"}, nil
	default:
		return nil, fmt.Errorf("outpoint sends funds to address with unrecognized role")
	}
}

func (s msgServer) SignTx(c context.Context, req *types.SignTxRequest) (*types.SignTxResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	unsignedTx, ok := s.GetUnsignedTx(ctx, req.KeyRole)
	if !ok || (!unsignedTx.Is(types.Created) && !unsignedTx.Is(types.Aborted)) {
		return nil, fmt.Errorf("no unsigned %s tx ready for signing", req.KeyRole.SimpleString())
	}

	s.Logger(ctx).Debug(fmt.Sprintf("signing %s consolidation transaction", req.KeyRole.SimpleString()))

	var maxLockTime *time.Time
	var outPointsToSign []types.OutPointToSign

	for _, txIn := range unsignedTx.GetTx().TxIn {
		outPointStr := txIn.PreviousOutPoint.String()
		outPointInfo, state, ok := s.BTCKeeper.GetOutPointInfo(ctx, txIn.PreviousOutPoint)
		if !ok || state != types.OutPointState_Spent {
			return nil, fmt.Errorf("out point info %s is not found or not spent", outPointStr)
		}

		addressInfo, ok := s.BTCKeeper.GetAddress(ctx, outPointInfo.Address)
		if !ok {
			return nil, fmt.Errorf("address for outpoint %s must be known", outPointStr)
		}

		outPointsToSign = append(outPointsToSign, types.OutPointToSign{OutPointInfo: outPointInfo, AddressInfo: addressInfo})

		if addressInfo.SpendingCondition.LockTime != nil && (maxLockTime == nil || addressInfo.SpendingCondition.LockTime.After(*maxLockTime)) {
			maxLockTime = addressInfo.SpendingCondition.LockTime
		}
	}

	tx := unsignedTx.GetTx()
	externalSigsRequired := false

	switch {
	// when some UTXOs cannot be spent without the external key
	case maxLockTime != nil && maxLockTime.After(ctx.BlockTime()):
		externalSigsRequired = true

		s.Logger(ctx).Debug(fmt.Sprintf("%s consolidation transaction requires external signatures", req.KeyRole.SimpleString()))
		fallthrough
	// when no UTXO is locked
	case maxLockTime == nil:
		tx.LockTime = 0
		tx = types.DisableTimelock(tx)

		s.Logger(ctx).Debug(fmt.Sprintf("disabled lock time on %s consolidation transaction", req.KeyRole.SimpleString()))
	// when all UTXOs can be spent without the external key
	default:
		tx.LockTime = uint32(maxLockTime.Unix())
		tx = types.EnableTimelock(tx)

		s.Logger(ctx).Debug(fmt.Sprintf("enabled lock time as %d on %s consolidation transaction", tx.LockTime, req.KeyRole.SimpleString()))
	}

	var sigHashes [][]byte
	// reset InputInfos each time to re-calculate the signatures that are needed
	unsignedTx.Info.InputInfos = []types.UnsignedTx_Info_InputInfo{}

	for i, outPointToSign := range outPointsToSign {
		sigHash, err := txscript.CalcWitnessSigHash(outPointToSign.RedeemScript, txscript.NewTxSigHashes(tx), txscript.SigHashAll, tx, i, int64(outPointToSign.Amount))
		if err != nil {
			return nil, err
		}

		sigHashes = append(sigHashes, sigHash)
		internalKeyIDs := outPointToSign.SpendingCondition.InternalKeyIds
		keyID := internalKeyIDs[0]
		// if the unsigned transaction has aborted due to signing failure, try signing with a different key if necessary and possible
		if unsignedTx.Is(types.Aborted) {
			keyID = internalKeyIDs[(utils.IndexOf(internalKeyIDs, unsignedTx.PrevAbortedKeyId)+1)%len(internalKeyIDs)]
		}

		unsignedTx.Info.InputInfos = append(unsignedTx.Info.InputInfos, types.UnsignedTx_Info_InputInfo{
			SigRequirements: []types.UnsignedTx_Info_InputInfo_SigRequirement{
				types.NewSigRequirement(keyID, sigHash),
			},
		})
	}

	if externalSigsRequired {
		// Verify that external keys have submitted all signatures
		for i, outpointToSign := range outPointsToSign {
			sigHash := sigHashes[i]

			requiredExternalSigCount := outpointToSign.AddressInfo.SpendingCondition.ExternalMultisigThreshold
			existingExternalSigCount := int64(0)

			for _, externalKeyID := range outpointToSign.AddressInfo.SpendingCondition.ExternalKeyIds {
				if existingExternalSigCount == requiredExternalSigCount {
					break
				}

				if _, status := s.signer.GetSig(ctx, getSigID(sigHash, externalKeyID)); status == tss.SigStatus_Signed {
					existingExternalSigCount++
					unsignedTx.Info.InputInfos[i].SigRequirements = append(
						unsignedTx.Info.InputInfos[i].SigRequirements,
						types.NewSigRequirement(externalKeyID, sigHash),
					)
				}

			}

			if existingExternalSigCount < requiredExternalSigCount {
				return nil, fmt.Errorf("not enough external signatures have been submitted yet for sig hash %s", hex.EncodeToString(sigHash))
			}
		}
	}

	for _, inputInfo := range unsignedTx.Info.InputInfos {
		for _, sigRequirement := range inputInfo.SigRequirements {
			sigID := getSigID(sigRequirement.SigHash, sigRequirement.KeyID)
			// if the signature already exists, ignore it
			if _, status := s.signer.GetSig(ctx, sigID); status == tss.SigStatus_Signed {
				s.Logger(ctx).Debug(fmt.Sprintf("signature %s for %s transaction exists already and therefore skipping", req.KeyRole.SimpleString(), sigID))
				continue
			}

			counter, ok := s.signer.GetSnapshotCounterForKeyID(ctx, sigRequirement.KeyID)
			if !ok {
				return nil, fmt.Errorf("no snapshot counter for key ID %s registered", sigRequirement.KeyID)
			}

			snapshot, ok := s.snapshotter.GetSnapshot(ctx, counter)
			if !ok {
				return nil, fmt.Errorf("no snapshot found for counter num %d", counter)
			}

			if _, err := s.signer.ScheduleSign(ctx, tss.SignInfo{
				KeyID:           sigRequirement.KeyID,
				SigID:           sigID,
				Msg:             sigRequirement.SigHash,
				SnapshotCounter: snapshot.Counter,
				RequestModule:   types.ModuleName,
				Metadata:        "",
			}); err != nil {
				return nil, err
			}
		}
	}

	unsignedTx.SetTx(tx)
	unsignedTx.Status = types.Signing
	s.SetUnsignedTx(ctx, req.KeyRole, unsignedTx)

	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeConsolidationTx,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueSigning),
		sdk.NewAttribute(types.AttributeKeyRole, req.KeyRole.SimpleString()),
	))

	return &types.SignTxResponse{}, nil
}

// CreateMasterTx creates a master key consolidation transaction
func (s msgServer) CreateMasterTx(c context.Context, req *types.CreateMasterTxRequest) (*types.CreateMasterTxResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	secondaryMin := s.GetMinOutputAmount(ctx)
	if req.SecondaryKeyAmount > 0 && req.SecondaryKeyAmount < secondaryMin {
		return nil, fmt.Errorf("cannot transfer %d to secondary key, it is below the minimum of %d", req.SecondaryKeyAmount, secondaryMin)
	}

	secondaryMax := s.GetMaxSecondaryOutputAmount(ctx)
	if req.SecondaryKeyAmount > secondaryMax {
		return nil, fmt.Errorf("cannot transfer %d to secondary key, it is above the maximum of %d", req.SecondaryKeyAmount, secondaryMax)
	}

	if _, ok := s.GetUnsignedTx(ctx, tss.MasterKey); ok {
		return nil, fmt.Errorf("consolidation in progress")
	}

	externalMultisigThreshold := s.GetExternalMultisigThreshold(ctx)
	externalKeys, err := getExternalKeys(ctx, s.BTCKeeper, s.signer)
	if err != nil {
		return nil, err
	}
	if len(externalKeys) != int(externalMultisigThreshold.Denominator) {
		return nil, fmt.Errorf("number of external keys does not match the threshold and re-register is needed")
	}

	consolidationKey, ok := s.signer.GetKey(ctx, req.KeyID)
	if !ok {
		return nil, fmt.Errorf("unkown key %s", req.KeyID)
	}

	currMasterKey, ok := s.signer.GetCurrentKey(ctx, exported.Bitcoin, tss.MasterKey)
	if !ok {
		return nil, fmt.Errorf("current %s key is not set", tss.MasterKey.SimpleString())
	}

	tx := types.CreateTx()

	inputsTotal, err := addInputs(ctx, s.BTCKeeper, tx, currMasterKey.ID)
	if err != nil {
		return nil, err
	}

	if err := types.AddOutput(tx, s.BTCKeeper.GetAnyoneCanSpendAddress(ctx).GetAddress(), s.BTCKeeper.GetMinOutputAmount(ctx)); err != nil {
		return nil, err
	}
	anyoneCanSpendVout := uint32(0)

	if req.SecondaryKeyAmount > 0 {
		var key tss.Key

		if nextKey, nextKeyFound := s.signer.GetNextKey(ctx, exported.Bitcoin, tss.SecondaryKey); nextKeyFound {
			key = nextKey
		} else if currKey, currKeyFound := s.signer.GetCurrentKey(ctx, exported.Bitcoin, tss.SecondaryKey); currKeyFound {
			key = currKey
		} else {
			return nil, fmt.Errorf("%s key not set", tss.SecondaryKey.SimpleString())
		}

		secondaryAddress, err := getSecondaryConsolidationAddress(ctx, s.BTCKeeper, key)
		if err != nil {
			return nil, err
		}

		if err := types.AddOutput(tx, secondaryAddress.GetAddress(), btcutil.Amount(req.SecondaryKeyAmount)); err != nil {
			return nil, err
		}
		s.SetAddress(ctx, secondaryAddress)
	}

	consolidationAddress, err := getMasterConsolidationAddress(ctx, s.BTCKeeper, s.signer, consolidationKey)
	if err != nil {
		return nil, err
	}

	txSizeUpperBound, err := estimateTxSizeWithOutputsTo(ctx, s, *tx, consolidationAddress.GetAddress())
	if err != nil {
		return nil, err
	}

	outputsTotal := types.GetOutputsTotal(*tx)
	// consolidation transactions always pay 1 satoshi/byte, which is the default minimum relay fee rate bitcoin-core sets
	fee := sdk.NewInt(txSizeUpperBound).MulRaw(types.MinRelayTxFeeSatoshiPerByte)
	change := inputsTotal.SubRaw(int64(outputsTotal)).Sub(fee)

	if change.Sign() <= 0 {
		return nil, fmt.Errorf("not enough inputs (%d) to cover the fee (%d) for master consolidation transaction", inputsTotal.Int64(), fee.Int64())
	}

	if err := types.AddOutput(tx, consolidationAddress.GetAddress(), btcutil.Amount(change.Int64())); err != nil {
		return nil, err
	}

	s.SetAddress(ctx, consolidationAddress)
	telemetry.SetGaugeWithLabels(
		[]string{types.ModuleName, "secondary", "address", "balance"},
		float32(change.Int64()),
		[]metrics.Label{
			telemetry.NewLabel("timestamp", strconv.FormatInt(time.Now().Unix(), 10)),
			telemetry.NewLabel("address", consolidationAddress.Address),
		})

	tx.LockTime = 0
	tx = types.DisableTimelock(tx)
	unsignedTx := types.NewUnsignedTx(tx, anyoneCanSpendVout, req.SecondaryKeyAmount)
	// If consolidating to a new key, that key has to be eligible for the role
	if currMasterKey.ID != consolidationKey.ID {
		if err := validateKeyAssignment(ctx, s.BTCKeeper, s.signer, s.snapshotter, currMasterKey, consolidationKey); err != nil {
			return nil, err
		}

		unsignedTx.Info.RotateKey = true
		if err := s.signer.AssignNextKey(ctx, exported.Bitcoin, consolidationKey.Role, consolidationKey.ID); err != nil {
			return nil, err
		}

		ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeKey,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueAssigned),
			sdk.NewAttribute(types.AttributeKeyRole, consolidationKey.Role.SimpleString()),
			sdk.NewAttribute(types.AttributeKeyKeyID, consolidationKey.ID),
		))
	}

	s.SetUnsignedTx(ctx, tss.MasterKey, unsignedTx)

	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeConsolidationTx,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueCreated),
		sdk.NewAttribute(types.AttributeKeyRole, tss.MasterKey.SimpleString()),
	))

	return &types.CreateMasterTxResponse{}, nil
}

// CreatePendingTransfersTx creates a secondary key consolidation transaction
func (s msgServer) CreatePendingTransfersTx(c context.Context, req *types.CreatePendingTransfersTxRequest) (*types.CreatePendingTransfersTxResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	masterMin := s.GetMinOutputAmount(ctx)
	if req.MasterKeyAmount > 0 && req.MasterKeyAmount < masterMin {
		return nil, fmt.Errorf("cannot transfer %d to the master key, it is below the minimum amount of %d", req.MasterKeyAmount, masterMin)
	}

	if _, ok := s.GetUnsignedTx(ctx, tss.SecondaryKey); ok {
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

	tx := types.CreateTx()

	inputsTotal, err := addInputs(ctx, s.BTCKeeper, tx, currSecondaryKey.ID)
	if err != nil {
		return nil, err
	}

	if err := types.AddOutput(tx, s.BTCKeeper.GetAnyoneCanSpendAddress(ctx).GetAddress(), s.BTCKeeper.GetMinOutputAmount(ctx)); err != nil {
		return nil, err
	}
	anyoneCanSpendVout := uint32(0)

	if req.MasterKeyAmount > 0 {
		var key tss.Key

		if nextKey, nextKeyFound := s.signer.GetNextKey(ctx, exported.Bitcoin, tss.MasterKey); nextKeyFound {
			key = nextKey
		} else if currKey, currKeyFound := s.signer.GetCurrentKey(ctx, exported.Bitcoin, tss.MasterKey); currKeyFound {
			key = currKey
		} else {
			return nil, fmt.Errorf("%s key not set", tss.MasterKey.SimpleString())
		}

		masterAddress, err := getMasterConsolidationAddress(ctx, s.BTCKeeper, s.signer, key)
		if err != nil {
			return nil, err
		}

		if err := types.AddOutput(tx, masterAddress.GetAddress(), btcutil.Amount(req.MasterKeyAmount)); err != nil {
			return nil, err
		}
		s.SetAddress(ctx, masterAddress)
	}

	consolidationAddress, err := getSecondaryConsolidationAddress(ctx, s.BTCKeeper, consolidationKey)
	if err != nil {
		return nil, err
	}

	if err := addWithdrawalOutputs(ctx, s.BTCKeeper, s.nexus, tx, consolidationAddress.GetAddress()); err != nil {
		return nil, err
	}

	txSizeUpperBound, err := estimateTxSizeWithOutputsTo(ctx, s, *tx, consolidationAddress.GetAddress())
	if err != nil {
		return nil, err
	}

	outputsTotal := types.GetOutputsTotal(*tx)
	// consolidation transactions always pay 1 satoshi/byte, which is the default minimum relay fee rate bitcoin-core sets
	fee := sdk.NewInt(txSizeUpperBound).MulRaw(types.MinRelayTxFeeSatoshiPerByte)
	change := inputsTotal.SubRaw(int64(outputsTotal)).Sub(fee)

	if change.Sign() <= 0 {
		return nil, fmt.Errorf("not enough deposits (%s) to make all withdrawals (%s) with a transaction fee of %s",
			inputsTotal.String(), outputsTotal.String(), btcutil.Amount(fee.Int64()).String(),
		)
	}

	if err := types.AddOutput(tx, consolidationAddress.GetAddress(), btcutil.Amount(change.Int64())); err != nil {
		return nil, err
	}

	s.SetAddress(ctx, consolidationAddress)
	telemetry.SetGaugeWithLabels(
		[]string{types.ModuleName, "secondary", "address", "balance"},
		float32(change.Int64()),
		[]metrics.Label{
			telemetry.NewLabel("timestamp", strconv.FormatInt(time.Now().Unix(), 10)),
			telemetry.NewLabel("address", consolidationAddress.Address),
		})

	tx.LockTime = 0
	tx = types.DisableTimelock(tx)
	unsignedTx := types.NewUnsignedTx(tx, anyoneCanSpendVout, req.MasterKeyAmount)
	// If consolidating to a new key, that key has to be eligible for the role
	if currSecondaryKey.ID != consolidationKey.ID {
		if err := validateKeyAssignment(ctx, s.BTCKeeper, s.signer, s.snapshotter, currSecondaryKey, consolidationKey); err != nil {
			return nil, err
		}

		unsignedTx.Info.RotateKey = true
		if err := s.signer.AssignNextKey(ctx, exported.Bitcoin, consolidationKey.Role, consolidationKey.ID); err != nil {
			return nil, err
		}

		ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeKey,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueAssigned),
			sdk.NewAttribute(types.AttributeKeyRole, consolidationKey.Role.SimpleString()),
			sdk.NewAttribute(types.AttributeKeyKeyID, consolidationKey.ID),
		))
	}

	s.SetUnsignedTx(ctx, tss.SecondaryKey, unsignedTx)

	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeConsolidationTx,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueCreated),
		sdk.NewAttribute(types.AttributeKeyRole, tss.SecondaryKey.SimpleString()),
	))

	return &types.CreatePendingTransfersTxResponse{}, nil
}

func getExternalKeys(ctx sdk.Context, k types.BTCKeeper, signer types.Signer) ([]tss.Key, error) {
	externalKeyIDs, ok := k.GetExternalKeyIDs(ctx)
	if !ok {
		return nil, fmt.Errorf("external keys not registered yet")
	}

	externalKeys := make([]tss.Key, len(externalKeyIDs))
	for i, externalKeyID := range externalKeyIDs {
		externalKey, ok := signer.GetKey(ctx, externalKeyID)
		if !ok || externalKey.Role != tss.ExternalKey {
			return nil, fmt.Errorf("external key %s not found", externalKeyID)
		}

		externalKeys[i] = externalKey
	}

	return externalKeys, nil
}

func estimateTxSizeWithOutputsTo(ctx sdk.Context, k types.BTCKeeper, tx wire.MsgTx, addresses ...btcutil.Address) (int64, error) {
	var outPointsToSign []types.OutPointToSign

	for _, txIn := range tx.TxIn {
		outPointStr := txIn.PreviousOutPoint.String()
		outPointInfo, _, ok := k.GetOutPointInfo(ctx, txIn.PreviousOutPoint)
		if !ok {
			return 0, fmt.Errorf("out point info %s is not found", outPointStr)
		}

		addressInfo, ok := k.GetAddress(ctx, outPointInfo.Address)
		if !ok {
			return 0, fmt.Errorf("address for outpoint %s must be known", outPointStr)
		}

		outPointsToSign = append(outPointsToSign, types.OutPointToSign{OutPointInfo: outPointInfo, AddressInfo: addressInfo})
	}

	for _, address := range addresses {
		if err := types.AddOutput(&tx, address, btcutil.Amount(0)); err != nil {
			return 0, err
		}
	}

	return types.EstimateTxSize(tx, outPointsToSign), nil
}

func addWithdrawalOutputs(ctx sdk.Context, k types.BTCKeeper, n types.Nexus, tx *wire.MsgTx, changeAddress btcutil.Address) error {
	total := sdk.ZeroInt()
	outputCount := 0
	minAmount := sdk.NewInt(int64(k.GetMinOutputAmount(ctx)))
	pendingTransfers := n.GetTransfersForChain(ctx, exported.Bitcoin, nexus.Pending)
	network := k.GetNetwork(ctx).Params()
	maxTxSize := k.GetMaxTxSize(ctx)

	addressToTransfers := make(map[string][]nexus.CrossChainTransfer)
	for _, transfer := range pendingTransfers {
		recipient, err := btcutil.DecodeAddress(transfer.Recipient.Address, network)
		if err != nil {
			continue
		}

		encodedAddress := recipient.EncodeAddress()
		addressToTransfers[encodedAddress] = append(addressToTransfers[encodedAddress], transfer)
	}

	getRecipient := func(transfer nexus.CrossChainTransfer) string {
		return transfer.Recipient.Address
	}

	// Combine output to same destination address
	for _, combinedTransfer := range nexus.MergeTransfersBy(pendingTransfers, getRecipient) {
		recipient, err := btcutil.DecodeAddress(combinedTransfer.Recipient.Address, network)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("%s is not a valid address", combinedTransfer.Recipient.Address))
			continue
		}

		encodedAddress := recipient.EncodeAddress()
		amount := combinedTransfer.Asset.Amount

		// Check if the recipient has unsent dust amount
		unsentDust := k.GetDustAmount(ctx, encodedAddress)
		k.DeleteDustAmount(ctx, encodedAddress)

		amount = amount.Add(sdk.NewInt(int64(unsentDust)))
		if amount.LT(minAmount) {
			// Set and continue
			k.SetDustAmount(ctx, encodedAddress, btcutil.Amount(amount.Int64()))

			ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeWithdrawal,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
				sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueFailed),
				sdk.NewAttribute(types.AttributeKeyDestinationAddress, encodedAddress),
				sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
				sdk.NewAttribute(sdk.EventTypeMessage, fmt.Sprintf("Withdrawal below minmum amount %s", minAmount)),
			))

			continue
		}

		if txSize, err := estimateTxSizeWithOutputsTo(ctx, k, *tx, recipient, changeAddress); err != nil {
			return err
		} else if txSize > maxTxSize {
			// stop if transaction size is above the limit after adding the ouput
			break
		}

		for _, transfer := range addressToTransfers[encodedAddress] {
			n.ArchivePendingTransfer(ctx, transfer)
		}

		total = total.Add(amount)
		outputCount++

		if err := types.AddOutput(tx, recipient, btcutil.Amount(amount.Int64())); err != nil {
			return err
		}
	}

	telemetry.IncrCounter(float32(total.Int64()), types.ModuleName, "total", "withdrawal")
	telemetry.IncrCounter(float32(outputCount), types.ModuleName, "total", "withdrawal", "count")

	if outputCount == 0 {
		k.Logger(ctx).Info("creating consolidation transaction without any withdrawals")
	}

	return nil
}

func addInputs(ctx sdk.Context, k types.BTCKeeper, tx *wire.MsgTx, keyID string) (sdk.Int, error) {
	total := sdk.ZeroInt()
	// TODO: the confirmed outpoint info queue should by ordered by value desc instead of block height asc
	confirmedOutpointInfoQueue := k.GetConfirmedOutpointInfoQueueForKey(ctx, keyID)
	maxInputCount := k.GetMaxInputCount(ctx)

	var info types.OutPointInfo
	for len(tx.TxIn) < int(maxInputCount) && confirmedOutpointInfoQueue.Dequeue(&info) {
		if err := types.AddInput(tx, info.OutPoint); err != nil {
			return total, err
		}

		total = total.AddRaw(int64(info.Amount))

		k.DeleteOutpointInfo(ctx, info.GetOutPoint())
		k.SetSpentOutpointInfo(ctx, info)
	}

	return total, nil
}

func getSigID(sigHash []byte, keyID string) string {
	return fmt.Sprintf("%s-%s", hex.EncodeToString(sigHash), keyID)
}

func validateKeyAssignment(ctx sdk.Context, k types.BTCKeeper, signer types.Signer, snapshotter types.Snapshotter, currKey tss.Key, nextKey tss.Key) error {
	var otherKeyRole tss.KeyRole

	switch currKey.Role {
	case tss.MasterKey:
		otherKeyRole = tss.SecondaryKey
	case tss.SecondaryKey:
		otherKeyRole = tss.MasterKey
	default:
		return fmt.Errorf("unknown key role %s", currKey.Role.SimpleString())
	}

	if unsignedTx, ok := k.GetUnsignedTx(ctx, otherKeyRole); ok && unsignedTx.InternalTransferAmount > 0 {
		return fmt.Errorf("cannot assign the next %s key while a %s transaction is sending coin to the current %s address",
			currKey.Role.SimpleString(),
			otherKeyRole.SimpleString(),
			currKey.Role.SimpleString(),
		)
	}

	if err := signer.AssertMatchesRequirements(ctx, snapshotter, exported.Bitcoin, nextKey.ID, currKey.Role); err != nil {
		return sdkerrors.Wrapf(err, "key %s does not match requirements for role %s", nextKey.ID, currKey.Role.SimpleString())
	}

	// TODO: How do we prevent the queue being always non-empty?
	if !k.GetConfirmedOutpointInfoQueueForKey(ctx, currKey.ID).IsEmpty() {
		return fmt.Errorf("key %s still has outpoints to be signed and therefore it cannot be rotated out yet", currKey.ID)
	}

	if k.GetUnconfirmedAmount(ctx, currKey.ID) > 0 {
		return fmt.Errorf("key %s still has unconfirmed outpoints and therefore it cannot be rotated out yet", currKey.ID)
	}

	return nil
}

func getSecondaryConsolidationAddress(ctx sdk.Context, k types.BTCKeeper, key tss.Key) (types.AddressInfo, error) {
	if key.Role != tss.SecondaryKey {
		return types.AddressInfo{}, fmt.Errorf("given key %s is not for a %s key", key.ID, tss.SecondaryKey.SimpleString())
	}

	consolidationAddress := types.NewSecondaryConsolidationAddress(key, k.GetNetwork(ctx))

	return consolidationAddress, nil
}

func getMasterConsolidationAddress(ctx sdk.Context, k types.BTCKeeper, s types.Signer, key tss.Key) (types.AddressInfo, error) {
	if key.Role != tss.MasterKey {
		return types.AddressInfo{}, fmt.Errorf("given key %s is not for a %s key", key.ID, tss.MasterKey.SimpleString())
	}

	currMasterKey, ok := s.GetCurrentKey(ctx, exported.Bitcoin, tss.MasterKey)
	if !ok {
		return types.AddressInfo{}, fmt.Errorf("%s key not set", tss.MasterKey.SimpleString())
	}

	oldMasterKey, ok := getOldMasterKey(ctx, k, s)
	if !ok {
		return types.AddressInfo{}, fmt.Errorf("cannot find the old %s key", tss.MasterKey.SimpleString())
	}

	externalMultisigThreshold := k.GetExternalMultisigThreshold(ctx)
	externalKeys, err := getExternalKeys(ctx, k, s)
	if err != nil {
		return types.AddressInfo{}, err
	}
	if len(externalKeys) != int(externalMultisigThreshold.Denominator) {
		return types.AddressInfo{}, fmt.Errorf("number of external keys does not match the threshold and re-register is needed")
	}

	internalKeyLockTime := currMasterKey.RotatedAt.Add(k.GetMasterAddressInternalKeyLockDuration(ctx))
	externalKeyLockTime := currMasterKey.RotatedAt.Add(k.GetMasterAddressExternalKeyLockDuration(ctx))
	consolidationAddress := types.NewMasterConsolidationAddress(key, oldMasterKey, externalMultisigThreshold.Numerator, externalKeys, internalKeyLockTime, externalKeyLockTime, k.GetNetwork(ctx))

	return consolidationAddress, nil
}

func getOldMasterKey(ctx sdk.Context, k types.BTCKeeper, signer types.Signer) (tss.Key, bool) {
	currRotationCount := signer.GetRotationCount(ctx, exported.Bitcoin, tss.MasterKey)
	oldMasterKeyRotationCount := currRotationCount - k.GetMasterKeyRetentionPeriod(ctx)
	if oldMasterKeyRotationCount < 1 {
		oldMasterKeyRotationCount = 1
	}

	return signer.GetKeyByRotationCount(ctx, exported.Bitcoin, tss.MasterKey, oldMasterKeyRotationCount)
}
