package keeper

import (
	"encoding/binary"
	"fmt"

	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/rpc/client/http"

	"github.com/axelarnetwork/axelar-core/x/broadcast/exported"
	"github.com/axelarnetwork/axelar-core/x/broadcast/types"
)

var _ exported.Broadcaster = Keeper{}

const (
	proxyCountKey = "proxyCount"
	seqNoKey      = "seqNo"
)

type Keeper struct {
	stakingKeeper staking.Keeper
	storeKey      sdk.StoreKey
	from          sdk.AccAddress
	keybase       keys.Keybase
	keeper        auth.AccountKeeper
	encodeTx      sdk.TxEncoder
	config        types.ClientConfig
	rpc           *http.HTTP
	fromName      string
}

func NewKeeper(
	conf types.ClientConfig,
	storeKey sdk.StoreKey,
	keybase keys.Keybase,
	authKeeper auth.AccountKeeper,
	stakingKeeper staking.Keeper,
	encoder sdk.TxEncoder,
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
		stakingKeeper: stakingKeeper,
		storeKey:      storeKey,
		from:          from,
		keybase:       keybase,
		keeper:        authKeeper,
		encodeTx:      encoder,
		config:        conf,
		rpc:           rpc,
		fromName:      fromName,
	}, nil
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
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

func (k Keeper) Broadcast(ctx sdk.Context, valMsgs []exported.ValidatorMsg) error {
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
	k.setSeqNo(ctx, stdSignMsg.Sequence+1)
	go func() {
		k.Logger(ctx).Debug("inside broadcasting goroutine")
		res, err := k.rpc.BroadcastTxSync(txBytes)
		if err != nil {
			k.Logger(ctx).Error(err.Error())
		}
		if res != nil && res.Log != "" {
			k.Logger(ctx).Info(res.Log)
		}
	}()
	return nil
}

func (k Keeper) prepareMsgForSigning(ctx sdk.Context, msgs []sdk.Msg) (auth.StdSignMsg, error) {
	if k.config.ChainID == "" {
		return auth.StdSignMsg{}, sdkerrors.Wrap(types.ErrInvalidChain, "chain ID required but not specified")
	}

	acc := k.keeper.GetAccount(ctx, k.from)
	seqNo := k.getSeqNo(ctx)
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

func (k Keeper) RegisterProxy(ctx sdk.Context, principal sdk.ValAddress, proxy sdk.AccAddress) error {
	_, found := k.stakingKeeper.GetValidator(ctx, principal)
	if !found {
		k.Logger(ctx).Error("could not find validator")
		return types.ErrInvalidValidator
	}
	k.Logger(ctx).Debug("getting proxy count")
	count := k.GetProxyCount(ctx)
	k.Logger(ctx).Debug(fmt.Sprintf("count: %v", count))
	storedProxy := ctx.KVStore(k.storeKey).Get(principal)
	if storedProxy != nil {
		ctx.KVStore(k.storeKey).Delete(storedProxy)
		count -= 1
	}
	k.Logger(ctx).Debug("setting proxy")
	ctx.KVStore(k.storeKey).Set(proxy, principal)
	count += 1
	k.Logger(ctx).Debug("setting proxy count")
	k.SetProxyCount(ctx, count)
	k.Logger(ctx).Debug("done")
	return nil
}

func (k Keeper) GetLocalPrincipal(ctx sdk.Context) sdk.ValAddress {
	return k.GetPrincipal(ctx, k.from)
}

func (k Keeper) GetPrincipal(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress {
	if proxy == nil {
		return nil
	}
	return ctx.KVStore(k.storeKey).Get(proxy)
}

func (k Keeper) GetProxyCount(ctx sdk.Context) uint32 {
	countRaw := ctx.KVStore(k.storeKey).Get([]byte(proxyCountKey))
	if countRaw == nil {
		k.Logger(ctx).Error("count was not set, this is an issue with the genesis init")
		return 0
	}
	return binary.LittleEndian.Uint32(countRaw)
}

func (k Keeper) SetProxyCount(ctx sdk.Context, count uint32) {
	bz := make([]byte, 4)
	binary.LittleEndian.PutUint32(bz, count)
	k.Logger(ctx).Debug(fmt.Sprintf("count to set: %v", count))
	k.Logger(ctx).Debug(fmt.Sprintf("count bytes: %v", bz))
	ctx.KVStore(k.storeKey).Set([]byte(proxyCountKey), bz)
}

func (k Keeper) getSeqNo(ctx sdk.Context) uint64 {
	seqNo := ctx.KVStore(k.storeKey).Get([]byte(seqNoKey))
	if seqNo == nil {
		return 0
	}
	return binary.LittleEndian.Uint64(seqNo)
}

func (k Keeper) setSeqNo(ctx sdk.Context, seqNo uint64) {
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, seqNo)
	ctx.KVStore(k.storeKey).Set([]byte(seqNoKey), bz)
}
