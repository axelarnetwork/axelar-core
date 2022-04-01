package utils

import (
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"

	"github.com/axelarnetwork/axelar-core/x/permission/exported"
)

// GetPermissionRole returns the role that is defined for the given message. Returns ROLE_CHAIN_MANAGEMENT if none is set.
func GetPermissionRole(message descriptor.Message) exported.Role {
	_, d := descriptor.ForMessage(message)
	v, err := proto.GetExtension(d.GetOptions(), E_PermissionRole)
	if err != nil {
		return exported.ROLE_CHAIN_MANAGEMENT
	}
	return *v.(*exported.Role)
}
