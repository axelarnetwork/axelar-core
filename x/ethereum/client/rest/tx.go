package rest

import (
	"fmt"
	clientUtils "github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/ethereum/keeper"
	"github.com/axelarnetwork/axelar-core/x/ethereum/types"
	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/gorilla/mux"
	"net/http"
)

func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router ) {
	// @TODO set tx routes using module constants
	r.HandleFunc(fmt.Sprintf("/tx/%s/signTx", types.RestRoute), signTxHandlerFn(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/tx/%s/verifyTx", types.RestRoute), verifyTxHandlerFn(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/tx/%s/signPending", types.RestRoute), signPendingTransfersHandlerFn(cliCtx)).Methods("POST")

	r.HandleFunc(fmt.Sprintf("/query/%s/%s", types.RestRoute, keeper.QueryMasterKey), QueryMasterAddress(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/query/%s/%s/{contractAddr}/{recipeint}", types.RestRoute, keeper.CreateMintTx), QueryCreateMintTx(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/query/%s/%s/{byteCode}", types.RestRoute, keeper.CreateDeployTx), QueryCreateDeployTx(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/query/%s/%s/{txID}", types.RestRoute, keeper.SendTx), QuerySendTx(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/query/%s/%s/{commandId}/{contractAddr}", types.RestRoute, keeper.SendMintTx), QuerySendMintTx(cliCtx)).Methods("GET")
}

type ReqSignTx struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	TxJson string `json:"tx_json" yaml:"tx_json"`
}

type ReqVerifyTx struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	TxInfo string `json:"tx_info" yaml:"tx_info"`
}

type ReqSignPendingTransfers struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
}

func signTxHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
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

		json := []byte(req.TxJson)
		err := types.ValidTxJson(json)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		msg := types.NewMsgSignTx(fromAddr, json)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

func verifyTxHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
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

		json := []byte(req.TxJson)
		err := types.ValidTxJson(json)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		msg := types.NewMsgVerifyTx(fromAddr, json)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

func signPendingTransfersHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
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

		msg := types.NewMsgSignPendingTransfersTx(fromAddr)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}
