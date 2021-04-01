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
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

const (
	gatewayKey = "gateway"

	unsignedPrefix         = "unsigned_"
	pendingTokenPrefix     = "pending_token_"
	pendingDepositPrefix   = "pending_deposit_"
	confirmedDepositPrefix = "confirmed_deposit_"
	burnedDepositPrefix    = "burned_deposit_"
	commandPrefix          = "command_"
	symbolPrefix           = "symbol_"
	burnerAddrPrefix       = "burnerAddr_"
	tokenAddrPrefix        = "tokenAddr_"
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

// GetRevoteLockingPeriod returns the lock period for revoting
func (k Keeper) GetRevoteLockingPeriod(ctx sdk.Context) int64 {
	var result int64
	k.params.Get(ctx, types.KeyRevoteLockingPeriod, &result)

	return result
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
	key := append([]byte(burnerAddrPrefix), burnerAddr.Bytes()...)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(burnerInfo)

	ctx.KVStore(k.storeKey).Set(key, bz)
}

// GetBurnerInfo retrieves the burner info for a given address
func (k Keeper) GetBurnerInfo(ctx sdk.Context, burnerAddr common.Address) *types.BurnerInfo {
	key := append([]byte(burnerAddrPrefix), burnerAddr.Bytes()...)

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
	bz := ctx.KVStore(k.storeKey).Get([]byte(tokenAddrPrefix + symbol))
	if bz != nil {
		return common.BytesToAddress(bz), nil
	}

	tokenInfo := k.getTokenInfo(ctx, symbol)
	if tokenInfo == nil {
		return common.Address{}, fmt.Errorf("symbol not found/confirmed")
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
	ctx.KVStore(k.storeKey).Set([]byte(tokenAddrPrefix+symbol), tokenAddr.Bytes())
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

// SetPendingTokenDeploy stores a pending ERC20 token deployment
func (k Keeper) SetPendingTokenDeploy(ctx sdk.Context, poll exported.PollMeta, token types.ERC20TokenDeploy) {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(token)
	ctx.KVStore(k.storeKey).Set([]byte(pendingTokenPrefix+poll.String()), bz)
}

// SetTokenInfo stores the token info
func (k Keeper) SetTokenInfo(ctx sdk.Context, msg types.MsgSignDeployToken) {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(msg)
	ctx.KVStore(k.storeKey).Set([]byte(symbolPrefix+msg.Symbol), bz)
}

// getTokenInfo retrieves the token info
func (k Keeper) getTokenInfo(ctx sdk.Context, symbol string) *types.MsgSignDeployToken {
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

func (k Keeper) getUnsignedTx(ctx sdk.Context, txID string) *ethTypes.Transaction {
	bz := ctx.KVStore(k.storeKey).Get([]byte(unsignedPrefix + txID))
	if bz == nil {
		return nil
	}
	var tx *ethTypes.Transaction
	k.cdc.MustUnmarshalJSON(bz, &tx)

	return tx
}

// SetUnsignedTx stores an unsigned transaction by hash
func (k Keeper) SetUnsignedTx(ctx sdk.Context, txID string, tx *ethTypes.Transaction) {
	bz := k.cdc.MustMarshalJSON(tx)
	ctx.KVStore(k.storeKey).Set([]byte(unsignedPrefix+txID), bz)
}

// SetPendingDeposit stores a pending deposit
func (k Keeper) SetPendingDeposit(ctx sdk.Context, poll exported.PollMeta, deposit *types.ERC20Deposit) {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(deposit)
	ctx.KVStore(k.storeKey).Set([]byte(pendingDepositPrefix+poll.String()), bz)
}

// GetDeposit retrieves a confirmed/burned deposit
func (k Keeper) GetDeposit(ctx sdk.Context, txID string, burnAddr string) (types.ERC20Deposit, types.DepositState, bool) {
	var deposit types.ERC20Deposit
	prefixes := map[types.DepositState]string{
		types.CONFIRMED: confirmedDepositPrefix,
		types.BURNED:    burnedDepositPrefix,
	}

	// the order of this iteration is non-deterministic,
	// the only reason this is correct is because exactly one of the cases is true
	for state, prefix := range prefixes {
		bz := ctx.KVStore(k.storeKey).Get([]byte(prefix + txID + "_" + burnAddr))
		if bz != nil {
			k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &deposit)
			return deposit, state, true
		}

	}

	return types.ERC20Deposit{}, 0, false
}

// GetConfirmedDeposits retrieves all the confirmed ERC20 deposits
func (k Keeper) GetConfirmedDeposits(ctx sdk.Context) []types.ERC20Deposit {
	var deposits []types.ERC20Deposit
	iter := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), []byte(confirmedDepositPrefix))

	for ; iter.Valid(); iter.Next() {
		bz := iter.Value()

		var deposit types.ERC20Deposit
		k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &deposit)
		deposits = append(deposits, deposit)
	}

	return deposits
}

// AssembleEthTx sets a signature for a previously stored raw transaction
func (k Keeper) AssembleEthTx(ctx sdk.Context, txID string, pk ecdsa.PublicKey, sig tss.Signature) (*ethTypes.Transaction, error) {
	rawTx := k.getUnsignedTx(ctx, txID)
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
	rawTx := k.getUnsignedTx(ctx, txID)
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

// DeletePendingToken deletes the token associated with the given poll
func (k Keeper) DeletePendingToken(ctx sdk.Context, poll exported.PollMeta) {
	ctx.KVStore(k.storeKey).Delete([]byte(pendingTokenPrefix + poll.String()))
}

// GetPendingTokenDeploy returns the token associated with the given poll
func (k Keeper) GetPendingTokenDeploy(ctx sdk.Context, poll exported.PollMeta) (types.ERC20TokenDeploy, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(pendingTokenPrefix + poll.String()))
	if bz == nil {
		return types.ERC20TokenDeploy{}, false
	}
	var tokenDeploy types.ERC20TokenDeploy
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &tokenDeploy)

	return tokenDeploy, true
}

// DeletePendingDeposit deletes the deposit associated with the given poll
func (k Keeper) DeletePendingDeposit(ctx sdk.Context, poll exported.PollMeta) {
	ctx.KVStore(k.storeKey).Delete([]byte(pendingTokenPrefix + poll.String()))
}

// GetPendingDeposit returns the deposit associated with the given poll
func (k Keeper) GetPendingDeposit(ctx sdk.Context, poll exported.PollMeta) (types.ERC20Deposit, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(pendingDepositPrefix + poll.String()))
	if bz == nil {
		return types.ERC20Deposit{}, false
	}
	var deposit types.ERC20Deposit
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &deposit)

	return deposit, true
}

// SetDeposit stores confirmed or burned deposits
func (k Keeper) SetDeposit(ctx sdk.Context, deposit types.ERC20Deposit, state types.DepositState) {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(deposit)

	switch state {
	case types.CONFIRMED:
		ctx.KVStore(k.storeKey).Set([]byte(confirmedDepositPrefix+deposit.TxID.Hex()+"_"+deposit.BurnerAddr), bz)
	case types.BURNED:
		ctx.KVStore(k.storeKey).Set([]byte(burnedDepositPrefix+deposit.TxID.Hex()+"_"+deposit.BurnerAddr), bz)
	default:
		panic("invalid deposit state")
	}
}

// DeleteDeposit deletes the given deposit
func (k Keeper) DeleteDeposit(ctx sdk.Context, deposit types.ERC20Deposit) {
	ctx.KVStore(k.storeKey).Delete([]byte(confirmedDepositPrefix + deposit.TxID.Hex() + "_" + deposit.BurnerAddr))
	ctx.KVStore(k.storeKey).Delete([]byte(burnedDepositPrefix + deposit.TxID.Hex() + "_" + deposit.BurnerAddr))
}
