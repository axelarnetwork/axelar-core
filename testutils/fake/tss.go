package fake

import (
	"crypto/ecdsa"
	"crypto/rand"
	"sync"

	"github.com/btcsuite/btcd/btcec/v2"
	ec "github.com/btcsuite/btcd/btcec/v2/ecdsa"

	"github.com/axelarnetwork/utils/funcs"
)

// Tofnd is a thread-safe fake that emulates the external tofnd process
type Tofnd struct {
	keyMutex *sync.RWMutex
	keys     map[string]*ecdsa.PrivateKey
	sigMutex *sync.RWMutex
	sigs     map[string][]byte
}

// NewTofnd returns a new Tofnd instance
func NewTofnd() *Tofnd {
	return &Tofnd{
		keyMutex: &sync.RWMutex{},
		keys:     map[string]*ecdsa.PrivateKey{},
		sigMutex: &sync.RWMutex{},
		sigs:     map[string][]byte{},
	}
}

// KeyGen simulates a distributed key generation. Only the first call with the same keyID creates a new key, every consecutive call returns the same one
func (t *Tofnd) KeyGen(keyID string) []byte {
	t.keyMutex.Lock()
	defer t.keyMutex.Unlock()

	var err error
	sk, ok := t.keys[keyID]
	if !ok {
		sk, err = ecdsa.GenerateKey(btcec.S256(), rand.Reader)
		if err != nil {
			panic(err)
		}
		t.keys[keyID] = sk
	}

	var x, y btcec.FieldVal
	x.SetByteSlice(sk.PublicKey.X.Bytes())
	y.SetByteSlice(sk.PublicKey.Y.Bytes())
	pk := btcec.NewPublicKey(&x, &y)
	return pk.SerializeCompressed()
}

// Sign simulates a distributed signature generation. Only the first call with the same sigID creates a new signature from the given key,
// every consecutive call returns the same one
func (t *Tofnd) Sign(sigID string, keyID string, msg []byte) []byte {
	sk := t.getPrivateKey(keyID)
	t.sigMutex.Lock()
	defer t.sigMutex.Unlock()
	sig, ok := t.sigs[sigID]
	if !ok {
		sig = createSignature(sk, msg)
		t.sigs[sigID] = sig
	}
	return sig
}

// HasKey returns true if it holds the key associated with the specified ID
func (t *Tofnd) HasKey(keyID string) bool {
	return t.getPrivateKey(keyID) != nil
}

func (t *Tofnd) getPrivateKey(keyID string) *ecdsa.PrivateKey {
	t.keyMutex.RLock()
	defer t.keyMutex.RUnlock()

	sk, ok := t.keys[keyID]
	if !ok {
		return nil
	}
	return sk
}

func createSignature(key *ecdsa.PrivateKey, msg []byte) []byte {
	encodedSig, err := ecdsa.SignASN1(rand.Reader, key, msg)
	if err != nil {
		panic(err)
	}
	sig := funcs.Must(ec.ParseDERSignature(encodedSig)).Serialize()
	return sig
}
