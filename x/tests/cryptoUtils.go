package tests

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"sync"
	"time"

	"github.com/axelarnetwork/tssd/convert"
	tssd "github.com/axelarnetwork/tssd/pb"
	"google.golang.org/grpc"

	tssdMock "github.com/axelarnetwork/axelar-core/x/tss/types/mock"
)

// ensure all nodes call .Send() , .Recv() and then .CloseSend()
func prepareKeygen(keygen *tssdMock.TSSDKeyGenClientMock, keyID string, key ecdsa.PublicKey) (successful <-chan bool) {
	closeTimeout, closeCancel := context.WithTimeout(context.Background(), 5*time.Second)

	sendSuccessful := false
	recvSuccessful := false
	closeSuccessful := false

	doneSend := make(chan struct{})
	keygen.SendFunc = func(in *tssd.MessageIn) error {
		defer close(doneSend)
		sendSuccessful = keyID == in.GetKeygenInit().NewKeyUid
		return nil
	}

	keygen.RecvFunc = func() (*tssd.MessageOut, error) {
		// keygen should only receive a response after sending something
		<-doneSend

		pk, err := convert.PubkeyToBytes(key)
		if err != nil {
			panic(err)
		}
		recvSuccessful = true
		return &tssd.MessageOut{Data: &tssd.MessageOut_KeygenResult{KeygenResult: pk}}, nil
	}

	keygen.CloseSendFunc = func() error {
		defer closeCancel()
		// close must be called last
		if recvSuccessful {
			closeSuccessful = true
		}

		return nil
	}

	allSuccessful := make(chan bool)
	go func() {
		<-closeTimeout.Done()
		allSuccessful <- sendSuccessful && recvSuccessful && closeSuccessful
	}()

	return allSuccessful
}

func prepareSign(mock *tssdMock.TSSDClientMock, keyID string, key *ecdsa.PrivateKey, signatureCache []*syncedBytes) <-chan bool {
	allSuccessful := make(chan bool, len(signatureCache))

	var msgToSign []byte
	mock.SignFunc = func(ctx context.Context, opts ...grpc.CallOption) (tssd.GG18_SignClient, error) {
		closeTimeout, closeCancel := context.WithTimeout(context.Background(), 5*time.Second)
		doneSend := make(chan struct{})

		sendSuccessful := false
		recvSuccessful := false
		closeSuccessful := false

		go func() {
			// assert sign was properly called
			<-closeTimeout.Done()
			allSuccessful <- sendSuccessful && recvSuccessful && closeSuccessful
		}()

		sig := signatureCache[len(mock.SignCalls())-1]
		return &tssdMock.TSSDSignClientMock{
			SendFunc: func(msg *tssd.MessageIn) error {
				defer close(doneSend)

				sendSuccessful = keyID == msg.GetSignInit().KeyUid
				msgToSign = msg.GetSignInit().MessageToSign

				sig.Set(createSignature(key, msgToSign))

				return nil
			},
			RecvFunc: func() (*tssd.MessageOut, error) {
				// keygen should only receive a response after sending something
				<-doneSend
				recvSuccessful = true
				return &tssd.MessageOut{Data: &tssd.MessageOut_SignResult{SignResult: sig.Get()}}, nil
			},
			CloseSendFunc: func() error {
				defer closeCancel()
				// close must be called last
				if recvSuccessful {
					closeSuccessful = true
				}

				return nil
			}}, nil
	}

	return allSuccessful
}

func createSignature(key *ecdsa.PrivateKey, msg []byte) []byte {
	r, s, err := ecdsa.Sign(rand.Reader, key, msg)
	if err != nil {
		panic(err)
	}
	sig, err := convert.SigToBytes(r.Bytes(), s.Bytes())
	if err != nil {
		panic(err)
	}
	return sig
}

// NewSignatureCache returns an empty cache of length sigCount. Each entry in the cache can be written once, but read many times.
func NewSignatureCache(sigCount int) []*syncedBytes {
	var sigCache []*syncedBytes
	for i := 0; i < sigCount; i++ {
		sigCache = append(sigCache, newSyncedBytes())
	}
	return sigCache
}

type syncedBytes struct {
	once  *sync.Once
	isSet chan struct{}
	value []byte
}

func newSyncedBytes() *syncedBytes {
	return &syncedBytes{
		once:  &sync.Once{},
		isSet: make(chan struct{}, 1),
		value: nil,
	}
}

// Set stores the byte value exactly once and ignores all consecutive calls
func (s *syncedBytes) Set(v []byte) {
	s.once.Do(func() {
		s.value = v
		s.isSet <- struct{}{}
	})
}

// Get returns the set value. Blocks if value is not set.
func (s *syncedBytes) Get() []byte {
	<-s.isSet
	s.isSet <- struct{}{}
	return s.value
}
