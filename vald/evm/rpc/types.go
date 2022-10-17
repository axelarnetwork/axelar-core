package rpc

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

// Header represents a block header in any EVM blockchain
type Header struct {
	ParentHash   common.Hash      `json:"parentHash"       gencodec:"required"`
	UncleHash    common.Hash      `json:"sha3Uncles"       gencodec:"required"`
	Coinbase     common.Address   `json:"miner"`
	Root         common.Hash      `json:"stateRoot"        gencodec:"required"`
	TxHash       common.Hash      `json:"transactionsRoot" gencodec:"required"`
	ReceiptHash  common.Hash      `json:"receiptsRoot"     gencodec:"required"`
	Bloom        types.Bloom      `json:"logsBloom"        gencodec:"required"`
	Difficulty   *hexutil.Big     `json:"difficulty"       gencodec:"required"`
	Number       *hexutil.Big     `json:"number"           gencodec:"required"`
	GasLimit     hexutil.Uint64   `json:"gasLimit"         gencodec:"required"`
	GasUsed      hexutil.Uint64   `json:"gasUsed"          gencodec:"required"`
	Time         hexutil.Uint64   `json:"timestamp"        gencodec:"required"`
	Extra        hexutil.Bytes    `json:"extraData"        gencodec:"required"`
	MixDigest    common.Hash      `json:"mixHash"`
	Nonce        types.BlockNonce `json:"nonce"`
	BaseFee      *hexutil.Big     `json:"baseFeePerGas" rlp:"optional"`
	Hash         common.Hash      `json:"hash"`
	Transactions []common.Hash    `json:"transactions"     gencodec:"required"`
}

// MoonbeamHeader represents a block header in the Moonbeam blockchain
type MoonbeamHeader struct {
	ParentHash     common.Hash  `json:"parentHash"       gencodec:"required"`
	ExtrinsicsRoot common.Hash  `json:"extrinsicsRoot"   gencodec:"required"`
	StateRoot      common.Hash  `json:"stateRoot"        gencodec:"required"`
	Number         *hexutil.Big `json:"number"           gencodec:"required"`
}
