package exported

import (
	"crypto/ecdsa"

	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcutil"
)

func PubkeyToAddress(pk ecdsa.PublicKey, network string) (string, error) {

	net := types.Network(network)

	err := net.Validate()

	if err != nil {
		return "", err
	}

	btcPK := btcec.PublicKey(pk)

	addr, err := btcutil.NewAddressPubKeyHash(btcutil.Hash160(btcPK.SerializeCompressed()), net.Params())

	if err != nil {
		return "", err
	}

	return addr.String(), nil
}
