package wallet

import (
	"fmt"
	cliKeyring "github.com/cosmos/cosmos-sdk/client/keys"
	keyring "github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"io/ioutil"
	"os"
	"strings"

	broadcastTypes "github.com/axelarnetwork/axelar-core/x/broadcast/types"
)

// Wallet is a similar context to the broadcast keeper
type Wallet struct {
	keybase keyring.Keybase
	//EncodeTx sdk.TxEncoder
	Config WalletConfig

	FromAddr sdk.AccAddress
	//Account  Account
	AccountNumber uint64
	SequenceNumber     uint64
}

type WalletConfig struct {
	broadcastTypes.ClientConfig
	AppName       string // keybase app name
	RootDir       string // keybase root dir
	Gas string
	GasFees sdk.Coins
	GasPrices sdk.DecCoins
}

// Temporary placeholder for proper account store
type Account struct {
}

func ReadMnemonicFromFile(fname string) (string,error) {
	file, err := os.Open(fname)
	if file != nil {
		defer file.Close()
	}
	if err != nil {
		return "", err
	}

	b, err := ioutil.ReadAll(file)
	return strings.TrimSpace(string(b)), nil
}

func DefaultConfig() *WalletConfig {
	return &WalletConfig{
		ClientConfig: broadcastTypes.ClientConfig{
			KeyringBackend:    "test",
			TendermintNodeUri: "",
			ChainID:           "axelar",
			BroadcastConfig: broadcastTypes.BroadcastConfig{
				From:              "",
				KeyringPassphrase: "",
				GasAdjustment:     0,
			},
		},
		AppName:      "abtcd",
		RootDir:      "keytest",
		Gas: "",
		//GasPrices: "",
		//GasFees: "",
	}
}

func CreateWallet(config WalletConfig) (Wallet, error) {
	//defaultConfig := *DefaultConfig()
	//if config == WalletConfig{} {
	//	config = defaultConfig
	//}

	keybase, err := keyring.NewKeyring(config.AppName, config.KeyringBackend, config.RootDir, os.Stdin)
	if err != nil {
		return Wallet{}, err
	}

	return NewWallet(keybase, config, sdk.AccAddress("cosmos1tvz9j7lll27mcfdtk85j24dutk53m3pjfzaxsq"),4, 1), nil
}

func NewWallet(keybase keyring.Keybase, config WalletConfig, fromAddr sdk.AccAddress, accountNumber uint64, sequenceNumber uint64) Wallet {
	return Wallet{
		keybase: keybase,
		Config:  config,
		FromAddr: fromAddr,
		AccountNumber: accountNumber,
		SequenceNumber: sequenceNumber,
	}
}

func (w *Wallet) ImportMnemonicFromFile (mnemonicFile string) error {
	mnemonic, err := ReadMnemonicFromFile(mnemonicFile)
	if err != nil {
		return err
	}

	// Empty algo parameter will default to keys.Secp256k1
	keyInfo, err := w.keybase.CreateAccount(w.Config.From, mnemonic, keyring.DefaultBIP39Passphrase, cliKeyring.DefaultKeyPass,"", keyring.Secp256k1)
	if err != nil {
		return err
	}

	fmt.Printf("Key info from mnemonic file:\n%+v\n", keyInfo)
	return nil
}

// Import account using an encrypted ASCII armor private key file
//func (w *Wallet) ImportArmorFromFile() error {
//}

// SignStdTx appends a signature to a StdTx and returns a copy of it. If append
// is false, it replaces the signatures already attached with the new signature.
//func (w Wallet) SignStdTx(name, passphrase string, stdTx auth.StdTx, appendSig bool) (signedStdTx auth.StdTx, err error) {
func (w Wallet) SignStdTx(stdTx auth.StdTx, appendSig bool) (signedStdTx auth.StdTx, err error) {
	if w.Config.ChainID == "" {
		return auth.StdTx{}, fmt.Errorf("chain ID required but not specified")
	}

	stdSignature, err := w.makeSignature(auth.StdSignMsg{
		ChainID:       w.Config.ChainID,
		AccountNumber: w.AccountNumber,
		Sequence:      w.SequenceNumber,
		Fee:           stdTx.Fee,
		Msgs:          stdTx.GetMsgs(),
		Memo:          stdTx.GetMemo(),
	})
	if err != nil {
		return
	}

	sigs := stdTx.Signatures
	if len(sigs) == 0 || !appendSig {
		sigs = []auth.StdSignature{stdSignature}
	} else {
		sigs = append(sigs, stdSignature)
	}
	signedStdTx = auth.NewStdTx(stdTx.GetMsgs(), stdTx.Fee, sigs, stdTx.GetMemo())
	return
}

func (w Wallet) Sign(msg auth.StdSignMsg) (auth.StdTx, error) {
	sig, err := w.makeSignature(msg)
	if err != nil {
		return auth.StdTx{}, err
	}

	return auth.NewStdTx(msg.Msgs, msg.Fee, []auth.StdSignature{sig}, msg.Memo), nil
}

func (w Wallet) makeSignature(msg auth.StdSignMsg) (auth.StdSignature, error) {
	sigBytes, pubkey, err := w.keybase.Sign(w.Config.From, w.Config.KeyringPassphrase, msg.Bytes())
	if err != nil {
		return auth.StdSignature{}, err
	}

	return auth.StdSignature{
		PubKey:    pubkey,
		Signature: sigBytes,
	}, nil
}