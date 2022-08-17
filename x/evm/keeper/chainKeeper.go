package keeper

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	evmTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/utils/funcs"
)

var (
	gatewayKey             = utils.KeyFromStr("gateway")
	unsignedBatchIDKey     = utils.KeyFromStr("unsigned_command_batch_id")
	latestSignedBatchIDKey = utils.KeyFromStr("latest_signed_command_batch_id")

	tokenMetadataByAssetPrefix  = utils.KeyFromStr("token_deployment_by_asset")
	tokenMetadataBySymbolPrefix = utils.KeyFromStr("token_deployment_by_symbol")
	confirmedDepositPrefix      = utils.KeyFromStr("confirmed_deposit")
	burnedDepositPrefix         = utils.KeyFromStr("burned_deposit")
	commandBatchPrefix          = utils.KeyFromStr("batched_commands")
	commandPrefix               = utils.KeyFromStr("command")
	burnerAddrPrefix            = utils.KeyFromStr("burnerAddr")
	eventPrefix                 = utils.KeyFromStr("event")

	commandQueueName        = "cmd_queue"
	confirmedEventQueueName = "confirmed_event_queue"
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

func (k chainKeeper) GetChainID(ctx sdk.Context) (sdk.Int, bool) {
	network, ok := k.GetNetwork(ctx)
	if !ok {
		return sdk.Int{}, false
	}

	return k.GetChainIDByNetwork(ctx, network)
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
func (k chainKeeper) getTokenAddress(ctx sdk.Context, details types.TokenDetails, gatewayAddr types.Address) (types.Address, error) {
	var saltToken [32]byte
	copy(saltToken[:], crypto.Keccak256Hash([]byte(details.Symbol)).Bytes())

	uint8Type, err := abi.NewType("uint8", "uint8", nil)
	if err != nil {
		return types.Address{}, err
	}

	uint256Type, err := abi.NewType("uint256", "uint256", nil)
	if err != nil {
		return types.Address{}, err
	}

	stringType, err := abi.NewType("string", "string", nil)
	if err != nil {
		return types.Address{}, err
	}

	arguments := abi.Arguments{{Type: stringType}, {Type: stringType}, {Type: uint8Type}, {Type: uint256Type}}
	packed, err := arguments.Pack(details.TokenName, details.Symbol, details.Decimals, details.Capacity.BigInt())
	if err != nil {
		return types.Address{}, err
	}

	bytecode, ok := k.GetTokenByteCode(ctx)
	if !ok {
		return types.Address{}, fmt.Errorf("bytecode for token contract not found")
	}

	tokenInitCode := append(bytecode, packed...)
	tokenInitCodeHash := crypto.Keccak256Hash(tokenInitCode)

	tokenAddr := types.Address(crypto.CreateAddress2(common.Address(gatewayAddr), saltToken, tokenInitCodeHash.Bytes()))
	return tokenAddr, nil
}

// GenerateSalt calculates a salt based on network params and recipient address to use for burner address generation
func (k chainKeeper) GenerateSalt(ctx sdk.Context, recipient string) types.Hash {
	nonce := utils.GetNonce(ctx.HeaderHash(), ctx.BlockGasMeter())
	bz := []byte(recipient)
	bz = append(bz, nonce[:]...)
	salt := types.Hash(common.BytesToHash(crypto.Keccak256Hash(bz).Bytes()))
	return salt
}

// GetBurnerAddress calculates a burner address for the given token address, salt, and axelar gateway address
func (k chainKeeper) GetBurnerAddress(ctx sdk.Context, token types.ERC20Token, salt types.Hash, gatewayAddr types.Address) (types.Address, error) {
	var tokenBurnerCodeHash types.Hash
	if token.IsExternal() {
		// always use the latest burner byte code for external token
		burnerCode, ok := k.GetBurnerByteCode(ctx)
		if !ok {
			return types.Address{}, fmt.Errorf("burner code not found for chain %s", k.chainLowerKey)
		}
		tokenBurnerCodeHash = types.Hash(crypto.Keccak256Hash(burnerCode))
	} else {
		tokenBurnerCodeHash = funcs.MustOk(token.GetBurnerCodeHash())
	}

	var initCodeHash types.Hash
	switch tokenBurnerCodeHash.Hex() {
	case types.BurnerCodeHashV1:
		addressType, err := abi.NewType("address", "address", nil)
		if err != nil {
			return types.Address{}, err
		}

		bytes32Type, err := abi.NewType("bytes32", "bytes32", nil)
		if err != nil {
			return types.Address{}, err
		}

		arguments := abi.Arguments{{Type: addressType}, {Type: bytes32Type}}
		params, err := arguments.Pack(token.GetAddress(), salt)
		if err != nil {
			return types.Address{}, err
		}

		initCodeHash = types.Hash(crypto.Keccak256Hash(append(token.GetBurnerCode(), params...)))
	case types.BurnerCodeHashV2, types.BurnerCodeHashV3, types.BurnerCodeHashV4, types.BurnerCodeHashV5:
		initCodeHash = tokenBurnerCodeHash
	default:
		return types.Address{}, fmt.Errorf("unsupported burner code with hash %s for chain %s", tokenBurnerCodeHash.Hex(), k.chainLowerKey)
	}

	return types.Address(crypto.CreateAddress2(common.Address(gatewayAddr), salt, initCodeHash.Bytes())), nil
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

func (k chainKeeper) CreateERC20Token(ctx sdk.Context, asset string, details types.TokenDetails, address types.Address) (types.ERC20Token, error) {
	metadata, err := k.initTokenMetadata(ctx, asset, details, address)
	if err != nil {
		return types.NilToken, err
	}

	return types.CreateERC20Token(func(m types.ERC20TokenMetadata) {
		k.setTokenMetadata(ctx, m)
	}, metadata), nil
}

// GetERC20TokenBySymbol returns the erc20 token by asset
func (k chainKeeper) GetERC20TokenByAsset(ctx sdk.Context, asset string) types.ERC20Token {
	metadata, ok := k.getTokenMetadataByAsset(ctx, asset)
	if !ok {
		return types.NilToken
	}

	return types.CreateERC20Token(func(m types.ERC20TokenMetadata) {
		k.setTokenMetadata(ctx, m)
	}, metadata)
}

// GetERC20TokenBySymbol returns the erc20 token by symbol
func (k chainKeeper) GetERC20TokenBySymbol(ctx sdk.Context, symbol string) types.ERC20Token {
	metadata, ok := k.getTokenMetadataBySymbol(ctx, symbol)
	if !ok {
		return types.NilToken
	}

	return types.CreateERC20Token(func(m types.ERC20TokenMetadata) {
		k.setTokenMetadata(ctx, m)
	}, metadata)
}

// GetConfirmedEventQueue returns a queue of all the confirmed events
func (k chainKeeper) GetConfirmedEventQueue(ctx sdk.Context) utils.KVQueue {
	blockHeightBz := make([]byte, 8)
	binary.BigEndian.PutUint64(blockHeightBz, uint64(ctx.BlockHeight()))

	return utils.NewGeneralKVQueue(
		confirmedEventQueueName,
		k.getStore(ctx, k.chainLowerKey),
		k.Logger(ctx),
		func(value codec.ProtoMarshaler) utils.Key {
			event := value.(*types.Event)

			indexBz := make([]byte, 8)
			binary.BigEndian.PutUint64(indexBz, event.Index)

			return utils.KeyFromBz(blockHeightBz).
				Append(utils.KeyFromBz(event.TxID.Bytes())).
				Append(utils.KeyFromBz(indexBz))
		},
	)
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

// GetDeposit retrieves a confirmed/burned deposit
func (k chainKeeper) GetDeposit(ctx sdk.Context, txID types.Hash, burnAddr types.Address) (types.ERC20Deposit, types.DepositStatus, bool) {
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
	chainID, found := k.GetChainIDByNetwork(ctx, network)

	if !found {
		panic(fmt.Sprintf("could not find chain ID for network '%s'", network))
	}
	return evmTypes.NewEIP155Signer(chainID.BigInt())
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

// GetNetworkByID returns the network name for a given chain and network ID
func (k chainKeeper) GetNetworkByID(ctx sdk.Context, id sdk.Int) (string, bool) {
	subspace, ok := k.getSubspace(ctx)
	if !ok {
		return "", false
	}

	var p types.Params
	subspace.GetParamSet(ctx, &p)
	for _, n := range p.Networks {
		if n.Id == id {
			return n.Name, true
		}
	}

	return "", false
}

// GetChainIDByNetwork returns the network name for a given chain and network name
func (k chainKeeper) GetChainIDByNetwork(ctx sdk.Context, network string) (sdk.Int, bool) {
	if network == "" {
		return sdk.Int{}, false
	}
	subspace, ok := k.getSubspace(ctx)
	if !ok {
		return sdk.Int{}, false
	}

	var p types.Params
	subspace.GetParamSet(ctx, &p)
	for _, n := range p.Networks {
		if n.Name == network {
			return n.Id, true
		}
	}

	return sdk.Int{}, false
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
func (k chainKeeper) CreateNewBatchToSign(ctx sdk.Context) (types.CommandBatch, error) {
	var firstCmd types.Command
	ok := k.getCommandQueue(ctx).Dequeue(&firstCmd)
	if !ok {
		return types.CommandBatch{}, nil
	}

	chainID := sdk.NewIntFromBigInt(k.getSigner(ctx).ChainID())
	gasLimit := k.getCommandsGasLimit(ctx)
	gasCost := uint32(firstCmd.MaxGasCost)
	keyID := firstCmd.KeyID
	filter := func(value codec.ProtoMarshaler) bool {
		cmd, ok := value.(*types.Command)
		gasCost += cmd.MaxGasCost

		return ok && cmd.KeyID == keyID && gasCost <= gasLimit
	}

	commands := []types.Command{firstCmd.Clone()}
	for {
		var cmd types.Command
		ok := k.getCommandQueue(ctx).DequeueIf(&cmd, filter)
		if !ok {
			break
		}

		commands = append(commands, cmd.Clone())
	}

	commandBatch, err := types.NewCommandBatchMetadata(ctx.BlockHeight(), chainID, keyID, commands)
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
func (k chainKeeper) getCommandQueue(ctx sdk.Context) utils.BlockHeightKVQueue {
	return utils.NewBlockHeightKVQueue(
		commandQueueName,
		k.getStore(ctx, k.chainLowerKey),
		ctx.BlockHeight(),
		k.Logger(ctx),
	)
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
			BurnerCode:   nil,
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

// SetGateway sets the gateway
func (k chainKeeper) SetGateway(ctx sdk.Context, address types.Address) {
	k.setGateway(ctx, types.Gateway{Address: address})
}

// GetGatewayAddress returns the confirmed address of gateway
func (k chainKeeper) GetGatewayAddress(ctx sdk.Context) (types.Address, bool) {
	if gateway := k.getGateway(ctx); !gateway.Address.IsZeroAddress() {
		return gateway.Address, true
	}

	return types.Address{}, false
}

func (k chainKeeper) setGateway(ctx sdk.Context, gateway types.Gateway) {
	k.getStore(ctx, k.chainLowerKey).Set(gatewayKey, &gateway)
}

func (k chainKeeper) getGateway(ctx sdk.Context) types.Gateway {
	var gateway types.Gateway
	k.getStore(ctx, k.chainLowerKey).Get(gatewayKey, &gateway)

	return gateway
}

func getEventKey(eventID types.EventID) utils.Key {
	return eventPrefix.Append(utils.LowerCaseKey(string(eventID)))
}

func (k chainKeeper) setEvent(ctx sdk.Context, event types.Event) {
	k.getStore(ctx, k.chainLowerKey).Set(getEventKey(event.GetID()), &event)
}

func (k chainKeeper) getEvents(ctx sdk.Context) []types.Event {
	iter := k.getStore(ctx, k.chainLowerKey).Iterator(eventPrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	var events []types.Event
	for ; iter.Valid(); iter.Next() {
		var event types.Event
		iter.UnmarshalValue(&event)
		events = append(events, event)
	}

	return events
}

// GetEvent returns the event for the given event ID
func (k chainKeeper) GetEvent(ctx sdk.Context, eventID types.EventID) (event types.Event, ok bool) {
	k.getStore(ctx, k.chainLowerKey).Get(getEventKey(eventID), &event)

	return event, event.Status != types.EventNonExistent
}

// SetConfirmedEvent sets the event as confirmed
func (k chainKeeper) SetConfirmedEvent(ctx sdk.Context, event types.Event) error {
	eventID := event.GetID()
	if _, ok := k.GetEvent(ctx, eventID); ok {
		return fmt.Errorf("event %s is already confirmed", eventID)
	}

	event.Status = types.EventConfirmed

	switch event.GetEvent().(type) {
	case *types.Event_ContractCall, *types.Event_ContractCallWithToken, *types.Event_TokenSent,
		*types.Event_Transfer, *types.Event_TokenDeployed, *types.Event_MultisigOperatorshipTransferred:
		k.GetConfirmedEventQueue(ctx).Enqueue(getEventKey(eventID), &event)
	default:
		return fmt.Errorf("unsupported event type %T", event)
	}

	return nil
}

// SetEventCompleted sets the event as completed
func (k chainKeeper) SetEventCompleted(ctx sdk.Context, eventID types.EventID) error {
	event, ok := k.GetEvent(ctx, eventID)
	if !ok || event.Status != types.EventConfirmed {
		return fmt.Errorf("event %s is not confirmed", eventID)
	}

	event.Status = types.EventCompleted
	k.setEvent(ctx, event)

	return nil
}

// SetEventFailed sets the event as invalid
func (k chainKeeper) SetEventFailed(ctx sdk.Context, eventID types.EventID) error {
	event, ok := k.GetEvent(ctx, eventID)
	if !ok || event.Status != types.EventConfirmed {
		return fmt.Errorf("event %s is not confirmed", eventID)
	}

	event.Status = types.EventFailed
	k.setEvent(ctx, event)

	return nil
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

// validateCommandQueueState checks if the keys of the given map have the correct format to be imported as command queue state.
func (k chainKeeper) validateCommandQueueState(state utils.QueueState, queueName ...string) error {
	if err := state.ValidateBasic(queueName...); err != nil {
		return err
	}

	for _, item := range state.Items {
		var command types.Command
		if err := k.cdc.UnmarshalLengthPrefixed(item.Value, &command); err != nil {
			return err
		}

		if err := command.KeyID.ValidateBasic(); err != nil {
			return err
		}
	}

	return nil
}

// validateConfirmedEventQueueState checks if the keys of the given map have the correct format to be imported as confirmed event state.
func (k chainKeeper) validateConfirmedEventQueueState(state utils.QueueState, queueName ...string) error {
	if err := state.ValidateBasic(queueName...); err != nil {
		return err
	}

	for _, item := range state.Items {
		var event types.Event
		if err := k.cdc.UnmarshalLengthPrefixed(item.Value, &event); err != nil {
			return err
		}

		if err := event.ValidateBasic(); err != nil {
			return err
		}
	}

	return nil
}
