package server

import (
	"github.com/haarchri/zpa-go-client/pkg/client/app_server_controller"
)

// IsNotFound returns whether the given error is of type NotFound or not.
func IsNotFound(err error) bool {
	// 404 NotFound is in API Response 400 BadRequest Server NotFound
	_, ok := err.(*app_server_controller.GetAppServerUsingGET1BadRequest)
	return ok
}
