package rest

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"

	clientUtils "github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

const (
	TxMethodKeygenStart         = "start"
	TxMethodMasterKeyAssignNext = "assign"
	TxMethodMasterKeyRotate     = "rotate"
)

// ReqKeygenStart represents a key-gen request
type ReqKeygenStart struct {
	BaseReq    rest.BaseReq `json:"base_req" yaml:"base_req"`
	NewKeyId   string       `json:"key_id" yaml:"key_id"`
	SubsetSize int64        `json:"validator_count" yaml:"validator_count"`
}

// ReqKeyAssignNext represents a request to assign a new key
type ReqKeyAssignNext struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	KeyId   string       `json:"key_id" yaml:"key_id"`
	KeyRole string       `json:"key_role" yaml:"key_role"`
}

// ReqKeyRotate represents a request to rotate a key
type ReqKeyRotate struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	Chain   string       `json:"chain" yaml:"chain"`
	KeyRole string       `json:"key_role" yaml:"key_role"`
}

// RegisterRoutes registers all REST routes with the given router
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	registerTx := clientUtils.RegisterTxHandlerFn(r, types.RestRoute)
	registerTx(GetHandlerKeygenStart(cliCtx), TxMethodKeygenStart)
	registerTx(GetHandlerKeyAssignNext(cliCtx), TxMethodMasterKeyAssignNext, clientUtils.PathVarChain)
	registerTx(GetHandlerKeyRotate(cliCtx), TxMethodMasterKeyRotate, clientUtils.PathVarChain)
}

func GetHandlerKeygenStart(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqKeygenStart
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
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

		msg := types.NewMsgKeygenStart(sender, req.NewKeyId, req.SubsetSize)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

func GetHandlerKeyAssignNext(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqKeyAssignNext
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
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

		msg := types.NewMsgAssignNextKey(sender, mux.Vars(r)[clientUtils.PathVarChain], req.KeyId, keyRole)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

func GetHandlerKeyRotate(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqKeyRotate
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
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

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}
