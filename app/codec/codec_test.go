package codec_test

import (
	"encoding/base64"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/axelarnetwork/axelar-core/app"
)

// TestRealTransactionFromBlock451 tests decoding a real transaction from block 451
// that contains a HeartBeatRequest with the old type URL /tss.v1beta1.HeartBeatRequest
// This was failing with:
// "no concrete type registered for type URL /tss.v1beta1.HeartBeatRequest against interface *exported.Refundable"
func TestRealTransactionFromBlock451(t *testing.T) {
	// This is the base64-encoded transaction bytes from block 451 (first tx)
	// Query: http://archive-node-lcd.axelarscan.io/blocks/451
	txBase64 := "CnUKcwogL3Jld2FyZC52MWJldGExLlJlZnVuZE1zZ1JlcXVlc3QSTwoUKnkaFofQifOcAd8S4mOdnUSccPsSNwodL3Rzcy52MWJldGExLkhlYXJ0QmVhdFJlcXVlc3QSFgoUKnkaFofQifOcAd8S4mOdnUSccPsSZwpQCkYKHy9jb3Ntb3MuY3J5cHRvLnNlY3AyNTZrMS5QdWJLZXkSIwohAq/NQXF1uAwAk81syNM4LYWL16kn/UsV7MOADy6lVVv+EgQKAggBGAkSEwoNCgR1YXhsEgUxNjQ3ORDkjhQaQNVlqdpVNALeVHpQlM1juQvtv/XDJCEZ89ZzLWI5C1dONm+NFCqBQHExP6dldeOdh8/UZrVACqSOy9moT5+7LzU="

	// Decode base64
	txBytes, err := base64.StdEncoding.DecodeString(txBase64)
	require.NoError(t, err, "failed to decode base64")

	// Create encoding config with our legacy registrations
	encodingConfig := app.MakeEncodingConfig()

	// Decode the transaction - this is where the bug was happening
	tx, err := encodingConfig.TxConfig.TxDecoder()(txBytes)
	require.NoError(t, err, "failed to decode transaction - this is the bug we're fixing!")

	// Verify we can get the messages
	msgs := tx.GetMsgs()
	require.Len(t, msgs, 1, "should have 1 message")

	// The message should be a RefundMsgRequest
	require.Equal(t, "/reward.v1beta1.RefundMsgRequest", sdk.MsgTypeURL(msgs[0]))

	t.Logf("Successfully decoded transaction from block 451 containing /tss.v1beta1.HeartBeatRequest")
}

// TestRealTransactionFromBlock1000000 tests decoding a real transaction from block 1000000
// that contains a SubmitMultisigSignaturesRequest with the old type URL /tss.v1beta1.SubmitMultisigSignaturesRequest
// This was failing with:
// "no concrete type registered for type URL /tss.v1beta1.SubmitMultisigSignaturesRequest against interface *exported.Refundable"
func TestRealTransactionFromBlock1000000(t *testing.T) {
	// This is the base64-encoded transaction bytes from block 1000000 (first tx)
	// Query: http://archive-node-lcd.axelarscan.io/blocks/1000000
	txBase64 := "CpICCo8CCiAvcmV3YXJkLnYxYmV0YTEuUmVmdW5kTXNnUmVxdWVzdBLqAQoU/MqFLK5tfs4IaAG+ALnsGuF2TQAS0QEKLC90c3MudjFiZXRhMS5TdWJtaXRNdWx0aXNpZ1NpZ25hdHVyZXNSZXF1ZXN0EqABChT8yoUsrm1+zghoAb4Auewa4XZNABJAOTI3N2IzNDNiNGNlN2M2YWE2YTYzYjdmMDA0N2ExNjkyODRkYTJmNTAxMmQ4NzZjYWQ5MjEyY2ZlYzYyNzQwMxpGMEQCIHA6WlZXCdvyyoD5NCGTdjNRAW7ZIMehopJCcsI795QwAiBinkB8CAZlSZw3rR0WzF0K0Q7fPaPeV7zRVzso5f9f2BJpClIKRgofL2Nvc21vcy5jcnlwdG8uc2VjcDI1NmsxLlB1YktleRIjCiED5RS28VNU1giuek0c+ONeFgzMyKkH81xF8bFUUR4s7FUSBAoCCAEY8f4JEhMKDQoEdWF4bBIFMTc2MzEQ6MIVGkAhmdiXotz46aKgN5bqNA28HNPioodiJbjk5N1Dg+0AVTItt/pGZJhZxn1lcxLtqk7cpm1l/ugRemftuI3Uok8r"

	// Decode base64
	txBytes, err := base64.StdEncoding.DecodeString(txBase64)
	require.NoError(t, err, "failed to decode base64")

	// Create encoding config with our legacy registrations
	encodingConfig := app.MakeEncodingConfig()

	// Decode the transaction - this is where the bug was happening
	tx, err := encodingConfig.TxConfig.TxDecoder()(txBytes)
	require.NoError(t, err, "failed to decode transaction - this is the bug we're fixing!")

	// Verify we can get the messages
	msgs := tx.GetMsgs()
	require.Len(t, msgs, 1, "should have 1 message")

	// The message should be a RefundMsgRequest
	require.Equal(t, "/reward.v1beta1.RefundMsgRequest", sdk.MsgTypeURL(msgs[0]))

	t.Logf("Successfully decoded transaction from block 1000000 containing /tss.v1beta1.SubmitMultisigSignaturesRequest")
}
