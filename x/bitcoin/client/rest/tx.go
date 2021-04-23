package rest

import (
	"net/http"

	"github.com/btcsuite/btcutil"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"

	clientUtils "github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
)

// rest routes
const (
	TxMethodLink                   = "link"
	TxMethodConfirmTx              = "confirm"
	TxMethodSignPendingTransfersTx = "sign"

	QueryMethodDepositAddress     = keeper.QueryDepositAddress
	QueryMethodKeyAddress         = keeper.QueryKeyAddress
	QueryMethodGetConsolidationTx = keeper.GetTx
)

// RegisterRoutes registers this module's REST routes with the given router
func RegisterRoutes(cliCtx client.Context, r *mux.Router) {
	registerTx := clientUtils.RegisterTxHandlerFn(r, types.RestRoute)
	registerTx(GetHandlerLink(cliCtx), TxMethodLink, clientUtils.PathVarChain)
	registerTx(GetHandlerConfirmTx(cliCtx), TxMethodConfirmTx)
	registerTx(GetHandlerSignPendingTransfersTx(cliCtx), TxMethodSignPendingTransfersTx)

	registerQuery := clientUtils.RegisterQueryHandlerFn(r, types.RestRoute)
	registerQuery(QueryDepositAddress(cliCtx), QueryMethodDepositAddress, clientUtils.PathVarChain, clientUtils.PathVarEthereumAddress)
	registerQuery(QueryKeyAddress(cliCtx), QueryMethodKeyAddress, clientUtils.PathVarKeyRole)
	registerQuery(QueryGetConsolidationTx(cliCtx), QueryMethodGetConsolidationTx)
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

// ReqSignPendingTransfersTx represents a request to sign pending token transfers from Ethereum
type ReqSignPendingTransfersTx struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	Fee     string       `json:"fee" yaml:"fee"`
}

// GetHandlerSignPendingTransfersTx returns the handler to sign pending transfers to Bitcoin
func GetHandlerSignPendingTransfersTx(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqSignPendingTransfersTx
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

		satoshi, err := types.ParseSatoshi(req.Fee)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		msg := types.NewMsgSignPendingTransfers(fromAddr, btcutil.Amount(satoshi.Amount.Int64()))
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

// GetHandlerLink returns the handler to link a Bitcoin address to a cross-chain address
func GetHandlerLink(cliCtx client.Context) http.HandlerFunc {
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

		msg := types.NewMsgLink(fromAddr, mux.Vars(r)[clientUtils.PathVarChain], req.Address)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

// GetHandlerConfirmTx returns the handler to confirm a tx outpoint
func GetHandlerConfirmTx(cliCtx client.Context) http.HandlerFunc {
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

		msg := types.NewMsgConfirmOutpoint(fromAddr, out)

		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}
