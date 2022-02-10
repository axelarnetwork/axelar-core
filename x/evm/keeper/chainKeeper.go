package keeper

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	evmTypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

var (
	gatewayKey             = utils.KeyFromStr("gateway")
	unsignedBatchIDKey     = utils.KeyFromStr("unsigned_command_batch_id")
	latestSignedBatchIDKey = utils.KeyFromStr("latest_signed_command_batch_id")

	tokenMetadataByAssetPrefix  = utils.KeyFromStr("token_deployment_by_asset")
	tokenMetadataBySymbolPrefix = utils.KeyFromStr("token_deployment_by_symbol")
	pendingDepositPrefix        = utils.KeyFromStr("pending_deposit")
	confirmedDepositPrefix      = utils.KeyFromStr("confirmed_deposit")
	burnedDepositPrefix         = utils.KeyFromStr("burned_deposit")
	commandBatchPrefix          = utils.KeyFromStr("batched_commands")
	commandPrefix               = utils.KeyFromStr("command")
	burnerAddrPrefix            = utils.KeyFromStr("burnerAddr")
	pendingTransferKeyPrefix    = utils.KeyFromStr("pending_transfer_key")
	archivedTransferKeyPrefix   = utils.KeyFromStr("archived_transfer_key")

	commandQueueName = "cmd_queue"
)

var _ types.ChainKeeper = chainKeeper{}

type chainKeeper struct {
	BaseKeeper
	chainLowerKey string
}

func (k chainKeeper) GetName() string {
	return k.chainLowerKey
}

// SetParams sets the evm module's parameters
func (k chainKeeper) SetParams(ctx sdk.Context, params types.Params) {
	// set the chain before calling the subspace so it is recognized as an existing chain
	k.getBaseStore(ctx).SetRaw(subspacePrefix.AppendStr(k.chainLowerKey), []byte(params.Chain))
	subspace, ok := k.getSubspace(ctx)
	if !ok {
		panic(fmt.Sprintf("param subspace for chain %s should exist", params.Chain))
	}

	subspace.SetParamSet(ctx, &params)
}

// GetParams gets the evm module's parameters
func (k chainKeeper) GetParams(ctx sdk.Context) types.Params {
	subspace, ok := k.getSubspace(ctx)
	if !ok {
		panic(fmt.Sprintf("params for chain %s not set", k.chainLowerKey))
	}

	var p types.Params
	subspace.GetParamSet(ctx, &p)
	return p
}

// returns the EVM network's gas limist for batched commands
func (k chainKeeper) getCommandsGasLimit(ctx sdk.Context) uint32 {
	var commandsGasLimit uint32
	subspace, ok := k.getSubspace(ctx)

	// the subspace must exist, if not we have a catastrophic failure
	if !ok {
		panic(fmt.Sprintf("subspace for chain '%s' not set", k.chainLowerKey))
	}

	subspace.Get(ctx, types.KeyCommandsGasLimit, &commandsGasLimit)

	return commandsGasLimit
}

func (k chainKeeper) GetChainID(ctx sdk.Context) (*big.Int, bool) {
	network, ok := k.GetNetwork(ctx)
	if !ok {
		return nil, false
	}
	return k.GetChainIDByNetwork(ctx, network), true
}

// GetNetwork returns the EVM network Axelar-Core is expected to connect to
func (k chainKeeper) GetNetwork(ctx sdk.Context) (string, bool) {
	var network string
	subspace, ok := k.getSubspace(ctx)
	if !ok {
		return network, false
	}

	subspace.Get(ctx, types.KeyNetwork, &network)
	return network, true
}

// GetRequiredConfirmationHeight returns the required block confirmation height
func (k chainKeeper) GetRequiredConfirmationHeight(ctx sdk.Context) (uint64, bool) {
	var h uint64

	subspace, ok := k.getSubspace(ctx)
	if !ok {
		return h, false
	}

	subspace.Get(ctx, types.KeyConfirmationHeight, &h)
	return h, true
}

// GetRevoteLockingPeriod returns the lock period for revoting
func (k chainKeeper) GetRevoteLockingPeriod(ctx sdk.Context) (int64, bool) {
	var result int64

	subspace, ok := k.getSubspace(ctx)
	if !ok {
		return result, false
	}

	subspace.Get(ctx, types.KeyRevoteLockingPeriod, &result)
	return result, true
}

// GetVotingThreshold returns voting threshold
func (k chainKeeper) GetVotingThreshold(ctx sdk.Context) (utils.Threshold, bool) {
	var threshold utils.Threshold

	subspace, ok := k.getSubspace(ctx)
	if !ok {
		return threshold, false
	}

	subspace.Get(ctx, types.KeyVotingThreshold, &threshold)
	return threshold, true
}

// GetMinVoterCount returns minimum voter count for voting
func (k chainKeeper) GetMinVoterCount(ctx sdk.Context) (int64, bool) {
	var minVoterCount int64

	subspace, ok := k.getSubspace(ctx)
	if !ok {
		return minVoterCount, false
	}

	subspace.Get(ctx, types.KeyMinVoterCount, &minVoterCount)
	return minVoterCount, true
}

// GetTransactionFeeRate returns the transaction fee rate for evm
func (k chainKeeper) GetTransactionFeeRate(ctx sdk.Context) (sdk.Dec, bool) {
	var feeRate sdk.Dec

	subspace, ok := k.getSubspace(ctx)
	if !ok {
		return feeRate, false
	}

	subspace.Get(ctx, types.KeyTransactionFeeRate, &feeRate)
	return feeRate, true
}

// SetBurnerInfo saves the burner info for a given address
func (k chainKeeper) SetBurnerInfo(ctx sdk.Context, burnerInfo types.BurnerInfo) {
	key := burnerAddrPrefix.AppendStr(burnerInfo.BurnerAddress.Hex())
	k.getStore(ctx, k.chainLowerKey).Set(key, &burnerInfo)
}

// GetBurnerInfo retrieves the burner info for a given address
func (k chainKeeper) GetBurnerInfo(ctx sdk.Context, burnerAddr types.Address) *types.BurnerInfo {
	key := burnerAddrPrefix.AppendStr(burnerAddr.Hex())
	var result types.BurnerInfo
	if !k.getStore(ctx, k.chainLowerKey).Get(key, &result) {
		return nil
	}

	return &result
}

func (k chainKeeper) getBurnerInfos(ctx sdk.Context) []types.BurnerInfo {
	iter := k.getStore(ctx, k.chainLowerKey).Iterator(burnerAddrPrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	var burners []types.BurnerInfo
	for ; iter.Valid(); iter.Next() {
		var burner types.BurnerInfo
		iter.UnmarshalValue(&burner)
		burners = append(burners, burner)
	}

	return burners
}

// calculates the token address for some asset with the provided axelar gateway address
func (k chainKeeper) getTokenAddress(ctx sdk.Context, details types.TokenDetails, gatewayAddr common.Address) (common.Address, error) {
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

	bytecode, ok := k.GetTokenByteCode(ctx)
	if !ok {
		return common.Address{}, fmt.Errorf("bytecode for token contract not found")
	}

	tokenInitCode := append(bytecode, packed...)
	tokenInitCodeHash := crypto.Keccak256Hash(tokenInitCode)

	tokenAddr := crypto.CreateAddress2(gatewayAddr, saltToken, tokenInitCodeHash.Bytes())
	return tokenAddr, nil
}

// GetBurnerAddressAndSalt calculates a burner address and the corresponding salt for the given token address, recipient and axelar gateway address
func (k chainKeeper) GetBurnerAddressAndSalt(ctx sdk.Context, token types.ERC20Token, recipient string, gatewayAddr common.Address) (types.Address, types.Hash, error) {
	nonce := utils.GetNonce(ctx.HeaderHash(), ctx.BlockGasMeter())
	bz := []byte(recipient)
	bz = append(bz, nonce[:]...)
	salt := types.Hash(common.BytesToHash(crypto.Keccak256Hash(bz).Bytes()))

	var initCodeHash types.Hash
	tokenBurnerCodeHash := token.GetBurnerCodeHash().Hex()
	switch tokenBurnerCodeHash {
	case types.BurnerCodeHashV1:
		addressType, err := abi.NewType("address", "address", nil)
		if err != nil {
			return types.Address{}, types.Hash{}, err
		}

		bytes32Type, err := abi.NewType("bytes32", "bytes32", nil)
		if err != nil {
			return types.Address{}, types.Hash{}, err
		}

		arguments := abi.Arguments{{Type: addressType}, {Type: bytes32Type}}
		params, err := arguments.Pack(token.GetAddress(), salt)
		if err != nil {
			return types.Address{}, types.Hash{}, err
		}

		initCodeHash = types.Hash(crypto.Keccak256Hash(append(token.GetBurnerCode(), params...)))
	case types.BurnerCodeHashV2:
		initCodeHash = token.GetBurnerCodeHash()
	default:
		return types.Address{}, types.Hash{}, fmt.Errorf("unsupported burner code with hash %s for chain %s", tokenBurnerCodeHash, k.chainLowerKey)
	}

	return types.Address(crypto.CreateAddress2(gatewayAddr, salt, initCodeHash.Bytes())), salt, nil
}

// GetBurnerByteCode returns the bytecode for the burner contract
func (k chainKeeper) GetBurnerByteCode(ctx sdk.Context) ([]byte, bool) {
	var b []byte
	subspace, ok := k.getSubspace(ctx)
	if !ok {
		return nil, false
	}
	subspace.Get(ctx, types.KeyBurnable, &b)
	return b, true
}

// GetTokenByteCode returns the bytecodes for the token contract
func (k chainKeeper) GetTokenByteCode(ctx sdk.Context) ([]byte, bool) {
	var b []byte
	subspace, ok := k.getSubspace(ctx)
	if !ok {
		return nil, false
	}
	subspace.Get(ctx, types.KeyToken, &b)
	return b, ok
}

// GetGatewayByteCode retrieves the byte codes for the Axelar Gateway smart contract
func (k chainKeeper) GetGatewayByteCode(ctx sdk.Context) ([]byte, bool) {
	var b []byte
	subspace, ok := k.getSubspace(ctx)
	if !ok {
		return b, false
	}

	subspace.Get(ctx, types.KeyGateway, &b)
	return b, true
}

func (k chainKeeper) CreateERC20Token(ctx sdk.Context, asset string, details types.TokenDetails, address types.Address) (types.ERC20Token, error) {
	metadata, err := k.initTokenMetadata(ctx, asset, details, address)
	if err != nil {
		return types.NilToken, err
	}

	return types.CreateERC20Token(func(m types.ERC20TokenMetadata) {
		k.setTokenMetadata(ctx, m)
	}, metadata), nil
}

func (k chainKeeper) GetERC20TokenByAsset(ctx sdk.Context, asset string) types.ERC20Token {
	metadata, ok := k.getTokenMetadataByAsset(ctx, asset)
	if !ok {
		return types.NilToken
	}

	return types.CreateERC20Token(func(m types.ERC20TokenMetadata) {
		k.setTokenMetadata(ctx, m)
	}, metadata)
}

func (k chainKeeper) GetERC20TokenBySymbol(ctx sdk.Context, symbol string) types.ERC20Token {
	metadata, ok := k.getTokenMetadataBySymbol(ctx, symbol)
	if !ok {
		return types.NilToken
	}

	return types.CreateERC20Token(func(m types.ERC20TokenMetadata) {
		k.setTokenMetadata(ctx, m)
	}, metadata)
}

// EnqueueCommand stores the given command; note that overwriting is not allowed
func (k chainKeeper) EnqueueCommand(ctx sdk.Context, command types.Command) error {
	key := commandPrefix.AppendStr(command.ID.Hex())
	if k.getStore(ctx, k.chainLowerKey).Has(key) {
		return fmt.Errorf("command %s already exists", command.ID.Hex())
	}

	k.getCommandQueue(ctx).Enqueue(key, &command)
	return nil
}

// GetCommand returns the command specified by the given ID
func (k chainKeeper) GetCommand(ctx sdk.Context, id types.CommandID) (types.Command, bool) {
	key := commandPrefix.AppendStr(id.Hex())
	var cmd types.Command
	found := k.getStore(ctx, k.chainLowerKey).Get(key, &cmd)

	return cmd, found
}

// GetPendingCommands returns the list of commands not yet added to any batch
func (k chainKeeper) GetPendingCommands(ctx sdk.Context) []types.Command {
	var commands []types.Command

	keys := k.getCommandQueue(ctx).Keys()
	for _, key := range keys {
		var cmd types.Command
		ok := k.getStore(ctx, k.chainLowerKey).Get(key, &cmd)
		if ok {
			commands = append(commands, cmd)
		}
	}

	return commands
}

// SetPendingDeposit stores a pending deposit
func (k chainKeeper) SetPendingDeposit(ctx sdk.Context, key exported.PollKey, deposit *types.ERC20Deposit) {
	k.getStore(ctx, k.chainLowerKey).Set(pendingDepositPrefix.AppendStr(key.String()), deposit)
}

// GetDeposit retrieves a confirmed/burned deposit
func (k chainKeeper) GetDeposit(ctx sdk.Context, txID common.Hash, burnAddr common.Address) (types.ERC20Deposit, types.DepositStatus, bool) {
	var deposit types.ERC20Deposit

	if k.getStore(ctx, k.chainLowerKey).Get(confirmedDepositPrefix.AppendStr(txID.Hex()).AppendStr(burnAddr.Hex()), &deposit) {
		return deposit, types.DepositStatus_Confirmed, true
	}
	if k.getStore(ctx, k.chainLowerKey).Get(burnedDepositPrefix.AppendStr(txID.Hex()).AppendStr(burnAddr.Hex()), &deposit) {
		return deposit, types.DepositStatus_Burned, true
	}

	return types.ERC20Deposit{}, 0, false
}

// GetConfirmedDeposits retrieves all the confirmed ERC20 deposits
func (k chainKeeper) GetConfirmedDeposits(ctx sdk.Context) []types.ERC20Deposit {
	var deposits []types.ERC20Deposit
	iter := k.getStore(ctx, k.chainLowerKey).Iterator(confirmedDepositPrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var deposit types.ERC20Deposit
		iter.UnmarshalValue(&deposit)
		deposits = append(deposits, deposit)
	}

	return deposits
}

// getBurnedDeposits retrieves all the burned ERC20 deposits
func (k chainKeeper) getBurnedDeposits(ctx sdk.Context) []types.ERC20Deposit {
	var deposits []types.ERC20Deposit
	iter := k.getStore(ctx, k.chainLowerKey).Iterator(burnedDepositPrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var deposit types.ERC20Deposit
		iter.UnmarshalValue(&deposit)
		deposits = append(deposits, deposit)
	}

	return deposits
}

func (k chainKeeper) getSigner(ctx sdk.Context) evmTypes.EIP155Signer {
	// both chain, subspace, and network must be valid if the chain keeper was instantiated,
	// so a nil value here must be a catastrophic failure

	var network string
	subspace, ok := k.getSubspace(ctx)
	if !ok {
		panic(fmt.Sprintf("could not find subspace for network '%s'", k.chainLowerKey))
	}

	subspace.Get(ctx, types.KeyNetwork, &network)
	chainID := k.GetChainIDByNetwork(ctx, network)

	if chainID == nil {
		panic(fmt.Sprintf("could not find chain ID for network '%s'", network))
	}
	return evmTypes.NewEIP155Signer(chainID)
}

// DeletePendingDeposit deletes the deposit associated with the given poll
func (k chainKeeper) DeletePendingDeposit(ctx sdk.Context, key exported.PollKey) {
	k.getStore(ctx, k.chainLowerKey).Delete(pendingDepositPrefix.AppendStr(key.String()))
}

// GetPendingDeposit returns the deposit associated with the given poll
func (k chainKeeper) GetPendingDeposit(ctx sdk.Context, key exported.PollKey) (types.ERC20Deposit, bool) {
	var deposit types.ERC20Deposit
	found := k.getStore(ctx, k.chainLowerKey).Get(pendingDepositPrefix.AppendStr(key.String()), &deposit)

	return deposit, found
}

// SetDeposit stores confirmed or burned deposits
func (k chainKeeper) SetDeposit(ctx sdk.Context, deposit types.ERC20Deposit, state types.DepositStatus) {
	switch state {
	case types.DepositStatus_Confirmed:
		k.getStore(ctx, k.chainLowerKey).Set(confirmedDepositPrefix.AppendStr(deposit.TxID.Hex()).AppendStr(deposit.BurnerAddress.Hex()), &deposit)
	case types.DepositStatus_Burned:
		k.getStore(ctx, k.chainLowerKey).Set(burnedDepositPrefix.AppendStr(deposit.TxID.Hex()).AppendStr(deposit.BurnerAddress.Hex()), &deposit)
	default:
		panic("invalid deposit state")
	}
}

// DeleteDeposit deletes the given deposit
func (k chainKeeper) DeleteDeposit(ctx sdk.Context, deposit types.ERC20Deposit) {
	k.getStore(ctx, k.chainLowerKey).Delete(confirmedDepositPrefix.AppendStr(deposit.TxID.Hex()).AppendStr(deposit.BurnerAddress.Hex()))
	k.getStore(ctx, k.chainLowerKey).Delete(burnedDepositPrefix.AppendStr(deposit.TxID.Hex()).AppendStr(deposit.BurnerAddress.Hex()))
}

// SetPendingTransferKey stores a pending transfer ownership/operatorship
func (k chainKeeper) SetPendingTransferKey(ctx sdk.Context, key exported.PollKey, transferKey *types.TransferKey) {
	k.getStore(ctx, k.chainLowerKey).Set(pendingTransferKeyPrefix.AppendStr(key.String()), transferKey)
}

// DeletePendingTransferKey deletes a pending transfer ownership/operatorship
func (k chainKeeper) DeletePendingTransferKey(ctx sdk.Context, key exported.PollKey) {
	k.getStore(ctx, k.chainLowerKey).Delete(pendingTransferKeyPrefix.AppendStr(key.String()))
}

// ArchiveTransferKey archives an ownership transfer so it is no longer pending but can still be queried
func (k chainKeeper) ArchiveTransferKey(ctx sdk.Context, key exported.PollKey) {
	var transferKey types.TransferKey
	if k.getStore(ctx, k.chainLowerKey).Get(pendingTransferKeyPrefix.AppendStr(key.String()), &transferKey) {
		k.DeletePendingTransferKey(ctx, key)
		k.getStore(ctx, k.chainLowerKey).Set(archivedTransferKeyPrefix.AppendStr(key.String()), &transferKey)
	}
}

// GetArchivedTransferKey returns an archived transfer of ownership/operatorship associated with the given poll
func (k chainKeeper) GetArchivedTransferKey(ctx sdk.Context, key exported.PollKey) (types.TransferKey, bool) {
	var transferKey types.TransferKey
	found := k.getStore(ctx, k.chainLowerKey).Get(archivedTransferKeyPrefix.AppendStr(key.String()), &transferKey)

	return transferKey, found
}

// GetPendingTransferKey returns the transfer ownership/operatorship associated with the given poll
func (k chainKeeper) GetPendingTransferKey(ctx sdk.Context, key exported.PollKey) (types.TransferKey, bool) {
	var transferKey types.TransferKey
	found := k.getStore(ctx, k.chainLowerKey).Get(pendingTransferKeyPrefix.AppendStr(key.String()), &transferKey)

	return transferKey, found
}

// GetNetworkByID returns the network name for a given chain and network ID
func (k chainKeeper) GetNetworkByID(ctx sdk.Context, id *big.Int) (string, bool) {
	if id == nil {
		return "", false
	}
	subspace, ok := k.getSubspace(ctx)
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
func (k chainKeeper) GetChainIDByNetwork(ctx sdk.Context, network string) *big.Int {
	if network == "" {
		return nil
	}
	subspace, ok := k.getSubspace(ctx)
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

func (k chainKeeper) popCommand(ctx sdk.Context, filters ...func(value codec.ProtoMarshaler) bool) (types.Command, bool) {
	var cmd types.Command
	ok := k.getCommandQueue(ctx).Dequeue(&cmd, filters...)
	return cmd, ok
}

func (k chainKeeper) setCommandBatchMetadata(ctx sdk.Context, meta types.CommandBatchMetadata) {
	k.getStore(ctx, k.chainLowerKey).Set(commandBatchPrefix.AppendStr(string(meta.ID)), &meta)
}

// GetBatchByID retrieves the specified batch if it exists
func (k chainKeeper) GetBatchByID(ctx sdk.Context, id []byte) types.CommandBatch {
	batch := k.getCommandBatchMetadata(ctx, id)

	setter := func(m types.CommandBatchMetadata) {
		k.setCommandBatchMetadata(ctx, m)
	}

	return types.NewCommandBatch(batch, setter)
}

func (k chainKeeper) getCommandBatchMetadata(ctx sdk.Context, id []byte) types.CommandBatchMetadata {
	var batch types.CommandBatchMetadata
	k.getStore(ctx, k.chainLowerKey).Get(commandBatchPrefix.AppendStr(string(id)), &batch)
	return batch
}

// GetLatestCommandBatch returns the latest batch of signed commands, if it exists
func (k chainKeeper) GetLatestCommandBatch(ctx sdk.Context) types.CommandBatch {
	if batch := k.getLatestCommandBatchMetadata(ctx); batch.Status != types.BatchNonExistent {
		setter := func(m types.CommandBatchMetadata) {
			k.setCommandBatchMetadata(ctx, m)
		}
		return types.NewCommandBatch(batch, setter)
	}

	return types.NonExistentCommand
}

func (k chainKeeper) getLatestCommandBatchMetadata(ctx sdk.Context) types.CommandBatchMetadata {
	if batch := k.getUnsignedCommandBatch(ctx); batch.Status != types.BatchNonExistent {
		return batch
	}

	if id := k.getLatestSignedCommandBatchID(ctx); id != nil {
		return k.getCommandBatchMetadata(ctx, id)
	}
	return types.CommandBatchMetadata{Status: types.BatchNonExistent}
}

func (k chainKeeper) getCommandBatchesMetadata(ctx sdk.Context) []types.CommandBatchMetadata {
	iter := k.getStore(ctx, k.chainLowerKey).Iterator(commandBatchPrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	var batches []types.CommandBatchMetadata
	for ; iter.Valid(); iter.Next() {
		var batch types.CommandBatchMetadata
		iter.UnmarshalValue(&batch)
		batches = append(batches, batch)
	}

	return batches
}

func (k chainKeeper) getLatestSignedCommandBatchID(ctx sdk.Context) []byte {
	return k.getStore(ctx, k.chainLowerKey).GetRaw(latestSignedBatchIDKey)
}

// SetLatestSignedCommandBatchID stores the latest signed command batch ID
func (k chainKeeper) SetLatestSignedCommandBatchID(ctx sdk.Context, id []byte) {
	k.getStore(ctx, k.chainLowerKey).SetRaw(latestSignedBatchIDKey, id)
}

func (k chainKeeper) setLatestBatchMetadata(ctx sdk.Context, batch types.CommandBatchMetadata) {
	switch batch.Status {
	case types.BatchNonExistent:
		return
	case types.BatchSigning, types.BatchAborted:
		k.setUnsignedCommandBatchID(ctx, batch.ID)
	case types.BatchSigned:
		k.SetLatestSignedCommandBatchID(ctx, batch.ID)
	default:
		panic(fmt.Sprintf("batch status %s is not handled", batch.Status.String()))
	}
}

// CreateNewBatchToSign creates a new batch of commands to be signed
func (k chainKeeper) CreateNewBatchToSign(ctx sdk.Context, signer types.Signer) (types.CommandBatch, error) {
	command, ok := k.popCommand(ctx)
	if !ok {
		return types.CommandBatch{}, nil
	}

	chainID := sdk.NewIntFromBigInt(k.getSigner(ctx).ChainID())
	gasLimit := k.getCommandsGasLimit(ctx)
	gasCost := uint32(command.MaxGasCost)
	keyID := command.KeyID
	filter := func(value codec.ProtoMarshaler) bool {
		cmd, ok := value.(*types.Command)
		gasCost += cmd.MaxGasCost

		return ok && cmd.KeyID == keyID && gasCost <= gasLimit
	}

	commands := []types.Command{command.Clone()}
	for {
		cmd, ok := k.popCommand(ctx, filter)
		if !ok {
			break
		}
		commands = append(commands, cmd.Clone())
	}

	keyRole := signer.GetKeyRole(ctx, keyID)
	commandBatch, err := types.NewCommandBatchMetadata(chainID.BigInt(), keyID, keyRole, commands)
	if err != nil {
		return types.CommandBatch{}, err
	}

	latest := k.GetLatestCommandBatch(ctx)
	if !latest.Is(types.BatchSigned) && !latest.Is(types.BatchNonExistent) {
		return types.CommandBatch{}, fmt.Errorf("latest command batch %s is still being processed", hex.EncodeToString(latest.GetID()))
	}

	commandBatch.PrevBatchedCommandsID = latest.GetID()
	k.setCommandBatchMetadata(ctx, commandBatch)
	k.setUnsignedCommandBatchID(ctx, commandBatch.ID)

	setter := func(m types.CommandBatchMetadata) {
		k.setCommandBatchMetadata(ctx, m)
	}
	return types.NewCommandBatch(commandBatch, setter), nil
}

// DeleteUnsignedCommandBatchID deletes the unsigned command batch ID
func (k chainKeeper) DeleteUnsignedCommandBatchID(ctx sdk.Context) {
	k.getStore(ctx, k.chainLowerKey).Delete(unsignedBatchIDKey)
}

func (k chainKeeper) getUnsignedCommandBatch(ctx sdk.Context) types.CommandBatchMetadata {
	if id := k.getStore(ctx, k.chainLowerKey).GetRaw(unsignedBatchIDKey); id != nil {
		return k.getCommandBatchMetadata(ctx, id)
	}

	return types.CommandBatchMetadata{}
}

func (k chainKeeper) setUnsignedCommandBatchID(ctx sdk.Context, id []byte) {
	k.getStore(ctx, k.chainLowerKey).SetRaw(unsignedBatchIDKey, id)
}

// returns the queue of commands
func (k chainKeeper) getCommandQueue(ctx sdk.Context) utils.GeneralKVQueue {
	return utils.NewGeneralKVQueue(
		commandQueueName,
		k.getStore(ctx, k.chainLowerKey),
		k.Logger(ctx),
		func(value codec.ProtoMarshaler) utils.Key {
			command, ok := value.(*types.Command)
			if !ok {
				panic(fmt.Errorf("unexpected type of command %T", command))
			}

			bz := make([]byte, 8)

			switch command.Command {
			case types.AxelarGatewayCommandBurnToken:
			default:
				binary.BigEndian.PutUint64(bz, uint64(ctx.BlockHeight()))
			}

			return utils.KeyFromBz(bz)
		},
	)
}

func (k chainKeeper) serializeCommandQueue(ctx sdk.Context) map[string]types.Command {
	iter := k.getStore(ctx, k.chainLowerKey).Iterator(utils.KeyFromStr(commandQueueName))
	defer utils.CloseLogError(iter, k.Logger(ctx))

	commands := make(map[string]types.Command)
	for ; iter.Valid(); iter.Next() {
		var command types.Command
		iter.UnmarshalValue(&command)
		key := string(iter.Key())
		key = strings.TrimPrefix(key, commandQueueName)
		key = strings.TrimPrefix(key, "_")
		commands[key] = command
	}

	return commands
}

func (k chainKeeper) setCommandQueue(ctx sdk.Context, queueState map[string]types.Command) {
	state := make(map[string]codec.ProtoMarshaler, len(queueState))
	for key, value := range queueState {
		// need to create a new variable inside the loop because the state map takes its reference,
		// otherwise all entries would refer to a single &value pointer
		v := value
		state[key] = &v
	}

	k.getCommandQueue(ctx).ImportState(state, types.ValidateCommandQueueState)
}

func (k chainKeeper) setTokenMetadata(ctx sdk.Context, meta types.ERC20TokenMetadata) {
	// lookup by asset
	key := tokenMetadataByAssetPrefix.Append(utils.LowerCaseKey(meta.Asset))
	k.getStore(ctx, k.chainLowerKey).Set(key, &meta)

	// lookup by symbol
	key = tokenMetadataBySymbolPrefix.Append(utils.LowerCaseKey(meta.Details.Symbol))
	k.getStore(ctx, k.chainLowerKey).Set(key, &meta)
}

func (k chainKeeper) getTokenMetadataByAsset(ctx sdk.Context, asset string) (types.ERC20TokenMetadata, bool) {
	var result types.ERC20TokenMetadata
	key := tokenMetadataByAssetPrefix.Append(utils.LowerCaseKey(asset))
	found := k.getStore(ctx, k.chainLowerKey).Get(key, &result)

	return result, found
}

func (k chainKeeper) getTokenMetadataBySymbol(ctx sdk.Context, symbol string) (types.ERC20TokenMetadata, bool) {
	var result types.ERC20TokenMetadata
	key := tokenMetadataBySymbolPrefix.Append(utils.LowerCaseKey(symbol))
	found := k.getStore(ctx, k.chainLowerKey).Get(key, &result)

	return result, found
}

func (k chainKeeper) GetTokens(ctx sdk.Context) []types.ERC20Token {
	tokensMetadata := k.getTokensMetadata(ctx)
	tokens := make([]types.ERC20Token, len(tokensMetadata))

	for i, tokenMetadata := range tokensMetadata {
		tokens[i] = types.CreateERC20Token(func(m types.ERC20TokenMetadata) {
			k.setTokenMetadata(ctx, m)
		}, tokenMetadata)
	}

	return tokens
}

func (k chainKeeper) getTokensMetadata(ctx sdk.Context) []types.ERC20TokenMetadata {
	iter := k.getStore(ctx, k.chainLowerKey).Iterator(tokenMetadataByAssetPrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	var tokens []types.ERC20TokenMetadata
	for ; iter.Valid(); iter.Next() {
		var token types.ERC20TokenMetadata
		iter.UnmarshalValue(&token)
		tokens = append(tokens, token)
	}
	return tokens
}

func (k chainKeeper) initTokenMetadata(ctx sdk.Context, asset string, details types.TokenDetails, address types.Address) (types.ERC20TokenMetadata, error) {
	if err := details.Validate(); err != nil {
		return types.ERC20TokenMetadata{}, err
	}

	// perform a few checks now, so that it is impossible to get errors later
	if token := k.GetERC20TokenByAsset(ctx, asset); !token.Is(types.NonExistent) {
		return types.ERC20TokenMetadata{}, fmt.Errorf("token for asset '%s' already set", asset)
	}

	if token := k.GetERC20TokenBySymbol(ctx, details.Symbol); !token.Is(types.NonExistent) {
		return types.ERC20TokenMetadata{}, fmt.Errorf("token with symbol '%s' already set", details.Symbol)
	}

	chainID := k.getSigner(ctx).ChainID()

	burnerCode, ok := k.GetBurnerByteCode(ctx)
	if !ok {
		return types.ERC20TokenMetadata{}, fmt.Errorf("burner code not found for chain %s", k.chainLowerKey)
	}

	if !address.IsZeroAddress() {
		meta := types.ERC20TokenMetadata{
			Asset:        asset,
			Details:      details,
			TokenAddress: address,
			ChainID:      sdk.NewIntFromBigInt(chainID),
			Status:       types.Initialized,
			IsExternal:   true,
			BurnerCode:   burnerCode,
		}
		k.setTokenMetadata(ctx, meta)

		return meta, nil
	}

	gatewayAddr, found := k.GetGatewayAddress(ctx)
	if !found {
		return types.ERC20TokenMetadata{}, fmt.Errorf("axelar gateway address for chain '%s' not set", k.chainLowerKey)
	}

	_, found = k.GetTokenByteCode(ctx)
	if !found {
		return types.ERC20TokenMetadata{}, fmt.Errorf("bytecodes for token contract for chain '%s' not found", k.chainLowerKey)
	}

	tokenAddr, err := k.getTokenAddress(ctx, details, gatewayAddr)
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
		IsExternal:   false,
		BurnerCode:   burnerCode,
	}
	k.setTokenMetadata(ctx, meta)

	return meta, nil
}

// SetPendingGateway sets the pending gateway
func (k chainKeeper) SetPendingGateway(ctx sdk.Context, address common.Address) {
	gateway := types.Gateway{Address: types.Address(address), Status: types.GatewayStatusPending}
	k.setGateway(ctx, gateway)
}

func (k chainKeeper) setGateway(ctx sdk.Context, gateway types.Gateway) {
	k.getStore(ctx, k.chainLowerKey).Set(gatewayKey, &gateway)
}

// ConfirmPendingGateway confirms the pending gateway
func (k chainKeeper) ConfirmPendingGateway(ctx sdk.Context) error {
	if gateway := k.getGateway(ctx); gateway.Status == types.GatewayStatusPending {
		gateway.Status = types.GatewayStatusConfirmed
		k.getStore(ctx, k.chainLowerKey).Set(gatewayKey, &gateway)

		return nil
	}

	return fmt.Errorf("no pending gateway found for chain %s", k.chainLowerKey)
}

func (k chainKeeper) getGateway(ctx sdk.Context) types.Gateway {
	var gateway types.Gateway
	k.getStore(ctx, k.chainLowerKey).Get(gatewayKey, &gateway)
	return gateway
}

// DeletePendingGateway deletes the pending gateway
func (k chainKeeper) DeletePendingGateway(ctx sdk.Context) error {
	if gateway := k.getGateway(ctx); gateway.Status == types.GatewayStatusPending {
		k.getStore(ctx, k.chainLowerKey).Delete(gatewayKey)

		return nil
	}

	return fmt.Errorf("no pending gateway found for chain %s", k.chainLowerKey)
}

// GetPendingGatewayAddress returns the pending addres of gateway
func (k chainKeeper) GetPendingGatewayAddress(ctx sdk.Context) (common.Address, bool) {
	if gateway := k.getGateway(ctx); gateway.Status == types.GatewayStatusPending {
		return common.Address(gateway.Address), true
	}

	return common.Address{}, false
}

// GetGatewayAddress returns the confirmed address of gateway
func (k chainKeeper) GetGatewayAddress(ctx sdk.Context) (common.Address, bool) {
	if gateway := k.getGateway(ctx); gateway.Status == types.GatewayStatusConfirmed {
		return common.Address(gateway.Address), true
	}

	return common.Address{}, false
}

func (k chainKeeper) getSubspace(ctx sdk.Context) (params.Subspace, bool) {
	// When a node restarts or joins the network after genesis, it might not have all EVM subspaces initialized.
	// The following check has to be done regardless, if we would only do it dependent on the existence of a subspace
	// different nodes would consume different amounts of gas and it would result in a consensus failure
	if !k.getBaseStore(ctx).Has(subspacePrefix.AppendStr(k.chainLowerKey)) {
		return params.Subspace{}, false
	}

	chainKey := types.ModuleName + "_" + k.chainLowerKey
	subspace, ok := k.subspaces[chainKey]
	if !ok {
		subspace = k.paramsKeeper.Subspace(chainKey).WithKeyTable(types.KeyTable())
		k.subspaces[chainKey] = subspace
	}
	return subspace, true
}
