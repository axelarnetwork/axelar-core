package utils

import (
	"testing"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogoprototypes "github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

var stringGen = rand.Strings(5, 50).Distinct()

func createBlockHeightKVQueue(name string) BlockHeightKVQueue {
	interfaceRegistry := types.NewInterfaceRegistry()
	interfaceRegistry.RegisterImplementations((*codec.ProtoMarshaler)(nil),
		&gogoprototypes.StringValue{},
	)
	marshaler := codec.NewProtoCodec(interfaceRegistry)
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	store := NewNormalizedStore(ctx.KVStore(sdk.NewKVStoreKey(stringGen.Next())), marshaler)

	return NewBlockHeightKVQueue(store, ctx, name).(BlockHeightKVQueue)
}

func TestNewBlockHeightKVQueue(t *testing.T) {
	repeats := 20

	t.Run("enqueue and dequeue", testutils.Func(func(t *testing.T) {
		kvQueue := createBlockHeightKVQueue("test-enqueue-dequeue")

		blockHeight := rand.I64Between(1, 10000)
		kvQueue.ctx = kvQueue.ctx.WithBlockHeight(blockHeight)
		itemCount := rand.I64Between(10, 1000)
		items := make([]string, itemCount)

		for i := 0; i < int(itemCount); i++ {
			items[i] = rand.Str(10)
		}

		for _, item := range items {
			kvQueue.Enqueue(RegularKey(item), &gogoprototypes.StringValue{Value: item})
			blockHeight += rand.I64Between(1, 1000)
			kvQueue.ctx = kvQueue.ctx.WithBlockHeight(blockHeight)
		}

		actualItems := []string{}
		var actualItem gogoprototypes.StringValue

		for kvQueue.Dequeue(&actualItem) {
			actualItems = append(actualItems, actualItem.Value)
		}

		assert.Equal(t, items, actualItems)
	}).Repeat(repeats))
}
