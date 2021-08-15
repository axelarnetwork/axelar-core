package keeper

import (
	"crypto/ecdsa"
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

const (
	gatewayKey      = "gateway"
	pendingChainKey = "pending_chain_asset"

	chainPrefix            = "chain_"
	subspacePrefix         = "subspace_"
	unsignedPrefix         = "unsigned_"
	pendingTokenPrefix     = "pending_token_"
	pendingDepositPrefix   = "pending_deposit_"
	confirmedDepositPrefix = "confirmed_deposit_"
	burnedDepositPrefix    = "burned_deposit_"
	commandPrefix          = "command_"
	assetPrefix            = "asset_"
	burnerAddrPrefix       = "burnerAddr_"
	tokenAddrPrefix        = "tokenAddr_"

	pendingTransferOwnershipPrefix  = "pending_transfer_ownership_"
	archivedTransferOwnershipPrefix = "archived_transfer_ownership_"
)

var _ types.BaseKeeper = keeper{}
var _ types.ChainKeeper = keeper{}

// Keeper implements both the base keeper and chain keeper
type keeper struct {
	chain        string
	storeKey     sdk.StoreKey
	cdc          codec.BinaryMarshaler
	paramsKeeper types.ParamsKeeper
	subspaces    map[string]params.Subspace
}

// NewKeeper returns a new EVM base keeper
func NewKeeper(cdc codec.BinaryMarshaler, storeKey sdk.StoreKey, paramsKeeper types.ParamsKeeper) types.BaseKeeper {
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

// GetChain returns the keeper associated to the given chain
func (k keeper) ForChain(ctx sdk.Context, chain string) types.ChainKeeper {
	k.chain = strings.ToLower(chain)
	return k
}

// SetPendingChain stores the chain pending for confirmation
func (k keeper) SetPendingChain(ctx sdk.Context, chain nexus.Chain) {
	k.getStore(ctx, chain.Name).Set([]byte(pendingChainKey), k.cdc.MustMarshalBinaryLengthPrefixed(&chain))
}

// GetPendingChain returns the chain object with the given name, false if the chain is either unknown or confirmed
func (k keeper) GetPendingChain(ctx sdk.Context, chainName string) (nexus.Chain, bool) {
	bz := k.getStore(ctx, chainName).Get([]byte(pendingChainKey))
	if bz == nil {
		return nexus.Chain{}, false
	}
	var chain nexus.Chain
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &chain)
	return chain, true
}

// DeletePendingChain deletes a chain that is not registered yet
func (k keeper) DeletePendingChain(ctx sdk.Context, chain string) {
	k.getStore(ctx, chain).Delete([]byte(pendingChainKey))
}

// SetParams sets the evm module's parameters
func (k keeper) SetParams(ctx sdk.Context, params ...types.Params) {
	for _, p := range params {
		chain := strings.ToLower(p.Chain)

		// set the chain before calling the subspace so it is recognized as an existing chain
		ctx.KVStore(k.storeKey).Set([]byte(subspacePrefix+chain), []byte(p.Chain))
		subspace, _ := k.getSubspace(ctx, chain)
		subspace.SetParamSet(ctx, &p)
	}
}

// GetParams gets the evm module's parameters
func (k keeper) GetParams(ctx sdk.Context) []types.Params {
	params := make([]types.Params, 0)
	iter := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), []byte(subspacePrefix))
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

// SetGatewayAddress sets the contract address for Axelar Gateway
func (k keeper) SetGatewayAddress(ctx sdk.Context, addr common.Address) {
	k.getStore(ctx, k.chain).Set([]byte(gatewayKey), addr.Bytes())
}

// GetGatewayAddress gets the contract address for Axelar Gateway
func (k keeper) GetGatewayAddress(ctx sdk.Context) (common.Address, bool) {
	bz := k.getStore(ctx, k.chain).Get([]byte(gatewayKey))
	if bz == nil {
		return common.Address{}, false
	}
	return common.BytesToAddress(bz), true
}

// SetBurnerInfo saves the burner info for a given address
func (k keeper) SetBurnerInfo(ctx sdk.Context, burnerAddr common.Address, burnerInfo *types.BurnerInfo) {
	key := append([]byte(burnerAddrPrefix), burnerAddr.Bytes()...)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(burnerInfo)

	k.getStore(ctx, k.chain).Set(key, bz)
}

// GetBurnerInfo retrieves the burner info for a given address
func (k keeper) GetBurnerInfo(ctx sdk.Context, burnerAddr common.Address) *types.BurnerInfo {
	key := append([]byte(burnerAddrPrefix), burnerAddr.Bytes()...)

	bz := k.getStore(ctx, k.chain).Get(key)
	if bz == nil {
		return nil
	}

	var result types.BurnerInfo
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &result)

	return &result
}

// GetTokenSymbol returns the symbol for a given token
func (k keeper) GetTokenSymbol(ctx sdk.Context, assetName string) (string, bool) {
	tokenInfo := k.getTokenInfo(ctx, assetName)
	if tokenInfo == nil {
		return "", false
	}

	return tokenInfo.Symbol, true
}

// GetTokenAddress calculates the token address for some asset with the provided axelar gateway address
func (k keeper) GetTokenAddress(ctx sdk.Context, assetName string, gatewayAddr common.Address) (common.Address, error) {
	assetName = strings.ToLower(assetName)

	bz := k.getStore(ctx, k.chain).Get([]byte(tokenAddrPrefix + assetName))
	if bz != nil {
		return common.BytesToAddress(bz), nil
	}

	tokenInfo := k.getTokenInfo(ctx, assetName)
	if tokenInfo == nil {
		return common.Address{}, fmt.Errorf("symbol not found")
	}

	var saltToken [32]byte
	copy(saltToken[:], crypto.Keccak256Hash([]byte(tokenInfo.Symbol)).Bytes())

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
	packed, err := arguments.Pack(tokenInfo.TokenName, tokenInfo.Symbol, tokenInfo.Decimals, tokenInfo.Capacity.BigInt())
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
	k.getStore(ctx, k.chain).Set([]byte(tokenAddrPrefix+assetName), tokenAddr.Bytes())
	return tokenAddr, nil
}

// GetBurnerAddressAndSalt calculates a burner address and the corresponding salt for the given token address, recipient and axelar gateway address
func (k keeper) GetBurnerAddressAndSalt(ctx sdk.Context, tokenAddr common.Address, recipient string, gatewayAddr common.Address) (common.Address, common.Hash, error) {
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

// SetPendingTokenDeployment stores a pending ERC20 token deployment
func (k keeper) SetPendingTokenDeployment(ctx sdk.Context, key exported.PollKey, token types.ERC20TokenDeployment) {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(&token)
	k.getStore(ctx, k.chain).Set([]byte(pendingTokenPrefix+key.String()), bz)
}

// SetTokenInfo stores the token info
func (k keeper) SetTokenInfo(ctx sdk.Context, assetName string, msg *types.SignDeployTokenRequest) {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(msg)
	k.getStore(ctx, k.chain).Set([]byte(assetPrefix+strings.ToLower(assetName)), bz)
}

// getTokenInfo retrieves the token info
func (k keeper) getTokenInfo(ctx sdk.Context, assetName string) *types.SignDeployTokenRequest {
	bz := k.getStore(ctx, k.chain).Get([]byte(assetPrefix + strings.ToLower(assetName)))
	if bz == nil {
		return nil
	}
	var msg types.SignDeployTokenRequest
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &msg)

	return &msg
}

// SetCommandData stores command data by ID
func (k keeper) SetCommandData(ctx sdk.Context, commandID types.CommandID, commandData []byte) {
	key := append([]byte(commandPrefix), commandID[:]...)

	k.getStore(ctx, k.chain).Set(key, commandData)
}

// GetCommandData retrieves command data by ID
func (k keeper) GetCommandData(ctx sdk.Context, commandID types.CommandID) []byte {
	key := append([]byte(commandPrefix), commandID[:]...)

	return k.getStore(ctx, k.chain).Get(key)
}

func (k keeper) getUnsignedTx(ctx sdk.Context, txID string) *evmTypes.Transaction {
	bz := k.getStore(ctx, k.chain).Get([]byte(unsignedPrefix + txID))
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

	k.getStore(ctx, k.chain).Set([]byte(unsignedPrefix+txID), bz)
}

// SetPendingDeposit stores a pending deposit
func (k keeper) SetPendingDeposit(ctx sdk.Context, key exported.PollKey, deposit *types.ERC20Deposit) {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(deposit)
	k.getStore(ctx, k.chain).Set([]byte(pendingDepositPrefix+key.String()), bz)
}

// GetDeposit retrieves a confirmed/burned deposit
func (k keeper) GetDeposit(ctx sdk.Context, txID common.Hash, burnAddr common.Address) (types.ERC20Deposit, types.DepositState, bool) {
	var deposit types.ERC20Deposit

	bz := k.getStore(ctx, k.chain).Get([]byte(confirmedDepositPrefix + txID.Hex() + "_" + burnAddr.Hex()))
	if bz != nil {
		k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &deposit)
		return deposit, types.CONFIRMED, true
	}

	bz = k.getStore(ctx, k.chain).Get([]byte(burnedDepositPrefix + txID.Hex() + "_" + burnAddr.Hex()))
	if bz != nil {
		k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &deposit)
		return deposit, types.BURNED, true
	}

	return types.ERC20Deposit{}, 0, false
}

// GetConfirmedDeposits retrieves all the confirmed ERC20 deposits
func (k keeper) GetConfirmedDeposits(ctx sdk.Context) []types.ERC20Deposit {
	var deposits []types.ERC20Deposit
	iter := sdk.KVStorePrefixIterator(k.getStore(ctx, k.chain), []byte(confirmedDepositPrefix))
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		bz := iter.Value()

		var deposit types.ERC20Deposit
		k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &deposit)
		deposits = append(deposits, deposit)
	}

	return deposits
}

// AssembleTx sets a signature for a previously stored raw transaction
func (k keeper) AssembleTx(ctx sdk.Context, txID string, pk ecdsa.PublicKey, sig tss.Signature) (*evmTypes.Transaction, error) {
	rawTx := k.getUnsignedTx(ctx, txID)
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
	rawTx := k.getUnsignedTx(ctx, txID)
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

// DeletePendingToken deletes the token associated with the given poll
func (k keeper) DeletePendingToken(ctx sdk.Context, key exported.PollKey) {
	k.getStore(ctx, k.chain).Delete([]byte(pendingTokenPrefix + key.String()))
}

// GetPendingTokenDeployment returns the token associated with the given poll
func (k keeper) GetPendingTokenDeployment(ctx sdk.Context, key exported.PollKey) (types.ERC20TokenDeployment, bool) {
	bz := k.getStore(ctx, k.chain).Get([]byte(pendingTokenPrefix + key.String()))
	if bz == nil {
		return types.ERC20TokenDeployment{}, false
	}
	var tokenDeployment types.ERC20TokenDeployment
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &tokenDeployment)

	return tokenDeployment, true
}

// DeletePendingDeposit deletes the deposit associated with the given poll
func (k keeper) DeletePendingDeposit(ctx sdk.Context, key exported.PollKey) {
	k.getStore(ctx, k.chain).Delete([]byte(pendingDepositPrefix + key.String()))
}

// GetPendingDeposit returns the deposit associated with the given poll
func (k keeper) GetPendingDeposit(ctx sdk.Context, key exported.PollKey) (types.ERC20Deposit, bool) {
	bz := k.getStore(ctx, k.chain).Get([]byte(pendingDepositPrefix + key.String()))
	if bz == nil {
		return types.ERC20Deposit{}, false
	}
	var deposit types.ERC20Deposit
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &deposit)

	return deposit, true
}

// SetDeposit stores confirmed or burned deposits
func (k keeper) SetDeposit(ctx sdk.Context, deposit types.ERC20Deposit, state types.DepositState) {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(&deposit)

	switch state {
	case types.CONFIRMED:
		k.getStore(ctx, k.chain).Set([]byte(confirmedDepositPrefix+deposit.TxID.Hex()+"_"+deposit.BurnerAddress.Hex()), bz)
	case types.BURNED:
		k.getStore(ctx, k.chain).Set([]byte(burnedDepositPrefix+deposit.TxID.Hex()+"_"+deposit.BurnerAddress.Hex()), bz)
	default:
		panic("invalid deposit state")
	}
}

// DeleteDeposit deletes the given deposit
func (k keeper) DeleteDeposit(ctx sdk.Context, deposit types.ERC20Deposit) {
	k.getStore(ctx, k.chain).Delete([]byte(confirmedDepositPrefix + deposit.TxID.Hex() + "_" + deposit.BurnerAddress.Hex()))
	k.getStore(ctx, k.chain).Delete([]byte(burnedDepositPrefix + deposit.TxID.Hex() + "_" + deposit.BurnerAddress.Hex()))
}

// SetPendingTransferOwnership stores a pending transfer ownership
func (k keeper) SetPendingTransferOwnership(ctx sdk.Context, key exported.PollKey, transferOwnership *types.TransferOwnership) {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(transferOwnership)
	k.getStore(ctx, k.chain).Set([]byte(pendingTransferOwnershipPrefix+key.String()), bz)
}

// DeletePendingTransferOwnership deletes a pending transfer ownership
func (k keeper) DeletePendingTransferOwnership(ctx sdk.Context, key exported.PollKey) {
	k.getStore(ctx, k.chain).Delete([]byte(pendingTransferOwnershipPrefix + key.String()))
}

// ArchiveTransferOwnership archives an ownership transfer so it is no longer pending but can still be queried
func (k keeper) ArchiveTransferOwnership(ctx sdk.Context, key exported.PollKey) {
	transfer := k.getStore(ctx, k.chain).Get([]byte(pendingTransferOwnershipPrefix + key.String()))
	if transfer != nil {
		k.DeletePendingTransferOwnership(ctx, key)
		k.getStore(ctx, k.chain).Set([]byte(archivedTransferOwnershipPrefix+key.String()), transfer)
	}
}

// GetArchivedTransferOwnership returns an archived transfer of ownership associated with the given poll
func (k keeper) GetArchivedTransferOwnership(ctx sdk.Context, key exported.PollKey) (types.TransferOwnership, bool) {
	bz := k.getStore(ctx, k.chain).Get([]byte(archivedTransferOwnershipPrefix + key.String()))
	if bz == nil {
		return types.TransferOwnership{}, false
	}
	var transferOwnership types.TransferOwnership
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &transferOwnership)

	return transferOwnership, true
}

// GetPendingTransferOwnership returns the transfer ownership associated with the given poll
func (k keeper) GetPendingTransferOwnership(ctx sdk.Context, key exported.PollKey) (types.TransferOwnership, bool) {
	bz := k.getStore(ctx, k.chain).Get([]byte(pendingTransferOwnershipPrefix + key.String()))
	if bz == nil {
		return types.TransferOwnership{}, false
	}
	var transferOwnership types.TransferOwnership
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &transferOwnership)

	return transferOwnership, true
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

func (k keeper) getStore(ctx sdk.Context, chain string) prefix.Store {
	pre := []byte(chainPrefix + strings.ToLower(chain) + "_")
	return prefix.NewStore(ctx.KVStore(k.storeKey), pre)
}

func (k keeper) getSubspace(ctx sdk.Context, chain string) (params.Subspace, bool) {
	chainLower := strings.ToLower(chain)

	// When a node restarts or joins the network after genesis, it might not have all EVM subspaces initialized.
	// The following checks has to be done regardless, if we would only do it dependent on the existence of a subspace
	// different nodes would consume different amounts of gas and it would result in a consensus failure
	if !ctx.KVStore(k.storeKey).Has([]byte(subspacePrefix + chainLower)) {
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
