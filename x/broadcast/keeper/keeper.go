package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/rpc/client/http"

	"github.com/axelarnetwork/axelar-core/x/broadcast/exported"
	"github.com/axelarnetwork/axelar-core/x/broadcast/types"
)

var _ exported.Broadcaster = broadcaster{}

type broadcaster struct {
	from     sdk.AccAddress
	keybase  keys.Keybase
	keeper   auth.AccountKeeper
	encodeTx sdk.TxEncoder
	config   types.ClientConfig
	rpc      *http.HTTP
	fromName string
}

func NewKeeper(conf types.ClientConfig, keybase keys.Keybase, keeper auth.AccountKeeper, encoder sdk.TxEncoder) (exported.Broadcaster, error) {
	from, fromName, err := getAccountAddress(conf.From, keybase)
	if err != nil {
		return nil, err
	}
	rpc, err := http.New(conf.TendermintNodeUri, "/websocket")
	if err != nil {
		return nil, err
	}

	return broadcaster{
		from:     from,
		fromName: fromName,
		rpc:      rpc,
		config:   conf,
		keybase:  keybase,
		keeper:   keeper,
		encodeTx: encoder,
	}, nil
}

// Logger returns a module-specific logger.
func (b broadcaster) Logger(ctx sdk.Context) log.Logger {
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

func (b broadcaster) Broadcast(ctx sdk.Context, valMsgs []exported.ValidatorMsg) error {
	b.Logger(ctx).Debug("setting sender")
	msgs := make([]sdk.Msg, 0, len(valMsgs))
	for _, msg := range valMsgs {
		b.Logger(ctx).Debug(fmt.Sprintf("b.from: %v", b.from))
		msg.SetSender(b.from)
		msgs = append(msgs, msg)
	}
	b.Logger(ctx).Debug(fmt.Sprintf("preparing to sign:%v", msgs))
	stdSignMsg, err := b.prepareMsgForSigning(ctx, msgs)
	if err != nil {
		return err
	}

	b.Logger(ctx).Debug("signing")
	tx, err := b.sign(stdSignMsg)
	if err != nil {
		return err
	}

	b.Logger(ctx).Debug("encoding tx")
	txBytes, err := b.encodeTx(tx)
	if err != nil {
		b.Logger(ctx).Info(err.Error())
		return err
	}
	b.Logger(ctx).Debug("broadcasting")
	go func() {
		_, err := b.rpc.BroadcastTxSync(txBytes)
		if err != nil {
			b.Logger(ctx).Error(err.Error())
		}
	}()
	return nil
}

func (b broadcaster) prepareMsgForSigning(ctx sdk.Context, msgs []sdk.Msg) (auth.StdSignMsg, error) {
	if b.config.ChainID == "" {
		return auth.StdSignMsg{}, sdkerrors.Wrap(types.ErrInvalidChain, "chain ID required but not specified")
	}

	acc := b.keeper.GetAccount(ctx, b.from)

	return auth.StdSignMsg{
		ChainID:       b.config.ChainID,
		AccountNumber: acc.GetAccountNumber(),
		Sequence:      acc.GetSequence(),
		Msgs:          msgs,
		Fee:           auth.NewStdFee(50000, nil),
	}, nil
}

func (b broadcaster) sign(msg auth.StdSignMsg) (auth.StdTx, error) {
	sig, err := b.makeSignature(msg)
	if err != nil {
		return auth.StdTx{}, err
	}

	return auth.NewStdTx(msg.Msgs, msg.Fee, []auth.StdSignature{sig}, msg.Memo), nil
}

func (b broadcaster) makeSignature(msg auth.StdSignMsg) (auth.StdSignature, error) {
	sigBytes, pubkey, err := b.keybase.Sign(b.fromName, b.config.KeyringPassphrase, msg.Bytes())
	if err != nil {
		return auth.StdSignature{}, err
	}

	return auth.StdSignature{
		PubKey:    pubkey,
		Signature: sigBytes,
	}, nil
}
