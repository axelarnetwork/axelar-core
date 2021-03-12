package keeper

import (
	"fmt"
	"testing"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types/mock"
)

// TestQuerier_TxInfo_CorrectMarshalling is a regression test that ensures OutPointInfo is correctly marshalled and unmarshalled
func TestQuerier_TxInfo_CorrectMarshalling(t *testing.T) {
	for i := 0; i < 100; i++ {
		var bz []byte
		for _, b := range rand.I64GenBetween(0, 256).Take(chainhash.HashSize) {
			bz = append(bz, byte(b))
		}
		txHash, err := chainhash.NewHash(bz)
		assert.NoError(t, err)
		blockHash, err := chainhash.NewHash(rand.Bytes(chainhash.HashSize))
		assert.NoError(t, err)

		info := types.OutPointInfo{
			OutPoint: &wire.OutPoint{
				Hash:  *txHash,
				Index: uint32(rand.I64Between(0, 100)),
			},
			BlockHash:     blockHash,
			Amount:        btcutil.Amount(rand.I64Between(0, 100000000)),
			Address:       rand.Strings(5, 20).Take(1)[0],
			Confirmations: uint64(rand.I64Between(0, 10000)),
		}

		query := NewQuerier(Keeper{}, &mock.SignerMock{}, &mock.NexusMock{}, &mock.RPCClientMock{
			GetOutPointInfoFunc: func(_ *chainhash.Hash, out *wire.OutPoint) (types.OutPointInfo, error) {
				if out.Hash.IsEqual(&info.OutPoint.Hash) {
					return info, nil
				}
				return types.OutPointInfo{}, fmt.Errorf("not found")
			}})
		bz, err = query(sdk.Context{}, []string{QueryOutInfo, info.BlockHash.String()}, abci.RequestQuery{Data: testutils.Codec().MustMarshalJSON(info.OutPoint)})
		assert.NoError(t, err)

		var unmarshaled types.OutPointInfo
		testutils.Codec().MustUnmarshalJSON(bz, &unmarshaled)
		assert.True(t, info.Equals(unmarshaled))
	}
}
