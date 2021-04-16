package keeper

import (
	"encoding/binary"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/rpc/client"

	broadcast "github.com/axelarnetwork/axelar-core/x/broadcast/exported"
	"github.com/axelarnetwork/axelar-core/x/broadcast/types"
)

var _ broadcast.Broadcaster = Keeper{}

const (
	proxyCountKey = "proxyCount"
	seqNoKey      = "seqNo"
)

// Keeper - the broadcast keeper
type Keeper struct {
	staker          types.Staker
	storeKey        sdk.StoreKey
	from            sdk.AccAddress
	kr              keyring.Keyring
	authKeeper      authkeeper.AccountKeeper
	config          types.ClientConfig
	rpc             client.ABCIClient
	fromName        string
	subjectiveStore sdk.KVStore
	cdc             *codec.LegacyAmino
}

// NewKeeper constructs a broadcast keeper
func NewKeeper(
	cdc *codec.LegacyAmino,
	storeKey sdk.StoreKey,
	subjectiveStore sdk.KVStore,
	kr keyring.Keyring,
	authKeeper authkeeper.AccountKeeper,
	stakingKeeper types.Staker,
	client client.ABCIClient,
	conf types.ClientConfig,
	logger log.Logger,
) (Keeper, error) {
	logger.With("module", fmt.Sprintf("x/%s", types.ModuleName)).Debug("creating broadcast keeper")
	from, fromName, err := types.GetAccountAddress(conf.From, kr)
	if err != nil {
		return Keeper{}, err
	}
	logger.With("module", fmt.Sprintf("x/%s", types.ModuleName)).Debug("broadcast keeper created")
	return Keeper{
		subjectiveStore: subjectiveStore,
		staker:          stakingKeeper,
		storeKey:        storeKey,
		from:            from,
		kr:              kr,
		authKeeper:      authKeeper,
		cdc:             cdc,
		config:          conf,
		rpc:             client,
		fromName:        fromName,
	}, nil
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// RegisterProxy registers a proxy address for a given principal, which can broadcast messages in the principal's name
func (k Keeper) RegisterProxy(ctx sdk.Context, principal sdk.ValAddress, proxy sdk.AccAddress) error {
	val := k.staker.Validator(ctx, principal)
	if val == nil {
		return fmt.Errorf("validator %s is unknown", principal.String())
	}
	k.Logger(ctx).Debug("getting proxy count")
	count := k.getProxyCount(ctx)

	storedProxy := ctx.KVStore(k.storeKey).Get(principal)
	if storedProxy != nil {
		ctx.KVStore(k.storeKey).Delete(storedProxy)
		count -= 1
	}
	k.Logger(ctx).Debug("setting proxy")
	ctx.KVStore(k.storeKey).Set(proxy, principal)
	// Creating a reverse lookup
	ctx.KVStore(k.storeKey).Set(principal, proxy)
	count += 1
	k.Logger(ctx).Debug("setting proxy count")
	k.setProxyCount(ctx, count)
	k.Logger(ctx).Debug("done")
	return nil
}

// GetLocalPrincipal returns the address of the local validator account. Returns nil if not set.
//
// WARNING: Handle with care, this call is non-deterministic because it exposes local information that is DIFFERENT for each validator
func (k Keeper) GetLocalPrincipal(ctx sdk.Context) sdk.ValAddress {
	return k.GetPrincipal(ctx, k.from)
}

// GetPrincipal returns the proxy address for a given principal address. Returns nil if not set.
func (k Keeper) GetPrincipal(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress {
	if proxy == nil {
		return nil
	}
	return ctx.KVStore(k.storeKey).Get(proxy)
}

// GetProxy returns the proxy address for a given principal address. Returns nil if not set.
func (k Keeper) GetProxy(ctx sdk.Context, principal sdk.ValAddress) sdk.AccAddress {
	return ctx.KVStore(k.storeKey).Get(principal)
}

func (k Keeper) setProxyCount(ctx sdk.Context, count int) {
	k.Logger(ctx).Debug(fmt.Sprintf("number of known proxies: %v", count))
	ctx.KVStore(k.storeKey).Set([]byte(proxyCountKey), k.cdc.MustMarshalBinaryLengthPrefixed(count))
}

func (k Keeper) getProxyCount(ctx sdk.Context) int {
	bz := ctx.KVStore(k.storeKey).Get([]byte(proxyCountKey))
	if bz == nil {
		return 0
	}
	var count int
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &count)
	return count
}

func (k Keeper) prepareMsgForSigning(ctx sdk.Context, msgs []sdk.Msg) (legacytx.StdSignMsg, error) {
	if k.config.ChainID == "" {
		return legacytx.StdSignMsg{}, sdkerrors.Wrap(types.ErrInvalidChain, "chain ID required but not specified")
	}

	acc := k.authKeeper.GetAccount(ctx, k.from)
	seqNo := k.getSeqNo()
	if acc.GetSequence() > seqNo {
		seqNo = acc.GetSequence()
	}

	return legacytx.StdSignMsg{
		ChainID:       k.config.ChainID,
		AccountNumber: acc.GetAccountNumber(),
		Sequence:      seqNo,
		Msgs:          msgs,
		Fee:           legacytx.NewStdFee(10000000, nil),
	}, nil
}

func (k Keeper) sign(msg legacytx.StdSignMsg) (legacytx.StdTx, error) {
	sig, err := k.makeSignature(msg)
	if err != nil {
		return legacytx.StdTx{}, err
	}

	return legacytx.NewStdTx(msg.Msgs, msg.Fee, []legacytx.StdSignature{sig}, msg.Memo), nil
}

func (k Keeper) makeSignature(msg legacytx.StdSignMsg) (legacytx.StdSignature, error) {
	sigBytes, pubkey, err := k.kr.Sign(k.fromName, msg.Bytes())
	if err != nil {
		return legacytx.StdSignature{}, err
	}

	return legacytx.StdSignature{
		PubKey:    pubkey,
		Signature: sigBytes,
	}, nil
}

func (k Keeper) getSeqNo() uint64 {
	seqNo := k.subjectiveStore.Get([]byte(seqNoKey))
	if seqNo == nil {
		return 0
	}
	return binary.LittleEndian.Uint64(seqNo)
}

func (k Keeper) setSeqNo(seqNo uint64) {
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, seqNo)
	k.subjectiveStore.Set([]byte(seqNoKey), bz)
}
