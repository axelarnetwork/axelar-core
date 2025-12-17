package codec_test

import (
	"encoding/base64"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/axelarnetwork/axelar-core/app"
	evmtypes "github.com/axelarnetwork/axelar-core/x/evm/types"
)

// HistoricalTransaction represents a test case for decoding historical transactions
type HistoricalTransaction struct {
	Name        string // Test name/description
	Block       int64  // Block number
	TxBase64    string // Base64-encoded transaction bytes
	ExpectedMsg string // Expected message type URL
}

// TestHistoricalTransactions tests decoding real transactions from historical blocks
// that contain messages with deprecated fields. This is a regression test suite
// to ensure backward compatibility is maintained.
func TestHistoricalTransactions(t *testing.T) {
	testCases := []HistoricalTransaction{
		{
			Name:        "Block451_HeartBeatRequest_LegacyTypeURL",
			Block:       451,
			ExpectedMsg: "/axelar.reward.v1beta1.RefundMsgRequest",
			TxBase64:    "CnUKcwogL3Jld2FyZC52MWJldGExLlJlZnVuZE1zZ1JlcXVlc3QSTwoUKnkaFofQifOcAd8S4mOdnUSccPsSNwodL3Rzcy52MWJldGExLkhlYXJ0QmVhdFJlcXVlc3QSFgoUKnkaFofQifOcAd8S4mOdnUSccPsSZwpQCkYKHy9jb3Ntb3MuY3J5cHRvLnNlY3AyNTZrMS5QdWJLZXkSIwohAq/NQXF1uAwAk81syNM4LYWL16kn/UsV7MOADy6lVVv+EgQKAggBGAkSEwoNCgR1YXhsEgUxNjQ3ORDkjhQaQNVlqdpVNALeVHpQlM1juQvtv/XDJCEZ89ZzLWI5C1dONm+NFCqBQHExP6dldeOdh8/UZrVACqSOy9moT5+7LzU=",
		},
		{
			Name:        "Block1000000_SubmitMultisigSignaturesRequest_LegacyTypeURL",
			Block:       1000000,
			ExpectedMsg: "/axelar.reward.v1beta1.RefundMsgRequest",
			TxBase64:    "CpICCo8CCiAvcmV3YXJkLnYxYmV0YTEuUmVmdW5kTXNnUmVxdWVzdBLqAQoU/MqFLK5tfs4IaAG+ALnsGuF2TQAS0QEKLC90c3MudjFiZXRhMS5TdWJtaXRNdWx0aXNpZ1NpZ25hdHVyZXNSZXF1ZXN0EqABChT8yoUsrm1+zghoAb4Auewa4XZNABJAOTI3N2IzNDNiNGNlN2M2YWE2YTYzYjdmMDA0N2ExNjkyODRkYTJmNTAxMmQ4NzZjYWQ5MjEyY2ZlYzYyNzQwMxpGMEQCIHA6WlZXCdvyyoD5NCGTdjNRAW7ZIMehopJCcsI795QwAiBinkB8CAZlSZw3rR0WzF0K0Q7fPaPeV7zRVzso5f9f2BJpClIKRgofL2Nvc21vcy5jcnlwdG8uc2VjcDI1NmsxLlB1YktleRIjCiED5RS28VNU1giuek0c+ONeFgzMyKkH81xF8bFUUR4s7FUSBAoCCAEY8f4JEhMKDQoEdWF4bBIFMTc2MzEQ6MIVGkAhmdiXotz46aKgN5bqNA28HNPioodiJbjk5N1Dg+0AVTItt/pGZJhZxn1lcxLtqk7cpm1l/ugRemftuI3Uok8r",
		},
		{
			Name:        "Block3198695_VoteRequest_DeprecatedPollKeyAndVote",
			Block:       3198695,
			ExpectedMsg: "/axelar.reward.v1beta1.RefundMsgRequest",
			TxBase64:    "Cv0CCvoCCicvYXhlbGFyLnJld2FyZC52MWJldGExLlJlZnVuZE1zZ1JlcXVlc3QSzgIKFPzKhSyubX7OCGgBvgC57Brhdk0AErUCCiAvYXhlbGFyLnZvdGUudjFiZXRhMS5Wb3RlUmVxdWVzdBKQAgoU/MqFLK5tfs4IaAG+ALnsGuF2TQASdAoDZXZtEm0weGI0NzFmMWQyODRlMDRkYTE4YmZjYzg4ZDc4YWFjOGE4YzJlOTIzNDYyZDFhMmRhMTM2YzBhODZhYzI1MmRhOTZfMHhDNUJCNUEyM2I4ZDJBNTFEODdhN2YxNTYwOGNhNkM5OUY4RDBDMDIwGoEBEn8KHi9heGVsYXIuZXZtLnYxYmV0YTEuVm90ZUV2ZW50cxJdCghFdGhlcmV1bRJRCghFdGhlcmV1bRIgtHHx0oTgTaGL/MiNeKrIqMLpI0YtGi2hNsCoasJS2pZCIwoUxbtaI7jSpR2Hp/FWCMpsmfjQwCASCzIwMDQwMDAwMDAwEmYKUgpGCh8vY29zbW9zLmNyeXB0by5zZWNwMjU2azEuUHViS2V5EiMKIQPlFLbxU1TWCK56TRz4414WDMzIqQfzXEXxsVRRHizsVRIECgIIARjEixQSEAoKCgR1YXhsEgIyMxDgmhsaQPOviYb1rvKt0w1AXhomq2mPyyP+gJt/j/+Du/CXY0lMcLBMongNvoUs2WeXZRNcAA1+PZYGzKN80jCfs2OHxnQ=",
		},
		{
			Name:        "Block2777636_ConfirmTransferKeyRequest_TransferTypeAndKeyID",
			Block:       2777636,
			ExpectedMsg: "/axelar.evm.v1beta1.ConfirmTransferKeyRequest",
			TxBase64:    "CpcBCpQBCi0vYXhlbGFyLmV2bS52MWJldGExLkNvbmZpcm1UcmFuc2ZlcktleVJlcXVlc3QSYwoUXMoe7FB79JBAKaJQZcqx5BzxkicSCWF2YWxhbmNoZRogpPGKKnEbUFaOjhZG/Izqcp/9D3reU3/GnvbLuPPGKLsgASocbWFzdGVyLWV2bS1hdmFsYW5jaGUtMjc3NzYyMhKUAQpRCkYKHy9jb3Ntb3MuY3J5cHRvLnNlY3AyNTZrMS5QdWJLZXkSIwohApdvAGNAZvGRk7YisdOloATew6TjUgC4FbxnCUkjr2uREgQKAggBGMEQEj8KCgoEdWF4bBICMzcQk4ktIi1heGVsYXIxcHUyc3djMG4wdHJmdGxkaHo1N3B5cWt3NmQ4N2hhaG43ZzY5N2MaQGjhbdepSVfV87pD4+PJUICEeAKWST9HM3scVS3/6tpuTW+VB6uhnAd+5PM6NTEyxtdQvDgCxg4HM//mywsxS+k=",
		},
		// Add more test cases here as you find blocks with deprecated message types
	}

	// Create encoding config once for all tests
	encodingConfig := app.MakeEncodingConfig()

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Decode base64
			txBytes, err := base64.StdEncoding.DecodeString(tc.TxBase64)
			require.NoError(t, err, "failed to decode base64 for block %d", tc.Block)

			// Decode the transaction
			tx, err := encodingConfig.TxConfig.TxDecoder()(txBytes)
			require.NoError(t, err, "failed to decode transaction from block %d - deprecated fields should allow decoding!", tc.Block)

			// Verify we can get the messages
			msgs := tx.GetMsgs()
			require.NotEmpty(t, msgs, "should have at least one message")

			// Verify the expected message type
			require.Equal(t, tc.ExpectedMsg, sdk.MsgTypeURL(msgs[0]), "unexpected message type in block %d", tc.Block)

			t.Logf("✓ Successfully decoded transaction from block %d", tc.Block)
		})
	}
}

// TestBulkHistoricalTransactions allows testing multiple transactions at once
// This is useful when you have a batch of transaction data to validate
func TestBulkHistoricalTransactions(t *testing.T) {
	// This test can be populated with bulk data from block scanning
	// Leave empty for now - can be populated as more test cases are discovered
	t.Skip("Populate this test with bulk transaction data from block scanning")
}

// TestHistoricalTransactionsGetSigners tests that GetMsgV1Signers works correctly
// for both historical transactions (using sender_deprecated) and new transactions (using sender).
// This is a regression test for rosetta compatibility.
func TestHistoricalTransactionsGetSigners(t *testing.T) {
	testCases := []struct {
		Name           string
		Block          int64
		TxBase64       string
		ExpectedSigner string // hex-encoded expected signer address
	}{
		// Old format: sender in sender_deprecated field (bytes)
		{
			Name:           "Block2777636_ConfirmTransferKeyRequest_SenderDeprecated",
			Block:          2777636,
			TxBase64:       "CpcBCpQBCi0vYXhlbGFyLmV2bS52MWJldGExLkNvbmZpcm1UcmFuc2ZlcktleVJlcXVlc3QSYwoUXMoe7FB79JBAKaJQZcqx5BzxkicSCWF2YWxhbmNoZRogpPGKKnEbUFaOjhZG/Izqcp/9D3reU3/GnvbLuPPGKLsgASocbWFzdGVyLWV2bS1hdmFsYW5jaGUtMjc3NzYyMhKUAQpRCkYKHy9jb3Ntb3MuY3J5cHRvLnNlY3AyNTZrMS5QdWJLZXkSIwohApdvAGNAZvGRk7YisdOloATew6TjUgC4FbxnCUkjr2uREgQKAggBGMEQEj8KCgoEdWF4bBICMzcQk4ktIi1heGVsYXIxcHUyc3djMG4wdHJmdGxkaHo1N3B5cWt3NmQ4N2hhaG43ZzY5N2MaQGjhbdepSVfV87pD4+PJUICEeAKWST9HM3scVS3/6tpuTW+VB6uhnAd+5PM6NTEyxtdQvDgCxg4HM//mywsxS+k=",
			ExpectedSigner: "5cca1eec507bf4904029a25065cab1e41cf19227",
		},
		{
			Name:           "Block451_HeartBeatRequest_SenderDeprecated",
			Block:          451,
			TxBase64:       "CnUKcwogL3Jld2FyZC52MWJldGExLlJlZnVuZE1zZ1JlcXVlc3QSTwoUKnkaFofQifOcAd8S4mOdnUSccPsSNwodL3Rzcy52MWJldGExLkhlYXJ0QmVhdFJlcXVlc3QSFgoUKnkaFofQifOcAd8S4mOdnUSccPsSZwpQCkYKHy9jb3Ntb3MuY3J5cHRvLnNlY3AyNTZrMS5QdWJLZXkSIwohAq/NQXF1uAwAk81syNM4LYWL16kn/UsV7MOADy6lVVv+EgQKAggBGAkSEwoNCgR1YXhsEgUxNjQ3ORDkjhQaQNVlqdpVNALeVHpQlM1juQvtv/XDJCEZ89ZzLWI5C1dONm+NFCqBQHExP6dldeOdh8/UZrVACqSOy9moT5+7LzU=",
			ExpectedSigner: "2a791a1687d089f39c01df12e2639d9d449c70fb",
		},
	}

	encodingConfig := app.MakeEncodingConfig()

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			txBytes, err := base64.StdEncoding.DecodeString(tc.TxBase64)
			require.NoError(t, err)

			tx, err := encodingConfig.TxConfig.TxDecoder()(txBytes)
			require.NoError(t, err)

			msgs := tx.GetMsgs()
			require.NotEmpty(t, msgs)

			// Test that GetMsgV1Signers works with sender_deprecated
			signers, _, err := encodingConfig.Codec.GetMsgV1Signers(msgs[0])
			require.NoError(t, err, "GetMsgV1Signers should work with sender_deprecated field")
			require.Len(t, signers, 1, "should have exactly one signer")
			require.Equal(t, tc.ExpectedSigner, fmt.Sprintf("%x", signers[0]), "signer address mismatch")

			t.Logf("✓ GetMsgV1Signers returned correct signer for block %d", tc.Block)
		})
	}
}

// TestNewSenderFieldGetSigners tests that GetMsgV1Signers works correctly
// for messages using the new sender field (string format).
func TestNewSenderFieldGetSigners(t *testing.T) {
	encodingConfig := app.MakeEncodingConfig()

	// Create a message with the new sender field set
	senderAddr := sdk.AccAddress("test_sender_address1")
	msg := &evmtypes.ConfirmGatewayTxsRequest{
		Sender: senderAddr.String(),
		Chain:  "ethereum",
		TxIDs:  []evmtypes.Hash{evmtypes.Hash(make([]byte, 32))},
	}

	signers, _, err := encodingConfig.Codec.GetMsgV1Signers(msg)
	require.NoError(t, err, "GetMsgV1Signers should work with new sender field")
	require.Len(t, signers, 1, "should have exactly one signer")
	require.Equal(t, senderAddr.Bytes(), signers[0], "signer address should match")

	t.Logf("✓ GetMsgV1Signers works with new sender field format")
}
