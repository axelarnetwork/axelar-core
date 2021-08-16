package rest

import (
	"net/http"

	"github.com/axelarnetwork/axelar-core/x/tss/keeper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/types/rest"

	clientUtils "github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

// rest routes
const (
	TxKeygenStart     = "start"
	TxMasterKeyRotate = "rotate"

	QuerySigStatus   = keeper.QuerySigStatus
	QueryKeyStatus   = keeper.QueryKeyStatus
	QueryRecovery    = keeper.QueryRecovery
	QueryDeactivated = keeper.QueryDeactivated
)

// ReqKeygenStart represents a key-gen request
type ReqKeygenStart struct {
	BaseReq                    rest.BaseReq `json:"base_req" yaml:"base_req"`
	NewKeyID                   string       `json:"key_id" yaml:"key_id"`
	SubsetSize                 int64        `json:"validator_count" yaml:"validator_count"`
	KeyShareDistributionPolicy string       `json:"key_share_distribution_policy" yaml:"key_share_distribution_policy"`
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
	registerTx := clientUtils.RegisterTxHandlerFn(r, types.RestRoute)
	registerTx(GetHandlerKeygenStart(cliCtx), TxKeygenStart)
	registerTx(GetHandlerKeyRotate(cliCtx), TxMasterKeyRotate, clientUtils.PathVarChain)

	registerQuery := clientUtils.RegisterQueryHandlerFn(r, types.RestRoute)
	registerQuery(QueryHandlerSigStatus(cliCtx), QuerySigStatus, clientUtils.PathVarSigID)
	registerQuery(QueryHandlerKeyStatus(cliCtx), QueryKeyStatus, clientUtils.PathVarKeyID)
	registerQuery(QueryHandlerRecovery(cliCtx), QueryRecovery)
	registerQuery(QueryHandlerDeactivatedOperator(cliCtx), QueryDeactivated)
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

		keyShareDistributionPolicy, err := exported.KeyShareDistributionPolicyFromSimpleStr(req.KeyShareDistributionPolicy)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		msg := types.NewStartKeygenRequest(sender, req.NewKeyID, req.SubsetSize, keyShareDistributionPolicy)
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
