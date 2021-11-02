package rest

import (
	"encoding/hex"
	clientUtils "github.com/axelarnetwork/axelar-core/utils"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"
	"net/http"

	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
)

// rest routes
const (
	TxLink                    = "link"
	TxConfirmDeposit          = "confirm-deposit"
	TxExecutePendingTransfers = "execute-pending"
	TxRegisterIBCPath         = "register-path"
	TxAddCosmosBasedChain     = "add-cosmos-based-chain"
	TxRegisterAsset           = "register-asset"
	TXRegisterFeeCollector    = "register-fee-collector"
)

// ReqLink represents a request to link a cross-chain address to an EVM chain address
type ReqLink struct {
	BaseReq        rest.BaseReq `json:"base_req" yaml:"base_req"`
	RecipientChain string       `json:"chain" yaml:"chain"`
	RecipientAddr  string       `json:"recipient" yaml:"recipient"`
	Asset          string       `json:"asset" yaml:"asset"`
}

// ReqConfirmDeposit represents a request to confirm a deposit
type ReqConfirmDeposit struct {
	BaseReq        rest.BaseReq `json:"base_req" yaml:"base_req"`
	TxID           string       `json:"tx_id" yaml:"tx_id"`
	Amount         string       `json:"amount" yaml:"amount"`
	DepositAddress string       `json:"deposit_address" yaml:"deposit_address"`
}

// ReqExecutePendingTransfers represents a request to execute pending token transfers
type ReqExecutePendingTransfers struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
}

// ReqRegisterIBCPath represents a request to register an IBC tracing path for a cosmos chain
type ReqRegisterIBCPath struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	Chain   string       `json:"chain" yaml:"chain"`
	Path    string       `json:"path" yaml:"path"`
}

// ReqAddCosmosBasedChain represents a request to add a cosmos based chain to nexus
type ReqAddCosmosBasedChain struct {
	BaseReq     rest.BaseReq `json:"base_req" yaml:"base_req"`
	Name        string       `json:"name" yaml:"name"`
	NativeAsset string       `json:"native_asset" yaml:"native_asset"`
}

// ReqRegisterAsset represents a request to register an asset to a cosmos based chain
type ReqRegisterAsset struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	Chain   string       `json:"chain" yaml:"chain"`
	Denom   string       `json:"denom" yaml:"denom"`
}

// ReqRegisterFeeCollector represents a request to register axelarnet fee collector account
type ReqRegisterFeeCollector struct {
	BaseReq      rest.BaseReq `json:"base_req" yaml:"base_req"`
	FeeCollector string       `json:"fee_collector" yaml:"fee_collector"`
}

// RegisterRoutes registers this module's REST routes with the given router
func RegisterRoutes(cliCtx client.Context, r *mux.Router) {
	registerTx := clientUtils.RegisterTxHandlerFn(r, types.RestRoute)
	registerTx(TxHandlerLink(cliCtx), TxLink, clientUtils.PathVarChain)
	registerTx(TxHandlerConfirmDeposit(cliCtx), TxConfirmDeposit)
	registerTx(TxHandlerExecutePendingTransfers(cliCtx), TxExecutePendingTransfers)
	registerTx(TxHandlerRegisterIBCPath(cliCtx), TxRegisterIBCPath)
	registerTx(TxHandlerAddCosmosBasedChain(cliCtx), TxAddCosmosBasedChain)
	registerTx(TxHandlerRegisterAsset(cliCtx), TxRegisterAsset)
	registerTx(TxHandlerRegisterFeeCollector(cliCtx), TXRegisterFeeCollector)
}

// TxHandlerLink returns the handler to link an Axelar address to a cross-chain address
func TxHandlerLink(cliCtx client.Context) http.HandlerFunc {
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

		msg := types.NewLinkRequest(fromAddr, req.RecipientChain, req.RecipientAddr, req.Asset)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

// TxHandlerConfirmDeposit returns the handler to confirm a deposit
func TxHandlerConfirmDeposit(cliCtx client.Context) http.HandlerFunc {
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

		txID, err := hex.DecodeString(req.TxID)

		coin, err := sdk.ParseCoinNormalized(req.Amount)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		depositAddr, err := sdk.AccAddressFromBech32(req.DepositAddress)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		msg := types.NewConfirmDepositRequest(fromAddr, txID, coin, depositAddr)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

// TxHandlerExecutePendingTransfers returns the handler to execute pending transfers to Axelar
func TxHandlerExecutePendingTransfers(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqExecutePendingTransfers
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

		msg := types.NewExecutePendingTransfersRequest(fromAddr)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

// TxHandlerRegisterIBCPath returns the handler to register an IBC tracing path for a cosmos chain
func TxHandlerRegisterIBCPath(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqRegisterIBCPath
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

		msg := types.NewRegisterIBCPathRequest(fromAddr, req.Chain, req.Path)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

// TxHandlerAddCosmosBasedChain returns the handler to add a cosmos based chain to nexus
func TxHandlerAddCosmosBasedChain(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqAddCosmosBasedChain
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

		msg := types.NewAddCosmosBasedChainRequest(fromAddr, req.Name, req.NativeAsset)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

// TxHandlerRegisterAsset returns the handler to register an asset to a cosmos based chain
func TxHandlerRegisterAsset(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqRegisterAsset
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

		msg := types.NewRegisterAssetRequest(fromAddr, req.Chain, req.Denom)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}

// TxHandlerRegisterFeeCollector returns the handler to register fee collector account
func TxHandlerRegisterFeeCollector(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReqRegisterFeeCollector
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

		feeCollector, err := sdk.AccAddressFromBech32(req.FeeCollector)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		msg := types.NewRegisterFeeCollectorRequest(fromAddr, feeCollector)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		tx.WriteGeneratedTxResponse(cliCtx, w, req.BaseReq, msg)
	}
}
