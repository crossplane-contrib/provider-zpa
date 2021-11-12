package application

import (
	"github.com/haarchri/zpa-go-client/pkg/client/application_controller"
)

// IsNotFound returns whether the given error is of type NotFound or not.
func IsNotFound(err error) bool {
	_, ok := err.(*application_controller.GetApplicationUsingGET1NotFound)
	return ok
}
