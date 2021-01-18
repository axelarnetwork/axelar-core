package rest

import (
	"fmt"
	"net/http"
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"

	"github.com/axelarnetwork/axelar-core/x/tss/types"
	"github.com/axelarnetwork/axelar-core/x/balance/exported"
	clientUtils "github.com/axelarnetwork/axelar-core/utils"
)

type ReqKeygenStart struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	NewKeyId string `json:"key_id" yaml:"key_id"`
	Threshold int `json:"threshold" yaml:"threshold"`
}

type ReqMasterkey struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
}

type ReqMasterkeyAssign struct {
	ReqMasterkey
	KeyId string `json:"key_id" yaml:"key_id"`
	Threshold int `json:"threshold" yaml:"threshold"`
}

type ReqMasterkeyRotate struct {
	ReqMasterkey
}

func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/tx/%s/keygen/start", types.ModuleName), keygenStartHandlerFn(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/tx/%s/masterkey/assign/{chain}", types.ModuleName), masterkeyAssignHandlerFn(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/tx/%s/masterkey/rotate/{chain}", types.ModuleName), masterkeyRotateHandlerFn(cliCtx)).Methods("POST")
}

func masterkeyRotateHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		// 1. Build request object
		baseReq, ok := clientUtils.GetValidatedBaseReq(w, r, cliCtx)
		if !ok {
			return
		}
		var req ReqMasterkeyRotate
		req.BaseReq = *baseReq

		// 2. Extract request params
		fromAddr, ok := clientUtils.GetBaseReqFromAddress(w, req.BaseReq)
		if !ok {
			return
		}

		// 3. Extract router params
		chain := mux.Vars(r)["chain"]

		// 4. Build tx message
		msg := types.MsgRotateMasterKey{
			Sender: fromAddr,
			Chain: exported.ChainFromString(chain),
		}
		if err := msg.ValidateBasic(); err != nil {
			// @fix {"error":"chain-id required but not specified"}
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

func masterkeyAssignHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		// 1. Build request object
		baseReq, ok := clientUtils.GetValidatedBaseReq(w, r, cliCtx)
		if !ok {
			return
		}
		var req ReqMasterkeyAssign
		req.BaseReq = *baseReq

		// 2. Extract request params
		fromAddr, ok := clientUtils.GetBaseReqFromAddress(w, req.BaseReq)
		if !ok {
			return
		}

		// 3. Extract router params
		chain := mux.Vars(r)["chain"]

		// 4. Build tx message
		msg := types.MsgAssignNextMasterKey{
			Sender: fromAddr,
			Chain: exported.ChainFromString(chain),
			KeyID: req.KeyId,
		}
		if err := msg.ValidateBasic(); err != nil {
			// @fix {"error":"chain-id required but not specified"}
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

func keygenStartHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		var req ReqKeygenStart
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}
		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		fromAddr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		msg := types.MsgKeygenStart{
			Sender:    fromAddr,
			NewKeyID:  req.NewKeyId,
			Threshold: req.Threshold,
		}
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

