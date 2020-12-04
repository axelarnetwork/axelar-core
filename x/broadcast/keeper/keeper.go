package keeper

import (
	"encoding/binary"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/rpc/client/http"

	"github.com/axelarnetwork/axelar-core/store"
	broadcast "github.com/axelarnetwork/axelar-core/x/broadcast/exported"
	"github.com/axelarnetwork/axelar-core/x/broadcast/types"
)

var _ broadcast.Broadcaster = Keeper{}

const (
	proxyCountKey = "proxyCount"
	seqNoKey      = "seqNo"
)

type Keeper struct {
	stakingKeeper   types.Staker
	storeKey        sdk.StoreKey
	from            sdk.AccAddress
	keybase         keys.Keybase
	authKeeper      auth.AccountKeeper
	encodeTx        sdk.TxEncoder
	config          types.ClientConfig
	rpc             *http.HTTP
	fromName        string
	subjectiveStore store.SubjectiveStore
	cdc             *codec.Codec
}

// NewKeeper constructs a broadcast keeper
func NewKeeper(
	cdc *codec.Codec,
	storeKey sdk.StoreKey,
	subjectiveStore store.SubjectiveStore,
	keybase keys.Keybase,
	authKeeper auth.AccountKeeper,
	stakingKeeper types.Staker,
	conf types.ClientConfig,
	logger log.Logger,
) (Keeper, error) {
	logger.With("module", fmt.Sprintf("x/%s", types.ModuleName)).Debug("creating broadcast keeper")
	from, fromName, err := getAccountAddress(conf.From, keybase)
	if err != nil {
		return Keeper{}, err
	}
	rpc, err := http.New(conf.TendermintNodeUri, "/websocket")
	if err != nil {
		return Keeper{}, err
	}
	logger.With("module", fmt.Sprintf("x/%s", types.ModuleName)).Debug("broadcast keeper created")
	return Keeper{
		subjectiveStore: subjectiveStore,
		stakingKeeper:   stakingKeeper,
		storeKey:        storeKey,
		from:            from,
		keybase:         keybase,
		authKeeper:      authKeeper,
		encodeTx:        utils.GetTxEncoder(cdc),
		cdc:             cdc,
		config:          conf,
		rpc:             rpc,
		fromName:        fromName,
	}, nil
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// Broadcast sends the passed message to the network. Needs to be called asynchronously or it will block
func (k Keeper) BroadcastSync(ctx sdk.Context, valMsgs []broadcast.MsgWithSenderSetter) error {
	if k.GetLocalPrincipal(ctx) == nil {
		return fmt.Errorf("broadcaster is not registered as a proxy")
	}

	k.Logger(ctx).Debug("setting sender")
	msgs := make([]sdk.Msg, 0, len(valMsgs))
	for _, msg := range valMsgs {
		k.Logger(ctx).Debug(fmt.Sprintf("k.from: %v", k.from))
		msg.SetSender(k.from)
		msgs = append(msgs, msg)
	}
	k.Logger(ctx).Debug(fmt.Sprintf("preparing to sign:%v", msgs))
	stdSignMsg, err := k.prepareMsgForSigning(ctx, msgs)
	if err != nil {
		return err
	}

	k.Logger(ctx).Debug("signing")
	tx, err := k.sign(stdSignMsg)
	if err != nil {
		return err
	}

	k.Logger(ctx).Debug(fmt.Sprintf("from address: %v, acc no.: %d, seq no.: %d, chainId: %s", k.from, stdSignMsg.AccountNumber, stdSignMsg.Sequence, stdSignMsg.ChainID))

	k.Logger(ctx).Debug("encoding tx")
	txBytes, err := k.encodeTx(tx)
	if err != nil {
		k.Logger(ctx).Info(err.Error())
		return err
	}
	k.Logger(ctx).Debug("broadcasting")
	k.setSeqNo(stdSignMsg.Sequence + 1)
	res, err := k.rpc.BroadcastTxSync(txBytes)
	if err != nil {
		k.Logger(ctx).Error(err.Error())
	}
	if res != nil && res.Log != "" && res.Log != "[]" {
		k.Logger(ctx).Info("broadcast msg log: " + res.Log)
	}
	return nil
}

// RegisterProxy registers a proxy address for a given principal, which can broadcast messages in the principal's name
func (k Keeper) RegisterProxy(ctx sdk.Context, principal sdk.ValAddress, proxy sdk.AccAddress) error {
	_, ok := k.stakingKeeper.Validator(ctx, principal)
	if !ok {
		k.Logger(ctx).Error("could not find validator")
		return types.ErrInvalidValidator
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
	ctx.KVStore(k.storeKey).Set([]byte(proxyCountKey), k.cdc.MustMarshalBinaryBare(count))
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

func getAccountAddress(from string, keybase keys.Keybase) (sdk.AccAddress, string, error) {
	var info keys.Info
	if addr, err := sdk.AccAddressFromBech32(from); err == nil {
		info, err = keybase.GetByAddress(addr)
		if err != nil {
			return nil, "", err
		}
	} else {
		info, err = keybase.Get(from)
		if err != nil {
			return nil, "", err
		}
	}

	return info.GetAddress(), info.GetName(), nil
}

func (k Keeper) prepareMsgForSigning(ctx sdk.Context, msgs []sdk.Msg) (auth.StdSignMsg, error) {
	if k.config.ChainID == "" {
		return auth.StdSignMsg{}, sdkerrors.Wrap(types.ErrInvalidChain, "chain ID required but not specified")
	}

	acc := k.authKeeper.GetAccount(ctx, k.from)
	seqNo := k.getSeqNo()
	if acc.GetSequence() > seqNo {
		seqNo = acc.GetSequence()
	}

	return auth.StdSignMsg{
		ChainID:       k.config.ChainID,
		AccountNumber: acc.GetAccountNumber(),
		Sequence:      seqNo,
		Msgs:          msgs,
		Fee:           auth.NewStdFee(2000000, nil),
	}, nil
}

func (k Keeper) sign(msg auth.StdSignMsg) (auth.StdTx, error) {
	sig, err := k.makeSignature(msg)
	if err != nil {
		return auth.StdTx{}, err
	}

	return auth.NewStdTx(msg.Msgs, msg.Fee, []auth.StdSignature{sig}, msg.Memo), nil
}

func (k Keeper) makeSignature(msg auth.StdSignMsg) (auth.StdSignature, error) {
	sigBytes, pubkey, err := k.keybase.Sign(k.fromName, k.config.KeyringPassphrase, msg.Bytes())
	if err != nil {
		return auth.StdSignature{}, err
	}

	return auth.StdSignature{
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
