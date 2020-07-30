package rest

import (
	"crypto/sha256"
	"encoding/hex"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/cosmos/sdk-tutorials/scavenge/x/scavenge/types"
	"github.com/gorilla/mux"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
)

func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(
	"/scavenge/create",
	CreateScavengeRequestHandlerFn(cliCtx),
	).Methods("POST")
	r.HandleFunc(
		"/scavenge/commit",
		CommitSolutionRequestHandlerFn(cliCtx),
	).Methods("POST")
	r.HandleFunc(
		"/scavenge/reveal",
		RevealScavengeRequestHandlerFn(cliCtx),
	).Methods("POST")
}

// Action TX body
type ScavengeReq struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	Reward  sdk.Coins    `json:"reward" yaml:"reward"`
	Solution string  	 `json:"solution" yaml:"solution"`
	Description string   `json:"description" yaml:"description"`
}

func CreateScavengeRequestHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ScavengeReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		var solution = req.Solution
		var solutionHash = sha256.Sum256([]byte(solution))
		var solutionHashString = hex.EncodeToString(solutionHash[:])

		fromAddr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		msg := types.NewMsgCreateScavenge(fromAddr, req.Description, solutionHashString, req.Reward)

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

func CommitSolutionRequestHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ScavengeReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		var solution = req.Solution
		var solutionHash = sha256.Sum256([]byte(solution))
		var solutionHashString = hex.EncodeToString(solutionHash[:])

		fromAddr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		var scavenger = req.BaseReq.From
		var solutionScavengerHash = sha256.Sum256([]byte(solution + scavenger))
		var solutionScavengerHashString = hex.EncodeToString(solutionScavengerHash[:])

		msg := types.NewMsgCommitSolution(fromAddr, solutionHashString, solutionScavengerHashString)

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

func RevealScavengeRequestHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ScavengeReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		var solution = req.Solution

		fromAddr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		msg := types.NewMsgRevealSolution(fromAddr, solution)

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}
