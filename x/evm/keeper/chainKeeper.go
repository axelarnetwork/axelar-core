package keeper

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	evmTypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

var (
	gatewayKey             = utils.KeyFromStr("gateway")
	unsignedBatchIDKey     = utils.KeyFromStr("unsigned_command_batch_id")
	latestSignedBatchIDKey = utils.KeyFromStr("latest_signed_command_batch_id")

	unsignedTxPrefix          = utils.KeyFromStr("unsigned_tx")
	tokenMetadataPrefix       = utils.KeyFromStr("token_deployment")
	pendingDepositPrefix      = utils.KeyFromStr("pending_deposit")
	confirmedDepositPrefix    = utils.KeyFromStr("confirmed_deposit")
	burnedDepositPrefix       = utils.KeyFromStr("burned_deposit")
	commandBatchPrefix        = utils.KeyFromStr("command_batch")
	commandPrefix             = utils.KeyFromStr("command")
	burnerAddrPrefix          = utils.KeyFromStr("burnerAddr")
	pendingTransferKeyPrefix  = utils.KeyFromStr("pending_transfer_key")
	archivedTransferKeyPrefix = utils.KeyFromStr("archived_transfer_key")

	commandQueueName = "command_queue"
)

var _ types.ChainKeeper = chainKeeper{}

type chainKeeper struct {
	baseKeeper
	chain string
}

// GetName returns the chain name
func (k chainKeeper) GetName() string {
	return k.chain
}

// returns the EVM network's gas limist for batched commands
func (k chainKeeper) getCommandsGasLimit(ctx sdk.Context) uint32 {
	var commandsGasLimit uint32
	subspace, ok := k.getSubspace(ctx, k.chain)

	// the subspace must exist, if not we have a catastrophic failure
	if !ok {
		panic(fmt.Sprintf("subspace for chain '%s' not set", k.chain))
	}

	subspace.Get(ctx, types.KeyCommandsGasLimit, &commandsGasLimit)

	return commandsGasLimit
}

// GetNetwork returns the EVM network Axelar-Core is expected to connect to
func (k chainKeeper) GetNetwork(ctx sdk.Context) (string, bool) {
	var network string
	subspace, ok := k.getSubspace(ctx, k.chain)
	if !ok {
		return network, false
	}

	subspace.Get(ctx, types.KeyNetwork, &network)
	return network, true
}

// GetRequiredConfirmationHeight returns the required block confirmation height
func (k chainKeeper) GetRequiredConfirmationHeight(ctx sdk.Context) (uint64, bool) {
	var h uint64

	subspace, ok := k.getSubspace(ctx, k.chain)
	if !ok {
		return h, false
	}

	subspace.Get(ctx, types.KeyConfirmationHeight, &h)
	return h, true
}

// GetRevoteLockingPeriod returns the lock period for revoting
func (k chainKeeper) GetRevoteLockingPeriod(ctx sdk.Context) (int64, bool) {
	var result int64

	subspace, ok := k.getSubspace(ctx, k.chain)
	if !ok {
		return result, false
	}

	subspace.Get(ctx, types.KeyRevoteLockingPeriod, &result)
	return result, true
}

// GetVotingThreshold returns voting threshold
func (k chainKeeper) GetVotingThreshold(ctx sdk.Context) (utils.Threshold, bool) {
	var threshold utils.Threshold

	subspace, ok := k.getSubspace(ctx, k.chain)
	if !ok {
		return threshold, false
	}

	subspace.Get(ctx, types.KeyVotingThreshold, &threshold)
	return threshold, true
}

// GetMinVoterCount returns minimum voter count for voting
func (k chainKeeper) GetMinVoterCount(ctx sdk.Context) (int64, bool) {
	var minVoterCount int64

	subspace, ok := k.getSubspace(ctx, k.chain)
	if !ok {
		return minVoterCount, false
	}

	subspace.Get(ctx, types.KeyMinVoterCount, &minVoterCount)
	return minVoterCount, true
}

// SetGatewayAddress sets the contract address for Axelar Gateway
func (k chainKeeper) SetGatewayAddress(ctx sdk.Context, addr common.Address) {
	k.getStore(ctx, k.chain).SetRaw(gatewayKey, addr.Bytes())
}

// GetGatewayAddress gets the contract address for Axelar Gateway
func (k chainKeeper) GetGatewayAddress(ctx sdk.Context) (common.Address, bool) {
	bz := k.getStore(ctx, k.chain).GetRaw(gatewayKey)
	return common.BytesToAddress(bz), bz != nil
}

// SetBurnerInfo saves the burner info for a given address
func (k chainKeeper) SetBurnerInfo(ctx sdk.Context, burnerAddr common.Address, burnerInfo *types.BurnerInfo) {
	key := burnerAddrPrefix.AppendStr(burnerAddr.Hex())
	k.getStore(ctx, k.chain).Set(key, burnerInfo)
}

// GetBurnerInfo retrieves the burner info for a given address
func (k chainKeeper) GetBurnerInfo(ctx sdk.Context, burnerAddr common.Address) *types.BurnerInfo {
	key := burnerAddrPrefix.AppendStr(burnerAddr.Hex())
	var result types.BurnerInfo
	if !k.getStore(ctx, k.chain).Get(key, &result) {
		return nil
	}

	return &result
}

// calculates the token address for some asset with the provided axelar gateway address
func (k chainKeeper) getTokenAddress(ctx sdk.Context, assetName string, details types.TokenDetails, gatewayAddr common.Address) (common.Address, error) {
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
func (k chainKeeper) GetBurnerAddressAndSalt(ctx sdk.Context, tokenAddr types.Address, recipient string, gatewayAddr common.Address) (common.Address, common.Hash, error) {
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
func (k chainKeeper) GetBurnerByteCodes(ctx sdk.Context) ([]byte, bool) {
	var b []byte
	subspace, ok := k.getSubspace(ctx, k.chain)
	if !ok {
		return nil, false
	}
	subspace.Get(ctx, types.KeyBurnable, &b)
	return b, true
}

// GetTokenByteCodes returns the bytecodes for the token contract
func (k chainKeeper) GetTokenByteCodes(ctx sdk.Context) ([]byte, bool) {
	var b []byte
	subspace, ok := k.getSubspace(ctx, k.chain)
	if !ok {
		return nil, false
	}
	subspace.Get(ctx, types.KeyToken, &b)
	return b, ok
}

// GetGatewayByteCodes retrieves the byte codes for the Axelar Gateway smart contract
func (k chainKeeper) GetGatewayByteCodes(ctx sdk.Context) ([]byte, bool) {
	var b []byte
	subspace, ok := k.getSubspace(ctx, k.chain)
	if !ok {
		return b, false
	}

	subspace.Get(ctx, types.KeyGateway, &b)
	return b, true
}

func (k chainKeeper) CreateERC20Token(ctx sdk.Context, asset string, details types.TokenDetails) (types.ERC20Token, error) {
	metadata, err := k.initTokenMetadata(ctx, asset, details)
	if err != nil {
		return types.NilToken, err
	}

	return types.CreateERC20Token(func(m types.ERC20TokenMetadata) {
		k.setTokenMetadata(ctx, asset, m)
	}, metadata), nil
}

func (k chainKeeper) GetERC20Token(ctx sdk.Context, asset string) types.ERC20Token {
	metadata, ok := k.getTokenMetadata(ctx, asset)
	if !ok {
		return types.NilToken
	}

	return types.CreateERC20Token(func(m types.ERC20TokenMetadata) {
		k.setTokenMetadata(ctx, asset, m)
	}, metadata)
}

// EnqueueCommand stores the given command; note that overwriting is not allowed
func (k chainKeeper) EnqueueCommand(ctx sdk.Context, command types.Command) error {
	key := commandPrefix.AppendStr(command.ID.Hex())
	if k.getStore(ctx, k.chain).Has(key) {
		return fmt.Errorf("command %s already exists", command.ID.Hex())
	}

	k.getCommandQueue(ctx).Enqueue(key, &command)
	return nil
}

// SetUnsignedTx stores an unsigned transaction
func (k chainKeeper) SetUnsignedTx(ctx sdk.Context, txID string, rawTx *evmTypes.Transaction, pk ecdsa.PublicKey) error {
	bzTX, err := rawTx.MarshalBinary()
	if err != nil {
		return err
	}

	btcecPK := btcec.PublicKey(pk)

	meta := types.TransactionMetadata{
		RawTX:  bzTX,
		PubKey: btcecPK.SerializeCompressed(),
	}

	k.getStore(ctx, k.chain).Set(unsignedTxPrefix.AppendStr(txID), &meta)

	return nil
}

// SetPendingDeposit stores a pending deposit
func (k chainKeeper) SetPendingDeposit(ctx sdk.Context, key exported.PollKey, deposit *types.ERC20Deposit) {
	k.getStore(ctx, k.chain).Set(pendingDepositPrefix.AppendStr(key.String()), deposit)
}

// GetDeposit retrieves a confirmed/burned deposit
func (k chainKeeper) GetDeposit(ctx sdk.Context, txID common.Hash, burnAddr common.Address) (types.ERC20Deposit, types.DepositState, bool) {
	var deposit types.ERC20Deposit

	if k.getStore(ctx, k.chain).Get(confirmedDepositPrefix.AppendStr(txID.Hex()).AppendStr(burnAddr.Hex()), &deposit) {
		return deposit, types.CONFIRMED, true
	}
	if k.getStore(ctx, k.chain).Get(burnedDepositPrefix.AppendStr(txID.Hex()).AppendStr(burnAddr.Hex()), &deposit) {
		return deposit, types.BURNED, true
	}

	return types.ERC20Deposit{}, 0, false
}

// GetConfirmedDeposits retrieves all the confirmed ERC20 deposits
func (k chainKeeper) GetConfirmedDeposits(ctx sdk.Context) []types.ERC20Deposit {
	var deposits []types.ERC20Deposit
	iter := k.getStore(ctx, k.chain).Iterator(confirmedDepositPrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var deposit types.ERC20Deposit
		iter.UnmarshalValue(&deposit)
		deposits = append(deposits, deposit)
	}

	return deposits
}

// AssembleTx returns the data structure resulting from a unsigned tx and the provided signature
func (k chainKeeper) AssembleTx(ctx sdk.Context, txID string, sig tss.Signature) (*evmTypes.Transaction, error) {
	var meta types.TransactionMetadata
	if !k.getStore(ctx, k.chain).Get(unsignedTxPrefix.AppendStr(txID), &meta) {
		return nil, fmt.Errorf("raw tx for ID %s has not been prepared yet", txID)
	}

	btcecPK, err := btcec.ParsePubKey(meta.PubKey, btcec.S256())
	// the setter is controlled by the keeper alone, so an error here should be a catastrophic failure
	if err != nil {
		panic(err)
	}

	pk := btcecPK.ToECDSA()

	var rawTx evmTypes.Transaction
	err = rawTx.UnmarshalBinary(meta.RawTX)
	// the setter is controlled by the keeper alone, so an error here should be a catastrophic failure
	if err != nil {
		panic(err)
	}

	signer := k.getSigner(ctx)

	recoverableSig, err := types.ToSignature(sig, signer.Hash(&rawTx), *pk)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEVM, fmt.Sprintf("could not create recoverable signature: %v", err))
	}

	return rawTx.WithSignature(signer, recoverableSig[:])
}

// GetHashToSign returns the hash to sign of the given raw transaction
func (k chainKeeper) GetHashToSign(ctx sdk.Context, rawTx *evmTypes.Transaction) common.Hash {
	signer := k.getSigner(ctx)
	return signer.Hash(rawTx)
}

func (k chainKeeper) getSigner(ctx sdk.Context) evmTypes.EIP155Signer {
	// both chain, subspace, and network must be valid if the chain keeper was instantiated,
	// so a nil value here must be a catastrophic failure

	var network string
	subspace, ok := k.getSubspace(ctx, k.chain)
	if !ok {
		panic(fmt.Sprintf("could not find subspace for network '%s'", k.chain))
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
	k.getStore(ctx, k.chain).Delete(pendingDepositPrefix.AppendStr(key.String()))
}

// GetPendingDeposit returns the deposit associated with the given poll
func (k chainKeeper) GetPendingDeposit(ctx sdk.Context, key exported.PollKey) (types.ERC20Deposit, bool) {
	var deposit types.ERC20Deposit
	found := k.getStore(ctx, k.chain).Get(pendingDepositPrefix.AppendStr(key.String()), &deposit)

	return deposit, found
}

// SetDeposit stores confirmed or burned deposits
func (k chainKeeper) SetDeposit(ctx sdk.Context, deposit types.ERC20Deposit, state types.DepositState) {
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
func (k chainKeeper) DeleteDeposit(ctx sdk.Context, deposit types.ERC20Deposit) {
	k.getStore(ctx, k.chain).Delete(confirmedDepositPrefix.AppendStr(deposit.TxID.Hex()).AppendStr(deposit.BurnerAddress.Hex()))
	k.getStore(ctx, k.chain).Delete(burnedDepositPrefix.AppendStr(deposit.TxID.Hex()).AppendStr(deposit.BurnerAddress.Hex()))
}

// SetPendingTransferKey stores a pending transfer ownership/operatorship
func (k chainKeeper) SetPendingTransferKey(ctx sdk.Context, key exported.PollKey, transferKey *types.TransferKey) {
	k.getStore(ctx, k.chain).Set(pendingTransferKeyPrefix.AppendStr(key.String()), transferKey)
}

// DeletePendingTransferKey deletes a pending transfer ownership/operatorship
func (k chainKeeper) DeletePendingTransferKey(ctx sdk.Context, key exported.PollKey) {
	k.getStore(ctx, k.chain).Delete(pendingTransferKeyPrefix.AppendStr(key.String()))
}

// ArchiveTransferKey archives an ownership transfer so it is no longer pending but can still be queried
func (k chainKeeper) ArchiveTransferKey(ctx sdk.Context, key exported.PollKey) {
	var transferKey types.TransferKey
	if !k.getStore(ctx, k.chain).Get(pendingTransferKeyPrefix.AppendStr(key.String()), &transferKey) {
		k.DeletePendingTransferKey(ctx, key)
		k.getStore(ctx, k.chain).Set(archivedTransferKeyPrefix.AppendStr(key.String()), &transferKey)
	}
}

// GetArchivedTransferKey returns an archived transfer of ownership/operatorship associated with the given poll
func (k chainKeeper) GetArchivedTransferKey(ctx sdk.Context, key exported.PollKey) (types.TransferKey, bool) {
	var transferKey types.TransferKey
	found := k.getStore(ctx, k.chain).Get(archivedTransferKeyPrefix.AppendStr(key.String()), &transferKey)

	return transferKey, found
}

// GetPendingTransferKey returns the transfer ownership/operatorship associated with the given poll
func (k chainKeeper) GetPendingTransferKey(ctx sdk.Context, key exported.PollKey) (types.TransferKey, bool) {
	var transferKey types.TransferKey
	found := k.getStore(ctx, k.chain).Get(pendingTransferKeyPrefix.AppendStr(key.String()), &transferKey)

	return transferKey, found
}

// GetNetworkByID returns the network name for a given chain and network ID
func (k chainKeeper) GetNetworkByID(ctx sdk.Context, id *big.Int) (string, bool) {
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
func (k chainKeeper) GetChainIDByNetwork(ctx sdk.Context, network string) *big.Int {
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

func (k chainKeeper) getUnsigned(ctx sdk.Context) types.CommandBatchMetadata {
	var unsigned types.CommandBatchMetadata
	if id := k.getStore(ctx, k.chain).GetRaw(unsignedBatchIDKey); id != nil {
		k.getStore(ctx, k.chain).Get(commandBatchPrefix.AppendStr(string(id)), &unsigned)
	}

	return unsigned
}

func (k chainKeeper) popCommand(ctx sdk.Context, filters ...func(value codec.ProtoMarshaler) bool) (types.Command, bool) {
	var cmd types.Command
	ok := k.getCommandQueue(ctx).Dequeue(&cmd, filters...)
	return cmd, ok
}

func (k chainKeeper) setCommandBatchMetadata(ctx sdk.Context, meta types.CommandBatchMetadata) {
	k.getStore(ctx, k.chain).Set(commandBatchPrefix.AppendStr(string(meta.ID)), &meta)
}

// GetBatchByID retrieves the specified batch if it exists
func (k chainKeeper) GetBatchByID(ctx sdk.Context, id []byte) types.CommandBatch {
	var batch types.CommandBatchMetadata
	k.getStore(ctx, k.chain).Get(commandBatchPrefix.AppendStr(string(id)), &batch)

	setter := func(m types.CommandBatchMetadata) {
		k.setCommandBatchMetadata(ctx, m)
	}

	return types.NewCommandBatch(batch, setter)
}

// GetLatestCommandBatch returns the latest batch of signed commands, if it exists
func (k chainKeeper) GetLatestCommandBatch(ctx sdk.Context) types.CommandBatch {
	if batch := k.getUnsigned(ctx); batch.Status != types.BatchNonExistent {
		setter := func(m types.CommandBatchMetadata) {}
		if batch.Status != types.BatchSigned {
			setter = func(m types.CommandBatchMetadata) {
				k.setCommandBatchMetadata(ctx, m)
			}
		}
		return types.NewCommandBatch(batch, setter)
	}

	var batch types.CommandBatchMetadata
	if id := k.getStore(ctx, k.chain).GetRaw(latestSignedBatchIDKey); id != nil {
		k.getStore(ctx, k.chain).Get(commandBatchPrefix.AppendStr(string(id)), &batch)
	}

	return types.NewCommandBatch(batch, func(types.CommandBatchMetadata) {})
}

// CreateNewBatchToSign creates a new batch of commands to be signed
func (k chainKeeper) CreateNewBatchToSign(ctx sdk.Context) ([]byte, error) {
	unsigned := k.getUnsigned(ctx)
	switch k.getUnsigned(ctx).Status {
	case types.BatchSigning:
		fallthrough
	case types.BatchAborted:
		return nil, fmt.Errorf("signing for command batch '%s' is still in progress", unsigned.ID)
	case types.BatchSigned:
		k.getStore(ctx, k.chain).SetRaw(latestSignedBatchIDKey, unsigned.ID)
	default:
		// first batch to be created
	}

	command, ok := k.popCommand(ctx)
	if !ok {
		return nil, fmt.Errorf("no commands to sign found")
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
	// TODO: limit the number of commands that are signed each time to avoid going above the gas limit
	for {
		cmd, ok := k.popCommand(ctx, filter)
		if !ok {
			break
		}
		commands = append(commands, cmd.Clone())
	}

	batchedCommands, err := types.NewCommandBatchMetadata(chainID.BigInt(), keyID, commands)
	if err != nil {
		return nil, err
	}

	if latestSignedBatchedCommands := k.GetLatestCommandBatch(ctx); latestSignedBatchedCommands.Is(types.BatchSigned) {
		batchedCommands.PrevBatchedCommandsID = latestSignedBatchedCommands.GetID()
	}

	k.setCommandBatchMetadata(ctx, batchedCommands)
	k.getStore(ctx, k.chain).SetRaw(unsignedBatchIDKey, batchedCommands.ID)

	return batchedCommands.ID, nil
}

// returns the queue of commands
func (k chainKeeper) getCommandQueue(ctx sdk.Context) utils.KVQueue {
	return utils.NewBlockHeightKVQueue(commandQueueName, k.getStore(ctx, k.chain), ctx.BlockHeight(), k.Logger(ctx))
}

func (k chainKeeper) setTokenMetadata(ctx sdk.Context, asset string, meta types.ERC20TokenMetadata) {
	key := tokenMetadataPrefix.Append(utils.LowerCaseKey(asset))
	k.getStore(ctx, k.chain).Set(key, &meta)
}

func (k chainKeeper) getTokenMetadata(ctx sdk.Context, asset string) (types.ERC20TokenMetadata, bool) {
	var result types.ERC20TokenMetadata
	key := tokenMetadataPrefix.Append(utils.LowerCaseKey(asset))
	found := k.getStore(ctx, k.chain).Get(key, &result)

	return result, found
}

func (k chainKeeper) initTokenMetadata(ctx sdk.Context, asset string, details types.TokenDetails) (types.ERC20TokenMetadata, error) {
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

	chainID := k.getSigner(ctx).ChainID()
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
