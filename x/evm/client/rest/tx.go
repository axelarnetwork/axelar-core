package rest

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"

	clientUtils "github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// rest routes
const (
	TxLink                        = "link"
	TxConfirmTokenDeploy          = "confirm-erc20-deploy"
	TxConfirmDeposit              = "confirm-erc20-deposit"
	TxConfirmTransferOperatorship = "confirm-transfer-operatorship"
	TxCreatePendingTransfers      = "create-pending-transfers"
	TxCreateDeployToken           = "create-deploy-token"
	TxCreateBurnTokens            = "create-burn-token"
	TxCreateTransferOperatorship  = "create-transfer-operatorship"
	TxSignCommands                = "sign-commands"
	TxAddChain                    = "add-chain"

	QueryAddress      = "query-address"
	QueryTokenAddress = "token-address"
)

// RegisterRoutes registers this module's REST routes with the given router
func RegisterRoutes(cliCtx client.Context, r *mux.Router) {
	registerTx := clientUtils.RegisterTxHandlerFn(r, types.RestRoute)
	registerTx(GetHandlerLink(cliCtx), TxLink, clientUtils.PathVarChain)
	registerTx(GetHandlerConfirmTokenDeploy(cliCtx), TxConfirmTokenDeploy, clientUtils.PathVarChain)
	registerTx(GetHandlerConfirmDeposit(cliCtx), TxConfirmDeposit, clientUtils.PathVarChain)
	registerTx(GetHandlerConfirmTransferKey(cliCtx), TxConfirmTransferOperatorship, clientUtils.PathVarChain)
	registerTx(GetHandlerCreatePendingTransfers(cliCtx), TxCreatePendingTransfers, clientUtils.PathVarChain)
	registerTx(GetHandlerCreateDeployToken(cliCtx), TxCreateDeployToken, clientUtils.PathVarChain)
	registerTx(GetHandlerCreateBurnTokens(cliCtx), TxCreateBurnTokens, clientUtils.PathVarChain)
	registerTx(GetHandlerCreateTransferOperatorship(cliCtx), TxCreateTransferOperatorship, clientUtils.PathVarChain)
	registerTx(GetHandlerSignCommands(cliCtx), TxSignCommands, clientUtils.PathVarChain)
	registerTx(GetHandlerAddChain(cliCtx), TxAddChain)

	registerQuery := clientUtils.RegisterQueryHandlerFn(r, types.RestRoute)
	registerQuery(GetHandlerQueryTokenAddress(cliCtx), QueryTokenAddress, clientUtils.PathVarChain)
}

// ReqLink represents a request to link a cross-chain address to an EVM chain address
type ReqLink struct {
	BaseReq        rest.BaseReq `json:"base_req" yaml:"base_req"`
	RecipientChain string       `json:"chain" yaml:"chain"`
	RecipientAddr  string       `json:"recipient" yaml:"recipient"`
	Asset          string       `json:"asset" yaml:"asset"`
}

// ReqConfirmChain represents a request to confirm a new chain
type ReqConfirmChain struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	Chain   string       `json:"chain" yaml:"chain"`
}

// ReqConfirmTokenDeploy represents a request to confirm a token deployment
type ReqConfirmTokenDeploy struct {
	BaseReq     rest.BaseReq `json:"base_req" yaml:"base_req"`
	OriginChain string       `json:"origin_chain" yaml:"origin_chain"`
	OriginAsset string       `json:"origin_asset" yaml:"origin_asset"`
	TxID        string       `json:"tx_id" yaml:"tx_id"`
}

// ReqConfirmDeposit represents a request to confirm a deposit
type ReqConfirmDeposit struct {
	BaseReq       rest.BaseReq `json:"base_req" yaml:"base_req"`
	TxID          string       `json:"tx_id" yaml:"tx_id"`
	Amount        string       `json:"amount" yaml:"amount"`
	BurnerAddress string       `json:"burner_address" yaml:"burner_address"`
}

// ReqConfirmTransferKey represents a request to confirm a transfer operatorship
type ReqConfirmTransferKey struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	TxID    string       `json:"tx_id" yaml:"tx_id"`
}

// ReqCreatePendingTransfers represents a request to create commands for all pending transfers
type ReqCreatePendingTransfers struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
}

// ReqCreateDeployToken represents a request to create a deploy token command
type ReqCreateDeployToken struct {
	BaseReq        rest.BaseReq `json:"base_req" yaml:"base_req"`
	OriginChain    string       `json:"origin_chain" yaml:"origin_chain"`
	OriginAsset    string       `json:"origin_asset" yaml:"origin_asset"`
	Symbol         string       `json:"symbol" yaml:"symbol"`
	TokenName      string       `json:"token_name" yaml:"token_name"`
	Decimals       string       `json:"decimals" yaml:"decimals"`
	Capacity       string       `json:"capacity" yaml:"capacity"`
	Address        string       `json:"address" yaml:"address"`
	DailyMintLimit string       `json:"daily_mint_limit" yaml:"daily_mint_limit"`
}

// ReqCreateBurnTokens represents a request to create commands for all outstanding burns
type ReqCreateBurnTokens struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
}

// ReqCreateTransferOperatorship represents a request to create transfer operatorship command
type ReqCreateTransferOperatorship struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	KeyID   string       `json:"key_id" yaml:"key_id"`
}

// ReqSignCommands represents a request to sign pending commands
type ReqSignCommands struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
}

// ReqAddChain represents a request to add a new evm chain command
type ReqAddChain struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	Name    string       `json:"name" yaml:"name"`
	KeyType string       `json:"key_type" yaml:"key_type"`
	Params  types.Params `json:"params" yaml:"params"`
}

// GetHandlerLink returns the handler to link addresses
func GetHandlerLink(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqLink
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

		msg := &types.LinkRequest{
			Chain:          nexus.ChainName(mux.Vars(r)[clientUtils.PathVarChain]),
			Sender:         fromAddr,
			RecipientChain: nexus.ChainName(req.RecipientChain),
			RecipientAddr:  req.RecipientAddr,
			Asset:          req.Asset,
		}

		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

// GetHandlerConfirmTokenDeploy returns a handler to confirm a token deployment
func GetHandlerConfirmTokenDeploy(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqConfirmTokenDeploy
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

		txID := common.HexToHash(req.TxID)
		asset := types.NewAsset(req.OriginChain, req.OriginAsset)
		msg := types.NewConfirmTokenRequest(fromAddr, mux.Vars(r)[clientUtils.PathVarChain], asset, txID)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

// GetHandlerConfirmDeposit returns a handler to confirm a deposit
func GetHandlerConfirmDeposit(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqConfirmDeposit
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

		txID := common.HexToHash(req.TxID)
		burnerAddr := common.HexToAddress(req.BurnerAddress)

		msg := types.NewConfirmDepositRequest(fromAddr, mux.Vars(r)[clientUtils.PathVarChain], txID, burnerAddr)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

// GetHandlerConfirmTransferKey returns a handler to confirm a transfer operatorship
func GetHandlerConfirmTransferKey(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqConfirmTransferKey
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

		txID := common.HexToHash(req.TxID)

		msg := types.NewConfirmTransferKeyRequest(fromAddr, mux.Vars(r)[clientUtils.PathVarChain], txID)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

// GetHandlerCreatePendingTransfers returns a handler to create commands for all pending transfers
func GetHandlerCreatePendingTransfers(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqCreatePendingTransfers
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

		msg := types.NewCreatePendingTransfersRequest(fromAddr, mux.Vars(r)[clientUtils.PathVarChain])
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

// GetHandlerCreateDeployToken returns a handler to create a deploy token command
func GetHandlerCreateDeployToken(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqCreateDeployToken
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

		decs, err := strconv.ParseUint(req.Decimals, 10, 8)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, errors.New("could not parse decimals").Error())
			return
		}

		capacity, ok := sdk.NewIntFromString(req.Capacity)
		if !ok {
			rest.WriteErrorResponse(w, http.StatusBadRequest, errors.New("could not parse capacity").Error())
			return
		}

		if req.Address == "" {
			req.Address = types.ZeroAddress.Hex()
		}

		if !common.IsHexAddress(req.Address) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, errors.New("could not parse address").Error())
			return
		}

		asset := types.NewAsset(req.OriginChain, req.OriginAsset)
		tokenDetails := types.NewTokenDetails(req.TokenName, req.Symbol, uint8(decs), capacity)
		msg := types.NewCreateDeployTokenRequest(fromAddr, mux.Vars(r)[clientUtils.PathVarChain], asset, tokenDetails, types.Address(common.HexToAddress(req.Address)), req.DailyMintLimit)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

// GetHandlerCreateBurnTokens returns a handler to create commands for all outstanding burns
func GetHandlerCreateBurnTokens(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqCreateBurnTokens
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

		msg := types.NewCreateBurnTokensRequest(fromAddr, mux.Vars(r)[clientUtils.PathVarChain])
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

// GetHandlerCreateTransferOperatorship returns a handler to create transfer operatoship command
func GetHandlerCreateTransferOperatorship(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqCreateTransferOperatorship
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
		msg := types.NewCreateTransferOperatorshipRequest(fromAddr, mux.Vars(r)[clientUtils.PathVarChain], req.KeyID)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

// GetHandlerSignCommands returns a handler to sign pending commands
func GetHandlerSignCommands(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqSignCommands
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
		msg := types.NewSignCommandsRequest(fromAddr, mux.Vars(r)[clientUtils.PathVarChain])
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

// GetHandlerAddChain returns a handler to add a new evm chain command
func GetHandlerAddChain(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqAddChain
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

		msg := types.NewAddChainRequest(fromAddr, req.Name, req.Params)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}
