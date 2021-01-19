package rest

import (
	"fmt"
	"github.com/axelarnetwork/axelar-core/x/broadcast/types"
	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/gorilla/mux"
	"net/http"
)

// SendReq defines the properties of a send request's body.
type SendReq struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
}

func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/tx/%s/registerProxy/{voter}", types.ModuleName), registerProxyHandlerFn(cliCtx)).Methods("POST")
}

func registerProxyHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		bech32Addr := vars["voter"]

		// Extract the validator address from the request path
		// -	Use GetFromFields to allow use of key names
		voter, err := sdk.AccAddressFromBech32(bech32Addr)
		if err != nil {
			//  - Should validate address using CLI context?
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// Read the request parameters
		var req SendReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}

		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		// Extract the sender address from the request params (instead of using cliCtx)
		fromAddr, err := sdk.AccAddressFromBech32(baseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// Val address is validator's operator
		// @? Changing the fromAddr in the request params does not seem to affect the message - validator operator address is always used
		msg := types.NewMsgRegisterProxy(sdk.ValAddress(fromAddr), voter)
		utils.WriteGenerateStdTxResponse(w, cliCtx, baseReq, []sdk.Msg{msg})
	}
}