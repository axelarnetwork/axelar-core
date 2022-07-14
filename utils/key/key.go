package key

import (
	"bytes"
	"strings"

	"github.com/axelarnetwork/utils/convert"
	"golang.org/x/exp/constraints"
)

const DefaultDelimiter = "_"

type Key interface {
	Append(Key) Key
	Bytes(delimiter ...string) []byte
}

type basicKey struct {
	particles [][]byte
}

func (k basicKey) Append(suffix Key) Key {
	var bz [][]byte
	switch suffix := suffix.(type) {
	case basicKey:
		bz = suffix.particles
	default:
		bz = [][]byte{suffix.Bytes()}
	}
	return basicKey{
		particles: append(k.particles, bz...),
	}
}

func (k basicKey) Bytes(delimiter ...string) []byte {
	del := DefaultDelimiter
	if len(delimiter) == 1 {
		del = delimiter[0]
	}
	return bytes.Join(k.particles, []byte(del))
}

func FromBz(key []byte) Key {
	return &basicKey{particles: [][]byte{key}}
}

func FromStr(key string) Key {
	return FromBz([]byte(strings.ToLower(key)))
}

func FromUInt[T constraints.Unsigned](key T) Key {
	return FromBz(convert.IntToBytes[T](key))
}
