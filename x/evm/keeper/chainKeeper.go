package keeper

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	evmTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/utils/events"
	"github.com/axelarnetwork/axelar-core/utils/key"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/funcs"
)

var (
	gatewayKey                       = key.FromStr("gateway")
	unsignedBatchIDKey               = key.FromStr("unsigned_command_batch_id")
	latestSignedBatchIDKey           = key.FromStr("latest_signed_command_batch_id")
	tokenMetadataByAssetPrefix       = "token_deployment_by_asset"
	tokenMetadataBySymbolPrefix      = key.FromStr("token_deployment_by_symbol")
	confirmedDepositPrefixDeprecated = "confirmed_deposit" // Deprecated
	burnedDepositPrefixDeprecated    = "burned_deposit"    // Deprecated
	commandBatchPrefix               = "batched_commands"
	commandPrefix                    = "command"
	eventPrefix                      = utils.KeyFromStr("event")
	confirmedEventQueueName          = "confirmed_event_queue"
	commandQueueName                 = "cmd_queue"

	burnerAddrPrefix       = key.RegisterStaticKey(types.ModuleName+types.ChainNamespace, 1)
	confirmedDepositPrefix = key.RegisterStaticKey(types.ModuleName+types.ChainNamespace, 2)
	burnedDepositPrefix    = key.RegisterStaticKey(types.ModuleName+types.ChainNamespace, 3)
)

var _ types.ChainKeeper = chainKeeper{}

type chainKeeper struct {
	internalKeeper
	chain nexus.ChainName
}

func (k chainKeeper) GetName() nexus.ChainName {
	return k.chain
}

// GetParams gets the evm module's parameters
func (k chainKeeper) GetParams(ctx sdk.Context) types.Params {
	var p types.Params
	k.getSubspace().GetParamSet(ctx, &p)
	return p
}

// returns the EVM network's gas limist for batched commands
func (k chainKeeper) getCommandsGasLimit(ctx sdk.Context) uint32 {
	return getParam[uint32](k, ctx, types.KeyCommandsGasLimit)
}

func (k chainKeeper) GetChainID(ctx sdk.Context) (sdk.Int, bool) {
	network := k.GetNetwork(ctx)
	return k.GetChainIDByNetwork(ctx, network)
}

// GetNetwork returns the EVM network Axelar-Core is expected to connect to
func (k chainKeeper) GetNetwork(ctx sdk.Context) string {
	return getParam[string](k, ctx, types.KeyNetwork)
}

// GetRequiredConfirmationHeight returns the required block confirmation height
func (k chainKeeper) GetRequiredConfirmationHeight(ctx sdk.Context) uint64 {
	return getParam[uint64](k, ctx, types.KeyConfirmationHeight)
}

// GetRevoteLockingPeriod returns the lock period for revoting
func (k chainKeeper) GetRevoteLockingPeriod(ctx sdk.Context) int64 {
	return getParam[int64](k, ctx, types.KeyRevoteLockingPeriod)
}

// GetVotingThreshold returns voting threshold
func (k chainKeeper) GetVotingThreshold(ctx sdk.Context) utils.Threshold {
	return getParam[utils.Threshold](k, ctx, types.KeyVotingThreshold)
}

// GetMinVoterCount returns minimum voter count for voting
func (k chainKeeper) GetMinVoterCount(ctx sdk.Context) int64 {
	return getParam[int64](k, ctx, types.KeyMinVoterCount)
}

// SetBurnerInfo saves the burner info for a given address
func (k chainKeeper) SetBurnerInfo(ctx sdk.Context, burnerInfo types.BurnerInfo) {
	funcs.MustNoErr(
		k.getStore(ctx).SetNewValidated(burnerAddrPrefix.Append(key.FromStr(burnerInfo.BurnerAddress.Hex())), &burnerInfo))
}

// GetBurnerInfo retrieves the burner info for a given address
func (k chainKeeper) GetBurnerInfo(ctx sdk.Context, burnerAddr types.Address) *types.BurnerInfo {
	var result types.BurnerInfo
	if !k.getStore(ctx).GetNew(burnerAddrPrefix.Append(key.FromStr(burnerAddr.Hex())), &result) {
		return nil
	}

	return &result
}

func (k chainKeeper) getBurnerInfos(ctx sdk.Context) []types.BurnerInfo {
	iter := k.getStore(ctx).IteratorNew(burnerAddrPrefix)
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

	bytecode := k.GetTokenByteCode(ctx)
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
		burnerCode := k.GetBurnerByteCode(ctx)
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
		return types.Address{}, fmt.Errorf("unsupported burner code with hash %s for chain %s", tokenBurnerCodeHash.Hex(), k.chain)
	}

	return types.Address(crypto.CreateAddress2(common.Address(gatewayAddr), salt, initCodeHash.Bytes())), nil
}

// GetBurnerByteCode returns the bytecode for the burner contract
func (k chainKeeper) GetBurnerByteCode(ctx sdk.Context) []byte {
	return getParam[[]byte](k, ctx, types.KeyBurnable)
}

// GetTokenByteCode returns the bytecodes for the token contract
func (k chainKeeper) GetTokenByteCode(ctx sdk.Context) []byte {
	return getParam[[]byte](k, ctx, types.KeyToken)
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

// GetERC20TokenByAsset returns the erc20 token by asset
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
		k.getStore(ctx),
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
	if k.getStore(ctx).HasNew(key.FromStr(commandPrefix).Append(key.FromStr(command.ID.Hex()))) {
		return fmt.Errorf("command %s already exists", command.ID.Hex())
	}

	k.getCommandQueue(ctx).Enqueue(utils.LowerCaseKey(commandPrefix).AppendStr(command.ID.Hex()), &command)
	return nil
}

// GetCommand returns the command specified by the given ID
func (k chainKeeper) GetCommand(ctx sdk.Context, id types.CommandID) (types.Command, bool) {
	var cmd types.Command
	found := k.getStore(ctx).GetNew(key.FromStr(commandPrefix).Append(key.FromStr(id.Hex())), &cmd)

	return cmd, found
}

// GetPendingCommands returns the list of commands not yet added to any batch
func (k chainKeeper) GetPendingCommands(ctx sdk.Context) []types.Command {
	var commands []types.Command

	keys := k.getCommandQueue(ctx).Keys()
	for _, queueKey := range keys {
		var cmd types.Command
		ok := k.getStore(ctx).GetNew(key.FromBz(queueKey.AsKey()), &cmd)
		if ok {
			commands = append(commands, cmd)
		}
	}

	return commands
}

// Deprecated: Use SetDeposit instead
func (k chainKeeper) setLegacyDeposit(ctx sdk.Context, deposit types.ERC20Deposit, state types.DepositStatus) {
	switch state {
	case types.DepositStatus_Confirmed:
		funcs.MustNoErr(
			k.getStore(ctx).SetNewValidated(
				key.FromStr(confirmedDepositPrefixDeprecated).Append(key.FromStr(deposit.TxID.Hex())).Append(key.FromStr(deposit.BurnerAddress.Hex())), &deposit))
	case types.DepositStatus_Burned:
		funcs.MustNoErr(
			k.getStore(ctx).SetNewValidated(
				key.FromStr(burnedDepositPrefixDeprecated).Append(key.FromStr(deposit.TxID.Hex())).Append(key.FromStr(deposit.BurnerAddress.Hex())), &deposit))
	default:
		panic("invalid deposit state")
	}
}

// Deprecated: Use getConfirmedDeposits and getBurnedDeposits instead
func (k chainKeeper) getLegacyDeposits(ctx sdk.Context, state types.DepositStatus) (deposits []types.ERC20Deposit) {
	var prefix key.Key
	switch state {
	case types.DepositStatus_Confirmed:
		prefix = key.FromStr(confirmedDepositPrefixDeprecated)
	case types.DepositStatus_Burned:
		prefix = key.FromStr(burnedDepositPrefixDeprecated)
	default:
		panic("invalid deposit state")
	}

	iter := k.getStore(ctx).IteratorNew(prefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var deposit types.ERC20Deposit
		iter.UnmarshalValue(&deposit)

		deposits = append(deposits, deposit)
	}

	return deposits
}

// GetLegacyDeposit retrieves a confirmed/burned deposit by tx id and burner address
// Deprecated: Use GetDeposit instead
func (k chainKeeper) GetLegacyDeposit(ctx sdk.Context, txID types.Hash, burnAddr types.Address) (types.ERC20Deposit, types.DepositStatus, bool) {
	var deposit types.ERC20Deposit

	if k.getStore(ctx).GetNew(key.FromStr(confirmedDepositPrefixDeprecated).Append(key.FromStr(txID.Hex())).Append(key.FromStr(burnAddr.Hex())), &deposit) {
		return deposit, types.DepositStatus_Confirmed, true
	}
	if k.getStore(ctx).GetNew(key.FromStr(burnedDepositPrefixDeprecated).Append(key.FromStr(txID.Hex())).Append(key.FromStr(burnAddr.Hex())), &deposit) {
		return deposit, types.DepositStatus_Burned, true
	}

	return types.ERC20Deposit{}, types.DepositStatus_None, false
}

// GetDepositsByTxID returns all deposits for the given tx ID and status
func (k chainKeeper) GetDepositsByTxID(ctx sdk.Context, txID types.Hash, status types.DepositStatus) ([]types.ERC20Deposit, error) {
	var prefix key.Key
	switch status {
	case types.DepositStatus_Confirmed:
		prefix = confirmedDepositPrefix
	case types.DepositStatus_Burned:
		prefix = burnedDepositPrefix
	default:
		return nil, fmt.Errorf("unsupported deposit status %s", status.String())
	}

	iter := k.getStore(ctx).IteratorNew(prefix.Append(key.FromStr(txID.Hex())))
	defer utils.CloseLogError(iter, k.Logger(ctx))

	var deposits []types.ERC20Deposit
	for ; iter.Valid(); iter.Next() {
		var deposit types.ERC20Deposit
		iter.UnmarshalValue(&deposit)

		deposits = append(deposits, deposit)
	}

	return deposits, nil
}

// GetDeposit retrieves a confirmed/burned deposit
func (k chainKeeper) GetDeposit(ctx sdk.Context, txID types.Hash, logIndex uint64) (types.ERC20Deposit, types.DepositStatus, bool) {
	var deposit types.ERC20Deposit

	if k.getStore(ctx).GetNew(confirmedDepositPrefix.Append(key.FromStr(txID.Hex())).Append(key.FromUInt(logIndex)), &deposit) {
		return deposit, types.DepositStatus_Confirmed, true
	}
	if k.getStore(ctx).GetNew(burnedDepositPrefix.Append(key.FromStr(txID.Hex())).Append(key.FromUInt(logIndex)), &deposit) {
		return deposit, types.DepositStatus_Burned, true
	}

	return types.ERC20Deposit{}, 0, false
}

// getConfirmedDeposits retrieves all the confirmed ERC20 deposits
func (k chainKeeper) getConfirmedDeposits(ctx sdk.Context) []types.ERC20Deposit {
	var deposits []types.ERC20Deposit
	iter := k.getStore(ctx).IteratorNew(confirmedDepositPrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var deposit types.ERC20Deposit
		iter.UnmarshalValue(&deposit)
		deposits = append(deposits, deposit)
	}

	return deposits
}

// GetConfirmedDepositsPaginated retrieves all the confirmed ERC20 deposits with the given pagination properties
func (k chainKeeper) GetConfirmedDepositsPaginated(ctx sdk.Context, pageRequest *query.PageRequest) ([]types.ERC20Deposit, *query.PageResponse, error) {
	var deposits []types.ERC20Deposit

	// TODO: refactor iteration over values using a prefix to avoid collisions
	resp, err := query.Paginate(prefix.NewStore(k.getStore(ctx).KVStore, append(confirmedDepositPrefix.Bytes(), []byte(key.DefaultDelimiter)...)), pageRequest, func(key []byte, value []byte) error {
		var deposit types.ERC20Deposit
		k.cdc.MustUnmarshalLengthPrefixed(value, &deposit)
		deposits = append(deposits, deposit)
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return deposits, resp, nil
}

// getBurnedDeposits retrieves all the burned ERC20 deposits
func (k chainKeeper) getBurnedDeposits(ctx sdk.Context) []types.ERC20Deposit {
	var deposits []types.ERC20Deposit
	iter := k.getStore(ctx).IteratorNew(burnedDepositPrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var deposit types.ERC20Deposit
		iter.UnmarshalValue(&deposit)
		deposits = append(deposits, deposit)
	}

	return deposits
}

func (k chainKeeper) getSigner(ctx sdk.Context) evmTypes.EIP155Signer {

	network := getParam[string](k, ctx, types.KeyNetwork)
	chainID, found := k.GetChainIDByNetwork(ctx, network)

	// both chain, subspace, and network must be valid if the chain keeper was instantiated,
	// so a nil value here must be a catastrophic failure
	if !found {
		panic(fmt.Sprintf("could not find chain ID for network '%s'", network))
	}
	return evmTypes.NewEIP155Signer(chainID.BigInt())
}

// SetDeposit stores confirmed or burned deposits
func (k chainKeeper) SetDeposit(ctx sdk.Context, deposit types.ERC20Deposit, state types.DepositStatus) {
	switch state {
	case types.DepositStatus_Confirmed:
		funcs.MustNoErr(
			k.getStore(ctx).SetNewValidated(
				confirmedDepositPrefix.Append(key.FromStr(deposit.TxID.Hex())).Append(key.FromUInt(deposit.LogIndex)), &deposit))
	case types.DepositStatus_Burned:
		funcs.MustNoErr(
			k.getStore(ctx).SetNewValidated(
				burnedDepositPrefix.Append(key.FromStr(deposit.TxID.Hex())).Append(key.FromUInt(deposit.LogIndex)), &deposit))
	default:
		panic("invalid deposit state")
	}
}

// DeleteDeposit deletes the given deposit
func (k chainKeeper) DeleteDeposit(ctx sdk.Context, deposit types.ERC20Deposit) {
	k.getStore(ctx).DeleteNew(
		confirmedDepositPrefix.Append(key.FromStr(deposit.TxID.Hex())).Append(key.FromUInt(deposit.LogIndex)))
	k.getStore(ctx).DeleteNew(
		burnedDepositPrefix.Append(key.FromStr(deposit.TxID.Hex())).Append(key.FromUInt(deposit.LogIndex)))
}

// GetNetworkByID returns the network name for a given chain and network ID
func (k chainKeeper) GetNetworkByID(ctx sdk.Context, id sdk.Int) (string, bool) {
	p := k.GetParams(ctx)
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
	p := k.GetParams(ctx)
	for _, n := range p.Networks {
		if n.Name == network {
			return n.Id, true
		}
	}

	return sdk.Int{}, false
}

func (k chainKeeper) setCommandBatchMetadata(ctx sdk.Context, meta types.CommandBatchMetadata) {
	funcs.MustNoErr(
		k.getStore(ctx).SetNewValidated(key.FromStr(commandBatchPrefix).Append(key.FromBz(meta.ID)), &meta))
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
	k.getStore(ctx).GetNew(key.FromStr(commandBatchPrefix).Append(key.FromBz(id)), &batch)
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
	iter := k.getStore(ctx).Iterator(utils.KeyFromStr(commandBatchPrefix))
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
	return k.getStore(ctx).GetRawNew(latestSignedBatchIDKey)
}

// SetLatestSignedCommandBatchID stores the latest signed command batch ID
func (k chainKeeper) SetLatestSignedCommandBatchID(ctx sdk.Context, id []byte) {
	k.getStore(ctx).SetRawNew(latestSignedBatchIDKey, id)
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
	gasCost := firstCmd.MaxGasCost
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
	k.getStore(ctx).DeleteNew(unsignedBatchIDKey)
}

func (k chainKeeper) getUnsignedCommandBatch(ctx sdk.Context) types.CommandBatchMetadata {
	if id := k.getStore(ctx).GetRawNew(unsignedBatchIDKey); id != nil {
		return k.getCommandBatchMetadata(ctx, id)
	}

	return types.CommandBatchMetadata{}
}

func (k chainKeeper) setUnsignedCommandBatchID(ctx sdk.Context, id []byte) {
	k.getStore(ctx).SetRawNew(unsignedBatchIDKey, id)
}

// returns the queue of commands
func (k chainKeeper) getCommandQueue(ctx sdk.Context) utils.BlockHeightKVQueue {
	return utils.NewBlockHeightKVQueue(
		commandQueueName,
		k.getStore(ctx),
		ctx.BlockHeight(),
		k.Logger(ctx),
	)
}

func (k chainKeeper) setTokenMetadata(ctx sdk.Context, meta types.ERC20TokenMetadata) {
	// lookup by asset
	funcs.MustNoErr(
		k.getStore(ctx).SetNewValidated(key.FromStr(tokenMetadataByAssetPrefix).Append(key.FromStr(meta.Asset)), &meta))

	// lookup by symbol
	funcs.MustNoErr(
		k.getStore(ctx).SetNewValidated(tokenMetadataBySymbolPrefix.Append(key.FromStr(meta.Details.Symbol)), &meta))
}

func (k chainKeeper) getTokenMetadataByAsset(ctx sdk.Context, asset string) (types.ERC20TokenMetadata, bool) {
	var result types.ERC20TokenMetadata
	found := k.getStore(ctx).GetNew(key.FromStr(tokenMetadataByAssetPrefix).Append(key.FromStr(asset)), &result)

	return result, found
}

func (k chainKeeper) getTokenMetadataBySymbol(ctx sdk.Context, symbol string) (types.ERC20TokenMetadata, bool) {
	var result types.ERC20TokenMetadata
	found := k.getStore(ctx).GetNew(tokenMetadataBySymbolPrefix.Append(key.FromStr(symbol)), &result)

	return result, found
}

// GetTokenByAddress finds a token's information by its address
func (k chainKeeper) GetERC20TokenByAddress(ctx sdk.Context, address types.Address) types.ERC20Token {
	for _, tokenMetadata := range k.getTokensMetadata(ctx) {
		if tokenMetadata.TokenAddress == address {
			return types.CreateERC20Token(func(m types.ERC20TokenMetadata) {
				k.setTokenMetadata(ctx, m)
			}, tokenMetadata)
		}
	}

	return types.ERC20Token{}
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
	iter := k.getStore(ctx).Iterator(utils.LowerCaseKey(tokenMetadataByAssetPrefix))
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

	burnerCode := k.GetBurnerByteCode(ctx)

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
		return types.ERC20TokenMetadata{}, fmt.Errorf("axelar gateway address for chain '%s' not set", k.chain)
	}

	tokenAddr, err := k.getTokenAddress(ctx, details, gatewayAddr)
	if err != nil {
		return types.ERC20TokenMetadata{}, err
	}

	// all good
	meta := types.ERC20TokenMetadata{
		Asset:        asset,
		Details:      details,
		TokenAddress: tokenAddr,
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
	// TODO: remove this guard clause once genesis state can have nil Gateway
	if gateway.Address.IsZeroAddress() {
		return
	}

	funcs.MustNoErr(
		k.getStore(ctx).SetNewValidated(gatewayKey, &gateway))
}

func (k chainKeeper) getGateway(ctx sdk.Context) types.Gateway {
	var gateway types.Gateway
	k.getStore(ctx).GetNew(gatewayKey, &gateway)

	return gateway
}

func getEventKey(eventID types.EventID) utils.Key {
	return eventPrefix.Append(utils.LowerCaseKey(string(eventID)))
}

func (k chainKeeper) setEvent(ctx sdk.Context, event types.Event) {
	funcs.MustNoErr(
		k.getStore(ctx).SetNewValidated(key.FromBz(getEventKey(event.GetID()).AsKey()), &event))
}

func (k chainKeeper) getEvents(ctx sdk.Context) []types.Event {
	iter := k.getStore(ctx).Iterator(eventPrefix)
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
	k.getStore(ctx).GetNew(key.FromBz(getEventKey(eventID).AsKey()), &event)

	return event, event.Status != types.EventNonExistent
}

func (k chainKeeper) SetConfirmedEvent(ctx sdk.Context, event types.Event) error {
	eventID := event.GetID()
	if _, ok := k.GetEvent(ctx, eventID); ok {
		return fmt.Errorf("event %s is already confirmed", eventID)
	}

	event.Status = types.EventConfirmed
	k.setEvent(ctx, event)

	events.Emit(ctx, &types.EVMEventConfirmed{
		Chain:   event.Chain,
		EventID: event.GetID(),
		Type:    event.GetEventType(),
	})

	return nil
}

// EnqueueConfirmedEvent enqueues the confirmed event
func (k chainKeeper) EnqueueConfirmedEvent(ctx sdk.Context, id types.EventID) error {
	event, ok := k.GetEvent(ctx, id)
	if !ok {
		return fmt.Errorf("event %s does not exist", id)
	}
	if event.Status != types.EventConfirmed {
		return fmt.Errorf("event %s is not confirmed", id)
	}

	switch event.GetEvent().(type) {
	// the missing Event_ContractCall is no longer allowed to be enqueued in the
	// EVM module, it must be routed through the nexus module instead
	case *types.Event_ContractCallWithToken,
		*types.Event_TokenSent,
		*types.Event_Transfer,
		*types.Event_TokenDeployed,
		*types.Event_MultisigOperatorshipTransferred:
		k.GetConfirmedEventQueue(ctx).Enqueue(getEventKey(id), &event)
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

	events.Emit(ctx,
		&types.EVMEventCompleted{
			Chain:   event.Chain,
			EventID: event.GetID(),
			Type:    event.GetEventType(),
		})

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

	k.Logger(ctx).Debug("failed handling event",
		"chain", event.Chain,
		"eventID", event.GetID(),
	)

	events.Emit(ctx,
		&types.EVMEventFailed{
			Chain:   event.Chain,
			EventID: event.GetID(),
			Type:    event.GetEventType(),
		})

	return nil
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

func (k chainKeeper) getStore(ctx sdk.Context) utils.KVStore {
	pre := string(chainPrefix.Append(utils.LowerCaseKey(k.chain.String())).AsKey()) + "_"
	return utils.NewNormalizedStore(prefix.NewStore(ctx.KVStore(k.storeKey), []byte(pre)), k.cdc)
}

func (k chainKeeper) getSubspace() params.Subspace {
	chainKey := key.FromStr(types.ModuleName).Append(key.From(k.chain))
	subspace, ok := k.paramsKeeper.GetSubspace(chainKey.String())
	if !ok {
		panic(fmt.Sprintf("subspace for chain %s does not exist", k.chain))
	}
	return subspace
}

func getParam[T any](k chainKeeper, ctx sdk.Context, paramKey []byte) T {
	var value T
	k.getSubspace().Get(ctx, paramKey, &value)
	return value
}
