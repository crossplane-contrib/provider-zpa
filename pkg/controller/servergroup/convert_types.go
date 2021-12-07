package servergroup

import (
	"github.com/haarchri/zpa-go-client/pkg/client/server_group_controller"
)

// IsNotFound returns whether the given error is of type NotFound or not.
func IsNotFound(err error) bool {
	// 404 NotFound is in API Response 400 BadRequest ServerGroup NotFound
	_, ok := err.(*server_group_controller.GetServerGroupUsingGET1BadRequest)
	return ok
}
