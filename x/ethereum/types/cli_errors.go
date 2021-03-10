package types

const (
	ErrFMasterKey      = "could not resolve master key: %s\n"
	ErrFGatewayAddress = "could not resolve gateway address: %s\n"
	ErrFTokenAddress   = "could not resolve token address: %s\n"
	ErrFDeployTx       = "could not send the command transaction with txID %s"
	ErrFSendTx         = "could not send the deploy transaction with txID %s"
	ErrFSendCommandTx  = "could not send Ethereum transaction executing command %s"
	ErrFSendMintTx     = "could not send Ethereum transaction executing mint command %s"
)
