package rest

import (
	"context"
	"net/http"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"
)

// GetHandlerQueryDepositAddress returns a handler to query the state of a deposit address
func GetHandlerQueryDepositAddress(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		recipientChain := mux.Vars(r)[utils.PathVarRecipientChain]
		recipientAddr := mux.Vars(r)[utils.PathVarLinkedAddress]

		queryClient := types.NewQueryServiceClient(cliCtx)

		res, err := queryClient.DepositAddress(context.Background(),
			&types.DepositAddressRequest{
				RecipientAddr:  recipientAddr,
				RecipientChain: recipientChain,
			})
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrap(err, "error querying deposit address").Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}
