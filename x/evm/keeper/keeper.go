package keeper

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	params "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tendermint/tendermint/libs/log"

	evmTypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/cosmos/cosmos-sdk/store/prefix"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

var (
	gatewayKey                       = utils.KeyFromStr("gateway")
	pendingChainKey                  = utils.KeyFromStr("pending_chain_asset")
	unsignedBatchedCommandsKey       = utils.KeyFromStr("unsigned_batched_commands")
	latestSignedBatchedCommandsIDKey = utils.KeyFromStr("latest_signed_batched_commands_id")

	chainPrefix                 = utils.KeyFromStr("chain_")
	subspacePrefix              = utils.KeyFromStr("subspace_")
	unsignedPrefix              = utils.KeyFromStr("unsigned_")
	tokenMetadataPrefix         = utils.KeyFromStr("token_deployment_")
	pendingDepositPrefix        = utils.KeyFromStr("pending_deposit_")
	confirmedDepositPrefix      = utils.KeyFromStr("confirmed_deposit_")
	burnedDepositPrefix         = utils.KeyFromStr("burned_deposit_")
	commandPrefix               = utils.KeyFromStr("command_")
	burnerAddrPrefix            = utils.KeyFromStr("burnerAddr_")
	pendingTransferKeyPrefix    = utils.KeyFromStr("pending_transfer_key_")
	archivedTransferKeyPrefix   = utils.KeyFromStr("archived_transfer_key_")
	signedBatchedCommandsPrefix = utils.KeyFromStr("signed_batched_commands_")

	commandQueueName = "command_queue"
)

var _ types.BaseKeeper = keeper{}
var _ types.ChainKeeper = keeper{}

// Keeper implements both the base keeper and chain keeper
type keeper struct {
	chain        string
	storeKey     sdk.StoreKey
	cdc          codec.BinaryCodec
	paramsKeeper types.ParamsKeeper
	subspaces    map[string]params.Subspace
}

// NewKeeper returns a new EVM base keeper
func NewKeeper(cdc codec.BinaryCodec, storeKey sdk.StoreKey, paramsKeeper types.ParamsKeeper) types.BaseKeeper {
	return keeper{
		chain:        "",
		cdc:          cdc,
		storeKey:     storeKey,
		paramsKeeper: paramsKeeper,
		subspaces:    make(map[string]params.Subspace),
	}
}

// Logger returns a module-specific logger.
func (k keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// ForChain returns the keeper associated to the given chain
func (k keeper) ForChain(ctx sdk.Context, chain string) types.ChainKeeper {
	k.chain = strings.ToLower(chain)
	return k
}

// SetPendingChain stores the chain pending for confirmation
func (k keeper) SetPendingChain(ctx sdk.Context, chain nexus.Chain) {
	k.getStore(ctx, chain.Name).Set(pendingChainKey, &chain)
}

// GetPendingChain returns the chain object with the given name, false if the chain is either unknown or confirmed
func (k keeper) GetPendingChain(ctx sdk.Context, chainName string) (nexus.Chain, bool) {
	var chain nexus.Chain
	found := k.getStore(ctx, chainName).Get(pendingChainKey, &chain)

	return chain, found
}

// DeletePendingChain deletes a chain that is not registered yet
func (k keeper) DeletePendingChain(ctx sdk.Context, chain string) {
	k.getStore(ctx, chain).Delete(pendingChainKey)
}

// SetParams sets the evm module's parameters
func (k keeper) SetParams(ctx sdk.Context, params ...types.Params) {
	for _, p := range params {
		chain := strings.ToLower(p.Chain)

		// set the chain before calling the subspace so it is recognized as an existing chain
		ctx.KVStore(k.storeKey).Set(subspacePrefix.AppendStr(chain).AsKey(), []byte(p.Chain))
		subspace, _ := k.getSubspace(ctx, chain)
		subspace.SetParamSet(ctx, &p)
	}
}

// GetParams gets the evm module's parameters
func (k keeper) GetParams(ctx sdk.Context) []types.Params {
	params := make([]types.Params, 0)
	iter := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), subspacePrefix.AsKey())
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		chain := string(iter.Value())
		subspace, _ := k.getSubspace(ctx, chain)

		var p types.Params
		subspace.GetParamSet(ctx, &p)
		params = append(params, p)
	}

	return params
}

// GetName returns the chain name
func (k keeper) GetName() string {
	return k.chain
}

// GetCommandsGasLimit returns the EVM network's gas limist for batched commands
func (k keeper) GetCommandsGasLimit(ctx sdk.Context) (uint32, bool) {
	var commandsGasLimit uint32
	subspace, ok := k.getSubspace(ctx, k.chain)
	if !ok {
		return 0, false
	}

	subspace.Get(ctx, types.KeyCommandsGasLimit, &commandsGasLimit)

	return commandsGasLimit, true
}

// GetNetwork returns the EVM network Axelar-Core is expected to connect to
func (k keeper) GetNetwork(ctx sdk.Context) (string, bool) {
	var network string
	subspace, ok := k.getSubspace(ctx, k.chain)
	if !ok {
		return network, false
	}

	subspace.Get(ctx, types.KeyNetwork, &network)
	return network, true
}

// GetRequiredConfirmationHeight returns the required block confirmation height
func (k keeper) GetRequiredConfirmationHeight(ctx sdk.Context) (uint64, bool) {
	var h uint64

	subspace, ok := k.getSubspace(ctx, k.chain)
	if !ok {
		return h, false
	}

	subspace.Get(ctx, types.KeyConfirmationHeight, &h)
	return h, true
}

// GetRevoteLockingPeriod returns the lock period for revoting
func (k keeper) GetRevoteLockingPeriod(ctx sdk.Context) (int64, bool) {
	var result int64

	subspace, ok := k.getSubspace(ctx, k.chain)
	if !ok {
		return result, false
	}

	subspace.Get(ctx, types.KeyRevoteLockingPeriod, &result)
	return result, true
}

// GetVotingThreshold returns voting threshold
func (k keeper) GetVotingThreshold(ctx sdk.Context) (utils.Threshold, bool) {
	var threshold utils.Threshold

	subspace, ok := k.getSubspace(ctx, k.chain)
	if !ok {
		return threshold, false
	}

	subspace.Get(ctx, types.KeyVotingThreshold, &threshold)
	return threshold, true
}

// GetMinVoterCount returns minimum voter count for voting
func (k keeper) GetMinVoterCount(ctx sdk.Context) (int64, bool) {
	var minVoterCount int64

	subspace, ok := k.getSubspace(ctx, k.chain)
	if !ok {
		return minVoterCount, false
	}

	subspace.Get(ctx, types.KeyMinVoterCount, &minVoterCount)
	return minVoterCount, true
}

// SetGatewayAddress sets the contract address for Axelar Gateway
func (k keeper) SetGatewayAddress(ctx sdk.Context, addr common.Address) {
	k.getStore(ctx, k.chain).SetRaw(gatewayKey, addr.Bytes())
}

// GetGatewayAddress gets the contract address for Axelar Gateway
func (k keeper) GetGatewayAddress(ctx sdk.Context) (common.Address, bool) {
	bz := k.getStore(ctx, k.chain).GetRaw(gatewayKey)
	if bz == nil {
		return common.Address{}, false
	}
	return common.BytesToAddress(bz), true
}

// SetBurnerInfo saves the burner info for a given address
func (k keeper) SetBurnerInfo(ctx sdk.Context, burnerAddr common.Address, burnerInfo *types.BurnerInfo) {
	key := burnerAddrPrefix.AppendStr(burnerAddr.Hex())
	k.getStore(ctx, k.chain).Set(key, burnerInfo)
}

// GetBurnerInfo retrieves the burner info for a given address
func (k keeper) GetBurnerInfo(ctx sdk.Context, burnerAddr common.Address) *types.BurnerInfo {
	key := burnerAddrPrefix.AppendStr(burnerAddr.Hex())
	var result types.BurnerInfo
	if !k.getStore(ctx, k.chain).Get(key, &result) {
		return nil
	}

	return &result
}

// calculates the token address for some asset with the provided axelar gateway address
func (k keeper) getTokenAddress(ctx sdk.Context, assetName string, details types.TokenDetails, gatewayAddr common.Address) (common.Address, error) {
	assetName = strings.ToLower(assetName)

	var saltToken [32]byte
	copy(saltToken[:], crypto.Keccak256Hash([]byte(details.Symbol)).Bytes())

	uint8Type, err := abi.NewType("uint8", "uint8", nil)
	if err != nil {
		return common.Address{}, err
	}

	uint256Type, err := abi.NewType("uint256", "uint256", nil)
	if err != nil {
		return common.Address{}, err
	}

	stringType, err := abi.NewType("string", "string", nil)
	if err != nil {
		return common.Address{}, err
	}

	arguments := abi.Arguments{{Type: stringType}, {Type: stringType}, {Type: uint8Type}, {Type: uint256Type}}
	packed, err := arguments.Pack(details.TokenName, details.Symbol, details.Decimals, details.Capacity.BigInt())
	if err != nil {
		return common.Address{}, err
	}

	bytecodes, ok := k.GetTokenByteCodes(ctx)
	if !ok {
		return common.Address{}, fmt.Errorf("bytecodes for token contract not found")
	}

	tokenInitCode := append(bytecodes, packed...)
	tokenInitCodeHash := crypto.Keccak256Hash(tokenInitCode)

	tokenAddr := crypto.CreateAddress2(gatewayAddr, saltToken, tokenInitCodeHash.Bytes())
	return tokenAddr, nil
}

// GetBurnerAddressAndSalt calculates a burner address and the corresponding salt for the given token address, recipient and axelar gateway address
func (k keeper) GetBurnerAddressAndSalt(ctx sdk.Context, tokenAddr types.Address, recipient string, gatewayAddr common.Address) (common.Address, common.Hash, error) {
	addressType, err := abi.NewType("address", "address", nil)
	if err != nil {
		return common.Address{}, common.Hash{}, err
	}

	bytes32Type, err := abi.NewType("bytes32", "bytes32", nil)
	if err != nil {
		return common.Address{}, common.Hash{}, err
	}

	saltBurn := common.BytesToHash(crypto.Keccak256Hash([]byte(recipient)).Bytes())

	arguments := abi.Arguments{{Type: addressType}, {Type: bytes32Type}}
	packed, err := arguments.Pack(tokenAddr, saltBurn)
	if err != nil {
		return common.Address{}, common.Hash{}, err
	}

	bytecodes, ok := k.GetBurnerByteCodes(ctx)
	if !ok {
		return common.Address{}, common.Hash{}, fmt.Errorf("bytecodes for burner address no found")
	}

	burnerInitCode := append(bytecodes, packed...)
	burnerInitCodeHash := crypto.Keccak256Hash(burnerInitCode)

	return crypto.CreateAddress2(gatewayAddr, saltBurn, burnerInitCodeHash.Bytes()), saltBurn, nil
}

// GetBurnerByteCodes returns the bytecodes for the burner contract
func (k keeper) GetBurnerByteCodes(ctx sdk.Context) ([]byte, bool) {
	var b []byte
	subspace, ok := k.getSubspace(ctx, k.chain)
	if !ok {
		return nil, false
	}
	subspace.Get(ctx, types.KeyBurnable, &b)
	return b, true
}

// GetTokenByteCodes returns the bytecodes for the token contract
func (k keeper) GetTokenByteCodes(ctx sdk.Context) ([]byte, bool) {
	var b []byte
	subspace, ok := k.getSubspace(ctx, k.chain)
	if !ok {
		return nil, false
	}
	subspace.Get(ctx, types.KeyToken, &b)
	return b, ok
}

// GetGatewayByteCodes retrieves the byte codes for the Axelar Gateway smart contract
func (k keeper) GetGatewayByteCodes(ctx sdk.Context) ([]byte, bool) {
	var b []byte
	subspace, ok := k.getSubspace(ctx, k.chain)
	if !ok {
		return b, false
	}

	subspace.Get(ctx, types.KeyGateway, &b)
	return b, true
}

func (k keeper) CreateERC20Token(ctx sdk.Context, asset string, details types.TokenDetails) (types.ERC20Token, error) {
	metadata, err := k.initTokenMetadata(ctx, asset, details)
	if err != nil {
		return types.NilToken, err
	}

	return types.CreateERC20Token(func(m types.ERC20TokenMetadata) {
		k.setTokenMetadata(ctx, asset, m)
	}, metadata), nil
}

func (k keeper) GetERC20Token(ctx sdk.Context, asset string) types.ERC20Token {
	metadata, ok := k.getTokenMetadata(ctx, asset)
	if !ok {
		return types.NilToken
	}

	return types.CreateERC20Token(func(m types.ERC20TokenMetadata) {
		k.setTokenMetadata(ctx, asset, m)
	}, metadata)
}

// SetCommand stores the given command; note that overwriting is not allowed
func (k keeper) SetCommand(ctx sdk.Context, command types.Command) error {
	key := commandPrefix.AppendStr(command.ID.Hex())
	if k.getStore(ctx, k.chain).Has(key) {
		return fmt.Errorf("command %s already exists", command.ID.Hex())
	}

	k.GetCommandQueue(ctx).Enqueue(key, &command)
	return nil
}

// GetCommand retrieves the command for the given ID
func (k keeper) GetCommand(ctx sdk.Context, commandID types.CommandID) *types.Command {
	var command types.Command
	if !k.getStore(ctx, k.chain).Get(commandPrefix.AppendStr(commandID.Hex()), &command) {
		return nil
	}

	return &command
}

func (k keeper) GetUnsignedTx(ctx sdk.Context, txID string) *evmTypes.Transaction {
	bz := k.getStore(ctx, k.chain).GetRaw(unsignedPrefix.AppendStr(txID))
	if bz == nil {
		return nil
	}

	var tx evmTypes.Transaction
	err := tx.UnmarshalBinary(bz)
	if err != nil {
		panic(err)
	}

	return &tx
}

// SetUnsignedTx stores an unsigned transaction by hash
func (k keeper) SetUnsignedTx(ctx sdk.Context, txID string, tx *evmTypes.Transaction) {
	bz, err := tx.MarshalBinary()
	if err != nil {
		panic(err)
	}

	k.getStore(ctx, k.chain).SetRaw(unsignedPrefix.AppendStr(txID), bz)
}

// SetPendingDeposit stores a pending deposit
func (k keeper) SetPendingDeposit(ctx sdk.Context, key exported.PollKey, deposit *types.ERC20Deposit) {
	k.getStore(ctx, k.chain).Set(pendingDepositPrefix.AppendStr(key.String()), deposit)
}

// GetDeposit retrieves a confirmed/burned deposit
func (k keeper) GetDeposit(ctx sdk.Context, txID common.Hash, burnAddr common.Address) (types.ERC20Deposit, types.DepositState, bool) {
	var deposit types.ERC20Deposit

	if !k.getStore(ctx, k.chain).Get(confirmedDepositPrefix.AppendStr(txID.Hex()).AppendStr(burnAddr.Hex()), &deposit) {
		return deposit, types.CONFIRMED, true
	}
	if !k.getStore(ctx, k.chain).Get(burnedDepositPrefix.AppendStr(txID.Hex()).AppendStr(burnAddr.Hex()), &deposit) {
		return deposit, types.BURNED, true
	}

	return types.ERC20Deposit{}, 0, false
}

// GetConfirmedDeposits retrieves all the confirmed ERC20 deposits
func (k keeper) GetConfirmedDeposits(ctx sdk.Context) []types.ERC20Deposit {
	var deposits []types.ERC20Deposit
	iter := k.getStore(ctx, k.chain).Iterator(confirmedDepositPrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		bz := iter.Value()

		var deposit types.ERC20Deposit
		k.cdc.MustUnmarshalLengthPrefixed(bz, &deposit)
		deposits = append(deposits, deposit)
	}

	return deposits
}

// AssembleTx sets a signature for a previously stored raw transaction
func (k keeper) AssembleTx(ctx sdk.Context, txID string, pk ecdsa.PublicKey, sig tss.Signature) (*evmTypes.Transaction, error) {
	rawTx := k.GetUnsignedTx(ctx, txID)
	if rawTx == nil {
		return nil, fmt.Errorf("raw tx for ID %s has not been prepared yet", txID)
	}

	signer := k.getSigner(ctx)

	recoverableSig, err := types.ToSignature(sig, signer.Hash(rawTx), pk)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("could not create recoverable signature: %v", err))
	}

	return rawTx.WithSignature(signer, recoverableSig[:])
}

// GetHashToSign returns the hash to sign of a previously stored raw transaction
func (k keeper) GetHashToSign(ctx sdk.Context, txID string) (common.Hash, error) {
	rawTx := k.GetUnsignedTx(ctx, txID)
	if rawTx == nil {
		return common.Hash{}, fmt.Errorf("raw tx with id %s not found", txID)
	}
	signer := k.getSigner(ctx)
	return signer.Hash(rawTx), nil
}

func (k keeper) getSigner(ctx sdk.Context) evmTypes.EIP155Signer {
	var network string
	subspace, _ := k.getSubspace(ctx, k.chain)
	subspace.Get(ctx, types.KeyNetwork, &network)
	return evmTypes.NewEIP155Signer(k.GetChainIDByNetwork(ctx, network))
}

// DeletePendingDeposit deletes the deposit associated with the given poll
func (k keeper) DeletePendingDeposit(ctx sdk.Context, key exported.PollKey) {
	k.getStore(ctx, k.chain).Delete(pendingDepositPrefix.AppendStr(key.String()))
}

// GetPendingDeposit returns the deposit associated with the given poll
func (k keeper) GetPendingDeposit(ctx sdk.Context, key exported.PollKey) (types.ERC20Deposit, bool) {
	var deposit types.ERC20Deposit
	found := k.getStore(ctx, k.chain).Get(pendingDepositPrefix.AppendStr(key.String()), &deposit)

	return deposit, found
}

// SetDeposit stores confirmed or burned deposits
func (k keeper) SetDeposit(ctx sdk.Context, deposit types.ERC20Deposit, state types.DepositState) {
	switch state {
	case types.CONFIRMED:
		k.getStore(ctx, k.chain).Set(confirmedDepositPrefix.AppendStr(deposit.TxID.Hex()).AppendStr(deposit.BurnerAddress.Hex()), &deposit)
	case types.BURNED:
		k.getStore(ctx, k.chain).Set(burnedDepositPrefix.AppendStr(deposit.TxID.Hex()).AppendStr(deposit.BurnerAddress.Hex()), &deposit)
	default:
		panic("invalid deposit state")
	}
}

// DeleteDeposit deletes the given deposit
func (k keeper) DeleteDeposit(ctx sdk.Context, deposit types.ERC20Deposit) {
	k.getStore(ctx, k.chain).Delete(confirmedDepositPrefix.AppendStr(deposit.TxID.Hex()).AppendStr(deposit.BurnerAddress.Hex()))
	k.getStore(ctx, k.chain).Delete(burnedDepositPrefix.AppendStr(deposit.TxID.Hex()).AppendStr(deposit.BurnerAddress.Hex()))
}

// SetPendingTransferKey stores a pending transfer ownership/operatorship
func (k keeper) SetPendingTransferKey(ctx sdk.Context, key exported.PollKey, transferKey *types.TransferKey) {
	k.getStore(ctx, k.chain).Set(pendingTransferKeyPrefix.AppendStr(key.String()), transferKey)
}

// DeletePendingTransferKey deletes a pending transfer ownership/operatorship
func (k keeper) DeletePendingTransferKey(ctx sdk.Context, key exported.PollKey) {
	k.getStore(ctx, k.chain).Delete(pendingTransferKeyPrefix.AppendStr(key.String()))
}

// ArchiveTransferKey archives an ownership transfer so it is no longer pending but can still be queried
func (k keeper) ArchiveTransferKey(ctx sdk.Context, key exported.PollKey) {
	var transferKey types.TransferKey
	if !k.getStore(ctx, k.chain).Get(pendingTransferKeyPrefix.AppendStr(key.String()), &transferKey) {
		k.DeletePendingTransferKey(ctx, key)
		k.getStore(ctx, k.chain).Set(archivedTransferKeyPrefix.AppendStr(key.String()), &transferKey)
	}
}

// GetArchivedTransferKey returns an archived transfer of ownership/operatorship associated with the given poll
func (k keeper) GetArchivedTransferKey(ctx sdk.Context, key exported.PollKey) (types.TransferKey, bool) {
	var transferKey types.TransferKey
	found := k.getStore(ctx, k.chain).Get(archivedTransferKeyPrefix.AppendStr(key.String()), &transferKey)

	return transferKey, found
}

// GetPendingTransferKey returns the transfer ownership/operatorship associated with the given poll
func (k keeper) GetPendingTransferKey(ctx sdk.Context, key exported.PollKey) (types.TransferKey, bool) {
	var transferKey types.TransferKey
	found := k.getStore(ctx, k.chain).Get(pendingTransferKeyPrefix.AppendStr(key.String()), &transferKey)

	return transferKey, found
}

// GetNetworkByID returns the network name for a given chain and network ID
func (k keeper) GetNetworkByID(ctx sdk.Context, id *big.Int) (string, bool) {
	if id == nil {
		return "", false
	}
	subspace, ok := k.getSubspace(ctx, k.chain)
	if !ok {
		return "", false
	}

	var p types.Params
	subspace.GetParamSet(ctx, &p)
	for _, n := range p.Networks {
		if n.Id.BigInt().Cmp(id) == 0 {
			return n.Name, true
		}
	}

	return "", false
}

// GetChainIDByNetwork returns the network name for a given chain and network name
func (k keeper) GetChainIDByNetwork(ctx sdk.Context, network string) *big.Int {
	if network == "" {
		return nil
	}
	subspace, ok := k.getSubspace(ctx, k.chain)
	if !ok {
		return nil
	}

	var p types.Params
	subspace.GetParamSet(ctx, &p)
	for _, n := range p.Networks {
		if n.Name == network {
			return n.Id.BigInt()
		}
	}

	return nil
}

// GetCommandQueue returns the queue of commands
func (k keeper) GetCommandQueue(ctx sdk.Context) utils.KVQueue {
	return utils.NewBlockHeightKVQueue(commandQueueName, k.getStore(ctx, k.chain), ctx.BlockHeight(), k.Logger(ctx))
}

// SetUnsignedBatchedCommands stores the given unsigned batched commands
func (k keeper) SetUnsignedBatchedCommands(ctx sdk.Context, batchedCommands types.BatchedCommands) {
	k.getStore(ctx, k.chain).Set(unsignedBatchedCommandsKey, &batchedCommands)
}

// GetUnsignedBatchedCommands retrieves the unsigned batched commands
func (k keeper) GetUnsignedBatchedCommands(ctx sdk.Context) (types.BatchedCommands, bool) {
	var batchedCommands types.BatchedCommands
	found := k.getStore(ctx, k.chain).Get(unsignedBatchedCommandsKey, &batchedCommands)

	return batchedCommands, found
}

// DeleteUnsignedBatchedCommands deletes the unsigned batched commands
func (k keeper) DeleteUnsignedBatchedCommands(ctx sdk.Context) {
	k.getStore(ctx, k.chain).Delete(unsignedBatchedCommandsKey)
}

// SetSignedBatchedCommands stores the signed batched commands
func (k keeper) SetSignedBatchedCommands(ctx sdk.Context, batchedCommands types.BatchedCommands) {
	batchedCommands.Status = types.Signed
	key := signedBatchedCommandsPrefix.AppendStr(hex.EncodeToString(batchedCommands.ID))

	k.getStore(ctx, k.chain).Set(key, &batchedCommands)
}

// GetSignedBatchedCommands retrieves the signed batched commands of given ID
func (k keeper) GetSignedBatchedCommands(ctx sdk.Context, id []byte) (types.BatchedCommands, bool) {
	key := signedBatchedCommandsPrefix.AppendStr(hex.EncodeToString(id))
	var batchedCommands types.BatchedCommands
	found := k.getStore(ctx, k.chain).Get(key, &batchedCommands)

	return batchedCommands, found
}

// SetLatestSignedBatchedCommandsID stores the ID of the latest signed batched commands
func (k keeper) SetLatestSignedBatchedCommandsID(ctx sdk.Context, id []byte) {
	k.getStore(ctx, k.chain).SetRaw(latestSignedBatchedCommandsIDKey, id)
}

// GetLatestSignedBatchedCommandsID retrieves the ID of the latest signed batched commands
func (k keeper) GetLatestSignedBatchedCommandsID(ctx sdk.Context) ([]byte, bool) {
	bz := k.getStore(ctx, k.chain).GetRaw(latestSignedBatchedCommandsIDKey)

	return bz, bz != nil
}

func (k keeper) getStore(ctx sdk.Context, chain string) utils.KVStore {
	pre := string(chainPrefix.Append(utils.LowerCaseKey(chain)).AsKey()) + "_"
	return utils.NewNormalizedStore(prefix.NewStore(ctx.KVStore(k.storeKey), []byte(pre)), k.cdc)
}

func (k keeper) getSubspace(ctx sdk.Context, chain string) (params.Subspace, bool) {
	chainLower := strings.ToLower(chain)

	// When a node restarts or joins the network after genesis, it might not have all EVM subspaces initialized.
	// The following checks has to be done regardless, if we would only do it dependent on the existence of a subspace
	// different nodes would consume different amounts of gas and it would result in a consensus failure
	if !ctx.KVStore(k.storeKey).Has(subspacePrefix.AppendStr(chainLower).AsKey()) {
		return params.Subspace{}, false
	}

	chainKey := types.ModuleName + "_" + chainLower
	subspace, ok := k.subspaces[chainKey]
	if !ok {
		subspace = k.paramsKeeper.Subspace(chainKey).WithKeyTable(types.KeyTable())
		k.subspaces[chainKey] = subspace
	}
	return subspace, true
}

func (k keeper) setTokenMetadata(ctx sdk.Context, asset string, meta types.ERC20TokenMetadata) {
	key := tokenMetadataPrefix.Append(utils.LowerCaseKey(asset))
	k.getStore(ctx, k.chain).Set(key, &meta)
}

func (k keeper) getTokenMetadata(ctx sdk.Context, asset string) (types.ERC20TokenMetadata, bool) {
	var result types.ERC20TokenMetadata
	key := tokenMetadataPrefix.Append(utils.LowerCaseKey(asset))
	found := k.getStore(ctx, k.chain).Get(key, &result)

	return result, found
}

func (k keeper) initTokenMetadata(ctx sdk.Context, asset string, details types.TokenDetails) (types.ERC20TokenMetadata, error) {
	// perform a few checks now, so that it is impossible to get errors later
	if token := k.GetERC20Token(ctx, asset); !token.Is(types.NonExistent) {
		return types.ERC20TokenMetadata{}, fmt.Errorf("token '%s' already set", asset)
	}

	gatewayAddr, found := k.GetGatewayAddress(ctx)
	if !found {
		return types.ERC20TokenMetadata{}, fmt.Errorf("axelar gateway address for chain '%s' not set", k.chain)
	}

	_, found = k.GetTokenByteCodes(ctx)
	if !found {
		return types.ERC20TokenMetadata{}, fmt.Errorf("bytecodes for token contract for chain '%s' not found", k.chain)
	}

	if err := details.Validate(); err != nil {
		return types.ERC20TokenMetadata{}, err
	}

	var network string
	subspace, ok := k.getSubspace(ctx, k.chain)
	if !ok {
		return types.ERC20TokenMetadata{}, fmt.Errorf("could not find subspace for chain '%s'", k.chain)
	}

	subspace.Get(ctx, types.KeyNetwork, &network)

	chainID := k.GetChainIDByNetwork(ctx, network)
	if chainID == nil {
		return types.ERC20TokenMetadata{}, fmt.Errorf("could not find chain ID for chain '%s'", k.chain)
	}

	tokenAddr, err := k.getTokenAddress(ctx, asset, details, gatewayAddr)
	if err != nil {
		return types.ERC20TokenMetadata{}, err
	}

	// all good
	meta := types.ERC20TokenMetadata{
		Asset:        asset,
		Details:      details,
		TokenAddress: types.Address(tokenAddr),
		ChainID:      sdk.NewIntFromBigInt(chainID),
		Status:       types.Initialized,
	}
	k.setTokenMetadata(ctx, asset, meta)
	return meta, nil
}
