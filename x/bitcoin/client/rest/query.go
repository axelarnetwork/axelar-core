package rest

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/rest"
)

func QueryMasterAddressHandlerFn(cliCtx context.CLIContext, queryRoute string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		path := fmt.Sprintf("custom/%s/%s", queryRoute, "MasterAddress") // @TODO use cli route constant
		res, _, err := cliCtx.QueryWithData(path, nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		if len(res) == 0 {
			rest.PostProcessResponse(w, cliCtx, "")
			// rest.PostProcessResponse(w, cliCtx, types.BtcAddress{})
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}