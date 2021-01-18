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

var (
	Mainnet  = Network{&chaincfg.MainNetParams}
	Testnet3 = Network{&chaincfg.TestNet3Params}
	Regtest  = Network{&chaincfg.RegressionNetParams}
)

// OutPointInfo describes all the necessary information to verify the outPoint of a transaction
type OutPointInfo struct {
	OutPoint      *wire.OutPoint `json:"outpoint" yaml:"outpoint"`
	Amount        btcutil.Amount `json:"amount" yaml:"amount"`
	DepositAddr   string `json:"deposit_address" yaml:"deposit_address"`
	Confirmations uint64 `json:"confirmations" yaml:"confirmations"`
}

// Validate ensures that all fields are filled with sensible values
func (i OutPointInfo) Validate() error {
	if i.OutPoint == nil {
		return fmt.Errorf("missing outpoint")
	}
	if i.Amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}
	if i.DepositAddr == "" {
		return fmt.Errorf("invalid address to track")
	}
	return nil
}

// Equals checks if two OutPointInfo objects are semantically equal
func (i OutPointInfo) Equals(other OutPointInfo) bool {
	return i.OutPoint.Hash.IsEqual(&other.OutPoint.Hash) &&
		i.OutPoint.Index == other.OutPoint.Index &&
		i.Amount == other.Amount &&
		i.DepositAddr == other.DepositAddr
}

// Network provides additional functionality based on the bitcoin network name
type Network struct {
	// Params returns the configuration parameters associated with the chain
	Params *chaincfg.Params
}

func NetworkFromStr(net string) (Network, error) {
	switch net {
	case "main":
		return Network{&chaincfg.MainNetParams}, nil
	case "test":
		return Network{&chaincfg.TestNet3Params}, nil
	case "regtest":
		return Network{&chaincfg.RegressionNetParams}, nil
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

type RawTxParams struct {
	OutPoint    *wire.OutPoint
	DepositAddr string
	Satoshi     sdk.Coin
}

// CreateTx returns a new unsigned Bitcoin transaction
func CreateTx(outPoint *wire.OutPoint, satoshi sdk.Coin, recipient btcutil.Address) (*wire.MsgTx, error) {
	addrScript, err := txscript.PayToAddrScript(recipient)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not create pay-to-address script for destination address")
	}

	/*
		Creating a Bitcoin transaction one step at a time:
			1. Create the transaction message
			2. Get the output of the deposit transaction and convert it into the transaction input
			3. Create a new output
		See https://blog.hlongvu.com/post/t0xx5dejn3-Understanding-btcd-Part-4-Create-and-Sign-a-Bitcoin-transaction-with-btcd
	*/
	tx := wire.NewMsgTx(wire.TxVersion)

	// The signature script or witness will be set later
	txIn := wire.NewTxIn(outPoint, nil, nil)
	tx.AddTxIn(txIn)
	txOut := wire.NewTxOut(satoshi.Amount.Int64(), addrScript)
	tx.AddTxOut(txOut)
	return tx, nil
}

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
