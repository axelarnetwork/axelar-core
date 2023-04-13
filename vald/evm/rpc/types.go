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

// zkEvmPolygonHeader represents a batch header in the zkEVM polygon blockchain
type zkEvmPolygonHeader struct {
	VerifyBatchTxHash   *common.Hash   `json:"verifyBatchTxHash"     gencodec:"required"`
	SendSequencesTxHash *common.Hash   `json:"sendSequencesTxHash"   gencodec:"required"`
	AccInputHash        common.Hash    `json:"accInputHash"          gencodec:"required"`
	LocalExitRoot       common.Hash    `json:"localExitRoot"         gencodec:"required"`
	GlobalExitRoot      common.Hash    `json:"globalExitRoot"        gencodec:"required"`
	StateRoot           common.Hash    `json:"stateRoot"             gencodec:"required"`
	Transactions        []common.Hash  `json:"transactions"          gencodec:"required"`
	Time                hexutil.Uint64 `json:"timestamp"             gencodec:"required"`
	Number              *hexutil.Big   `json:"number"                gencodec:"required"`
}
