package application

import (
	"errors"

	"github.com/haarchri/zpa-go-client/pkg/client/application_controller"
)

// IsNotFound returns whether the given error is of type NotFound or not.
func IsNotFound(err error) bool {
	// 404 NotFound is in API Response 400 BadRequest ApplicationSegment NotFound
	var badRequest *application_controller.GetApplicationUsingGET1BadRequest
	return errors.As(err, &badRequest)
}
