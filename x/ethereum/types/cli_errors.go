package types

const (
	ErrFMasterKey     = "could not resolve master key: %s"
	ErrFTxInfo        = "could not resolve transaction: %s"
	ErrFDeployTx      = "could not send the command transaction with txID %s"
	ErrFSendTx        = "could not send the deploy transaction with txID %s"
	ErrFSendCommandTx = "could not send Ethereum transaction executing command %s"
	ErrFSendMintTx    = "could not send Ethereum transaction executing mint command %s"
)
