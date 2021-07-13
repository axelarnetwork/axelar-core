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
	QDepositAddress          = "depositAddr"
	QMasterAddress           = "masterAddr"
	QKeyConsolidationAddress = "keyConsolidationAddress"
	QNextMasterKeyID         = "nextMasterKeyID"
	QMinimumWithdrawAmount   = "minWithdrawAmount"
	QTxState                 = "txState"
	QConsolidationTx         = "getConsolidationTx"
	QConsolidationTxState    = "QConsolidationTxState"
	QPayForConsolidationTx   = "getPayForConsolidationTx"
)

// NewQuerier returns a new querier for the Bitcoin module
func NewQuerier(rpc types.RPCClient, k types.BTCKeeper, s types.Signer, n types.Nexus) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		var res []byte
		var err error
		switch path[0] {
		case QueryDepositAddress:
			res, err = queryDepositAddress(ctx, k, s, n, req.Data)
		case QueryMasterAddress:
			res, err = queryMasterAddress(ctx, k, s)
		case QueryKeyConsolidationAddress:
			res, err = queryKeyConsolidationAddress(ctx, k, s, req.Data)
		case QNextMasterKeyID:
			res, err = queryNextMasterKeyID(ctx, s)
		case QMinimumWithdrawAmount:
			res = queryMinimumWithdrawAmount(ctx, k)
		case QueryTxState:
			res, err = queryTxState(ctx, k, req.Data)
		case GetConsolidationTx:
			res, err = getRawConsolidationTx(ctx, k)
		case GetConsolidationTxState:
			res, err = getConsolidationTxState(ctx, k)
		case GetPayForConsolidationTx:
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

func queryDepositAddress(ctx sdk.Context, k types.BTCKeeper, s types.Signer, n types.Nexus, data []byte) ([]byte, error) {
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

	depositAddr := types.NewLinkedAddress(masterKey, secondaryKey, k.GetNetwork(ctx), recipient)

	_, ok = n.GetRecipient(ctx, depositAddr.ToCrossChainAddr())
	if !ok {
		return nil, fmt.Errorf("deposit address is not linked with recipient address")
	}

	return []byte(depositAddr.Address), nil
}

func queryMasterAddress(ctx sdk.Context, k types.BTCKeeper, s types.Signer) ([]byte, error) {
	masterKey, ok := s.GetCurrentKey(ctx, exported.Bitcoin, tss.MasterKey)
	if !ok {
		return nil, fmt.Errorf("masterKey not found")
	}

	addr := types.NewConsolidationAddress(masterKey, k.GetNetwork(ctx))

	// After a key rotation, the master address is not known to Axelar Core until a consolidation
	// transaction is completed, during which k.SetAddress() is called.
	if _, ok := k.GetAddress(ctx, addr.Address); !ok {
		return nil, fmt.Errorf("no address found for current %s key %s", tss.MasterKey.SimpleString(), masterKey.ID)
	}

	resp := types.QueryMasterAddressResponse{
		Address: addr.Address,
		KeyId:   masterKey.ID,
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

	addr := types.NewConsolidationAddress(key, k.GetNetwork(ctx))
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

func queryTxState(ctx sdk.Context, k types.BTCKeeper, data []byte) ([]byte, error) {

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

func getConsolidationTxState(ctx sdk.Context, k types.BTCKeeper) ([]byte, error) {

	tx, ok := k.GetSignedTx(ctx)
	if !ok {
		return nil, fmt.Errorf("could not find the signed consolidation transaction")
	}
	txID := tx.TxHash()
	vout, ok := k.GetMasterKeyVout(ctx)
	if !ok {
		return nil, fmt.Errorf("could not find the consolidation transaction vout")
	}

	outpointByte := []byte(wire.NewOutPoint(&txID, vout).String())

	stateMsg, err := queryTxState(ctx, k, outpointByte)
	if err != nil {
		return nil, err
	}

	return stateMsg, nil
}

func getRawConsolidationTx(ctx sdk.Context, k types.BTCKeeper) ([]byte, error) {
	if _, ok := k.GetUnsignedTx(ctx); ok {
		rawTxResponse := &types.QueryRawTxResponse{StateOrTx: &types.QueryRawTxResponse_State{State: types.Signing}}
		return rawTxResponse.Marshal()
	}

	if tx, ok := k.GetSignedTx(ctx); ok {
		rawTxResponse := &types.QueryRawTxResponse{StateOrTx: &types.QueryRawTxResponse_RawTx{RawTx: hex.EncodeToString(types.MustEncodeTx(tx))}}
		return rawTxResponse.Marshal()
	}

	rawTxResponse := &types.QueryRawTxResponse{StateOrTx: &types.QueryRawTxResponse_State{State: types.Ready}}
	return rawTxResponse.Marshal()
}

func payForConsolidationTx(ctx sdk.Context, k types.BTCKeeper, rpc types.RPCClient, data []byte) ([]byte, error) {
	feeRate := int64(binary.LittleEndian.Uint64(data))

	consolidationTx, ok := k.GetSignedTx(ctx)
	if !ok {
		return nil, fmt.Errorf("no signed consolidation transaction ready")
	}

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
	// By setting an input's sequence to be (wire.MaxTxInSequenceNum - 2), it makes the transaction opt-in to transaction replacement (https://github.com/bitcoin/bips/blob/master/bip-0125.mediawiki)
	tx.TxIn[0].Sequence = wire.MaxTxInSequenceNum - 2
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
