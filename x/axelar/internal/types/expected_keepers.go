package types

type BridgeKeeper interface {
	TrackAddress(address []byte)
}
