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
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/gorilla/mux"

	clientUtils "github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/evm/keeper"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
)

// rest routes
const (
	TxConfirmChain             = "confirm-chain"
	TxLink                     = "link"
	TxConfirmTokenDeploy       = "confirm-erc20-deploy"
	TxConfirmDeposit           = "confirm-erc20-deposit"
	TxConfirmTransferOwnership = "confirm-transfer-ownership"
	TxSignTx                   = "sign-tx"
	TxSignPending              = "sign-pending"
	TxSignDeployToken          = "sign-deploy-token"
	TxSignBurnTokens           = "sign-burn"
	TxSignTransferOwnership    = "sign-transfer-ownership"
	TxAddChain                 = "add-chain"

	QueryMasterAddress        = keeper.QMasterAddress
	QueryNextMasterAddress    = keeper.QNextMasterAddress
	QueryKeyAddress           = keeper.QKeyAddress
	QueryAxelarGatewayAddress = keeper.QAxelarGatewayAddress
	QueryCommandData          = keeper.QCommandData
	QueryBytecode             = keeper.QBytecode
	QuerySignedTx             = keeper.QSignedTx
	QueryCreateDeployTx       = keeper.CreateDeployTx
	QuerySendTx               = keeper.SendTx
	QuerySendCommand          = keeper.SendCommand
)

// RegisterRoutes registers this module's REST routes with the given router
func RegisterRoutes(cliCtx client.Context, r *mux.Router) {
	registerTx := clientUtils.RegisterTxHandlerFn(r, types.RestRoute)
	registerTx(GetHandlerLink(cliCtx), TxLink, clientUtils.PathVarChain)
	registerTx(GetHandlerConfirmTokenDeploy(cliCtx), TxConfirmTokenDeploy, clientUtils.PathVarChain, clientUtils.PathVarSymbol)
	registerTx(GetHandlerConfirmDeposit(cliCtx), TxConfirmDeposit, clientUtils.PathVarChain)
	registerTx(GetHandlerConfirmTransferOwnership(cliCtx), TxConfirmTransferOwnership, clientUtils.PathVarChain)
	registerTx(GetHandlerSignTx(cliCtx), TxSignTx, clientUtils.PathVarChain)
	registerTx(GetHandlerSignPendingTransfers(cliCtx), TxSignPending, clientUtils.PathVarChain)
	registerTx(GetHandlerSignDeployToken(cliCtx), TxSignDeployToken, clientUtils.PathVarChain, clientUtils.PathVarSymbol)
	registerTx(GetHandlerSignBurnTokens(cliCtx), TxSignBurnTokens, clientUtils.PathVarChain)
	registerTx(GetHandlerSignTransferOwnership(cliCtx), TxSignTransferOwnership, clientUtils.PathVarChain)
	registerTx(GetHandlerConfirmChain(cliCtx), TxConfirmChain)
	registerTx(GetHandlerAddChain(cliCtx), TxAddChain)

	registerQuery := clientUtils.RegisterQueryHandlerFn(r, types.RestRoute)
	registerQuery(GetHandlerQueryMasterAddress(cliCtx), QueryMasterAddress, clientUtils.PathVarChain)
	registerQuery(GetHandlerQueryNextMasterAddress(cliCtx), QueryNextMasterAddress, clientUtils.PathVarChain)
	registerQuery(GetHandlerQueryKeyAddress(cliCtx), QueryKeyAddress, clientUtils.PathVarKeyID)
	registerQuery(GetHandlerQueryAxelarGatewayAddress(cliCtx), QueryAxelarGatewayAddress, clientUtils.PathVarChain)
	registerQuery(GetHandlerQueryCommandData(cliCtx), QueryCommandData, clientUtils.PathVarChain, clientUtils.PathVarCommandID)
	registerQuery(GetHandlerQueryBytecode(cliCtx), QueryBytecode, clientUtils.PathVarChain, clientUtils.PathVarContract)
	registerQuery(GetHandlerQuerySignedTx(cliCtx), QuerySignedTx, clientUtils.PathVarChain, clientUtils.PathVarTxID)
	registerQuery(GetHandlerQueryCreateDeployTx(cliCtx), QueryCreateDeployTx, clientUtils.PathVarChain)
	registerQuery(GetHandlerQuerySendTx(cliCtx), QuerySendTx, clientUtils.PathVarChain, clientUtils.PathVarTxID)
	registerQuery(GetHandlerQuerySendCommandTx(cliCtx), QuerySendCommand, clientUtils.PathVarChain)
}

// ReqLink represents a request to link a cross-chain address to an EVM chain address
type ReqLink struct {
	BaseReq        rest.BaseReq `json:"base_req" yaml:"base_req"`
	RecipientChain string       `json:"chain" yaml:"chain"`
	RecipientAddr  string       `json:"recipient" yaml:"recipient"`
	Symbol         string       `json:"symbol" yaml:"symbol"`
}

// ReqConfirmChain represents a request to confirm a token deployment
type ReqConfirmChain struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	Chain   string       `json:"chain" yaml:"chain"`
}

// ReqConfirmTokenDeploy represents a request to confirm a token deployment
type ReqConfirmTokenDeploy struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	TxID    string       `json:"tx_id" yaml:"tx_id"`
}

// ReqConfirmDeposit represents a request to confirm a deposit
type ReqConfirmDeposit struct {
	BaseReq       rest.BaseReq `json:"base_req" yaml:"base_req"`
	TxID          string       `json:"tx_id" yaml:"tx_id"`
	Amount        string       `json:"amount" yaml:"amount"`
	BurnerAddress string       `json:"burner_address" yaml:"burner_address"`
}

// ReqConfirmTransferOwnership represents a request to confirm a transfer ownership
type ReqConfirmTransferOwnership struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	TxID    string       `json:"tx_id" yaml:"tx_id"`
	KeyID   string       `json:"key_id" yaml:"key_id"`
}

// ReqSignTx represents a request to sign a transaction
type ReqSignTx struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	TxJSON  string       `json:"tx_json" yaml:"tx_json"`
}

// ReqSignPendingTransfers represents a request to sign all pending transfers
type ReqSignPendingTransfers struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
}

// ReqSignDeployToken represents a request to sign a deploy token command
type ReqSignDeployToken struct {
	BaseReq  rest.BaseReq `json:"base_req" yaml:"base_req"`
	Name     string       `json:"name" yaml:"name"`
	Decimals string       `json:"decimals" yaml:"decimals"`
	Capacity string       `json:"capacity" yaml:"capacity"`
}

// ReqSignBurnTokens represents a request to sign all outstanding burn commands
type ReqSignBurnTokens struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
}

// ReqSignTransferOwnership represents a request to sign transfer ownership command
type ReqSignTransferOwnership struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	KeyID   string       `json:"key_id" yaml:"key_id"`
}

// ReqAddChain represents a request to add a new evm chain command
type ReqAddChain struct {
	BaseReq         rest.BaseReq       `json:"base_req" yaml:"base_req"`
	Name            string             `json:"name" yaml:"name"`
	NativeAsset     string             `json:"native_asset" yaml:"native_asset"`
	KeyRequirements tss.KeyRequirement `json:"key_requirement" yaml:"key_requirement"`
	Params          types.Params       `json:"params" yaml:"params"`
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
			Chain:          mux.Vars(r)[clientUtils.PathVarChain],
			Sender:         fromAddr,
			RecipientChain: req.RecipientChain,
			RecipientAddr:  req.RecipientAddr,
			Symbol:         req.Symbol,
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
		msg := types.NewConfirmTokenRequest(fromAddr, mux.Vars(r)[clientUtils.PathVarChain], mux.Vars(r)[clientUtils.PathVarSymbol], txID)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

// GetHandlerConfirmChain returns a handler to confirm an EVM chain
func GetHandlerConfirmChain(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqConfirmChain
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

		msg := types.NewConfirmChainRequest(fromAddr, req.Chain)
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
		amount := sdk.NewUintFromString(req.Amount)
		burnerAddr := common.HexToAddress(req.BurnerAddress)

		msg := types.NewConfirmDepositRequest(fromAddr, mux.Vars(r)[clientUtils.PathVarChain], txID, amount, burnerAddr)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

// GetHandlerConfirmTransferOwnership returns a handler to confirm a transfer ownership
func GetHandlerConfirmTransferOwnership(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqConfirmTransferOwnership
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

		msg := types.NewConfirmTransferOwnershipRequest(fromAddr, mux.Vars(r)[clientUtils.PathVarChain], txID, req.KeyID)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

// GetHandlerSignTx returns a handler to sign a transaction
func GetHandlerSignTx(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqSignTx
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

		txJSON := []byte(req.TxJSON)
		var ethtx *ethTypes.Transaction
		err := cliCtx.LegacyAmino.UnmarshalJSON(txJSON, &ethtx)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		msg := types.NewSignTxRequest(fromAddr, mux.Vars(r)[clientUtils.PathVarChain], txJSON)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

// GetHandlerSignPendingTransfers returns a handler to sign all pending transfers
func GetHandlerSignPendingTransfers(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqSignPendingTransfers
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

		msg := types.NewSignPendingTransfersRequest(fromAddr, mux.Vars(r)[clientUtils.PathVarChain])
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

// GetHandlerSignDeployToken returns a handler to sign a deploy token command
func GetHandlerSignDeployToken(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqSignDeployToken
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

		symbol := mux.Vars(r)[clientUtils.PathVarSymbol]
		decs, err := strconv.ParseUint(req.Decimals, 10, 8)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, errors.New("could not parse decimals").Error())
		}
		capacity, ok := sdk.NewIntFromString(req.Capacity)
		if !ok {
			rest.WriteErrorResponse(w, http.StatusBadRequest, errors.New("could not parse capacity").Error())
		}

		msg := types.NewSignDeployTokenRequest(fromAddr, mux.Vars(r)[clientUtils.PathVarChain], req.Name, symbol, uint8(decs), capacity)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

// GetHandlerSignBurnTokens returns a handler to sign all outstanding burn commands
func GetHandlerSignBurnTokens(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqSignBurnTokens
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

		msg := types.NewSignBurnTokensRequest(fromAddr, mux.Vars(r)[clientUtils.PathVarChain])
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

// GetHandlerSignTransferOwnership returns a handler to sign transfer ownership command
func GetHandlerSignTransferOwnership(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqSignTransferOwnership
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
		msg := types.NewSignTransferOwnershipRequest(fromAddr, mux.Vars(r)[clientUtils.PathVarChain], req.KeyID)
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

		msg := types.NewAddChainRequest(fromAddr, req.Name, req.NativeAsset, req.KeyRequirements, req.Params)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}
