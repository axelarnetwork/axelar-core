package exported

import (
	"reflect"
	"sync"

	"github.com/cosmos/gogoproto/proto"
	"github.com/cosmos/gogoproto/protoc-gen-gogo/descriptor"
)

// roleCache memoizes the permission role per concrete message type.
var roleCache sync.Map // reflect.Type -> Role

// GetPermissionRole returns the role that is defined for the given message. Returns ROLE_UNSPECIFIED if none is set.
func GetPermissionRole(message descriptor.Message) Role {
	key := reflect.TypeOf(message)
	if cached, ok := roleCache.Load(key); ok {
		return cached.(Role)
	}

	role := resolvePermissionRole(message)
	roleCache.Store(key, role)

	return role
}

func resolvePermissionRole(message descriptor.Message) Role {
	_, d := descriptor.ForMessage(message)
	v, err := proto.GetExtension(d.GetOptions(), E_PermissionRole)
	if err != nil {
		return ROLE_UNSPECIFIED
	}
	return *v.(*Role)
}
