package broadcast

import (
	"fmt"
	"math"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	rpc "github.com/tendermint/tendermint/rpc/client"
	"github.com/tendermint/tendermint/rpc/client/http"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/axelarnetwork/axelar-core/cmd/vald/broadcast/types"
	broadcastTypes "github.com/axelarnetwork/axelar-core/x/broadcast/types"
)

// Broadcaster submits transactions to a tendermint node
type Broadcaster struct {
	rpc        types.Client
	logger     log.Logger
	seqNo      uint64
	broadcasts chan func()
	signer     types.Sign
	chainID    string
	gas        uint64
}

// NewBroadcaster returns a broadcaster to submit transactions to the blockchain.
// Only one instance of a broadcaster should be run for a given account, otherwise risk conflicting sequence numbers for submitted transactions.
func NewBroadcaster(signer types.Sign, client types.Client, conf broadcastTypes.ClientConfig, logger log.Logger) (*Broadcaster, error) {
	if conf.ChainID == "" {
		return nil, sdkerrors.Wrap(broadcastTypes.ErrInvalidChain, "chain ID required but not specified")
	}

	broadcaster := &Broadcaster{
		signer:     signer,
		chainID:    conf.ChainID,
		gas:        conf.Gas,
		rpc:        client,
		logger:     logger,
		seqNo:      0,
		broadcasts: make(chan func(), 1000),
	}

	// call broadcast functions sequentially
	go func() {
		// this is expected to run for the full life time of the process, so there is no need to be able to escape the loop
		for broadcast := range broadcaster.broadcasts {
			broadcast()
		}
	}()

	return broadcaster, nil
}

// Broadcast sends the passed messages to the network. This function in thread-safe.
func (b *Broadcaster) Broadcast(msgs ...sdk.Msg) error {
	errChan := make(chan error, 1)
	// push the "intent to run broadcast" into a channel so it can be executed sequentially,
	// even if the public Broadcast function is called concurrently
	b.broadcasts <- func() { errChan <- b.broadcast(msgs) }
	// block until the broadcast call has actually been run
	return <-errChan
}

func (b *Broadcaster) broadcast(msgs []sdk.Msg) error {
	if len(msgs) == 0 {
		return fmt.Errorf("call broadcast with at least one message")
	}

	// By convention the first signer of a tx pays the fees
	if len(msgs[0].GetSigners()) == 0 {
		return fmt.Errorf("messages must have at least one signer")
	}

	accNo, seqNo, err := b.updateAccountNumberSequence(msgs[0].GetSigners()[0])
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			// broadcast has been successful, so increment sequence number
			b.seqNo += 1
		}
	}()

	stdSignMsg := auth.StdSignMsg{
		ChainID:       b.chainID,
		AccountNumber: accNo,
		Sequence:      seqNo,
		Msgs:          msgs,
		Fee:           auth.NewStdFee(b.gas, nil),
	}

	tx, err := sign(b.signer, stdSignMsg)
	if err != nil {
		return err
	}

	b.logger.Debug(fmt.Sprintf("broadcasting %d messages from address: %.20s, acc no.: %d, seq no.: %d, chainId: %s",
		len(msgs), msgs[0].GetSigners()[0], stdSignMsg.AccountNumber, stdSignMsg.Sequence, stdSignMsg.ChainID))

	res, err := b.rpc.BroadcastTxSync(tx)
	if err != nil {
		return err
	}
	if res.Code != abci.CodeTypeOK {
		return fmt.Errorf(res.Log)
	}

	return nil
}

func (b *Broadcaster) updateAccountNumberSequence(addr sdk.AccAddress) (uint64, uint64, error) {
	accNo, seqNo, err := b.rpc.GetAccountNumberSequence(addr)
	if err != nil {
		return 0, 0, err
	}
	if seqNo > b.seqNo {
		b.seqNo = seqNo
	}
	return accNo, b.seqNo, nil
}

func sign(sign types.Sign, msg auth.StdSignMsg) (auth.StdTx, error) {
	var sigs []auth.StdSignature
	for i, m := range msg.Msgs {
		if len(m.GetSigners()) == 0 {
			return auth.StdTx{}, fmt.Errorf("signing failed: msg at idx [%d] without signers", i)
		}
		for _, s := range m.GetSigners() {
			sig, err := sign(s, msg)
			if err != nil {
				return auth.StdTx{}, err
			}
			sigs = append(sigs, sig)
		}
	}

	return auth.NewStdTx(msg.Msgs, msg.Fee, sigs, msg.Memo), nil
}

type client struct {
	rpc.ABCIClient
	encodeTx sdk.TxEncoder
}

// NewClient returns a new rpc client to a tendermint node
func NewClient(encoder sdk.TxEncoder, tendermintURI string) (types.Client, error) {
	abciClient, err := http.New(tendermintURI, "/websocket")
	if err != nil {
		return nil, err
	}

	return client{ABCIClient: abciClient, encodeTx: encoder}, nil
}

// BroadcastTxSync submits a transaction synchronously
func (c client) BroadcastTxSync(tx auth.StdTx) (*coretypes.ResultBroadcastTx, error) {
	txBytes, err := c.encodeTx(tx)
	if err != nil {
		return nil, err
	}
	return c.ABCIClient.BroadcastTxSync(txBytes)
}

// GetAccountNumberSequence returns the account and sequence number of the given address
func (c client) GetAccountNumberSequence(addr sdk.AccAddress) (uint64, uint64, error) {
	return auth.NewAccountRetriever(c).GetAccountNumberSequence(addr)
}

// QueryWithData submits a generic abci query
func (c client) QueryWithData(path string, data []byte) ([]byte, int64, error) {
	res, err := c.ABCIClient.ABCIQuery(path, data)
	if err != nil {
		return nil, 0, err
	}
	if !res.Response.IsOK() {
		return nil, 0, fmt.Errorf(res.Response.Log)
	}

	return res.Response.Value, res.Response.Height, nil
}

// NewSigner unlocks the given keybase, so messages can be signed by the returned sign function
func NewSigner(keybase keys.Keybase, passphrase string) types.Sign {
	return func(from sdk.AccAddress, msg auth.StdSignMsg) (auth.StdSignature, error) {
		info, err := keybase.GetByAddress(from)
		if err != nil {
			return auth.StdSignature{}, err
		}
		sig, pk, err := keybase.Sign(info.GetName(), passphrase, msg.Bytes())
		if err != nil {
			return auth.StdSignature{}, err
		}
		return auth.StdSignature{
			PubKey:    pk,
			Signature: sig,
		}, nil
	}
}

// XBOBroadcaster is a broadcast wrapper that adds retries with exponential backoff
type XBOBroadcaster struct {
	broadcaster *Broadcaster
	timeout     time.Duration
	maxRetries  int
}

// WithExponentialBackoff wraps a broadcaster so that failed broadcasts are retried with exponential backoff
func WithExponentialBackoff(b *Broadcaster, minTimeout time.Duration, maxRetries int) *XBOBroadcaster {
	return &XBOBroadcaster{
		broadcaster: b,
		timeout:     minTimeout,
		maxRetries:  maxRetries,
	}
}

// Broadcast submits messages synchronously and retries with exponential backoff.
// This function is thread-safe but might block for a long time depending on the exponential backoff parameters.
func (b *XBOBroadcaster) Broadcast(msgs ...sdk.Msg) error {
	errChan := make(chan error)
	go func() {
		defer close(errChan)
		errChan <- b.broadcastWithBackoff(msgs)
	}()
	return <-errChan
}

func (b *XBOBroadcaster) broadcastWithBackoff(msgs []sdk.Msg) error {
	for i := 0; i <= b.maxRetries; i++ {
		err := b.broadcaster.Broadcast(msgs...)
		if err == nil {
			return nil
		}

		if i >= b.maxRetries {
			return sdkerrors.Wrap(err, fmt.Sprintf("aborting broadcast after %d retries", b.maxRetries))
		}

		// exponential backoff
		timeout := time.Duration(math.Pow(2, float64(i))) * b.timeout
		b.broadcaster.logger.Error(sdkerrors.Wrapf(err, "exponentially backing off (retry in %v )", timeout).Error())
		time.Sleep(timeout)
	}

	panic("this should be unreachable")
}
