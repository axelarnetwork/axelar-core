package exported

import (
	"fmt"
)

// Validate validates the Role
func (x Role) Validate() error {
	_, ok := Role_name[int32(x)]
	if !ok {
		return fmt.Errorf("unknown gov role")
	}

	if x == ROLE_UNSPECIFIED || x == ROLE_UNRESTRICTED {
		return fmt.Errorf("unspecified gov role")
	}

	return nil
}
