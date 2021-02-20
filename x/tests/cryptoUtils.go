package tests

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"sync"
	"time"

	"github.com/axelarnetwork/tssd/convert"
	tssd "github.com/axelarnetwork/tssd/pb"

	tssdMock "github.com/axelarnetwork/axelar-core/x/tss/types/mock"
)

// ensure all nodes call .Send() , .Recv() and then .CloseSend()
func prepareKeygen(keygen *tssdMock.TSSDKeyGenClientMock, keyID string, key ecdsa.PublicKey) (successful <-chan bool) {
	closeTimeout, closeCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)

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

func prepareSign(sign *tssdMock.TSSDSignClientMock, keyID string, key *ecdsa.PrivateKey, syncedSig *syncedBytes) <-chan bool {
	closeTimeout, closeCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)

	sendSuccessful := false
	recvSuccessful := false
	closeSuccessful := false

	doneSend := make(chan struct{})
	var msgToSign []byte
	sign.SendFunc = func(msg *tssd.MessageIn) error {
		defer close(doneSend)

		sendSuccessful = keyID == msg.GetSignInit().KeyUid
		msgToSign = msg.GetSignInit().MessageToSign

		return nil
	}

	sign.RecvFunc = func() (*tssd.MessageOut, error) {
		// keygen should only receive a response after sending something
		<-doneSend

		syncedSig.Set(createSignature(key, msgToSign))
		recvSuccessful = true
		return &tssd.MessageOut{Data: &tssd.MessageOut_SignResult{SignResult: syncedSig.Get()}}, nil
	}

	sign.CloseSendFunc = func() error {
		defer closeCancel()
		// close must be called last
		if recvSuccessful {
			closeSuccessful = true
		}

		return nil
	}

	allSuccessful := make(chan bool)
	go func() {
		// assert tssd was properly called
		<-closeTimeout.Done()
		allSuccessful <- sendSuccessful && recvSuccessful && closeSuccessful
	}()

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

type syncedBytes struct {
	once  *sync.Once
	isSet chan struct{}
	value []byte
}

// NewSyncedBytes returns a new syncedBytes object. It is a write once, read many times structure.
func NewSyncedBytes() *syncedBytes {
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
