package broadcast

import (
	"encoding/binary"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/rpc/client"

	"github.com/axelarnetwork/axelar-core/x/broadcast/exported"
	"github.com/axelarnetwork/axelar-core/x/broadcast/types"
)

var (
	proxyKey     = []byte("proxy")
	proxyNameKey = []byte("proxyName")
	seqNoKey     = []byte("seqNo")
	gasKey       = []byte("gas")
)

type Broadcaster struct {
	keybase          keys.Keybase
	encodeTx         sdk.TxEncoder
	cdc              *codec.Codec
	config           types.ClientConfig
	rpc              client.ABCIClient
	logger           log.Logger
	store            sdk.KVStore
	accountRetriever auth.AccountRetriever
}

func NewBroadcaster(cdc *codec.Codec, keybase keys.Keybase, store sdk.KVStore, client client.ABCIClient, conf types.ClientConfig, logger log.Logger) (Broadcaster, error) {
	from, fromName, err := types.GetAccountAddress(conf.From, keybase)
	if err != nil {
		return Broadcaster{}, err
	}
	store.Set(proxyKey, from)
	store.Set(proxyNameKey, []byte(fromName))

	b := Broadcaster{
		keybase:          keybase,
		encodeTx:         utils.GetTxEncoder(cdc),
		cdc:              cdc,
		config:           conf,
		rpc:              client,
		logger:           logger,
		store:            store,
		accountRetriever: auth.NewAccountRetriever(querier{client}),
	}
	b.setGas(uint64(conf.Gas))

	return b, nil
}

// Broadcast sends the passed message to the network. Needs to be called asynchronously or it will block
func (b Broadcaster) Broadcast(msgsWithoutSender []exported.MsgWithSenderSetter) error {
	var msgs []sdk.Msg
	for _, msg := range msgsWithoutSender {
		msg.SetSender(b.store.Get(proxyKey))
		msgs = append(msgs, msg)
	}

	stdSignMsg, err := b.prepareMsgForSigning(msgs)
	if err != nil {
		return err
	}

	tx, err := b.sign(stdSignMsg)
	if err != nil {
		return err
	}

	b.logger.Debug(fmt.Sprintf("broadcasting %d messages from address: %.20s, acc no.: %d, seq no.: %d, chainId: %s",
		len(msgs), sdk.AccAddress(b.store.Get(proxyKey)).String(), stdSignMsg.AccountNumber, stdSignMsg.Sequence, stdSignMsg.ChainID))

	txBytes, err := b.encodeTx(tx)
	if err != nil {
		return err
	}

	b.setSeqNo(stdSignMsg.Sequence + 1)
	res, err := b.rpc.BroadcastTxSync(txBytes)
	if err != nil {
		return err
	}
	if res.Code != abci.CodeTypeOK {
		return fmt.Errorf(res.Log)
	}

	return nil
}

func (b Broadcaster) prepareMsgForSigning(msgs []sdk.Msg) (auth.StdSignMsg, error) {
	if b.config.ChainID == "" {
		return auth.StdSignMsg{}, sdkerrors.Wrap(types.ErrInvalidChain, "chain ID required but not specified")
	}

	accNo, seqNo, err := b.accountRetriever.GetAccountNumberSequence(b.store.Get(proxyKey))
	if err != nil {
		return auth.StdSignMsg{}, err
	}
	localSeqNo := b.getSeqNo()
	if seqNo > localSeqNo {
		localSeqNo = seqNo
		b.setSeqNo(localSeqNo)
	}

	return auth.StdSignMsg{
		ChainID:       b.config.ChainID,
		AccountNumber: accNo,
		Sequence:      localSeqNo,
		Msgs:          msgs,
		Fee:           auth.NewStdFee(b.getGas(), nil),
	}, nil
}

func (b Broadcaster) sign(msg auth.StdSignMsg) (auth.StdTx, error) {
	name := b.store.Get(proxyNameKey)
	if name == nil {
		return auth.StdTx{}, fmt.Errorf("name of the sender account unknown")
	}
	sigBytes, pubkey, err := b.keybase.Sign(string(name), b.config.KeyringPassphrase, msg.Bytes())
	if err != nil {
		return auth.StdTx{}, err
	}

	sig := auth.StdSignature{PubKey: pubkey, Signature: sigBytes}

	return auth.NewStdTx(msg.Msgs, msg.Fee, []auth.StdSignature{sig}, msg.Memo), nil
}

func (b Broadcaster) getSeqNo() uint64 {
	seqNo := b.store.Get(seqNoKey)
	if seqNo == nil {
		return 0
	}
	return binary.LittleEndian.Uint64(seqNo)
}

func (b Broadcaster) setSeqNo(seqNo uint64) {
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, seqNo)
	b.store.Set(seqNoKey, bz)
}

func (b Broadcaster) getGas() uint64 {
	gas := b.store.Get(gasKey)
	if gas == nil {
		return 0
	}
	return binary.LittleEndian.Uint64(gas)
}

func (b Broadcaster) setGas(gas uint64) {
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, gas)
	b.store.Set(gasKey, bz)
}

type querier struct {
	client.ABCIClient
}

func (q querier) QueryWithData(path string, data []byte) ([]byte, int64, error) {
	res, err := q.ABCIQuery(path, data)
	if err != nil {
		return nil, 0, err
	}
	if !res.Response.IsOK() {
		return nil, 0, fmt.Errorf(res.Response.Log)
	}

	return res.Response.Value, res.Response.Height, nil
}
