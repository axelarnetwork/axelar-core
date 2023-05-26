package rpc

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

//go:generate stringer -type=FinalityOverride

// Header represents a block header in any EVM blockchain
type Header struct {
	ParentHash    common.Hash    `json:"parentHash"       gencodec:"required"`
	Number        *hexutil.Big   `json:"number"           gencodec:"required"`
	Time          hexutil.Uint64 `json:"timestamp"        gencodec:"required"`
	Hash          common.Hash    `json:"hash"`
	Transactions  []common.Hash  `json:"transactions"     gencodec:"required"`
	L1BlockNumber *hexutil.Big   `json:"l1BlockNumber"`
}

type FinalityOverride int

const (
	NoOverride FinalityOverride = iota
	Confirmation
)

func ParseFinalityOverride(s string) (FinalityOverride, error) {
	switch strings.ToLower(s) {
	case "":
		return NoOverride, nil
	case strings.ToLower(Confirmation.String()):
		return Confirmation, nil
	default:
		return -1, fmt.Errorf("invalid finality override option")
	}
}
