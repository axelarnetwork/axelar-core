package rest

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/tss/keeper"
	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
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
		if err := types.ModuleCdc.UnmarshalLengthPrefixed(bz, &res); err != nil {
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
		if err := types.ModuleCdc.UnmarshalLengthPrefixed(bz, &res); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrapf(err, "failed to get key status").Error())
			return
		}

		// force the rotatedAt field to be nil, if the timestamp is for Jan 1, 1970
		if res.RotatedAt != nil && res.RotatedAt.Unix() == 0 {
			res.RotatedAt = nil
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

		if err := r.ParseForm(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		keyIDs := r.Form[QueryParamKeyID]
		validator := r.URL.Query().Get(QueryParamValidator)
		address, err := sdk.ValAddressFromBech32(validator)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrapf(err, "failed to parse validator address").Error())
			return
		}

		requests := make([]tofnd.RecoverRequest, len(keyIDs))
		for i, keyID := range keyIDs {
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s/%s", types.QuerierRoute, keeper.QueryRecovery, keyID, address.String()), nil)
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
					MyPartyIndex:     uint32(index),
				},
				KeygenOutput: recResponse.KeygenOutput,
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

// QueryHandlerActiveOldKeys returns a handler to query for a list of active old key IDs held by a validator address
func QueryHandlerActiveOldKeys(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		chain := mux.Vars(r)[utils.PathVarChain]
		role := mux.Vars(r)[utils.PathVarKeyRole]

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s/%s", types.QuerierRoute, keeper.QueryActiveOldKeys, chain, role), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		var keyShareResponse types.QueryActiveOldKeysResponse
		err = keyShareResponse.Unmarshal(res)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrapf(err, "failed to get key share information").Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, keyShareResponse.KeyIDs)
	}
}

// QueryHandlerActiveOldKeysByValidator returns a handler to query for a list of active old key IDs held by a validator address
func QueryHandlerActiveOldKeysByValidator(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		validatorAddress := mux.Vars(r)[utils.PathVarCosmosAddress]

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", types.QuerierRoute, keeper.QueryActiveOldKeysByValidator, validatorAddress), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		var keyShareResponse types.QueryActiveOldKeysValidatorResponse
		err = keyShareResponse.Unmarshal(res)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, sdkerrors.Wrapf(err, "failed to get key share information").Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, keyShareResponse.KeysInfo)
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

		rest.PostProcessResponse(w, cliCtx, keyShareResponse.ShareInfos)
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
		types.ModuleCdc.MustUnmarshalLengthPrefixed(bz, &res)

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// QueryHandlerExternalKeyID returns a handler to query the keyIDs of the current set of external keys
func QueryHandlerExternalKeyID(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		chain := mux.Vars(r)[utils.PathVarChain]
		path := fmt.Sprintf("custom/%s/%s/%s", types.QuerierRoute, keeper.QExternalKeyID, chain)

		bz, _, err := cliCtx.Query(path)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "could not resolve the external key IDs")
			return
		}

		var res types.QueryExternalKeyIDResponse
		types.ModuleCdc.MustUnmarshalLengthPrefixed(bz, &res)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}
