package types

const (
	ErrFMasterKey     = "could not resolve master key: %s\n"
	ErrFMintTx        = ErrFMasterKey
	ErrFDeployTx      = ErrFMintTx
	ErrFSendTx        = "could not send the transaction spending transaction %s"
	ErrFSendCommandTx = "could not send Ethereum transaction executing command %s"
)
