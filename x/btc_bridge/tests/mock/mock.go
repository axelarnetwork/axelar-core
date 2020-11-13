package mock

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	sdkTypes "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/axelar/exported"
	"github.com/axelarnetwork/axelar-core/x/btc_bridge/types"
	tssTypes "github.com/axelarnetwork/axelar-core/x/tss/types"
)

var _ types.Voter = &TestVoter{}

type TestVoter struct {
	Vote *exported.FutureVote
}

func (t *TestVoter) SetFutureVote(ctx sdkTypes.Context, vote exported.FutureVote) {
	t.Vote = &vote
}

func (t TestVoter) IsVerified(ctx sdkTypes.Context, tx exported.ExternalTx) bool {
	panic("implement me")
}

var _ types.RPCClient = &TestRPC{}

type TestRPC struct {
	TrackedAddress string ``
	Cancel         context.CancelFunc
	RawTxs         map[string]*btcjson.TxRawResult
}

func (t *TestRPC) ImportAddressRescan(address string, account string, rescan bool) error {
	t.TrackedAddress = address
	t.Cancel()
	return nil
}

func (t *TestRPC) ImportAddress(address string) error {
	t.TrackedAddress = address
	t.Cancel()
	return nil
}

func (t TestRPC) GetRawTransactionVerbose(hash *chainhash.Hash) (*btcjson.TxRawResult, error) {
	if res, ok := t.RawTxs[hash.String()]; !ok {
		return nil, fmt.Errorf("tx not found")
	} else {
		return res, nil
	}
}

func (t TestRPC) SendRawTransaction(tx *wire.MsgTx, b bool) (*chainhash.Hash, error) {
	panic("implement me")
}

var _ types.Signer = TestSigner{}

type TestSigner struct {
}

func (t TestSigner) StartSign(ctx sdkTypes.Context, info tssTypes.MsgSignStart) error {
	panic("implement me")
}

func (t TestSigner) GetSig(ctx sdkTypes.Context, sigID string) (r *big.Int, s *big.Int, err error) {
	panic("implement me")
}

func (t TestSigner) GetKey(ctx sdkTypes.Context, keyID string) (ecdsa.PublicKey, error) {
	panic("implement me")
}
