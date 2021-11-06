package rest

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"

	clientUtils "github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/snapshot/keeper"
	"github.com/axelarnetwork/axelar-core/x/snapshot/types"
)

// ReqRegisterProxy defines the properties of a tx request's body
type ReqRegisterProxy struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
}

// ReqDeactivateProxy defines the properties of a tx request's body
type ReqDeactivateProxy struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
}

// RegisterRoutes registers rest routes for this module
func RegisterRoutes(cliCtx client.Context, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/tx/%s/registerProxy/{voter}", types.ModuleName), registerProxyHandlerFn(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/tx/%s/deactivateProxy", types.ModuleName), deactivateProxyHandlerFn(cliCtx)).Methods("POST")

	registerQuery := clientUtils.RegisterQueryHandlerFn(r, types.RestRoute)
	registerQuery(GetHandlerQueryProxy(cliCtx), keeper.QProxy, clientUtils.PathVarCosmosAddress)
	registerQuery(GetHandlerQueryOperator(cliCtx), keeper.QOperator, clientUtils.PathVarCosmosAddress)
	registerQuery(GetHandlerQuerySnapshot(cliCtx), keeper.QInfo, clientUtils.PathVarCounter)
}

func registerProxyHandlerFn(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract voter address path variable
		vars := mux.Vars(r)
		bech32Addr := vars["voter"]
		voter, err := sdk.AccAddressFromBech32(bech32Addr)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		var req ReqRegisterProxy
		if !rest.ReadRESTReq(w, r, cliCtx.LegacyAmino, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}
		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}
		fromAddr, ok := clientUtils.ExtractReqSender(w, req.BaseReq)
		if !ok {
			return
		}

		msg := types.NewRegisterProxyRequest(sdk.ValAddress(fromAddr), voter)
		tx.WriteGeneratedTxResponse(cliCtx, w, baseReq, msg)
	}
}

func deactivateProxyHandlerFn(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqDeactivateProxy
		if !rest.ReadRESTReq(w, r, cliCtx.LegacyAmino, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}
		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}
		fromAddr, ok := clientUtils.ExtractReqSender(w, req.BaseReq)
		if !ok {
			return
		}

		msg := types.NewDeactivateProxyRequest(sdk.ValAddress(fromAddr))
		tx.WriteGeneratedTxResponse(cliCtx, w, baseReq, msg)
	}
}
