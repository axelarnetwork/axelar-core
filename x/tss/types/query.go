package types

import "github.com/axelarnetwork/axelar-core/x/tss/exported"

// NewQueryNextKeyIDRequest returns a new QueryNextKeyIDRequest
func NewQueryNextKeyIDRequest(chain string, roleStr string) (QueryNextKeyIDRequest, error) {
	role, err := exported.KeyRoleFromSimpleStr(roleStr)
	if err != nil {
		return QueryNextKeyIDRequest{}, err
	}

	request := QueryNextKeyIDRequest{
		Chain:   chain,
		KeyRole: role,
	}
	return request, nil
}
