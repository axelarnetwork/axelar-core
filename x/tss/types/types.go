package types

import (
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
