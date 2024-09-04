package keeper

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvm "github.com/CosmWasm/wasmvm"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type WasmerEngine struct {
	wasmtypes.WasmerEngine
	txIDGenerator types.TxIDGenerator
}

func NewWasmerEngine(inner wasmtypes.WasmerEngine, txIDGenerator types.TxIDGenerator) wasmtypes.WasmerEngine {
	return &WasmerEngine{WasmerEngine: inner, txIDGenerator: txIDGenerator}
}

func getCtx(querier wasmvm.Querier) sdk.Context {
	return querier.(wasmkeeper.QueryHandler).Ctx
}

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
) (*wasmvmtypes.Response, uint64, error) {
	defer w.txIDGenerator.Next(getCtx(querier))

	return w.WasmerEngine.Instantiate(checksum, env, info, initMsg, store, goapi, querier, gasMeter, gasLimit, deserCost)
}

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
) (*wasmvmtypes.Response, uint64, error) {
	defer w.txIDGenerator.Next(getCtx(querier))

	return w.WasmerEngine.Execute(code, env, info, executeMsg, store, goapi, querier, gasMeter, gasLimit, deserCost)
}

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
) (*wasmvmtypes.Response, uint64, error) {
	defer w.txIDGenerator.Next(getCtx(querier))

	return w.WasmerEngine.Migrate(checksum, env, migrateMsg, store, goapi, querier, gasMeter, gasLimit, deserCost)
}

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
) (*wasmvmtypes.Response, uint64, error) {
	defer w.txIDGenerator.Next(getCtx(querier))

	return w.WasmerEngine.Sudo(checksum, env, sudoMsg, store, goapi, querier, gasMeter, gasLimit, deserCost)
}

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
) (*wasmvmtypes.Response, uint64, error) {
	defer w.txIDGenerator.Next(getCtx(querier))

	return w.WasmerEngine.Reply(checksum, env, reply, store, goapi, querier, gasMeter, gasLimit, deserCost)
}
