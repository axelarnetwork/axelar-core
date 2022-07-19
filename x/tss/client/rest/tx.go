package rest

import (
	"encoding/hex"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"

	clientUtils "github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/keeper"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

// rest routes
const (
	TxKeygenStart          = "start"
	TxMasterKeyRotate      = "rotate"
	TxRegisterExternalKeys = "register-external-keys"

	QuerySignature                = keeper.QuerySignature
	QueryKey                      = keeper.QueryKey
	QueryRecovery                 = keeper.QueryRecovery
	QueryKeyID                    = keeper.QueryKeyID
	QueryKeySharesByKeyID         = keeper.QueryKeySharesByKeyID
	QueryActiveOldKeys            = keeper.QueryActiveOldKeys
	QueryActiveOldKeysByValidator = keeper.QueryActiveOldKeysByValidator
	QueryKeySharesByValidator     = keeper.QueryKeySharesByValidator
	QueryDeactivated              = keeper.QueryDeactivated
	QueryExternalKeyID            = "external-key-id"
)

// ReqRegisterExternalKey represents a request to register external keys for a chain
type ReqRegisterExternalKey struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	KeyIDs  []string     `json:"key_ids" yaml:"key_ids"`
	PubKeys []string     `json:"pub_keys" yaml:"pub_keys"`
}

// ReqKeygenStart represents a key-gen request
type ReqKeygenStart struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	KeyID   string       `json:"key_id" yaml:"key_id"`
	KeyRole string       `json:"key_role" yaml:"key_role"`
	KeyType string       `json:"key_type" yaml:"key_type"`
}

// ReqKeyRotate represents a request to rotate a key
type ReqKeyRotate struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	Chain   string       `json:"chain" yaml:"chain"`
	KeyRole string       `json:"key_role" yaml:"key_role"`
	KeyID   string       `json:"key_id" yaml:"key_id"`
}

// RegisterRoutes registers all REST routes with the given router
func RegisterRoutes(cliCtx client.Context, r *mux.Router) {
	registerQuery := clientUtils.RegisterQueryHandlerFn(r, types.RestRoute)
	registerQuery(QueryHandlerSigStatus(cliCtx), QuerySignature, clientUtils.PathVarSigID)
	registerQuery(QueryHandlerKeyStatus(cliCtx), QueryKey, clientUtils.PathVarKeyID)
	registerQuery(QueryHandlerRecovery(cliCtx), QueryRecovery)
	registerQuery(QueryHandlerKeyID(cliCtx), QueryKeyID, clientUtils.PathVarChain, clientUtils.PathVarKeyRole)
	registerQuery(QueryHandlerKeySharesByKeyID(cliCtx), QueryKeySharesByKeyID, clientUtils.PathVarKeyID)
	registerQuery(QueryHandlerActiveOldKeys(cliCtx), QueryActiveOldKeys, clientUtils.PathVarChain, clientUtils.PathVarKeyRole)
	registerQuery(QueryHandlerActiveOldKeysByValidator(cliCtx), QueryActiveOldKeysByValidator, clientUtils.PathVarCosmosAddress)
	registerQuery(QueryHandlerKeySharesByValidator(cliCtx), QueryKeySharesByValidator, clientUtils.PathVarCosmosAddress)
	registerQuery(QueryHandlerDeactivatedOperator(cliCtx), QueryDeactivated)
	registerQuery(QueryHandlerExternalKeyID(cliCtx), QueryExternalKeyID, clientUtils.PathVarChain)
}

// GetHandlerKeygenStart returns the handler to start a keygen
func GetHandlerKeygenStart(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqKeygenStart
		if !rest.ReadRESTReq(w, r, cliCtx.LegacyAmino, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}
		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		sender, ok := clientUtils.ExtractReqSender(w, req.BaseReq)
		if !ok {
			return
		}

		keyRole, err := exported.KeyRoleFromSimpleStr(req.KeyRole)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		keyType, err := exported.KeyTypeFromSimpleStr(req.KeyType)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		if !types.TSSEnabled && keyType == exported.Multisig {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "threshold signing is disabled")
			return
		}

		msg := types.NewStartKeygenRequest(sender, req.KeyID, keyRole, keyType)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

// GetHandlerKeyRotate returns a handler that rotates the active keys to the next assigned ones
func GetHandlerKeyRotate(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqKeyRotate
		if !rest.ReadRESTReq(w, r, cliCtx.LegacyAmino, &req) {
			return
		}
		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		sender, ok := clientUtils.ExtractReqSender(w, req.BaseReq)
		if !ok {
			return
		}

		keyRole, err := exported.KeyRoleFromSimpleStr(req.KeyRole)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		msg := types.NewRotateKeyRequest(sender, mux.Vars(r)[clientUtils.PathVarChain], keyRole, req.KeyID)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

// GetHandlerRegisterExternalKeys returns the handler to register an external key
func GetHandlerRegisterExternalKeys(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqRegisterExternalKey
		if !rest.ReadRESTReq(w, r, cliCtx.LegacyAmino, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		fromAddr, ok := clientUtils.ExtractReqSender(w, req.BaseReq)
		if !ok {
			return
		}

		if len(req.KeyIDs) != len(req.PubKeys) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "length mismatch between key IDs and pub keys")
			return
		}

		externalKeys := make([]types.RegisterExternalKeysRequest_ExternalKey, len(req.KeyIDs))
		for i, keyID := range req.KeyIDs {
			pubKeyBytes, err := hex.DecodeString(req.PubKeys[i])
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}

			externalKeys[i] = types.RegisterExternalKeysRequest_ExternalKey{ID: exported.KeyID(keyID), PubKey: pubKeyBytes}
		}

		msg := types.NewRegisterExternalKeysRequest(fromAddr, mux.Vars(r)[clientUtils.PathVarChain], externalKeys...)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}
