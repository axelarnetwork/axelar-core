package types

import (
	"github.com/axelarnetwork/axelar-core/utils"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// HexSignature represents a tss signature as hex encoded bytes for use in responses from the query client.
type HexSignature struct {
	R string `json:"r"`
	S string `json:"s"`
}

// NewHexSignatureFromQuerySigResponse converts a QuerySigResponse to a HexSignature
func NewHexSignatureFromQuerySigResponse(sigResp *QuerySigResponse) HexSignature {
	return HexSignature{
		R: hexutil.Encode(sigResp.Signature.R),
		S: hexutil.Encode(sigResp.Signature.S),
	}
}

// ComputeCorruptionThreshold returns corruption threshold to be used by tss.
// (threshold + 1) shares are required to sign
func ComputeCorruptionThreshold(threshold utils.Threshold, totalShareCount sdk.Int) int64 {
	return totalShareCount.MulRaw(threshold.Numerator).QuoRaw(threshold.Denominator).Int64() - 1
}
