package rest

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/axelarnetwork/axelar-core/utils"

	"github.com/axelarnetwork/axelar-core/x/tss/keeper"
	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	"github.com/axelarnetwork/axelar-core/x/tss/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gorilla/mux"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
)

// query parameters
const (
	QueryParamKeyID     = "key_id"
	QueryParamValidator = "validator"
)

// QueryHandlerSigStatus returns a handler to query a signature's vote status by its sigID
func QueryHandlerSigStatus(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		sigID := mux.Vars(r)[utils.PathVarSigID]
		bz, _, err := cliCtx.Query(fmt.Sprintf("custom/%s/%s/%s", types.QuerierRoute, keeper.QuerySignature, sigID))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		var res types.QuerySignatureResponse
		if err := types.ModuleCdc.UnmarshalBinaryLengthPrefixed(bz, &res); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrapf(err, "failed to get sig status").Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// QueryHandlerKeyStatus returns a handler to query a key's vote status by its keyID
func QueryHandlerKeyStatus(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		keyID := mux.Vars(r)[utils.PathVarKeyID]
		bz, _, err := cliCtx.Query(fmt.Sprintf("custom/%s/%s/%s", types.QuerierRoute, keeper.QueryKey, keyID))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		var res types.QueryKeyResponse
		if err := types.ModuleCdc.UnmarshalBinaryLengthPrefixed(bz, &res); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrapf(err, "failed to get key status").Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// QueryHandlerRecovery returns a handler to query the recovery data for some operator and key IDs
func QueryHandlerRecovery(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		r.ParseForm()
		keyIDs := r.Form[QueryParamKeyID]
		validator := r.URL.Query().Get(QueryParamValidator)
		address, err := sdk.ValAddressFromBech32(validator)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrapf(err, "failed to parse validator address").Error())
			return
		}

		requests := make([]tofnd.RecoverRequest, len(keyIDs))
		for i, keyID := range keyIDs {
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", types.QuerierRoute, keeper.QueryRecovery, keyID), nil)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrapf(err, "failed to get recovery data").Error())
				return
			}

			var recResponse types.QueryRecoveryResponse
			err = recResponse.Unmarshal(res)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrapf(err, "failed to get recovery data").Error())
				return
			}

			index := utils.IndexOf(recResponse.PartyUids, address.String())
			if index == -1 {
				// not participating
				rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrapf(err, "recovery data does not contain address %s", address.String()).Error())
				return
			}

			requests[i] = tofnd.RecoverRequest{
				KeygenInit: &tofnd.KeygenInit{
					NewKeyUid:        keyID,
					Threshold:        recResponse.Threshold,
					PartyUids:        recResponse.PartyUids,
					PartyShareCounts: recResponse.PartyShareCounts,
					MyPartyIndex:     int32(index),
				},
				ShareRecoveryInfos: recResponse.ShareRecoveryInfos,
			}
		}

		rest.PostProcessResponse(w, cliCtx, requests)
	}
}

// QueryHandlerKeyID returns a handler to query the keyID of the most recent key given the keyChain and keyRole
func QueryHandlerKeyID(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		chain := mux.Vars(r)[utils.PathVarChain]
		role := mux.Vars(r)[utils.PathVarKeyRole]

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s/%s", types.QuerierRoute, keeper.QueryKeyID, chain, role), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, string(res))
	}
}

// QueryHandlerKeySharesByKeyID returns a handler to query for a list of key shares for a given keyID
func QueryHandlerKeySharesByKeyID(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		keyID := mux.Vars(r)[utils.PathVarKeyID]

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", types.QuerierRoute, keeper.QueryKeySharesByKeyID, keyID), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		var keyShareResponse types.QueryKeyShareResponse
		err = keyShareResponse.Unmarshal(res)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrapf(err, "failed to get key share information").Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, keyShareResponse)
	}
}

// QueryHandlerKeySharesByValidator returns a handler to query for a list of key shares held by a validator address
func QueryHandlerKeySharesByValidator(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		validatorAddress := mux.Vars(r)[utils.PathVarCosmosAddress]

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", types.QuerierRoute, keeper.QueryKeySharesByValidator, validatorAddress), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		var keyShareResponse types.QueryKeyShareResponse
		err = keyShareResponse.Unmarshal(res)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrapf(err, "failed to get key share information").Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, keyShareResponse)
	}
}

// QueryHandlerDeactivatedOperator returns a list of deactivated operator addresses
func QueryHandlerDeactivatedOperator(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		path := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, keeper.QueryDeactivated)

		bz, _, err := cliCtx.Query(path)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrapf(err, "could not get deactivated operator addresses").Error())
			return
		}

		var res types.QueryDeactivatedOperatorsResponse
		types.ModuleCdc.MustUnmarshalBinaryLengthPrefixed(bz, &res)

		rest.PostProcessResponse(w, cliCtx, res)
	}
}
