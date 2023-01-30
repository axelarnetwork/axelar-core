package rest

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/evm/keeper"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
)

// query parameters
const (
	QueryParamKeyRole = "key_role"
	QueryParamKeyID   = "key_id"
	QueryParamSymbol  = keeper.BySymbol
	QueryParamAsset   = keeper.ByAsset
)

// GetHandlerQueryTokenAddress returns a handler to query an EVM chain address
func GetHandlerQueryTokenAddress(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		chain := mux.Vars(r)[utils.PathVarChain]

		symbol := r.URL.Query().Get(QueryParamSymbol)
		asset := r.URL.Query().Get(QueryParamAsset)

		var bz []byte
		var err error
		switch {
		case symbol != "" && asset == "":
			bz, _, err = cliCtx.Query(fmt.Sprintf("custom/%s/%s/%s/%s", types.QuerierRoute, keeper.QTokenAddressBySymbol, chain, symbol))
		case symbol == "" && asset != "":
			bz, _, err = cliCtx.Query(fmt.Sprintf("custom/%s/%s/%s/%s", types.QuerierRoute, keeper.QTokenAddressByAsset, chain, asset))
		default:
			rest.WriteErrorResponse(w, http.StatusBadRequest, "lookup must be either by asset name or symbol")
			return
		}

		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrap(err, types.ErrFTokenAddress).Error())
			return
		}

		var res types.QueryTokenAddressResponse
		types.ModuleCdc.MustUnmarshalLengthPrefixed(bz, &res)

		rest.PostProcessResponse(w, cliCtx, res)
	}
}
