package rpc

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// Header represents a block header in any EVM blockchain
type Header struct {
	ParentHash    common.Hash    `json:"parentHash"       gencodec:"required"`
	Number        *hexutil.Big   `json:"number"           gencodec:"required"`
	Time          hexutil.Uint64 `json:"timestamp"        gencodec:"required"`
	Hash          common.Hash    `json:"hash"`
	Transactions  []common.Hash  `json:"transactions"     gencodec:"required"`
	L1BlockNumber *hexutil.Big   `json:"l1BlockNumber"`
}

// moonbeamHeader represents a block header in the Moonbeam blockchain
type moonbeamHeader struct {
	ParentHash     common.Hash  `json:"parentHash"       gencodec:"required"`
	ExtrinsicsRoot common.Hash  `json:"extrinsicsRoot"   gencodec:"required"`
	StateRoot      common.Hash  `json:"stateRoot"        gencodec:"required"`
	Number         *hexutil.Big `json:"number"           gencodec:"required"`
}

type optimismRollupGasPrices struct {
	L1GasPrice *hexutil.Big `json:"l1GasPrice"           gencodec:"required"`
	L2GasPrice *hexutil.Big `json:"l2GasPrice"           gencodec:"required"`
}
