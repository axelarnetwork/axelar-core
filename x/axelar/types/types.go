package types

type ExternalChainAddress struct {
	Chain   string
	Address string
}

func (a ExternalChainAddress) IsValid() bool {
	return a.Chain != "" && a.Address != ""
}
