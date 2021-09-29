package rest

import (
	"encoding/hex"
	"net/http"

	"github.com/btcsuite/btcutil"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"

	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"

	clientUtils "github.com/axelarnetwork/axelar-core/utils"
)

// rest routes
const (
	TxLink                        = "link"
	TxConfirmTx                   = "confirm"
	TxCreatePendingTransfersTx    = "create-pending-transfers-tx"
	TxCreateMasterConsolidationTx = "create-master-consolidation-tx"
	TxCreateRescueTx              = "create-rescue-tx"
	TxSignTx                      = "sign-tx"
	TxSubmitExternalSignature     = "submit-external-signature"

	QueryDepositAddress       = "deposit-address"
	QueryDepositStatus        = "deposit-status"
	QueryConsolidationAddress = "consolidation-address"
	QueryMinOutputAmount      = "min-output-amount"
	QueryNextKeyID            = "next-key-id"
	QueryLatestTx             = "latest-tx"
	QuerySignedTx             = "signed-tx"
)

// RegisterRoutes registers this module's REST routes with the given router
func RegisterRoutes(cliCtx client.Context, r *mux.Router) {
	registerTx := clientUtils.RegisterTxHandlerFn(r, types.RestRoute)
	registerTx(TxHandlerLink(cliCtx), TxLink, clientUtils.PathVarChain)
	registerTx(TxHandlerConfirmTx(cliCtx), TxConfirmTx)
	registerTx(TxHandlerCreatePendingTransfersTx(cliCtx), TxCreatePendingTransfersTx)
	registerTx(TxHandlerCreateMasterConsolidationTx(cliCtx), TxCreateMasterConsolidationTx)
	registerTx(TxHandlerCreateRescueTx(cliCtx), TxCreateRescueTx)
	registerTx(TxHandlerSignTx(cliCtx), TxSignTx)
	registerTx(TxHandlerSubmitExternalSignature(cliCtx), TxSubmitExternalSignature)

	registerQuery := clientUtils.RegisterQueryHandlerFn(r, types.RestRoute)
	registerQuery(QueryHandlerDepositAddress(cliCtx), QueryDepositAddress, clientUtils.PathVarChain, clientUtils.PathVarEthereumAddress)
	registerQuery(QueryHandlerDepositStatus(cliCtx), QueryDepositStatus, clientUtils.PathVarOutpoint)
	registerQuery(QueryHandlerConsolidationAddress(cliCtx), QueryConsolidationAddress)
	registerQuery(QueryHandlerNextKeyID(cliCtx), QueryNextKeyID, clientUtils.PathVarKeyRole)
	registerQuery(QueryHandlerMinOutputAmount(cliCtx), QueryMinOutputAmount)
	registerQuery(QueryHandlerLatestTx(cliCtx), QueryLatestTx, clientUtils.PathVarTxType)
	registerQuery(QueryHandlerSignedTx(cliCtx), QuerySignedTx, clientUtils.PathVarTxID)
}

// ReqLink represents a request to link a cross-chain address to a Bitcoin address
type ReqLink struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	Address string       `json:"address" yaml:"address"`
}

// ReqConfirmOutPoint represents a request to confirm a Bitcoin outpoint
type ReqConfirmOutPoint struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	TxInfo  string       `json:"tx_info" yaml:"tx_info"`
}

// ReqCreatePendingTransfersTx represents a request to create a secondary key consolidation transaction handling all pending transfers
type ReqCreatePendingTransfersTx struct {
	BaseReq         rest.BaseReq `json:"base_req" yaml:"base_req"`
	KeyID           string       `json:"key_id" yaml:"key_id"`
	MasterKeyAmount string       `json:"master_key_amount" yaml:"master_key_amount"`
}

// ReqCreateMasterConsolidationTx represents a request to create a master key consolidation transaction
type ReqCreateMasterConsolidationTx struct {
	BaseReq            rest.BaseReq `json:"base_req" yaml:"base_req"`
	KeyID              string       `json:"key_id" yaml:"key_id"`
	SecondaryKeyAmount string       `json:"secondary_key_amount" yaml:"secondary_key_amount"`
}

// ReqCreateRescueTx represents a request to create a rescue transaction
type ReqCreateRescueTx struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
}

// ReqSignTx represents a request to sign a consolidation transaction
type ReqSignTx struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	TxType  string       `json:"tx_type" yaml:"tx_type"`
}

// ReqSubmitExternalSignature represents a request to submit a signature from an external key
type ReqSubmitExternalSignature struct {
	BaseReq   rest.BaseReq `json:"base_req" yaml:"base_req"`
	KeyID     string       `json:"key_id" yaml:"key_id"`
	Signature string       `json:"signature" yaml:"signature"`
	SigHash   string       `json:"sig_hash" yaml:"sig_hash"`
}

// TxHandlerLink returns the handler to link a Bitcoin address to a cross-chain address
func TxHandlerLink(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqLink
		if !rest.ReadRESTReq(w, r, cliCtx.LegacyAmino, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		fromAddr, ok := clientUtils.ExtractReqSender(w, req.BaseReq)
		if !ok {
			return
		}

		msg := types.NewLinkRequest(fromAddr, req.Address, mux.Vars(r)[clientUtils.PathVarChain])
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

// TxHandlerConfirmTx returns the handler to confirm a tx outpoint
func TxHandlerConfirmTx(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqConfirmOutPoint
		if !rest.ReadRESTReq(w, r, cliCtx.LegacyAmino, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		fromAddr, ok := clientUtils.ExtractReqSender(w, req.BaseReq)
		if !ok {
			return
		}

		var out types.OutPointInfo
		if err := cliCtx.LegacyAmino.UnmarshalJSON([]byte(req.TxInfo), &out); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		msg := types.NewConfirmOutpointRequest(fromAddr, out)

		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

// TxHandlerCreatePendingTransfersTx returns the handler to create a secondary key consolidation transaction handling all pending transfers
func TxHandlerCreatePendingTransfersTx(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqCreatePendingTransfersTx
		if !rest.ReadRESTReq(w, r, cliCtx.LegacyAmino, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		fromAddr, ok := clientUtils.ExtractReqSender(w, req.BaseReq)
		if !ok {
			return
		}

		masterKeyAmount, err := types.ParseSatoshi(req.MasterKeyAmount)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		msg := types.NewCreatePendingTransfersTxRequest(fromAddr, req.KeyID, btcutil.Amount(masterKeyAmount.Amount.Int64()))
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

// TxHandlerCreateMasterConsolidationTx returns the handler to create a master key consolidation transaction
func TxHandlerCreateMasterConsolidationTx(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqCreateMasterConsolidationTx
		if !rest.ReadRESTReq(w, r, cliCtx.LegacyAmino, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		fromAddr, ok := clientUtils.ExtractReqSender(w, req.BaseReq)
		if !ok {
			return
		}

		secondaryKeyAmount, err := types.ParseSatoshi(req.SecondaryKeyAmount)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		msg := types.NewCreateMasterTxRequest(fromAddr, req.KeyID, btcutil.Amount(secondaryKeyAmount.Amount.Int64()))
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

// TxHandlerCreateRescueTx returns the handler to create a rescue transaction
func TxHandlerCreateRescueTx(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqCreateRescueTx
		if !rest.ReadRESTReq(w, r, cliCtx.LegacyAmino, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		fromAddr, ok := clientUtils.ExtractReqSender(w, req.BaseReq)
		if !ok {
			return
		}

		msg := types.NewCreateRescueTxRequest(fromAddr)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

// TxHandlerSignTx returns the handler to sign a consolidation transaction
func TxHandlerSignTx(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqSignTx
		if !rest.ReadRESTReq(w, r, cliCtx.LegacyAmino, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		fromAddr, ok := clientUtils.ExtractReqSender(w, req.BaseReq)
		if !ok {
			return
		}

		txType, err := types.TxTypeFromSimpleStr(req.TxType)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		msg := types.NewSignTxRequest(fromAddr, txType)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

// TxHandlerSubmitExternalSignature returns the handler to submit a signature from an external key
func TxHandlerSubmitExternalSignature(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqSubmitExternalSignature
		if !rest.ReadRESTReq(w, r, cliCtx.LegacyAmino, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		fromAddr, ok := clientUtils.ExtractReqSender(w, req.BaseReq)
		if !ok {
			return
		}

		signature, err := hex.DecodeString(req.Signature)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		sigHash, err := hex.DecodeString(req.SigHash)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		msg := types.NewSubmitExternalSignatureRequest(fromAddr, req.KeyID, signature, sigHash)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}
