package application

import (
	"github.com/haarchri/zpa-go-client/pkg/client/application_controller"
)

// IsNotFound returns whether the given error is of type NotFound or not.
func IsNotFound(err error) bool {
	// 404 NotFound is in API Response 400 BadRequest ApplicationSegment NotFound
	_, ok := err.(*application_controller.GetApplicationUsingGET1BadRequest)
	return ok
}
