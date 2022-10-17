package rpc

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

// Header represents a block header in any EVM blockchain
type Header struct {
	types.Header
	Transactions []common.Hash `json:"transactions" gencodec:"required"`
}

// MoonbeamHeader represents a block header in the Moonbeam blockchain
type MoonbeamHeader struct {
	ParentHash     common.Hash  `json:"parentHash"       gencodec:"required"`
	ExtrinsicsRoot common.Hash  `json:"extrinsicsRoot"   gencodec:"required"`
	StateRoot      common.Hash  `json:"stateRoot"        gencodec:"required"`
	Number         *hexutil.Big `json:"number"           gencodec:"required"`
}
