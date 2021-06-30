package rest

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"net/http"

	"github.com/axelarnetwork/axelar-core/utils"

	"github.com/axelarnetwork/axelar-core/x/tss/keeper"
	"github.com/axelarnetwork/axelar-core/x/tss/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/types/rest"
)

// QueryHandlerGetSig returns a handler to query a signature by its sigID
func QueryHandlerGetSig(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		queryData := []byte(mux.Vars(r)[utils.PathVarSigID])
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, keeper.QueryGetSig), queryData)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		var sigResponse types.QuerySigResponse
		err = sigResponse.Unmarshal(res)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrapf(err, "failed to get sig").Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, sigResponse)
	}
}
