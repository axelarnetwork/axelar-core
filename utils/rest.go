package utils

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"net/http"
)

// Extract the sender address from an SDK base request
func ExtractReqSender(w http.ResponseWriter, req rest.BaseReq) (sdk.AccAddress, bool) {
	sender, err := sdk.AccAddressFromBech32(req.From)
	if err != nil {
		rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return nil, false
	}

	return sender, true
}
