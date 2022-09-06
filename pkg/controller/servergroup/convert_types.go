package servergroup

import (
	"errors"

	"github.com/haarchri/zpa-go-client/pkg/client/server_group_controller"
)

// IsNotFound returns whether the given error is of type NotFound or not.
func IsNotFound(err error) bool {
	// 404 NotFound is in API Response 400 BadRequest ServerGroup NotFound
	var badRequest *server_group_controller.GetServerGroupUsingGET1BadRequest
	return errors.As(err, &badRequest)
}
