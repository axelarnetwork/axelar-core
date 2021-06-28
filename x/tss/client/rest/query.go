package rest

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"net/http"

	"github.com/axelarnetwork/axelar-core/utils"

	"github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/types/rest"
)

// query parameters
const (
	QueryParamFeeRate = "fee_rate"
)

// QueryHandlerDepositAddress returns a handler to query a deposit address
func QueryHandlerDepositAddress(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		vars := mux.Vars(r)
		queryData, err := cliCtx.LegacyAmino.MarshalJSON(types.DepositQueryParams{Chain: vars[utils.PathVarChain], Address: vars[utils.PathVarEthereumAddress]})
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, keeper.QueryDepositAddress), queryData)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrap(err, types.ErrFDepositAddress).Error())
			return
		}

		if len(res) == 0 {
			rest.PostProcessResponse(w, cliCtx, "")
			return
		}

		rest.PostProcessResponse(w, cliCtx, string(res))
	}
}
