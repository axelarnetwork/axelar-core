package rest

import (
	"fmt"
	"github.com/axelarnetwork/axelar-core/x/balance/exported"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	balance "github.com/axelarnetwork/axelar-core/x/balance/exported"
	"github.com/gorilla/mux"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/rest"
)

type CrossChainAddress struct {
	Chain   exported.Chain `json:"chain" yaml:"chain"`
	Address string `json:"addres" yaml:"addres"`
}

func QueryDepositAddress(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		vars := mux.Vars(r)
		path := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, keeper.QueryDepositAddress)

		chain := balance.ChainFromString(vars["chain"])
		res, _, err := cliCtx.QueryWithData(path, cliCtx.Codec.MustMarshalJSON(CrossChainAddress{Chain: chain, Address: vars["address"]}))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		if len(res) == 0 {
			// @TODO use appropriate response struct
			rest.PostProcessResponse(w, cliCtx, "")
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

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