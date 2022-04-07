package exported

import (
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
)

// GetPermissionRole returns the role that is defined for the given message. Returns ROLE_UNSPECIFIED if none is set.
func GetPermissionRole(message descriptor.Message) Role {
	_, d := descriptor.ForMessage(message)
	v, err := proto.GetExtension(d.GetOptions(), E_PermissionRole)
	if err != nil {
		return ROLE_UNSPECIFIED
	}
	return *v.(*Role)
}
