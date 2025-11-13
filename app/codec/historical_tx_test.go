package codec_test

import (
	"encoding/base64"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/axelarnetwork/axelar-core/app"
)

// TestHistoricalBlock3198695 tests decoding a real transaction from block 3198695
// This block is from before v0.20 and contains a VoteRequest with the old poll_key
// and vote fields (fields 2 and 3) that were later reserved.
//
// This test verifies that by marking those fields as deprecated instead of reserved,
// we can still decode historical transactions.
func TestHistoricalBlock3198695(t *testing.T) {
	// This is the base64-encoded transaction bytes from block 3198695 (first tx)
	// This transaction has a VoteRequest with fields 2 and 3 populated
	txBase64 := "Cv0CCvoCCicvYXhlbGFyLnJld2FyZC52MWJldGExLlJlZnVuZE1zZ1JlcXVlc3QSzgIKFPzKhSyubX7OCGgBvgC57Brhdk0AErUCCiAvYXhlbGFyLnZvdGUudjFiZXRhMS5Wb3RlUmVxdWVzdBKQAgoU/MqFLK5tfs4IaAG+ALnsGuF2TQASdAoDZXZtEm0weGI0NzFmMWQyODRlMDRkYTE4YmZjYzg4ZDc4YWFjOGE4YzJlOTIzNDYyZDFhMmRhMTM2YzBhODZhYzI1MmRhOTZfMHhDNUJCNUEyM2I4ZDJBNTFEODdhN2YxNTYwOGNhNkM5OUY4RDBDMDIwGoEBEn8KHi9heGVsYXIuZXZtLnYxYmV0YTEuVm90ZUV2ZW50cxJdCghFdGhlcmV1bRJRCghFdGhlcmV1bRIgtHHx0oTgTaGL/MiNeKrIqMLpI0YtGi2hNsCoasJS2pZCIwoUxbtaI7jSpR2Hp/FWCMpsmfjQwCASCzIwMDQwMDAwMDAwEmYKUgpGCh8vY29zbW9zLmNyeXB0by5zZWNwMjU2azEuUHViS2V5EiMKIQPlFLbxU1TWCK56TRz4414WDMzIqQfzXEXxsVRRHizsVRIECgIIARjEixQSEAoKCgR1YXhsEgIyMxDgmhsaQPOviYb1rvKt0w1AXhomq2mPyyP+gJt/j/+Du/CXY0lMcLBMongNvoUs2WeXZRNcAA1+PZYGzKN80jCfs2OHxnQ="

	// Decode base64
	txBytes, err := base64.StdEncoding.DecodeString(txBase64)
	require.NoError(t, err, "failed to decode base64")

	// Create encoding config with our updated proto definitions
	encodingConfig := app.MakeEncodingConfig()

	// Decode the transaction - this should succeed with deprecated fields
	tx, err := encodingConfig.TxConfig.TxDecoder()(txBytes)
	require.NoError(t, err, "failed to decode transaction - deprecated fields should allow decoding!")

	// Verify we can get the messages
	msgs := tx.GetMsgs()
	require.Len(t, msgs, 1, "should have 1 message")

	// The message should be a RefundMsgRequest
	require.Equal(t, "/axelar.reward.v1beta1.RefundMsgRequest", sdk.MsgTypeURL(msgs[0]))

	t.Logf("Successfully decoded transaction from block 3198695 with deprecated VoteRequest fields!")
}
