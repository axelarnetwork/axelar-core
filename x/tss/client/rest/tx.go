package rest

import (
	"net/http"

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
	TxMethodKeygenStart         = "start"
	TxMethodMasterKeyAssignNext = "assign"
	TxMethodMasterKeyRotate     = "rotate"
)

// ReqKeygenStart represents a key-gen request
type ReqKeygenStart struct {
	BaseReq    rest.BaseReq `json:"base_req" yaml:"base_req"`
	NewKeyID   string       `json:"key_id" yaml:"key_id"`
	SubsetSize int64        `json:"validator_count" yaml:"validator_count"`
}

// ReqKeyAssignNext represents a request to assign a new key
type ReqKeyAssignNext struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	KeyID   string       `json:"key_id" yaml:"key_id"`
	KeyRole string       `json:"key_role" yaml:"key_role"`
}

// ReqKeyRotate represents a request to rotate a key
type ReqKeyRotate struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	Chain   string       `json:"chain" yaml:"chain"`
	KeyRole string       `json:"key_role" yaml:"key_role"`
}

// RegisterRoutes registers all REST routes with the given router
func RegisterRoutes(cliCtx client.Context, r *mux.Router) {
	registerTx := clientUtils.RegisterTxHandlerFn(r, types.RestRoute)
	registerTx(GetHandlerKeygenStart(cliCtx), TxMethodKeygenStart)
	registerTx(GetHandlerKeyAssignNext(cliCtx), TxMethodMasterKeyAssignNext, clientUtils.PathVarChain)
	registerTx(GetHandlerKeyRotate(cliCtx), TxMethodMasterKeyRotate, clientUtils.PathVarChain)
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

		msg := types.NewMsgKeygenStart(sender, req.NewKeyID, req.SubsetSize)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

// GetHandlerKeyAssignNext returns the handler to assign a role to an existing key
func GetHandlerKeyAssignNext(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqKeyAssignNext
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

		keyRole, err := exported.KeyRoleFromStr(req.KeyRole)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		msg := types.NewMsgAssignNextKey(sender, mux.Vars(r)[clientUtils.PathVarChain], req.KeyID, keyRole)
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

		keyRole, err := exported.KeyRoleFromStr(req.KeyRole)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		msg := types.NewMsgRotateKey(sender, mux.Vars(r)[clientUtils.PathVarChain], keyRole)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}
