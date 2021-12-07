package keeper

type NoCommandsToSignError struct{}

func (e *NoCommandsToSignError) Error() string {
	return "no commands to sign found"
}
