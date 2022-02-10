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

	if err := validateChainActivated(ctx, s.nexus, exported.Bitcoin); err != nil {
		return nil, err
	}

	externalKey, ok := s.signer.GetKey(ctx, req.KeyID)
	if !ok || externalKey.Role != tss.ExternalKey {
		return nil, fmt.Errorf("external key %s not found", req.KeyID)
	}

	pk, err := externalKey.GetECDSAPubKey()
	if err != nil {
		return nil, err
	}

	sig, err := btcec.ParseDERSignature(req.Signature, btcec.S256())
	if err != nil {
		return nil, err
	}
	if !ecdsa.Verify(&pk, req.SigHash, sig.R, sig.S) {
		return nil, fmt.Errorf("invalid signature for external key %s received", req.KeyID)
	}

	sigID := getSigID(req.SigHash, req.KeyID)
	btcecPK := btcec.PublicKey(pk)
	s.signer.SetSig(ctx, tss.Signature{
		SigID: sigID,
		Sig: &tss.Signature_SingleSig_{
			SingleSig: &tss.Signature_SingleSig{
				SigKeyPair: tss.SigKeyPair{
					PubKey:    btcecPK.SerializeCompressed(),
					Signature: req.Signature,
				},
			},
		},
		SigStatus: tss.SigStatus_Signed,
	})

	info := tss.SignInfo{
		KeyID: req.KeyID,
		SigID: sigID,
	}
	s.signer.SetInfoForSig(ctx, sigID, info)

	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeExternalSignature,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueSubmitted),
		sdk.NewAttribute(types.AttributeKeyKeyID, string(req.KeyID)),
		sdk.NewAttribute(types.AttributeKeySigID, sigID),
	))

	return &types.SubmitExternalSignatureResponse{}, nil
}

// Link handles address linking
func (s msgServer) Link(c context.Context, req *types.LinkRequest) (*types.LinkResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	if err := validateChainActivated(ctx, s.nexus, exported.Bitcoin); err != nil {
		return nil, err
	}

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

	if !s.nexus.IsAssetRegistered(ctx, recipientChain, types.Satoshi) {
		return nil, fmt.Errorf("asset '%s' not registered for chain '%s'", types.Satoshi, recipientChain.Name)
	}

	recipient := nexus.CrossChainAddress{Chain: recipientChain, Address: req.RecipientAddr}

	depositAddressInfo, err := getDepositAddress(ctx, s.BTCKeeper, s.signer, secondaryKey, recipient)
	if err != nil {
		return nil, err
	}

	addr, err := btcutil.DecodeAddress(depositAddressInfo.Address, s.GetNetwork(ctx).Params())
	if err != nil {
		return nil, err
	}

	err = s.nexus.LinkAddresses(ctx, depositAddressInfo.ToCrossChainAddr(), recipient)
	if err != nil {
		return nil, fmt.Errorf("could not link addresses: %s", err.Error())
	}

	s.SetAddressInfo(ctx, depositAddressInfo)
	s.SetDepositAddress(ctx, recipient, addr)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeLink,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyMasterKeyID, string(masterKey.ID)),
			sdk.NewAttribute(types.AttributeKeySecondaryKeyID, string(secondaryKey.ID)),
			sdk.NewAttribute(types.AttributeKeyDepositAddress, depositAddressInfo.Address),
			sdk.NewAttribute(types.AttributeKeySourceChain, exported.Bitcoin.Name),
			sdk.NewAttribute(types.AttributeKeyDestinationChain, recipient.Chain.Name),
			sdk.NewAttribute(types.AttributeKeyDestinationAddress, recipient.Address),
			sdk.NewAttribute(types.AttributeKeyAsset, types.Satoshi),
		),
	)

	return &types.LinkResponse{DepositAddr: depositAddressInfo.Address}, nil
}

// ConfirmOutpoint handles the confirmation of a Bitcoin outpoint
func (s msgServer) ConfirmOutpoint(c context.Context, req *types.ConfirmOutpointRequest) (*types.ConfirmOutpointResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	if err := validateChainActivated(ctx, s.nexus, exported.Bitcoin); err != nil {
		return nil, err
	}

	_, state, ok := s.GetOutPointInfo(ctx, req.OutPointInfo.GetOutPoint())
	switch {
	case !ok:
		break
	case state == types.OutPointState_Confirmed:
		return nil, fmt.Errorf("already confirmed")
	case state == types.OutPointState_Spent:
		return nil, fmt.Errorf("already spent")
	}

	if _, ok := s.GetAddressInfo(ctx, req.OutPointInfo.Address); !ok {
		return nil, fmt.Errorf("outpoint address unknown, aborting deposit confirmation")
	}

	pollKey := vote.NewPollKey(types.ModuleName, fmt.Sprintf("%s_%s_%d", req.OutPointInfo.OutPoint, req.OutPointInfo.Address, req.OutPointInfo.Amount))
	if err := s.voter.InitializePoll(
		ctx,
		pollKey,
		s.nexus.GetChainMaintainers(ctx, exported.Bitcoin),
		vote.ExpiryAt(ctx.BlockHeight()+s.BTCKeeper.GetRevoteLockingPeriod(ctx)),
		vote.Threshold(s.GetVotingThreshold(ctx)),
		vote.MinVoterCount(s.GetMinVoterCount(ctx)),
		vote.RewardPool(exported.Bitcoin.Name),
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

	if err := validateChainActivated(ctx, s.nexus, exported.Bitcoin); err != nil {
		return nil, err
	}

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

	s.Logger(ctx).Info(fmt.Sprintf("outpoint %s was confirmed ", req.OutPoint))
	event = event.AppendAttributes(sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueConfirm))

	addr, ok := s.GetAddressInfo(ctx, pendingOutPointInfo.Address)
	if !ok {
		return nil, fmt.Errorf("cannot confirm outpoint of unknown address")
	}

	s.SetConfirmedOutpointInfo(ctx, addr.KeyID, pendingOutPointInfo)

	key, ok := s.signer.GetKey(ctx, addr.KeyID)
	if !ok {
		return nil, fmt.Errorf("cannot key %s", addr.KeyID)
	}

	rotationCount, ok := s.signer.GetRotationCountOfKeyID(ctx, addr.KeyID)
	if !ok {
		return nil, fmt.Errorf("cannot find rotation count of key %s", addr.KeyID)
	}

	currRotationCount := s.signer.GetRotationCount(ctx, exported.Bitcoin, key.Role)
	_, nextKeyFound := s.signer.GetNextKey(ctx, exported.Bitcoin, key.Role)
	if nextKeyFound {
		currRotationCount++
	}

	if currRotationCount-rotationCount > s.signer.GetKeyUnbondingLockingKeyRotationCount(ctx) {
		return &types.VoteConfirmOutpointResponse{
			Status: fmt.Sprintf("cannot confirm outpoint of the old key %s anymore", addr.KeyID)}, nil
	}

	switch addr.Role {
	case types.Deposit:
		// handle cross-chain transfer
		depositAddr := nexus.CrossChainAddress{Address: pendingOutPointInfo.Address, Chain: exported.Bitcoin}
		amount := sdk.NewInt64Coin(types.Satoshi, int64(pendingOutPointInfo.Amount))
		if err := s.nexus.EnqueueForTransfer(ctx, depositAddr, amount, s.GetTransactionFeeRate(ctx)); err != nil {
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

	if err := validateChainActivated(ctx, s.nexus, exported.Bitcoin); err != nil {
		return nil, err
	}

	unsignedTx, ok := s.GetUnsignedTx(ctx, req.TxType)
	if !ok || (!unsignedTx.Is(types.Created) && !unsignedTx.Is(types.Aborted)) {
		return nil, fmt.Errorf("no unsigned %s tx ready for signing", req.TxType.SimpleString())
	}

	s.Logger(ctx).Debug(fmt.Sprintf("signing %s transaction", req.TxType.SimpleString()))

	var maxLockTime *time.Time
	var outPointsToSign []types.OutPointToSign

	for _, txIn := range unsignedTx.GetTx().TxIn {
		outPointStr := txIn.PreviousOutPoint.String()
		outPointInfo, state, ok := s.BTCKeeper.GetOutPointInfo(ctx, txIn.PreviousOutPoint)
		if !ok || state != types.OutPointState_Spent {
			return nil, fmt.Errorf("out point info %s is not found or not spent", outPointStr)
		}

		addressInfo, ok := s.BTCKeeper.GetAddressInfo(ctx, outPointInfo.Address)
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

		s.Logger(ctx).Debug(fmt.Sprintf("%s  transaction requires external signatures", req.TxType.SimpleString()))
		fallthrough
	// when no UTXO is locked
	case maxLockTime == nil:
		tx = types.DisableTimelock(tx)

		s.Logger(ctx).Debug(fmt.Sprintf("disabled lock time on %s  transaction", req.TxType.SimpleString()))
	// when all UTXOs can be spent without the external key
	default:
		tx = types.EnableTimelock(tx, uint32(maxLockTime.Unix()))

		s.Logger(ctx).Debug(fmt.Sprintf("enabled lock time as %d on %s  transaction", tx.LockTime, req.TxType.SimpleString()))
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
			keyID = internalKeyIDs[(utils.IndexOf(tss.KeyIDsToStrings(internalKeyIDs), string(unsignedTx.PrevAbortedKeyId))+1)%len(internalKeyIDs)]
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
	// track sign info position in queue
	pos := int64(0)
	var err error
	for _, inputInfo := range unsignedTx.Info.InputInfos {
		for _, sigRequirement := range inputInfo.SigRequirements {
			sigID := getSigID(sigRequirement.SigHash, sigRequirement.KeyID)
			// if the signature already exists, ignore it
			if _, status := s.signer.GetSig(ctx, sigID); status == tss.SigStatus_Signed {
				s.Logger(ctx).Debug(fmt.Sprintf("signature %s for %s transaction exists already and therefore skipping", req.TxType.SimpleString(), sigID))
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

			err = s.signer.StartSign(ctx, tss.SignInfo{
				KeyID:           sigRequirement.KeyID,
				SigID:           sigID,
				Msg:             sigRequirement.SigHash,
				SnapshotCounter: snapshot.Counter,
				RequestModule:   types.ModuleName,
				Metadata:        "",
			}, s.snapshotter, s.voter)
			if err != nil {
				return nil, err
			}
		}
	}

	unsignedTx.SetTx(tx)
	unsignedTx.Status = types.Signing
	s.SetUnsignedTx(ctx, unsignedTx)

	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeConsolidationTx,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueSigning),
		sdk.NewAttribute(types.AttributeTxType, req.TxType.SimpleString()),
	))

	return &types.SignTxResponse{Position: pos}, nil
}

// CreateRescueTx creates a rescue transaction
func (s msgServer) CreateRescueTx(c context.Context, req *types.CreateRescueTxRequest) (*types.CreateRescueTxResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	if err := validateChainActivated(ctx, s.nexus, exported.Bitcoin); err != nil {
		return nil, err
	}

	if _, ok := s.GetUnsignedTx(ctx, types.Rescue); ok {
		return nil, fmt.Errorf("rescue in progress")
	}

	var latestSecondaryKey tss.Key
	if nextKey, nextKeyFound := s.signer.GetNextKey(ctx, exported.Bitcoin, tss.SecondaryKey); nextKeyFound {
		latestSecondaryKey = nextKey
	} else if currKey, currKeyFound := s.signer.GetCurrentKey(ctx, exported.Bitcoin, tss.SecondaryKey); currKeyFound {
		latestSecondaryKey = currKey
	} else {
		return nil, fmt.Errorf("%s key not set", tss.SecondaryKey.SimpleString())
	}

	// always rescue to the latest secondary key
	secondaryAddress, err := getSecondaryConsolidationAddress(ctx, s.BTCKeeper, latestSecondaryKey)
	if err != nil {
		return nil, err
	}

	tx := types.CreateTx()
	inputsTotal := sdk.ZeroInt()

	for _, keyRole := range tss.GetKeyRoles() {
		oldActiveKeys, err := s.signer.GetOldActiveKeys(ctx, exported.Bitcoin, keyRole)
		if err != nil {
			return nil, err
		}

		for _, oldActiveKey := range oldActiveKeys {
			total, err := addInputs(ctx, s.BTCKeeper, tx, oldActiveKey.ID)
			if err != nil {
				return nil, err
			}

			if total.IsPositive() {
				s.Logger(ctx).Debug("rescuing UTXOs of old %s key %s", keyRole.SimpleString(), oldActiveKey.ID)
				inputsTotal = inputsTotal.Add(total)
			}
		}
	}

	if len(tx.TxIn) == 0 {
		return nil, fmt.Errorf("no rescue needed")
	}

	if err := types.AddOutput(tx, s.BTCKeeper.GetAnyoneCanSpendAddress(ctx).GetAddress(), s.BTCKeeper.GetMinOutputAmount(ctx)); err != nil {
		return nil, err
	}
	anyoneCanSpendVout := uint32(0)

	txSizeUpperBound, err := estimateTxSizeWithOutputsTo(ctx, s, *tx, secondaryAddress.GetAddress())
	if err != nil {
		return nil, err
	}

	outputsTotal := types.GetOutputsTotal(*tx)
	// rescue transactions always pay 1 satoshi/byte, which is the default minimum relay fee rate bitcoin-core sets
	fee := sdk.NewInt(txSizeUpperBound).MulRaw(types.MinRelayTxFeeSatoshiPerByte)
	change := inputsTotal.SubRaw(int64(outputsTotal)).Sub(fee)

	if !change.IsPositive() {
		return nil, fmt.Errorf("not enough inputs (%d) to cover the fee (%d) for the %s transaction", inputsTotal.Int64(), fee.Int64(), types.Rescue.SimpleString())
	}

	if err := types.AddOutput(tx, secondaryAddress.GetAddress(), btcutil.Amount(change.Int64())); err != nil {
		return nil, err
	}

	s.SetAddressInfo(ctx, secondaryAddress)

	tx = types.DisableTimelock(tx)
	unsignedTx := types.NewUnsignedTx(types.Rescue, tx, anyoneCanSpendVout, btcutil.Amount(change.Int64()))

	s.SetUnsignedTx(ctx, unsignedTx)

	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeConsolidationTx,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueCreated),
		sdk.NewAttribute(types.AttributeTxType, types.Rescue.SimpleString()),
	))

	s.Logger(ctx).Debug(fmt.Sprintf("successfully created %s transaction", types.Rescue.SimpleString()))

	return &types.CreateRescueTxResponse{}, nil
}

// CreateMasterTx creates a master key consolidation transaction
func (s msgServer) CreateMasterTx(c context.Context, req *types.CreateMasterTxRequest) (*types.CreateMasterTxResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	if err := validateChainActivated(ctx, s.nexus, exported.Bitcoin); err != nil {
		return nil, err
	}

	secondaryMin := s.GetMinOutputAmount(ctx)
	if req.SecondaryKeyAmount > 0 && req.SecondaryKeyAmount < secondaryMin {
		return nil, fmt.Errorf("cannot transfer %d to secondary key, it is below the minimum of %d", req.SecondaryKeyAmount, secondaryMin)
	}

	secondaryMax := s.GetMaxSecondaryOutputAmount(ctx)
	if req.SecondaryKeyAmount > secondaryMax {
		return nil, fmt.Errorf("cannot transfer %d to secondary key, it is above the maximum of %d", req.SecondaryKeyAmount, secondaryMax)
	}

	if _, ok := s.GetUnsignedTx(ctx, types.MasterConsolidation); ok {
		return nil, fmt.Errorf("consolidation in progress")
	}

	externalMultisigThreshold := s.signer.GetExternalMultisigThreshold(ctx)
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
		s.SetAddressInfo(ctx, secondaryAddress)
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

	if !change.IsPositive() {
		return nil, fmt.Errorf("not enough inputs (%d) to cover the fee (%d) for master consolidation transaction", inputsTotal.Int64(), fee.Int64())
	}

	if err := types.AddOutput(tx, consolidationAddress.GetAddress(), btcutil.Amount(change.Int64())); err != nil {
		return nil, err
	}

	s.SetAddressInfo(ctx, consolidationAddress)
	telemetry.SetGaugeWithLabels(
		[]string{types.ModuleName, "secondary", "address", "balance"},
		float32(change.Int64()),
		[]metrics.Label{
			telemetry.NewLabel("timestamp", strconv.FormatInt(ctx.BlockTime().Unix(), 10)),
			telemetry.NewLabel("address", consolidationAddress.Address),
		})

	tx = types.DisableTimelock(tx)
	unsignedTx := types.NewUnsignedTx(types.MasterConsolidation, tx, anyoneCanSpendVout, req.SecondaryKeyAmount)
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
			sdk.NewAttribute(types.AttributeKeyKeyID, string(consolidationKey.ID)),
		))
	}

	s.SetUnsignedTx(ctx, unsignedTx)

	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeConsolidationTx,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueCreated),
		sdk.NewAttribute(types.AttributeTxType, types.MasterConsolidation.SimpleString()),
	))

	return &types.CreateMasterTxResponse{}, nil
}

// CreatePendingTransfersTx creates a secondary key consolidation transaction
func (s msgServer) CreatePendingTransfersTx(c context.Context, req *types.CreatePendingTransfersTxRequest) (*types.CreatePendingTransfersTxResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	if err := validateChainActivated(ctx, s.nexus, exported.Bitcoin); err != nil {
		return nil, err
	}

	masterMin := s.GetMinOutputAmount(ctx)
	if req.MasterKeyAmount > 0 && req.MasterKeyAmount < masterMin {
		return nil, fmt.Errorf("cannot transfer %d to the master key, it is below the minimum amount of %d", req.MasterKeyAmount, masterMin)
	}

	if _, ok := s.GetUnsignedTx(ctx, types.SecondaryConsolidation); ok {
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
		s.SetAddressInfo(ctx, masterAddress)
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

	if !change.IsPositive() {
		return nil, fmt.Errorf("not enough deposits (%s) to make all withdrawals (%s) with a transaction fee of %s",
			inputsTotal.String(), outputsTotal.String(), btcutil.Amount(fee.Int64()).String(),
		)
	}

	if err := types.AddOutput(tx, consolidationAddress.GetAddress(), btcutil.Amount(change.Int64())); err != nil {
		return nil, err
	}

	s.SetAddressInfo(ctx, consolidationAddress)
	telemetry.SetGaugeWithLabels(
		[]string{types.ModuleName, "secondary", "address", "balance"},
		float32(change.Int64()),
		[]metrics.Label{
			telemetry.NewLabel("timestamp", strconv.FormatInt(ctx.BlockTime().Unix(), 10)),
			telemetry.NewLabel("address", consolidationAddress.Address),
		})

	tx = types.DisableTimelock(tx)
	unsignedTx := types.NewUnsignedTx(types.SecondaryConsolidation, tx, anyoneCanSpendVout, req.MasterKeyAmount)
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
			sdk.NewAttribute(types.AttributeKeyKeyID, string(consolidationKey.ID)),
		))
	}

	s.SetUnsignedTx(ctx, unsignedTx)

	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeConsolidationTx,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueCreated),
		sdk.NewAttribute(types.AttributeTxType, types.SecondaryConsolidation.SimpleString()),
	))

	return &types.CreatePendingTransfersTxResponse{}, nil
}

func getExternalKeys(ctx sdk.Context, k types.BTCKeeper, signer types.Signer) ([]tss.Key, error) {
	externalKeyIDs, ok := signer.GetExternalKeyIDs(ctx, exported.Bitcoin)
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

		addressInfo, ok := k.GetAddressInfo(ctx, outPointInfo.Address)
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

	// Combine output to same destination address
	for _, transfer := range pendingTransfers {
		recipient, err := btcutil.DecodeAddress(transfer.Recipient.Address, network)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("%s is not a valid address", transfer.Recipient.Address))
			continue
		}

		encodedAddress := recipient.EncodeAddress()
		amount := transfer.Asset.Amount

		// Check if the recipient has unsent dust amount
		if amount.LT(minAmount) {
			// Emit event and continue
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

func addInputs(ctx sdk.Context, k types.BTCKeeper, tx *wire.MsgTx, keyID tss.KeyID) (sdk.Int, error) {
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

func getSigID(sigHash []byte, keyID tss.KeyID) string {
	return fmt.Sprintf("%s-%s", hex.EncodeToString(sigHash), keyID)
}

func validateKeyAssignment(ctx sdk.Context, k types.BTCKeeper, signer types.Signer, snapshotter types.Snapshotter, currKey tss.Key, nextKey tss.Key) error {
	// Validate that the other key is not in the process of sending coin to the key that is to be rotated out
	var txType types.TxType

	switch currKey.Role {
	case tss.MasterKey:
		txType = types.SecondaryConsolidation
	case tss.SecondaryKey:
		txType = types.MasterConsolidation
	default:
		return fmt.Errorf("unknown key role %s", currKey.Role.SimpleString())
	}

	if unsignedTx, ok := k.GetUnsignedTx(ctx, txType); ok && unsignedTx.InternalTransferAmount > 0 {
		return fmt.Errorf("cannot assign the next %s key while a %s transaction is sending coin to the current %s address",
			currKey.Role.SimpleString(),
			txType.SimpleString(),
			currKey.Role.SimpleString(),
		)
	}

	// If rotating secondary key, validate that the secondary key whose validators are about to be allowed to unbond does not have any unspent outpoint
	if currKey.Role == tss.SecondaryKey {
		rotationCount := signer.GetRotationCount(ctx, exported.Bitcoin, currKey.Role)
		unbondingLockingKeyRotationCount := signer.GetKeyUnbondingLockingKeyRotationCount(ctx)
		if rotationCount > unbondingLockingKeyRotationCount {
			key, ok := signer.GetKeyByRotationCount(ctx, exported.Bitcoin, currKey.Role, rotationCount-unbondingLockingKeyRotationCount)
			if !ok {
				return fmt.Errorf("cannot find the %s key of rotation count %d", currKey.Role, rotationCount-unbondingLockingKeyRotationCount)
			}

			if !k.GetConfirmedOutpointInfoQueueForKey(ctx, key.ID).IsEmpty() {
				return fmt.Errorf("the %s key %s still has confirmed outpoints to spend, and resuce is required before key rotation is allowed", key.Role, key.ID)
			}

			if k.GetUnconfirmedAmount(ctx, key.ID) > 0 {
				return fmt.Errorf("the %s key %s still has unconfirmed outpoints to confirm, and confirm and resuce are required before key rotation is allowed", key.Role, key.ID)
			}
		}
	}

	if err := signer.AssertMatchesRequirements(ctx, snapshotter, exported.Bitcoin, nextKey.ID, currKey.Role); err != nil {
		return sdkerrors.Wrapf(err, "key %s does not match requirements for role %s", nextKey.ID, currKey.Role.SimpleString())
	}

	// TODO: How do we prevent the queue being always non-empty?
	if !k.GetConfirmedOutpointInfoQueueForKey(ctx, currKey.ID).IsEmpty() {
		return fmt.Errorf("the %s key %s still has confirmed outpoints to spend, and spend is required before key rotation is allowed", currKey.Role, currKey.ID)
	}

	if k.GetUnconfirmedAmount(ctx, currKey.ID) > 0 {
		return fmt.Errorf("the %s key %s still has unconfirmed outpoints to confirm, and confirm and spend is required before key rotation is allowed", currKey.Role, currKey.ID)
	}

	return nil
}

func getDepositAddress(ctx sdk.Context, k types.BTCKeeper, s types.Signer, key tss.Key, recipient nexus.CrossChainAddress) (types.AddressInfo, error) {
	if key.Role != tss.SecondaryKey {
		return types.AddressInfo{}, fmt.Errorf("given key %s is not for a %s key", key.ID, tss.SecondaryKey.SimpleString())
	}

	externalMultisigThreshold := s.GetExternalMultisigThreshold(ctx)
	externalKeys, err := getExternalKeys(ctx, k, s)
	if err != nil {
		return types.AddressInfo{}, err
	}
	if len(externalKeys) != int(externalMultisigThreshold.Denominator) {
		return types.AddressInfo{}, fmt.Errorf("number of external keys does not match the threshold and re-register is needed")
	}

	if key.RotatedAt == nil {
		return types.AddressInfo{}, fmt.Errorf("cannot get deposit address of key %s which is not rotated yet", key.ID)
	}

	nonce := utils.GetNonce(ctx.HeaderHash(), ctx.BlockGasMeter())
	externalKeyLockTime := key.RotatedAt.Add(k.GetMasterAddressExternalKeyLockDuration(ctx))
	scriptNonce := btcutil.Hash160([]byte(recipient.String() + hex.EncodeToString(nonce[:])))

	return types.NewDepositAddress(
		key,
		externalMultisigThreshold.Numerator,
		externalKeys,
		externalKeyLockTime,
		scriptNonce,
		k.GetNetwork(ctx),
	)
}

func getSecondaryConsolidationAddress(ctx sdk.Context, k types.BTCKeeper, key tss.Key) (types.AddressInfo, error) {
	if key.Role != tss.SecondaryKey {
		return types.AddressInfo{}, fmt.Errorf("given key %s is not for a %s key", key.ID, tss.SecondaryKey.SimpleString())
	}

	return types.NewSecondaryConsolidationAddress(
		key,
		k.GetNetwork(ctx),
	)
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

	externalMultisigThreshold := s.GetExternalMultisigThreshold(ctx)
	externalKeys, err := getExternalKeys(ctx, k, s)
	if err != nil {
		return types.AddressInfo{}, err
	}
	if len(externalKeys) != int(externalMultisigThreshold.Denominator) {
		return types.AddressInfo{}, fmt.Errorf("number of external keys does not match the threshold and re-register is needed")
	}

	internalKeyLockTime := currMasterKey.RotatedAt.Add(k.GetMasterAddressInternalKeyLockDuration(ctx))
	externalKeyLockTime := currMasterKey.RotatedAt.Add(k.GetMasterAddressExternalKeyLockDuration(ctx))

	return types.NewMasterConsolidationAddress(
		key,
		oldMasterKey,
		externalMultisigThreshold.Numerator,
		externalKeys,
		internalKeyLockTime,
		externalKeyLockTime,
		k.GetNetwork(ctx),
	)
}

func getOldMasterKey(ctx sdk.Context, k types.BTCKeeper, signer types.Signer) (tss.Key, bool) {
	currRotationCount := signer.GetRotationCount(ctx, exported.Bitcoin, tss.MasterKey)
	oldMasterKeyRotationCount := currRotationCount - k.GetMasterKeyRetentionPeriod(ctx)
	if oldMasterKeyRotationCount < 1 {
		oldMasterKeyRotationCount = 1
	}

	return signer.GetKeyByRotationCount(ctx, exported.Bitcoin, tss.MasterKey, oldMasterKeyRotationCount)
}

func validateChainActivated(ctx sdk.Context, n types.Nexus, chain nexus.Chain) error {
	if !n.IsChainActivated(ctx, chain) {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest,
			fmt.Sprintf("chain %s is not activated yet", chain.Name))
	}

	return nil
}
