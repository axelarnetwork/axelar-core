package keeper

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/mempool"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
)

// Query paths
const (
	QDepositAddress                = "depositAddr"
	QSecondaryConsolidationAddress = "masterAddr"
	QKeyConsolidationAddress       = "keyConsolidationAddress"
	QNextMasterKeyID               = "nextMasterKeyID"
	QMinimumWithdrawAmount         = "minWithdrawAmount"
	QTxState                       = "txState"
	QConsolidationTx               = "getConsolidationTx"
	QConsolidationTxState          = "QConsolidationTxState"
	QPayForConsolidationTx         = "getPayForConsolidationTx"
)

// NewQuerier returns a new querier for the Bitcoin module
func NewQuerier(rpc types.RPCClient, k types.BTCKeeper, s types.Signer, n types.Nexus) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		var res []byte
		var err error
		switch path[0] {
		case QDepositAddress:
			res, err = QueryDepositAddress(ctx, k, s, n, req.Data)
		case QSecondaryConsolidationAddress:
			res, err = QuerySecondaryConsolidationAddress(ctx, k, s)
		case QKeyConsolidationAddress:
			res, err = queryKeyConsolidationAddress(ctx, k, s, req.Data)
		case QNextMasterKeyID:
			res, err = queryNextMasterKeyID(ctx, s)
		case QMinimumWithdrawAmount:
			res = queryMinimumWithdrawAmount(ctx, k)
		case QTxState:
			res, err = QueryTxState(ctx, k, req.Data)
		case QConsolidationTx:
			res, err = GetRawConsolidationTx(ctx, k)
		case QConsolidationTxState:
			res, err = GetConsolidationTxState(ctx, k)
		case QPayForConsolidationTx:
			res, err = payForConsolidationTx(ctx, k, rpc, req.Data)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, fmt.Sprintf("unknown btc-bridge query endpoint: %s", path[1]))
		}

		if err != nil {
			return nil, sdkerrors.Wrap(types.ErrBitcoin, err.Error())
		}
		return res, nil
	}
}

// QueryDepositAddress returns deposit address
func QueryDepositAddress(ctx sdk.Context, k types.BTCKeeper, s types.Signer, n types.Nexus, data []byte) ([]byte, error) {
	var params types.DepositQueryParams
	if err := types.ModuleCdc.UnmarshalJSON(data, &params); err != nil {
		return nil, fmt.Errorf("could not parse the recipient")
	}

	chain, ok := n.GetChain(ctx, params.Chain)
	if !ok {
		return nil, fmt.Errorf("recipient chain not found")
	}

	recipient := nexus.CrossChainAddress{Chain: chain, Address: params.Address}

	masterKey, ok := s.GetCurrentKey(ctx, exported.Bitcoin, tss.MasterKey)
	if !ok {
		return nil, fmt.Errorf("master key not set")
	}

	secondaryKey, ok := s.GetCurrentKey(ctx, exported.Bitcoin, tss.SecondaryKey)
	if !ok {
		return nil, fmt.Errorf("secondary key not set")
	}

	depositAddr := types.NewDepositAddress(masterKey, secondaryKey, k.GetNetwork(ctx), recipient)

	_, ok = n.GetRecipient(ctx, depositAddr.ToCrossChainAddr())
	if !ok {
		return nil, fmt.Errorf("deposit address is not linked with recipient address")
	}

	return []byte(depositAddr.Address), nil
}

// QuerySecondaryConsolidationAddress returns the master address
func QuerySecondaryConsolidationAddress(ctx sdk.Context, k types.BTCKeeper, s types.Signer) ([]byte, error) {
	secondaryKey, ok := s.GetCurrentKey(ctx, exported.Bitcoin, tss.SecondaryKey)
	if !ok {
		return nil, fmt.Errorf("secondaryKey not found")
	}

	resp := types.QuerySecondaryConsolidationAddressResponse{
		Address: types.NewSecondaryConsolidationAddress(secondaryKey, k.GetNetwork(ctx)).Address,
		KeyId:   secondaryKey.ID,
	}

	return resp.Marshal()
}

func queryKeyConsolidationAddress(ctx sdk.Context, k types.BTCKeeper, s types.Signer, keyIDBytes []byte) ([]byte, error) {
	keyID := string(keyIDBytes)

	key, ok := s.GetKey(ctx, keyID)
	if !ok {
		return nil, fmt.Errorf("no key with keyID %s found", keyID)
	}
	if key.Role != tss.MasterKey {
		return nil, fmt.Errorf("key %s does not have the role %s", keyID, tss.MasterKey)
	}

	addr := types.NewSecondaryConsolidationAddress(key, k.GetNetwork(ctx))
	return []byte(addr.Address), nil
}

func queryNextMasterKeyID(ctx sdk.Context, s types.Signer) ([]byte, error) {
	next, nextAssigned := s.GetNextKey(ctx, exported.Bitcoin, tss.MasterKey)
	if !nextAssigned {
		return []byte{}, nil
	}

	return []byte(next.ID), nil
}

func queryMinimumWithdrawAmount(ctx sdk.Context, k types.BTCKeeper) []byte {
	amount := make([]byte, 8)
	binary.LittleEndian.PutUint64(amount, uint64(k.GetMinimumWithdrawalAmount(ctx)))
	return amount
}

// QueryTxState returns the state of given transaction
func QueryTxState(ctx sdk.Context, k types.BTCKeeper, data []byte) ([]byte, error) {
	outpoint, err := types.OutPointFromStr(string(data))
	if err != nil {
		return nil, err
	}

	_, state, ok := k.GetOutPointInfo(ctx, *outpoint)
	var message string

	switch {
	case !ok:
		return nil, fmt.Errorf("bitcoin transaction is not tracked")
	case state == types.CONFIRMED:
		message = "bitcoin transaction state is confirmed"
	case state == types.SPENT:
		message = "bitcoin transaction state is spent"
	default:
		message = "bitcoin transaction state is not confirmed"
	}

	return []byte(message), nil
}

// GetConsolidationTxState returns the state of consolidqtion transaction
func GetConsolidationTxState(ctx sdk.Context, k types.BTCKeeper) ([]byte, error) {
	txHash, ok := k.GetLatestSignedTxHash(ctx)
	if !ok {
		return nil, fmt.Errorf("could not find the signed consolidation transaction")
	}

	outpointByte := []byte(wire.NewOutPoint(txHash, 0).String())

	stateMsg, err := QueryTxState(ctx, k, outpointByte)
	if err != nil {
		return nil, err
	}

	return stateMsg, nil
}

// GetRawConsolidationTx returns the consolidation transaction in bytes
func GetRawConsolidationTx(ctx sdk.Context, k types.BTCKeeper) ([]byte, error) {
	txHash, ok := k.GetLatestSignedTxHash(ctx)
	if !ok {
		rawTxResponse := &types.QueryRawTxResponse{StateOrTx: &types.QueryRawTxResponse_State{State: types.Ready}}
		return rawTxResponse.Marshal()
	}

	tx, _ := k.GetSignedTx(ctx, *txHash)

	rawTxResponse := &types.QueryRawTxResponse{StateOrTx: &types.QueryRawTxResponse_RawTx{RawTx: hex.EncodeToString(types.MustEncodeTx(tx))}}
	return rawTxResponse.Marshal()
}

func payForConsolidationTx(ctx sdk.Context, k types.BTCKeeper, rpc types.RPCClient, data []byte) ([]byte, error) {
	txHash, ok := k.GetLatestSignedTxHash(ctx)
	if !ok {
		return nil, fmt.Errorf("no signed consolidation transaction ready")
	}

	feeRate := int64(binary.LittleEndian.Uint64(data))
	consolidationTx, _ := k.GetSignedTx(ctx, *txHash)

	utxos, err := rpc.ListUnspent()
	if err != nil {
		return nil, err
	}

	if len(utxos) <= 0 {
		return nil, fmt.Errorf("no UTXO available to pay for consolidation transaction")
	}

	if feeRate <= 0 {
		estimateSmartFeeResult, err := rpc.EstimateSmartFee(1, &btcjson.EstimateModeEconomical)
		if err != nil {
			return nil, err
		}

		// feeRate here is measured in satoshi/byte
		feeRate = int64(math.Ceil(*estimateSmartFeeResult.FeeRate * btcutil.SatoshiPerBitcoin / 1000))
		if feeRate <= types.MinRelayTxFeeSatoshiPerByte {
			return nil, fmt.Errorf("no need to pay for consolidation transaction")
		}
	}

	network := k.GetNetwork(ctx)
	consolidationTxHash := consolidationTx.TxHash()
	anyoneCanSpendAddress := k.GetAnyoneCanSpendAddress(ctx)
	inputs := []types.OutPointToSign{
		{
			OutPointInfo: types.NewOutPointInfo(
				// TODO: anyone-can-spend output is no longer harded-coded at vout 1. Will fix soon.
				wire.NewOutPoint(&consolidationTxHash, 1),
				k.GetMinimumWithdrawalAmount(ctx),
				anyoneCanSpendAddress.Address,
			),
			AddressInfo: types.AddressInfo{
				Address:      anyoneCanSpendAddress.Address,
				RedeemScript: anyoneCanSpendAddress.RedeemScript,
			},
		},
	}
	inputTotal := sdk.NewInt(int64(k.GetMinimumWithdrawalAmount(ctx)))

	for _, utxo := range utxos {
		hash, err := chainhash.NewHashFromStr(utxo.TxID)
		if err != nil {
			return nil, err
		}

		address, err := btcutil.DecodeAddress(utxo.Address, network.Params())
		if err != nil {
			return nil, err
		}

		redeemScript, err := hex.DecodeString(utxo.RedeemScript)
		if err != nil {
			return nil, err
		}

		amount := btcutil.Amount(utxo.Amount * btcutil.SatoshiPerBitcoin)
		outPointInfo := types.NewOutPointInfo(
			wire.NewOutPoint(hash, utxo.Vout),
			amount,
			utxo.Address,
		)
		addressInfo := types.AddressInfo{
			Address:      address.EncodeAddress(),
			RedeemScript: redeemScript,
		}

		input := types.OutPointToSign{
			OutPointInfo: outPointInfo,
			AddressInfo:  addressInfo,
		}
		inputs = append(inputs, input)
		inputTotal = inputTotal.AddRaw(int64(amount))
	}

	address, err := btcutil.DecodeAddress(utxos[0].Address, network.Params())
	if err != nil {
		return nil, err
	}
	txSizeUpperBound, err := estimateTxSize(inputs, []types.Output{{Amount: 0, Recipient: address}})
	if err != nil {
		return nil, err
	}

	consolidationTxSize := mempool.GetTxVirtualSize(btcutil.NewTx(consolidationTx))
	fee := (txSizeUpperBound+consolidationTxSize)*feeRate - consolidationTxSize*types.MinRelayTxFeeSatoshiPerByte
	amount := btcutil.Amount(inputTotal.SubRaw(fee).Int64())
	if amount < 0 {
		return nil, fmt.Errorf("not enough UTXOs to execute child-pay-for-parent with fee rate %d", feeRate)
	}

	outputs := []types.Output{
		{Amount: btcutil.Amount(inputTotal.SubRaw(fee).Int64()), Recipient: address},
	}

	tx, err := types.CreateTx(inputs, outputs)
	if err != nil {
		return nil, err
	}

	tx.TxIn[0].Witness = wire.TxWitness{anyoneCanSpendAddress.RedeemScript}
	tx = types.EnableTimelockAndRBF(tx)

	tx, _, err = rpc.SignRawTransactionWithWallet(tx)
	if err != nil {
		return nil, err
	}

	return []byte(hex.EncodeToString(types.MustEncodeTx(tx))), nil
}

func estimateTxSize(inputs []types.OutPointToSign, outputs []types.Output) (int64, error) {
	tx, err := types.CreateTx(inputs, outputs)
	if err != nil {
		return 0, err
	}

	return types.EstimateTxSize(*tx, inputs), nil
}
