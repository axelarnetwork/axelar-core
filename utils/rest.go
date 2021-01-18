package utils

import (
	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"net/http"
)

// Unmarshal, sanitize and validate rest request into a base request object
func GetValidatedBaseReq(w http.ResponseWriter, r *http.Request, cliCtx context.CLIContext) (*rest.BaseReq, bool) {
	var baseReq rest.BaseReq
	if !rest.ReadRESTReq(w, r, cliCtx.Codec, &baseReq) {
		return nil, false
	}

	baseReq = baseReq.Sanitize()
	if !baseReq.ValidateBasic(w) {
		return nil, false
	}

	return &baseReq, true
}

// Extract the sender address from an SDK base request
func GetBaseReqFromAddress(w http.ResponseWriter, req rest.BaseReq) (sdk.AccAddress, bool) {
	fromAddr, err := sdk.AccAddressFromBech32(req.From)
	if err != nil {
		rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return nil, false
	}

	return fromAddr, true
}
