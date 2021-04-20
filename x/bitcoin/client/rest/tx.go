package rest

import (
	"net/http"

	"github.com/btcsuite/btcutil"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
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
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	registerTx := clientUtils.RegisterTxHandlerFn(r, types.RestRoute)
	registerTx(HandlerLink(cliCtx), TxMethodLink, clientUtils.PathVarChain)
	registerTx(HandlerConfirmTx(cliCtx), TxMethodConfirmTx)
	registerTx(HandlerSignPendingTransfersTx(cliCtx), TxMethodSignPendingTransfersTx)

	registerQuery := clientUtils.RegisterQueryHandlerFn(r, types.RestRoute)
	registerQuery(HandlerQueryDepositAddress(cliCtx), QueryMethodDepositAddress, clientUtils.PathVarChain, clientUtils.PathVarEthereumAddress)
	registerQuery(HandlerQueryKeyAddress(cliCtx), QueryMethodKeyAddress, clientUtils.PathVarKeyRole)
	registerQuery(HandlerQueryGetConsolidationTx(cliCtx), QueryMethodGetConsolidationTx)
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

// HandlerSignPendingTransfersTx returns the handler to sign pending transfers to Bitcoin
func HandlerSignPendingTransfersTx(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqSignPendingTransfersTx
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
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

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

// HandlerLink returns the handler to link a Bitcoin address to a cross-chain address
func HandlerLink(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqLink
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
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

		msg := types.MsgLink{Sender: fromAddr, RecipientChain: mux.Vars(r)[clientUtils.PathVarChain], RecipientAddr: req.Address}
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

// HandlerConfirmTx returns the handler to confirm a tx outpoint
func HandlerConfirmTx(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqConfirmOutPoint
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
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
		if err := cliCtx.Codec.UnmarshalJSON([]byte(req.TxInfo), &out); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		msg := types.MsgConfirmOutpoint{Sender: fromAddr, OutPointInfo: out}

		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}
