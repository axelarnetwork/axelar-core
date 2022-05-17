package exported2_test

import (
	"fmt"
	"testing"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	. "github.com/axelarnetwork/axelar-core/x/tss/exported2"
)

func TestOneOf(t *testing.T) {
	sig := Signature{
		SigID: rand.Str(10),
		Sig: &Signature_MultiSig_{MultiSig: &Signature_MultiSig{SigKeyPairs: []SigKeyPair{{
			PubKey:    rand.Bytes(20),
			Signature: rand.Bytes(30),
		}}}},
		SigStatus: 0,
	}

	cdc := app.MakeEncodingConfig().Codec
	bz := cdc.MustMarshalLengthPrefixed(&sig)

	var sig2 exported.Signature
	cdc.MustUnmarshalLengthPrefixed(bz, &sig2)

	fmt.Println(sig2)
}
