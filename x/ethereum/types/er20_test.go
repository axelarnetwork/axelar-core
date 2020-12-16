package types

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

const (
	erc20Transfer    = "transfer(address,uint256)"
	erc20TransferSel = "0xa9059cbb"
	erc20Addr        = "0x000000000000000000000000337c67618968370907da31daef3020238d01c9de"
	erc20Val         = "0x0000000000000000000000000000000000000000000000008ac7230489e80000"
)

/*
This test is based in the following tutorial about ERC20 parameter serialization:

https://medium.com/swlh/understanding-data-payloads-in-ethereum-transactions-354dbe995371
https://medium.com/mycrypto/why-do-we-need-transaction-data-39c922930e92
*/
func TestERC20Marshal(t *testing.T) {

	// test function selector
	assert.Equal(t, erc20TransferSel, calcSelector(erc20Transfer))

	addr := EthAddress{
		Chain:         "fubar",
		EncodedString: "0x337c67618968370907da31daef3020238d01c9de",
	}

	// test first parameter (the address)
	paddedAddr := hexutil.Encode(common.LeftPadBytes(addr.Convert().Bytes(), 32))

	assert.Equal(t, erc20Addr, paddedAddr)

	// test second parameter (the amount)
	val, ok := big.NewInt(0).SetString("10000000000000000000", 10)
	assert.True(t, ok)

	paddedVal := hexutil.Encode(common.LeftPadBytes(val.Bytes(), 32))

	assert.Equal(t, erc20Val, paddedVal)

	var data []byte

	data = append(data, erc20TransferSel...)
	data = append(data, paddedAddr...)
	data = append(data, paddedVal...)

	// The number of bytes could have been slashed in half if we were supposed
	// to covert the hexadecimal value to bytes, instead of typecasting...
	assert.Equal(t, 142, len(data))

	concat := []byte(erc20TransferSel + erc20Addr + erc20Val)

	assert.Equal(t, concat, data)

}
