package keeper

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvm "github.com/CosmWasm/wasmvm/v2"
	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/nexus/types"
)

// WasmerEngine is a wrapper around the WasmerEngine to add a transaction ID generator
type WasmerEngine struct {
	wasmtypes.WasmEngine
	msgIDGenerator types.MsgIDGenerator
}

// NewWasmerEngine wraps the given engine with a transaction ID generator
func NewWasmerEngine(inner wasmtypes.WasmEngine, msgIDGenerator types.MsgIDGenerator) wasmtypes.WasmEngine {
	return &WasmerEngine{WasmEngine: inner, msgIDGenerator: msgIDGenerator}
}

func getCtx(querier wasmvm.Querier) sdk.Context {
	// wasmd passes a reference to the querier only for the Migrate method
	// https://github.com/CosmWasm/wasmd/blob/21ec15a5c025bc0fa8c634691dc839ab77b9a7d2/x/wasm/keeper/keeper.go#L433
	if querier, ok := querier.(*wasmkeeper.QueryHandler); ok {
		return querier.Ctx
	}

	return querier.(wasmkeeper.QueryHandler).Ctx
}

// Instantiate calls the inner engine and increments the transaction ID
// func (w *WasmerEngine) Instantiate(
//
//	checksum wasmvm.Checksum,
//	env wasmvmtypes.Env,
//	info wasmvmtypes.MessageInfo,
//	initMsg []byte,
//	store wasmvm.KVStore,
//	goapi wasmvm.GoAPI,
//	querier wasmvm.Querier,
//	gasMeter wasmvm.GasMeter,
//	gasLimit uint64,
//	deserCost wasmvmtypes.UFraction,
//
// ) (*wasmvmtypes.Response, uint64, error) {
func (w *WasmerEngine) Instantiate(
	checksum wasmvm.Checksum,
	env wasmvmtypes.Env,
	info wasmvmtypes.MessageInfo,
	initMsg []byte,
	store wasmvm.KVStore,
	goapi wasmvm.GoAPI,
	querier wasmvm.Querier,
	gasMeter wasmvm.GasMeter,
	gasLimit uint64,
	deserCost wasmvmtypes.UFraction,
) (*wasmvmtypes.ContractResult, uint64, error) {
	defer w.msgIDGenerator.IncrID(getCtx(querier))

	return w.WasmEngine.Instantiate(checksum, env, info, initMsg, store, goapi, querier, gasMeter, gasLimit, deserCost)
}

// Execute calls the inner engine and increments the transaction ID
func (w *WasmerEngine) Execute(
	code wasmvm.Checksum,
	env wasmvmtypes.Env,
	info wasmvmtypes.MessageInfo,
	executeMsg []byte,
	store wasmvm.KVStore,
	goapi wasmvm.GoAPI,
	querier wasmvm.Querier,
	gasMeter wasmvm.GasMeter,
	gasLimit uint64,
	deserCost wasmvmtypes.UFraction,
) (*wasmvmtypes.ContractResult, uint64, error) {
	defer w.msgIDGenerator.IncrID(getCtx(querier))

	return w.WasmEngine.Execute(code, env, info, executeMsg, store, goapi, querier, gasMeter, gasLimit, deserCost)
}

// Migrate calls the inner engine and increments the transaction ID
func (w *WasmerEngine) Migrate(
	checksum wasmvm.Checksum,
	env wasmvmtypes.Env,
	migrateMsg []byte,
	store wasmvm.KVStore,
	goapi wasmvm.GoAPI,
	querier wasmvm.Querier,
	gasMeter wasmvm.GasMeter,
	gasLimit uint64,
	deserCost wasmvmtypes.UFraction,
) (*wasmvmtypes.ContractResult, uint64, error) {
	defer w.msgIDGenerator.IncrID(getCtx(querier))

	return w.WasmEngine.Migrate(checksum, env, migrateMsg, store, goapi, querier, gasMeter, gasLimit, deserCost)
}

// Sudo calls the inner engine and increments the transaction ID
func (w *WasmerEngine) Sudo(
	checksum wasmvm.Checksum,
	env wasmvmtypes.Env,
	sudoMsg []byte,
	store wasmvm.KVStore,
	goapi wasmvm.GoAPI,
	querier wasmvm.Querier,
	gasMeter wasmvm.GasMeter,
	gasLimit uint64,
	deserCost wasmvmtypes.UFraction,
) (*wasmvmtypes.ContractResult, uint64, error) {
	defer w.msgIDGenerator.IncrID(getCtx(querier))

	return w.WasmEngine.Sudo(checksum, env, sudoMsg, store, goapi, querier, gasMeter, gasLimit, deserCost)
}

// Reply calls the inner engine and increments the transaction ID
func (w *WasmerEngine) Reply(
	checksum wasmvm.Checksum,
	env wasmvmtypes.Env,
	reply wasmvmtypes.Reply,
	store wasmvm.KVStore,
	goapi wasmvm.GoAPI,
	querier wasmvm.Querier,
	gasMeter wasmvm.GasMeter,
	gasLimit uint64,
	deserCost wasmvmtypes.UFraction,
) (*wasmvmtypes.ContractResult, uint64, error) {
	defer w.msgIDGenerator.IncrID(getCtx(querier))

	return w.WasmEngine.Reply(checksum, env, reply, store, goapi, querier, gasMeter, gasLimit, deserCost)
}
