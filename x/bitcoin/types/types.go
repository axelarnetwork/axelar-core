package types

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Bitcoin network types
var (
	Mainnet  = Network{&chaincfg.MainNetParams}
	Testnet3 = Network{&chaincfg.TestNet3Params}
	Regtest  = Network{&chaincfg.RegressionNetParams}
)

// OutPointInfo describes all the necessary information to verify the outPoint of a transaction
type OutPointInfo struct {
	OutPoint      *wire.OutPoint
	Amount        btcutil.Amount
	BlockHash     *chainhash.Hash
	Address       string
	Confirmations uint64
}

// Validate ensures that all fields are filled with sensible values
func (i OutPointInfo) Validate() error {
	if i.OutPoint == nil {
		return fmt.Errorf("missing outpoint")
	}
	if i.BlockHash == nil {
		return fmt.Errorf("missing block hash")
	}
	if i.Amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}
	if i.Address == "" {
		return fmt.Errorf("invalid address to track")
	}
	return nil
}

// Equals checks if two OutPointInfo objects are semantically equal
func (i OutPointInfo) Equals(other OutPointInfo) bool {
	return i.OutPoint.Hash.IsEqual(&other.OutPoint.Hash) &&
		i.OutPoint.Index == other.OutPoint.Index &&
		i.Amount == other.Amount &&
		i.Address == other.Address
}

// Network provides additional functionality based on the bitcoin network name
type Network struct {
	// Params returns the configuration parameters associated with the chain
	Params *chaincfg.Params
}

// NetworkFromStr returns network given string
func NetworkFromStr(net string) (Network, error) {
	switch net {
	case "main":
		return Mainnet, nil
	case "test":
		return Testnet3, nil
	case "regtest":
		return Regtest, nil
	default:
		return Network{}, fmt.Errorf("unknown network: %s", net)
	}
}

// Validate checks if the object is a valid chain
func (n Network) Validate() error {
	if n.Params == nil {
		return fmt.Errorf("network could not be parsed, choose main, test, or regtest")
	}
	return nil
}

// RawTxParams describe the parameters used to create a raw unsigned transaction for Bitcoin
type RawTxParams struct {
	OutPoint    *wire.OutPoint
	DepositAddr string
	Satoshi     sdk.Coin
}

// CreateTx returns a new unsigned Bitcoin transaction
func CreateTx(prevOuts []*wire.OutPoint, outputs []Output) (*wire.MsgTx, error) {
	tx := wire.NewMsgTx(wire.TxVersion)
	for _, outPoint := range prevOuts {
		// The signature script or witness will be set later
		txIn := wire.NewTxIn(outPoint, nil, nil)
		tx.AddTxIn(txIn)
	}
	for _, out := range outputs {
		addrScript, err := txscript.PayToAddrScript(out.Recipient)
		if err != nil {
			return nil, sdkerrors.Wrap(err, "could not create pay-to-address script for destination address")
		}
		txOut := wire.NewTxOut(int64(out.Amount), addrScript)
		tx.AddTxOut(txOut)
	}

	return tx, nil
}

// OutPointFromStr returns the parsed outpoint from a string of the form "txID:voutIdx"
func OutPointFromStr(outStr string) (*wire.OutPoint, error) {
	outParams := strings.Split(outStr, ":")
	if len(outParams) != 2 {
		return nil, fmt.Errorf("you must pass the outpoint identifier in the form of \"txID:voutIdx\"")
	}

	v, err := strconv.ParseUint(outParams[1], 10, 32)
	if err != nil {
		return nil, err
	}
	hash, err := chainhash.NewHashFromStr(outParams[0])
	if err != nil {
		return nil, err
	}

	out := wire.NewOutPoint(hash, uint32(v))
	return out, nil
}

type Output struct {
	Amount    btcutil.Amount
	Recipient btcutil.Address
}
