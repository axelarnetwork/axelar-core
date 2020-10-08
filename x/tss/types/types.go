package types

type TSSPartyID []byte

type TSSParty struct {
	Moniker string     // convenience only - can be anything (even left blank)
	ID      TSSPartyID // unique identifying key for this peer (such as its p2p public key)
}
