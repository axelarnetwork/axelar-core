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
	"github.com/axelarnetwork/axelar-core/x/broadcast/types"
)

// ReqRegisterProxy defines the properties of a tx request's body
type ReqRegisterProxy struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
}

func RegisterRoutes(cliCtx client.Context, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/tx/%s/registerProxy/{voter}", types.ModuleName), registerProxyHandlerFn(cliCtx)).Methods("POST")
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

		msg := types.NewMsgRegisterProxy(sdk.ValAddress(fromAddr), voter)
		tx.WriteGeneratedTxResponse(cliCtx, w, baseReq, msg)
	}
}
