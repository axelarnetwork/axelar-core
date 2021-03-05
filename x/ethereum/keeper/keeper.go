package keeper

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tendermint/tendermint/libs/log"

	ethTypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/axelarnetwork/axelar-core/x/ethereum/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
)

const (
	gatewayKey = "gateway"

	rawPrefix                  = "raw_"
	verifiedTokenPrefix        = "verified_token_"
	pendingTokenPrefix         = "pending_token_"
	pendingErc20DepositPrefix  = "pending_erc20_deposit_"
	verifiedErc20DepositPrefix = "verified_erc20_deposit_"
	archivedErc20DepositPrefix = "archived_erc20_deposit_"
	commandPrefix              = "command_"
	symbolPrefix               = "symbol_"
	burnerPrefix               = "burner_"
	tokenPrefix                = "token_"
)

// Keeper represents the ethereum keeper
type Keeper struct {
	storeKey sdk.StoreKey
	cdc      *codec.Codec
	params   params.Subspace
}

// NewEthKeeper returns a new ethereum keeper
func NewEthKeeper(cdc *codec.Codec, storeKey sdk.StoreKey, paramSpace params.Subspace) Keeper {
	return Keeper{cdc: cdc, storeKey: storeKey, params: paramSpace.WithKeyTable(types.KeyTable())}
}

// SetParams sets the eth module's parameters
func (k Keeper) SetParams(ctx sdk.Context, p types.Params) {
	k.params.SetParamSet(ctx, &p)
}

// GetParams gets the eth module's parameters
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	var p types.Params
	k.params.GetParamSet(ctx, &p)
	return p
}

// GetNetwork returns the Ethereum network Axelar-Core is expected to connect to
func (k Keeper) GetNetwork(ctx sdk.Context) types.Network {
	var network types.Network
	k.params.Get(ctx, types.KeyNetwork, &network)
	return network
}

// GetERC20TokenDeploySignature returns the signature of the ERC20 transfer method
func (k Keeper) GetERC20TokenDeploySignature(ctx sdk.Context) common.Hash {
	var tokenSig []byte
	k.params.Get(ctx, types.KeyTokenDeploySig, &tokenSig)
	return common.BytesToHash(tokenSig)
}

// Codec returns the codec
func (k Keeper) Codec() *codec.Codec {
	return k.cdc
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetRequiredConfirmationHeight returns the required block confirmation height
func (k Keeper) GetRequiredConfirmationHeight(ctx sdk.Context) uint64 {
	var h uint64
	k.params.Get(ctx, types.KeyConfirmationHeight, &h)
	return h
}

// SetGatewayAddress sets the contract address for Axelar Gateway
func (k Keeper) SetGatewayAddress(ctx sdk.Context, addr common.Address) {
	ctx.KVStore(k.storeKey).Set([]byte(gatewayKey), addr.Bytes())
}

// GetGatewayAddress gets the contract address for Axelar Gateway
func (k Keeper) GetGatewayAddress(ctx sdk.Context) (common.Address, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(gatewayKey))
	if bz == nil {
		return common.Address{}, false
	}
	return common.BytesToAddress(bz), true
}

// SetBurnerInfo saves the burner info for a given address
func (k Keeper) SetBurnerInfo(ctx sdk.Context, burnerAddr common.Address, burnerInfo *types.BurnerInfo) {
	key := append([]byte(burnerPrefix), burnerAddr.Bytes()...)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(burnerInfo)

	ctx.KVStore(k.storeKey).Set(key, bz)
}

// GetBurnerInfo retrieves the burner info for a given address
func (k Keeper) GetBurnerInfo(ctx sdk.Context, burnerAddr common.Address) *types.BurnerInfo {
	key := append([]byte(burnerPrefix), burnerAddr.Bytes()...)

	bz := ctx.KVStore(k.storeKey).Get(key)
	if bz == nil {
		return nil
	}

	var result *types.BurnerInfo
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &result)

	return result
}

// GetTokenAddress calculates the token address given symbol and axelar gateway address
func (k Keeper) GetTokenAddress(ctx sdk.Context, symbol string, gatewayAddr common.Address) (common.Address, error) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(tokenPrefix + symbol))
	if bz != nil {
		return common.BytesToAddress(bz), nil
	}

	tokenInfo := k.GetTokenInfo(ctx, symbol)
	if tokenInfo == nil {
		return common.Address{}, fmt.Errorf("symbol not found/verified")
	}

	var saltToken [32]byte
	copy(saltToken[:], crypto.Keccak256Hash([]byte(symbol)).Bytes())

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
	packed, err := arguments.Pack(tokenInfo.TokenName, symbol, tokenInfo.Decimals, tokenInfo.Capacity.BigInt())
	if err != nil {
		return common.Address{}, err
	}

	tokenInitCode := append(k.getTokenBC(ctx), packed...)
	tokenInitCodeHash := crypto.Keccak256Hash(tokenInitCode)

	tokenAddr := crypto.CreateAddress2(gatewayAddr, saltToken, tokenInitCodeHash.Bytes())
	ctx.KVStore(k.storeKey).Set([]byte(tokenPrefix+symbol), tokenAddr.Bytes())
	return tokenAddr, nil
}

// GetBurnerAddressAndSalt calculates a burner address and the corresponding salt for the given token address, recipient and axelar gateway address
func (k Keeper) GetBurnerAddressAndSalt(ctx sdk.Context, tokenAddr common.Address, recipient string, gatewayAddr common.Address) (common.Address, common.Hash, error) {
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

	burnerInitCode := append(k.getBurnerBC(ctx), packed...)
	burnerInitCodeHash := crypto.Keccak256Hash(burnerInitCode)

	return crypto.CreateAddress2(gatewayAddr, saltBurn, burnerInitCodeHash.Bytes()), saltBurn, nil
}

func (k Keeper) getBurnerBC(ctx sdk.Context) []byte {
	var b []byte
	k.params.Get(ctx, types.KeyBurnable, &b)
	return b
}

func (k Keeper) getTokenBC(ctx sdk.Context) []byte {
	var b []byte
	k.params.Get(ctx, types.KeyToken, &b)
	return b
}

// GetGatewayByteCodes retrieves the byte codes for the Axelar Gateway smart contract
func (k Keeper) GetGatewayByteCodes(ctx sdk.Context) []byte {
	var b []byte
	k.params.Get(ctx, types.KeyGateway, &b)
	return b
}

// SetUnverifiedErc20TokenDeploy stores and unverified erc20 token
func (k Keeper) SetUnverifiedErc20TokenDeploy(ctx sdk.Context, token *types.Erc20TokenDeploy) {
	txID := common.BytesToHash(token.TxID[:]).String()
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(token)
	ctx.KVStore(k.storeKey).Set([]byte(pendingTokenPrefix+txID), bz)
}

// SetTokenInfo stores the token info
func (k Keeper) SetTokenInfo(ctx sdk.Context, msg types.MsgSignDeployToken) {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(msg)
	ctx.KVStore(k.storeKey).Set([]byte(symbolPrefix+msg.Symbol), bz)
}

// GetTokenInfo retrieves the token info
func (k Keeper) GetTokenInfo(ctx sdk.Context, symbol string) *types.MsgSignDeployToken {
	bz := ctx.KVStore(k.storeKey).Get([]byte(symbolPrefix + symbol))
	if bz == nil {
		return nil
	}
	var msg *types.MsgSignDeployToken
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &msg)

	return msg
}

// SetCommandData stores command data by ID
func (k Keeper) SetCommandData(ctx sdk.Context, commandID types.CommandID, commandData []byte) {
	key := append([]byte(commandPrefix), commandID[:]...)

	ctx.KVStore(k.storeKey).Set(key, commandData)
}

// GetCommandData retrieves command data by ID
func (k Keeper) GetCommandData(ctx sdk.Context, commandID types.CommandID) []byte {
	key := append([]byte(commandPrefix), commandID[:]...)

	return ctx.KVStore(k.storeKey).Get(key)
}

func (k Keeper) getRawTx(ctx sdk.Context, txID string) *ethTypes.Transaction {
	bz := ctx.KVStore(k.storeKey).Get([]byte(rawPrefix + txID))
	if bz == nil {
		return nil
	}
	var tx *ethTypes.Transaction
	k.cdc.MustUnmarshalJSON(bz, &tx)

	return tx
}

// SetRawTx stores a raw transaction by hash
func (k Keeper) SetRawTx(ctx sdk.Context, txID string, tx *ethTypes.Transaction) {
	bz := k.cdc.MustMarshalJSON(tx)
	ctx.KVStore(k.storeKey).Set([]byte(rawPrefix+txID), bz)
}

// HasUnverifiedErc20Deposit returns true if an unverified erc20 deposit has been stored
func (k Keeper) HasUnverifiedErc20Deposit(ctx sdk.Context, txID string) bool {
	return ctx.KVStore(k.storeKey).Has([]byte(pendingErc20DepositPrefix + txID))
}

// SetUnverifiedErc20Deposit stores and unverified erc20 deposit
func (k Keeper) SetUnverifiedErc20Deposit(ctx sdk.Context, txID string, deposit *types.Erc20Deposit) {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(deposit)
	ctx.KVStore(k.storeKey).Set([]byte(pendingErc20DepositPrefix+txID), bz)
}

// GetVerifiedErc20Deposit retrieves the verified erc20 deposit given txID if found
func (k Keeper) GetVerifiedErc20Deposit(ctx sdk.Context, txID string) *types.Erc20Deposit {
	bz := ctx.KVStore(k.storeKey).Get([]byte(verifiedErc20DepositPrefix + txID))
	if bz == nil {
		return nil
	}

	var result *types.Erc20Deposit
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &result)

	return result
}

// GetVerifiedErc20Deposits retrieves all the verified erc20 deposits
func (k Keeper) GetVerifiedErc20Deposits(ctx sdk.Context) []types.Erc20Deposit {
	var deposits []types.Erc20Deposit
	iter := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), []byte(verifiedErc20DepositPrefix))

	for ; iter.Valid(); iter.Next() {
		bz := iter.Value()

		var deposit types.Erc20Deposit
		k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &deposit)
		deposits = append(deposits, deposit)
	}

	return deposits
}

// ArchiveErc20Deposit marks a deposit as archived
func (k Keeper) ArchiveErc20Deposit(ctx sdk.Context, txID string) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(verifiedErc20DepositPrefix + txID))
	if bz == nil {
		return
	}

	ctx.KVStore(k.storeKey).Delete([]byte(verifiedErc20DepositPrefix + txID))
	ctx.KVStore(k.storeKey).Set([]byte(archivedErc20DepositPrefix+txID), bz)
}

// HasUnverifiedToken returns true if an unverified transaction has been stored
func (k Keeper) HasUnverifiedToken(ctx sdk.Context, txID string) bool {
	return ctx.KVStore(k.storeKey).Has([]byte(pendingTokenPrefix + txID))
}

// ProcessVerificationTokenResult stores the TX permanently if confirmed or discards the data otherwise
func (k Keeper) ProcessVerificationTokenResult(ctx sdk.Context, txID string, verified bool) {
	ok := k.processVerificationResult(ctx, []byte(pendingTokenPrefix+txID), []byte(verifiedTokenPrefix+txID), verified)
	if !ok {
		k.Logger(ctx).Debug(fmt.Sprintf("tx %s not found", txID))
	}
}

// ProcessVerificationErc20DepositResult stores the TX permanently if confirmed or discards the data otherwise
func (k Keeper) ProcessVerificationErc20DepositResult(ctx sdk.Context, txID string, verified bool) {
	ok := k.processVerificationResult(ctx, []byte(pendingErc20DepositPrefix+txID), []byte(verifiedErc20DepositPrefix+txID), verified)
	if !ok {
		k.Logger(ctx).Debug(fmt.Sprintf("tx %s not found", txID))
	}
}

func (k Keeper) processVerificationResult(ctx sdk.Context, pendingKey, verifiedKey []byte, verified bool) bool {
	bz := ctx.KVStore(k.storeKey).Get(pendingKey)
	if bz == nil {
		return false
	}
	if verified {
		ctx.KVStore(k.storeKey).Set(verifiedKey, bz)
	}
	ctx.KVStore(k.storeKey).Delete(pendingKey)
	return true
}

// AssembleEthTx sets a signature for a previously stored raw transaction
func (k Keeper) AssembleEthTx(ctx sdk.Context, txID string, pk ecdsa.PublicKey, sig tss.Signature) (*ethTypes.Transaction, error) {
	rawTx := k.getRawTx(ctx, txID)
	if rawTx == nil {
		return nil, fmt.Errorf("raw tx for ID %s has not been prepared yet", txID)
	}

	signer := k.getSigner(ctx)

	recoverableSig, err := types.ToEthSignature(sig, signer.Hash(rawTx), pk)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrEthereum, fmt.Sprintf("could not create recoverable signature: %v", err))
	}

	return rawTx.WithSignature(signer, recoverableSig[:])
}

// GetHashToSign returns the hash to sign of a previously stored raw transaction
func (k Keeper) GetHashToSign(ctx sdk.Context, txID string) (common.Hash, error) {
	rawTx := k.getRawTx(ctx, txID)
	if rawTx == nil {
		return common.Hash{}, fmt.Errorf("raw tx with id %s not found", txID)
	}
	signer := k.getSigner(ctx)
	return signer.Hash(rawTx), nil
}

func (k Keeper) getSigner(ctx sdk.Context) ethTypes.EIP155Signer {
	var network types.Network
	k.params.Get(ctx, types.KeyNetwork, &network)
	return ethTypes.NewEIP155Signer(network.Params().ChainID)
}
