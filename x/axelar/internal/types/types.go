package types

type ExternalChainAddress struct {
	Chain   string
	Address []byte
}

func (a ExternalChainAddress) IsValid() bool {
	return a.Chain != "" && a.Address != nil
}
