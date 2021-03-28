package rest

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"

	clientUtils "github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

const (
	TxMethodKeygenStart         = "start"
	TxMethodMasterKeyAssignNext = "assign"
	TxMethodMasterKeyRotate     = "rotate"
)

// ReqKeygenStart represents a key-gen request
type ReqKeygenStart struct {
	BaseReq        rest.BaseReq `json:"base_req" yaml:"base_req"`
	NewKeyId       string       `json:"key_id" yaml:"key_id"`
	ValidatorCount int64        `json:"validator_count" yaml:"validator_count"`
}

// ReqMasterKeyAssignNext represents a request to assign a new master key
type ReqMasterKeyAssignNext struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	KeyId   string       `json:"key_id" yaml:"key_id"`
}

// ReqMasterKeyRotate represents a request to rotate a master key
type ReqMasterKeyRotate struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	Chain   string       `json:"chain" yaml:"chain"`
}

// RegisterRoutes registers all REST routes with the given router
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	registerTx := clientUtils.RegisterTxHandlerFn(r, types.RestRoute)
	registerTx(GetHandlerKeygenStart(cliCtx), TxMethodKeygenStart)
	registerTx(GetHandlerMasterKeyAssignNext(cliCtx), TxMethodMasterKeyAssignNext, clientUtils.PathVarChain)
	registerTx(GetHandlerMasterKeyRotate(cliCtx), TxMethodMasterKeyRotate, clientUtils.PathVarChain)
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

		msg := types.NewMsgKeygenStart(sender, req.NewKeyId, req.ValidatorCount)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

func GetHandlerMasterKeyAssignNext(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqMasterKeyAssignNext
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

		msg := types.MsgAssignNextMasterKey{
			Sender: sender,
			Chain:  mux.Vars(r)[clientUtils.PathVarChain],
			KeyID:  req.KeyId,
		}
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

func GetHandlerMasterKeyRotate(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqMasterKeyRotate
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

		msg := types.MsgRotateMasterKey{
			Sender: sender,
			Chain:  mux.Vars(r)[clientUtils.PathVarChain],
		}
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}
