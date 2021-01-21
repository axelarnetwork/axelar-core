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
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types/mock"
)

// TestQuerier_TxInfo_CorrectMarshalling is a regression test that ensures OutPointInfo is correctly marshalled and unmarshalled
func TestQuerier_TxInfo_CorrectMarshalling(t *testing.T) {
	for i := 0; i < 100; i++ {
		var bz []byte
		for _, b := range testutils.RandIntsBetween(0, 256).Take(chainhash.HashSize) {
			bz = append(bz, byte(b))
		}
		hash, err := chainhash.NewHash(bz)
		assert.NoError(t, err)
		info := types.OutPointInfo{
			OutPoint: &wire.OutPoint{
				Hash:  *hash,
				Index: uint32(testutils.RandIntBetween(0, 100)),
			},
			Amount:        btcutil.Amount(testutils.RandIntBetween(0, 100000000)),
			DepositAddr:   testutils.RandStrings(5, 20).Take(1)[0],
			Confirmations: uint64(testutils.RandIntBetween(0, 10000)),
		}

		query := NewQuerier(Keeper{}, &mock.SignerMock{}, &mock.BalancerMock{}, &mock.RPCClientMock{
			GetOutPointInfoFunc: func(out *wire.OutPoint) (types.OutPointInfo, error) {
				if out.Hash.IsEqual(&info.OutPoint.Hash) {
					return info, nil
				}
				return types.OutPointInfo{}, fmt.Errorf("not found")
			}})
		bz, err = query(sdk.Context{}, []string{QueryOutInfo}, abci.RequestQuery{Data: testutils.Codec().MustMarshalJSON(info.OutPoint)})
		assert.NoError(t, err)

		var unmarshaled types.OutPointInfo
		testutils.Codec().MustUnmarshalJSON(bz, &unmarshaled)
		assert.True(t, info.Equals(unmarshaled))
	}
}
