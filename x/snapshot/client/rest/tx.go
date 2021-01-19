package rest

import (
	"fmt"
	clientUtils "github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/snapshot/types"
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
	r.HandleFunc(fmt.Sprintf("/tx/%s/now", types.ModuleName), regsiterSnapshotHandlerFn(cliCtx)).Methods("POST")
}

func regsiterSnapshotHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		var req SendReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}
		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		//fromAddr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		//if err != nil {
		//	rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		//	return
		//}
		fromAddr, ok := clientUtils.ExtractReqSender(w, req.BaseReq)
		if !ok {
			return
		}

		msg := types.MsgSnapshot{ Sender: fromAddr }
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}
