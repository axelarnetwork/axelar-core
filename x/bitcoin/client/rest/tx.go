package rest

import (
	"fmt"
	"net/http"

	"github.com/btcsuite/btcd/wire"
	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/gorilla/mux"

	clientUtils "github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
)

// RegisterRoutes registers this module's REST routes with the given router
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/tx/%s/link/{chain}", types.RestRoute), linkHandlerFn(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/tx/%s/sign", types.RestRoute), signRawTxHandlerFn(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/tx/%s/verify", types.RestRoute), verifyTxHandlerFn(cliCtx)).Methods("POST")

	r.HandleFunc(fmt.Sprintf("/query/%s/%s/{chain}/{address}", types.RestRoute, keeper.QueryDepositAddress), QueryDepositAddress(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/query/%s/%s/{txID}", types.RestRoute, keeper.QueryConsolidationAddress), QueryConsolidationAddress(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/query/%s/%s/{txID}", types.RestRoute, keeper.QueryOutInfo), QueryTxInfo(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/query/%s/%s/{txID}", types.RestRoute, keeper.QueryRawTx), QueryRawTx(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/query/%s/%s/{txID}", types.RestRoute, keeper.SendTx), QuerySendTx(cliCtx)).Methods("GET")
}

// ReqLink represents a request to link a cross-chain address to a Bitcoin address
type ReqLink struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	Address string       `json:"address" yaml:"address"`
}

// ReqVerifyTx represents a request to verify a Bitcoin transaction
type ReqVerifyTx struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	TxInfo  string       `json:"tx_info" yaml:"tx_info"`
}

// ReqSignTx represents a request to sign a Bitcoin transaction
type ReqSignTx struct {
	BaseReq  rest.BaseReq `json:"base_req" yaml:"base_req"`
	OutPoint string       `json:"outpoint" yaml:"outpoint"`
	TxJson   string       `json:"tx_json" yaml:"tx_json"`
}

func linkHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
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

		vars := mux.Vars(r)
		msg := types.MsgLink{Sender: fromAddr, RecipientChain: vars["chain"], RecipientAddr: req.Address}
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

func signRawTxHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqSignTx
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

		var tx *wire.MsgTx
		types.ModuleCdc.MustUnmarshalJSON([]byte(req.TxJson), &tx)

		outpoint, err := types.OutPointFromStr(req.OutPoint)
		if err != nil {
			return
		}

		msg := types.NewMsgSignTx(fromAddr, outpoint, tx)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

func verifyTxHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqVerifyTx
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

		msg := types.MsgVerifyTx{Sender: fromAddr, OutPointInfo: out}

		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}
