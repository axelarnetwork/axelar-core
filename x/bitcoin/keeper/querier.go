package keeper

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// Query paths
const (
	QDepositAddress                = "depositAddr"
	QConsolidationAddressByKeyRole = "consolidationAddrByKeyRole"
	QConsolidationAddressByKeyID   = "consolidationAddrByKeyID"
	QNextKeyID                     = "nextKeyID"
	QMinOutputAmount               = "minOutputAmount"
	QLatestTxByTxType              = "latestTxByKeyRole"
	QSignedTx                      = "signedTx"
	QDepositStatus                 = "depositStatus"
)

// NewQuerier returns a new querier for the Bitcoin module
func NewQuerier(k types.BTCKeeper, s types.Signer, n types.Nexus) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		var res []byte
		var err error
		switch path[0] {
		case QDepositAddress:
			res, err = QueryDepositAddress(ctx, k, s, n, req.Data)
		case QDepositStatus:
			res, err = QueryDepositStatus(ctx, k, path[1])
		case QConsolidationAddressByKeyRole:
			res, err = QueryConsolidationAddressByKeyRole(ctx, k, s, path[1])
		case QConsolidationAddressByKeyID:
			keyID := tss.KeyID(path[1])
			err = keyID.Validate()
			if err != nil {
				break
			}
			res, err = QueryConsolidationAddressByKeyID(ctx, k, s, keyID)
		case QNextKeyID:
			res, err = QueryNextKeyID(ctx, s, path[1])
		case QMinOutputAmount:
			res = QueryMinOutputAmount(ctx, k)
		case QLatestTxByTxType:
			res, err = QueryLatestTxByTxType(ctx, k, path[1])
		case QSignedTx:
			res, err = QuerySignedTx(ctx, k, path[1])
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, fmt.Sprintf("unknown query endpoint: %s", path[0]))
		}

		if err != nil {
			return nil, sdkerrors.Wrap(types.ErrBitcoin, err.Error())
		}

		return res, nil
	}
}

// QueryDepositStatus returns the status of the queried depoist
func QueryDepositStatus(ctx sdk.Context, k types.BTCKeeper, outpointStr string) ([]byte, error) {
	outpoint, err := types.OutPointFromStr(outpointStr)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "cannot parse outpoint")
	}

	key := vote.NewPollKey(types.ModuleName, outpointStr)

	var resp types.QueryDepositStatusResponse

	_, pending := k.GetPendingOutPointInfo(ctx, key)
	_, state, ok := k.GetOutPointInfo(ctx, *outpoint)

	switch {
	case pending:
		resp = types.QueryDepositStatusResponse{Status: types.OutPointState_Pending, Log: "deposit is waiting for confirmation"}
	case !pending && !ok:
		resp = types.QueryDepositStatusResponse{Status: types.OutPointState_None, Log: "deposit is unknown"}
	case state == types.OutPointState_Confirmed:
		resp = types.QueryDepositStatusResponse{Status: types.OutPointState_Confirmed, Log: "deposit has been confirmed and is pending for transfer"}
	case state == types.OutPointState_Spent:
		resp = types.QueryDepositStatusResponse{Status: types.OutPointState_Spent, Log: "deposit has been transferred to the destination address"}
	default:
		return nil, fmt.Errorf("deposit is in an unexpected state")
	}

	return types.ModuleCdc.MarshalLengthPrefixed(&resp)
}

// QueryDepositAddress returns deposit address
func QueryDepositAddress(ctx sdk.Context, k types.BTCKeeper, s types.Signer, n types.Nexus, data []byte) ([]byte, error) {
	var params types.DepositQueryParams
	if err := types.ModuleCdc.UnmarshalLengthPrefixed(data, &params); err != nil {
		return nil, fmt.Errorf("could not parse the recipient")
	}

	chain, ok := n.GetChain(ctx, params.Chain)
	if !ok {
		return nil, fmt.Errorf("recipient chain not found")
	}

	secondaryKey, ok := s.GetCurrentKey(ctx, exported.Bitcoin, tss.SecondaryKey)
	if !ok {
		return nil, fmt.Errorf("secondary key not set")
	}

	recipient := nexus.CrossChainAddress{Chain: chain, Address: params.Address}
	depositAddr, err := getDepositAddress(ctx, k, s, secondaryKey, recipient)
	if err != nil {
		return nil, err
	}

	_, ok = n.GetRecipient(ctx, depositAddr.ToCrossChainAddr())
	if !ok {
		return nil, fmt.Errorf("deposit address is not linked with recipient address")
	}

	resp := types.QueryAddressResponse{
		Address: depositAddr.Address,
		KeyID:   depositAddr.KeyID,
	}

	return types.ModuleCdc.MarshalLengthPrefixed(&resp)
}

// QueryConsolidationAddressByKeyRole returns the current consolidation address of the given key role
func QueryConsolidationAddressByKeyRole(ctx sdk.Context, k types.BTCKeeper, s types.Signer, keyRoleStr string) ([]byte, error) {
	keyRole, err := tss.KeyRoleFromSimpleStr(keyRoleStr)
	if err != nil {
		return nil, err
	}

	keyID, ok := s.GetCurrentKeyID(ctx, exported.Bitcoin, keyRole)
	if !ok {
		return nil, fmt.Errorf("%s key not found", keyRoleStr)
	}

	return QueryConsolidationAddressByKeyID(ctx, k, s, keyID)
}

// QueryConsolidationAddressByKeyID returns the consolidation address of the given key ID
func QueryConsolidationAddressByKeyID(ctx sdk.Context, k types.BTCKeeper, s types.Signer, keyID tss.KeyID) ([]byte, error) {

	key, ok := s.GetKey(ctx, keyID)
	if !ok {
		return nil, fmt.Errorf("no key with keyID %s found", keyID)
	}

	var addressInfo types.AddressInfo
	var err error

	switch key.Role {
	case tss.MasterKey:
		addressInfo, err = getMasterConsolidationAddress(ctx, k, s, key)
	case tss.SecondaryKey:
		addressInfo, err = getSecondaryConsolidationAddress(ctx, k, key)
	default:
		return nil, fmt.Errorf("no consolidation address supported for key %s of key role %s", keyID, key.Role.SimpleString())
	}

	if err != nil {
		return nil, err
	}

	resp := types.QueryAddressResponse{Address: addressInfo.Address, KeyID: addressInfo.KeyID}

	return types.ModuleCdc.MarshalLengthPrefixed(&resp)
}

// QueryNextKeyID returns the next key ID of the given key role
func QueryNextKeyID(ctx sdk.Context, s types.Signer, keyRoleStr string) ([]byte, error) {
	keyRole, err := tss.KeyRoleFromSimpleStr(keyRoleStr)
	if err != nil {
		return nil, err
	}

	next, nextAssigned := s.GetNextKey(ctx, exported.Bitcoin, keyRole)
	if !nextAssigned {
		return []byte{}, nil
	}

	return []byte(next.ID), nil
}

// QueryMinOutputAmount returns the minimum amount allowed for any transaction output
func QueryMinOutputAmount(ctx sdk.Context, k types.BTCKeeper) []byte {
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, uint64(k.GetMinOutputAmount(ctx)))

	return bz
}

// QueryLatestTxByTxType returns the latest consolidation transaction of the given tx type
func QueryLatestTxByTxType(ctx sdk.Context, k types.BTCKeeper, txTypeStr string) ([]byte, error) {
	txType, err := types.TxTypeFromSimpleStr(txTypeStr)
	if err != nil {
		return nil, err
	}

	unsignedTx, ok := k.GetUnsignedTx(ctx, txType)
	if ok {
		prevSignedTxHashHex := ""

		prevSignedTxHash, ok := k.GetLatestSignedTxHash(ctx, txType)
		if ok {
			prevSignedTxHashHex = prevSignedTxHash.String()
		}

		var signingInfos []*types.QueryTxResponse_SigningInfo

		for _, txIn := range unsignedTx.GetTx().TxIn {
			outPointStr := txIn.PreviousOutPoint.String()
			outPointInfo, state, ok := k.GetOutPointInfo(ctx, txIn.PreviousOutPoint)
			if !ok || state != types.OutPointState_Spent {
				return nil, fmt.Errorf("out point info %s is not found or not spent", outPointStr)
			}

			addressInfo, ok := k.GetAddress(ctx, outPointInfo.Address)
			if !ok {
				return nil, fmt.Errorf("unknown outpoint address %s", outPointInfo.Address)
			}

			signingInfos = append(signingInfos, &types.QueryTxResponse_SigningInfo{
				RedeemScript: hex.EncodeToString(addressInfo.RedeemScript),
				Amount:       int64(outPointInfo.Amount),
			})
		}

		resp := types.QueryTxResponse{
			Tx:                   hex.EncodeToString(types.MustEncodeTx(unsignedTx.GetTx())),
			Status:               unsignedTx.Status,
			ConfirmationRequired: unsignedTx.ConfirmationRequired,
			PrevSignedTxHash:     prevSignedTxHashHex,
			AnyoneCanSpendVout:   unsignedTx.AnyoneCanSpendVout,
			SigningInfos:         signingInfos,
		}

		return types.ModuleCdc.MarshalLengthPrefixed(&resp)
	}

	latestSignedTxHash, ok := k.GetLatestSignedTxHash(ctx, txType)
	if !ok {
		return nil, fmt.Errorf("no %s transaction exists", txType.SimpleString())
	}

	signedTx, ok := k.GetSignedTx(ctx, *latestSignedTxHash)
	if !ok {
		return nil, fmt.Errorf("cannot find the latest signed %s transaction", txType.SimpleString())
	}

	prevSignedTxHashHex := ""
	if signedTx.PrevSignedTxHash != nil {
		prevSignedTxHash, err := chainhash.NewHash(signedTx.PrevSignedTxHash)
		if err != nil {
			return nil, err
		}

		prevSignedTxHashHex = prevSignedTxHash.String()
	}

	resp := types.QueryTxResponse{
		Tx:                   hex.EncodeToString(types.MustEncodeTx(signedTx.GetTx())),
		Status:               types.Signed,
		ConfirmationRequired: signedTx.ConfirmationRequired,
		PrevSignedTxHash:     prevSignedTxHashHex,
		AnyoneCanSpendVout:   signedTx.AnyoneCanSpendVout,
		SigningInfos:         nil,
	}

	return types.ModuleCdc.MarshalLengthPrefixed(&resp)
}

// QuerySignedTx returns the signed consolidation transaction of the given transaction hash
func QuerySignedTx(ctx sdk.Context, k types.BTCKeeper, txHashHex string) ([]byte, error) {
	txHash, err := chainhash.NewHashFromStr(txHashHex)
	if err != nil {
		return nil, err
	}

	signedTx, ok := k.GetSignedTx(ctx, *txHash)
	if !ok {
		return nil, fmt.Errorf("cannot find signed consolidation transaction for the given transaction hash %s", txHash.String())
	}

	prevSignedTxHashHex := ""
	if signedTx.PrevSignedTxHash != nil {
		prevSignedTxHash, err := chainhash.NewHash(signedTx.PrevSignedTxHash)
		if err != nil {
			return nil, err
		}

		prevSignedTxHashHex = prevSignedTxHash.String()
	}

	resp := types.QueryTxResponse{
		Tx:                   hex.EncodeToString(types.MustEncodeTx(signedTx.GetTx())),
		Status:               types.Signed,
		ConfirmationRequired: signedTx.ConfirmationRequired,
		PrevSignedTxHash:     prevSignedTxHashHex,
		AnyoneCanSpendVout:   signedTx.AnyoneCanSpendVout,
		SigningInfos:         nil,
	}

	return types.ModuleCdc.MarshalLengthPrefixed(&resp)
}
